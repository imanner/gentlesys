package userinfo

import (
	"fmt"
	"gentlesys/store"
	"os"

	"github.com/golang/protobuf/proto"
)

//用户评论管理。也是使用.proto结构进行管理
//用户发布的话题相关的功能在此
//相关的功能在此
type Comment struct {
	Fd *os.File
	//不需要锁，因为只有用户自己才能操作自己的几率，不存在并发的可能性
}

//读取指定块的评论内容
func (c *Comment) ReadCommentBlockByIndex(blockNums int) (*store.UserComments, bool) {
	aMeta := &store.McDataIndexHead{}
	if !store.ReadMetaData(c.Fd, blockNums, aMeta) {
		return nil, false
	}
	m2 := &store.UserComments{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !store.ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return m2, true
}

//读当前的用户帖子块，每个块包含OnePageCommentNum条记录
func (c *Comment) ReadCurUserCommentBlock() (*store.UserComments, bool) {
	index, ok := store.GetCurUsedId(c.Fd)

	if !ok || index >= store.MaxMetaMcSize {
		return nil, false
	}
	return c.ReadCommentBlockByIndex(int(index))

}

//禁用一条评论。
func (c *Comment) DisableOneComment(pageNums int, id int) (bool, int) {
	if srcData, ok := c.ReadCommentBlockByIndex(pageNums); ok {
		for _, v := range srcData.Usercommentdata {
			if int(*v.Commentdata.Id) == id {
				//找到并屏蔽
				if v.Commentdata.IsDel != nil && *v.Commentdata.IsDel {
					//已经是禁用的
					return false, 1
				}
				v.Commentdata.IsDel = proto.Bool(true)
				mdata, err := proto.Marshal(srcData)
				if err != nil {
					panic(err)
				}
				fmt.Printf("第%d楼用户中心已经删除回复\n", id)

				return store.UpdateBlockToStore(c.Fd, mdata, pageNums)
			}
		}
	}
	return false, 0
}

//增加一条发帖，返回最后评论页面index
func (c *Comment) AddOneUserComment(data *store.UserCommentData) (bool, int) {

	srcData, ok := c.ReadCurUserCommentBlock()
	if !ok {
		return false, 0
	}

	srcData.Usercommentdata = append(srcData.Usercommentdata, data)
	mdata, err := proto.Marshal(srcData)
	if err != nil {
		panic(err)
	}
	var isCurMcFull bool = false
	if len(srcData.Usercommentdata) >= store.OnePageCommentNum {
		isCurMcFull = true
	}
	//fmt.Printf("update len %d ", len(srcData.Commentdata))
	return store.UpdateTailBlockToStore(c.Fd, mdata, isCurMcFull)

}

//获取一页帖子
func (c *Comment) GetOnePageComment(pageNums int) (*[]*store.UserCommentData, bool) {
	index, ok := store.GetCurUsedId(c.Fd)

	if !ok || pageNums > int(index) || index >= store.MaxMetaMcSize {
		return nil, false
	}

	aMeta := &store.McDataIndexHead{}
	if !store.ReadMetaData(c.Fd, pageNums, aMeta) {
		return nil, false
	}
	m2 := &store.UserComments{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !store.ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return &m2.Usercommentdata, true
}
