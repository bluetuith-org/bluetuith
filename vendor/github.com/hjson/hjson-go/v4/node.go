package hjson

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Comments struct {
	// Comment/whitespace on line(s) before the value, and before the value on
	// the same line. If not empty, is expected to end with a line feed +
	// indentation for the value.
	Before string
	// Comment/whitespace between the key and this value, if this value is an
	// element in a map/object.
	Key string
	// Comment/whitespace after (but still on the same line as) the leading
	// bracket ({ or [) for this value, if this value is a slice/array or
	// map/object. Is not expected to contain any line feed.
	InsideFirst string
	// Comment/whitespace from the beginning of the first line after all child
	// values belonging to this value, until the closing bracket (} or ]), if
	// this value is a slice/array or map/object. If not empty, is expected to
	// end with a line feed + indentation for the closing bracket.
	InsideLast string
	// Comment/whitespace after (but still on the same line as) the value. Is not
	// expected to contain any line feed. hjson.Unmarshal() will try to assign
	// comments/whitespace from lines between values to `Before` on the value
	// after those lines, or to `InsideLast` on the slice/map if the lines
	// containing comments appear after the last element inside a slice/map.
	After string
}

// Node must be used as destination for Unmarshal() or UnmarshalWithOptions()
// whenever comments should be read from the input. The struct is simply a
// wrapper for the actual values and a helper struct containing any comments.
// The Value in the destination Node will be overwritten in the call to
// Unmarshal() or UnmarshalWithOptions(), i.e. node trees are not merged.
// After the unmarshal, Node.Value will contain any of these types:
//
//	nil (no type)
//	float64 (if UseJSONNumber == false)
//	json.Number (if UseJSONNumber == true)
//	string
//	bool
//	[]interface{}
//	*hjson.OrderedMap
//
// All elements in an []interface{} or *hjson.OrderedMap will be of the type
// *hjson.Node, so that they can contain comments.
//
// This example shows unmarshalling input with comments, changing the value on
// a single key (the input is assumed to have an object/map as root) and then
// marshalling the node tree again, including comments and with preserved key
// order in the object/map.
//
//	var node hjson.Node
//	err := hjson.Unmarshal(input, &node)
//	if err != nil {
//	  return err
//	}
//	_, err = node.SetKey("setting1", 3)
//	if err != nil {
//	  return err
//	}
//	output, err := hjson.Marshal(node)
//	if err != nil {
//	  return err
//	}
type Node struct {
	Value interface{}
	Cm    Comments
}

// Len returns the length of the value wrapped by this Node, if the value is of
// type *hjson.OrderedMap, []interface{} or string. Otherwise 0 is returned.
func (c *Node) Len() int {
	if c == nil {
		return 0
	}
	switch cont := c.Value.(type) {
	case *OrderedMap:
		return cont.Len()
	case []interface{}:
		return len(cont)
	case string:
		return len(cont)
	}
	return 0
}

// AtIndex returns the key (if any) and value (unwrapped from its Node) found
// at the specified index, if this Node contains a value of type
// *hjson.OrderedMap or []interface{}. Returns an error for unexpected types.
// Panics if index < 0 or index >= Len().
func (c *Node) AtIndex(index int) (string, interface{}, error) {
	if c == nil {
		return "", nil, fmt.Errorf("Node is nil")
	}
	var key string
	var elem interface{}
	switch cont := c.Value.(type) {
	case *OrderedMap:
		key = cont.Keys[index]
		elem = cont.Map[key]
	case []interface{}:
		elem = cont[index]
	default:
		return "", nil, fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
	}
	node, ok := elem.(*Node)
	if !ok {
		return "", nil, fmt.Errorf("Unexpected element type: %v", reflect.TypeOf(elem))
	}
	return key, node.Value, nil
}

// AtKey returns the value (unwrapped from its Node) found for the specified
// key, if this Node contains a value of type *hjson.OrderedMap. An error is
// returned for unexpected types. The second returned value is true if the key
// was found, false otherwise.
func (c *Node) AtKey(key string) (interface{}, bool, error) {
	if c == nil {
		return nil, false, nil
	}
	om, ok := c.Value.(*OrderedMap)
	if !ok {
		return nil, false, fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
	}
	elem, ok := om.Map[key]
	if !ok {
		return nil, false, nil
	}
	node, ok := elem.(*Node)
	if !ok {
		return nil, false, fmt.Errorf("Unexpected element type: %v", reflect.TypeOf(elem))
	}
	return node.Value, true, nil
}

