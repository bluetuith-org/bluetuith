package hjson

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const maxPointerDepth = 512

// This limits the max nesting depth to prevent stack overflow.
const maxNestingDepth = 10000

type commentInfo struct {
	hasComment bool
	cmStart    int
	cmEnd      int
}

// If a destination type implements ElemTyper, Unmarshal() will call ElemType()
// on the destination when unmarshalling an array or an object, to see if any
// array element or leaf node should be of type string even if it can be treated
// as a number, boolean or null. This is most useful if the destination also
// implements the json.Unmarshaler interface, because then there is no other way
// for Unmarshal() to know the type of the elements on the destination. If a
// destination implements ElemTyper all of its elements must be of the same
// type.
type ElemTyper interface {
	// Returns the desired type of any elements. If ElemType() is implemented
	// using a pointer receiver it must be possible to call with nil as receiver.
	ElemType() reflect.Type
}

// DecoderOptions defines options for decoding Hjson.
type DecoderOptions struct {
	// UseJSONNumber causes the Decoder to unmarshal a number into an interface{} as a
	// json.Number instead of as a float64.
	UseJSONNumber bool
	// DisallowUnknownFields causes an error to be returned when the destination
	// is a struct and the input contains object keys which do not match any
	// non-ignored, exported fields in the destination.
	DisallowUnknownFields bool
	// DisallowDuplicateKeys causes an error to be returned if an object (map) in
	// the Hjson input contains duplicate keys. If DisallowDuplicateKeys is set
	// to false, later values will silently overwrite previous values for the
	// same key.
	DisallowDuplicateKeys bool
	// WhitespaceAsComments only has any effect when an hjson.Node struct (or
	// an *hjson.Node pointer) is used as target for Unmarshal. If
	// WhitespaceAsComments is set to true, all whitespace and comments are stored
	// in the Node structs so that linefeeds and custom indentation is kept. If
	// WhitespaceAsComments instead is set to false, only actual comments are
	// stored as comments in Node structs.
	WhitespaceAsComments bool
}

// DefaultDecoderOptions returns the default decoding options.
func DefaultDecoderOptions() DecoderOptions {
	return DecoderOptions{
		UseJSONNumber:         false,
		DisallowUnknownFields: false,
		DisallowDuplicateKeys: false,
		WhitespaceAsComments:  true,
	}
}

type hjsonParser struct {
	DecoderOptions
	data              []byte
	at                int  // The index of the current character
	ch                byte // The current character
	structTypeCache   map[reflect.Type]structFieldMap
	willMarshalToJSON bool
	nodeDestination   bool
	nestingDepth      int
}

var unmarshalerText = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
var elemTyper = reflect.TypeOf((*ElemTyper)(nil)).Elem()

func (p *hjsonParser) setComment1(pCm *string, ci commentInfo) {
	if ci.hasComment {
		*pCm = string(p.data[ci.cmStart:ci.cmEnd])
	}
}

func (p *hjsonParser) setComment2(pCm *string, ciA, ciB commentInfo) {
	if ciA.hasComment && ciB.hasComment {
		*pCm = string(p.data[ciA.cmStart:ciA.cmEnd]) + string(p.data[ciB.cmStart:ciB.cmEnd])
	} else {
		p.setComment1(pCm, ciA)
		p.setComment1(pCm, ciB)
	}
}

func (p *hjsonParser) resetAt() {
	p.at = 0
	p.nestingDepth = 0
	p.next()
}

func isPunctuatorChar(c byte) bool {
	return c == '{' || c == '}' || c == '[' || c == ']' || c == ',' || c == ':'
}

