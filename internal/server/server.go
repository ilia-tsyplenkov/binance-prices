package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ilia-tsyplenkov/binance-prices/internal/storage"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	storage  *storage.Storage
	upgrader websocket.Upgrader
	address  string
	ctx      context.Context
}

func New(ctx context.Context, addr string, storage *storage.Storage) *Server {
	return &Server{
		ctx:     ctx,
		address: addr,
		storage: storage,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Пропускаем любой запрос
			},
		},
	}
}

func (srv *Server) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error: %v", err)
	}
	defer conn.Close()

	for tick := time.Tick(time.Second); ; {
		select {
		case <-tick:
			msg, err := srv.storage.MarshalJSON()
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("marhaling error: %v", err)))
			}
			conn.WriteMessage(websocket.TextMessage, msg)
		case <-r.Context().Done():
			log.Infof("request context done")
			return
		case <-srv.ctx.Done():
			log.Infof("server context done")
			return
		}
	}

}

func (srv *Server) Run() {

	l := log.WithField("action", "server.Run")

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handler)

	httpServer := http.Server{
		Addr:    srv.address,
		Handler: mux,
	}

	go func() {
		<-srv.ctx.Done()
		httpServer.Shutdown(context.Background())
	}()

	if err := httpServer.ListenAndServe(); err != nil {
		l.Error(err)

	}
}
