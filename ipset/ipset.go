/*
Copyright 2015 Jan Broer All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package ipset is a library providing a wrapper to the IPtables ipset userspace utility
package ipset

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const minIPSetVersion = "6.0.0"

// Params defines optional parameters for creating a new set.
type Params struct {
	HashFamily string
	HashSize   int
	MaxElem    int
	Timeout    int
}

// Cmd is an interface to ipset command
type IPSet struct {
	path string
}

// New returns new ipset command instance
func New() (*IPSet, error) {
	path, err := exec.LookPath("ipset")
	if err != nil {
		return nil, err
	}
	cmd := &IPSet{path: path}
	supported, err := cmd.checkVersion()
	if err != nil {
		return nil, fmt.Errorf("error validating ipset version: %s", err)
	}
	if !supported {
		return nil, fmt.Errorf("ipset version is not supported")
	}
	return cmd, nil
}

// Create creates new ipset
func (ips *IPSet) Create(name string, hashType string, p *Params, opts ...string) error {
	// set default ipset values
	if p.HashSize == 0 {
		p.HashSize = 1024
	}
	if p.MaxElem == 0 {
		p.MaxElem = 65536
	}
	if p.HashFamily == "" {
		p.HashFamily = "inet"
	}

	// check hash type is in form 'hash:<TYPE>'
	if !strings.HasPrefix(hashType, "hash:") {
		return fmt.Errorf("not a hash type: %s", hashType)
	}

	cmd := append([]string{"create", name, hashType, "family", p.HashFamily, "hashsize", strconv.Itoa(p.HashSize),
		"maxelem", strconv.Itoa(p.MaxElem), "timeout", strconv.Itoa(p.Timeout)}, opts...)

	_, stderr, err := ips.run(nil /* in */, cmd /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}

	return nil
}

// Add adds an entry to the set
// opts are additional parameters, for example "timeout 10"
func (ips *IPSet) Add(name, entry string, opts ...string) error {
	cmd := append([]string{"add", name, entry}, opts...)
	_, stderr, err := ips.run(nil /* in */, cmd /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// Del deletes entry from the set
func (ips *IPSet) Del(name, entry string, opts ...string) error {
	cmd := append([]string{"del", name, entry}, opts...)
	_, stderr, err := ips.run(nil /* in */, cmd /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// Test checks if set contains an entry
func (ips *IPSet) Test(name, entry string) (bool, error) {
	stdout, stderr, err := ips.run(nil /* in */, []string{"test", name, entry} /* cmd */)
	if err != nil {
		return false, fmt.Errorf("%v: %s", err, stderr)
	}
	if regexp.MustCompile(`is in set`).Match(stdout) {
		return true, nil
	}
	return false, fmt.Errorf("%s", stderr)
}

// Destroy destroys the set.
func (ips *IPSet) Destroy(name string, opts ...string) error {
	cmd := append([]string{"destroy", name}, opts...)
	_, stderr, err := ips.run(nil /* in */, cmd /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// DestroyAll destroys all sets.
func (ips *IPSet) DestroyAll() error {
	_, stderr, err := ips.run(nil /* in */, []string{"destroy"} /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// List returns members of a set
func (ips *IPSet) List(name string) ([]string, error) {
	stdout, stderr, err := ips.run(nil /* in */, []string{"list", name} /* cmd */)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr)
	}
	listFull := regexp.MustCompile(`(?m)^(.*\n)*Members:\n`).ReplaceAll(stdout[:], nil)
	listAddrs := regexp.MustCompile(`([^\s]+).*`).FindAllSubmatch(listFull, -1)
	var ret []string
	for _, b := range listAddrs {
		ret = append(ret, string(b[1]))
	}
	return ret, nil
}

// ListSorted same as List, but returns sorted slice
func (ips *IPSet) ListSorted(name string) ([]string, error) {
	stdout, stderr, err := ips.run(nil /* in */, []string{"list", name, "-sorted"} /* cmd */)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr)
	}
	listFull := regexp.MustCompile(`(?m)^(.*\n)*Members:\n`).ReplaceAll(stdout[:], nil)
	listAddrs := regexp.MustCompile(`([^\s]+).*`).FindAllSubmatch(listFull, -1)
	var ret []string
	for _, b := range listAddrs {
		ret = append(ret, string(b[1]))
	}
	return ret, nil
}

// ListSets returns all sets
func (ips *IPSet) ListSets() ([]string, error) {
	stdout, stderr, err := ips.run(nil /* in */, []string{"list", "-n"} /* cmd */)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr)
	}
	return strings.Split(string(stdout), "\n"), nil
}

// Flush removes all entries from the set
func (ips *IPSet) Flush(name string, opts ...string) error {
	cmd := append([]string{"flush", name}, opts...)
	_, stderr, err := ips.run(nil /* in */, cmd /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// FlushAll removes all entries from all sets
func (ips *IPSet) FlushAll() error {
	_, stderr, err := ips.run(nil /* in */, []string{"flush"} /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// Swap swaps the content of two existing sets
func (ips *IPSet) Swap(from, to string) error {
	_, stderr, err := ips.run(nil /* in */, []string{"swap", from, to} /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// Replace overwrites the set with new entries
func (ips *IPSet) Replace(name string, entries []string) error {
	if err := ips.Flush(name); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := ips.Add(name, entry); err != nil {
			return err
		}
	}
	return nil
}

// Save returns ipset save output as []byte
func (ips *IPSet) Save() ([]byte, error) {
	stdout, stderr, err := ips.run(nil /* in */, []string{"save"} /* cmd */)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr)
	}

	return stdout, nil
}

// Restore invokes ipset restore with stdin data
func (ips *IPSet) Restore(data []byte) error {
	_, stderr, err := ips.run(data /* in */, []string{"restore"} /* cmd */)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}

	return nil
}
