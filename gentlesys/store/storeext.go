package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

/*我需要学硬盘一样，在头部编写一个零号扇区，这个扇区管制这所以的数据。
每个存储体包含OneConPageNum个页面，每个页面包含OnePageObjNum个对象
做一个可以自动无限扩展的存储结构。目前该函数不是线程安全的。并发时，需要外部进行锁的
*/
func CheckExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

const OnePageObjNum = 20                           //一页最多多少个对象。//如果该值越大，那么每次更新评论时的负载就高，我认为20-50比较合适。
const OneConPageNum = 32                           //一个存储体多少页面
const OneConObjNum = OnePageObjNum * OneConPageNum //一个存储体多少个对象
const MaxObjPages = 0x00FFFFFF                     //最大页数，超过该页不能再写入

const ErrMetaId = 0x00FFFFFF

const McDataHeadSize = 8 //McDataIndexHead结构体的长度

const MCDATAOFF = 4

const Int32Bytes = 4

//文件名称格式id必须是唯一标识对象的，id_x1 x1表示在哪个存储体中。而第一个id_0的前面4个字节，表示当前最新使用的是那个存储体
type Store struct {
	PathDir   string //文件路径目录
	Id        string //存储对象的id，比如是用户评论，那么id就使用用户userId等可以唯一标识的前缀
	curUserId int    //当前使用页id
	//curPageUserNum int      //当前页被使用的空间个数,每页有OnePageObjNum个对象
	FirstFd *os.File //第一个存储体
	Fd      *os.File //当前操作的存储体,如果当前操作第一个存储体，则与FirstFd相同
}

func (s *Store) Init(path string, id string) {
	s.PathDir = path
	s.Id = id
}

type MetaData struct {
	UsedId int32 //下一个可以使用的index
	Metas  [OneConPageNum]McDataHead
}

type McDataHead struct {
	Start  uint32 /*开始偏移量*/
	Length uint32 /*一块data的长度。OnePageCommentNum条评论共用该空间，目前看是够了。*/
}

//开始
func (s *Store) begin(askPages int) bool {
	firstCon := fmt.Sprintf("%s/%s_0", s.PathDir, s.Id)
	var err error
	if CheckExists(firstCon) {
		//如果文件存在，则读取当前使用体
		s.FirstFd, err = os.OpenFile(firstCon, os.O_RDWR, 0644)
		if err != nil {
			return false
		}
		if pageNums, _, ok := s.getCurUsedId(); ok {
			s.curUserId = pageNums
			//s.curPageUserNum = curIndex
			if askPages == -1 {
				//操作最新页
				askPages = pageNums
			} else if askPages > pageNums {
				return false //超过当前最大值
			}
			//fmt.Printf("操作存储体%s_%d\n", s.Id, askPages/OneConPageNum)
			if askPages > 0 {
				conId := askPages / OneConPageNum //定位到存储体
				curCon := fmt.Sprintf("%s/%s_%d", s.PathDir, s.Id, conId)
				mode := os.O_RDWR
				//如果不存在则创建
				exist := CheckExists(curCon)
				if !exist {
					mode |= os.O_CREATE
				}
				s.Fd, err = os.OpenFile(curCon, mode, 0644)
				if err != nil {
					s.FirstFd.Close()
					return false
				}

				//凡是新建的，都需要初始化
				if !exist {
					s.InitMcData()
				}
			} else {
				s.Fd = s.FirstFd //就是操作第一个存储体
			}
		}
	} else {
		//第一个块都不存在，直接创建
		s.FirstFd, err = os.OpenFile(firstCon, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return false
		}
		s.Fd = s.FirstFd
		s.InitMcData()
	}
	return true
}

//写结束
func (s *Store) end() {
	if s.Fd == s.FirstFd {
		s.Fd.Close()
	} else {
		s.FirstFd.Close()
		s.Fd.Close()
	}
}

func (s *Store) InitMcData() {
	cur_offset, _ := s.Fd.Seek(0, os.SEEK_CUR)
	metaOff := Int32Bytes + OneConPageNum*McDataHeadSize
	content := make([]byte, metaOff)

	s.Fd.WriteAt(content, cur_offset)
}

