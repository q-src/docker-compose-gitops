package dockercompose

import (
	"github.com/op/go-logging"
	"os"
	"os/exec"
)

var defaultArgs = []string{"up", "-d"}
var log = logging.MustGetLogger("gitops")

func Run(path string, customArgs []string) error {
	args := append(defaultArgs, customArgs...)

	cmd := createCommand(path, "pull")
	err := cmd.Run()
	if nil != err {
		log.Warningf("docker-compose pull failed: '%s'", err)
	}

	cmd = createCommand(path, "up", args...)
	err = cmd.Run()
	if nil != err {
		log.Errorf("docker-compose failed: '%s'", err)
	}
	return err
}

func createCommand(path string, subcommand string, subargs ...string) *exec.Cmd {
	args := append([]string{subcommand}, subargs...)
	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Debugf("Running docker-compose %s with args '%s'",subcommand, subargs)
	return cmd
}
