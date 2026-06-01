// Package jsonschema contains Workbench's small JSON Schema construction
// primitives. It deliberately builds on Google's implementation instead of
// inventing a parallel schema model.
package jsonschema

import gjsonschema "github.com/google/jsonschema-go/jsonschema"

type Schema = gjsonschema.Schema

func EmptyObject() *gjsonschema.Schema {
	return StrictObject(nil)
}

func StrictObject(properties map[string]*gjsonschema.Schema, required ...string) *gjsonschema.Schema {
	return &gjsonschema.Schema{
		Type:                 "object",
		Properties:           properties,
		Required:             required,
		AdditionalProperties: False(),
	}
}

func String() *gjsonschema.Schema {
	return &gjsonschema.Schema{Type: "string"}
}

func Bool() *gjsonschema.Schema {
	return &gjsonschema.Schema{Type: "boolean"}
}

func Null() *gjsonschema.Schema {
	return &gjsonschema.Schema{Type: "null"}
}

func ArrayOf(items *gjsonschema.Schema) *gjsonschema.Schema {
	return &gjsonschema.Schema{
		Type:  "array",
		Items: items,
	}
}

func StringMap() *gjsonschema.Schema {
	return &gjsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: String(),
	}
}

func Nullable(schema *gjsonschema.Schema) *gjsonschema.Schema {
	return &gjsonschema.Schema{AnyOf: []*gjsonschema.Schema{schema, Null()}}
}

func NullableString() *gjsonschema.Schema {
	return &gjsonschema.Schema{Types: []string{"string", "null"}}
}

func False() *gjsonschema.Schema {
	return &gjsonschema.Schema{Not: &gjsonschema.Schema{}}
}
