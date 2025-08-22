package hjson

import (
	"bytes"
	"encoding/json"
)

// OrderedMap wraps a map and a slice containing all of the keys from the map,
// so that the order of the keys can be specified. The Keys slice can be sorted
// or rearranged like any other slice, but do not add or remove keys manually
// on it. Use OrderedMap.Insert(), OrderedMap.Set(), OrderedMap.DeleteIndex() or
// OrderedMap.DeleteKey() instead.
//
// Example of how to iterate through the elements of an OrderedMap in order:
//
//	for _, key := range om.Keys {
//	  fmt.Printf("%v\n", om.Map[key])
//	}
//
// Always use the functions Insert() or Set() instead of setting values
// directly on OrderedMap.Map, because any new keys must also be added to
// OrderedMap.Keys. Otherwise those keys will be ignored when iterating through
// the elements of the OrderedMap in order, as for example happens in the
// function hjson.Marshal().
type OrderedMap struct {
	Keys []string
	Map  map[string]interface{}
}

// KeyValue is only used as input to NewOrderedMapFromSlice().
type KeyValue struct {
	Key   string
	Value interface{}
}

// NewOrderedMap returns a pointer to a new OrderedMap. An OrderedMap should
// always be passed by reference, never by value. If an OrderedMap is passed
// by value then appending new keys won't affect all of the copies of the
// OrderedMap.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys: nil,
		Map:  map[string]interface{}{},
	}
}

// NewOrderedMapFromSlice is like NewOrderedMap but with initial values.
// Example:
//
//	om := NewOrderedMapFromSlice([]KeyValue{
//	 {"B", "first"},
//	 {"A", "second"},
//	})
func NewOrderedMapFromSlice(args []KeyValue) *OrderedMap {
	c := NewOrderedMap()
	for _, elem := range args {
		c.Set(elem.Key, elem.Value)
	}
	return c
}

// Len returns the number of values contained in the OrderedMap.
func (c *OrderedMap) Len() int {
	return len(c.Keys)
}

// AtIndex returns the value found at the specified index. Panics if
// index < 0 or index >= c.Len().
func (c *OrderedMap) AtIndex(index int) interface{} {
	return c.Map[c.Keys[index]]
}

// AtKey returns the value found for the specified key, and true if the value
// was found. Returns nil and false if the value was not found.
func (c *OrderedMap) AtKey(key string) (interface{}, bool) {
	ret, ok := c.Map[key]
	return ret, ok
}

// Insert inserts a new key/value pair at the specified index. Panics if
// index < 0 or index > c.Len(). If the key already exists in the OrderedMap,
// the new value is set but the position of the key is not changed. Returns
// the old value and true if the key already exists in the OrderedMap, nil and
// false otherwise.
func (c *OrderedMap) Insert(index int, key string, value interface{}) (interface{}, bool) {
	oldValue, exists := c.Map[key]
	c.Map[key] = value
	if exists {
		return oldValue, true
	}
	if index == len(c.Keys) {
		c.Keys = append(c.Keys, key)
	} else {
		c.Keys = append(c.Keys[:index+1], c.Keys[index:]...)
		c.Keys[index] = key
	}
	return nil, false
}

// Set sets the specified value for the specified key. If the key does not
// already exist in the OrderedMap it is appended to the end of the OrderedMap.
// If the key already exists in the OrderedMap, the new value is set but the
// position of the key is not changed. Returns the old value and true if the
// key already exists in the OrderedMap, nil and false otherwise.
func (c *OrderedMap) Set(key string, value interface{}) (interface{}, bool) {
	return c.Insert(len(c.Keys), key, value)
}

// DeleteIndex deletes the key/value pair found at the specified index.
// Returns the deleted key and value. Panics if index < 0 or index >= c.Len().
func (c *OrderedMap) DeleteIndex(index int) (string, interface{}) {
	key := c.Keys[index]
	value := c.Map[key]
	delete(c.Map, key)
	c.Keys = append(c.Keys[:index], c.Keys[index+1:]...)
	return key, value
}

// DeleteKey deletes the key/value pair with the specified key, if found.
// Returns the deleted value and true if the key was found, nil and false
// otherwise.
func (c *OrderedMap) DeleteKey(key string) (interface{}, bool) {
	for index, ck := range c.Keys {
		if ck == key {
			_, value := c.DeleteIndex(index)
			return value, true
		}
	}
	return nil, false
}

// MarshalJSON is an implementation of the json.Marshaler interface, enabling
// hjson.OrderedMap to be used as input for json.Marshal().
func (c *OrderedMap) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString("{")

	for index, key := range c.Keys {
		if index > 0 {
			b.WriteString(",")
		}
		jbuf, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		b.Write(jbuf)
		b.WriteString(":")
		jbuf, err = json.Marshal(c.Map[key])
		if err != nil {
			return nil, err
		}
		b.Write(jbuf)
	}

	b.WriteString("}")

	return b.Bytes(), nil
}

// UnmarshalJSON is an implementation of the json.Unmarshaler interface,
// enabling hjson.OrderedMap to be used as destination for json.Unmarshal().
func (c *OrderedMap) UnmarshalJSON(b []byte) error {
	c.Keys = nil
	c.Map = map[string]interface{}{}
	return Unmarshal(b, c)
}
