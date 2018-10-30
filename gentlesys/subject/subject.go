package subject

import (
	"fmt"
	"gentlesys/global"
)

type subNode struct {
	Name     string
	Href     string
	Desc     string
	SubNums  int
	TalkNums int
}

var subNodes *[]subNode = nil

func GetMaxSubjectId() int {
	return len(*subNodes)
}

func GetSubjectById(id int) *subNode {
	return &(*subNodes)[id]
}

func GetMainPageSubjectData() *[]subNode {
	if subNodes == nil {
		subNodes = getMainPageSubjectCfg()
	}
	return subNodes
}

func getMainPageSubjectCfg() *[]subNode {
	nums := global.GetIntFromCfg("subject::nums", 0)
	if nums > 0 {
		pageSubNodes := make([]subNode, nums)

		for i := 0; i < nums; i++ {
			name_ := fmt.Sprintf("subject::id.%d.name", i)
			desc_ := fmt.Sprintf("subject::id.%d.desc", i)

			pageSubNodes[i].Name = global.GetStringFromCfg(name_, "检查是否有名称")
			pageSubNodes[i].Href = fmt.Sprintf("/sjt%d", i)
			pageSubNodes[i].Desc = global.GetStringFromCfg(desc_, "检查是否有描述")
		}
		return &pageSubNodes
	}
	return nil
}
