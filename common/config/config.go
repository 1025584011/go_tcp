package config

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go_tcp/common/logging"
)

type ConfigMap map[string]string
type ConfigMapList map[string][]string

var (
	global *Config
)

type Config struct {
	configmap     ConfigMap
	configmaplist ConfigMapList
}

func NewConfig() *Config {
	config := &Config{
		configmap:     make(ConfigMap),
		configmaplist: make(ConfigMapList),
	}
	return config
}
func (self *Config) GetConfig() *ConfigMap {
	return &self.configmap
}
func (self *Config) SetConfig(key, value string) {
	self.configmap[key] = value
}
func (self *Config) GetConfigList() *ConfigMapList {
	return &self.configmaplist
}
func (self *Config) SetConfigList(key, value string) {
	self.configmaplist[key] = append(self.configmaplist[key], value)
}

func (self *Config) GetConfigStr(key string) string {
	return self.configmap[key]
}
func (self *Config) GetConfigInt(key string) int {
	ret, _ := strconv.Atoi(self.configmap[key])
	return ret
}

func (self *Config) GetConfigStrList(key string) []string {
	return self.configmaplist[key]
}
func (self *Config) GetConfigIntList(key string) []int {
	var ilist []int
	for _, v := range self.configmaplist[key] {
		ret, _ := strconv.Atoi(v)
		ilist = append(ilist, ret)
	}
	return ilist
}

func (self *Config) ListConfig() {
	for k, v := range self.configmap {
		logging.Debug("%s,%s", k, v)
	}
	for k, v := range self.configmaplist {
		for _, v1 := range v {
			logging.Debug("%s,%s", k, v1)
		}
	}
}
func (self *Config) LoadFromFile(filename, node string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return load(self, f, node, false)
}
func (self *Config) LoadFromNet(addr, node string) error {
	resp, err := http.Get(addr)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return &NetError{resp.Status}
	}

	defer resp.Body.Close()
	return load(self, resp.Body, node, false)
}
func (self *Config) LoadListFromFile(filename, node string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return load(self, f, node, true)
}
func (self *Config) LoadListFromNet(addr, node string) error {
	resp, err := http.Get(addr)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return &NetError{resp.Status}
	}

	defer resp.Body.Close()
	return load(self, resp.Body, node, true)
}
func init() {
	global = NewConfig()
}
func load(cfg *Config, r io.Reader, node string, list bool) error {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = charsetReader
	mynode := false
	for {
		t, err := dec.Token()
		if err != nil {
			if err.Error() != "EOF" {
				return err
			}
			break
		}
		switch value := t.(type) {
		case xml.StartElement:
			switch {
			case value.Name.Local == node:
				mynode = true
			case mynode == true:
				enc_base64 := false
				for _, v := range value.Attr {
					if v.Name.Local == "encode" && v.Value == "yes" {
						enc_base64 = true
					}
					if list == true {
						cfg.SetConfigList(v.Name.Local, v.Value)
					} else {
						cfg.SetConfig(v.Name.Local, v.Value)
					}
				}
				td, err := dec.Token()
				if err != nil {
					fmt.Println("td err:", err)
					continue
				}
				switch vd := td.(type) {
				case xml.CharData:
					if enc_base64 {
						bvd, err := base64.URLEncoding.DecodeString(string([]byte(vd)))
						if err != nil {
							if list == true {
								cfg.SetConfigList(value.Name.Local, string([]byte(bvd)))
							} else {
								cfg.SetConfig(value.Name.Local, string([]byte(bvd)))
							}
						} else {
							if list == true {
								cfg.SetConfigList(value.Name.Local, string(vd))
							} else {
								cfg.SetConfig(value.Name.Local, string(vd))
							}
						}

					} else {
						if list == true {
							cfg.SetConfigList(value.Name.Local, string([]byte(vd)))
						} else {
							cfg.SetConfig(value.Name.Local, string([]byte(vd)))
						}
					}
				}
			}
		case xml.EndElement:
			if value.Name.Local == node {
				mynode = false
			}
		}
	}
	return nil
}

type NetError struct {
	err string
}

func (n *NetError) Error() string {
	return n.err
}
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "gb23121":
		buf := make([]byte, 10240)

		reader := bytes.NewBuffer(nil)
		n, err := reader.ReadFrom(input)
		fmt.Println(charset, len(buf), cap(buf), n)
		return reader, err

	}
	return input, nil
}
func LoadFromFile(filename, node string) error {
	return global.LoadFromFile(filename, node)
}
func LoadFromNet(addr, node string) error {
	return global.LoadFromNet(addr, node)
}
func SetConfig(key, value string) {
	global.SetConfig(key, value)
}

func GetConfigStr(key string) string {
	return global.GetConfigStr(key)
}
func GetConfigInt(key string) int {
	return global.GetConfigInt(key)
}

func GetConfig() *ConfigMap {
	return global.GetConfig()
}

func ListConfig() {
	global.ListConfig()
}
