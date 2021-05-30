package internal

import (
	"context"
	"log"
	"net"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

var (
	cache, _ = lru.New(10240)
)

var OnQuestion func(question dns.Question, w dns.ResponseWriter, r *dns.Msg) bool

func StartDNSServer(addr string, upstream ...string) {
	dns.HandleFunc(".", route(upstream...))
	log.Println(dns.ListenAndServe(addr, "udp", nil))
}

func route(upstream ...string) dns.HandlerFunc {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		qs := r.Question
		if len(qs) == 0 {
			return
		}
		q := r.Question[0]
		if OnQuestion != nil && OnQuestion(q, w, r) {
			return
		}
		key := r.String()
		if v, ok := cache.Get(key); ok {
			w.WriteMsg(v.(*dns.Msg))
			logrus.WithField("cache host", q.Name).WithField("type", dns.TypeToString[q.Qtype]).Info("exchange host")
			return
		}
		var (
			msg *dns.Msg
			err error
		)
		for i := 0; i < 3; i++ {
			for _, ups := range upstream {
				msg, err = dns.Exchange(r, ups)
				if err != nil {
					continue
				}
			}
			if err != nil {
				continue
			}
		}

		if err != nil {
			m := new(dns.Msg)
			m.SetRcode(r, dns.RcodeServerFailure)
			w.WriteMsg(m)
			logrus.Error(err)
		} else {
			cache.Add(key, msg)
			w.WriteMsg(msg)
			logrus.WithField("host", q.Name).WithField("type", dns.TypeToString[q.Qtype]).Info("exchange host")
		}
	}
}

// NewHttpClient with default resolver
func NewHttpClient(network string, resolverAddress string) *http.Client {
	client := http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Resolver: &net.Resolver{
					PreferGo:     true,
					StrictErrors: false,
					Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
						return net.Dial("udp", resolverAddress)
					},
				},
			}).Dial,
		},
	}

	return &client
}
