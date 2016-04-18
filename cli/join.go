package cli

import (
	"os"
	"regexp"
    "io/ioutil"
	"time"
    URL "net/url"
    "os/exec"
	"strings"
    "path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/swarm/discovery"
)

func checkAddrFormat(addr string) bool {
	m, _ := regexp.MatchString("^[0-9a-zA-Z._-]+:[0-9]{1,5}", addr)
	return m
}

//join  --addr 172.16.150.11:2375 172.16.150.11:2375?weight=1
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
    
    var url=addr
    var nodeAddr = trueAddr(url)

	for {
        var data=refreshData(url)
		log.WithFields(log.Fields{"addr": nodeAddr, "discovery": dflag}).Infof("Registering on the discovery service every %s...", hb)

		if err := d.RegisterWithData(nodeAddr,data); err != nil {
			log.Error(err)
		}

		time.Sleep(hb)
	}
}

func trueAddr(url string) string{

    u,_ := URL.Parse("http://"+url)
    
    return u.Host
}

//refreshData 如果环境变量有值,则以环境变量为准
//如果环境变量没值,则以加入时的状态为准
func refreshData(joinedUrl string) map[string]string {
	var setValue = strings.TrimSpace(weight())
    u,_ := URL.Parse("http://"+joinedUrl)
    var query=u.Query()
    
	if setValue != "" {
        query.Set("weight",setValue)
		return parseValuesToData(query)
	}

	return parseValuesToData(query)
}

func parseValuesToData(vs URL.Values) map[string]string {
    
    var data=map[string]string{}
    for k,v :=range vs{
        data[k]=v[0]
    }
    return data
}

func setWeightValue(url string, value string) string {

    u,_ := URL.Parse("http://"+url)
    var query = u.Query()
    query.Set("weight",value)
    
    var res = u.Host+"?"+query.Encode()
    return res
}

func weight() string{
    p,err := exec.LookPath(os.Args[0])
    if err!=nil{
       log.Error("LookPath error ",err)
       os.Exit(1) 
    }
    absPath,err := filepath.Abs(p)
    if err!=nil{
       log.Errorf("Abs path-> %s error  %s",p, err.Error())
       os.Exit(1)
    }
    
    segments := strings.Split(absPath,"/")
    segments[len(segments)-1]="join-weight"
    path := strings.Join(segments,"/")
    
    f,err := os.Open(path)
    if err!=nil{
        if os.IsNotExist(err){
            if f,err = os.OpenFile(path,os.O_CREATE,0666);err!=nil{
                log.Error("创建.join-weight 失败",err)
                return ""   
            }
        }else{
            log.Error(".join-weight Open失败",err)
            return ""
        }
    }
    
    if bts,err := ioutil.ReadAll(f);err==nil{
        return string(bts)
    }
    log.Error(".join-weight read error ",err.Error())
    return ""
}
