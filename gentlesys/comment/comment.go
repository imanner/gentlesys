package comment

import (
	"fmt"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"
)

var commentHandlerManager sync.Map

func init() {

}

//所有获取Comment都必须通过该接口，防止异步读写文件冲突，但是commentHandlerManager的清理是个问题
func GetCommentHandlerByPath(filePath string) *Comment {
	obj, _ := commentHandlerManager.LoadOrStore(filePath, new(Comment))
	return obj.(*Comment)
}

//map里面存放的是Comment地址，就算删除地址，不会影响Comment本身的存在。防止commentHandlerManager过大爆炸
func DelCommentHandlerByPath(filePath string) {
	commentHandlerManager.Delete(filePath)
}

//相关的功能在此
type Comment struct {
	FilePath string
	Fd       *os.File
	Mutex    sync.RWMutex //用于保护结构体的锁，保护文件的读写，防止异步写
}

//读当前的评论块，每个块包含OnePageCommentNum条记录
func (c *Comment) ReadCurCommentBlock() (*CommentStory, bool) {
	index, ok := GetCurUsedId(c.Fd)

	if !ok || index >= MaxMetaMcSize {
		return nil, false
	}
	aMeta := &McDataIndexHead{}
	if !ReadMetaData(c.Fd, int(index), aMeta) {
		return nil, false
	}
	m2 := &CommentStory{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return m2, true
}

func (c *Comment) InitMcData() {
	cur_offset, _ := c.Fd.Seek(0, os.SEEK_CUR)
	metaOff := Int32Bytes + MaxMetaMcSize*McDataIndexHeadSize
	content := make([]byte, metaOff)

	c.Fd.WriteAt(content, cur_offset)
}

//增加一条评论
func (c *Comment) AddOneComment(data *CommentData) bool {

	srcData, ok := c.ReadCurCommentBlock()
	if !ok {
		return false
	}

	srcData.Commentdata = append(srcData.Commentdata, data)
	mdata, err := proto.Marshal(srcData)
	if err != nil {
		panic(err)
	}
	var isCurMcFull bool = false
	if len(srcData.Commentdata) >= OnePageCommentNum {
		isCurMcFull = true
		fmt.Printf("full len %d ", len(srcData.Commentdata))
	}
	fmt.Printf("update len %d ", len(srcData.Commentdata))
	return UpdateTailBlockToStore(c.Fd, mdata, isCurMcFull)

}

//获取一页评论
func (c *Comment) GetOnePageComments(pageNums int) (*[]*CommentData, bool) {
	index, ok := GetCurUsedId(c.Fd)

	if !ok || pageNums > int(index) || index >= MaxMetaMcSize {
		return nil, false
	}

	aMeta := &McDataIndexHead{}
	if !ReadMetaData(c.Fd, pageNums, aMeta) {
		return nil, false
	}
	m2 := &CommentStory{}
	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !ReadOneBlockMemory(c.Fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}

		proto.Unmarshal(buf, m2) //反序列化
	}
	return &m2.Commentdata, true
}
