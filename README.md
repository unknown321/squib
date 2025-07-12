squib: MGSV:TPP save decoder
====

```
squib: MGSV:TPP save decoder

Usage of ./squib-linux-amd64:
        ./squib-linux-amd64 [OPTION] FILE

Options:
  -encode
        encode file
  -key string
        decryption key, see GAME_SAVE_FILE_NAME in TppDefine.lua

Provide more keys by adding them to dict.txt in the same directory as executable.
Decoded file is created in the same directory with '_decoded' suffix.
Encoded file is created in the same directory without '_decoded' suffix. Encoding will overwrite existing files.
```

### Examples

```shell
$ ./squib TPP_GAME_DATA
2025/07/12 23:14:59 INFO decoding
2025/07/12 23:14:59 INFO decoded "output file"=TPP_GAME_DATA_decoded
SVAR, Type: CONFIG, version: 0x0001006a
Group 0, 50 entries
	fobPickup (1): [1 (0x00000001)]
	dominatedCpFlagsAfgh (16): [255, 255, 254, 125, 59, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
	dominatedCpFlagsMafr (16): [255, 63, 255, 254, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
	totalAlertCount (1): [776 (0x00000308)]
	playerOgrePointOnStartMission (1): [180 (0x000000b4)]
...
```

#### Custom key

```shell
$ ./squib-linux-amd64 -key DYNAMITE_PERSONAL_DATA DYNAMITE_PERSONAL_DATA
2025/07/12 23:16:36 INFO decoding
2025/07/12 23:16:36 INFO decoded "output file"=DYNAMITE_PERSONAL_DATA_decoded
SVAR, Type: CONFIG, version: 0x00010063
Group 0, 61 entries
	eulaVersion (1): [6 (0x00000006)]
```

#### Encoding

```shell
$ ./squib-linux-amd64 TPP_GAME_DATA > /dev/null
2025/07/12 23:18:22 INFO decoding
2025/07/12 23:18:22 INFO decoded "output file"=TPP_GAME_DATA_decoded

$ mv TPP_GAME_DATA TPP_GAME_DATA_orig

$ ./squib-linux-amd64 -encode TPP_GAME_DATA_decoded
2025/07/12 23:18:49 INFO encoding
2025/07/12 23:18:49 INFO encoded "output file"=TPP_GAME_DATA

$ md5sum TPP_GAME_DATA TPP_GAME_DATA_orig
77467d5c188c857bb57d48371e7118b9  TPP_GAME_DATA
77467d5c188c857bb57d48371e7118b9  TPP_GAME_DATA_orig
```