package cli

import (
	"os"
	"regexp"
	"time"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/swarm/discovery"
)

func checkAddrFormat(addr string) bool {
	m, _ := regexp.MatchString("^[0-9a-zA-Z._-]+:[0-9]{1,5}$", addr)
	return m
}

//join  --addr 172.16.150.11:2375 --weight -1    172.16.150.11:2375/weight/-1
//join  --addr 172.16.150.11:2375

func join(c *cli.Context) {
	dflag := getDiscovery(c)
	if dflag == "" {
		log.Fatalf("discovery required to join a cluster. See '%s join --help'.", c.App.Name)
	}

	addr := c.String("advertise")
	if addr == "" {
		log.Fatal("missing mandatory --advertise flag")
	}
	if !checkAddrFormat(addr) {
		log.Fatal("--advertise should be of the form ip:port or hostname:port")
	}

	hb, err := time.ParseDuration(c.String("heartbeat"))
	if err != nil {
		log.Fatalf("invalid --heartbeat: %v", err)
	}
	if hb < 1*time.Second {
		log.Fatal("--heartbeat should be at least one second")
	}
	ttl, err := time.ParseDuration(c.String("ttl"))
	if err != nil {
		log.Fatalf("invalid --ttl: %v", err)
	}
	if ttl <= hb {
		log.Fatal("--ttl must be strictly superior to the heartbeat value")
	}

	d, err := discovery.New(dflag, hb, ttl, getDiscoveryOpt(c))
	if err != nil {
		log.Fatal(err)
	}

	for {
        var refreshedAddr = refreshURL(addr)
		log.WithFields(log.Fields{"addr": refreshedAddr, "discovery": dflag}).Infof("Registering on the discovery service every %s...", hb)

		if err := d.Register(refreshedAddr); err != nil {
			log.Error(err)
		}

		time.Sleep(hb)
	}
}

//如果环境变量有值,则以环境变量为准
//如果环境变量没值,则以加入时的状态为准
func refreshURL(joinedUrl string) string {
	var envValue = strings.TrimSpace(os.Getenv("SWARM_JOIN_WEIGHT"))

	if envValue != "" {
		return appendWeightValue(joinedUrl, envValue)
	}

	return joinedUrl
}

func appendWeightValue(url string, value string) string {

	if strings.Index(url, "/weight/") == -1 {

		if !strings.HasSuffix(url, "/") {
			url = url + "/"
		}

		if !strings.HasSuffix(url, "weight/") {
			url = url + "weight/"
		}

		url = url + value
		return url
	}

	var segments = strings.Split(url, "/")
	var index = -1

	for i, item := range segments {
		if item == "weight" {
			index = i
			break
		}
	}

	segments[index+1]=value
    return strings.Join(segments,"/")
}
