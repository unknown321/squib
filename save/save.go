package save

import (
	"crypto/md5"
	"encoding/binary"
	"os"
	"reflect"
	"squib/dictionary"
	"squib/scriptvar"
	"squib/scriptvarscompositeslot"
)

// 0x140B285A7
// gvar name: {utf8@[rbx]}, hash: 0x{mem;4@rbx+8}, varsize: 0x{mem;1@rbx+10}, category 0x{mem;1@rbx+11}, arraysize 0x{mem;2@rbx+0xc}

// 0x140B288C7
// savevar name: {utf8@[r9+8]}, hash: 0x{mem;4@r9+10}, arraysize: 0x{mem;2@r9+14}, varsize: 0x{mem;1@r9+16}, category: 0x{mem;1@r9+17}

// varsize = varsize & 0x7

// 0x1401af4e0
// fox::FoxGameSaveCommon::DecodeSaveData
func Decrypt(key string, data []byte) {
	hash := md5.Sum([]byte(key))

	hashState := binary.LittleEndian.Uint32(hash[:])

	for i := 0; i <= len(data)-4; i += 4 {
		hashState ^= hashState << 0xd
		hashState ^= hashState >> 7
		hashState ^= hashState << 5

		block := binary.LittleEndian.Uint32(data[i:])
		decrypted := block ^ hashState
		binary.LittleEndian.PutUint32(data[i:], decrypted)
	}

	remaining := len(data) % 4
	if remaining > 0 {
		start := len(data) - remaining
		hashState ^= hashState << 0xd
		hashState ^= hashState >> 7
		tempHash := hashState ^ (hashState << 5)

		for i := 0; i < remaining; i++ {
			data[start+i] ^= byte(tempHash >> (8 * i))
		}
	}
}

type Save struct {
	Header    [16]byte // md5sum(TPP_GAME_DATA), see Decrypt
	Magic     [4]byte
	Svcs      scriptvarscompositeslot.ScriptVarsCompositeSlot
	ScriptVar []scriptvar.ScriptVar
}

func (s *Save) Parse(rawData []byte, dict dictionary.Dictionary) error {
	var err error
	off := 0

	if _, err = binary.Decode(rawData, binary.LittleEndian, &s.Header); err != nil {
		return err
	}
	off += len(s.Header)

	for {
		magic := rawData[off : off+4]
		off += 4

		switch string(magic) {
		case scriptvarscompositeslot.Magic:
			if err = s.Svcs.Parse(rawData[off:]); err != nil {
				return err
			}
			off += int(reflect.TypeOf(s.Svcs).Size())
		case scriptvar.Magic:
			if s.Svcs.Count == 0 {
				sv := scriptvar.ScriptVar{}
				if err = sv.Parse(rawData[off:], dict); err != nil {
					return err
				}
				s.ScriptVar = append(s.ScriptVar, sv)
			} else {
				for _, e := range s.Svcs.Entries {
					sv := scriptvar.ScriptVar{}
					if err = sv.Parse(rawData[e.Offset+16+4:], dict); err != nil {
						return err
					}
					s.ScriptVar = append(s.ScriptVar, sv)
				}
			}
			os.Exit(1)
		}
	}

	return nil
}
