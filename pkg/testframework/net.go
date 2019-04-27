package testframework

import (
	"errors"
	"net"
	"strconv"
)

var ErrNoDefaultIP = errors.New("no suitable IP address")

func DefaultLocalIP4() (net.IP, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	devices, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, dev := range devices {
		if (dev.Flags&net.FlagUp != 0) && (dev.Flags&net.FlagLoopback == 0) {
			addrs, err := dev.Addrs()
			if err != nil {
				continue
			}
			for i := range addrs {
				if ip, ok := addrs[i].(*net.IPNet); ok {
					if ip.IP.To4() != nil {
						return ip.IP, nil
					}
				}
			}
		}
	}
	return nil, ErrNoDefaultIP
}
func FindFreeLocalPort() (int, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()
	_, portStr, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return port, nil
}
