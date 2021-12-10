package port_scan

import "testing"

func TestScan_Run(t *testing.T) {
	scan := New("172.18.36.24", 22, 9999)
	scan.Run()
}
