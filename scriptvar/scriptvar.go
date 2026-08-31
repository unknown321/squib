package scriptvar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"strings"

	"github.com/unknown321/squib/dictionary"
	"github.com/unknown321/squib/savetype"
	"github.com/unknown321/squib/size"
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
	Skip      byte
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
	ScriptVersion uint32 // 0x0010063 for PERSONAL_DATA, 0x0001006A for GAME_DATA. Actual value is `ScriptVersion - TppDefine.PROGRAM_SAVE_FILE_VERSION_OFFSET`

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
			vp := table.ValueParams[i]
			key := table.Keys[i]
			var start, end int
			if vp.Size == size.Bool {
				startByte := int(vp.Offset / 8)
				start = entriesParamOffset + startByte
				end = start + (int(vp.Offset%8)+int(vp.ArraySize)+7)/8
			} else {
				fullsize := int(vp.ArraySize)
				switch vp.Size {
				case size.Int16, size.UInt16:
					fullsize *= 2
				case size.UInt32, size.Int32, size.Float:
					fullsize *= 4
				}

				start = -4 + int(vp.Offset) // sizeOffset = -4
				end = start + fullsize
			}

			value := rawData[start:end]
			name, ok := dict[key.Hash]
			if !ok {
				slog.Info("hash not found", "hash", fmt.Sprintf("%08x", key.Hash))
				name = []byte(fmt.Sprintf("%x", key.Hash))
			}

			out := strings.Builder{}
			fmt.Fprintf(&out, "\t%s (%d): [", name, vp.ArraySize)
			switch vp.Size {
			case size.Bool:
				startBit := uint64(vp.Offset % 8)
				for j := range int(vp.ArraySize) {
					bitOffset := startBit + uint64(j)
					qq := (value[bitOffset/8] & (1 << (bitOffset % 8))) != 0
					fmt.Fprintf(&out, "%t, ", qq)
				}
			case size.UInt32:
				for j := range int(vp.ArraySize) {
					qq := binary.LittleEndian.Uint32(value[j*4 : (j+1)*4])
					fmt.Fprintf(&out, "%d (%#08x), ", qq, qq)
				}
			case size.Int32:
				for j := range int(vp.ArraySize) {
					var qq int32
					if _, err = binary.Decode(value[j*4:(j+1)*4], binary.LittleEndian, &qq); err != nil {
						return err
					}

					fmt.Fprintf(&out, "%d (%#08x), ", qq, qq)
				}
			case size.UInt16:
				for j := range int(vp.ArraySize) {
					qq := binary.LittleEndian.Uint16(value[j*2 : (j+1)*2])
					fmt.Fprintf(&out, "%d (%#08x), ", qq, qq)
				}
			case size.Int16:
				for j := range int(vp.ArraySize) {
					var qq int16
					if _, err = binary.Decode(value[j*2:(j+1)*2], binary.LittleEndian, &qq); err != nil {
						return err
					}

					fmt.Fprintf(&out, "%d (%#08x), ", qq, qq)
				}
			case size.UInt8:
				if string(name) == "personalName" {
					out.Write(bytes.TrimRight(value, "\x00"))
				} else {
					for j := range int(vp.ArraySize) {
						qq := int(value[j*1 : (j+1)*1][0])
						fmt.Fprintf(&out, "%d, ", qq)
					}
				}
			case size.Int8:
				for j := range int(vp.ArraySize) {
					var qq int8
					if _, err = binary.Decode(value[j*1:(j+1)*1], binary.LittleEndian, &qq); err != nil {
						return err
					}

					fmt.Fprintf(&out, "%d, ", qq)
				}
			case size.Float:
				for j := range int(vp.ArraySize) {
					var qq float32
					if _, err = binary.Decode(value[j*4:(j+1)*4], binary.LittleEndian, &qq); err != nil {
						return nil
					}

					fmt.Fprintf(&out, "%f, ", qq)
				}
			}

			fmt.Print(strings.TrimSuffix(out.String(), ", ") + "]\n")
		}

		s.Table = append(s.Table, table)
	}

	return nil
}
