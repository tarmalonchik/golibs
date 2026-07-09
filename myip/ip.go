package myip

import (
	"io"
	"net/http"
	"net/netip"
)

type Config struct {
	Debug bool `env:"DEBUG" envDefault:"false"`
}

func My(config Config) netip.Addr {
	if config.Debug {
		return netip.MustParseAddr("72.56.232.186")
	}

	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return netip.MustParseAddr(string(ip))
}
