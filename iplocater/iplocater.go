package iplocater

import (
	"net"
)

type Location struct {
	ID   uint64
	Name string
}

type ISP struct {
	ID   uint32
	Name string
}

type IPLocater interface {
	Locate(ip net.IP) (country, province, city *Location, isp []*ISP, err error)
}
