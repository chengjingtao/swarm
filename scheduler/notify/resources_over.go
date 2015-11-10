package notify

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/eventemitter"
	"github.com/docker/swarm/scheduler/node"
	"github.com/spf13/viper"
	"strings"
	"sync"
	"time"
)

var lastNotifyTime time.Time

var m *sync.RWMutex

var NotifyInterval time.Duration

var (
	email_user     = "chengjtdebug@sina.com"
	email_pwd      = "*******"
	email_host     = "smtp.sina.com:25"
	email_to       = "*******"
	email_subject  = "docker 集群资源超限提醒"
	email_interval = 2
)

func init() {
	m = new(sync.RWMutex)

	eventemitter.On("/cluster/resources/over", resourceOver)
	log.Debugln("register  /cluster/resources/over ")
	NotifyInterval = time.Duration(email_interval) * time.Hour

	viper.SetConfigFile("config.json")
	viper.SetConfigType("json")
	var err = viper.ReadInConfig()
	if err != nil {
		log.Error("notify load config.js error  ,will using default values : " + err.Error())
		return
	}

	setEmailArgs()
	NotifyInterval = time.Duration(email_interval) * time.Hour
}

func setEmailArgs() {
	if str := viper.GetString("notify.email_user"); str != "" {
		email_user = str
	}
	if str := viper.GetString("notify.email_host"); str != "" {
		email_host = str
	}
	if str := viper.GetString("notify.email_pwd"); str != "" {
		email_pwd = str
	}
	if str := viper.GetString("notify.email_to"); str != "" {
		email_to = str
	}
	if str := viper.GetString("notify.email_subject"); str != "" {
		email_subject = str
	}
	if v := viper.GetInt("notify.email_interval"); v != 0 {
		email_interval = v
	}
}

func resourceOver(evt string, args interface{}) {
	log.Info("/cluster/resources/over event ")

	nodes, ok := args.([]*node.Node)
	if ok == false {
		log.Warn("/cluster/resources/over event, but args is not []*node.Node")
		return
	}

	if isNeedToNotify() == false {
		log.Infof("/cluster/resources/over event , last notify is less than  %d h ,  no need to notify.", email_interval)
		return
	}

	m.Lock()

	err := notify(nodes)
	if err != nil {
		log.Info("/cluster/resources/over event , notify error ," + err.Error())

		m.Unlock()
		return
	}

	lastNotifyTime = time.Now()
	m.Unlock()

	return
}

func isNeedToNotify() bool {
	m.RLock()
	defer m.RUnlock()

	if NotifyInterval.Nanoseconds() == 0 {
		log.Errorln("email notify fail , need to set  EMAIL_INTERVAL .")
		return false
	}

	//从未曾通知过
	if lastNotifyTime.IsZero() {
		return true
	}

	//曾经通知过,距离上次的通知时间已经超过了2小时
	if lastNotifyTime.IsZero() == false && time.Now().Sub(lastNotifyTime) > NotifyInterval {
		return true
	}

	return false
}

func notify(nodes []*node.Node) error {

	var data = "<ul>"

	for _, n := range nodes {
		data += "<li>" + n.Addr
		data += "<ul>"
		data += fmt.Sprintf("<li> <strong> Containers Count  : </strong>  %d / %d </li>", len(n.Containers_Start), len(n.Containers))
		data += fmt.Sprintf("<li> <strong> Cpus  : </strong>  %d / %d </li>", n.UsedCpus, n.TotalCpus)
		data += fmt.Sprintf("<li> <strong> Memory  : </strong>  %.2f G / %.2f G </li>", float64(n.UsedMemory)/(1024*1024*1024), float64(n.TotalMemory)/(1024*1024*1024))

		data += "<li><table><tr><th>Id</th><th>Name</th><th>Image</th><th>CpuShares</th><th>Memory</th><th>State</th></tr>"

		for _, c := range n.Containers {
			data += "<tr>"
			data += "<td>" + c.Container.Id + "</td>"
			data += "<td>" + strings.Join(c.Container.Names, ",") + "</td>"
			data += "<td>" + c.Container.Image + "</td>"
			data += "<td>" + fmt.Sprintf("%d", c.Config.CpuShares) + "</td>"
			data += "<td>" + fmt.Sprintf("%.2f Mb", float64(c.Config.Memory)/1024/1024) + "</td>"
			data += "<td>" + c.Info.State.StateString() + "</td>"
			data += "</tr>"
		}

		data += "</table></li>"
		data += "</ul></li>"
	}
	data += "</ul>"

	body := fmt.Sprintf(`
	<html>
		<body>
			<h2>docker 集群超限提醒</h2>
			%s
		</body>
	</html>
	`, data)

	log.Debugf("email_user : %s , email_host: %s , email_to: %s , email_subject : %s  ,email_interval: %d ", email_user, email_host, email_to, email_subject, email_interval)

	var err = sendEmail(email_user, email_pwd, email_host, email_to, email_subject, body, "html")
	if err != nil {
		log.Debug("email notify fail  , " + err.Error())
		return err
	}
	log.Debug("email notify success")
	return nil
}
