package evictor

import "time"

type Config struct {
	PrometheusAddress string
	PdAddress         string
	MaxEvicted        uint
	Interval          time.Duration
	Threshold         time.Duration
	PendingForEvict   time.Duration
	PendingForRecover time.Duration
}

func (it Config) RequiredMaxTimeRange() time.Duration {
	var duration time.Duration

	if it.Interval > it.PendingForEvict {
		duration = it.Interval
	} else {
		duration = it.PendingForEvict
	}

	if it.PendingForRecover > duration {
		duration = it.PendingForRecover
	}
	return duration
}
