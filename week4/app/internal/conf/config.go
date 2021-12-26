package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type config struct {
	GrpcNetWork string
	GrpcAddr    string
}

var configFilePath = "./app/configs/config.json"

var conf config

func GetGrpcConfig() (string, string) {
	return conf.GrpcNetWork, conf.GrpcAddr

}

func init() {
	configByte, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	if len(configByte) <= 0 {
		log.Fatalln("config file invalid")
	}

	err = json.Unmarshal(configByte, &conf)
	if err != nil {
		log.Fatalln(err)
	}
}
