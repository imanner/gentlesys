package userinfo

import (
	"gentlesys/store"
	"os"

	"github.com/golang/protobuf/proto"
)

//用户发布的帖子，也是用这个结构体来保存的。所以当时这个结构体只用来保存评论，名称取的不太好。

//用户发布的话题相关的功能在此
//相关的功能在此
type Topic struct {
	//FilePath string
	Fd *os.File
	//不需要锁，因为只有用户自己才能操作自己的几率，不存在并发的可能性
}

//读当前的用户帖子块，每个块包含OnePageCommentNum条记录
func (c *Topic) ReadCurUserTopicBlock() (*store.UserTopics, bool) {
	index, ok := store.GetCurUsedId(c.Fd)

	if !ok || index >= store.MaxMetaMcSize {
		return nil, false
	}
	aMeta := &store.McDataIndexHead{}
	if !store.ReadMetaData(c.Fd, int(index), aMeta) {
		return nil, false
	}
	m2 := &store.UserTopics{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !store.ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return m2, true
}

//增加一条发帖，返回最后评论页面index
func (c *Topic) AddOneUserTopic(data *store.UserTopicData) (bool, int) {

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
	if len(srcData.Usertopicdata) >= store.OnePageCommentNum {
		isCurMcFull = true
		//fmt.Printf("full len %d ", len(srcData.Commentdata))
	}
	//fmt.Printf("update len %d ", len(srcData.Commentdata))
	return store.UpdateTailBlockToStore(c.Fd, mdata, isCurMcFull)

}

//获取一页帖子
func (c *Topic) GetOnePageTopics(pageNums int) (*[]*store.UserTopicData, bool) {
	index, ok := store.GetCurUsedId(c.Fd)

	if !ok || pageNums > int(index) || index >= store.MaxMetaMcSize {
		return nil, false
	}

	aMeta := &store.McDataIndexHead{}
	if !store.ReadMetaData(c.Fd, pageNums, aMeta) {
		return nil, false
	}
	m2 := &store.UserTopics{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !store.ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return &m2.Usertopicdata, true
}
