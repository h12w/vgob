package vgob

import (
	"fmt"
	"os"
	"testing"
)

func TestAddField(t *testing.T) {
	schemaFile := "test1.schema"
	defer os.RemoveAll(schemaFile)

	var data []byte
	{
		type S struct {
			V string
		}
		schemaStore, err := NewSchemaStore(schemaFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.RegisterName("S", S{}); err != nil {
			t.Fatal(err)
		}
		m, err := schemaStore.NewTypeMarshaler("S")
		if err != nil {
			t.Fatal(err)
		}
		data, err = m.Marshal(&S{V: "a"})
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.Save(); err != nil {
			t.Fatal(err)
		}
	}
	{
		type S struct {
			V  string
			V1 string
		}
		schemaStore, err := NewSchemaStore(schemaFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.RegisterName("S", S{}); err != nil {
			t.Fatal(err)
		}
		u, err := schemaStore.NewTypeUnmarshaler("S")
		if err != nil {
			t.Fatal(err)
		}
		s := new(S)
		if err := u.Unmarshal(data, s); err != nil {
			t.Fatal(err)
		}
		actual := fmt.Sprintf("%#v", *s)
		expected := `vgob.S{V:"a", V1:""}`
		if actual != expected {
			t.Fatalf("expect %s got %s", expected, actual)
		}
	}
}

func TestRemoveField(t *testing.T) {
	schemaFile := "test2.schema"
	defer os.RemoveAll(schemaFile)

	var data []byte
	{
		type S struct {
			V  string
			V1 string
		}
		schemaStore, err := NewSchemaStore(schemaFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.RegisterName("S", S{}); err != nil {
			t.Fatal(err)
		}
		m, err := schemaStore.NewTypeMarshaler("S")
		if err != nil {
			t.Fatal(err)
		}
		data, err = m.Marshal(&S{V: "a", V1: "b"})
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.Save(); err != nil {
			t.Fatal(err)
		}
	}
	{
		type S struct {
			V string
		}
		schemaStore, err := NewSchemaStore(schemaFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.RegisterName("S", S{}); err != nil {
			t.Fatal(err)
		}
		u, err := schemaStore.NewTypeUnmarshaler("S")
		if err != nil {
			t.Fatal(err)
		}
		s := new(S)
		if err := u.Unmarshal(data, s); err != nil {
			t.Fatal(err)
		}
		actual := fmt.Sprintf("%#v", *s)
		expected := `vgob.S{V:"a"}`
		if actual != expected {
			t.Fatalf("expect %s got %s", expected, actual)
		}
	}
}

func TestRenameType(t *testing.T) {
	schemaFile := "test3.schema"
	defer os.RemoveAll(schemaFile)

	var data []byte
	{
		type S struct {
			V string
		}
		schemaStore, err := NewSchemaStore(schemaFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.RegisterName("S", S{}); err != nil {
			t.Fatal(err)
		}
		m, err := schemaStore.NewTypeMarshaler("S")
		if err != nil {
			t.Fatal(err)
		}
		data, err = m.Marshal(&S{V: "a"})
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.Save(); err != nil {
			t.Fatal(err)
		}
	}
	{
		type T struct {
			V string
		}
		schemaStore, err := NewSchemaStore(schemaFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := schemaStore.RegisterName("S", T{}); err != nil {
			t.Fatal(err)
		}
		u, err := schemaStore.NewTypeUnmarshaler("S")
		if err != nil {
			t.Fatal(err)
		}
		s := new(T)
		if err := u.Unmarshal(data, s); err != nil {
			t.Fatal(err)
		}
		actual := fmt.Sprintf("%#v", *s)
		expected := `vgob.T{V:"a"}`
		if actual != expected {
			t.Fatalf("expect %s got %s", expected, actual)
		}
	}
}

func TestTypeMismatchError(t *testing.T) {
	type T1 struct{}
	type T2 struct{}

	schemaFile := "test4.schema"
	schemaStore, err := NewSchemaStore(schemaFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := schemaStore.RegisterName("T1", T1{}); err != nil {
		t.Fatal(err)
	}
	if err := schemaStore.RegisterName("T2", T2{}); err != nil {
		t.Fatal(err)
	}

	{
		var err error
		m, err := schemaStore.NewTypeMarshaler("T1")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := m.Marshal(T2{}); err == nil {
			t.Fatal("expect type mismatch error")
		}
	}
	{
		u, err := schemaStore.NewTypeUnmarshaler("T1")
		if err != nil {
			t.Fatal(err)
		}
		if u.Unmarshal(nil, &T2{}) == nil {
			t.Fatal("expect type mismatch error")
		}
	}
}
