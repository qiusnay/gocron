package model

import (
	"fmt"
)

func Getcron() {

	var cron []Cron
	// fmt.Printf("%v", cron)

	DB().First(&cron)
	fmt.Printf("%v", cron)

	// if lib.TimeDifference(d.ExDate) {
	// 	model.DB().Model(model.CoreQueryOrder{}).Where("username =?", user).Update(&model.CoreQueryOrder{QueryPer: 3})
	// }

	// return c.JSON(http.StatusOK, map[string]interface{}{"status": d.QueryPer, "export": model.GloOther.Export, "idc": d.IDC})
}