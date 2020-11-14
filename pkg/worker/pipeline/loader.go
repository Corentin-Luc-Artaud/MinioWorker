package pipeline

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

//Load loads the pipelines from a given directory
func Load(dir string) ([]Pipeline, error) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	res := make([]Pipeline, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			pipeline, err := handlePipeline(entry.Name(), dir)
			if err != nil {
				return nil, err
			}
			res = append(res, *pipeline)
		}
	}

	return res, nil
}

func handlePipeline(name string, basedir string) (*Pipeline, error) {
	path := complementPath(name, basedir)
	if err := populateEmptyFile(path); err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pipeline := &Pipeline{}
	err = json.Unmarshal(bytes, pipeline)
	if err != nil {
		return nil, err
	}
	for index, script := range pipeline.Scripts {
		scriptpath := complementPath(script.Script, basedir)
		if !checkpath(path) {
			return nil, errors.New(path + " is not a script")
		}
		script.Script = scriptpath
		pipeline.Scripts[index] = script
	}
	return pipeline, nil
}

func complementPath(path, basedir string) string {
	if strings.HasSuffix(basedir, "/") {
		return basedir + path
	}
	return basedir + "/" + path
}

func populateEmptyFile(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.Size() <= 1 {
		squeleton := &Pipeline{}
		bytes, _ := json.MarshalIndent(squeleton, "", "  ")
		ioutil.WriteFile(path, bytes, 0660)
		return errors.New(path + " was empty")
	}
	return nil
}

func checkpath(path string) bool {
	regex, _ := regexp.Compile("[aA-zZ]([aA-zZ]*[0-9]*(-*_*)*)*(/[aA-zZ]([aA-zZ]*[0-9]*(-*_*)*)*)+")
	if !regex.MatchString(path) {
		return false
	}
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}
	return true
}
