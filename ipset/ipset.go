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
	out, err := exec.Command(ips.path, cmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	return nil
}

// Add adds an entry to the set
// opts are additional parameters, for example "timeout 10"
func (ips *IPSet) Add(name, entry string, opts ...string) error {
	cmd := append([]string{"add", name, entry}, opts...)
	out, err := exec.Command(ips.path, cmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	return nil
}

// Del deletes entry from the set
func (ips *IPSet) Del(name, entry string, opts ...string) error {
	cmd := append([]string{"del", name, entry}, opts...)
	out, err := exec.Command(ips.path, cmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	return nil
}

// Test checks if set contains an entry
func (ips *IPSet) Test(name, entry string) (bool, error) {
	out, err := exec.Command(ips.path, "test", name, entry).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%v: %s", err, out)
	}
	if regexp.MustCompile(`is in set`).Match(out) {
		return true, nil
	}
	return false, fmt.Errorf("%s", out)
}

// Destroy destroys the set.
func (ips *IPSet) Destroy(name string, opts ...string) error {
	cmd := append([]string{"destroy", name}, opts...)
	out, err := exec.Command(ips.path, cmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	return nil
}

// DestroyAll destroys all sets.
func (ips *IPSet) DestroyAll() error {
	out, err := exec.Command(ips.path, "destroy").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	return nil
}

// List returns members of a set
func (ips *IPSet) List(name string) ([]string, error) {
	out, err := exec.Command(ips.path, "list", name).CombinedOutput()
	if err != nil {
		return []string{}, fmt.Errorf("%v: %s", err, out)
	}
	listFull := regexp.MustCompile(`(?m)^(.*\n)*Members:\n`).ReplaceAll(out[:], nil)
	listAddrs := regexp.MustCompile(`([^\s]+).*`).FindAllSubmatch(listFull, -1)
	var ret []string
	for _, b := range listAddrs {
		ret = append(ret, string(b[1]))
	}
	return ret, nil
}

// ListSorted same as List, but returns sorted slice
func (ips *IPSet) ListSorted(name string) ([]string, error) {
	out, err := exec.Command(ips.path, "list", name, "-sorted").CombinedOutput()
	if err != nil {
		return []string{}, fmt.Errorf("%v: %s", err, out)
	}
	listFull := regexp.MustCompile(`(?m)^(.*\n)*Members:\n`).ReplaceAll(out[:], nil)
	listAddrs := regexp.MustCompile(`([^\s]+).*`).FindAllSubmatch(listFull, -1)
	var ret []string
	for _, b := range listAddrs {
		ret = append(ret, string(b[1]))
	}
	return ret, nil
}

// ListSets returns all sets
func (ips *IPSet) ListSets() ([]string, error) {
	out, err := exec.Command(ips.path, "list", "-n").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, out)
	}
	return strings.Split(string(out), "\n"), nil
}

// Flush removes all entries from the set
func (ips *IPSet) Flush(name string, opts ...string) error {
	cmd := append([]string{"flush", name}, opts...)
	out, err := exec.Command(ips.path, cmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	return nil
}

// FlushAll removes all entries from all sets
func (ips *IPSet) FlushAll() error {
	out, err := exec.Command(ips.path, "flush").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	return nil
}

// Swap swaps the content of two existing sets
func (ips *IPSet) Swap(from, to string) error {
	out, err := exec.Command(ips.path, "swap", from, to).Output()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
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
	out, err := exec.Command(ips.path, []string{"save"}...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, out)
	}
	return out, nil
}

// Restore invokes ipset restore with stdin data
func (ips *IPSet) Restore(data []byte) error {
	cmd := exec.Command(ips.path, []string{"restore"}...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %s", err)
	}

	go func() {
		defer stdin.Close()
		stdin.Write(data)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	return nil
}
