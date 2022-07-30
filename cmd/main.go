package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/seventv/compactdisc/internal/discord/handler"

	"github.com/bugsnag/panicwrap"
	"github.com/seventv/common/mongo"
	"github.com/seventv/common/redis"
	"github.com/seventv/common/structures/v3/query"
	"github.com/seventv/compactdisc/internal/api"
	"github.com/seventv/compactdisc/internal/configure"
	"github.com/seventv/compactdisc/internal/discord"
	"github.com/seventv/compactdisc/internal/global"
	"github.com/seventv/compactdisc/internal/health"
	"go.uber.org/zap"
)

var (
	Version = "development"
	Unix    = ""
	Time    = "unknown"
	User    = "unknown"
)

func init() {
	if i, err := strconv.Atoi(Unix); err == nil {
		Time = time.Unix(int64(i), 0).Format(time.RFC3339)
	}
}

func main() {
	config := configure.New()

	exitStatus, err := panicwrap.BasicWrap(func(s string) {
		zap.S().Errorw("panic detected",
			"panic", s,
		)
	})
	if err != nil {
		zap.S().Errorw("failed to setup panic handler",
			"error", err,
		)
		os.Exit(2)
	}

	if exitStatus >= 0 {
		os.Exit(exitStatus)
	}

	if !config.NoHeader {
		zap.S().Info("7TV CD")
		zap.S().Infof("Version: %s", Version)
		zap.S().Infof("build.Time: %s", Time)
		zap.S().Infof("build.User: %s", User)
	}

	zap.S().Debugf("MaxProcs: %d", runtime.GOMAXPROCS(0))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	gctx, cancel := global.WithCancel(global.New(context.Background(), config))

	{
		gctx.Inst().Redis, err = redis.Setup(gctx, redis.SetupOptions{
			Username:   config.Redis.Username,
			Password:   config.Redis.Password,
			Database:   config.Redis.Database,
			Sentinel:   config.Redis.Sentinel,
			Addresses:  config.Redis.Addresses,
			MasterName: config.Redis.MasterName,
			EnableSync: true,
		})
		if err != nil {
			zap.S().Fatalw("failed to setup redis handler",
				"error", err,
			)
		}

		zap.S().Infow("redis, ok")
	}

	{
		gctx.Inst().Mongo, err = mongo.Setup(gctx, mongo.SetupOptions{
			URI:    config.Mongo.URI,
			DB:     config.Mongo.DB,
			Direct: config.Mongo.Direct,
		})
		if err != nil {
			zap.S().Fatalw("failed to setup mongo handler",
				"error", err,
			)
		}

		zap.S().Infow("mongo, ok")
	}

	{
		handler.DefaultRoleId = config.Discord.DefaultRoleId
		gctx.Inst().Discord, err = discord.New(gctx, config.Discord.Token)
		if err != nil {
			zap.S().Fatalw("failed to setup discord", "error", err)
		}

		zap.S().Infow("discord, ok")
	}

	{
		gctx.Inst().Query = query.New(gctx.Inst().Mongo, gctx.Inst().Redis)
	}

	wg := sync.WaitGroup{}

	if gctx.Config().Health.Enabled {
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-health.New(gctx)
		}()
	}

	apiDone, err := api.Start(gctx)
	if err != nil {
		zap.S().Fatalw("failed to start api", "error", err)
	}

	done := make(chan struct{})

	go func() {
		<-sig
		cancel()

		go func() {
			select {
			case <-time.After(time.Minute):
			case <-sig:
			}
			zap.S().Fatal("force shutdown")
		}()

		zap.S().Info("shutting down")

		<-apiDone
		wg.Wait()

		close(done)
	}()

	zap.S().Info("running")

	<-done

	zap.S().Info("shutdown")
	os.Exit(0)
}
