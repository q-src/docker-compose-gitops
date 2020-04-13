package main

import (
	"./dockercompose"
	"./git"
	"github.com/op/go-logging"
	"github.com/q-src/envconf/envconf"
	"github.com/robfig/cron"
	"os"
)

var UpdateCron = envconf.Env{Name: "CRON", DefaultValue: "* * * * *"}
var Remote = envconf.Env{Name: "REMOTE", DefaultValue: "origin"}
var RepositoryPath = envconf.Env{Name: "REPOSITORY_PATH", DefaultValue: "/tmp"}
var SshPrivKeyFile = envconf.Env{Name: "SSH_PRIV_KEY_FILE", DefaultValue: "~/.ssh/id_rsa"}
var LogFormat = envconf.Env{Name: "LOG_FORMAT", DefaultValue: `%{color}%{time:2006-01-02 15:04:05.000} [%{level:.4s} / %{id:03x}] %{message}%{color:reset}`}
var log = logging.MustGetLogger("gitops")
var path = RepositoryPath.Get()

func main() {
	initLogger()
	startCron()
}

func initLogger() {
	format := logging.MustStringFormatter(LogFormat.Get())
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	formatted := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatted)
}

func update() {
	deplyomentConfig, err := git.Open(path, Remote.Get(), SshPrivKeyFile.Get())
	if nil != err {
		return
	}
	oldRef, err := deplyomentConfig.Update()
	if nil != err {
		return
	}
	err = dockercompose.Run(path, os.Args[1:], true)
	if nil == err {
		return
	}
	log.Infof("Recovering old commit '%s", oldRef.Hash().String())
	_, err = deplyomentConfig.Apply(oldRef)
	if nil != err {
		log.Errorf("Unable to recover old commit.")
		return
	}
	err = dockercompose.Run(path, os.Args[1:], false)
	if nil != err {
		log.Errorf("Recovering of old commit was successful. However, bringing services up failed.")
		return
	}
}

func startCron() {
	cronSchedule := UpdateCron.Get()
	crond := cron.New()
	_, err := crond.AddFunc(cronSchedule, update)
	if nil != err {
		log.Errorf("Unable to start GitOps daemon with schedule '%s' : %s", cronSchedule, err)
		os.Exit(1)
	}
	crond.Run()
	log.Infof("GitOps daemon started. Using cron '%s'.", cronSchedule)
}
