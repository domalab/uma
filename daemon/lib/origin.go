package lib

import (
	"strings"

	"github.com/domalab/uma/daemon/common"
	"github.com/domalab/uma/daemon/dto"

	"gopkg.in/ini.v1"
)

func GetOrigin() (*dto.Origin, error) {
	var origin dto.Origin

	nginx, err := ini.Load(common.Nginx)
	if err != nil {
		origin.ErrorCode = "nginx-state"
		origin.ErrorText = err.Error()
		return nil, err
	}

	vars, err := ini.Load(common.Variables)
	if err != nil {
		origin.ErrorCode = "var-state"
		origin.ErrorText = err.Error()
		return nil, err
	}

	origin.Name = getValueOrDefault(vars, "NAME", "Tower")
	origin.Address = getValueOrDefault(nginx, "NGINX_LANIP", "")

	usessl := getValueOrDefault(vars, "USE_SSL", "no")
	if usessl == "no" {
		origin.Protocol = "http"
		origin.Host = getValueOrDefault(nginx, "NGINX_LANNAME", origin.Address)
		origin.Port = getValueOrDefault(vars, "PORT", "80")
	} else {
		origin.Protocol = "https"
		origin.Host = getValueOrDefault(nginx, "NGINX_LANFQDN", getValueOrDefault(nginx, "NGINX_LANMDNS", origin.Address))
		origin.Port = getValueOrDefault(vars, "PORTSSL", "443")
	}

	return &origin, nil
}

func getValueOrDefault(file *ini.File, key string, def string) string {
	value := file.Section("").Key(key).MustString(def)
	value = strings.Replace(value, "\"", "", -1)
	return value
}
