package groupbit

type EGroupBit uint16

// tpp::gm::impl::BeginScriptVars

//go:generate stringer -type=EGroupBit
const (
	Vars  EGroupBit = 1
	CVars EGroupBit = 2
	GVars EGroupBit = 4
	SVars EGroupBit = 8
	All   EGroupBit = 0xf
)
