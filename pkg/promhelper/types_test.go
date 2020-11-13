package promhelper

import (
	"math/rand"
	"testing"
	"time"
)

func makeSeries(base, jitter, step time.Duration, start time.Time, number int) TimeSeries {
	var result []Sample
	for i := 0; i < number; i++ {
		item := Sample{
			Timestamp: start.Add(time.Duration(i) * step),
			Latency:   base + time.Duration((rand.Float64()*2-1)*float64(jitter)),
		}
		result = append(result, item)
	}
	return result
}

func TestTimeSeries_LatencyLargerThanThresholdFor(t *testing.T) {
	now := time.Now()
	type args struct {
		threshold  time.Duration
		atLeastFor time.Duration
	}
	tests := []struct {
		name string
		it   TimeSeries
		args args
		want bool
	}{
		{
			"nil time series",
			nil,
			args{
				threshold:  time.Second,
				atLeastFor: time.Minute,
			},
			false,
		}, {
			"only one point",
			[]Sample{{
				Timestamp: time.Time{},
				Latency:   0,
			}},
			args{
				threshold:  time.Second,
				atLeastFor: time.Minute,
			},
			false,
		}, {
			"not enough points",
			[]Sample{
				{
					Timestamp: now.Add(-30 * time.Second),
					Latency:   time.Millisecond,
				}, {
					Timestamp: now.Add(-15 * time.Second),
					Latency:   time.Millisecond,
				}, {
					Timestamp: now,
					Latency:   time.Millisecond,
				},
			},
			args{
				threshold:  time.Second,
				atLeastFor: time.Minute,
			},
			false,
		}, {
			"normal situation",
			[]Sample{
				{
					Timestamp: now.Add(-61 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now.Add(-15 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now,
					Latency:   time.Second,
				},
			},
			args{
				threshold:  time.Millisecond,
				atLeastFor: time.Minute,
			},
			true,
		}, {
			"boundary value",
			[]Sample{
				{
					// it could just on the atLeastFor boundary
					Timestamp: now.Add(-60 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now.Add(-15 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now,
					Latency:   time.Second,
				},
			},
			args{
				threshold:  time.Millisecond,
				atLeastFor: time.Minute,
			},
			true,
		}, {
			"complicated",
			makeSeries(2*time.Second, 500*time.Millisecond, time.Second, now.Add(-60*time.Second), 61),
			args{
				threshold:  time.Second,
				atLeastFor: time.Minute,
			},
			true,
		}, {
			"complicated_unstable",
			makeSeries(1*time.Second, 500*time.Millisecond, time.Second, now.Add(-60*time.Second), 61),
			args{
				threshold:  time.Second,
				atLeastFor: time.Minute,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.LatencyLargerThanThresholdFor(tt.args.threshold, tt.args.atLeastFor); got != tt.want {
				t.Errorf("LatencyLargerThanThresholdFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeSeries_LatencySmallerThanThresholdFor(t *testing.T) {
	now := time.Now()
	type args struct {
		threshold  time.Duration
		atLeastFor time.Duration
	}
	tests := []struct {
		name string
		it   TimeSeries
		args args
		want bool
	}{
		{
			"nil time series",
			nil,
			args{
				threshold:  time.Second,
				atLeastFor: time.Minute,
			},
			false,
		}, {
			"only one point",
			[]Sample{{
				Timestamp: time.Time{},
				Latency:   0,
			}},
			args{
				threshold:  2 * time.Second,
				atLeastFor: time.Minute,
			},
			false,
		}, {
			"not enough points",
			[]Sample{
				{
					Timestamp: now.Add(-30 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now.Add(-15 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now,
					Latency:   time.Second,
				},
			},
			args{
				threshold:  2 * time.Second,
				atLeastFor: time.Minute,
			},
			false,
		}, {
			"normal situation",
			[]Sample{
				{
					Timestamp: now.Add(-61 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now.Add(-15 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now,
					Latency:   time.Second,
				},
			},
			args{
				threshold:  2 * time.Second,
				atLeastFor: time.Minute,
			},
			true,
		}, {
			"boundary value",
			[]Sample{
				{
					// it could just on the atLeastFor boundary
					Timestamp: now.Add(-60 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now.Add(-15 * time.Second),
					Latency:   time.Second,
				}, {
					Timestamp: now,
					Latency:   time.Second,
				},
			},
			args{
				threshold:  2 * time.Second,
				atLeastFor: time.Minute,
			},
			true,
		}, {
			"complicated",
			makeSeries(2*time.Second, 500*time.Millisecond, time.Second, now.Add(-60*time.Second), 61),
			args{
				threshold:  3 * time.Second,
				atLeastFor: time.Minute,
			},
			true,
		}, {
			"complicated_unstable",
			makeSeries(time.Second, 500*time.Millisecond, time.Second, now.Add(-60*time.Second), 61),
			args{
				threshold:  time.Second,
				atLeastFor: time.Minute,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.LatencySmallerThanThresholdFor(tt.args.threshold, tt.args.atLeastFor); got != tt.want {
				t.Errorf("LatencySmallerThanThresholdFor() = %v, want %v", got, tt.want)
			}
		})
	}
}
