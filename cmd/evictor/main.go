package main

import (
	"auto-failover-tikv-leader-evict/cmd/evictor/command"
	"auto-failover-tikv-leader-evict/pkg/log"
	"go.uber.org/zap"
)

func main() {
	if err := command.NewRootCmd().Execute(); err != nil {
		log.L().With(zap.Error(err)).Error("failed to execute")
		return
	}
}
