package timework

//定时做一些工作
import (
	"github.com/jakecoffman/cron"
)

var pCron *cron.Cron

type taskCallback func()

var taskMap map[string]taskCallback //任务列表

func init() {
	taskMap = make(map[string]taskCallback)
	pCron = cron.New()
	pCron.Start()
	runDailyTask()
}

//每天凌晨1分开始做事
func runDailyTask() {
	pCron.AddFunc("0 0 1 * * *", func() {
		for _, k := range taskMap {
			k()
		}
	}, "Often")
}

func AddDailyTask(name string, c taskCallback) {
	taskMap[name] = c
}
