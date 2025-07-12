package main

import (
	"crypto/md5"
	"embed"
	_ "embed"
	"encoding/binary"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"squib/dictionary"
	"squib/save"
	"strings"
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

//go:embed savevars.txt
var f embed.FS

func main() {
	var keysOnly bool
	var key string
	var encode bool
	flag.CommandLine.SetOutput(os.Stdout)
	flag.BoolVar(&keysOnly, "keysOnly", false, "print only keys without address and values")
	flag.StringVar(&key, "key", "", "decryption key, see GAME_SAVE_FILE_NAME in TppDefine.lua")
	flag.BoolVar(&encode, "encode", false, "encode file")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Squib: MGSV:TPP save decoder")
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

	var filename string
	if len(flag.Args()) < 1 {
		slog.Error("please provide filename")
		os.Exit(1)
	}

	filename = flag.Args()[0]
	saveData, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("open file", "error", err.Error(), "filename", filename)
		os.Exit(1)
	}

	if err = dictionary.Init(&f); err != nil {
		slog.Error("open dict", "error", err.Error())
		os.Exit(1)
	}

	keys := []string{"TPP_GAME_DATA", "TPP_CONFIG_DATA", "PERSONAL_DATA", "MGO_GAME_DATA"}
	if key == "" {
		for _, k := range keys {
			if strings.Contains(filename, k) {
				key = k
				break
			}
		}
	}

	if key == "" {
		slog.Error("decryption key not provided")
		os.Exit(1)
	}

	if encode {
		slog.Info("encoding")
	} else {
		slog.Info("decoding")
	}

	Decrypt(key, saveData)

	//os.Exit(1)

	out := filename + "_decoded"
	if encode {
		out = strings.TrimSuffix(filename, "_decoded")
	}

	if err = os.WriteFile(out, saveData, 0644); err != nil {
		slog.Error("save", "error", err.Error(), "filename", out)
		os.Exit(1)
	}

	slog.Info("saved", "output file", out)

	s := &save.Save{}
	if err = s.Parse(saveData, dictionary.Dict); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if encode {
		os.Exit(0)
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
