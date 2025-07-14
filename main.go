package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/unknown321/squib/dictionary"
	"github.com/unknown321/squib/save"
	"log/slog"
	"os"
	"strings"
	"time"
)

//go:embed savevars.txt
var f embed.FS

func main() {
	var key string
	var encode bool
	var noBackup bool
	var js bool
	flag.CommandLine.SetOutput(os.Stdout)
	flag.StringVar(&key, "key", "", "decryption key, see GAME_SAVE_FILE_NAME in TppDefine.lua")
	flag.BoolVar(&encode, "encode", false, "encode file")
	flag.BoolVar(&js, "json", false, "dump to json")
	flag.BoolVar(&noBackup, "noBackup", false, "do not create backup while encoding")
	flag.Parse()

	flag.Usage = func() {
		fmt.Println("squib: MGSV:TPP save decoder")
		fmt.Println()
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("\t%s [OPTION] FILE\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Provide more keys by adding them to dict.txt in the same directory as executable.")
		fmt.Println("Decoded file is created in the same directory with '_decoded' suffix.")
		fmt.Println("Encoded file is created in the same directory without '_decoded' suffix. Encoding will overwrite existing files.")
	}

	if len(os.Args) < 2 {
		flag.Usage()
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

	// TPP_GRAPHICS_CONFIG is plain json
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
		save.Decode(key, saveData)
		out := strings.TrimSuffix(filename, "_decoded")
		if !noBackup {
			if _, err = os.Stat(out); err == nil {
				backupName := fmt.Sprintf("%s.%d.backup", out, time.Now().Unix())
				if err = os.Rename(out, backupName); err != nil {
					slog.Error("Cannot create backup", "original file", out, "error", err.Error())
					os.Exit(1)
				}
				slog.Info("Created backup", "filename", backupName)
			}
		}
		if err = os.WriteFile(out, saveData, 0644); err != nil {
			slog.Error("encode", "error", err.Error(), "filename", out)
			os.Exit(1)
		}
		slog.Info("encoded", "output file", out)
		os.Exit(0)
	}

	slog.Info("decoding")
	save.Decode(key, saveData)

	out := filename + "_decoded"
	if err = os.WriteFile(out, saveData, 0644); err != nil {
		slog.Error("save", "error", err.Error(), "filename", out)
		os.Exit(1)
	}

	slog.Info("decoded", "output file", out)

	s := &save.Save{}
	if err = s.Parse(saveData, dictionary.Dict); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if js && !encode {
		res, err := json.MarshalIndent(s, "", "    ")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		if err = os.WriteFile(filename+".json", res, 0644); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
}
