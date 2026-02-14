package scanner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/big"
	"net"
	"net/netip"
	"regexp"
	"strings"
)

const (
	_ = iota
	HostTypeIP
	HostTypeCIDR
	HostTypeDomain
)

type HostType int

type Host struct {
	IP     net.IP
	Origin string
	Type   HostType
}

func iterate(reader io.Reader) chan Host {
	scanner := bufio.NewScanner(reader)
	hostChan := make(chan Host)
	go func() {
		defer close(hostChan)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			ip := net.ParseIP(line)
			if ip != nil && ip.To4() != nil {
				hostChan <- Host{
					IP:     ip,
					Origin: line,
					Type:   HostTypeIP,
				}
				continue
			}
			_, _, err := net.ParseCIDR(line)
			if err == nil {
				// ip cidr
				p, err := netip.ParsePrefix(line)
				if err != nil {
					slog.Warn("Invalid cidr", "cidr", line, "err", err)
				}

				if !p.Addr().Is4() {
					continue
				}

				p = p.Masked()

				addr := p.Addr()

				for {
					if !p.Contains(addr) {
						break
					}
					ip = net.ParseIP(addr.String())
					if ip != nil {
						hostChan <- Host{
							IP:     ip,
							Origin: line,
							Type:   HostTypeCIDR,
						}
					}
					addr = addr.Next()
				}
				continue
			}
			if validateDomainName(line) {
				hostChan <- Host{
					IP:     nil,
					Origin: line,
					Type:   HostTypeDomain,
				}
				continue
			}
			slog.Warn("Not a valid IP, IP CIDR or domain", "line", line)
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
			slog.Error("Read file error", "err", err)
		}
	}()
	return hostChan
}

func validateDomainName(domain string) bool {
	r := regexp.MustCompile(`(?m)^[A-Za-z0-9\-.]+$`)
	return r.MatchString(domain)
}

func iterateAddr(addr string) <-chan Host {
	hostChan := make(chan Host)

	_, _, err := net.ParseCIDR(addr)
	if err == nil {
		return iterate(strings.NewReader(addr))
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		ip, err = lookupIP(addr)
		if err != nil {
			close(hostChan)
			slog.Error("Not a valid IP, IP CIDR or domain", "addr", addr)
			return hostChan
		}
	}
	go func() {
		slog.Info("Enable infinite mode", "init", ip.String())
		lowIP := ip
		highIP := ip
		hostChan <- Host{
			IP:     ip,
			Origin: addr,
			Type:   HostTypeIP,
		}
		for i := 0; i < math.MaxInt; i++ {
			if i%2 == 0 {
				lowIP = nextIP(lowIP, false)
				hostChan <- Host{
					IP:     lowIP,
					Origin: lowIP.String(),
					Type:   HostTypeIP,
				}
			} else {
				highIP = nextIP(highIP, true)
				hostChan <- Host{
					IP:     highIP,
					Origin: highIP.String(),
					Type:   HostTypeIP,
				}
			}
		}
	}()
	return hostChan
}

func lookupIP(addr string) (net.IP, error) {
	ips, err := net.LookupIP(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup: %w", err)
	}
	var arr []net.IP
	for _, ip := range ips {
		if ip.To4() != nil {
			arr = append(arr, ip)
		}
	}
	if len(arr) == 0 {
		return nil, errors.New("no IP found")
	}
	return arr[0], nil
}

func nextIP(ip net.IP, increment bool) net.IP {
	// Convert to big.Int and increment
	ipb := big.NewInt(0).SetBytes(ip)
	if increment {
		ipb.Add(ipb, big.NewInt(1))
	} else {
		ipb.Sub(ipb, big.NewInt(1))
	}

	// Add leading zeros
	b := ipb.Bytes()
	b = append(make([]byte, len(ip)-len(b)), b...)
	return b
}
