package vgob

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
)

type (
	encoder struct {
		typ    reflect.Type
		sw     *switchWriter
		gobEnc *gob.Encoder
		mu     sync.Mutex
	}
	decoder struct {
		typ    reflect.Type
		sr     *switchReader
		gobDec *gob.Decoder
		mu     sync.Mutex
	}

	switchWriter struct {
		w io.Writer
	}
	switchReader struct {
		r io.Reader
	}
)

func newEncoder(typ reflect.Type) (*encoder, error) {
	var schemaBuf bytes.Buffer
	sw := newSwitchWriter(&schemaBuf)
	enc := gob.NewEncoder(sw)
	if err := enc.Encode(reflect.New(typ).Interface()); err != nil {
		return nil, err
	}
	return &encoder{
		gobEnc: enc,
		sw:     sw,
		typ:    typ,
	}, nil
}

func newDecoder(typ reflect.Type, schemaData []byte) (*decoder, error) {
	sr := newSwitchReader(bytes.NewReader(schemaData))
	dec := gob.NewDecoder(sr)
	if err := dec.Decode(reflect.New(typ).Interface()); err != nil {
		return nil, err
	}
	return &decoder{
		sr:     sr,
		gobDec: dec,
		typ:    typ,
	}, nil
}

func (enc *encoder) encode(w io.Writer, v interface{}) error {
	if t := getType(v); t != enc.typ {
		return fmt.Errorf("expect type %v but got %v", enc.typ, t)
	}

	enc.mu.Lock()
	enc.sw.switchTo(w)
	if err := enc.gobEnc.Encode(v); err != nil {
		enc.mu.Unlock()
		return err
	}
	enc.mu.Unlock()
	return nil
}

func (dec *decoder) decode(r io.Reader, v interface{}) error {
	if t := getType(v); t != dec.typ {
		return fmt.Errorf("expect type %v but got %v", dec.typ, t)
	}

	dec.mu.Lock()
	dec.sr.switchTo(r)
	if err := dec.gobDec.Decode(v); err != nil {
		dec.mu.Unlock()
		return err
	}
	dec.mu.Unlock()
	return nil
}

func newSwitchWriter(w io.Writer) *switchWriter {
	return &switchWriter{w: w}
}

func newSwitchReader(r io.Reader) *switchReader {
	return &switchReader{r: r}
}

func (s *switchWriter) Write(data []byte) (int, error) {
	return s.w.Write(data)
}

func (s *switchReader) Read(data []byte) (int, error) {
	return s.r.Read(data)
}

func (s *switchWriter) switchTo(w io.Writer) {
	s.w = w
}

func (s *switchReader) switchTo(r io.Reader) {
	s.r = r
}

func encodeVersion(w io.Writer, version uint) (int, error) {
	if version == 0 {
		return 0, errors.New("version should not be zero")
	}
	buf := make([]byte, binary.MaxVarintLen64)
	return w.Write(buf[:binary.PutUvarint(buf, uint64(version))])
}
func decodeVersion(r io.ByteReader) (uint, error) {
	version, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, err
	}
	if version == 0 {
		return 0, errors.New("version should not be zero")
	}
	return uint(version), nil
}

func encodeSchema(t reflect.Type) ([]byte, error) {
	var schemaBuf bytes.Buffer
	if err := gob.NewEncoder(&schemaBuf).Encode(reflect.New(t).Interface()); err != nil {
		return nil, err
	}
	return schemaBuf.Bytes(), nil
}

func getType(v interface{}) reflect.Type {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
