# jsonmap
golang json serialization field mapper, using struct tag

## Usage

```go
package main

import (
    "fmt"

    "github.com/chenhg5/jsonmap"
)

func main() {

    var text = `{
		"type": "cow"
	}`

	type Animal struct {
		Type uint8 `json:"type" jsonmap:"0:dog;1:cat;2:cow;3:others"`
	}

	var animal = Animal{}
	json.Unmarshal([]byte(text), Wrap(&animal))

    // Animal{Type: 2}

    res, _ := json.MarshalIndent(Wrap(animal), "", "    ")
    fmt.Println(res)

    // {"type": "cow"}
}


```