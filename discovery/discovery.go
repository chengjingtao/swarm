package discovery

import (
	"errors"
	"fmt"
	"net"
    URL "net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// An Entry represents a swarm host.
type Entry struct {
	Host string
	Port string

	// -1 表示 备用机器
	// 0 表示 标记为退出集群(将不再使用向该机器上分配资源)
	// 1 默认值为1 表示默认权重
	// 其他整数值 表示分配权重
	Weight int
}

const defaultWeight = 1

// NewEntry creates a new entry.
func NewEntry(url string) (*Entry, error) {

	//	host, port, err := net.SplitHostPort(url)
	//	log.Debugf("NewEntry() host is %s ,port is %s", host, port)
	//	if err != nil {
	//		return nil, err
	//	}
	var host, port, w, err = getHostAndPortAndWeight(url)

	if err != nil {
		return nil, err
	}
	return &Entry{host, port, w}, nil
}

//url = 192.168.5.55:2375?weight=1
func getHostAndPortAndWeight(url string) (host string, port string, w int, ers error) {
    
	u, err := URL.Parse("http://" + url)
	if err != nil {
		//log.Warnf("%s url Parse error %s, 使用默认值 %d", url, err.Error(), defaultWeight)
		return "", "", defaultWeight, err
	}
    
	host, port, err = net.SplitHostPort(u.Host)

	if err != nil {
		//log.Warnf("%s weight 值为空, 使用默认值 %d", url, defaultWeight)
		return "", "", defaultWeight, err
	}
    
    if len(u.Query()["weight"])==0{
		//log.Warnf("%s  weight 值为空, 使用默认值 %d", url, defaultWeight)
		return host, port, defaultWeight, nil
    }
    

	var value = u.Query()["weight"][0]
	value = strings.TrimSpace(value)
	if value == "" {
		//log.Warnf("%s  weight 值为空, 使用默认值 %d", url, defaultWeight)
		return host, port, defaultWeight, nil
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		//log.Warnf("%s weight 值有误, 使用默认值 %d", url, defaultWeight)
		return host, port, defaultWeight, nil
	}

	//log.Infof("%s  weight 值为 %d", url, v)
	return host, port, v, nil
}

// String returns the string form of an entry.
func (e *Entry) String() string {
	return fmt.Sprintf("%s:%s", e.Host, e.Port)
}

// Equals returns true if cmp contains the same data.
func (e *Entry) Equals(cmp *Entry) bool {
	return e.Host == cmp.Host && e.Port == cmp.Port
}

// NeedUpdate 判断源Entry 是否要根据cmp 进行更新
func (e *Entry) NeedUpdate(cmp *Entry) bool{
    return e.Equals(cmp) && e.Weight!=cmp.Weight
}

// Update entry expect the field of host and port
func (e *Entry) Update(cmp *Entry){
    e.Weight=cmp.Weight
}

// Entries is a list of *Entry with some helpers.
type Entries []*Entry

// Equals returns true if cmp contains the same data.
func (e Entries) Equals(cmp Entries) bool {
	// Check if the file has really changed.
	if len(e) != len(cmp) {
		return false
	}
	for i := range e {
		if !e[i].Equals(cmp[i]) {
			return false
		}
	}
	return true
}

// Contains returns true if the Entries contain a given Entry.
func (e Entries) Contains(entry *Entry) bool {
	for _, curr := range e {
		if curr.Equals(entry) {
			return true
		}
	}
	return false
}
// Update  will update e by entry if item are equal
func (e Entries) Update(entry *Entry){
	for _, curr := range e {
		if curr.Equals(entry) && curr.NeedUpdate(entry) {
			curr.Update(entry)
		}
	}
}

// NeedUpdate  判断集合中于entry相同的元素是否需要更新
func (e Entries) NeedUpdate(entry *Entry) bool{
	for _, curr := range e {
		if curr.Equals(entry) {
			return curr.NeedUpdate(entry)
		}
	}
    
    return false
}

// Diff compares two entries and returns the added and removed entries.
func (e Entries) Diff(cmp Entries) (Entries, Entries,Entries) {
	added := Entries{}
    updated := Entries{}
    
	for _, entry := range cmp {
		if !e.Contains(entry) {
			added = append(added, entry)
		}else{
            if e.NeedUpdate(entry){
                updated=append(updated,entry)
            }
        }
	}

	removed := Entries{}
	for _, entry := range e {
		if !cmp.Contains(entry) {
			removed = append(removed, entry)
		}
	}

	return added, removed,updated
}

// The Discovery interface is implemented by Discovery backends which
// manage swarm host entries.
type Discovery interface {
	// Initialize the discovery with URIs, a heartbeat, a ttl and optional settings.
	Initialize(string, time.Duration, time.Duration, map[string]string) error

	// Watch the discovery for entry changes.
	// Returns a channel that will receive changes or an error.
	// Providing a non-nil stopCh can be used to stop watching.
	Watch(stopCh <-chan struct{}) (<-chan Entries, <-chan error)

	// Register to the discovery

	Register(string) error
    
    RegisterWithData(string,map[string]string) error
}

var (
	discoveries map[string]Discovery
	// ErrNotSupported is returned when a discovery service is not supported.
	ErrNotSupported = errors.New("discovery service not supported")
	// ErrNotImplemented is returned when discovery feature is not implemented
	// by discovery backend.
	ErrNotImplemented = errors.New("not implemented in this discovery service")
)

func init() {
	discoveries = make(map[string]Discovery)
}

// Register makes a discovery backend available by the provided scheme.
// If Register is called twice with the same scheme an error is returned.
func Register(scheme string, d Discovery) error {
	if _, exists := discoveries[scheme]; exists {
		return fmt.Errorf("scheme already registered %s", scheme)
	}
	log.WithField("name", scheme).Debug("Registering discovery service")
	discoveries[scheme] = d

	return nil
}

func parse(rawurl string) (string, string) {
	parts := strings.SplitN(rawurl, "://", 2)

	// nodes:port,node2:port => nodes://node1:port,node2:port
	if len(parts) == 1 {
		return "nodes", parts[0]
	}
	return parts[0], parts[1]
}

// New returns a new Discovery given a URL, heartbeat and ttl settings.
// Returns an error if the URL scheme is not supported.
func New(rawurl string, heartbeat time.Duration, ttl time.Duration, discoveryOpt map[string]string) (Discovery, error) {
	scheme, uri := parse(rawurl)
    
	if discovery, exists := discoveries[scheme]; exists {
		log.WithFields(log.Fields{"name": scheme, "uri": uri}).Debug("Initializing discovery service")
		err := discovery.Initialize(uri, heartbeat, ttl, discoveryOpt)
		return discovery, err
	}

	return nil, ErrNotSupported
}

// CreateEntries returns an array of entries based on the given addresses.
func CreateEntries(addrs []string) (Entries, error) {
	entries := Entries{}
	if addrs == nil {
		return entries, nil
	}

	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		entry, err := NewEntry(addr)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
