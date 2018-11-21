package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/astaxie/beego/logs"
)

/*我需要学硬盘一样，在头部编写一个零号扇区，这个扇区管制这所以的数据。但是最多可以有1024个结构体
即comment.mc的格式为，头部有一个1024的结构体体，控制着所有存储块。这是一个通用的存储结构
*/

type Metadata struct {
	UsedId int32 //下一个可以使用的index 从0-511
	Metas  [MaxMetaMcSize]McDataIndexHead
}

const MCDATAOFF = 4

const Int32Bytes = 4

const MaxMetaMcSize = 512

//如果该值越大，那么每次更新评论时的负载就高，我认为20-50比较合适。
const OnePageCommentNum = 20 //一页最多多少条评论。目前20*512=10240，最多支持10240条评论。超过则无法继续评论

const ErrMetaMcSize = 512

const McDataIndexHeadSize = 8 //McDataIndexHead结构体的长度

type McDataIndexHead struct {
	Start  uint32 /*开始偏移量*/
	Length uint32 /*一块data的长度。OnePageCommentNum条评论共用该空间，目前看是够了。*/
}

//获取对象的数量,不准确，只能在页范围内，不影响索引
func GetObjPageNums(filePath string) int {

	if CheckExists(filePath) {
		fd, _ := os.OpenFile(filePath, os.O_RDWR, 0644)
		defer fd.Close()
		if nums, ok := GetCurUsedId(fd); ok {
			return int(nums) + 1
		}
	}
	return 0
}

func InitMcData(fd *os.File) {
	cur_offset, _ := fd.Seek(0, os.SEEK_CUR)
	metaOff := Int32Bytes + MaxMetaMcSize*McDataIndexHeadSize
	content := make([]byte, metaOff)

	fd.WriteAt(content, cur_offset)
}

//返回当前的元数据
func GetCurUsedId(fd *os.File) (uint32, bool) {
	used := make([]byte, Int32Bytes)
	n, err := fd.ReadAt(used, 0)
	if n >= Int32Bytes {
		return binary.LittleEndian.Uint32(used), true
	} else if err == io.EOF {
		return 0, true
	}
	return ErrMetaMcSize, false
}

/*读取指定元数据*/
func ReadMetaData(fd *os.File, index int, mcHead *McDataIndexHead) bool {
	if index >= MaxMetaMcSize {
		return false
	}

	/*在调用外部都已经检查过
	useId, ok := GetCurUsedId(fd)
	if !ok || useId < uint32(index) {
		return false
	}*/

	start := Int32Bytes + index*McDataIndexHeadSize

	buf := make([]byte, McDataIndexHeadSize)

	n1, _ := fd.ReadAt(buf, int64(start))
	if n1 >= McDataIndexHeadSize {
		mcHead.Start = binary.LittleEndian.Uint32(buf[:Int32Bytes])
		mcHead.Length = binary.LittleEndian.Uint32(buf[Int32Bytes:])
		return true
	}
	return false
}

/*追加一块元数据*/
func AppendMetaData(fd *os.File, mcHead *McDataIndexHead) bool {

	useId, ok := GetCurUsedId(fd)
	if !ok {
		return false
	}

	if useId >= MaxMetaMcSize {
		return false //满了
	} else {
		start := Int32Bytes + useId*McDataIndexHeadSize

		//fd.Seek(int64(start), 0)
		buf := make([]byte, McDataIndexHeadSize)

		binary.LittleEndian.PutUint32(buf, mcHead.Start)
		binary.LittleEndian.PutUint32(buf[Int32Bytes:], mcHead.Length)

		//fmt.Fprintf(fd, "%x", buf)
		fd.WriteAt(buf, int64(start))
	}

	return true
}

/*更新指定的元数据*/
func UpdateMetaData(fd *os.File, index int, mcHead *McDataIndexHead) bool {
	if index >= MaxMetaMcSize {
		return false
	}

	start := Int32Bytes + index*McDataIndexHeadSize

	//fd.Seek(int64(start), 0)
	buf := make([]byte, McDataIndexHeadSize)

	binary.LittleEndian.PutUint32(buf, mcHead.Start)
	binary.LittleEndian.PutUint32(buf[Int32Bytes:], mcHead.Length)

	//fmt.Fprintf(fd, "%x", buf)
	fd.WriteAt(buf, int64(start))

	return true
}

func CheckExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

