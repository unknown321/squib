package save

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"squib/dictionary"
	"squib/size"
)

// 0x140B285A7
// gvar name: {utf8@[rbx]}, hash: 0x{mem;4@rbx+8}, varsize: 0x{mem;1@rbx+10}, category 0x{mem;1@rbx+11}, arraysize 0x{mem;2@rbx+0xc}

// 0x140B288C7
// savevar name: {utf8@[r9+8]}, hash: 0x{mem;4@r9+10}, arraysize: 0x{mem;2@r9+14}, varsize: 0x{mem;1@r9+16}, category: 0x{mem;1@r9+17}

// varsize = varsize & 0x7

type ScriptVars struct {
	Header []byte
}

type SaveGroup struct {
	GroupID      uint16 // not groupbit
	EntriesCount uint16
	DataOffset   uint32
}

type HashEntry struct {
	Hash   uint32
	Param1 uint16
	Param2 uint16
}

type EntryParam struct {
	Offset    uint32
	ArraySize uint16
	Size      size.ESize
	_         byte
}

type Table struct {
	Indexes     []uint16
	Entries     []HashEntry
	EntryParams []EntryParam
}

type Save struct {
	Header       [16]byte
	Magic        [4]byte
	Type         uint16
	MysteryByte  byte // manually set to zero
	GroupsCount  byte
	MysteryParam uint32 // 0x0010063 for PERSONAL_DATA

	Groups []SaveGroup
	Table  []Table
}

func (s *Save) Parse(rawData []byte, dict dictionary.Dictionary) error {
	var err error
	buf := bytes.NewReader(rawData)
	for i, v := range []any{
		&s.Header,
		&s.Magic,
		&s.Type,
		&s.MysteryByte,
		&s.GroupsCount,
		&s.MysteryParam,
	} {
		if err = binary.Read(buf, binary.LittleEndian, v); err != nil {
			return fmt.Errorf("binary read field %d: %w", i, err)
		}
	}

	for range int(s.GroupsCount) {
		g := SaveGroup{}
		if err = binary.Read(buf, binary.LittleEndian, &g); err != nil {
			return err
		}
		s.Groups = append(s.Groups, g)
	}

	for _, g := range s.Groups {
		fmt.Printf("Group %d, %d entries\n", g.GroupID, g.EntriesCount)
		table := Table{}
		if g.EntriesCount == 0 {
			s.Table = append(s.Table, table)
			continue
		}

		infoOffset := int((uint64(len(s.Header)) + uint64(g.DataOffset+3)) & 0xFFFFFFFFFFFFFFFC)
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

		hashesOffset := int((uint64(len(s.Header)) + uint64(g.DataOffset+3+uint32(g.EntriesCount)*2)) & 0xFFFFFFFFFFFFFFFC)
		entriesParamOffset := hashesOffset + int(g.EntriesCount)*8

		for range int(g.EntriesCount) {
			h := HashEntry{}
			if _, err = binary.Decode(rawData[hashesOffset:], binary.LittleEndian, &h); err != nil {
				return err
			}
			table.Entries = append(table.Entries, h)

			hashesOffset += 8
		}

		for range int(g.EntriesCount) {
			p := EntryParam{}
			if _, err = binary.Decode(rawData[entriesParamOffset:], binary.LittleEndian, &p); err != nil {
				return err
			}
			entriesParamOffset += 8
			table.EntryParams = append(table.EntryParams, p)
		}

		for i := range int(g.EntriesCount) {
			offset := table.EntryParams[i].Offset
			valueSize := table.EntryParams[i].Size
			fullsize := 0
			sizeOffset := 16
			switch valueSize {
			case size.Bool, size.UInt8, size.Int8:
				fullsize = 1 * int(table.EntryParams[i].ArraySize)
			case size.Int16, size.UInt16:
				fullsize = 2 * int(table.EntryParams[i].ArraySize)
			case size.UInt32, size.Int32, size.Float:
				fullsize = 4 * int(table.EntryParams[i].ArraySize)
			}
			o1 := sizeOffset + int(offset)
			o2 := int(offset) + fullsize + sizeOffset
			value := rawData[o1:o2]

			key, ok := dict[table.Entries[i].Hash]
			if !ok {
				slog.Info("hash not found", "hash", fmt.Sprintf("%08x", table.Entries[i].Hash))
				continue
			}

			arrs := table.EntryParams[i].ArraySize
			out := ""
			switch valueSize {
			case size.Bool:
				for index := range int(arrs) {
					qq := value[index]
					out += fmt.Sprintf("\t%s (%d/%d): %b\n", key, index+1, arrs, qq)
				}
			case size.UInt32:
				for index := range int(arrs) {
					qq := binary.LittleEndian.Uint32(value[index*4 : (index+1)*4])
					out += fmt.Sprintf("\t%s (%d/%d): %d (0x%08x)\n", key, index+1, arrs, qq, qq)
				}
			case size.Int32:
				for index := range int(arrs) {
					var qq int32
					if _, err = binary.Decode(value[index*4:(index+1)*4], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("\t%s (%d/%d): %d (0x%08x)\n", key, index+1, arrs, qq, qq)
				}
			case size.UInt16:
				for index := range int(arrs) {
					qq := binary.LittleEndian.Uint16(value[index*2 : (index+1)*2])
					out += fmt.Sprintf("\t%s (%d/%d): %d\n", key, index+1, arrs, qq)
				}
			case size.Int16:
				for index := range int(arrs) {
					var qq int16
					if _, err = binary.Decode(value[index*2:(index+1)*2], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("\t%s (%d/%d): %d\n", key, index+1, arrs, qq)
				}
			case size.UInt8:
				switch string(key) {
				case "personalName":
					out += fmt.Sprintf("\t%s: %s\n", key, bytes.TrimRight(value, "\x00"))
				case "emblemFlag", "avatarMotionFrame":
					out += fmt.Sprintf("\t%s: %v\n", key, value)
				default:
					for index := range int(arrs) {
						qq := int(value[index*1 : (index+1)*1][0])
						out += fmt.Sprintf("\t%s (%d/%d): %d\n", key, index+1, arrs, qq)
					}
				}
			case size.Int8:
				for index := range int(arrs) {
					var qq int8
					if _, err = binary.Decode(value[index*1:(index+1)*1], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("\t%s (%d/%d): %d\n", key, index+1, arrs, qq)
				}
			case size.Float:
				for index := range int(arrs) {
					var qq float32
					if _, err = binary.Decode(value[index*4:(index+1)*4], binary.LittleEndian, &qq); err != nil {
						return nil
					}
					out += fmt.Sprintf("\t%s (%d/%d): %f\n", key, index+1, arrs, qq)
				}
			}

			fmt.Print(out)
		}
		s.Table = append(s.Table, table)
	}

	return nil
}
