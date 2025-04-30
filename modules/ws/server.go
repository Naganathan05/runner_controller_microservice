package ws

import (
	"context"
	"errors"
	"evolve/util"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const liveLogDir = "live"

var globalHub *Hub

// Hub maintains the set of active clients
// and broadcasts messages to them.
type Hub struct {
	runs       map[string]map[*websocket.Conn]bool
	broadcast  chan *message
	register   chan *client
	unregister chan *client
	mu         sync.RWMutex
}

// message represents a message sent to the hub.
type message struct {
	runId   string
	payload []byte
	sender  *websocket.Conn
}

// client represents a WebSocket client.
type client struct {
	runId string
	conn  *websocket.Conn
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *message),
		register:   make(chan *client),
		unregister: make(chan *client),
		runs:       make(map[string]map[*websocket.Conn]bool),
	}
}

// run starts the hub's processing loop.
func (h *Hub) run(ctx context.Context, logger util.Logger) {
	logger.Info("[WS Hub] Starting")
	defer logger.Info("[WS Hub] Stopped")
	for {
		select {
		case <-ctx.Done():
			logger.Info("[WS Hub] Shutdown signal received")
			return

		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.runs[client.runId]; !ok {
				h.runs[client.runId] = make(map[*websocket.Conn]bool)
			}
			h.runs[client.runId][client.conn] = true
			logger.Info(fmt.Sprintf("[WS Hub] Client registered for run-id: %s. Total clients for run: %d", client.runId, len(h.runs[client.runId])))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if runClients, ok := h.runs[client.runId]; ok {
				if _, clientOk := runClients[client.conn]; clientOk {
					delete(runClients, client.conn)
					logger.Info(fmt.Sprintf("[WS Hub] Client unregistered for run-id: %s", client.runId))
					if len(runClients) == 0 {
						delete(h.runs, client.runId)
						logger.Info(fmt.Sprintf("[WS Hub] Run-id room removed: %s", client.runId))
					} else {
						logger.Info(fmt.Sprintf("[WS Hub] Remaining clients for run %s: %d", client.runId, len(runClients)))
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			runClients, exists := h.runs[message.runId]
			if !exists {
				h.mu.RUnlock()
				continue
			}

			for conn := range runClients {
				if conn == message.sender {
					continue
				}
				go func(ctx context.Context, c *websocket.Conn, payload []byte, runId string, l util.Logger) {
					writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
					defer writeCancel()

					err := c.Write(writeCtx, websocket.MessageText, payload)
					if err != nil {
						status := websocket.CloseStatus(err)
						if status == -1 && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
							l.Error(fmt.Sprintf("[WS Hub Write] Error writing to client for run %s: %v", runId, err))
						} else if errors.Is(err, context.DeadlineExceeded) {
							l.Warn(fmt.Sprintf("[WS Hub Write] Write timeout for client in run %s", runId))
						}
					}
				}(ctx, conn, message.payload, message.runId, logger)
			}
			h.mu.RUnlock()
		}
	}
}

// serveWsInternal handles the actual WebSocket upgrade and client management.
func serveWsInternal(ctx context.Context, hub *Hub, logger util.Logger, w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] != "live" || pathParts[1] == "" {
		logger.Warn(fmt.Sprintf("[WS Handler] Invalid path format: %s", r.URL.Path))
		http.Error(w, "Not found. Use /live/:run-id", http.StatusNotFound)
		return
	}
	runId := pathParts[1]
	logger.Info(fmt.Sprintf("[WS Handler] Incoming WebSocket connection request for run-id: %s from %s", runId, r.RemoteAddr))

	opts := websocket.AcceptOptions{
		// Nothing as of now.
	}

	conn, err := websocket.Accept(w, r, &opts)
	if err != nil {
		status := websocket.CloseStatus(err)
		if status == -1 && !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
			logger.Error(fmt.Sprintf("[WS Handler] Failed to upgrade connection for run %s: %v", runId, err))
		}
		return
	}
	logger.Info(fmt.Sprintf("[WS Handler] WebSocket connection upgraded for run-id: %s", runId))

	client := &client{runId: runId, conn: conn}

	select {
	case hub.register <- client:
		logger.Info(fmt.Sprintf("[WS Handler] Client registered for run %s", runId))
	case <-ctx.Done():
		logger.Warn(fmt.Sprintf("[WS Handler] Server shutting down, closing new connection for run %s", runId))
		conn.Close(websocket.StatusGoingAway, "Server shutting down")
		return
	case <-time.After(2 * time.Second):
		logger.Error(fmt.Sprintf("[WS Handler] Timeout registering client for run %s. Hub may be blocked.", runId))
		conn.Close(websocket.StatusInternalError, "Server busy")
		return
	}

	go readPump(ctx, hub, client, logger)
}

