package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/ilia-tsyplenkov/binance-prices/config"
	"github.com/ilia-tsyplenkov/binance-prices/internal/models"
	"github.com/ilia-tsyplenkov/binance-prices/internal/scanner"
	"github.com/ilia-tsyplenkov/binance-prices/internal/server"
	"github.com/ilia-tsyplenkov/binance-prices/internal/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	log.Infof("config: %v", cfg)
	connectionPool := make([]*websocket.Conn, 0, len(cfg.Tickers))
	for _ = range cfg.Tickers {
		conn, err := connectToBinance(cfg.BinanceWsURL)
		if err != nil {
			log.Fatal(err)
		}
		connectionPool = append(connectionPool, conn)
	}

	queue := make(chan models.DephMessage, 10)

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	go func() {
		signal.Notify(signals,
			os.Interrupt,
			syscall.SIGTERM,
		)
		sig := <-signals
		log.Infof("got signal: %v", sig)
		log.Info("shutting down....")
		cancel()
	}()

	priceStorage := storage.New(cfg.Tickers, queue)
	binanceScanner := scanner.New(
		ctx,
		connectionPool,
		queue,
		cfg.Tickers,
	)
	go priceStorage.Run(ctx)
	binanceScanner.Run()

	srv := server.New(ctx, cfg.Address, priceStorage)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		srv.Run()

	}()

	wg.Wait()

}

func readResponse(r *http.Response) (string, error) {
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, r.Body); err != nil {
		return "", err
	}
	return fmt.Sprintf("status: %s: %s", r.Status, buffer.String()), nil
}

func connectToBinance(wsURL string) (*websocket.Conn, error) {
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respMsg, err := readResponse(resp)
	if err != nil {
		log.Errorf("unable to read response: %v", err)
	} else {
		log.Info(respMsg)
	}

	return conn, nil
}
