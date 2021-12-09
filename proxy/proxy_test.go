package proxy

import (
	"testing"
	"time"
)

func TestProxy_Run(t *testing.T) {
	stop := make(chan bool)
	go func() {
		proxy := Proxy{
			local:         "127.0.0.1:6379",
			remote:        "172.20.110.11:6379",
			maxRemoteConn: 10,
		}
		proxy.Run()
		<-stop
		proxy.Close()
	}()
	go func() {
		proxy := Proxy{
			local:         "127.0.0.1:3306",
			remote:        "172.20.110.11:3306",
			maxRemoteConn: 10,
		}
		proxy.Run()
		<-stop
		proxy.Close()
	}()
	go func() {
		time.Sleep(time.Second * 10)
		stop <- true
	}()

	<-stop
	time.Sleep(time.Second * 2)
}

func TestManager_Add(t *testing.T) {
	m := new(Manager)
	m.Add("127.0.0.1:3306", "172.20.110.11:3306", 10)
}
