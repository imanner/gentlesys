package userinfo

import (
	"gentlesys/store"
	"os"

	"github.com/golang/protobuf/proto"
)

//用户发布的话题记录的相关功能在此
type Topic struct {
	//FilePath string
	Fd *os.File
	//不需要锁，因为只有用户自己才能操作自己的几率，不存在并发的可能性
}

//增加一条发帖，返回最后评论页面index
func (c *Topic) AddOneUserTopic(data *store.UserTopicData) (bool, int) {

	srcData, ok := c.ReadCurUserTopicBlock()
	if !ok {
		return false, 0
	} else if srcData == nil {
		srcData = &store.UserTopics{}
		srcData.Usertopicdata = []*store.UserTopicData{data}
	} else {
		srcData.Usertopicdata = append(srcData.Usertopicdata, data)
	}

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

//读当前的用户帖子块，每个块包含OnePageCommentNum条记录
func (c *Topic) ReadCurUserTopicBlock() (*store.UserTopics, bool) {
	if buf, ok := store.GetOnePageContent(c.Fd, -1); ok && buf != nil {
		m2 := &store.UserTopics{}
		proto.Unmarshal(*buf, m2) //反序列化
		return m2, true
	} else if ok {
		return nil, true
	} else {
		return nil, false
	}
}

//获取一页帖子
func (c *Topic) GetOnePageTopics(pageNums int) (*[]*store.UserTopicData, bool) {

	if buf, ok := store.GetOnePageContent(c.Fd, pageNums); ok && buf != nil {
		m2 := &store.UserTopics{}
		proto.Unmarshal(*buf, m2) //反序列化
		return &m2.Usertopicdata, true
	} else if ok {
		return nil, true
	} else {
		return nil, false
	}

}
