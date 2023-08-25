package jsonmap

import (
	"encoding/json"
	"testing"
)

func TestJSONMapMarshal(t *testing.T) {
	type Animal struct {
		Type uint8 `json:"type" jsonmap:"0:dog;1:cat;2:cow;3:others"`
	}

	var animal = Animal{Type: 2}
	res, err := json.MarshalIndent(Wrap(animal), "", "    ")
	t.Logf("\nres: %v\nerr %+v", string(res), err)
}

func TestJSONMapUnmarshal(t *testing.T) {
	var text = `{
		"type": "cow"
	}`

	type Animal struct {
		Type uint8 `json:"type" jsonmap:"0:dog;1:cat;2:cow;3:others"`
	}

	var animal = Animal{}
	err := json.Unmarshal([]byte(text), Wrap(&animal))
	t.Logf("\nres: %+v\nerr %+v", animal, err)
}
