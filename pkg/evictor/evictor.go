package evictor

import (
	"auto-failover-tikv-leader-evict/pkg/log"
	"auto-failover-tikv-leader-evict/pkg/pdhelper"
	"auto-failover-tikv-leader-evict/pkg/promhelper"
	"context"
	"fmt"
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
	pd := pdhelper.NewExecutorV3(config.PdAddress)
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
	pd     pdhelper.Executor
	prom   *promhelper.QueryClient
}

func (it *Evictor) Run(ctx context.Context) error {
	ticker := time.NewTicker(it.config.Interval)
	defer ticker.Stop()

	for {
		if err := it.loopForever(ctx); err != nil {
			log.L().With(zap.Error(err)).Error("failed to execute loop")
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
		log.L().With(zap.Error(err)).Error("failed to find out should evicted stores; it will not evict any nodes at this time")
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
		log.L().With(zap.Error(err)).Error("failed to find out should recovered stores; it will not recover any tikv nodes at this time")
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
	var shouldEvicts []pdhelper.Store
	for key, health := range nodes {
		if health != Unhealthy {
			continue
		}
		for _, store := range allStores {
			if strings.Contains(store.Address, key) {
				shouldEvicts = append(shouldEvicts, store)
			}
		}
	}

	evictedStores, err := it.pd.ListEvictedStore()

	// check max-evicted
	if uint(len(evictedStores)) >= it.config.MaxEvicted {
		log.L().With(zap.Uint("max-evicted", it.config.MaxEvicted)).With(zap.Any("already-evicted", evictedStores)).Warn("max-evicted exceed")
		return nil, fmt.Errorf("max-evicted exceed")
	}

	var result []pdhelper.Store

	for _, shouldEvictItem := range shouldEvicts {
		newToEvict := true
		for _, evictedStore := range evictedStores {
			if shouldEvictItem.Id == evictedStore.Id {
				newToEvict = false
				break
			}
		}
		if newToEvict {
			result = append(result, shouldEvictItem)
		}
	}
	if len(result) > 0 {
		log.L().With(zap.Any("already-evicted", evictedStores)).With(zap.Any("new-to-evicted", result)).Info("fetch new stores to evicted")
	} else {
		log.L().With(zap.Any("already-evicted", evictedStores)).With(zap.Any("new-to-evicted", result)).Debug("fetch new stores to evicted")
	}
	return result, nil
}

func (it *Evictor) findOutShouldRecover(healthMap map[string]NodeHealth) ([]pdhelper.Store, error) {
	evictedStores, err := it.getEvicted()
	if err != nil {
		return nil, err
	}
	var newToRecover []pdhelper.Store
	for _, store := range evictedStores {
		var ipAddress string
		if strings.Contains(store.Address, ":") {
			ipAddress = store.Address[:strings.LastIndex(store.Address, ":")]
		} else {
			ipAddress = store.Address
		}
		if value, ok := healthMap[ipAddress]; ok && value == Healthy {
			newToRecover = append(newToRecover, store)
		}
	}
	if len(newToRecover) > 0 {
		log.L().With(zap.Any("already-evicted", evictedStores)).With(zap.Any("new-to-recover", newToRecover)).Info("new stores to recover")
	} else {
		log.L().With(zap.Any("already-evicted", evictedStores)).With(zap.Any("new-to-recover", newToRecover)).Debug("new stores to recover")
	}
	return newToRecover, nil
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