func (p *hjsonParser) errAt(message string) error {
	if p.at <= len(p.data) {
		var i int
		col := 0
		line := 1
		for i = p.at - 1; i > 0 && p.data[i] != '\n'; i-- {
			col++
		}
		for ; i > 0; i-- {
			if p.data[i] == '\n' {
				line++
			}
		}
		samEnd := p.at - col + 20
		if samEnd > len(p.data) {
			samEnd = len(p.data)
		}
		return fmt.Errorf("%s at line %d,%d >>> %s", message, line, col, string(p.data[p.at-col:samEnd]))
	}
	return errors.New(message)
}

func (p *hjsonParser) next() bool {
	// get the next character.
	if p.at < len(p.data) {
		p.ch = p.data[p.at]
		p.at++
		return true
	}
	p.at++
	p.ch = 0
	return false
}

func (p *hjsonParser) prev() bool {
	// get the previous character.
	if p.at > 1 {
		p.ch = p.data[p.at-2]
		p.at--
		return true
	}

	return false
}

func (p *hjsonParser) peek(offs int) byte {
	pos := p.at + offs
	if pos >= 0 && pos < len(p.data) {
		return p.data[p.at+offs]
	}
	return 0
}

var escapee = map[byte]byte{
	'"':  '"',
	'\'': '\'',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

func unravelDestination(dest reflect.Value, t reflect.Type) (reflect.Value, reflect.Type) {
	if dest.IsValid() {
		for a := 0; a < maxPointerDepth && (dest.Kind() == reflect.Ptr ||
			dest.Kind() == reflect.Interface) && !dest.IsNil(); a++ {

			dest = dest.Elem()
		}

		if dest.IsValid() {
			t = dest.Type()
		}
	}

	for a := 0; a < maxPointerDepth && t != nil && t.Kind() == reflect.Ptr; a++ {
		t = t.Elem()
	}

	return dest, t
}

func (p *hjsonParser) readString(allowML bool) (string, error) {

	// Parse a string value.
	res := new(bytes.Buffer)

	// callers make sure that (ch === '"' || ch === "'")
	// When parsing for string values, we must look for " and \ characters.
	exitCh := p.ch
	for p.next() {
		if p.ch == exitCh {
			p.next()
			if allowML && exitCh == '\'' && p.ch == '\'' && res.Len() == 0 {
				// ''' indicates a multiline string
				p.next()
				return p.readMLString()
			} else {
				return res.String(), nil
			}
		}
		if p.ch == '\\' {
			p.next()
			if p.ch == 'u' {
				uffff := 0
				for i := 0; i < 4; i++ {
					p.next()
					var hex int
					if p.ch >= '0' && p.ch <= '9' {
						hex = int(p.ch - '0')
					} else if p.ch >= 'a' && p.ch <= 'f' {
						hex = int(p.ch - 'a' + 0xa)
					} else if p.ch >= 'A' && p.ch <= 'F' {
						hex = int(p.ch - 'A' + 0xa)
					} else {
						return "", p.errAt("Bad \\u char " + string(p.ch))
					}
					uffff = uffff*16 + hex
				}
				res.WriteRune(rune(uffff))
			} else if ech, ok := escapee[p.ch]; ok {
				res.WriteByte(ech)
			} else {
				return "", p.errAt("Bad escape \\" + string(p.ch))
			}
		} else if p.ch == '\n' || p.ch == '\r' {
			return "", p.errAt("Bad string containing newline")
		} else {
			res.WriteByte(p.ch)
		}
	}
	return "", p.errAt("Bad string")
}

func (p *hjsonParser) readMLString() (value string, err error) {

	// Parse a multiline string value.
	res := new(bytes.Buffer)
	triple := 0

	// we are at ''' +1 - get indent
	indent := 0
	for {
		c := p.peek(-indent - 5)
		if c == 0 || c == '\n' {
			break
		}
		indent++
	}

	skipIndent := func() {
		skip := indent
		for p.ch > 0 && p.ch <= ' ' && p.ch != '\n' && skip > 0 {
			skip--
			p.next()
		}
	}

	// skip white/to (newline)
	for p.ch > 0 && p.ch <= ' ' && p.ch != '\n' {
		p.next()
	}
	if p.ch == '\n' {
		p.next()
		skipIndent()
	}

	// When parsing multiline string values, we must look for ' characters.
	lastLf := false
	for {
		if p.ch == 0 {
			return "", p.errAt("Bad multiline string")
		} else if p.ch == '\'' {
			triple++
			p.next()
			if triple == 3 {
				sres := res.Bytes()
				if lastLf {
					return string(sres[0 : len(sres)-1]), nil // remove last EOL
				}
				return string(sres), nil
			}
			continue
		} else {
			for triple > 0 {
				res.WriteByte('\'')
				triple--
				lastLf = false
			}
		}
		if p.ch == '\n' {
			res.WriteByte('\n')
			lastLf = true
			p.next()
			skipIndent()
		} else {
			if p.ch != '\r' {
				res.WriteByte(p.ch)
				lastLf = false
			}
			p.next()
		}
	}
}

func (p *hjsonParser) readKeyname() (string, error) {

	// quotes for keys are optional in Hjson
	// unless they include {}[],: or whitespace.

	if p.ch == '"' || p.ch == '\'' {
		return p.readString(false)
	}

	name := new(bytes.Buffer)
	start := p.at
	space := -1
	for {
		if p.ch == ':' {
			if name.Len() == 0 {
				return "", p.errAt("Found ':' but no key name (for an empty key name use quotes)")
			} else if space >= 0 && space != name.Len() {
				p.at = start + space
				return "", p.errAt("Found whitespace in your key name (use quotes to include)")
			}
			return name.String(), nil
		} else if p.ch <= ' ' {
			if p.ch == 0 {
				return "", p.errAt("Found EOF while looking for a key name (check your syntax)")
			}
			if space < 0 {
				space = name.Len()
			}
		} else {
			if isPunctuatorChar(p.ch) {
				return "", p.errAt("Found '" + string(p.ch) + "' where a key name was expected (check your syntax or use quotes if the key name includes {}[],: or whitespace)")
			}
			name.WriteByte(p.ch)
		}
		p.next()
	}
}

func (p *hjsonParser) commonWhite(onlyAfter bool) (commentInfo, bool) {
	ci := commentInfo{
		false,
		p.at - 1,
		0,
	}
	var hasLineFeed bool

	for p.ch > 0 {
		// Skip whitespace.
		for p.ch > 0 && p.ch <= ' ' {
			if p.ch == '\n' {
				hasLineFeed = true
				if onlyAfter {
					ci.cmEnd = p.at - 1
					// Skip EOL.
					p.next()
					return ci, hasLineFeed
				}
			}
			p.next()
		}
		// Hjson allows comments
		if p.ch == '#' || p.ch == '/' && p.peek(0) == '/' {
			ci.hasComment = p.nodeDestination
			for p.ch > 0 && p.ch != '\n' {
				p.next()
			}
		} else if p.ch == '/' && p.peek(0) == '*' {
			ci.hasComment = p.nodeDestination
			p.next()
			p.next()
			for p.ch > 0 && !(p.ch == '*' && p.peek(0) == '/') {
				p.next()
			}
			if p.ch > 0 {
				p.next()
				p.next()
			}
		} else {
			break
		}
	}

	// cmEnd is the first char after the comment (i.e. not included in the comment).
	ci.cmEnd = p.at - 1

	return ci, hasLineFeed
}

func (p *hjsonParser) white() commentInfo {
	ci, _ := p.commonWhite(false)

	ci.hasComment = (ci.hasComment || (p.WhitespaceAsComments && (ci.cmEnd > ci.cmStart)))

	return ci
}

func (p *hjsonParser) whiteAfterComma() commentInfo {
	ci, hasLineFeed := p.commonWhite(true)

	ci.hasComment = (ci.hasComment || (p.WhitespaceAsComments &&
		hasLineFeed && (ci.cmEnd > ci.cmStart)))

	return ci
}

func (p *hjsonParser) getCommentAfter() commentInfo {
	ci, _ := p.commonWhite(true)

	ci.hasComment = (ci.hasComment || (p.WhitespaceAsComments && (ci.cmEnd > ci.cmStart)))

	return ci
}

func (p *hjsonParser) maybeWrapNode(n *Node, v interface{}) (interface{}, error) {
	if p.nodeDestination {
		n.Value = v
		return n, nil
	}
	return v, nil
}

func (p *hjsonParser) readTfnns(dest reflect.Value, t reflect.Type) (interface{}, error) {

	// Hjson strings can be quoteless
	// returns string, (json.Number or float64), true, false, or null.
	// Or wraps the value in a Node.

	if isPunctuatorChar(p.ch) {
		return nil, p.errAt("Found a punctuator character '" + string(p.ch) + "' when expecting a quoteless string (check your syntax)")
	}
	chf := p.ch
	var node Node
	value := new(bytes.Buffer)
	value.WriteByte(p.ch)

	var newT reflect.Type
	if !p.nodeDestination {
		// Keep the original dest and t, because we need to check if it implements
		// encoding.TextUnmarshaler.
		_, newT = unravelDestination(dest, t)
	}

	for {
		p.next()
		isEol := p.ch == '\r' || p.ch == '\n' || p.ch == 0
		if isEol ||
			p.ch == ',' || p.ch == '}' || p.ch == ']' ||
			p.ch == '#' ||
			p.ch == '/' && (p.peek(0) == '/' || p.peek(0) == '*') {

			// Do not output anything else than a string if our destination is a string.
			// Pointer methods can be called if the destination is addressable,
			// therefore we also check if dest.Addr() implements encoding.TextUnmarshaler.
			// But "null" is a special case: unmarshal it as nil if the original
			// destination type is a pointer.
			if chf == 'n' && !p.nodeDestination && t != nil && t.Kind() == reflect.Ptr &&
				strings.TrimSpace(value.String()) == "null" {

				return p.maybeWrapNode(&node, nil)
			}
			if (newT == nil || newT.Kind() != reflect.String) &&
				(t == nil || !(t.Implements(unmarshalerText) ||
					dest.CanAddr() && dest.Addr().Type().Implements(unmarshalerText))) {

				switch chf {
				case 'f':
					if strings.TrimSpace(value.String()) == "false" {
						return p.maybeWrapNode(&node, false)
					}
				case 'n':
					if strings.TrimSpace(value.String()) == "null" {
						return p.maybeWrapNode(&node, nil)
					}
				case 't':
					if strings.TrimSpace(value.String()) == "true" {
						return p.maybeWrapNode(&node, true)
					}
				default:
					if chf == '-' || chf >= '0' && chf <= '9' {
						// Always use json.Number if we will marshal to JSON.
						if n, err := tryParseNumber(
							value.Bytes(),
							false,
							p.willMarshalToJSON || p.DecoderOptions.UseJSONNumber,
						); err == nil {
							return p.maybeWrapNode(&node, n)
						}
					}
				}
			}

			if isEol {
				// remove any whitespace at the end (ignored in quoteless strings)
				return p.maybeWrapNode(&node, strings.TrimSpace(value.String()))
			}
		}
		value.WriteByte(p.ch)
	}
}

// t must not have been unraveled
func getElemTyperType(rv reflect.Value, t reflect.Type) reflect.Type {
	var elemType reflect.Type
	isElemTyper := false

	if t != nil && t.Implements(elemTyper) {
		isElemTyper = true
		if t.Kind() == reflect.Ptr {
			// If ElemType() has a value receiver we would get a panic if we call it
			// on a nil pointer.
			if !rv.IsValid() || rv.IsNil() {
				rv = reflect.New(t.Elem())
			}
		} else if !rv.IsValid() {
			rv = reflect.Zero(t)
		}
	}
	if !isElemTyper && rv.CanAddr() {
		rv = rv.Addr()
		if rv.Type().Implements(elemTyper) {
			isElemTyper = true
		}
	}
	if !isElemTyper && t != nil {
		pt := reflect.PtrTo(t)
		if pt.Implements(elemTyper) {
			isElemTyper = true
			rv = reflect.Zero(pt)
		}
	}
	if isElemTyper {
		elemType = rv.Interface().(ElemTyper).ElemType()
	}

	return elemType
}

func (p *hjsonParser) readArray(dest reflect.Value, t reflect.Type) (value interface{}, err error) {
	var node Node

	if p.nestingDepth > maxNestingDepth {
		return nil, p.errAt(fmt.Sprintf("Exceeded max depth (%d)", maxNestingDepth))
	}

	array := make([]interface{}, 0, 1)

	// Skip '['.
	p.next()
	ciBefore := p.getCommentAfter()
	p.setComment1(&node.Cm.InsideFirst, ciBefore)
	ciBefore = p.white()

	if p.ch == ']' {
		p.setComment1(&node.Cm.InsideLast, ciBefore)
		p.next()
		return p.maybeWrapNode(&node, array) // empty array
	}

	var elemType reflect.Type
	if !p.nodeDestination {
		elemType = getElemTyperType(dest, t)

		dest, t = unravelDestination(dest, t)

		// All elements in any existing slice/array will be removed, so we only care
		// about the type of the new elements that will be created.
		if elemType == nil && t != nil && (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) {
			elemType = t.Elem()
		}
	}

	for p.ch > 0 {
		var elemNode *Node
		var val interface{}
		if val, err = p.readValue(reflect.Value{}, elemType); err != nil {
			return nil, err
		}
		if p.nodeDestination {
			var ok bool
			if elemNode, ok = val.(*Node); ok {
				p.setComment1(&elemNode.Cm.Before, ciBefore)
			}
		}
		// Check white before comma because comma might be on other line.
		ciAfter := p.white()
		// in Hjson the comma is optional and trailing commas are allowed
		if p.ch == ',' {
			p.next()
			ciAfterComma := p.whiteAfterComma()
			if elemNode != nil {
				existingAfter := elemNode.Cm.After
				p.setComment2(&elemNode.Cm.After, ciAfter, ciAfterComma)
				elemNode.Cm.After = existingAfter + elemNode.Cm.After
			}
			// Any comments starting on the line after the comma.
			ciAfter = p.white()
		}
		if p.ch == ']' {
			p.setComment1(&node.Cm.InsideLast, ciAfter)
			array = append(array, val)
			p.next()
			return p.maybeWrapNode(&node, array)
		}
		array = append(array, val)
		ciBefore = ciAfter
	}

	return nil, p.errAt("End of input while parsing an array (did you forget a closing ']'?)")
}

func (p *hjsonParser) readObject(
	withoutBraces bool,
	dest reflect.Value,
	t reflect.Type,
	ciBefore commentInfo,
) (value interface{}, err error) {
	// Parse an object value.
	var node Node
	var elemNode *Node

	if p.nestingDepth > maxNestingDepth {
		return nil, p.errAt(fmt.Sprintf("Exceeded max depth (%d)", maxNestingDepth))
	}

	object := NewOrderedMap()

	// If withoutBraces == true we use the input argument ciBefore as
	// Before-comment on the first element of this obj, or as InnerLast-comment
	// on this obj if it doesn't contain any elements. If withoutBraces == false
	// we ignore the input ciBefore.

	if !withoutBraces {
		// assuming ch == '{'
		p.next()
		ciInsideFirst := p.getCommentAfter()
		p.setComment1(&node.Cm.InsideFirst, ciInsideFirst)
		ciBefore = p.white()
		if p.ch == '}' {
			p.setComment1(&node.Cm.InsideLast, ciBefore)
			p.next()
			return p.maybeWrapNode(&node, object) // empty object
		}
	}

	var stm structFieldMap

	var elemType reflect.Type
	if !p.nodeDestination {
		elemType = getElemTyperType(dest, t)

		dest, t = unravelDestination(dest, t)

		if elemType == nil && t != nil {
			switch t.Kind() {
			case reflect.Struct:
				var ok bool
				stm, ok = p.structTypeCache[t]
				if !ok {
					stm = getStructFieldInfoMap(t)
					p.structTypeCache[t] = stm
				}

			case reflect.Map:
				// For any key that we find in our loop here below, the new value fully
				// replaces any old value. So no need for us to dig down into a tree.
				// (This is because we are decoding into a map. If we were decoding into
				// a struct we would need to dig down into a tree, to match the behavior
				// of Golang's JSON decoder.)
				elemType = t.Elem()
			}
		}
	}

	for p.ch > 0 {
		var key string
		if key, err = p.readKeyname(); err != nil {
			return nil, err
		}
		ciKey := p.white()
		if p.ch != ':' {
			return nil, p.errAt("Expected ':' instead of '" + string(p.ch) + "'")
		}
		p.next()

		var newDest reflect.Value
		var newDestType reflect.Type
		if stm != nil {
			sfi, ok := stm.getField(key)
			if ok {
				// The field might be found on the root struct or in embedded structs.
				newDest, newDestType = dest, t
				for _, i := range sfi.indexPath {
					newDest, newDestType = unravelDestination(newDest, newDestType)

					if newDestType == nil {
						return nil, p.errAt("Internal error")
					}
					newDestType = newDestType.Field(i).Type
					elemType = newDestType

					if newDest.IsValid() {
						if newDest.Kind() != reflect.Struct {
							// We are only keeping track of newDest in case it contains a
							// tree that we will partially update. But here we have not found
							// any tree, so we can ignore newDest and just look at
							// newDestType instead.
							newDest = reflect.Value{}
						} else {
							newDest = newDest.Field(i)
						}
					}
				}
			}
		}

		// duplicate keys overwrite the previous value
		var val interface{}
		if val, err = p.readValue(newDest, elemType); err != nil {
			return nil, err
		}
		if p.nodeDestination {
			var ok bool
			if elemNode, ok = val.(*Node); ok {
				p.setComment1(&elemNode.Cm.Key, ciKey)
				elemNode.Cm.Key += elemNode.Cm.Before
				elemNode.Cm.Before = ""
				p.setComment1(&elemNode.Cm.Before, ciBefore)
			}
		}
		// Check white before comma because comma might be on other line.
		ciAfter := p.white()
		// in Hjson the comma is optional and trailing commas are allowed
		if p.ch == ',' {
			p.next()
			ciAfterComma := p.whiteAfterComma()
			if elemNode != nil {
				existingAfter := elemNode.Cm.After
				p.setComment2(&elemNode.Cm.After, ciAfter, ciAfterComma)
				elemNode.Cm.After = existingAfter + elemNode.Cm.After
			}
			ciAfter = p.white()
		}
		if p.ch == '}' && !withoutBraces {
			p.setComment1(&node.Cm.InsideLast, ciAfter)
			oldValue, isDuplicate := object.Set(key, val)
			if isDuplicate && p.DisallowDuplicateKeys {
				return nil, p.errAt(fmt.Sprintf("Found duplicate values ('%#v' and '%#v') for key '%v'",
					oldValue, val, key))
			}
			p.next()
			return p.maybeWrapNode(&node, object)
		}
		oldValue, isDuplicate := object.Set(key, val)
		if isDuplicate && p.DisallowDuplicateKeys {
			return nil, p.errAt(fmt.Sprintf("Found duplicate values ('%#v' and '%#v') for key '%v'",
				oldValue, val, key))
		}
		ciBefore = ciAfter
	}

	if withoutBraces {
		p.setComment1(&node.Cm.InsideLast, ciBefore)
		return p.maybeWrapNode(&node, object)
	}
	return nil, p.errAt("End of input while parsing an object (did you forget a closing '}'?)")
}

// dest and t must not have been unraveled yet here. In readTfnns we need
// to check if the original type (or a pointer to it) implements
// encoding.TextUnmarshaler.
func (p *hjsonParser) readValue(dest reflect.Value, t reflect.Type) (ret interface{}, err error) {
	ciBefore := p.white()
	// Parse an Hjson value. It could be an object, an array, a string, a number or a word.
	switch p.ch {
	case '{':
		p.nestingDepth++
		ret, err = p.readObject(false, dest, t, ciBefore)
		p.nestingDepth--
	case '[':
		p.nestingDepth++
		ret, err = p.readArray(dest, t)
		p.nestingDepth--
	case '"', '\'':
		s, err := p.readString(true)
		if err != nil {
			return nil, err
		}
		ret, err = p.maybeWrapNode(&Node{}, s)
	default:
		ret, err = p.readTfnns(dest, t)
		// Make sure that any comment will include preceding whitespace.
		if p.ch == '#' || p.ch == '/' {
			for p.prev() && p.ch <= ' ' {
			}
			p.next()
		}
	}

	ciAfter := p.getCommentAfter()
	if p.nodeDestination {
		if node, ok := ret.(*Node); ok {
			p.setComment1(&node.Cm.Before, ciBefore)
			p.setComment1(&node.Cm.After, ciAfter)
		}
	}

	return
}

func (p *hjsonParser) rootValue(dest reflect.Value) (ret interface{}, err error) {
	// Braces for the root object are optional

	// We have checked that dest is a pointer before calling rootValue().
	// Dereference here because readObject() etc will pass on child destinations
	// without creating pointers.
	dest = dest.Elem()
	t := dest.Type()

	var errSyntax error
	var ciAfter commentInfo
	ciBefore := p.white()

	switch p.ch {
	case '{':
		ret, err = p.readObject(false, dest, t, ciBefore)
		if err != nil {
			return
		}
		ciAfter, err = p.checkTrailing()
		if err != nil {
			return
		}
		if p.nodeDestination {
			if node, ok := ret.(*Node); ok {
				p.setComment1(&node.Cm.Before, ciBefore)
				p.setComment1(&node.Cm.After, ciAfter)
			}
		}
		return
	case '[':
		ret, err = p.readArray(dest, t)
		if err != nil {
			return
		}
		ciAfter, err = p.checkTrailing()
		if err != nil {
			return
		}
		if p.nodeDestination {
			if node, ok := ret.(*Node); ok {
				p.setComment1(&node.Cm.Before, ciBefore)
				p.setComment1(&node.Cm.After, ciAfter)
			}
		}
		return
	}

	if ret == nil {
		// Assume we have a root object without braces.
		ret, errSyntax = p.readObject(true, dest, t, ciBefore)
		ciAfter, err = p.checkTrailing()
		if errSyntax != nil || err != nil {
			// Syntax error, or maybe a single JSON value.
			ret = nil
			err = nil
		} else {
			if p.nodeDestination {
				if node, ok := ret.(*Node); ok {
					p.setComment1(&node.Cm.After, ciAfter)
				}
			}
			return
		}
	}

	if ret == nil {
		// test if we are dealing with a single JSON value instead (true/false/null/num/"")
		p.resetAt()
		ret, err = p.readValue(dest, t)
		if err == nil {
			ciAfter, err = p.checkTrailing()
		}
		if err == nil {
			if p.nodeDestination {
				if node, ok := ret.(*Node); ok {
					// ciBefore has been read again and set on the node inside the
					// function p.readValue().
					existingAfter := node.Cm.After
					p.setComment1(&node.Cm.After, ciAfter)
					if node.Cm.After != "" {
						existingAfter += "\n"
					}
					node.Cm.After = existingAfter + node.Cm.After
				}
			}

			return
		}
	}

	if errSyntax != nil {
		return nil, errSyntax
	}

	return
}

func (p *hjsonParser) checkTrailing() (commentInfo, error) {
	ci := p.white()
	if p.ch > 0 {
		return ci, p.errAt("Syntax error, found trailing characters")
	}
	return ci, nil
}

// Unmarshal parses the Hjson-encoded data using default options and stores the
// result in the value pointed to by v.
//
// See UnmarshalWithOptions.
func Unmarshal(data []byte, v interface{}) error {
	return UnmarshalWithOptions(data, v, DefaultDecoderOptions())
}

func orderedUnmarshal(
	data []byte,
	v interface{},
	options DecoderOptions,
	willMarshalToJSON bool,
	nodeDestination bool,
) (
	interface{},
	error,
) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("Cannot unmarshal into non-pointer %v", reflect.TypeOf(v))
	}

	parser := &hjsonParser{
		DecoderOptions:    options,
		data:              data,
		at:                0,
		ch:                ' ',
		structTypeCache:   map[reflect.Type]structFieldMap{},
		willMarshalToJSON: willMarshalToJSON,
		nodeDestination:   nodeDestination,
		nestingDepth:      0,
	}
	parser.resetAt()
	value, err := parser.rootValue(rv)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// UnmarshalWithOptions parses the Hjson-encoded data and stores the result
