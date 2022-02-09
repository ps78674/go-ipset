package ipset

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/coreos/go-semver/semver"
)

func (ips *IPSet) run(in []byte, cmd []string) ([]byte, []byte, error) {
	c := exec.Command(ips.path, cmd...)
	if in != nil {
		_stdin, err := c.StdinPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("error creating stdin pipe: %s", err)
		}

		go func() {
			defer _stdin.Close()
			_stdin.Write(in)
		}()
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (ips *IPSet) checkVersion() (bool, error) {
	minVersion, err := semver.NewVersion(minIPSetVersion)
	if err != nil {
		return false, fmt.Errorf("unable to parse minIPSetVersion: %s", err)
	}
	stdout, stderr, err := ips.run(nil /* in */, []string{"--version"} /* cmd */)
	if err != nil {
		return false, fmt.Errorf("unable to get ipset version: %v: %s", err, stderr)
	}
	match := regexp.MustCompile(`v[0-9]+\.[0-9]+`).FindSubmatch(stdout)
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