// Append adds the input value to the end of the []interface{} wrapped by this
// Node. If this Node contains nil without a type, an empty []interface{} is
// first created. If this Node contains a value of any other type, an error is
// returned.
func (c *Node) Append(value interface{}) error {
	if c == nil {
		return fmt.Errorf("Node is nil")
	}
	var arr []interface{}
	if c.Value == nil {
		arr = []interface{}{}
	} else {
		var ok bool
		arr, ok = c.Value.([]interface{})
		if !ok {
			return fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
		}
	}
	c.Value = append(arr, &Node{Value: value})
	return nil
}

// Insert inserts a new key/value pair at the specified index, if this Node
// contains a value of type *hjson.OrderedMap or []interface{}.  Returns an error
// for unexpected types. Panics if index < 0 or index > c.Len(). If the key
// already exists in the OrderedMap, the new value is set but the position of
// the key is not changed. Otherwise the value to insert is wrapped in a new
// Node. If this Node contains []interface{}, the key is ignored. Returns the
// old value and true if the key already exists in the/ OrderedMap, nil and
// false otherwise.
func (c *Node) Insert(index int, key string, value interface{}) (interface{}, bool, error) {
	if c == nil {
		return nil, false, fmt.Errorf("Node is nil")
	}
	var oldVal interface{}
	var found bool
	switch cont := c.Value.(type) {
	case *OrderedMap:
		oldVal, found := cont.Map[key]
		if found {
			if node, ok := oldVal.(*Node); ok {
				oldVal = node.Value
				node.Value = value
			} else {
				cont.Map[key] = &Node{Value: value}
			}
		} else {
			oldVal, found = cont.Insert(index, key, &Node{Value: value})
		}
	case []interface{}:
		value = &Node{Value: value}
		if index == len(cont) {
			c.Value = append(cont, value)
		} else {
			cont = append(cont[:index+1], cont[index:]...)
			cont[index] = value
			c.Value = cont
		}
	default:
		return nil, false, fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
	}
	return oldVal, found, nil
}

// SetIndex assigns the specified value to the child Node found at the specified
// index, if this Node contains a value of type *hjson.OrderedMap or
// []interface{}. Returns an error for unexpected types. Returns the key (if
// any) and value previously found at the specified index. Panics if index < 0
// or index >= Len().
func (c *Node) SetIndex(index int, value interface{}) (string, interface{}, error) {
	if c == nil {
		return "", nil, fmt.Errorf("Node is nil")
	}
	var key string
	var elem interface{}
	switch cont := c.Value.(type) {
	case *OrderedMap:
		key = cont.Keys[index]
		elem = cont.Map[key]
	case []interface{}:
		elem = cont[index]
	default:
		return "", nil, fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
	}
	var oldVal interface{}
	node, ok := elem.(*Node)
	if ok {
		oldVal = node.Value
		node.Value = value
	} else {
		oldVal = elem
		switch cont := c.Value.(type) {
		case *OrderedMap:
			cont.Map[key] = &Node{Value: value}
		case []interface{}:
			cont[index] = &Node{Value: value}
		}
	}
	return key, oldVal, nil
}

// SetKey assigns the specified value to the child Node identified by the
// specified key, if this Node contains a value of the type *hjson.OrderedMap.
// If this Node contains nil without a type, an empty *hjson.OrderedMap is
// first created. If this Node contains a value of any other type an error is
// returned. If the key cannot be found in the OrderedMap, a new Node is
// created, wrapping the specified value, and appended to the end of the
// OrderedMap. Returns the old value and true if the key already existed in
// the OrderedMap, nil and false otherwise.
func (c *Node) SetKey(key string, value interface{}) (interface{}, bool, error) {
	if c == nil {
		return nil, false, fmt.Errorf("Node is nil")
	}
	var om *OrderedMap
	if c.Value == nil {
		om = NewOrderedMap()
		c.Value = om
	} else {
		var ok bool
		om, ok = c.Value.(*OrderedMap)
		if !ok {
			return nil, false, fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
		}
	}
	var oldVal interface{}
	elem, ok := om.Map[key]
	if ok {
		var node *Node
		node, ok = elem.(*Node)
		if ok {
			oldVal = node.Value
			node.Value = value
		}
	}
	foundKey := true
	if !ok {
		oldVal, foundKey = om.Set(key, &Node{Value: value})
	}
	return oldVal, foundKey, nil
}

