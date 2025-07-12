package scriptvar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"squib/dictionary"
	"squib/savetype"
	"squib/size"
	"strings"
)

var Magic = "SVAR"

type Section struct {
	SectionID    uint16 // NOT groupbit
	EntriesCount uint16
	DataOffset   uint32
}

type Key struct {
	Hash   uint32
	Param1 uint16 // can be filled with index from indexes table
	Param2 uint16
}

type ValueParam struct {
	Offset    uint32
	ArraySize uint16
	Size      size.ESize
	_         byte
}

type CategoryTable struct {
	Indexes     []uint16
	Keys        []Key
	ValueParams []ValueParam
}

type ScriptVar struct {
	Type          savetype.ESaveType
	MysteryByte   byte // manually set to zero
	SectionsCount byte
	ScriptVersion uint32 // 0x0010063 for PERSONAL_DATA, 0x0001006A for GAME_DATA

	Sections []Section
	Table    []CategoryTable
}

func (s *ScriptVar) Parse(rawData []byte, dict dictionary.Dictionary) error {
	var err error
	buf := bytes.NewReader(rawData)
	for i, v := range []any{
		&s.Type,
		&s.MysteryByte,
		&s.SectionsCount,
		&s.ScriptVersion,
	} {
		if err = binary.Read(buf, binary.LittleEndian, v); err != nil {
			return fmt.Errorf("binary read field %d: %w", i, err)
		}
	}
	fmt.Printf("ScriptVar, type: %s, version: %#08x\n", s.Type, s.ScriptVersion)

	for range int(s.SectionsCount) {
		g := Section{}
		if err = binary.Read(buf, binary.LittleEndian, &g); err != nil {
			return err
		}
		s.Sections = append(s.Sections, g)
	}

	for _, g := range s.Sections {
		fmt.Printf("Section %d, %d entries\n", g.SectionID, g.EntriesCount)
		table := CategoryTable{}
		if g.EntriesCount == 0 {
			s.Table = append(s.Table, table)
			continue
		}

		infoOffset := int(uint64(g.DataOffset+3) & 0xFFFFFFFFFFFFFFFC)
		for range int(g.EntriesCount) {
			u := binary.LittleEndian.Uint16(rawData[infoOffset:])
			// u = entry number
			// u index = entry.hash % g.EntriesCount
			// if index used:
			//    go to hash at index
			//    if hash param1 == 0xffff, param1 = index
			table.Indexes = append(table.Indexes, u)
			infoOffset += 2
		}

		hashesOffset := int(uint64(g.DataOffset+3+uint32(g.EntriesCount)*2)&0xFFFFFFFFFFFFFFFC) - 4
		entriesParamOffset := hashesOffset + int(g.EntriesCount)*8

		for range int(g.EntriesCount) {
			h := Key{}
			if _, err = binary.Decode(rawData[hashesOffset:], binary.LittleEndian, &h); err != nil {
				return err
			}
			table.Keys = append(table.Keys, h)

			hashesOffset += 8
		}

		for range int(g.EntriesCount) {
			p := ValueParam{}
			if _, err = binary.Decode(rawData[entriesParamOffset:], binary.LittleEndian, &p); err != nil {
				return err
			}
			entriesParamOffset += 8
			table.ValueParams = append(table.ValueParams, p)
		}

		for i := range int(g.EntriesCount) {
			offset := table.ValueParams[i].Offset
			valueSize := table.ValueParams[i].Size
			fullsize := 0
			sizeOffset := -4
			switch valueSize {
			case size.Bool, size.UInt8, size.Int8:
				fullsize = 1 * int(table.ValueParams[i].ArraySize)
			case size.Int16, size.UInt16:
				fullsize = 2 * int(table.ValueParams[i].ArraySize)
			case size.UInt32, size.Int32, size.Float:
				fullsize = 4 * int(table.ValueParams[i].ArraySize)
			}
			o1 := sizeOffset + int(offset)
			o2 := int(offset) + fullsize + sizeOffset
			if o1 < 0 {
				o1 = 0
				o2 = fullsize
			}
			value := rawData[o1:o2]

			key, ok := dict[table.Keys[i].Hash]
			if !ok {
				slog.Info("hash not found", "hash", fmt.Sprintf("%08x", table.Keys[i].Hash))
				continue
			}

			arrs := table.ValueParams[i].ArraySize
			out := fmt.Sprintf("\t%s (%d): [", key, arrs)
			switch valueSize {
			case size.Bool:
				for index := range int(arrs) {
					qq := value[index]
					out += fmt.Sprintf("%t, ", qq > 0)
				}
			case size.UInt32:
				for index := range int(arrs) {
					qq := binary.LittleEndian.Uint32(value[index*4 : (index+1)*4])
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
				}
			case size.Int32:
				for index := range int(arrs) {
					var qq int32
					if _, err = binary.Decode(value[index*4:(index+1)*4], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
				}
			case size.UInt16:
				for index := range int(arrs) {
					qq := binary.LittleEndian.Uint16(value[index*2 : (index+1)*2])
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
				}
			case size.Int16:
				for index := range int(arrs) {
					var qq int16
					if _, err = binary.Decode(value[index*2:(index+1)*2], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
				}
			case size.UInt8:
				switch string(key) {
				case "personalName":
					out += fmt.Sprintf("%s", bytes.TrimRight(value, "\x00"))
				default:
					for index := range int(arrs) {
						qq := int(value[index*1 : (index+1)*1][0])
						out += fmt.Sprintf("%d, ", qq)
					}
				}
			case size.Int8:
				for index := range int(arrs) {
					var qq int8
					if _, err = binary.Decode(value[index*1:(index+1)*1], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("%d, ", qq)
				}
			case size.Float:
				for index := range int(arrs) {
					var qq float32
					if _, err = binary.Decode(value[index*4:(index+1)*4], binary.LittleEndian, &qq); err != nil {
						return nil
					}
					out += fmt.Sprintf("%f, ", qq)
				}
			}

			out = strings.TrimSuffix(out, ", ")
			out += "]\n"

			fmt.Print(out)
		}
		s.Table = append(s.Table, table)
	}

	return nil
}
