package dictionary

import (
	"bytes"
	"embed"
	"github.com/unknown321/hashing"
)

type Dictionary map[uint32][]byte

var Dict = make(map[uint32][]byte)

func Init(f *embed.FS) error {
	data, err := f.ReadFile("savevars.txt")
	if err != nil {
		return err
	}

	//var dictData []byte
	//dictData, _ = os.ReadFile("./dict.txt")
	//data = append(data, dictData...)

	dd := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	lines := bytes.Split(dd, []byte("\n"))

	for _, line := range lines {
		if len(line) > 100 {
			continue
		}
		v := uint32(hashing.StrCode64(bytes.TrimSuffix(line, []byte("\n"))) & 0xffffffff) // known as strcode32 in HashWrangler
		Dict[v] = line
	}

	return nil
}
