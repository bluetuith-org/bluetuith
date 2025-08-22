package hjson

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type fieldInfo struct {
	field   reflect.Value
	name    string
	comment string
}

type structFieldInfo struct {
	name      string
	tagged    bool
	comment   string
	omitEmpty bool
	indexPath []int
}

// Use lower key name as key. Values are arrays in case some fields only differ
// by upper/lower case.
type structFieldMap map[string][]structFieldInfo

func (s structFieldMap) insert(sfi structFieldInfo) {
	key := strings.ToLower(sfi.name)
	s[key] = append(s[key], sfi)
}

func (s structFieldMap) getField(name string) (structFieldInfo, bool) {
	key := strings.ToLower(name)
	if arr, ok := s[key]; ok {
		for _, elem := range arr {
			if elem.name == name {
				return elem, true
			}
		}
		return arr[0], true
	}

	return structFieldInfo{}, false
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// JSON tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []structFieldInfo) (structFieldInfo, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level, either both tagged or neither tagged.
	if len(fields) > 1 && len(fields[0].indexPath) == len(fields[1].indexPath) && fields[0].tagged == fields[1].tagged {
		return structFieldInfo{}, false
	}
	return fields[0], true
}

// byIndex sorts by index sequence.
type byIndex []structFieldInfo

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].indexPath {
		if k >= len(x[j].indexPath) {
			return false
		}
		if xik != x[j].indexPath[k] {
			return xik < x[j].indexPath[k]
		}
	}
	return len(x[i].indexPath) < len(x[j].indexPath)
}

