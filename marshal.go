package vgob

import (
	"bytes"
	"fmt"
	"reflect"
)

type (
	Marshaler struct {
		ms map[reflect.Type]*TypeMarshaler
	}
	Unmarshaler struct {
		us map[reflect.Type]*TypeUnmarshaler
	}
	TypeMarshaler struct {
		ver uint
		enc *encoder
	}
	TypeUnmarshaler struct {
		decs map[uint]*decoder
	}
)

func (m *Marshaler) Marshal(v interface{}) ([]byte, error) {
	t := getType(v)
	marshaler, ok := m.ms[t]
	if !ok {
		return nil, fmt.Errorf("marshaler of type %v cannot be found", t)
	}
	return marshaler.Marshal(v)
}

func (u *Unmarshaler) Unmarshal(data []byte, v interface{}) error {
	t := getType(v)
	unmarshaler, ok := u.us[t]
	if !ok {
		return fmt.Errorf("unmarshaler of type %v cannot be found", t)
	}
	return unmarshaler.Unmarshal(data, v)
}

// Marshal marshals v into []byte and returns the result
func (m *TypeMarshaler) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := encodeVersion(&buf, m.ver); err != nil {
		return nil, err
	}
	if err := m.enc.encode(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (u *TypeUnmarshaler) Unmarshal(data []byte, v interface{}) error {
	r := bytes.NewReader(data)
	ver, err := decodeVersion(r)
	if err != nil {
		return err
	}
	dec, ok := u.decs[uint(ver)]
	if !ok {
		return fmt.Errorf("missing decoder for type %v, version %d", reflect.TypeOf(v), ver)
	}
	return dec.decode(r, v)
}
