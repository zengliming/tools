package proxy

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Manager struct {
	configs []string
	/// 所有的代理对象
	proxies map[string]*Proxy
	/// 初始化标识
	init sync.Once
	/// 启动标识
	start sync.Once
}

func (m *Manager) Add(local string, remote string, maxRemoteConn uint, autoStart bool) {
	if _, ok := m.proxies[local]; ok {
		/// 存在一个
		return
	}
	m.init.Do(func() {
		m.proxies = make(map[string]*Proxy, 1024)
	})
	proxy := &Proxy{
		local:         local,
		remote:        remote,
		maxRemoteConn: maxRemoteConn,
		autoStart:     autoStart,
	}
	if proxy.autoStart {
		proxy.Run()
	}
	m.proxies[local] = proxy
}

func (m *Manager) Start() {
	m.start.Do(func() {
		for s := range m.proxies {
			p := m.proxies[s]
			if p.autoStart {
				p.Run()
			}
		}
	})
}

func (m *Manager) Run(local string) {
	p := m.proxies[local]
	if p.autoStart {
		p.Run()
	}
}

type Proxy struct {
	local             string
	remote            string
	listener          net.Listener
	maxRemoteConn     uint
	currentRemoteConn uint
	stops             []chan bool
	remoteConnects    []net.Conn
	run               sync.Once
	autoStart         bool
}

func (p *Proxy) Run() {
	p.run.Do(func() {
		listener, err := net.Listen("tcp", p.local)
		if err != nil {
			log.Println("", err)
			return
		}
		p.listener = listener
		p.stops = make([]chan bool, p.maxRemoteConn+1)
		for {
			conn, err := p.listener.Accept()
			if err != nil {
				log.Println("", err)
				return
			}
			go p.handlerConn(conn)
		}
	})

}

func (p *Proxy) handlerConn(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("关闭本地连接异常:", err)
		}
	}(conn)
	if p.currentRemoteConn >= p.maxRemoteConn {
		log.Println("超过最大数量")
	}
	d := net.Dialer{Timeout: time.Second * 10}
	remote, err := d.Dial("tcp", p.remote)
	if err != nil {
		log.Println("", err)
		return
	}
	if p.remoteConnects == nil {
		p.remoteConnects = make([]net.Conn, p.maxRemoteConn)
	}
	p.remoteConnects = append(p.remoteConnects, remote)
	p.currentRemoteConn++
	defer func(remote net.Conn) {
		err := remote.Close()
		if err != nil {
			log.Println("关闭远程连接异常:", err)
		}
	}(remote)
	stop := make(chan bool)
	p.stops = append(p.stops, stop)
	go func() {
		_, err := io.Copy(conn, remote)
		if err != nil {
			return
		}
		<-stop
	}()

	go func() {
		_, err := io.Copy(remote, conn)
		if err != nil {
			return
		}
		<-stop
	}()
	<-stop
	p.currentRemoteConn--
}

func (p *Proxy) Close() {
	log.Println("收到关闭当前连接消息")
	for _, stop := range p.stops {
		stop <- true
	}
}
