package main

import (
	"fmt"
	"github.com/qiusnay/gocron/model"
)

func main() {
	model := mysql{}
	model.DBinit()
	model.Getcron()
	// var cron []model.Cron
	// model.DB().First(&cron)
	fmt.Printf("1111")
}
