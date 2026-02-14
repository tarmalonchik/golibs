package scanner

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func scanTLS(host Host, out chan<- string, port int, timeout time.Duration) {
	if host.IP == nil {
		ip, err := lookupIP(host.Origin)
		if err != nil {
			slog.Debug("Failed to get IP from the origin", "origin", host.Origin, "err", err)
			return
		}
		host.IP = ip
	}
	hostPort := net.JoinHostPort(host.IP.String(), strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", hostPort, timeout)
	if err != nil {
		slog.Debug("Cannot dial", "target", hostPort)
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		slog.Error("Error setting deadline", "err", err)
		return
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
		CurvePreferences:   []tls.CurveID{tls.X25519},
	}

	if host.Type == HostTypeDomain {
		tlsCfg.ServerName = host.Origin
	}

	c := tls.Client(conn, tlsCfg)
	defer func() {
		_ = c.Close()
	}()

	if err = c.Handshake(); err != nil {
		slog.Debug("TLS handshake failed", "target", hostPort)
		return
	}

	state := c.ConnectionState()
	alpn := state.NegotiatedProtocol
	domain := state.PeerCertificates[0].Subject.CommonName
	issuers := strings.Join(state.PeerCertificates[0].Issuer.Organization, " | ")

	log := slog.Info

	feasible := true

	if state.Version != tls.VersionTLS13 || alpn != "h2" || len(domain) == 0 || len(issuers) == 0 {
		log = slog.Debug
		feasible = false
	} else {
		if isValidDomain(domain) {
			fmt.Println(domain)
			out <- domain
		}
	}

	log("Connected to target", "feasible", feasible, "ip", host.IP.String(),
		"origin", host.Origin,
		"tls", tls.VersionName(state.Version), "alpn", alpn, "cert-domain", domain, "cert-issuer", issuers)
}

func isValidDomain(domain string) bool {
	if strings.Contains(domain, "*") {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}

	resp, err := http.Get("https://" + domain)
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return false
	}

	return true
}
