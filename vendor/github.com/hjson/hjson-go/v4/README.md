# hjson-go

[![Build Status](https://github.com/hjson/hjson-go/workflows/test/badge.svg)](https://github.com/hjson/hjson-go/actions)
[![Go Pkg](https://img.shields.io/github/release/hjson/hjson-go.svg?style=flat-square&label=go-pkg)](https://github.com/hjson/hjson-go/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/hjson/hjson-go?style=flat-square)](https://goreportcard.com/report/github.com/hjson/hjson-go)
[![coverage](https://img.shields.io/badge/coverage-ok-brightgreen.svg?style=flat-square)](https://gocover.io/github.com/hjson/hjson-go/)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/hjson/hjson-go/v4)

![Hjson Intro](https://hjson.github.io/hjson1.gif)

```
{
  # specify rate in requests/second (because comments are helpful!)
  rate: 1000

  // prefer c-style comments?
  /* feeling old fashioned? */

  # did you notice that rate doesn't need quotes?
  hey: look ma, no quotes for strings either!

  # best of all
  notice: []
  anything: ?

  # yes, commas are optional!
}
```

The Go implementation of Hjson is based on [hjson-js](https://github.com/hjson/hjson-js). For other platforms see [hjson.github.io](https://hjson.github.io).

More documentation can be found at https://pkg.go.dev/github.com/hjson/hjson-go/v4

# Install

Instructions for installing a pre-built **hjson-cli** tool can be found at https://hjson.github.io/users-bin.html

If you instead want to build locally, make sure you have a working Go environment. See the [install instructions](https://golang.org/doc/install.html).

- In order to use Hjson from your own Go source code, just add an import line like the one here below. Before building your project, run `go mod tidy` in order to download the Hjson source files. The suffix `/v4` is required in the import path, unless you specifically want to use an older major version.
```go
import "github.com/hjson/hjson-go/v4"
```
- If you instead want to use the **hjson-cli** command line tool, run the command here below in your terminal. The executable will be installed into your `go/bin` folder, make sure that folder is included in your `PATH` environment variable.
```bash
go install github.com/hjson/hjson-go/v4/hjson-cli@latest
```
# Usage as command line tool
```
usage: hjson-cli [OPTIONS] [INPUT]
hjson can be used to convert JSON from/to Hjson.

hjson will read the given JSON/Hjson input file or read from stdin.

Options:
  -bracesSameLine
      Print braces on the same line.
  -c  Output as JSON.
  -h  Show this screen.
  -indentBy string
      The indent string. (default "  ")
  -j  Output as formatted JSON.
  -omitRootBraces
      Omit braces at the root.
  -preserveKeyOrder
      Preserve key order in objects/maps.
  -quoteAlways
      Always quote string values.
  -v
      Show version.
```

Sample:
- run `hjson-cli test.json > test.hjson` to convert to Hjson
- run `hjson-cli -j test.hjson > test.json` to convert to JSON

# Usage as a GO library

```go

package main

import (
    "github.com/hjson/hjson-go/v4"
    "fmt"
)

func main() {
    // Now let's look at decoding Hjson data into Go
    // values.
    sampleText := []byte(`
    {
        # specify rate in requests/second
        rate: 1000
        array:
        [
            foo
            bar
        ]
    }`)

    // We need to provide a variable where Hjson
    // can put the decoded data.
    var dat map[string]interface{}

    // Decode with default options and check for errors.
    if err := hjson.Unmarshal(sampleText, &dat); err != nil {
        panic(err)
    }
    // short for:
    // options := hjson.DefaultDecoderOptions()
    // err := hjson.UnmarshalWithOptions(sampleText, &dat, options)
    fmt.Println(dat)

    // In order to use the values in the decoded map,
    // we'll need to cast them to their appropriate type.

    rate := dat["rate"].(float64)
    fmt.Println(rate)

    array := dat["array"].([]interface{})
    str1 := array[0].(string)
    fmt.Println(str1)


    // To encode to Hjson with default options:
    sampleMap := map[string]int{"apple": 5, "lettuce": 7}
    hjson, _ := hjson.Marshal(sampleMap)
    // short for:
    // options := hjson.DefaultOptions()
    // hjson, _ := hjson.MarshalWithOptions(sampleMap, options)
    fmt.Println(string(hjson))
}
```

## Unmarshal to Go structs

If you prefer, you can also unmarshal to Go structs (including structs implementing the json.Unmarshaler interface or the encoding.TextUnmarshaler interface). The Go JSON package is used for this, so the same rules apply. Specifically for the "json" key in struct field tags. For more details about this type of unmarshalling, see the [documentation for json.Unmarshal()](https://pkg.go.dev/encoding/json#Unmarshal).

```go

package main

import (
    "github.com/hjson/hjson-go/v4"
    "fmt"
)

type Sample struct {
    Rate  int
    Array []string
}

type SampleAlias struct {
    Rett    int      `json:"rate"`
    Ashtray []string `json:"array"`
}

func main() {
    sampleText := []byte(`
    {
        # specify rate in requests/second
        rate: 1000
        array:
        [
            foo
            bar
        ]
    }`)

    // unmarshal
    var sample Sample
    hjson.Unmarshal(sampleText, &sample)

    fmt.Println(sample.Rate)
    fmt.Println(sample.Array)

    // unmarshal using json tags on struct fields
    var sampleAlias SampleAlias
    hjson.Unmarshal(sampleText, &sampleAlias)

    fmt.Println(sampleAlias.Rett)
    fmt.Println(sampleAlias.Ashtray)
}
```

## Comments on struct fields

By using key `comment` in struct field tags you can specify comments to be written on one or more lines preceding the struct field in the Hjson output. Another way to output comments is to use *hjson.Node* structs, more on than later.

```go

package main

import (
    "github.com/hjson/hjson-go/v4"
    "fmt"
)

type foo struct {
    A string `json:"x" comment:"First comment"`
    B int32  `comment:"Second comment\nLook ma, new lines"`
    C string
    D int32
}

func main() {
    a := foo{A: "hi!", B: 3, C: "some text", D: 5}
    buf, err := hjson.Marshal(a)
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(string(buf))
}
```

Output:

```
{
  # First comment
  x: hi!

  # Second comment
  # Look ma, new lines
  B: 3

  C: some text
  D: 5
}
```

## Read and write comments

The only way to read comments from Hjson input is to use a destination variable of type *hjson.Node* or *&ast;hjson.Node*. The *hjson.Node* must be the root destination, it won't work if you create a field of type *hjson.Node* in some other struct and use that struct as destination. An *hjson.Node* struct is simply a wrapper for a value and comments stored in an *hjson.Comments* struct. It also has several convenience functions, for example *AtIndex()* or *SetKey()* that can be used when you know that the node contains a value of type `[]interface{}` or *&ast;hjson.OrderedMap*. All of the elements in `[]interface{}` or *&ast;hjson.OrderedMap* will be of type *&ast;hjson.Node* in trees created by *hjson.Unmarshal*, but the *hjson.Node* convenience functions unpack the actual values from them.

When *hjson.Node* or *&ast;hjson.Node* is used as destination for Hjson unmarshal the output will be a tree of *&ast;hjson.Node* where all of the values contained in tree nodes will be of these types:

*	`nil` (no type)
*	`float64` &nbsp;&nbsp;(if *UseJSONNumber* == `false`)
*	*json.Number* &nbsp;&nbsp;(if *UseJSONNumber* == `true`)
*	`string`
*	`bool`
*	`[]interface{}`
*	*&ast;hjson.OrderedMap*

These are just the types used by Hjson unmarshal and the convenience functions, you are free to assign any type of values to nodes in your own code.

The comments will contain all whitespace chars too (including line feeds) so that an Hjson document can be read and written without altering the layout. This can be disabled by setting the decoding option *WhitespaceAsComments* to `false`.

```go

package main

import (
    "fmt"

    "github.com/hjson/hjson-go/v4"
)

func main() {
    // Now let's look at decoding Hjson data into hjson.Node.
    sampleText := []byte(`
    {
        # specify rate in requests/second
        rate: 1000
        array:
        [
            foo
            bar
        ]
    }`)

    var node hjson.Node
    if err := hjson.Unmarshal(sampleText, &node); err != nil {
        panic(err)
    }

    node.NK("array").Cm.Before = `        # please specify an array
        `

    if _, _, err := node.NKC("subMap").SetKey("subVal", 1); err != nil {
        panic(err)
    }

    outBytes, err := hjson.Marshal(node)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(outBytes))
}
```

Output:

```

    {
        # specify rate in requests/second
        rate: 1000
        # please specify an array
        array:
        [
            foo
            bar
        ]
  subMap: {
    subVal: 1
  }
    }
```


## Type ambiguity

Hjson allows quoteless strings. But if a value is a valid number, boolean or `null` then it will be unmarshalled into that type instead of a string when unmarshalling into `interface{}`. This can lead to unintended consequences if the creator of an Hjson file meant to write a string but didn't think of that the quoteless string they wrote also was a valid number.

The ambiguity can be avoided by using typed destinations when unmarshalling. A string destination will receive a string even if the quoteless string also was a valid number, boolean or `null`. Example:

```go

package main

import (
    "github.com/hjson/hjson-go/v4"
    "fmt"
)

type foo struct {
  A string
}

func main() {
    var dest foo
    err := hjson.Unmarshal([]byte(`a: 3`), &dest)
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(dest)
}
```

Output:

```
{3}
```

String pointer destinations can be set to `nil` by writing `null` in an Hjson file. The same goes for a pointer destination of any type that implements `UnmarshalText()`.

## ElemTyper interface

If a destination type implements hjson.ElemTyper, Unmarshal() will call ElemType() on the destination when unmarshalling an array or an object, to see if any array element or leaf node should be of type string even if it can be treated as a number, boolean or null. This is most useful if the destination also implements the json.Unmarshaler interface, because then there is no other way for Unmarshal() to know the type of the elements on the destination. If a destination implements ElemTyper all of its elements must be of the same type.

Example implementation for a generic ordered map:

```go

func (o *OrderedMap[T]) ElemType() reflect.Type {
  return reflect.TypeOf((*T)(nil)).Elem()
}
```

## Max nesting depth

In order to avoid stack overflow all unmarshal-functions return an error if the Hjson input contains a tree more than 10000 levels deep. This is the same limit as in the Go standard library package `encoding/json`.

# API

[![godoc](https://godoc.org/github.com/hjson/hjson-go/v4?status.svg)](https://godoc.org/github.com/hjson/hjson-go/v4)

# History

[see releases](https://github.com/hjson/hjson-go/releases)
