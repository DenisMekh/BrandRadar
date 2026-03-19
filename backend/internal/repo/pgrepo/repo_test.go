package pgrepo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type scannerStub struct {
	scanFn func(dest ...any) error
}

func (s scannerStub) Scan(dest ...any) error {
	return s.scanFn(dest...)
}

func TestConstructors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		fn   func() any
	}{
		{name: "brand", fn: func() any { return NewBrandRepo(nil) }},
		{name: "source", fn: func() any { return NewSourceRepo(nil) }},
		{name: "mention", fn: func() any { return NewMentionRepo(nil) }},
		{name: "alert", fn: func() any { return NewAlertRepo(nil) }},
		{name: "alert_config", fn: func() any { return NewAlertConfigRepo(nil) }},
		{name: "event", fn: func() any { return NewEventRepo(nil) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.NotNil(t, tc.fn())
		})
	}
}
