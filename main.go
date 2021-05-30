package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/tomhomekit/dnsproxy/internal"
)

var (
	client             *http.Client
	upStream           string
	homeIp             string
	mutex              sync.RWMutex
	homeDomain         string
	localServerAddress string
)

const ipURL = "http://ip.clearcode.cn/kv?key=ip"

func init() {
	flag.StringVar(&upStream, "s", "114.114.114.114:53,223.5.5.5:53", "upstream dns servers")
	flag.StringVar(&localServerAddress, "a", "0.0.0.0:53", "hosts")
	flag.StringVar(&homeDomain, "h", "mrj.com", "hosts")
}

func initHttpClient() {
	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				up0 := strings.Split(upStream, ",")[0]
				return net.Dial(network, up0)
			},
		},
	}

	client = &http.Client{
		Transport: &http.Transport{
			DialContext: nil,
			Dial:        dialer.Dial,
			DialTLS: func(network, addr string) (net.Conn, error) {
				return tls.DialWithDialer(dialer, network, addr, nil)
			},
		},
	}
}

func getIpRemote() {
	resp, err := client.Get(ipURL)
	if err != nil {
		logrus.Error(err)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.WithField("ip", string(data)).Info("get home ip")
	mutex.Lock()
	homeIp = string(data)
	mutex.Unlock()
}

func getIp() string {
	mutex.RLock()
	ip := homeIp
	mutex.RUnlock()
	return ip
}

func main() {
	flag.Parse()
	initHttpClient()
	getIpRemote()
	go func() {
		tk := time.NewTicker(1 * time.Minute)
		for range tk.C {
			getIpRemote()
		}
	}()
	go startLocalDnsServer()
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill, os.Interrupt)
	<-ch
}

func startLocalDnsServer() {
	internal.OnQuestion = func(q dns.Question, w dns.ResponseWriter, r *dns.Msg) bool {
		name := string(q.Name)
		name = name[:len(name)-1]
		if !strings.HasSuffix(name, homeDomain) {
			return false
		}
		x := fmt.Sprintf("%s IN A %s", q.Name, getIp())
		rr, err := dns.NewRR(x)
		if err != nil {
			log.Println(err)
			return true
		}
		m := new(dns.Msg)
		m.SetReply(r)
		m.Id = r.Id
		m.Answer = []dns.RR{
			rr,
		}
		m.Response = true
		err = w.WriteMsg(m)
		if err != nil {
			log.Println(err)
			return false
		}
		logrus.WithField("resolve", q.Name).WithField("ip", getIp()).Info("resolve home address")
		return true
	}
	ups := strings.Split(upStream, ",")
	internal.StartDNSServer(localServerAddress, ups...)
}
