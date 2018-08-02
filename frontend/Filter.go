package frontend

import (
	"net"
	"encoding/binary"
	"sync"
)

type IIPFilter interface {
	AllowIP(ip string)
	BlockIP(ip string)
	IsAllow(ip string) bool
}

func convertIPv4ToUint(ip net.IP) uint32 {
	if len (ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16]);
	}
	return binary.BigEndian.Uint32(ip)
}
func convertUintToIPv4(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}

type implIPFilter struct {
	IIPFilter
	AllowdIPs  []string
	BlockedIPs []string
	SubNets    []*subnet
	defaultAllowed bool

	mut sync.RWMutex
	IPs map[string]bool
}

type subnet struct {
	cidr    string
	allowed bool
	net *net.IPNet
}

func NewIPFilter(defaultAllowed bool) (IIPFilter) {
	var flt implIPFilter
	flt.IPs = make(map[string]bool)
	flt.defaultAllowed = defaultAllowed
	return &flt
}

func (this *implIPFilter) toggleIP(s string, allowed bool) (bool) {
	ip, snet, err := net.ParseCIDR(s);
	if err == nil { // 包含子网
		if n, total := snet.Mask.Size(); n == total {
			this.mut.Lock()
			this.IPs[ip.String()] = allowed
			this.mut.Unlock()
			return true
		}
		this.mut.Lock()
		found := false
		for _, subnet := range this.SubNets {
			if subnet.cidr == s {
				found = true
				subnet.allowed = allowed
			}
		}
		if !found {
			this.SubNets = append(this.SubNets, &subnet{cidr: s, allowed: allowed, net:snet})
		}
		this.mut.Unlock()
		return true
	}
	// Host IP
	if ip := net.ParseIP(s); ip != nil {
		this.mut.Lock()
		this.IPs[ip.String()] = allowed
		this.mut.Unlock()
		return true
	}

	return false
}

func (this *implIPFilter) AllowIP(ip string) {
	this.toggleIP(ip, true)
}

func (this *implIPFilter) BlockIP(ip string) {
	this.toggleIP(ip, false)
}

func (this *implIPFilter) IsAllow(ip string) bool {
	return this.isNetAllow(net.ParseIP(ip))
}

func (this *implIPFilter) isNetAllow(ip net.IP) bool {
	if ip == nil {
		return false
	}
	this.mut.Lock()
	defer this.mut.Unlock()

	allowed, ok := this.IPs[ip.String()]
	if ok {
		return allowed
	}

	for _, subnet := range this.SubNets {
		if subnet.net.Contains(ip) {
			return subnet.allowed
		}
	}

	return this.defaultAllowed
}