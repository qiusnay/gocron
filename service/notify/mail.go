package notify

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/google/logger"
	croninit "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/utils"
)

func SendMail(result model.TaskResult, task model.FlCron) interface{} {
	log := croninit.BASEPATH + "/log/mail.log"
	emailConfig := utils.GetConfig("alarm_mail_list", "")
	hostname, _ := os.Hostname()
	mailContent := fmt.Sprintf("Host[%s], Query Cmd[%s]:\n\n", hostname, task.Cmd)
	mailContent += fmt.Sprintf("作业ID:%d\n", task.Jobid)
	mailContent += fmt.Sprintf("作业名称:%s\n", task.JobName)
	mailContent += fmt.Sprintf("执行命令:%s\n", task.Cmd)
	mailContent += fmt.Sprintf("\n错误信息:[%s]:\n", result.Err)
	mailContent += fmt.Sprintf("\n执行结果:[%s]:\n", result.Result)
	mailContent += "\ncron系统地址 : " + emailConfig["cron_url"]
	command := fmt.Sprintf("echo -e \"%s\" | mail -s \"%s\" %s >> %s 2>&1", mailContent, "CRON ALARM MAIL", emailConfig["cron_mail"], log)
	e := exec.Command("/bin/bash", "-c", command)
	logger.Info("绝对路径 %+v", command)
	resultChan := make(chan interface{})
	go func() {
		Result, err := e.Output()
		if err != nil {
			resultChan <- "send mail error :" + err.Error()
		}
		resultChan <- Result
	}()
	return <-resultChan
}
