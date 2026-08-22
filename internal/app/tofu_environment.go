package app

import "github.com/projectious-work/ainfra/internal/security"

var tofuAllowedEnvironment = []string{
	"HOME",
	"PATH",
	"SSL_CERT_DIR",
	"SSL_CERT_FILE",
	"HCLOUD_TOKEN",
}

func buildTofuEnvironment(parent []string) (security.Environment, error) {
	return security.BuildEnvironment(parent, tofuAllowedEnvironment,
		map[string]string{"TF_IN_AUTOMATION": "1"})
}