//一个32位的数字，前面24位表示页面，后面8位表示当前页面使用索引
func getUsedPagesAndCurIndex(v uint32) (usedPage uint32, curIndex uint32) {
	return (v >> 8), (v & 0x000000FF)
}

func createUsedPagesAndCurIndex(usedPage uint32, curIndex uint32) uint32 {
	return (usedPage << 8) | curIndex
}

//返回当前的元数据
func (s *Store) getCurUsedId() (int, int, bool) {
	used := make([]byte, Int32Bytes)
	n, err := s.FirstFd.ReadAt(used, 0)
	if n >= Int32Bytes {
		v := binary.LittleEndian.Uint32(used)
		i, j := getUsedPagesAndCurIndex(v)
		return int(i), int(j), true
	} else if err == io.EOF {
		return 0, 0, true
	}
	return ErrMetaId, 0, false
}

//已经定位到了当前的存储块，读取其元数据，index是相对当前存储块而言
func (s *Store) readRelativeMetaData(index int, mcHead *McDataHead) bool {
	start := Int32Bytes + index*McDataHeadSize

	buf := make([]byte, McDataHeadSize)

	n1, _ := s.Fd.ReadAt(buf, int64(start))
	if n1 >= McDataHeadSize {
		mcHead.Start = binary.LittleEndian.Uint32(buf[:Int32Bytes])
		mcHead.Length = binary.LittleEndian.Uint32(buf[Int32Bytes:])
		return true
	}
	return false
}

/*追加一块元数据*/
func (s *Store) AppendMetaData(mcHead *McDataHead) bool {

	if s.begin(-1) {
		defer s.end()
		//定位到多少块或页
		useId := s.curUserId % OneConPageNum
		start := Int32Bytes + useId*McDataHeadSize

		buf := make([]byte, McDataHeadSize)

		binary.LittleEndian.PutUint32(buf, mcHead.Start)
		binary.LittleEndian.PutUint32(buf[Int32Bytes:], mcHead.Length)
		s.Fd.WriteAt(buf, int64(start))

		return true
	}
	return false
}

/*更新指定的元数据*/
func (s *Store) UpdateRelativeMetaData(index int, mcHead *McDataHead) bool {
	start := Int32Bytes + index*McDataHeadSize

	buf := make([]byte, McDataHeadSize)

	binary.LittleEndian.PutUint32(buf, mcHead.Start)
	binary.LittleEndian.PutUint32(buf[Int32Bytes:], mcHead.Length)

	s.Fd.WriteAt(buf, int64(start))
	return true

}

/*基本流程。当需要添加一条评论时，读取尾部最后一个数据块的内容。
1 如果尾部没有满，则加上该评论后，走尾部更新函数
2 如果尾部的已经满了，则更新头部索引，使用下一块
*/

//更新当前尾部的一块数据，前面的不能更新，因为写是追加的,故只能更新尾部的。
func (s *Store) UpdateTailBlockToStore(content []byte, elemSize int) (bool, int) {

	if s.begin(-1) {
		defer s.end()

		useId := s.curUserId % OneConPageNum

		//读取当前使用块

		aMeta := &McDataHead{}
		s.readRelativeMetaData(useId, aMeta)
		cur_offset := int64(aMeta.Start)
		/*一定要超过元数据，元数据是不能被写的*/
		metaOff := int64(Int32Bytes + OneConPageNum*McDataHeadSize)
		if cur_offset < metaOff+1 {
			cur_offset = metaOff + 1
		}

		s.Fd.WriteAt(content, cur_offset)

		aMeta.Start = uint32(cur_offset)
		aMeta.Length = uint32(len(content))

		//更新元数据
		s.UpdateRelativeMetaData(useId, aMeta)

		//fmt.Printf("更新存储体%s_%d的第%d块\n", s.Id, s.curUserId/OneConPageNum, useId)
		ret := s.curUserId
		if elemSize >= OnePageObjNum {
			//如果满了，则更新useId到下一块
			//fmt.Printf("存储体%s_%d的第%d块满了\n", s.Id, s.curUserId/OneConPageNum, useId)
			s.curUserId++
			v := createUsedPagesAndCurIndex(uint32(s.curUserId), 0)
			used := make([]byte, Int32Bytes)
			binary.LittleEndian.PutUint32(used, v)
			s.FirstFd.WriteAt(used, 0)

			//获取在当前体的所在页
			useId = s.curUserId % OneConPageNum //所在页

			//初始化新页的start,否则新块的start会从0开始。
			//只要不是首页，则都需要初始化新页的元元素。首页不需要，因为首页就是另外一个新体的第一页，相对就是0
			if useId != 0 {
				aMeta.Start += uint32(aMeta.Length)
				aMeta.Length = 0
				s.UpdateRelativeMetaData(useId, aMeta)
				//fmt.Printf("初始化存储体%s_%d的第%d块\n", s.Id, s.curUserId/OneConPageNum, useId)
			}

		} else {
			//没满更新页使用
			v := createUsedPagesAndCurIndex(uint32(s.curUserId), uint32(elemSize))
			used := make([]byte, Int32Bytes)
			binary.LittleEndian.PutUint32(used, v)
			s.FirstFd.WriteAt(used, 0)
		}
		return true, ret
	}
	return false, 0

}

