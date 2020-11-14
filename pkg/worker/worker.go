package worker

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/corentin-luc-artaud/MinioWorker/pkg/worker/network"
)

type worker struct {
	w        work
	endChan  chan string
	endpoint network.NetworkEndpoint
}

func (w *worker) work() {
	//notify end
	defer func() { w.endChan <- w.w.getLockKey() }()
	//ask and wait for download
	w.endpoint.Download(w.w.Bucket, w.w.File, w.getLocalPath(w.w.File))
	//run programs on file and collect output
	for pos, script := range w.w.Description.Scripts {
		err := w.runScript(script.GetCmd())
		if err != nil {
			log.Println("error when running script ", w.w.Description.Name, pos, err)
		}
	}
	os.RemoveAll(w.w.LocalDir)
}

func (w *worker) runScript(cmd *exec.Cmd) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	//start the command after redirecting I/O
	err = cmd.Start()
	if err != nil {
		return err
	}

	//handle input
	io.WriteString(stdin, w.w.LocalDir+"\n")
	io.WriteString(stdin, w.w.File+"\n")
	stdin.Close() //important to stop some script properly

	//handle output
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		file := w.getLocalPath(scanner.Text())
		destFile := w.getDestinationPath(file)
		//ask network endpoint to upload this file
		w.endpoint.Upload(w.w.Bucket, destFile, file)
	}
	w.endpoint.WaitAll()

	//handle error output
	io.Copy(os.Stderr, stderr)

	return cmd.Wait()
}

func (w *worker) getLocalPath(file string) string {
	parts := strings.Split(file, "/")
	name := parts[len(parts)-1]
	if len(w.w.LocalDir) == 0 {
		return name
	}
	if strings.HasSuffix(w.w.LocalDir, "/") {
		return w.w.LocalDir + name
	}
	return w.w.LocalDir + "/" + name
}

func (w *worker) getDestinationPath(file string) string {
	parts := strings.Split(file, "/")
	name := parts[len(parts)-1]
	if len(w.w.Description.DestinationPrefix) > 0 {
		if strings.HasSuffix(w.w.Description.DestinationPrefix, "/") {
			name = w.w.Description.DestinationPrefix + name
		}
		name = w.w.Description.DestinationPrefix + "/" + name
	}

	if len(w.w.ResultsDestination) > 0 {
		if strings.HasSuffix(w.w.ResultsDestination, "/") {
			name = w.w.ResultsDestination + name
		}
		name = w.w.ResultsDestination + "/" + name
	}
	return name
}

func runWorker(w work, endChan chan string, endpoint network.NetworkEndpoint) {
	task := worker{w, endChan, endpoint}
	go task.work()
}
