package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ilia-tsyplenkov/binance-prices/internal/models"
	log "github.com/sirupsen/logrus"
)

const (
	wsStreamURL = "wss://stream.binance.com:9443/stream"
	ts          = "1000ms"
	limit       = "5"
	ticker      = "btcusdt"
)

type Scanner struct {
	connPool []*websocket.Conn
	ctx      context.Context
	queue    chan models.DephMessage
	tickers  []string
}

func New(ctx context.Context, connPool []*websocket.Conn, queue chan models.DephMessage, tickers []string) *Scanner {

	return &Scanner{
		ctx:      ctx,
		connPool: connPool,
		queue:    queue,
		tickers:  tickers,
	}
}

func (s *Scanner) Run() {
	for i, ticker := range s.tickers {
		go s.worker(s.connPool[i], ticker)
	}
}

func (s *Scanner) worker(conn *websocket.Conn, ticker string) {
	defer conn.Close()
	if err := s.subscribe(conn, ticker); err != nil {
		log.Errorf("subscribe: %s: %v", ticker, err)
		return
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msgType, msgB, err := conn.ReadMessage()
			if err != nil {
				log.Error(err)
				break
			}

			msg := models.DephMessage{}
			if err := json.Unmarshal(msgB, &msg); err != nil || msg.Data.LastUpdateID == 0 {
				log.Infof("message: %d: %s: %v", msgType, bytes.NewBuffer(msgB).String(), err)
			} else {
				// log.Infof("message: %d: %s: %v", msgType, bytes.NewBuffer(msgB).String(), err)

				select {
				case s.queue <- msg:
					{
					}
				default:
					{
					}
				}

			}
		}
	}
}

func (s *Scanner) subscribe(conn *websocket.Conn, ticker string) error {
	msg := &models.Subscribe{
		ID:     uuid.New().String(),
		Params: []string{fmt.Sprintf("%s@depth%s@%s", ticker, limit, ts)},
		Method: "SUBSCRIBE",
	}
	log.Infof("subscribe: %v", msg)

	msgB, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, msgB); err != nil {
		return err
	}

	return nil
}
