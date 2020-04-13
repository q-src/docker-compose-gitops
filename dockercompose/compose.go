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
	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Debugf("Running docker-compose with args '%s'", args)
	err := cmd.Run()
	if nil != err {
		log.Errorf("docker-compose failed: '%s'", err)
	}
	return err
}
