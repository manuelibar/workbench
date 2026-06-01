package jsonschema

import (
	"encoding/json"
	"testing"
)

func TestStrictObjectRejectsAdditionalProperties(t *testing.T) {
	schema := StrictObject(map[string]*Schema{
		"name": String(),
	}, "name")
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["type"] != "object" {
		t.Fatalf("type = %v, want object", got["type"])
	}
	if got["additionalProperties"] != false {
		t.Fatalf("additionalProperties = %#v, want false", got["additionalProperties"])
	}
	required := got["required"].([]any)
	if len(required) != 1 || required[0] != "name" {
		t.Fatalf("required = %#v, want [name]", got["required"])
	}
}

func TestNullableString(t *testing.T) {
	raw, err := json.Marshal(NullableString())
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"type":["string","null"]}` {
		t.Fatalf("nullable string schema = %s", raw)
	}
}

func TestStringMap(t *testing.T) {
	raw, err := json.Marshal(StringMap())
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"type":"object","additionalProperties":{"type":"string"}}` {
		t.Fatalf("string map schema = %s", raw)
	}
}
