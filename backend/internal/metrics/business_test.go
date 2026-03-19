package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewBusiness_AndMethods_DoNotPanic(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	m := NewBusiness(reg)

	require.NotNil(t, m)

	testCases := []struct {
		name string
		fn   func()
	}{
		{
			name: "IncMentionsCreated",
			fn: func() {
				m.IncMentionsCreated("brand-1", "web")
			},
		},
		{
			name: "IncAlertsFired",
			fn: func() {
				m.IncAlertsFired("brand-1")
			},
		},
		{
			name: "ObserveIngestDuration",
			fn: func() {
				m.ObserveIngestDuration(0.42)
			},
		},
		{
			name: "AddIngestItems_positive",
			fn: func() {
				m.AddIngestItems(StatusSuccess, 3)
			},
		},
		{
			name: "AddIngestItems_zero",
			fn: func() {
				m.AddIngestItems(StatusSkipped, 0)
			},
		},
		{
			name: "SetSourcesActive",
			fn: func() {
				m.SetSourcesActive("web", 5)
			},
		},
		{
			name: "SetBrandsTotal",
			fn: func() {
				m.SetBrandsTotal(12)
			},
		},
		{
			name: "IncAPIErrors",
			fn: func() {
				m.IncAPIErrors("/api/v1/brands", "500")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.fn)
		})
	}

	families, err := reg.Gather()
	require.NoError(t, err)
	require.Len(t, families, 7)
}

func TestNopBusiness_DoNotPanic(t *testing.T) {
	t.Parallel()

	m := NopBusiness()
	require.NotNil(t, m)

	testCases := []struct {
		name string
		fn   func()
	}{
		{
			name: "IncMentionsCreated",
			fn: func() {
				m.IncMentionsCreated("brand-1", "telegram")
			},
		},
		{
			name: "IncAlertsFired",
			fn: func() {
				m.IncAlertsFired("brand-1")
			},
		},
		{
			name: "ObserveIngestDuration",
			fn: func() {
				m.ObserveIngestDuration(1.1)
			},
		},
		{
			name: "AddIngestItems",
			fn: func() {
				m.AddIngestItems(StatusError, 10)
			},
		},
		{
			name: "SetSourcesActive",
			fn: func() {
				m.SetSourcesActive("telegram", 2)
			},
		},
		{
			name: "SetBrandsTotal",
			fn: func() {
				m.SetBrandsTotal(3)
			},
		},
		{
			name: "IncAPIErrors",
			fn: func() {
				m.IncAPIErrors("/api/v1/sources", "400")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.fn)
		})
	}
}
