package functions

import "slices"

type RoleMapStruct struct {
	List map[string]RoleInfo
}

type RoleInfo struct {
	PushLimit    int
	AccountLimit int

	RoleScore       int
	AllowEndpoints  []string
	AllowWebsocket  bool
	AllowTranslator bool
	AllowTTS        bool
	TimelineLatency int
}

var ServiceRoleMap = RoleMapStruct{
	List: map[string]RoleInfo{
		"node": {
			PushLimit:      0,
			AccountLimit:   0,
			RoleScore:      0,
			AllowWebsocket: true,
		},
	},
}

func (roleMap *RoleMapStruct) GetQuota(role, _type string) int {
	return 0
}

func (roleMap *RoleMapStruct) AllowEndpoint(role, endpoint string) bool {
	if item, ok := roleMap.List[role]; ok {
		return slices.Contains(item.AllowEndpoints, endpoint)
	}
	return false
}

func (roleMap *RoleMapStruct) AllowWebsocket(role string) bool {
	if item, ok := roleMap.List[role]; ok {
		return item.AllowWebsocket
	}
	return false
}

func (roleMap *RoleMapStruct) ValidRole(role string) bool {
	_, ok := roleMap.List[role]
	return ok
}

func (roleMap *RoleMapStruct) GetLatency(role string) int {
	if item, ok := roleMap.List[role]; ok {
		return item.TimelineLatency
	}
	return 10000
}
