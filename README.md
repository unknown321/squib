TNT: MGSV:TPP save decoder
====

```
Usage of ./tnt:
	./tnt [OPTION] FILE

Options:
  -keysOnly
    	print only keys without address and values

Provide more keys by adding them to dict.txt in the same directory as executable.
Decoded file is created in the same directory with '_decoded' suffix.
```

#### Example

```shell
$ ./tnt /tmp/PERSONAL_DATA1 
0xb8            eulaVersion: 36
0xd0            infoId: 65535
0xd8            infoIdForMGO: 65535
0xe0            inquiryId: 42
0xe8            avatarFaceRaceIndex: 65535
0xf0            avatarFaceTypeIndex: 65535
...
```
