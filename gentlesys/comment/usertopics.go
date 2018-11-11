package comment

import (
	"github.com/golang/protobuf/proto"
)

//用户发布的帖子，也是用这个结构体来保存的。所以当时这个结构体只用来保存评论，名称取的不太好。

//用户发布的话题相关的功能在此

//读当前的用户帖子块，每个块包含OnePageCommentNum条记录
func (c *Comment) ReadCurUserTopicBlock() (*UserTopics, bool) {
	index, ok := GetCurUsedId(c.Fd)

	if !ok || index >= MaxMetaMcSize {
		return nil, false
	}
	aMeta := &McDataIndexHead{}
	if !ReadMetaData(c.Fd, int(index), aMeta) {
		return nil, false
	}
	m2 := &UserTopics{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return m2, true
}

//增加一条发帖，返回最后评论页面index
func (c *Comment) AddOneUserTopic(data *UserTopicData) (bool, int) {

	srcData, ok := c.ReadCurUserTopicBlock()
	if !ok {
		return false, 0
	}

	srcData.Usertopicdata = append(srcData.Usertopicdata, data)
	mdata, err := proto.Marshal(srcData)
	if err != nil {
		panic(err)
	}
	var isCurMcFull bool = false
	if len(srcData.Usertopicdata) >= OnePageCommentNum {
		isCurMcFull = true
		//fmt.Printf("full len %d ", len(srcData.Commentdata))
	}
	//fmt.Printf("update len %d ", len(srcData.Commentdata))
	return UpdateTailBlockToStore(c.Fd, mdata, isCurMcFull)

}

//获取一页帖子
func (c *Comment) GetOnePageTopics(pageNums int) (*[]*UserTopicData, bool) {
	index, ok := GetCurUsedId(c.Fd)

	if !ok || pageNums > int(index) || index >= MaxMetaMcSize {
		return nil, false
	}

	aMeta := &McDataIndexHead{}
	if !ReadMetaData(c.Fd, pageNums, aMeta) {
		return nil, false
	}
	m2 := &UserTopics{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return &m2.Usertopicdata, true
}