func getStructFieldInfo(rootType reflect.Type) []structFieldInfo {
	type structInfo struct {
		typ       reflect.Type
		indexPath []int
	}
	var sfis []structFieldInfo
	structsToInvestigate := []structInfo{structInfo{typ: rootType}}
	// Struct types already visited at an earlier depth.
	visited := map[reflect.Type]bool{}
	// Count the number of specific struct types on a specific depth.
	typeDepthCount := map[reflect.Type]int{}

	for len(structsToInvestigate) > 0 {
		curStructs := structsToInvestigate
		structsToInvestigate = []structInfo{}
		curTDC := typeDepthCount
		typeDepthCount = map[reflect.Type]int{}

		for _, curStruct := range curStructs {
			if visited[curStruct.typ] {
				// The struct type has already appeared on an earlier depth. Fields on
				// an earlier depth always have precedence over fields with identical
				// name on a later depth, so no point in investigating this type again.
				continue
			}
			visited[curStruct.typ] = true

			for i := 0; i < curStruct.typ.NumField(); i++ {
				sf := curStruct.typ.Field(i)

				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					// If the field is not exported and not a struct.
					if sf.PkgPath != "" && t.Kind() != reflect.Struct {
						// Ignore embedded fields of unexported non-struct types.
						continue
					}
					// Do not ignore embedded fields of unexported struct types
					// since they may have exported fields.
				} else if sf.PkgPath != "" {
					// Ignore unexported non-embedded fields.
					continue
				}

				jsonTag := sf.Tag.Get("json")
				if jsonTag == "-" {
					continue
				}

				sfi := structFieldInfo{
					name:    sf.Name,
					comment: sf.Tag.Get("comment"),
				}

				splits := strings.Split(jsonTag, ",")
				if splits[0] != "" {
					sfi.name = splits[0]
					sfi.tagged = true
				}
				if len(splits) > 1 {
					for _, opt := range splits[1:] {
						if opt == "omitempty" {
							sfi.omitEmpty = true
						}
					}
				}

				sfi.indexPath = make([]int, len(curStruct.indexPath)+1)
				copy(sfi.indexPath, curStruct.indexPath)
				sfi.indexPath[len(curStruct.indexPath)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}

				// If the current field should be included.
				if sfi.tagged || !sf.Anonymous || ft.Kind() != reflect.Struct {
					sfis = append(sfis, sfi)
					if curTDC[curStruct.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						sfis = append(sfis, sfi)
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				typeDepthCount[ft]++
				if typeDepthCount[ft] == 1 {
					structsToInvestigate = append(structsToInvestigate, structInfo{
						typ:       ft,
						indexPath: sfi.indexPath,
					})
				}
			}
		}
	}

	sort.Slice(sfis, func(i, j int) bool {
		// sort field by name, breaking ties with depth, then
		// breaking ties with "name came from json tag", then
		// breaking ties with index sequence.
		if sfis[i].name != sfis[j].name {
			return sfis[i].name < sfis[j].name
		}
		if len(sfis[i].indexPath) != len(sfis[j].indexPath) {
			return len(sfis[i].indexPath) < len(sfis[j].indexPath)
		}
		if sfis[i].tagged != sfis[j].tagged {
			return sfis[i].tagged
		}
		return byIndex(sfis).Less(i, j)
	})

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with JSON tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := sfis[:0]
	for advance, i := 0, 0; i < len(sfis); i += advance {
		// One iteration per name.
		// Find the sequence of sfis with the name of this first field.
		sfi := sfis[i]
		name := sfi.name
		for advance = 1; i+advance < len(sfis); advance++ {
			fj := sfis[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, sfi)
			continue
		}
		dominant, ok := dominantField(sfis[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	return out
}

func getStructFieldInfoSlice(rootType reflect.Type) []structFieldInfo {
	sfis := getStructFieldInfo(rootType)

	sort.Sort(byIndex(sfis))

	return sfis
}

func getStructFieldInfoMap(rootType reflect.Type) structFieldMap {
	sfis := getStructFieldInfo(rootType)

	out := structFieldMap{}
	for _, elem := range sfis {
		out.insert(elem)
	}

	return out
}

func (e *hjsonEncoder) writeFields(
	fis []fieldInfo,
	noIndent bool,
	separator string,
	isRootObject bool,
	isObjElement bool,
	cm Comments,
) error {
	indent1 := e.indent
	if !isRootObject || e.EmitRootBraces || len(fis) == 0 {
		e.bracesIndent(isObjElement, len(fis) == 0, cm, separator)
		e.WriteString("{" + cm.InsideFirst)

		if len(fis) == 0 {
			if cm.InsideFirst != "" || cm.InsideLast != "" {
				e.WriteString(e.Eol)
			}
			e.WriteString(cm.InsideLast)
			if cm.InsideLast != "" {
				endsInsideComment, endsWithLineFeed := investigateComment(cm.InsideLast)
				if endsInsideComment {
					e.writeIndent(e.indent)
				}
				if endsWithLineFeed {
					e.writeIndentNoEOL(e.indent)
				}
			} else if cm.InsideFirst != "" {
				e.writeIndentNoEOL(e.indent)
			}
			e.WriteString("}")
			return nil
		}

		e.indent++
	} else {
		e.WriteString(cm.InsideFirst)
	}

	// Join all of the member texts together, separated with newlines
	var elemCm Comments
	for i, fi := range fis {
		var elem reflect.Value
		elem, elemCm = e.unpackNode(fi.field, elemCm)
		if i > 0 || !isRootObject || e.EmitRootBraces {
			e.WriteString(e.Eol)
		}
		if len(fi.comment) > 0 {
			for _, line := range strings.Split(fi.comment, e.Eol) {
				e.writeIndentNoEOL(e.indent)
				e.WriteString(fmt.Sprintf("# %s\n", line))
			}
		}
		if elemCm.Before == "" {
			e.writeIndentNoEOL(e.indent)
		} else {
			e.WriteString(elemCm.Before)
		}
		e.WriteString(e.quoteName(fi.name))
		e.WriteString(":")
		e.WriteString(elemCm.Key)

		if err := e.str(elem, false, " ", false, true, elemCm); err != nil {
			return err
		}

		if len(fi.comment) > 0 && i < len(fis)-1 {
			e.WriteString(e.Eol)
		}

		e.WriteString(elemCm.After)
	}

	if cm.InsideLast != "" {
		e.WriteString(e.Eol + cm.InsideLast)
	}

	if !isRootObject || e.EmitRootBraces {
		if cm.InsideLast == "" {
			e.writeIndent(indent1)
		} else {
			endsInsideComment, endsWithLineFeed := investigateComment(cm.InsideLast)
			if endsInsideComment {
				e.writeIndent(indent1)
			} else if endsWithLineFeed {
				e.writeIndentNoEOL(indent1)
			}
		}
		e.WriteString("}")
	}

	e.indent = indent1

	return nil
}
