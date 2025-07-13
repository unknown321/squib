package scriptvarscompositeslot

import (
	"encoding/binary"
	"reflect"
)

var Magic = "SVCS"

type Entry struct {
	Offset uint32
	Size1  uint32
	Size2  uint32
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
