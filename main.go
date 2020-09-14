package main

import (
	"github.com/qiusnay/gocron"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
)

func main() {
	config.LoadConfig() // 读取DB配置
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
	fmt.Println("hello world")
}
