package promhelper

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"time"
)

// It is a metric exposed by blackbox_exporter.
const IcmpPingQuery = "probe_duration_seconds{ping!=\"\"}"
const LabelInstance = "instance"
const LabelPing = "ping"

type QueryClient struct {
	prom v1.API
}

func NewQueryClient(promAddr string) (*QueryClient, error) {
	promClient, err := api.NewClient(api.Config{
		Address: promAddr,
	})
	if err != nil {
		return nil, err
	}
	promV1 := v1.NewAPI(promClient)
	return &QueryClient{prom: promV1}, nil
}

func (it *QueryClient) FetchNodeLatencyMetrics(ctx context.Context, duration time.Duration) (map[Link]TimeSeries, error) {
	now := time.Now()
	// here is trick to avoid not enough samples during assertion on time series
	duration = duration + time.Minute
	values, err := it.prom.QueryRange(ctx, IcmpPingQuery, v1.Range{
		Start: now.Add(-duration),
		End:   now,
		Step:  time.Second,
	})

	if err != nil {
		return nil, err
	}
	switch values.Type() {
	case model.ValMatrix:
		result := make(map[Link]TimeSeries)
		for _, matrix := range values.(model.Matrix) {
			result[Link{
				From: string(matrix.Metric[LabelInstance]),
				To:   string(matrix.Metric[LabelPing]),
			}] = parseTimeSeries(matrix.Values)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("failed parse prometheus data with [%s]", values.Type().String())
	}
}
