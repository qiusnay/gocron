package utils

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	// "path/filepath"
	"github.com/google/logger"
)

var BASEPATH, _ = filepath.Abs(filepath.Dir("../"))

type Config struct {
	filepath string
	conflist []map[string]map[string]string
}

//To obtain corresponding value of the key values
func GetConfig(section, feilds string) map[string]string {
	//读取配置
	c := new(Config)
	c.filepath = BASEPATH + "/conf/conf.ini"
	conf := c.ReadList()
	logger.Infof("erro : %v", c.filepath)
	if feilds == "" { //如果不传具体的feilds
		for _, v := range conf {
			for key, value := range v {
				if key == section {
					// fmt.Printf("%v", value)
					return value
				}
			}
		}
	}
	//读取指定的feilds
	mFeilds := strings.Split(feilds, ",")
	dbConf := make(map[string]string, len(mFeilds))
	for _, feild := range mFeilds {
		for _, v := range conf {
			for key, value := range v {
				if key == section {
					dbConf[feild] = value[feild]
				}
			}
		}
	}
	return dbConf
}

//List all the configuration file
func (c *Config) ReadList() []map[string]map[string]string {
	file, err := os.Open(c.filepath)
	if err != nil {
		CheckErr(err)
	}
	defer file.Close()
	var data map[string]map[string]string
	var section string
	buf := bufio.NewReader(file)
	for {
		l, err := buf.ReadString('\n')
		line := strings.TrimSpace(l)
		if err != nil {
			if err != io.EOF {
				CheckErr(err)
			}
			if len(line) == 0 {
				break
			}
		}
		switch {
		case len(line) == 0:
		case string(line[0]) == "#": //增加配置文件备注
		case line[0] == '[' && line[len(line)-1] == ']':
			section = strings.TrimSpace(line[1 : len(line)-1])
			data = make(map[string]map[string]string)
			data[section] = make(map[string]string)
		default:
			i := strings.IndexAny(line, "=")
			if i == -1 {
				continue
			}
			value := strings.TrimSpace(line[i+1 : len(line)])
			data[section][strings.TrimSpace(line[0:i])] = value
			if c.uniquappend(section) == true {
				c.conflist = append(c.conflist, data)
			}
		}
	}
	return c.conflist
}

func CheckErr(err error) string {
	if err != nil {
		return fmt.Sprintf("Error is :'%s'", err.Error())
	}
	return "Notfound this error"
}

//Ban repeated appended to the slice method
func (c *Config) uniquappend(conf string) bool {
	for _, v := range c.conflist {
		for k, _ := range v {
			if k == conf {
				return false
			}
		}
	}
	return true
}

// 判断目录是否存在
func IsDir(fileAddr string) bool {
	s, err := os.Stat(fileAddr)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func InArray(need interface{}, needArr []string) bool {
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}
