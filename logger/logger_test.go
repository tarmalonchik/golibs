package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_Level(t *testing.T) {
	for _, val := range LevelValues() {
		_, err := zap.ParseAtomicLevel(val.String())
		require.NoError(t, err)
	}
}
