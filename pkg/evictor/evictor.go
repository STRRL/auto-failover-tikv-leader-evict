package evictor

import (
	"auto-failover-tikv-leader-evict/pkg/log"
	"auto-failover-tikv-leader-evict/pkg/pdhelper"
	"auto-failover-tikv-leader-evict/pkg/promhelper"
	"context"
	"go.uber.org/zap"
	"strings"
	"time"
)

type NodeHealth string

const (
	Healthy   NodeHealth = "healthy"
	Unhealthy NodeHealth = "unhealthy"
	Unstable  NodeHealth = "unstable"
)

func NewEvictor(config Config) (*Evictor, error) {
	queryClient, err := promhelper.NewQueryClient(config.PrometheusAddress)
	pd := pdhelper.NewExecutor(config.PdAddress)
	if err != nil {
		return nil, err
	}
	return &Evictor{
		config: config,
		prom:   queryClient,
		pd:     pd,
	}, nil
}

type Evictor struct {
	config Config
	pd     *pdhelper.Executor
	prom   *promhelper.QueryClient
}

func (it *Evictor) Run(ctx context.Context) error {
	ticker := time.NewTicker(it.config.Interval)
	defer ticker.Stop()

	for {
		if err := it.loopForever(ctx); err != nil {
			log.L().With(zap.Error(err)).Error("failed to evict tikv leader")
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			log.L().Info("evictor exiting")
			return nil
		}
	}
}

func (it *Evictor) loopForever(ctx context.Context) error {
	// it follows best-effort pattern
	metrics, err := it.prom.FetchNodeLatencyMetrics(ctx, it.config.RequiredMaxTimeRange())
	if err != nil {
		log.L().With(zap.Error(err)).Error("failed to fetch metrics; it will not do any operations")
		return err
	}
	if len(metrics) == 0 {
		log.L().Warn("could not found target metrics on prometheus")
	}

	healthMap := it.generateNodeHealthMap(metrics)
	log.L().With(zap.Any("status", healthMap)).Debug("nodes status")

	// evict
	if shouldEvict, err := it.findOutShouldEvict(healthMap); err != nil {
		log.L().With(zap.Error(err)).Error("failed to all stores; it will not evict any nodes at this time")
	} else {
		for _, store := range shouldEvict {
			err := it.pd.AddEvictScheduler(store.Id)
			if err != nil {
				log.L().With(zap.Error(err)).With(zap.Any("store", store)).Error("failed to evict node")
			} else {
				log.L().With(zap.Any("store", store)).Info("tikv node evicted")
			}
		}
	}

	// recover
	if shouldRecover, err := it.findOutShouldRecover(healthMap); err != nil {
		log.L().With(zap.Error(err)).Error("failed to fetch evicted node; it will not recover any tikv nodes at this time")
	} else {
		for _, store := range shouldRecover {
			err := it.pd.RemoveEvictScheduler(store.Id)
			if err != nil {
				log.L().With(zap.Error(err)).With(zap.Any("store", store)).Error("failed to recover node")
			} else {
				log.L().With(zap.Any("store", store)).Info("tikv node recovered")
			}
		}
	}
	return nil
}

func (it *Evictor) findOutShouldEvict(nodes map[string]NodeHealth) ([]pdhelper.Store, error) {
	allStores, err := it.pd.ListStores()
	if err != nil {
		return nil, err
	}
	var shouldEvict []pdhelper.Store
	for _, store := range allStores {
		for key, _ := range nodes {
			if strings.Contains(store.Address, key) {
				shouldEvict = append(shouldEvict, store)
			}
		}
	}
	return shouldEvict, nil
}

func (it *Evictor) findOutShouldRecover(healthMap map[string]NodeHealth, ) ([]pdhelper.Store, error) {
	evictedStore, err := it.getEvicted()
	if err != nil {
		return nil, err
	}
	var shouldRecover []pdhelper.Store
	for _, store := range evictedStore {
		var ipAddress string
		if strings.Contains(store.Address, ":") {
			ipAddress = store.Address[:strings.LastIndex(store.Address, ":")]
		} else {
			ipAddress = store.Address
		}
		if value, ok := healthMap[ipAddress]; ok && value == Healthy {
			shouldRecover = append(shouldRecover, store)
		}
	}
	return shouldRecover, nil
}

func (it *Evictor) generateNodeHealthMap(metrics map[promhelper.Link]promhelper.TimeSeries) map[string]NodeHealth {
	nodes := make(map[string]NodeHealth)

	for link, ts := range metrics {
		if ts.LatencyLargerThanThresholdFor(it.config.Threshold, it.config.PendingForEvict) {
			// As any one link performs as unhealthy, this node treads unhealthy.
			// It could overwrite existed Healthy and Unstable.
			nodes[link.To] = Unhealthy
		} else if ts.LatencySmallerThanThresholdFor(it.config.Threshold, it.config.PendingForRecover) {
			if _, ok := nodes[link.To]; !ok {
				nodes[link.To] = Healthy
			}
		} else {
			// Unstable could only overwrite Healthy.
			if value, ok := nodes[link.To]; !ok || value == Healthy {
				nodes[link.To] = Unstable
			}
		}
	}
	return nodes
}

func (it *Evictor) getEvicted() ([]pdhelper.Store, error) {
	return it.pd.ListEvictedStore()
}
