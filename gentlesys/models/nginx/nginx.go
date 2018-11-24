package nginx

import (
	"bytes"
	"fmt"
	"gentlesys/global"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
)

func ClearccessTimes() {
	exec_shell(fmt.Sprintf("echo 0 > %s", global.NginxAccessLogPath))
}

func exec_shell(s string) (string, error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("/bin/bash", "-c", s)

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()
	return out.String(), err
}

type callback func(int, int, int)

//获取每一个帖子的日访问量
func UpdateTopicDayAccessTimes(fn callback) bool {

	accessInfo, err := exec_shell(fmt.Sprintf("awk '{print $7}' %s|sort | uniq -c |sort -n -k 1 -r", global.NginxAccessLogPath))
	if err != nil {
		logs.Error("UpdateTopicDayAccessTimes err", err)
		return false
	}

	s := strings.Split(accessInfo, "\n")

	for _, v := range s {

		c := strings.Split(strings.TrimSpace(v), " ")
		if len(c) >= 2 && strings.HasPrefix(c[1], `/browse?sid=`) {

			rsid := regexp.MustCompile(`sid=\d+`)
			raid := regexp.MustCompile(`aid=\d+`)
			sid := rsid.FindString(c[1])
			aid := raid.FindString(c[1])

			if sid != "" && aid != "" {
				isid, _ := strconv.Atoi(sid[4:])
				iaid, _ := strconv.Atoi(aid[4:])
				times, _ := strconv.Atoi(c[0])
				fn(isid, iaid, times)
			}
		}
	}
	return true
}
