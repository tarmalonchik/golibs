package kafka

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestGetPartitioner(t *testing.T) {
	topic := "vpnchik-core-to-vpn-action"

	out, err := GetPartition(topic, "hereShouldBeIP", 10, 100)
	require.NoError(t, err)
	require.Equal(t, int32(3), out)
}

func TestGenerateKey(t *testing.T) {
	out, err := GenerateKey("someKe", -1)
	require.Error(t, err)

	v := make(map[int32]int)

	for range 500 {
		rd := lo.RandomString(10, []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g'})
		out, err = GenerateKey(rd, 10)
		require.NoError(t, err)
		require.GreaterOrEqual(t, out, int32(0))
		require.Less(t, out, int32(10))
		v[out]++
	}

	for i := range 10 {
		require.Greater(t, v[int32(i)], 10)
	}
}