// in the value pointed to by v.
//
// The Hjson input is internally converted to JSON, which is then used as input
// to the function json.Unmarshal(). Unless the input argument v is of any of
// these types:
//
//	*hjson.OrderedMap
//	**hjson.OrderedMap
//	*hjson.Node
//	**hjson.Node
//
// Comments can be read from the Hjson-encoded data, but only if the input
// argument v is of type *hjson.Node or **hjson.Node.
//
// For more details about the output from this function, see the documentation
// for json.Unmarshal().
func UnmarshalWithOptions(data []byte, v interface{}, options DecoderOptions) error {
	inOM, destinationIsOrderedMap := v.(*OrderedMap)
	if !destinationIsOrderedMap {
		pInOM, ok := v.(**OrderedMap)
		if ok {
			destinationIsOrderedMap = true
			inOM = &OrderedMap{}
			*pInOM = inOM
		}
	}

	inNode, destinationIsNode := v.(*Node)
	if !destinationIsNode {
		pInNode, ok := v.(**Node)
		if ok {
			destinationIsNode = true
			inNode = &Node{}
			*pInNode = inNode
		}
	}

	value, err := orderedUnmarshal(data, v, options, !(destinationIsOrderedMap ||
		destinationIsNode), destinationIsNode)
	if err != nil {
		return err
	}

	if destinationIsOrderedMap {
		if outOM, ok := value.(*OrderedMap); ok {
			*inOM = *outOM
			return nil
		}
		return fmt.Errorf("Cannot unmarshal into hjson.OrderedMap: Try %v as destination instead",
			reflect.TypeOf(v))
	}

	if destinationIsNode {
		if outNode, ok := value.(*Node); ok {
			*inNode = *outNode
			return nil
		}
	}

	// Convert to JSON so we can let json.Unmarshal() handle all destination
	// types (including interfaces json.Unmarshaler and encoding.TextUnmarshaler)
	// and merging.
	buf, err := json.Marshal(value)
	if err != nil {
		return errors.New("Internal error")
	}

	dec := json.NewDecoder(bytes.NewBuffer(buf))
	if options.UseJSONNumber {
		dec.UseNumber()
	}
	if options.DisallowUnknownFields {
		dec.DisallowUnknownFields()
	}

	err = dec.Decode(v)
	if err != nil {
		return err
	}

	return err
}
