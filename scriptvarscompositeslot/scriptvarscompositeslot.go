package scriptvarscompositeslot

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

var Magic = "SVCS"

type Entry struct {
	Offset uint32
	Size1  uint32
	Size2  uint32
}

func (e *Entry) Write(w io.Writer) error {
	v := []uint32{e.Offset, e.Size1, e.Size2}
	for _, vv := range v {
		if err := binary.Write(w, binary.LittleEndian, vv); err != nil {
			return err
		}
	}

	return nil
}

type ScriptVarsCompositeSlot struct {
	Type    uint16
	Skip    byte
	Count   int
	Entries []Entry
}

func (s *ScriptVarsCompositeSlot) Parse(rawData []byte) error {
	var err error
	offset := 0
	if _, err = binary.Decode(rawData, binary.LittleEndian, &s.Type); err != nil {
		return err
	}
	offset += 2
	offset += 1 // skip
	s.Count = int(rawData[offset : offset+1][0])
	offset += 1

	for range int(s.Count) {
		e := Entry{}
		if _, err = binary.Decode(rawData[offset:], binary.LittleEndian, &e); err != nil {
			return err
		}
		s.Entries = append(s.Entries, e)
		offset += int(reflect.TypeOf(e).Size())
	}

	return nil
}

func (s *ScriptVarsCompositeSlot) Write(w io.Writer) error {
	var err error
	v := []any{[]byte(Magic), s.Type, s.Skip, byte(s.Count)}
	for i, k := range v {
		if err = binary.Write(w, binary.LittleEndian, k); err != nil {
			return fmt.Errorf("composite slot struct key %d: %w", i, err)
		}
	}

	for _, e := range s.Entries {
		if err = e.Write(w); err != nil {
			return err
		}
	}

	return nil
}
