package fileHandler

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jalilbengoufa/ptp/internal"
	"github.com/jalilbengoufa/ptp/internal/database/sqlite"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 30 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 2 * time.Second
)

func ReadFileIfModified(lastMod time.Time) ([]byte, time.Time, error) {
	fi, err := os.Stat(internal.Filename)
	if err != nil {
		return nil, lastMod, err
	}
	p, err := os.ReadFile(filepath.Clean(internal.Filename))
	if err != nil {
		return nil, fi.ModTime(), err
	}
	return p, fi.ModTime(), nil
}

func Reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func Writer(ws *websocket.Conn, lastMod time.Time) {
	lastError := ""
	pingTicker := time.NewTicker(pingPeriod)
	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-fileTicker.C:
			var p []byte
			var err error

			p, lastMod, err = ReadFileIfModified(lastMod)

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					p = []byte(lastError)
				}
			} else {
				lastError = ""
			}
			if p != nil {

				// Save to sqlite
				file := sqlite.File{ID: "1", Name: internal.Filename, Content: string(p)}
				sqlite.DbInstanceSqlite.Save(&file)

				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
