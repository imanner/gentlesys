package timework

//定时做一些工作
import (
	"fmt"
	"gentlesys/global"

	"github.com/jakecoffman/cron"
)

var pCron *cron.Cron

var pCronMin *cron.Cron

type taskCallback func()

var dailyTaskMap map[string]taskCallback //天任务列表

var periodicMinTask map[string]taskCallback //多少分钟间隔做一次任务列表

func init() {
	dailyTaskMap = make(map[string]taskCallback)

	pCron = cron.New()
	pCron.Start()
	runDailyTask()

	if global.IsNginxCache {
		pCronMin = cron.New()
		pCronMin.Start()
		runPeriodicMinTask()
	}

}

//每天凌晨1分开始做事
func runDailyTask() {
	pCron.AddFunc("0 0 1 * * *", func() {
		for _, k := range dailyTaskMap {
			k()
		}
	}, "Often")
}

//每隔30分钟做事
func runPeriodicMinTask() {
	periodicMinTask = make(map[string]taskCallback)
	spec := fmt.Sprintf("*/%d * * * * *", global.NginxAccessFlushTimes)
	pCronMin.AddFunc(spec, func() {
		for _, k := range periodicMinTask {
			k()
		}
	}, "Often")
}

func AddPeriodicMinTask(name string, c taskCallback) {
	periodicMinTask[name] = c
}

func AddDailyTask(name string, c taskCallback) {
	dailyTaskMap[name] = c
}
