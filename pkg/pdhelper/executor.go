package pdhelper

import (
	"auto-failover-tikv-leader-evict/pkg/log"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

const evictLeaderScheduler = "evict-leader-scheduler"

type Executor struct {
	PdAddr string
}

func NewExecutor(pdAddr string) *Executor {
	return &Executor{PdAddr: pdAddr}
}

func (it *Executor) AddEvictScheduler(storeId uint) error {
	log.L().With(zap.String("command", fmt.Sprintf("pd-ctl -u %s scheduler add %s %s", it.PdAddr, evictLeaderScheduler, fmt.Sprintf("%d", storeId)))).Info("add an evict scheduler")

	out, err := exec.Command("pd-ctl", "-u", it.PdAddr, "scheduler", "add", evictLeaderScheduler, fmt.Sprintf("%d", storeId)).CombinedOutput()
	if err != nil {
		log.L().With(zap.String("out", string(out))).Error("failed to execute pd-ctl scheduler add")
		return err
	}
	log.L().With(zap.String("output", string(out))).Debug("pd-ctl scheduler add")
	if strings.Contains(string(out), "Success") {
		return nil
	} else {
		return fmt.Errorf("failed to add evict scheuler, %s", out)
	}
}

func (it *Executor) RemoveEvictScheduler(storeId uint) error {
	log.L().With(zap.String("command", fmt.Sprintf("pd-ctl -u %s scheduler remove %s", it.PdAddr, fmt.Sprintf("%s-%d", evictLeaderScheduler, storeId)))).Info("remove an evict scheduler")

	out, err := exec.Command("pd-ctl", "-u", it.PdAddr, "scheduler", "remove", fmt.Sprintf("%s-%d", evictLeaderScheduler, storeId)).CombinedOutput()
	if err != nil {
		log.L().With(zap.String("out", string(out))).Error("failed to execute pd-ctl scheduler remove")
		return err
	}
	log.L().With(zap.String("output", string(out))).Debug("pd-ctl scheduler remove")
	if strings.Contains(string(out), "Success") {
		return nil
	} else {
		return fmt.Errorf("failed to remove evict scheuler, %s", out)
	}
}

func (it *Executor) ListStores() ([]Store, error) {
	out, err := exec.Command("pd-ctl", "-u", it.PdAddr, "store").CombinedOutput()
	if err != nil {
		log.L().With(zap.String("out", string(out))).Error("failed to execute pd-ctl store")
		return nil, err
	}
	log.L().With(zap.String("output", string(out))).Debug("pd-ctl store")
	pdOutput := PdStore{}
	err = json.Unmarshal(out, &pdOutput)
	if err != nil {
		return nil, err
	}
	var result []Store
	for _, store := range pdOutput.Stores {
		result = append(result, store.Store)
	}
	return result, nil
}

func (it *Executor) ListEvictedStore() ([]Store, error) {
	out, err := exec.Command("pd-ctl", "-u", it.PdAddr, "scheduler", "config", "evict-leader-scheduler").CombinedOutput()
	if err != nil {
		log.L().With(zap.String("out", string(out))).Error("failed to execute pd-ctl store")
		return nil, err
	}

	log.L().With(zap.String("output", string(out))).Debug("pd-ctl scheduler config evict-leader-scheduler")

	// if there is no scheduler called "evict-leader-scheduler", pd-ctl will print something like "[404] 404 page not found"
	if strings.Contains(string(out), "404") {
		return nil, nil
	}

	var schedulerConfig PdSchedulerConfig
	err = json.Unmarshal(out, &schedulerConfig)
	if err != nil {
		log.L().With(zap.Error(err)).Error("failed to parse output for pd-ctl scheduler config evict-leader-scheduler")
		return nil, err
	}

	storeIds := schedulerConfig.FetchStoreIds()

	stores, err := it.ListStores()
	if err != nil {
		return nil, err
	}

	var result []Store
	for _, store := range stores {
		for _, id := range storeIds {
			if store.Id == id {
				result = append(result, store)
			}
		}
	}
	return result, nil
}
