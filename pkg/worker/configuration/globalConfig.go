package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

//GlobalConfig configuration of the worker
type GlobalConfig struct {
	Endpoint                string `json:"endpoint"`
	AccessKeyID             string `json:"accessKeyID"`
	SecretAccessKey         string `json:"secretAccessKey"`
	UseSSl                  bool   `json:"useSSL"`
	MaxConcurentWorker      int    `json:"maxConcurentWorker"`
	MaxConcurentUploadFiles int    `json:"maxConcurentUploadFiles"`
	PipelineDirectory       string `json:"pipelineDirectory"`
	ResultsDestination      string `json:"resultDestination"`
}

//LoadConfiguration load the configuration from the given file, if the file does not exists, create a squelton and save it
func LoadConfiguration(file string) *GlobalConfig {
	globalConfig := &GlobalConfig{}
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		createSqueltonConfiguration(file)
		log.Fatal("error reading configuration file ", err)
	}
	err = json.Unmarshal(bytes, globalConfig)
	if err != nil {
		log.Fatal("error while parsing configuration file", err)
	}
	return globalConfig
}

func createSqueltonConfiguration(file string) {
	squeltonConf := &GlobalConfig{}
	//errors should not happen here
	bytes, _ := json.MarshalIndent(squeltonConf, "", "  ")
	ioutil.WriteFile(file, bytes, os.FileMode(0667))
}
