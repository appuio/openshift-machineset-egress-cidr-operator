package controller

import (
	"regexp"
	"sort"
	"sync"

	v1 "github.com/openshift/api/network/v1"
)

type CIDRMap struct {
	entries map[string][]v1.HostSubnetEgressCIDR
	mutex   *sync.RWMutex
}

var (
	splitRe = regexp.MustCompile(`,\s*`)
)

func NewCIDRMap() *CIDRMap {
	return &CIDRMap{
		entries: make(map[string][]v1.HostSubnetEgressCIDR),
		mutex:   new(sync.RWMutex),
	}
}

// Set takes a list of (comma separated) values, splits and sorts them, and
// then inserts them into the cache for `machineSetName`.
func (m *CIDRMap) Set(machineSetName, s string) {
	cidrs := splitCIDRs(s)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.entries[machineSetName] = cidrs
}

func (m *CIDRMap) Delete(machineSetName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.entries, machineSetName)
}

func (m *CIDRMap) Exists(machineSetName string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.entries[machineSetName]) > 0
}

func (m *CIDRMap) Get(machineSetName string) []v1.HostSubnetEgressCIDR {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	entries := m.entries[machineSetName]

	if len(entries) == 1 && entries[0] == "none" {
		return []v1.HostSubnetEgressCIDR{}
	}

	return entries
}

// Equals returns true if the splitted, sorted value of v is equal to the entry
// in the cache for `machineSetName`.
func (m *CIDRMap) Equals(machineSetName, v string) bool {
	other := splitCIDRs(v)

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	this := m.entries[machineSetName]

	return compare(this, other)
}

func (m *CIDRMap) EqualCIRDs(machineSetName string, other []v1.HostSubnetEgressCIDR) bool {
	s := egressCIDRsToStrings(other)
	sort.Strings(s)
	other = stringsToEgressCIDRs(s)

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	this := m.entries[machineSetName]

	if len(this) == 1 && this[0] == "none" {
		return len(other) == 0
	}

	return compare(this, other)
}

func splitCIDRs(s string) []v1.HostSubnetEgressCIDR {
	// edge case: When splitting "", Split will return a slice with a single
	// empty string, whereas we want a slice of length 0.
	if s == "" {
		return make([]v1.HostSubnetEgressCIDR, 0)
	}

	v := splitRe.Split(s, -1)
	sort.Strings(v)
	return stringsToEgressCIDRs(v)
}

func stringsToEgressCIDRs(s []string) []v1.HostSubnetEgressCIDR {
	out := make([]v1.HostSubnetEgressCIDR, len(s))
	for i := range s {
		out[i] = v1.HostSubnetEgressCIDR(s[i])
	}
	return out
}

func egressCIDRsToStrings(s []v1.HostSubnetEgressCIDR) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = string(s[i])
	}
	return out
}

func compare(this, other []v1.HostSubnetEgressCIDR) bool {
	if len(this) != len(other) {
		return false
	}

	for i := range this {
		if this[i] != other[i] {
			return false
		}
	}

	return true
}
