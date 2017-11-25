package vgob

import (
	"os"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	type (
		S1 struct {
			V1 string
		}
		S2 struct {
			V2 string
		}
	)
	schemaFile := "test2.schema"
	defer os.RemoveAll(schemaFile)
	store, err := NewSchemaStore(schemaFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RegisterName("s1", S1{}); err != nil {
		t.Fatal(err)
	}
	if err := store.RegisterName("s2", S2{}); err != nil {
		t.Fatal(err)
	}
	marshaler, err := store.NewMarshaler()
	if err != nil {
		t.Fatal(err)
	}
	unmarshaler, err := store.NewUnmarshaler()
	if err != nil {
		t.Fatal(err)
	}
	{
		s1 := S1{"a"}
		buf, err := marshaler.Marshal(s1)
		if err != nil {
			t.Fatal(err)
		}
		var res S1
		if err := unmarshaler.Unmarshal(buf, &res); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(res, s1) {
			t.Fatalf("expect %v got %v", s1, res)
		}
	}
	{
		s2 := S2{"a"}
		buf, err := marshaler.Marshal(s2)
		if err != nil {
			t.Fatal(err)
		}
		var res S2
		if err := unmarshaler.Unmarshal(buf, &res); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(res, s2) {
			t.Fatalf("expect %v got %v", s2, res)
		}
	}
}
