package sqlsys

/*这地方是为主题数据库表预先定义的结构体，为了使用orm创建表，而不得已为之。
一是因为orm只能支持一个结构体创建一张表，不支持一个结构体创建多张表。
二来golang语言无法支持模板结构等动态创建结构体的方式。预先创建30个结构体，
可以创建30张数据库主题表。暂时这样解决。如果实际业务需要多余30个主题，可以自行
按照格式添加Subx结构体，并补充上GetInstanceById函数。
*/

type Sub0 struct {
	Subject
}

type Sub1 struct {
	Subject
}
type Sub2 struct {
	Subject
}
type Sub3 struct {
	Subject
}
type Sub4 struct {
	Subject
}
type Sub5 struct {
	Subject
}

type Sub6 struct {
	Subject
}
type Sub7 struct {
	Subject
}
type Sub8 struct {
	Subject
}
type Sub9 struct {
	Subject
}
type Sub10 struct {
	Subject
}

type Sub11 struct {
	Subject
}
type Sub12 struct {
	Subject
}
type Sub13 struct {
	Subject
}
type Sub14 struct {
	Subject
}
type Sub15 struct {
	Subject
}

type Sub16 struct {
	Subject
}
type Sub17 struct {
	Subject
}
type Sub18 struct {
	Subject
}
type Sub19 struct {
	Subject
}
type Sub20 struct {
	Subject
}

type Sub21 struct {
	Subject
}
type Sub22 struct {
	Subject
}
type Sub23 struct {
	Subject
}
type Sub24 struct {
	Subject
}
type Sub25 struct {
	Subject
}
type Sub26 struct {
	Subject
}
type Sub27 struct {
	Subject
}
type Sub28 struct {
	Subject
}
type Sub29 struct {
	Subject
}

//表示公告的数据结构，但是Type不表示类型，而是表示SubId,特此说明
type Sub1001 struct {
	Subject
}

/*将获取内含Subject地址的函数放在基类中，这样Subxx均包含该方法*/
func (s *Subject) GetSubject() *Subject {
	return s
}

/*从具体类subxx中获取它的组合结构Subject，让外部可以直接修改Subject值*/
type TranInterface interface {
	GetSubject() *Subject
}

func GetInstanceById(id int) TranInterface {
	switch id {
	case 0:
		return new(Sub0)

	case 1:
		return new(Sub1)

	case 2:
		return new(Sub2)

	case 3:
		return new(Sub3)

	case 4:
		return new(Sub4)

	case 5:
		return new(Sub5)

	case 6:
		return new(Sub6)

	case 7:
		return new(Sub7)

	case 8:
		return new(Sub8)

	case 9:
		return new(Sub9)
	case 10:
		return new(Sub10)

	case 11:
		return new(Sub11)

	case 12:
		return new(Sub12)

	case 13:
		return new(Sub13)

	case 14:
		return new(Sub14)

	case 15:
		return new(Sub15)

	case 16:
		return new(Sub16)

	case 17:
		return new(Sub17)

	case 18:
		return new(Sub18)

	case 19:
		return new(Sub19)
	case 20:
		return new(Sub20)

	case 21:
		return new(Sub21)
	case 22:
		return new(Sub22)

	case 23:
		return new(Sub23)

	case 24:
		return new(Sub24)

	case 25:
		return new(Sub25)

	case 26:
		return new(Sub26)

	case 27:
		return new(Sub27)
	case 28:
		return new(Sub28)
	case 29:
		return new(Sub29)
		//第1001号标签，表示是公告，其他主题不可占用该标签
	case 1001:
		return new(Sub1001)
	default:
		return nil
	}
}
