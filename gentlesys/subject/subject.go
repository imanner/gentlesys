package subject

import (
	"fmt"
	"gentlesys/global"
	"sync/atomic"
)

type subNode struct {
	Name          string
	Href          string
	Desc          string
	UniqueId      int   //在数据库中话题对于的唯一类型IsSubjectIdExist()
	CurTopicIndex int32 //主题中当前帖子的索引
}

//原子更新当前帖子的索引
func UpdateCurTopicIndex(subId int, value int) {
	atomic.StoreInt32(&((*subNodes)[subId].CurTopicIndex), int32(value))
}

//读写最好都是原子操作,这样可以避免并发读取错误
func GetCurTotalTopicNums(subId int) int {
	return int(atomic.LoadInt32(&(*subNodes)[subId].CurTopicIndex))
}

var subNodes *[]subNode = nil

//根据subNode的UniqueId快速反向定位到subNode.因为UniqueId可能不连续，故不能直接用subNodes的数组索引
var subNodesIndexMap map[int]*subNode

func init() {
	//预先做下初始化，避免一些Get空指针
	subNodes = getMainPageSubjectCfg()

	//初始化话题类型列表
	initTopicTypeList()

}

//检查主题id是否存在
func IsSubjectIdExist(id int) bool {
	if _, ok := subNodesIndexMap[id]; ok {
		return true
	}
	return false
}

//通过主题id得到主题结构体
func GetSubjectById(id int) *subNode {
	return subNodesIndexMap[id]
}

//基本是个包装函数，做了判空
func GetMainPageSubjectData() *[]subNode {
	if subNodes == nil {
		subNodes = getMainPageSubjectCfg()
	}
	return subNodes
}

//从配置文件中读取主题相关的配置，注意不要每次掉该函数，而用封装函数，不必每次读文件
func getMainPageSubjectCfg() *[]subNode {
	nums := global.GetIntFromCfg("subject::nums", 0)
	if nums > 0 {
		pageSubNodes := make([]subNode, nums)
		subNodesIndexMap = make(map[int]*subNode, nums)

		for i := 0; i < nums; i++ {
			name_ := fmt.Sprintf("subject::id.%d.name", i)
			desc_ := fmt.Sprintf("subject::id.%d.desc", i)
			uniqueId_ := fmt.Sprintf("subject::id.%d.uniqueId", i)
			pageSubNodes[i].Name = global.GetStringFromCfg(name_, "检查是否有名称")
			pageSubNodes[i].Desc = global.GetStringFromCfg(desc_, "检查是否有描述")
			pageSubNodes[i].UniqueId = global.GetIntFromCfg(uniqueId_, -1)
			if pageSubNodes[i].UniqueId == -1 {
				panic("error pageSubNodes[i].UniqueId -1\n")
			}
			pageSubNodes[i].Href = fmt.Sprintf("/subject%d", pageSubNodes[i].UniqueId)
			subNodesIndexMap[pageSubNodes[i].UniqueId] = &pageSubNodes[i]
		}
		return &pageSubNodes
	}
	return nil
}

//有关话题类型的操作。
var topicTyleList []string = nil

func GetTopicTyleList() *[]string {
	if topicTyleList != nil {
		return &topicTyleList
	}
	return nil
}

func GetTopicTyleById(id int) string {
	if topicTyleList != nil && id >= 0 && id < len(topicTyleList) {
		return topicTyleList[id]
	}
	return ""
}

func initTopicTypeList() {
	nums := global.GetIntFromCfg("topic::nums", 0)
	if nums > 0 {
		topicTyleList = make([]string, nums)
		for i := 0; i < nums; i++ {
			name_ := fmt.Sprintf("topic::id.%d.name", i)
			topicTyleList[i] = global.GetStringFromCfg(name_, "")
		}
	}
}
