package savevarcategory

type ESaveVarCategory uint8

// tpp::gm::impl::BeginScriptVars

//go:generate stringer -type=ESaveVarCategory
const (
	None               ESaveVarCategory = 0
	Config             ESaveVarCategory = 1
	MissionRestartable ESaveVarCategory = 2
	GameGlobal         ESaveVarCategory = 4
	Mission            ESaveVarCategory = 8
	Retry              ESaveVarCategory = 0x10
	MbManagement       ESaveVarCategory = 0x20
	Quest              ESaveVarCategory = 0x40
	Personal           ESaveVarCategory = 0x80
)