func (s *Store) GetPageNums() int {
	firstCon := fmt.Sprintf("%s/%s_0", s.PathDir, s.Id)
	var err error
	if CheckExists(firstCon) {
		//如果文件存在，则读取当前使用体
		s.FirstFd, err = os.OpenFile(firstCon, os.O_RDONLY, 0644)
		if err != nil {
			return 0
		}
		defer s.FirstFd.Close()
		//只要存在，哪怕现在使用的是0块，都表示存在存储数据。
		if nums, pageUsed, ok := s.getCurUsedId(); ok {
			if pageUsed == 0 {
				//fmt.Printf("现在实际使用页%d第0条", nums)
				return nums
			} else {
				//fmt.Printf("现在实际使用页%d第%d条", nums, pageUsed)
				return nums + 1
			}
		}
	}
	return 0
}

/*读取一块记忆体*/
func (s *Store) ReadOneBlockMemory(buf []byte, start int64, length int) bool {
	var count int
	var err error
	var index int
	var off int64

	for count < length {
		index, err = s.Fd.ReadAt(buf, start+off)
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
func (s *Store) UpdateBlockToStore(content []byte, blockNum int) (bool, int) {

	if s.begin(blockNum) {
		defer s.end()

		if blockNum == -1 {
			blockNum = s.curUserId
		}
		aMeta := &McDataHead{}

		useId := blockNum % OneConPageNum
		s.readRelativeMetaData(useId, aMeta)
		if len(content) > int(aMeta.Length) {
			//新内容不能超过旧现有内容
			return false, 1
		}
		cur_offset := int64(aMeta.Start)
		/*一定要超过元数据，元数据是不能被写的*/
		metaOff := int64(Int32Bytes + OneConPageNum*McDataHeadSize)
		if cur_offset < metaOff+1 {
			cur_offset = metaOff + 1
		}

		s.Fd.WriteAt(content, cur_offset)

		//更新元数据
		aMeta.Start = uint32(cur_offset)
		aMeta.Length = uint32(len(content))

		s.UpdateRelativeMetaData(useId, aMeta)
		return true, 0
	}
	return false, 0
}

//获取块中一页的内容,如果pageNums=-1，则获取最新一页，忽略pageNums
func (s *Store) GetOnePageContent(pageNums *int) (*[]byte, bool) {

	if s.begin(*pageNums) {
		defer s.end()
		if *pageNums == -1 {
			*pageNums = s.curUserId
		}
		aMeta := &McDataHead{}
		useId := *pageNums % OneConPageNum
		if !s.readRelativeMetaData(useId, aMeta) {
			return nil, false
		}

		if aMeta.Length > 0 {
			buf := make([]byte, aMeta.Length)
			if !s.ReadOneBlockMemory(buf, int64(aMeta.Start), int(aMeta.Length)) {
				return nil, false
			}
			return &buf, true
		} else {
			return nil, true
		}
	}
	return nil, false
}
