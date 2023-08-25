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

func BenchmarkJSONMapUnmarshal(b *testing.B) {
	var text = `{
		"type": "cow"
	}`

	type Animal struct {
		Type uint8 `json:"type" jsonmap:"0:dog;1:cat;2:cow;3:others"`
	}

	for i := 0; i < b.N; i++ {
		var animal = Animal{}
		json.Unmarshal([]byte(text), Wrap(&animal))
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	var text = `{
		"type": "cow"
	}`

	type Animal struct {
		Type uint8 `json:"type" jsonmap:"0:dog;1:cat;2:cow;3:others"`
	}

	for i := 0; i < b.N; i++ {
		var animal = Animal{}
		json.Unmarshal([]byte(text), &animal)
	}
}
