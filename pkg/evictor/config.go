package evictor

import "time"

const VersionV3 string = "v3"
const VersionV4 string = "v4"

type Config struct {
	PrometheusAddress string
	PdAddress         string
	MaxEvicted        uint
	PdVersion         string
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
