package port_scan

import (
	"fmt"
	"net"
	"sync"
)

type Scan struct {
	ip        string
	beginPort uint
	endPort   uint
	waitGroup sync.WaitGroup
}

func New(ip string, beginPort uint, endPort uint) *Scan {
	return &Scan{
		ip:        ip,
		beginPort: beginPort,
		endPort:   endPort,
	}
}

func (scan *Scan) Run() {
	for i := scan.beginPort; i <= scan.endPort; i++ {
		scan.waitGroup.Add(1)
		go func(i uint) {
			defer scan.waitGroup.Done()
			var address = fmt.Sprintf("%s:%d", scan.ip, i)
			conn, err := net.Dial("tcp", address)
			if err != nil {
				return
			}
			conn.Close()
			fmt.Println(address, "打开")
		}(i)
	}
	scan.waitGroup.Wait()
}
