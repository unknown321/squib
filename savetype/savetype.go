package savetype

type ESaveType uint

// tpp::gm::impl::BeginScriptVars

//go:generate stringer -type=ESaveType
const (
	GAME ESaveType = iota
	CONFIG
	PERSONAL
	GRAPHICS
	MGO
)
