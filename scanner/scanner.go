package scanner

import (
	"errors"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/tarmalonchik/golibs/trace"
)

func RunAndFilter(addr netip.Addr, port, parallel int, timeout time.Duration) []string {
	prefix, err := addr.Prefix(24)
	if err != nil {
		panic(err)
	}

	out, err := runScanner(prefix.String(), port, parallel, timeout)
	if err != nil {
		panic(err)
	}

	sort.Strings(out)

	return lo.Filter(out, func(item string, index int) bool {
		if strings.Contains(item, "beget") {
			return false
		}
		return true
	})
}

func runScanner(addr string, port, parallel int, timeout time.Duration) ([]string, error) {
	if _, _, err := net.ParseCIDR(addr); err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "parsing CIDR")
	}

	if parallel < 1 || timeout < 1*time.Second {
		return nil, trace.FuncNameWithError(errors.New("invalid inputs"))
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	var hostChan <-chan Host

	hostChan = iterateAddr(addr)

	outCh := make(chan string, 1000)

	once := sync.OnceFunc(func() {
		close(outCh)
	})

	defer func() {
		once()
	}()

	var wg sync.WaitGroup

	wg.Add(parallel)

	for i := 0; i < parallel; i++ {
		go func() {
			for ip := range hostChan {
				scanTLS(ip, outCh, port, timeout)
			}
			wg.Done()
		}()
	}

	t := time.Now()

	slog.Info("Started all scanning threads", "time", t)

	wg.Wait()

	once()

	slog.Info("Scanning completed", "time", time.Now(), "elapsed", time.Since(t).String())

	out := make([]string, 0, len(outCh))
	for i := range outCh {
		out = append(out, i)
	}
	return out, nil
}
