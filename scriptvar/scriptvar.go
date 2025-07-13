package scriptvar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"squib/dictionary"
	"squib/saveslot"
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

func (sect *Section) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, sect)
}

type Key struct {
	Hash   uint32
	Param1 uint16 // can be filled with index from indexes table
	Param2 uint16
}

func (k *Key) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, k)
}

type ValueParam struct {
	Offset    uint32
	ArraySize uint16
	Size      size.ESize
	Skip      byte
}

func (v *ValueParam) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, v)
}

type SectionTable struct {
	Indexes     []uint16
	Keys        []Key
	ValueParams []ValueParam
	Values      []any
}

func (st *SectionTable) Write(writer io.WriteSeeker) error {
	var err error

	if len(st.Keys) == 0 {
		if err = binary.Write(writer, binary.LittleEndian, make([]byte, 92)); err != nil {
			return err
		}
		return nil
	}

	if err = binary.Write(writer, binary.LittleEndian, st.Indexes); err != nil {
		return err
	}

	remainder := len(st.Indexes) * 2 % 8
	if remainder > 0 {
		v := make([]byte, remainder)
		if err = binary.Write(writer, binary.LittleEndian, v); err != nil {
			return err
		}
	}

	for _, k := range st.Keys {
		if err = k.Write(writer); err != nil {
			return err
		}
	}

	for _, v := range st.ValueParams {
		if err = v.Write(writer); err != nil {
			return err
		}
	}

	for i, v := range st.ValueParams {
		if _, err = writer.Seek(int64(v.Offset)+16, io.SeekStart); err != nil {
			return fmt.Errorf("%d %d, %w", v.Offset, 16, err)
		}

		if err = binary.Write(writer, binary.LittleEndian, st.Values[i]); err != nil {
			return err
		}
	}

	return nil
}

type ScriptVar struct {
	Type          savetype.ESaveType
	Skip          byte // manually set to zero
	SectionsCount byte
	ScriptVersion uint16 // see TppDefine, SAVE_FILE_INFO? some variables are missing
	SomeVar       uint16

	Sections []Section
	Tables   []SectionTable
}