//追加一块数据到存储文件，并更新元数据块
/*可以不使用该函数，因为UpdateTailBlockToStore更新used会自动在一页满时更新到下一页
func AppendOneBlockToStore(fd *os.File, content []byte) bool {

	//一定要超过元数据，元数据是不能被写的
	//定位到文件最后
	cur_offset, _ := fd.Seek(0, os.SEEK_END)
	metaOff := int64(Int32Bytes + MaxMetaMcSize*McDataIndexHeadSize)
	if cur_offset < metaOff+1 {
		cur_offset = metaOff + 1
	}

	fd.WriteAt(content, cur_offset)

	//追加元数据
	aMeta := &McDataIndexHead{}
	aMeta.Start = int32(cur_offset)
	aMeta.Length = int32(len(content))
	return AppendMetaData(fd, aMeta)
}
*/
/*基本流程。当需要添加一条评论时，读取尾部最后一个数据块的内容。
1 如果尾部没有满，则加上该评论后，走尾部更新函数
2 如果尾部的已经满了，则更新头部索引，使用下一块
*/

//更新当前尾部的一块数据，前面的不能更新，因为写是追加的,故只能更新尾部的。
func UpdateTailBlockToStore(fd *os.File, content []byte, isCurMcFill bool) (bool, int) {

	//走到该函数，说明尾部是没有满的
	//读取当前使用块
	useId, ok := GetCurUsedId(fd)
	if !ok {
		return false, 0
	}

	aMeta := &McDataIndexHead{}
	ReadMetaData(fd, int(useId), aMeta)
	cur_offset := int64(aMeta.Start)
	/*一定要超过元数据，元数据是不能被写的*/
	metaOff := int64(Int32Bytes + MaxMetaMcSize*McDataIndexHeadSize)
	if cur_offset < metaOff+1 {
		cur_offset = metaOff + 1
	}

	fd.WriteAt(content, cur_offset)

	//更新元数据
	aMeta.Start = uint32(cur_offset)
	aMeta.Length = uint32(len(content))

	ok = UpdateMetaData(fd, int(useId), aMeta)
	if !ok {
		return false, 0
	}

	curLastPage := int(useId)
	if isCurMcFill {
		//如果满了，则更新useId
		useId++
		if useId < MaxMetaMcSize {
			used := make([]byte, Int32Bytes)
			binary.LittleEndian.PutUint32(used, uint32(useId))
			fd.WriteAt(used, 0)
			//初始化新块的start,否则新块的start会从0开始。
			aMeta.Start += uint32(aMeta.Length)
			aMeta.Length = 0
			UpdateMetaData(fd, int(useId), aMeta)
		} else {
			logs.Error(fmt.Sprintf("%s 所有评论均已满", fd.Name()))
		}

	}
	return true, curLastPage
}

/*读取一块记忆体*/
func ReadOneBlockMemory(fd *os.File, buf []byte, start int64, length int) bool {
	var count int
	var err error
	var index int
	var off int64

	for count < length {
		index, err = fd.ReadAt(buf, start+off)
		count += index
		off += int64(index)
		if err != nil {
			return false
		}
	}
	return true
}

//修改一块区域，目前只修改一个评论的是否屏蔽位，不能随意修改最后一个块前面的块，因为大小已经定了。如果要修改
//只能减小块空间，千万不能增大块空间，否则会覆盖后面的数据
func UpdateBlockToStore(fd *os.File, content []byte, blockNum int) (bool, int) {
	//读取当前使用块
	useId, ok := GetCurUsedId(fd)
	if !ok || blockNum > int(useId) {
		return false, 0
	}

	aMeta := &McDataIndexHead{}
	ReadMetaData(fd, blockNum, aMeta)
	if len(content) > int(aMeta.Length) {
		//新内容不能超过旧现有内容
		return false, 1
	}
	cur_offset := int64(aMeta.Start)
	/*一定要超过元数据，元数据是不能被写的*/
	metaOff := int64(Int32Bytes + MaxMetaMcSize*McDataIndexHeadSize)
	if cur_offset < metaOff+1 {
		cur_offset = metaOff + 1
	}

	fd.WriteAt(content, cur_offset)

	//更新元数据
	aMeta.Start = uint32(cur_offset)
	aMeta.Length = uint32(len(content))

	ok = UpdateMetaData(fd, blockNum, aMeta)
	if !ok {
		return false, 0
	}
	return true, 0
}

//获取块中一页的内容,如果pageNums=-1，则获取最新一页，忽略pageNums
func GetOnePageContent(fd *os.File, pageNums int) (*[]byte, bool) {
	index, ok := GetCurUsedId(fd)

	if !ok || pageNums > int(index) || index >= MaxMetaMcSize {
		return nil, false
	}
	//pageNums=-1表示获取最新一页
	if pageNums == -1 {
		pageNums = int(index)
	}

	aMeta := &McDataIndexHead{}
	if !ReadMetaData(fd, pageNums, aMeta) {
		return nil, false
	}

	if aMeta.Length > 0 {
		buf := make([]byte, aMeta.Length)
		if !ReadOneBlockMemory(fd, buf, int64(aMeta.Start), int(aMeta.Length)) {
			return nil, false
		}
		return &buf, true
	} else {
		return nil, true
	}
}
