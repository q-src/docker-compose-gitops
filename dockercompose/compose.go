package dockercompose

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/q-src/envconf/envconf"
	"os"
	"os/exec"
)

var workdir = envconf.Env{Name: "WORKDIR", DefaultValue: ""}.Get()
var defaultArgs = []string{"-d"}
var log = logging.MustGetLogger("gitops")

func Run(path string, customArgs []string, pull bool) error {
	args := append(defaultArgs, customArgs...)

	if pull {
		cmd := createCommand(path, "pull")
		err := cmd.Run()
		if nil != err {
			log.Warningf("docker-compose pull failed: '%s'", err)
		}
	}

	cmd := createCommand(path, "up", args...)
	err := cmd.Run()
	if nil != err {
		log.Errorf("docker-compose failed: '%s'", err)
	}
	return err
}

func createCommand(path string, subcommand string, subargs ...string) *exec.Cmd {
	args := append([]string{subcommand}, subargs...)
	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = fmt.Sprintf("%s/%s", path, workdir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Debugf("Running docker-compose %s with args '%s'", subcommand, subargs)
	log.Debugf("Selected workdir is '%s'.", cmd.Dir)
	return cmd
}
