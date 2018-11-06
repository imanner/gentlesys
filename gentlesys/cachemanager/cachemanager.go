package cachemanager

import (
	"container/list"
	//	"fmt"
	"gentlesys/global"
	"gentlesys/models/sqlsys"
	"gentlesys/subject"
	"sync"
)

//一个用来管理缓存的文件，主要是对主题中的页面进行缓存，避免过多频繁查询数据库

func init() {
	//注意顺序
	initCacheSubject()
}

//与板块话题相关的缓存区
var CacheSubjectObjMaps map[int]*CacheObj

//每个主题的global.CachePagesNums个页面都是在内存中的，每页global.OnePageElementCount个帖子
func initCacheSubject() {
	subNodes := subject.GetMainPageSubjectData()

	CacheSubjectObjMaps = make(map[int]*CacheObj)
	for _, k := range *subNodes {
		CacheSubjectObjMaps[k.UniqueId] = new(CacheObj)
		CacheSubjectObjMaps[k.UniqueId].SubId = k.UniqueId
		CacheSubjectObjMaps[k.UniqueId].element = make([]*sqlsys.Subject, global.OnePageElementCount*global.CachePagesNums)
		CacheSubjectObjMaps[k.UniqueId].accessFlag = make([]int, global.OnePageElementCount*global.CachePagesNums)
		CacheSubjectObjMaps[k.UniqueId].InitCacheTopicListFromDb()
	}
}

type CacheObj struct {
	SubId      int               //所属的主题板块
	mutex      sync.Mutex        //用于保护结构体的锁，保护list
	newAddList *list.List        //在头部新加入的元素，先加入该列表
	list       *list.List        //缓存的结构体
	element    []*sqlsys.Subject //list的索引，加速定位，空间换时间
	accessFlag []int             //页面是否访问过的标识
}

//初始化操作时没有加锁，考虑到还在程序初始化期，不会并发
func (c *CacheObj) InitCacheTopicListFromDb() {
	var aSubject sqlsys.Subject
	nums := global.OnePageElementCount * global.CachePagesNums
	pTopicList := aSubject.GetTopicListSortByTime(c.SubId, nums)

	c.list = list.New()

	c.newAddList = list.New()

	//将数据保存在列表中topicList 是 *[]orm.Params
	if len(*pTopicList) > 0 {
		for i, _ := range *pTopicList {
			c.list.PushBack(&(*pTopicList)[i])
			c.element[i] = &(*pTopicList)[i]
			//处理匿名
			if (*pTopicList)[i].Anonymity {
				(*pTopicList)[i].UserName = "匿名网友"
			}
		}

		subject.UpdateCurTopicIndex(c.SubId, (*pTopicList)[0].Id)
	}

	//atomic.StoreUint32(&mysqlTool.ShareCureIndex, shareList[0].Id)

	//fmt.Printf("inital list count %d\n", c.list.Len())
}

/*
//每天还是更新一次
func (c *CacheObj) DayUpdateOneTimes() {

	var aShare mysqlTool.Share
	nums := global.OnePageElementCount * global.CachePagesNums //只缓存了6个页面的记录
	shareList := aShare.GetNShareListSortByTime(nums)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.list.Init()
	c.newAddList.Init()

	//将数据保存在列表中
	for i, k := range shareList {
		c.list.PushBack(k)
		c.element[i] = k

		//处理匿名
		if k.Anonymity {
			shareList[i].ArName = "晒方网友"
		}
	}
}
*/
//在链表头部加入一个数据。v 必须是 *Share 类型。返回值 是否清空过nginx缓存
func (c *CacheObj) AddElementAtFront(v interface{}) {
	c.mutex.Lock()
	//defer c.mutex.Unlock()

	c.newAddList.PushFront(v)

	if c.newAddList.Len() >= global.FlushNumsLimit {
		//新加的元素超过global.FlushNumsLimit个了。将二者合并
		c.list.PushFrontList(c.newAddList)
		c.newAddList.Init()

		//只缓存global.OnePageElementCount * global.CachePagesNums个元素，超过的删除
		if c.list.Len() > global.OnePageElementCount*global.CachePagesNums {
			for i := global.OnePageElementCount * global.CachePagesNums; i < c.list.Len(); i++ {
				e := c.list.Back()
				if e != nil {
					c.list.Remove(e)
				}
			}
		}
		//更新索引
		var i int
		for e := c.list.Front(); e != nil; e = e.Next() {
			c.element[i] = e.Value.(*sqlsys.Subject)
			i++
		}

		c.mutex.Unlock()
		//刷新此类的全部缓存，即让Nginx缓存失效
		//c.clearMainPageNginxCache()
		return
	}
	c.mutex.Unlock()
	//刷新主页。新元素不超过界限，只刷新首页
	//clearcache.ClearPath("/")
	return
}

/*
func (c *CacheObj) clearMainPageNginxCache() {
	//不刷新page=0，因为没有page=0，page=0是主页
	clearcache.ClearPath("/") // 这个就是page=0
	for i := 1; i < global.CachePagesNums; i++ {
		//只有访问过的才清除
		if c.accessFlag[i] == 1 {
			clearcache.ClearPath(fmt.Sprintf("/?page=%d", i))
			c.accessFlag[i] = 0
			//fmt.Printf("clear cache page %d ...\n", i)
		}

	}
	//fmt.Printf("clear cache ...\n")
}
*/

//读取一页的元素数据。一共有global.CachePagesNums页
func (c *CacheObj) ReadElementsWithPageNums(pageNums int) []*sqlsys.Subject {

	if pageNums < 0 || pageNums >= global.CachePagesNums {
		return nil
	}

	c.accessFlag[pageNums] = 1

	//每页global.OnePageElementCount
	start := pageNums * global.OnePageElementCount
	end := (pageNums + 1) * global.OnePageElementCount //不包含end

	//这里加mutexAddList锁，是因为合并中会访问，可能导致链表结构乱
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if start > c.list.Len() {
		return nil
	}

	if end >= c.list.Len() {
		end = c.list.Len()
	}

	//如果访问的是首页，则还有加上addList。故首页可能不止50条记录。
	var ret []*sqlsys.Subject

	j := 0

	if pageNums == 0 {
		ret = make([]*sqlsys.Subject, end-start+c.newAddList.Len())
		for e := c.newAddList.Front(); e != nil; e = e.Next() {
			ret[j] = e.Value.(*sqlsys.Subject) //保存的都是地址
			j++
		}
		//fmt.Printf("new add %d, end %d start %d len %d\n", j, end, start, c.newAddList.Len())
	} else {
		//非首页不需要访问addList
		ret = make([]*sqlsys.Subject, end-start)
		//fmt.Printf("2 list %d, end %d start %d\n", j, end, start)

	}

	for index := start; index < end; index++ {
		ret[j] = c.element[index] //保存的都是地址
		j++
	}
	return ret
}
