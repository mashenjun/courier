package wsStorage

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var upgrader websocket.Upgrader

func init() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

type WsStorage struct {
	connMap sync.Map
}

// NewStorage creates a sync.Map to store connections.
func NewStorage() (*WsStorage, error) {
	s := WsStorage{}
	s.connMap = sync.Map{}
	return &s, nil
}

// Delete deletes a connections form the storage.
func (s *WsStorage) Delete(key string) {
	s.connMap.Delete(key)
}

// Load loads a connection form the storage.
func (s *WsStorage) Load(key string) (*connection, error) {
	c, ok := s.connMap.Load(key)
	if !ok {
		return nil, ErrConnectionNotFound
	}
	return c.(*connection), nil
}

// Store stores a connection into the storage.
// It returns ErrIsExisted if the key is already used.
func (s *WsStorage) Store(key string, conn *connection) error {
	_, ok := s.connMap.LoadOrStore(key, conn)
	if ok {
		return ErrIsExisted
	}
	go func(key string, conn *connection) (err error){
		defer func() {
			s.Delete(key)
			conn.CloseWithMessage(err.Error())
		}()
		// ReadMessage returns error if the connection is closed from client side
		for {
			_, _, err = conn.ReadMessage()
			if err != nil {
				return err
			}
		}
	}(key, conn)
	return nil
}
