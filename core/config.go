package core

import (
	"github.com/cro4k/annotation/utils/array"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"
)

const configFile = ".ann/ann.yml"

type Config struct {
	Replace map[string][]string
}

var config *Config

func init() {
	var err error
	config, err = loadConfig(configFile)
	if err != nil {
		config = &Config{Replace: make(map[string][]string)}
	} else {

	}
}

func loadConfig(path string) (*Config, error) {
	var body io.Reader
	if strings.HasPrefix(path, "http") {
		if resp, err := http.Get(path); err != nil {
			return nil, err
		} else {
			defer resp.Body.Close()
			body = resp.Body
		}
	} else {
		if fi, err := os.Open(path); err != nil {
			return nil, err
		} else {
			defer fi.Close()
			body = fi
		}
	}
	var cfg = new(Config)
	err := yaml.NewDecoder(body).Decode(cfg)
	return cfg, err
}

func SetConfig(replace map[string][]string) {
	for k := range replace {
		config.Replace[k] = append(config.Replace[k], replace[k]...)
	}
}

func DelConfig(replace map[string][]string) {
	for k, val := range replace {
		if len(val) == 0 {
			delete(config.Replace, k)
		} else {
			if res := array.Remove(config.Replace[k], val); len(res) > 0 {
				config.Replace[k] = res
			} else {
				delete(config.Replace, k)
			}
		}
	}
}

func WriteConfigFile() error {
	os.Remove(configFile)
	os.Mkdir(".ann", 0777)
	fi, err := os.OpenFile(configFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fi.Close()
	return yaml.NewEncoder(fi).Encode(config)
}
