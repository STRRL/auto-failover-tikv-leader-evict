package pdhelper

import (
	"auto-failover-tikv-leader-evict/pkg/log"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type PdStore struct {
	Count  int         `json:"count"`
	Stores []StoreItem `json:"stores"`
}

type StoreItem struct {
	Store Store `json:"store"`
}

type Store struct {
	Id      uint   `json:"id"`
	Address string `json:"address"`
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
