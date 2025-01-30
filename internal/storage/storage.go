package storage

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/ilia-tsyplenkov/binance-prices/internal/models"
	log "github.com/sirupsen/logrus"
)

type Storage struct {
	prices  map[string]*models.Price
	updates map[string]int64
	queue   chan models.DephMessage
	mx      sync.Mutex
}

func New(tickers []string, queue chan models.DephMessage) *Storage {
	prices := make(map[string]*models.Price, len(tickers))
	updates := make(map[string]int64, len(tickers))
	for _, t := range tickers {
		prices[t] = &models.Price{
			Symbol: t,
		}
		updates[t] = 0
	}
	return &Storage{
		prices:  prices,
		updates: updates,
		queue:   queue,
	}
}

func (s *Storage) MarshalJSON() ([]byte, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	return json.Marshal(s.prices)
}

func (s *Storage) Add(msg *models.DephMessage) {
	if len(msg.Data.Asks) == 0 || len(msg.Data.Bids) == 0 {
		log.Infof("no asks or bids")
		return
	}
	ticker := strings.Split(msg.Stream, "@")[0]
	s.mx.Lock()
	defer s.mx.Unlock()
	if msg.Data.LastUpdateID < s.updates[ticker] {
		log.Infof("too old update")
		return
	}
	tickerPrice := s.prices[ticker]
	tickerPrice.Ask = msg.Data.Asks[0][0]
	tickerPrice.Bid = msg.Data.Bids[0][0]
}

func (s *Storage) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.queue:
			// log.Info("got message: %v", msg)
			s.Add(&msg)
		}
	}
}
