package size

type ESize byte

// tpp::gm::impl::BeginScriptVars

//go:generate stringer -type=ESize
const (
	Bool   ESize = 0
	Int32  ESize = 1
	UInt32 ESize = 2
	Float  ESize = 3
	Int8   ESize = 4
	UInt8  ESize = 5
	Int16  ESize = 6
	UInt16 ESize = 7
)
