package command

import (
	"auto-failover-tikv-leader-evict/pkg/evictor"
	"auto-failover-tikv-leader-evict/pkg/log"
	"context"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var config = evictor.Config{
	PrometheusAddress: "",
	PdAddress:         "",
	MaxEvicted:        0,
	Interval:          0,
	Threshold:         0,
	PendingForEvict:   0,
	PendingForRecover: 0,
}
var debug = false

const defaultInterval = 15 * time.Second

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Run: run,
	}
	rootCmd.Flags().StringVar(&config.PrometheusAddress, "prometheus", "", "address of prometheus")
	rootCmd.MarkFlagRequired("prometheus")
	rootCmd.Flags().StringVar(&config.PdAddress, "pd", "", "address of pd")
	rootCmd.MarkFlagRequired("pd")
	rootCmd.Flags().UintVar(&config.MaxEvicted, "max-evicted", 10, "max number of tikv which could be evicted leader by this tool")
	rootCmd.Flags().DurationVar(&config.Interval, "interval", defaultInterval, "interval for refresh latency metrics")
	rootCmd.Flags().DurationVar(&config.Threshold, "threshold", time.Second, "a node which hold a latency longer than threshold will be treated as unhealthy")
	rootCmd.Flags().DurationVar(&config.PendingForEvict, "pending-for-evict", time.Minute, "an unhealthy tikv node will be evicted after this duration")
	rootCmd.Flags().DurationVar(&config.PendingForRecover, "pending-for-recover", 2*defaultInterval, "an evicted tikv with stable latency will recover at least after this duration")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "print debug logs")
	return rootCmd
}

func run(cmd *cobra.Command, args []string) {
	if debug {
		log.EnableDebug()
	}
	log.L().With(zap.Any("config", config)).Info("evictor configurations")
	instance, err := evictor.NewEvictor(config)
	if err != nil {
		log.L().With(zap.Error(err)).Error("failed to initialize evictor")
	}
	err = instance.Run(makeContext())
	if err != nil {
		log.L().With(zap.Error(err)).Error("failed to execute evictor")
	}
}

func makeContext() context.Context {
	ctx, cancelFunc := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		select {
		case <-c:
			cancelFunc()
		case <-ctx.Done():
		}
	}()

	return ctx
}