func (s *ScriptVar) Parse(rawData []byte, dict dictionary.Dictionary) error {
	var err error
	buf := bytes.NewReader(rawData)
	for i, v := range []any{
		&s.Type,
		&s.Skip,
		&s.SectionsCount,
		&s.ScriptVersion,
		&s.SomeVar,
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

	for i, g := range s.Sections {
		if s.SectionsCount == 7 {
			fmt.Printf("Save slot %s (%d), %d entries\n", saveslot.GameData[i].Name, g.SectionID, g.EntriesCount)
		}
		table := SectionTable{}
		if g.EntriesCount == 0 {
			s.Tables = append(s.Tables, table)
			continue
		}

		//infoOffset := int(uint64(g.DataOffset+3) & 0xFFFFFFFFFFFFFFFC)
		for range int(g.EntriesCount) {
			var u uint16
			if err = binary.Read(buf, binary.LittleEndian, &u); err != nil {
				return err
			}
			//u := binary.LittleEndian.Uint16(rawData[infoOffset:])
			// u = entry number
			// u index = entry.hash % g.EntriesCount
			// if index used:
			//    go to hash at index
			//    if hash param1 == 0xffff, param1 = index
			table.Indexes = append(table.Indexes, u)
			//infoOffset += 2
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

		for ii := range int(g.EntriesCount) {
			offset := table.ValueParams[ii].Offset
			valueSize := table.ValueParams[ii].Size
			fullsize := 0
			sizeOffset := -4
			switch valueSize {
			case size.Bool, size.UInt8, size.Int8:
				fullsize = 1 * int(table.ValueParams[ii].ArraySize)
			case size.Int16, size.UInt16:
				fullsize = 2 * int(table.ValueParams[ii].ArraySize)
			case size.UInt32, size.Int32, size.Float:
				fullsize = 4 * int(table.ValueParams[ii].ArraySize)
			}
			o1 := sizeOffset + int(offset)
			o2 := int(offset) + fullsize + sizeOffset
			if o1 < 0 {
				o1 = 0
				o2 = fullsize
			}
			value := rawData[o1:o2]

			key, ok := dict[table.Keys[ii].Hash]
			if !ok {
				slog.Info("hash not found", "hash", fmt.Sprintf("%08x", table.Keys[ii].Hash))
				continue
			}

			arrs := table.ValueParams[ii].ArraySize
			out := fmt.Sprintf("\t%s (%d): [", key, arrs)
			switch valueSize {
			case size.Bool:
				var c []bool
				for index := range int(arrs) {
					qq := value[index]
					out += fmt.Sprintf("%t, ", qq > 0)
					c = append(c, qq > 0)
				}
				table.Values = append(table.Values, c)
			case size.UInt32:
				var c []uint32
				for index := range int(arrs) {
					qq := binary.LittleEndian.Uint32(value[index*4 : (index+1)*4])
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
					c = append(c, qq)
				}
				table.Values = append(table.Values, c)
			case size.Int32:
				var c []int32
				for index := range int(arrs) {
					var qq int32
					if _, err = binary.Decode(value[index*4:(index+1)*4], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
					c = append(c, qq)
				}
				table.Values = append(table.Values, c)
			case size.UInt16:
				var c []uint16
				for index := range int(arrs) {
					qq := binary.LittleEndian.Uint16(value[index*2 : (index+1)*2])
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
					c = append(c, qq)
				}
				table.Values = append(table.Values, c)
			case size.Int16:
				var c []int16
				for index := range int(arrs) {
					var qq int16
					if _, err = binary.Decode(value[index*2:(index+1)*2], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("%d (%#08x), ", qq, qq)
					c = append(c, qq)
				}
				table.Values = append(table.Values, c)
			case size.UInt8:
				var c []uint8
				switch string(key) {
				case "personalName":
					out += fmt.Sprintf("%s", bytes.TrimRight(value, "\x00"))
					for index := range int(arrs) {
						qq := uint8(value[index*1 : (index+1)*1][0])
						c = append(c, qq)
					}
				default:
					for index := range int(arrs) {
						qq := int(value[index*1 : (index+1)*1][0])
						out += fmt.Sprintf("%d, ", qq)
						c = append(c, uint8(value[index*1 : (index+1)*1][0]))
					}
				}
				table.Values = append(table.Values, c)
			case size.Int8:
				var c []int8
				for index := range int(arrs) {
					var qq int8
					if _, err = binary.Decode(value[index*1:(index+1)*1], binary.LittleEndian, &qq); err != nil {
						return err
					}
					out += fmt.Sprintf("%d, ", qq)
					c = append(c, qq)
				}
				table.Values = append(table.Values, c)
			case size.Float:
				var c []float32
				for index := range int(arrs) {
					var qq float32
					if _, err = binary.Decode(value[index*4:(index+1)*4], binary.LittleEndian, &qq); err != nil {
						return nil
					}
					out += fmt.Sprintf("%f, ", qq)
					c = append(c, qq)
				}
				table.Values = append(table.Values, c)
			}

			out = strings.TrimSuffix(out, ", ")
			out += "]\n"

			fmt.Print(out)
		}
		s.Tables = append(s.Tables, table)
	}

	return nil
}

func (s *ScriptVar) Write(w io.WriteSeeker, size int) error {
	var err error

	v := []any{[]byte(Magic), s.Type, s.Skip, s.SectionsCount, s.ScriptVersion, s.SomeVar}
	for i, k := range v {
		if err = binary.Write(w, binary.LittleEndian, k); err != nil {
			return fmt.Errorf("scriptvar struct key %d: %w", i, err)
		}
	}

	for _, e := range s.Sections {
		if err = e.Write(w); err != nil {
			return err
		}
	}

	for _, e := range s.Tables {
		if err = e.Write(w); err != nil {
			return err
		}
	}

	return nil
}
