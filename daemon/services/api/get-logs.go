package api

import (
	"github.com/domalab/omniraid/daemon/lib"
	"github.com/domalab/omniraid/daemon/logger"
)

var logLocation = map[string]string{
	"system": "/var/log/syslog",
	"docker": "/var/log/docker.log",
	"vm":     "/var/log/libvirt/libvirtd.log",
}

func (a *Api) getLogs(logType string) []string {
	log := make([]string, 0)

	exists, err := lib.Exists(logLocation[logType])
	if err != nil {
		logger.Yellow("unable to check for log existence: %s", err)
		return log
	}

	if !exists {
		logger.Yellow("log %s is not present in the system", logType)
		return log
	}

	cmd := "tail -n 40 " + logLocation[logType]

	lib.Shell(cmd, func(line string) {
		log = append(log, line)
	})

	return log
}
