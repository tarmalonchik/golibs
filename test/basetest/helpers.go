package basetest

import (
	"time"

	"github.com/stretchr/testify/require"
)

func RunWithTimeout(s *require.Assertions, timeout time.Duration, f func()) {
	start := time.Now()

	f()

	if time.Since(start) > timeout {
		s.Fail("failed due to timeout")
	}
}
