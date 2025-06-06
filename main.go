package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	_ "embed"
	"encoding/binary"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/unknown321/hashing"
)

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

//go:embed save_dict
//go:embed config_dict
//go:embed personal_dict
//go:embed graphics_dict
//go:embed mgo_dict
var f embed.FS

func main() {
	var keysOnly bool
	flag.CommandLine.SetOutput(os.Stdout)
	flag.BoolVar(&keysOnly, "keysOnly", false, "print only keys without address and values")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("TNT: MGSV:TPP save decoder")
		fmt.Println()
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("\t%s [OPTION] FILE\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Provide more keys by adding them to dict.txt in the same directory as executable.")
		fmt.Println("Decoded file is created in the same directory with '_decoded' suffix.")
		os.Exit(1)
	}

	data, _ := f.ReadFile("save_dict")
	configData, _ := f.ReadFile("config_dict")
	personalData, _ := f.ReadFile("personal_dict")
	graphicsData, _ := f.ReadFile("graphics_dict")
	mgoData, _ := f.ReadFile("mgo_dict")

	var dictData []byte
	dictData, _ = os.ReadFile("./dict.txt")

	data = append(data, dictData...)
	data = append(data, configData...)
	data = append(data, personalData...)
	data = append(data, graphicsData...)
	data = append(data, mgoData...)

	dd := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	lines := bytes.Split(dd, []byte("\n"))

	dict := map[uint32][]byte{}
	for _, line := range lines {
		if len(line) > 100 {
			continue
		}
		v := uint32(hashing.StrCode64(bytes.TrimSuffix(line, []byte("\n"))) & 0xffffffff) // known as strcode32 in HashWrangler
		dict[v] = line
	}

	var filename string
	if len(os.Args) < 2 {
		filename = "TPP_GAME_DATA1"
	} else {
		filename = os.Args[1]
		if os.Args[1] == "-keysOnly" {
			filename = os.Args[2]
		}
	}

	s, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("open file", "error", err.Error(), "filename", filename)
		os.Exit(1)
	}

	keys := []string{"TPP_GAME_DATA", "TPP_CONFIG_DATA", "PERSONAL_DATA", "TPP_GRAPHICS_CONFIG", "MGO_GAME_DATA"}
	var key string
	for _, k := range keys {
		if strings.Contains(filename, k) {
			key = k
			break
		}
	}
	if key == "" {
		slog.Error("file must have one of accepted names;", "names", fmt.Sprintf("%+v", keys))
		os.Exit(1)
	}

	Decrypt(key, s)

	out := filename + "_decoded"
	if err = os.WriteFile(out, s, 0644); err != nil {
		slog.Error("save decoded", "error", err.Error(), "filename", out)
		os.Exit(1)
	}

	ss := make([]byte, 4)
	var indexes []int
	for k := range dict {
		binary.LittleEndian.PutUint32(ss, k)
		ii := bytes.Index(s, ss)
		if ii > 0 {
			indexes = append(indexes, ii)
		}
	}

	sort.Ints(indexes)
	for _, v := range indexes {
		saveKey := dict[binary.LittleEndian.Uint32(s[v:])]
		saveValue := binary.LittleEndian.Uint32(s[v+4:])
		if keysOnly {
			fmt.Printf("%s\n", saveKey)
		} else {
			fmt.Printf("0x%x\t\t%s: %d\n", v, saveKey, saveValue)
		}
	}

	// encrypt, need to add md5
	//Decrypt(key, s)
	//os.WriteFile("/tmp/encoded", s, 0644)
	//
	//sold2, err := os.ReadFile("/tmp/decoded")
	//if err != nil {
	//	panic(err)
	//}
}
