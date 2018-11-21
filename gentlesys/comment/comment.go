package comment

//与评论相关的在此。这个文件

import (
	//	"fmt"
	"gentlesys/store"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"
)

var commentHandlerManager sync.Map

func init() {

}

//所有获取Comment都必须通过该接口，防止异步读写文件冲突，commentHandlerManager得注意清理
func GetCommentHandlerByPath(filePath string) *Comment {
	obj, _ := commentHandlerManager.LoadOrStore(filePath, new(Comment))
	return obj.(*Comment)
}

//map里面存放的是Comment地址，就算删除地址，不会影响Comment本身的存在。防止commentHandlerManager过大爆炸
func DelCommentHandlerByPath(filePath string) {
	commentHandlerManager.Delete(filePath)
}

//获取帖子的评论数量,不准确，只能在页范围内，不影响索引
func GetCommentNums(filePath string) int {

	if store.CheckExists(filePath) {
		fd, _ := os.OpenFile(filePath, os.O_RDWR, 0644)
		defer fd.Close()
		if nums, ok := store.GetCurUsedId(fd); ok {
			return int(nums) + 1
		}
	}
	return 0
}

//相关的功能在此
type Comment struct {
	FilePath string
	Fd       *os.File
	Mutex    sync.RWMutex //用于保护结构体的锁，保护文件的读写，防止异步写
}

//读取指定块的评论内容
func (c *Comment) ReadCommentBlockByIndex(blockNums int) (*store.CommentStory, uint32, bool) {

	if buf, ok := store.GetOnePageContent(c.Fd, blockNums); ok && buf != nil {
		m2 := &store.CommentStory{Commentdata: nil}
		proto.Unmarshal(*buf, m2) //反序列化
		return m2, uint32(blockNums), true
	} else if ok {
		//也是成功，不过块暂时是空的，新建块后第一次操作
		return nil, 0, true
	} else {
		return nil, 0, false
	}
}

//读当前的评论块，每个块包含OnePageCommentNum条记录
func (c *Comment) ReadCurCommentBlock() (*store.CommentStory, uint32, bool) {
	return c.ReadCommentBlockByIndex(-1)
}

//禁用一条评论。
func (c *Comment) DisableOneComment(pageNums int, id int) (bool, int) {
	if srcData, _, ok := c.ReadCommentBlockByIndex(pageNums); ok && srcData != nil {
		for _, v := range srcData.Commentdata {
			if int(*v.Id) == id {
				//找到并屏蔽
				if v.IsDel != nil && *v.IsDel {
					//已经是禁用的了，直接返回
					return false, 1
				}
				v.IsDel = proto.Bool(true)
				mdata, err := proto.Marshal(srcData)
				if err != nil {
					panic(err)
				}
				return store.UpdateBlockToStore(c.Fd, mdata, pageNums)
			}
		}
	}
	return false, 0
}

//增加一条评论，返回最后评论页面index
func (c *Comment) AddOneComment(data *store.CommentData) (bool, int) {

	var id int32
	srcData, curBlockNums, ok := c.ReadCurCommentBlock()
	if !ok {
		return false, 0
	} else if srcData == nil {
		//块的第一个元素
		srcData = &store.CommentStory{}
		id += int32(curBlockNums) * store.OnePageCommentNum
		data.Id = proto.Int32(id)
		srcData.Commentdata = []*store.CommentData{data}
	} else {
		id = *(srcData.Commentdata[len(srcData.Commentdata)-1].Id) + 1
		data.Id = proto.Int32(id)
		srcData.Commentdata = append(srcData.Commentdata, data)
	}

	mdata, err := proto.Marshal(srcData)
	if err != nil {
		panic(err)
	}
	var isCurMcFull bool = false
	//如果达到OnePageCommentNum，表示当前满了一页，要开始更新索引到下一页
	if len(srcData.Commentdata) >= store.OnePageCommentNum {
		isCurMcFull = true
		//fmt.Printf("full len %d ", len(srcData.Commentdata))
	}
	//fmt.Printf("update len %d ", len(srcData.Commentdata))
	return store.UpdateTailBlockToStore(c.Fd, mdata, isCurMcFull)

}

//获取一页评论
func (c *Comment) GetOnePageComments(pageNums int) (*[]*store.CommentData, bool) {

	if buf, ok := store.GetOnePageContent(c.Fd, pageNums); ok && buf != nil {
		m2 := &store.CommentStory{}
		proto.Unmarshal(*buf, m2) //反序列化
		return &m2.Commentdata, true
	} else if ok {
		return nil, true
	} else {
		return nil, false
	}
}
