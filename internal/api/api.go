package api

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/fasthttp/router"
	"github.com/seventv/common/utils"
	"github.com/seventv/compactdisc/client"
	"github.com/seventv/compactdisc/internal/api/handler"
	"github.com/seventv/compactdisc/internal/global"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func Start(gctx global.Context) (<-chan uint8, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", gctx.Config().Http.Addr, gctx.Config().Http.Port))
	if err != nil {
		return nil, err
	}

	router := router.New()
	router.POST("/", func(ctx *fasthttp.RequestCtx) {

	})

	srv := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if utils.B2S(ctx.Method()) != fasthttp.MethodPost {
				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
				return
			}

			body := client.Request[json.RawMessage]{}
			if err := json.Unmarshal(ctx.Request.Body(), &body); err != nil {
				_, _ = ctx.WriteString(err.Error())
				ctx.SetStatusCode(fasthttp.StatusBadRequest)
				return
			}

			zap.S().Infow("executing operation", "operation", body.Operation)

			switch body.Operation {
			case client.OperationNameSyncUser:
				err = handler.SyncUser(gctx, ctx, client.ConvertRequest[client.RequestPayloadSyncUser](body))
			}

			if err != nil {
				_, _ = ctx.WriteString(err.Error())
				ctx.SetStatusCode(fasthttp.StatusBadRequest)
			}
		},
		ReadTimeout:     time.Second * 20,
		IdleTimeout:     time.Second * 20,
		CloseOnShutdown: true,
	}

	done := make(chan uint8)

	go func() {
		<-gctx.Done()

		zap.S().Info("api is shutting down")

		_ = srv.Shutdown()
	}()

	go func() {
		defer close(done)

		if err = srv.Serve(listener); err != nil {
			zap.S().Fatalw("failed to start api", "error", err)
		}
	}()

	return done, nil
}