// readPump pumps messages from the websocket connection to the hub AND writes to log file.
func readPump(ctx context.Context, hub *Hub, c *client, logger util.Logger) {
	defer func() {
		logger.Info(fmt.Sprintf("[WS ReadPump] Exiting for run-id: %s, triggering unregister", c.runId))
		select {
		case hub.unregister <- c:
		case <-ctx.Done():
		case <-time.After(1 * time.Second):
			logger.Warn(fmt.Sprintf("[WS ReadPump] Timeout waiting for unregister confirmation for run-id: %s", c.runId))
		}
		c.conn.Close(websocket.StatusNormalClosure, "Read loop finished")
	}()

	for {
		readCtx := ctx
		messageType, payload, err := c.conn.Read(readCtx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway || errors.Is(err, io.EOF) {
				logger.Info(fmt.Sprintf("[WS ReadPump] Client disconnected normally for run %s (Status: %d)", c.runId, status))
			} else if errors.Is(err, context.Canceled) {
				logger.Info(fmt.Sprintf("[WS ReadPump] Read canceled for run %s (likely server shutdown)", c.runId))
			} else if errors.Is(err, net.ErrClosed) {
				logger.Info(fmt.Sprintf("[WS ReadPump] Connection closed for run %s.", c.runId))
			} else {
				logger.Error(fmt.Sprintf("[WS ReadPump] Read error for run %s: %v (Status: %d)", c.runId, err, status))
			}
			return
		}

		if messageType == websocket.MessageText {
			err := logMessageToFile(c.runId, payload, logger)
			if err != nil {
				// Log the error but continue processing (broadcasting) the message.
				logger.Error(fmt.Sprintf("[WS LogToFile] Failed for run %s: %v", c.runId, err))
			}

			// Broadcast the message.
			msg := &message{runId: c.runId, payload: payload, sender: c.conn}
			select {
			case hub.broadcast <- msg:
			case <-ctx.Done():
				logger.Warn(fmt.Sprintf("[WS ReadPump] Server shutting down, could not broadcast message from run %s", c.runId))
				return
			case <-time.After(1 * time.Second):
				logger.Warn(fmt.Sprintf("[WS ReadPump] Timeout broadcasting message from run %s. Hub may be slow.", c.runId))
			}
		} else {
			logger.Warn(fmt.Sprintf("[WS ReadPump] Received non-text message type (%d) for run %s", messageType, c.runId))
		}
	}
}

// logMessageToFile handles appending a message payload to the run-specific log file.
func logMessageToFile(runId string, payload []byte, logger util.Logger) error {
	err := os.MkdirAll(liveLogDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create log directory '%s': %w", liveLogDir, err)
	}

	fileName := fmt.Sprintf("%s.str", runId)
	filePath := filepath.Join(liveLogDir, fileName)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file '%s': %w", filePath, err)
	}

	// Ensure file is closed even if errors occur during write.
	defer func() {
		if cerr := file.Close(); cerr != nil {
			logger.Error(fmt.Sprintf("[WS LogToFile] Error closing file '%s': %v", filePath, cerr))
		}
	}()

	lineToWrite := append(payload, '\n')

	_, err = file.Write(lineToWrite)
	if err != nil {
		return fmt.Errorf("failed to write to log file '%s': %w", filePath, err)
	}

	// logger.Info(fmt.Sprintf("[WS LogToFile] Appended message to %s", filePath))

	return nil
}

// StartServer initializes and runs the central WebSocket Hub.
func StartServer(ctx context.Context, logger util.Logger) *Hub {
	if globalHub != nil {
		logger.Warn("[WS] Server already started.")
		return globalHub
	}
	globalHub = newHub()
	go globalHub.run(ctx, logger)
	return globalHub
}

// GetHandler returns an http.HandlerFunc that uses the provided logger and hub.
func GetHandler(ctx context.Context, hub *Hub, logger util.Logger) http.HandlerFunc {
	if hub == nil {
		logger.Warn("[WS] GetHandler called with nil Hub. Using globalHub.")
		if globalHub == nil {
			panic("[WS] Fatal: GetHandler called before StartServer or with nil Hub and globalHub is also nil.")
		}
		hub = globalHub
	}
	return func(w http.ResponseWriter, r *http.Request) {
		serveWsInternal(ctx, hub, logger, w, r.WithContext(ctx))
	}
}
