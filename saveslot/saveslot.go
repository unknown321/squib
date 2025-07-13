package saveslot

type Slot struct {
	Name string
	Size int
	ID   int
}

// TppVarInit.lua, e.InitializeOnStatingMainFrame

var GameData = []Slot{
	0: {
		Name: "GLOBAL",
		Size: 14 * 1024,
		ID:   0,
	},
	1: {
		Name: "CHECK_POINT",
		Size: 65 * 1024,
		ID:   1,
	},
	2: {
		Name: "MISSION_START",
		Size: 10 * 1024,
		ID:   5,
	},
	3: {
		Name: "CHECK_POINT_RESTARTABLE",
		Size: 10 * 1024,
		ID:   6,
	},
	4: {
		Name: "RETRY",
		Size: 11 * 1024,
		ID:   2,
	},
	5: {
		Name: "MB_MANAGEMENT",
		Size: 80.5*1024 + 2688,
		ID:   3,
	},
	6: {
		Name: "QUEST",
		Size: 2 * 1024,
		ID:   4,
	},
}

var PersonalData = []Slot{
	0: {
		Name: "PERSONAL_SAVE",
		Size: 3 * 1024,
		ID:   11,
	},
}

var MgoData = []Slot{
	0: {
		Name: "MGO_SAVE",
		Size: 16 * 1024,
		ID:   13,
	},
}

var Slots = []Slot{
	7: {
		Name: "CONFIG",
		Size: 2 * 1024,
		ID:   7,
	},
	8: {
		Name: "SAVING", // total file size =  sum(global, check_point, retry, mb_management, quest, mission_start, checkp_restartable) + 92
		Size: 0,
		ID:   8,
	},
	9: {
		Name: "CONFIG_SAVE",
		Size: 2 * 1024,
		ID:   9,
	},

	11: {
		Name: "PERSONAL",
		Size: 3 * 1024,
		ID:   10,
	},
	//12: {
	//	Name: "MGO",
	//	Size: 16 * 1024,
	//	ID:   12,
	//},
}
