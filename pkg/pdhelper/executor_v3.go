package pdhelper

import (
	"auto-failover-tikv-leader-evict/pkg/log"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"strconv"
	"strings"
)

type ExecutorV3 struct {
	PdAddr string
}

func NewExecutorV3(pdAddr string) *ExecutorV3 {
	return &ExecutorV3{PdAddr: pdAddr}
}

func (it *ExecutorV3) AddEvictScheduler(storeId uint) error {
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

func (it *ExecutorV3) RemoveEvictScheduler(storeId uint) error {
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

func (it *ExecutorV3) ListStores() ([]Store, error) {
	out, err := exec.Command("pd-ctl", "-u", it.PdAddr, "store").CombinedOutput()
	if err != nil {
		log.L().With(zap.String("out", string(out))).Error("failed to execute pd-ctl store")
		return nil, err
	}
	log.L().With(zap.String("output", string(out))).Debug("pd-ctl store")
	pdOutput := PdStore{}
	err = json.Unmarshal(out, &pdOutput)
	if err != nil {
		log.L().With(zap.Error(err)).With(zap.String("output", string(out))).Warn("failed to parse output for pd-ctl store")
		return nil, err
	}
	var result []Store
	for _, store := range pdOutput.Stores {
		result = append(result, store.Store)
	}
	return result, nil
}

func (it *ExecutorV3) ListEvictedStore() ([]Store, error) {
	out, err := exec.Command("pd-ctl", "-u", it.PdAddr, "scheduler", "show").CombinedOutput()
	if err != nil {
		log.L().With(zap.String("out", string(out))).Error("failed to execute pd-ctl scheduler show")
		return nil, err
	}

	log.L().With(zap.String("output", string(out))).Debug("pd-ctl scheduler show")

	// if there is no scheduler called "evict-leader-scheduler", pd-ctl will print something like "[404] 404 page not found"
	if strings.Contains(string(out), "404") {
		return nil, nil
	}

	var schedulers PdSchedulerShow
	err = json.Unmarshal(out, &schedulers)
	if err != nil {
		log.L().With(zap.Error(err)).With(zap.String("output", string(out))).Error("failed to parse output for pd-ctl scheduler config evict-leader-scheduler")
		return nil, err
	}

	storeIds := schedulers.FetchStoreIds()

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

type PdSchedulerShow []string

func (it *PdSchedulerShow) FetchStoreIds() []uint {
	var result []uint
	for _, item := range *it {
		if strings.Contains(item, evictLeaderScheduler) {
			parsed, err := strconv.ParseUint(item[strings.LastIndex(item, "-")+1:], 10, 32)
			if err != nil {
				log.L().With(zap.Any("output", *it)).Error("failed to parse store id from pd-ctl scheduler show")
				continue
			}
			evictedStoreId := uint(parsed)
			result = append(result, evictedStoreId)
		}
	}

	return result
}
