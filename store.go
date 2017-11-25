// vgob is versioned encoding/gob
package vgob

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
)

type (
	SchemaStore struct {
		schemas schemas
		file    string
	}
)

type (
	schemas map[string]*schema
	schema  struct {
		Versions schemaVersions
		typ      reflect.Type // unmarshaled in gob
	}
	schemaVersions map[string]uint
)

func NewSchemaStore(file string) (*SchemaStore, error) {
	s := &SchemaStore{file: file, schemas: make(schemas)}
	f, err := os.Open(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return s, nil
	}
	defer f.Close()
	if err := gob.NewDecoder(f).Decode(&s.schemas); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SchemaStore) RegisterName(name string, v interface{}) error {
	typ := getType(v)
	schemaBytes, err := encodeSchema(typ)
	if err != nil {
		return err
	}
	schemaStr := string(schemaBytes)
	if schema, schemaExists := s.schemas[name]; schemaExists {
		schema.typ = typ
		if schema.Versions[schemaStr] == 0 {
			schema.Versions[schemaStr] = uint(len(schema.Versions)) + 1
		}
		return nil
	}
	s.schemas[name] = &schema{
		typ: typ,
		Versions: schemaVersions{
			schemaStr: 1,
		},
	}
	return nil
}

func (s *SchemaStore) Save() error {
	tmpfile := s.file + ".tmp"
	f, err := os.Create(tmpfile)
	if err != nil {
		return nil
	}
	if err := gob.NewEncoder(f).Encode(&s.schemas); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmpfile, s.file)
}

func (s *SchemaStore) NewMarshaler() (*Marshaler, error) {
	ms := make(map[reflect.Type]*TypeMarshaler)
	for name, schema := range s.schemas {
		marshaler, err := s.NewTypeMarshaler(name)
		if err != nil {
			return nil, err
		}
		ms[schema.typ] = marshaler
	}
	return &Marshaler{ms: ms}, nil
}

func (s *SchemaStore) NewUnmarshaler() (*Unmarshaler, error) {
	us := make(map[reflect.Type]*TypeUnmarshaler)
	for name, schema := range s.schemas {
		unmarshaler, err := s.NewTypeUnmarshaler(name)
		if err != nil {
			return nil, err
		}
		us[schema.typ] = unmarshaler
	}
	return &Unmarshaler{us: us}, nil
}

// NewTypeMarshaler creates a new Marshaler for type of v
func (s *SchemaStore) NewTypeMarshaler(name string) (*TypeMarshaler, error) {
	schema, ok := s.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema for %s is not registered", name)
	}
	if schema.typ == nil {
		return nil, fmt.Errorf("type %s not registered", name)
	}

	enc, err := newEncoder(schema.typ)
	if err != nil {
		return nil, err
	}
	return &TypeMarshaler{
		enc: enc,
		ver: uint(len(schema.Versions)),
	}, nil
}

// NewTypeUnmarshaler creates a new Unmarshaler for type of v
func (s *SchemaStore) NewTypeUnmarshaler(name string) (*TypeUnmarshaler, error) {
	schema, ok := s.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema for %s is not registered", name)
	}
	if schema.typ == nil {
		return nil, fmt.Errorf("type %s not registered", name)
	}

	decs := make(map[uint]*decoder)
	for schemaData, version := range schema.Versions {
		dec, err := newDecoder(schema.typ, []byte(schemaData))
		if err != nil {
			return nil, err
		}
		decs[version] = dec
	}
	return &TypeUnmarshaler{
		decs: decs,
	}, nil
}
