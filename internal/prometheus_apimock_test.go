package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/api"
	promHttpC "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func NewPromApiMock() PromApiMock {
	m := PromApiMock{}
	m.SetQueryRangeOutput(nil, nil, nil)
	f := func(ctx context.Context, query string, r promHttpC.Range) {}
	m.SetQueryRangeCheckFunc(f)
	return m
}

type PromApiMock struct {
	// QueryRange
	queryRangeOutValue    model.Value
	queryRangeOutWarnings api.Warnings
	queryRangeOutError    error
	queryRangeCheckFunc   func(ctx context.Context, query string, r promHttpC.Range)
}

func (m *PromApiMock) SetQueryRangeOutput(v model.Value, w api.Warnings, err error) {
	m.queryRangeOutValue = v
	m.queryRangeOutWarnings = w
	m.queryRangeOutError = err
}

func (m *PromApiMock) SetQueryRangeCheckFunc(f func(ctx context.Context, query string, r promHttpC.Range)) {
	m.queryRangeCheckFunc = f
}

func (m PromApiMock) QueryRange(ctx context.Context, query string, r promHttpC.Range) (model.Value, api.Warnings, error) {
	m.queryRangeCheckFunc(ctx, query, r)
	return m.queryRangeOutValue, m.queryRangeOutWarnings, m.queryRangeOutError
}

func (m PromApiMock) Alerts(ctx context.Context) (promHttpC.AlertsResult, error) {
	return promHttpC.AlertsResult{}, nil
}

func (m PromApiMock) AlertManagers(ctx context.Context) (promHttpC.AlertManagersResult, error) {
	return promHttpC.AlertManagersResult{}, nil
}

func (m PromApiMock) CleanTombstones(ctx context.Context) error {
	return nil
}

func (m PromApiMock) Config(ctx context.Context) (promHttpC.ConfigResult, error) {
	return promHttpC.ConfigResult{}, nil
}

func (m PromApiMock) DeleteSeries(ctx context.Context, matches []string, startTime time.Time, endTime time.Time) error {
	return nil
}

func (m PromApiMock) Flags(ctx context.Context) (promHttpC.FlagsResult, error) {
	return nil, nil
}

func (m PromApiMock) LabelNames(ctx context.Context) ([]string, api.Warnings, error) {
	return nil, nil, nil
}

func (m PromApiMock) LabelValues(ctx context.Context, label string) (model.LabelValues, api.Warnings, error) {
	return nil, nil, nil
}

func (m PromApiMock) Query(ctx context.Context, query string, ts time.Time) (model.Value, api.Warnings, error) {
	return nil, nil, nil
}

func (m PromApiMock) Series(ctx context.Context, matches []string, startTime time.Time, endTime time.Time) ([]model.LabelSet, api.Warnings, error) {
	return nil, nil, nil
}

func (m PromApiMock) Snapshot(ctx context.Context, skipHead bool) (promHttpC.SnapshotResult, error) {
	return promHttpC.SnapshotResult{}, nil
}

func (m PromApiMock) Rules(ctx context.Context) (promHttpC.RulesResult, error) {
	return promHttpC.RulesResult{}, nil
}

func (m PromApiMock) Targets(ctx context.Context) (promHttpC.TargetsResult, error) {
	return promHttpC.TargetsResult{}, nil
}

func (m PromApiMock) TargetsMetadata(ctx context.Context, matchTarget string, metric string, limit string) ([]promHttpC.MetricMetadata, error) {
	return nil, nil
}

func TestQueryRange(t *testing.T) {
	m := NewPromApiMock()
	v, w, e := m.QueryRange(context.Background(), "", promHttpC.Range{})
	assert.Nil(t, v)
	assert.Nil(t, w)
	assert.Nil(t, e)
	mV := model.Matrix{}
	mW := api.Warnings{}
	mE := fmt.Errorf("a")
	m.SetQueryRangeOutput(mV, mW, mE)
	v, w, e = m.QueryRange(context.Background(), "", promHttpC.Range{})
	assert.NotNil(t, v)
	assert.NotNil(t, w)
	assert.NotNil(t, e)
}