// DeleteIndex deletes the value or key/value pair found at the specified index,
// if this Node contains a value of type *hjson.OrderedMap or []interface{}.
// Returns an error for unexpected types. Panics if index < 0 or
// index >= c.Len(). Returns the deleted key (if any) and value.
func (c *Node) DeleteIndex(index int) (string, interface{}, error) {
	if c == nil {
		return "", nil, fmt.Errorf("Node is nil")
	}
	var key string
	var value interface{}
	switch cont := c.Value.(type) {
	case *OrderedMap:
		key, value = cont.DeleteIndex(index)
	case []interface{}:
		value = cont[index]
		cont = append(cont[:index], cont[index+1:]...)
		c.Value = cont
	default:
		return "", nil, fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
	}
	if node, ok := value.(*Node); ok {
		value = node.Value
	}
	return key, value, nil
}

// DeleteKey deletes the key/value pair with the specified key, if found and if
// this Node contains a value of type *hjson.OrderedMap. Returns an error for
// unexpected types. Returns the deleted value and true if the key was found,
// nil and false otherwise.
func (c *Node) DeleteKey(key string) (interface{}, bool, error) {
	if c == nil {
		return nil, false, fmt.Errorf("Node is nil")
	}
	if om, ok := c.Value.(*OrderedMap); ok {
		oldValue, found := om.DeleteKey(key)
		if node, ok := oldValue.(*Node); ok {
			oldValue = node.Value
		}
		return oldValue, found, nil
	}
	return nil, false, fmt.Errorf("Unexpected value type: %v", reflect.TypeOf(c.Value))
}

// NI is an acronym formed from "get Node pointer by Index". Returns the *Node
// element found at the specified index, if this Node contains a value of type
// *hjson.OrderedMap or []interface{}. Returns nil otherwise. Panics if
// index < 0 or index >= Len(). Does not create or alter any value.
func (c *Node) NI(index int) *Node {
	if c == nil {
		return nil
	}
	var elem interface{}
	switch cont := c.Value.(type) {
	case *OrderedMap:
		elem = cont.AtIndex(index)
	case []interface{}:
		elem = cont[index]
	default:
		return nil
	}
	if node, ok := elem.(*Node); ok {
		return node
	}
	return nil
}

// NK is an acronym formed from "get Node pointer by Key". Returns the *Node
// element found for the specified key, if this Node contains a value of type
// *hjson.OrderedMap. Returns nil otherwise. Does not create or alter anything.
func (c *Node) NK(key string) *Node {
	if c == nil {
		return nil
	}
	om, ok := c.Value.(*OrderedMap)
	if !ok {
		return nil
	}
	if elem, ok := om.Map[key]; ok {
		if node, ok := elem.(*Node); ok {
			return node
		}
	}
	return nil
}

// NKC is an acronym formed from "get Node pointer by Key, Create if not found".
// Returns the *Node element found for the specified key, if this Node contains
// a value of type *hjson.OrderedMap. If this Node contains nil without a type,
// an empty *hjson.OrderedMap is first created. If this Node contains a value of
// any other type or if the element idendified by the specified key is not of
// type *Node, an error is returned. If the key cannot be found in the
// OrderedMap, a new Node is created for the specified key. Example usage:
//
//	var node hjson.Node
//	node.NKC("rootKey1").NKC("subKey1").SetKey("valKey1", "my value")
func (c *Node) NKC(key string) *Node {
	if c == nil {
		return nil
	}
	var om *OrderedMap
	if c.Value == nil {
		om = NewOrderedMap()
		c.Value = om
	} else {
		var ok bool
		om, ok = c.Value.(*OrderedMap)
		if !ok {
			return nil
		}
	}
	if elem, ok := om.Map[key]; ok {
		if node, ok := elem.(*Node); ok {
			return node
		}
	} else {
		node := &Node{}
		om.Set(key, node)
		return node
	}
	return nil
}

// MarshalJSON is an implementation of the json.Marshaler interface, enabling
// hjson.Node trees to be used as input for json.Marshal().
func (c Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

// UnmarshalJSON is an implementation of the json.Unmarshaler interface,
// enabling hjson.Node to be used as destination for json.Unmarshal().
func (c *Node) UnmarshalJSON(b []byte) error {
	return Unmarshal(b, c)
}
