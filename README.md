# jsonmap
golang json serialization field mapper, using struct tag

## Usage

```go
package main

import (
    "encoding/json"
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

    var animal = new(Animal)
    _ = json.Unmarshal([]byte(text), jsonmap.Wrap(animal))
    fmt.Println(animal)

    // Animal{Type: 2}

    res, _ := json.MarshalIndent(jsonmap.Wrap(animal), "", "    ")
    fmt.Println(res)

    // {"type": "cow"}
}
```