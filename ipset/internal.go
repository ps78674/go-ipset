package ipset

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/coreos/go-semver/semver"
)

func (ips *IPSet) checkVersion() (bool, error) {
	minVersion, err := semver.NewVersion(minIPSetVersion)
	if err != nil {
		return false, fmt.Errorf("unable to parse minIPSetVersion: %s", err)
	}
	out, err := exec.Command(ips.path, "--version").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("unable to get ipset version: %v: %s", err, out)
	}
	match := regexp.MustCompile(`v[0-9]+\.[0-9]+`).FindSubmatch(out)
	if match == nil {
		return false, fmt.Errorf("unable to parse ipset version")
	}
	vstring := string(match[0]) + ".0"
	version, err := semver.NewVersion(vstring[1:])
	if err != nil {
		return false, err
	}
	if version.LessThan(*minVersion) {
		return false, nil
	}
	return true, nil
}
