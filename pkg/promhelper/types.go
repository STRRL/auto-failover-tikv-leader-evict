package promhelper

import (
	"github.com/prometheus/common/model"
	"time"
)

type Sample struct {
	Timestamp time.Time
	Latency   time.Duration
}

type Link struct {
	From string
	To   string
}

func parseTimeSeries(origin []model.SamplePair) TimeSeries {
	result := make([]Sample, len(origin))
	for i, pair := range origin {
		result[i] =
			Sample{
				Timestamp: pair.Timestamp.Time(),
				Latency:   time.Duration((float64(pair.Value)) * float64(time.Second)),
			}
	}
	return result
}

type TimeSeries []Sample

func (it *TimeSeries) LatencyLargerThanThresholdFor(threshold, atLeastFor time.Duration) bool {
	if len(*it) < 2 {
		return false
	}
	if atLeastFor > ((*it)[len(*it)-1].Timestamp.Sub((*it)[0].Timestamp)) {
		return false
	}

	start := (*it)[len(*it)-1].Timestamp.Add(-atLeastFor)
	for _, sample := range *it {
		if sample.Timestamp.Before(start) {
			continue
		}
		if sample.Latency > threshold {
			continue
		} else {
			return false
		}
	}
	return true

}

func (it *TimeSeries) LatencySmallerThanThresholdFor(threshold, atLeastFor time.Duration) bool {
	if len(*it) < 2 {
		return false
	}
	if atLeastFor > ((*it)[len(*it)-1].Timestamp.Sub((*it)[0].Timestamp)) {
		return false
	}
	start := (*it)[len(*it)-1].Timestamp.Add(-atLeastFor)
	for _, sample := range *it {
		if sample.Timestamp.Before(start) {
			continue
		}
		if sample.Latency < threshold {
			continue
		} else {
			return false
		}
	}
	return true
}
