/*
 * Copyright (c) 2021 Nikita Krasnikov
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikitaksv/jgen/pkg/endpoint"
	"github.com/nikitaksv/jgen/pkg/resource"
	"github.com/nikitaksv/jgen/pkg/service"
	http2 "github.com/nikitaksv/jgen/pkg/transport/http"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()

	var res resource.Resource
	{
		res = resource.New(logger)
	}

	var svc service.Service
	{
		svc = service.NewService(res)
	}

	var ends endpoint.Endpoints
	{
		ends = endpoint.New(svc)
		ends.Generate = endpoint.LoggingMiddleware("generateTemplate", logger)(ends.Generate)
	}

	var h http.Handler
	{
		h = http2.MakeHTTPHandler(ends)
	}

	errs := make(chan error)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	httpAddr := "127.0.0.1:8081"
	hs := &http.Server{
		Addr:    httpAddr,
		Handler: h,
	}

	go func() {
		logger.Sugar().Infow("server started",
			zap.String("addr", httpAddr),
		)
		if err := hs.ListenAndServe(); err != http.ErrServerClosed {
			errs <- err
		}
	}()

	select {
	case err := <-errs:
		logger.Sugar().Panicw("server error occurred", zap.Error(err))
	case <-stop:
		timeout := 15 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		logger.Info("Shutting down...", zap.Duration("timeout", timeout))
		if err := hs.Shutdown(ctx); err != nil {
			logger.Sugar().Errorw("Shutdown failed", zap.Error(err))
		} else {
			logger.Sugar().Infow("Shutdown success")
		}
		cancel()
	}

}
