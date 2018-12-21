package wsStorage

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// connection warps the real ws connection
type connection struct {
	conn   *websocket.Conn
	wMutex sync.Mutex
	rMutex sync.Mutex
	ticker *time.Ticker
	done   chan struct{} // only used to notify keep alive should be stop
	once   sync.Once
}

// NewConn creates a connection which warp the real web socket connection.
func NewConn(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*connection, error) {
	c, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return &connection{
		conn: c,
		done: make(chan struct{}),
	}, nil
}

// WriteJSON sends the given data through the connection in json format.
func (conn *connection) WriteJSON(v interface{}) error {
	conn.wMutex.Lock()
	defer conn.wMutex.Unlock()
	return conn.conn.WriteJSON(v)
}

// ReadJSON reads and returns data form the connection and in json format.
func (conn *connection) ReadJSON(v interface{}) error {
	conn.rMutex.Lock()
	defer conn.rMutex.Unlock()
	return conn.conn.ReadJSON(v)
}

// ReadMessage reads and return data form the connection as bytes.
func (conn *connection) ReadMessage() (messageType int, p []byte, err error) {
	conn.rMutex.Lock()
	defer conn.rMutex.Unlock()
	return conn.conn.ReadMessage()
}

func (conn *connection) Close() error {
	if conn.ticker != nil {
		conn.ticker.Stop()
	}
	conn.sendClose("")
	conn.once.Do(
		func() {
			close(conn.done)
		})
	return conn.conn.Close()
}

func (conn *connection) CloseWithMessage(msg string) error {
	if conn.ticker != nil {
		conn.ticker.Stop()
	}
	conn.sendClose(msg)
	conn.once.Do(
		func() {
			close(conn.done)
		})
	return conn.conn.Close()
}

// KeepLive sends heartbeat message through the connection periodically
func (conn *connection) KeepLive(dur time.Duration) error {
	if dur <= 0 {
		return ErrInvalidDuration
	}
	conn.ticker = time.NewTicker(dur)
	// run a goroutine to keep connection alive
	go func(conn *connection) {
		for {
			select {
			case <-conn.ticker.C:
				conn.sendPing(2 * time.Second)
			case <-conn.done:
				return
			}
		}
	}(conn)
	return nil
}

// ----------------------------------------------

func (conn *connection) sendPing(dur time.Duration) error {
	return conn.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(dur))
}

func (conn *connection) sendClose(msg string) error {
	return conn.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg), time.Now().Add(time.Second))
}

//func (conn *connection) close() {
//	conn.once.Do(func(){
//		close(conn.done)
//		conn.conn.Close()
//	})
//}
