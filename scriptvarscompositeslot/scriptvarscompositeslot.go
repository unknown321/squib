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
	off := 0
	if _, err = binary.Decode(rawData, binary.LittleEndian, &s.Type); err != nil {
		return err
	}
	off += 2
	off += 1 // skip
	s.Count = int(rawData[off : off+1][0])
	off += 1

	for range int(s.Count) {
		e := Entry{}
		if _, err = binary.Decode(rawData[off:], binary.LittleEndian, &e); err != nil {
			return err
		}
		s.Entries = append(s.Entries, e)
		// NOTE: reflection is expensive and could be removed.
		off += int(reflect.TypeFor[Entry]().Size()) // off += 12
	}

	return nil
}
