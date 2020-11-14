package pipeline

import (
	"log"
	"os/exec"
	"regexp"
)

//Pipeline represent a pipeline a set of script to execute for objects in a given bucket corresponding to the filter
type Pipeline struct {
	Name string `json:"name"`
	//BucketSource is a regex
	BucketSource      string   `json:"bucketSource"`
	DestinationPrefix string   `json:"destinationPrefix"`
	Scripts           []Script `json:"scripts"`
	//regex
	Include string `json:"include"`
	Exclude string `json:"exclude"`
}

//IsHandelable return true if the pipeline can handle the given file
func (pipeline *Pipeline) IsHandelable(bucket, file string) bool {
	includeRegex, err := regexp.Compile(pipeline.Include)
	if err != nil {
		pipeline.logError(err)
		return false
	}
	excludeRegex, err := regexp.Compile(pipeline.Exclude)
	if err != nil {
		pipeline.logError(err)
		return false
	}
	bucketRegex, err := regexp.Compile(pipeline.BucketSource)
	if err != nil {
		pipeline.logError(err)
		return false
	}
	return includeRegex.MatchString(file) && !(excludeRegex.MatchString(file) && len(pipeline.Exclude) > 0) && bucketRegex.MatchString(bucket)
}

func (pipeline *Pipeline) logError(err error) {
	log.Println("exception in pipeline", pipeline.Name, err)
}

//Script represent a script to execute with the given runner
type Script struct {
	Script        string `json:"script"`
	Runner        string `json:"runner"`
	RunnerOptions string `json:"options"`
}

//GetCmd build the command to execute the script
func (script *Script) GetCmd() *exec.Cmd {
	if len(script.Runner) == 0 {
		return exec.Command(script.Script)
	}
	return exec.Command(script.Runner, script.RunnerOptions, script.Script)
}
