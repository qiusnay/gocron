package utils

import (
	"fmt"
	"os"
	"strings"
	"bufio"
	"io"
	// "github.com/google/logger"
)

type Config struct {
	filepath string
	conflist []map[string]map[string]string
}

//To obtain corresponding value of the key values
func GetConfig(section, feilds string) map[string]string {
	//读取配置
	c := new(Config)
	c.filepath = "./conf/conf.ini"
	conf := c.ReadList()
	// logger.Infof("database connect erro : %v", conf)
	if feilds == "" { //如果不传具体的feilds
		for _, v := range conf {
			for key, value := range v {
				if key == section {
					// fmt.Printf("%v", value)
					return value;
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
		case string(line[0]) == "#":	//增加配置文件备注
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