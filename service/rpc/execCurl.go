package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	// "github.com/qiusnay/gocron/model"
	// "github.com/qiusnay/gocron/init"
	// "github.com/google/logger"
)

type RpcServiceCurl struct {
	Result string
	Err    error
}

type ResponseWrapper struct {
	StatusCode int
	Body       string
	Header     http.Header
}

// http任务执行时间不超过300秒
const HttpExecTimeout = 300

func (h *RpcServiceCurl) ExecCurl(ctx context.Context, command string) CronResponse {
	var resp ResponseWrapper
	resp = h.Get(command, HttpExecTimeout)
	// 返回状态码非200，均为失败
	if resp.StatusCode != http.StatusOK {
		return CronResponse{"", CronError, "", fmt.Errorf("HTTP状态码非200-->%d")}
	}
	return CronResponse{"", CronSucess, resp.Body, nil}
}

func (h *RpcServiceCurl) Get(url string, timeout int) ResponseWrapper {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return createRequestError(err)
	}

	return request(req, timeout)
}

func (h *RpcServiceCurl) PostParams(url string, params string, timeout int) ResponseWrapper {
	buf := bytes.NewBufferString(params)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return createRequestError(err)
	}
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	return request(req, timeout)
}

func request(req *http.Request, timeout int) ResponseWrapper {
	wrapper := ResponseWrapper{StatusCode: 0, Body: "", Header: make(http.Header)}
	client := &http.Client{}
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}
	setRequestHeader(req)
	resp, err := client.Do(req)
	if err != nil {
		wrapper.Body = fmt.Sprintf("执行HTTP请求错误-%s", err.Error())
		return wrapper
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		wrapper.Body = fmt.Sprintf("读取HTTP请求返回值失败-%s", err.Error())
		return wrapper
	}
	wrapper.StatusCode = resp.StatusCode
	wrapper.Body = string(body)
	wrapper.Header = resp.Header
	return wrapper
}

func setRequestHeader(req *http.Request) {
	req.Header.Set("User-Agent", "golang/gocron")
}

func createRequestError(err error) ResponseWrapper {
	errorMessage := fmt.Sprintf("创建HTTP请求错误-%s", err.Error())
	return ResponseWrapper{0, errorMessage, make(http.Header)}
}
