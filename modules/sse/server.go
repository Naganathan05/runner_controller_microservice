package sse

import (
	"evolve/util"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/hpcloud/tail"
)

const (
	logDir       = "live"     // Directory where <runId>.str files are stored.
	runIdHeader  = "X-RUN-ID" // Header key for the run ID.
	retrySeconds = 3          // SSE retry interval suggestion for clients.
	sseDoneEvent = "done"     // Event name for the end of the stream.
)

// GetSSEHandler creates and returns an
// http.HandlerFunc for the SSE endpoint.
func GetSSEHandler(logger util.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveSSE(logger, w, r)
	}
}

// serveSSE handles incoming SSE requests.
func serveSSE(logger util.Logger, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get Run ID from Header.
	runId := r.Header.Get(runIdHeader)
	if runId == "" {
		logger.Warn(fmt.Sprintf("[SSE Handler] Missing %s header", runIdHeader))
		http.Error(w, fmt.Sprintf("Missing %s header", runIdHeader), http.StatusBadRequest)
		return
	}

	// Basic Sanitize Run ID (Prevent path traversal)
	if strings.Contains(runId, "..") || strings.ContainsAny(runId, "/\\") {
		logger.Warn(fmt.Sprintf("[SSE Handler] Invalid characters in %s header: %s", runIdHeader, runId))
		http.Error(w, "Invalid Run ID format", http.StatusBadRequest)
		return
	}

	logFileName := fmt.Sprintf("%s.str", runId)
	logFilePath := filepath.Join(logDir, logFileName)

	logger.Info(fmt.Sprintf("[SSE Handler] Request received for runId: %s (File: %s)", runId, logFilePath))

	// Set SSE Headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Suggest a retry interval to the client.
	fmt.Fprintf(w, "retry: %ds\n\n", retrySeconds)

	rc := http.NewResponseController(w)
	if rc == nil {
		logger.Error(fmt.Sprintf("[SSE Handler] Failed to get ResponseController for runId: %s", runId))
		http.Error(w, "Failed to get ResponseController", http.StatusInternalServerError)
		return
	}

	if err := rc.Flush(); err != nil {
		logger.Error(fmt.Sprintf("[SSE Handler] Error flushing response for runId %s: %v", runId, err))
		http.Error(w, "Error flushing response", http.StatusInternalServerError)
		return
	}

	// Configure and Start Tailing.
	tailConfig := tail.Config{
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekStart},
		ReOpen:    true,
		MustExist: false,
		Poll:      true,
		Follow:    true,
	}

	tailer, err := tail.TailFile(logFilePath, tailConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("[SSE Tailing] Failed to start tailing file %s for runId %s: %v", logFilePath, runId, err))
		// Don't send HTTP error here as headers are already sent. Client will retry or disconnect.
		return
	}

	// Ensure tailer is stopped when the handler exits.
	defer func() {
		logger.Info(fmt.Sprintf("[SSE Handler] Stopping tailer for runId: %s", runId))
		// Stopping the tailer closes its internal channels.
		if stopErr := tailer.Stop(); stopErr != nil {
			logger.Error(fmt.Sprintf("[SSE Tailing] Error stopping tailer for runId %s: %v", runId, stopErr))
		}
	}()

	logger.Info(fmt.Sprintf("[SSE Handler] Started tailing %s for runId: %s", logFilePath, runId))

	// Event Loop - Send lines and handle disconnect.
	for {
		select {
		case <-ctx.Done():
			// Client disconnected.
			logger.Info(fmt.Sprintf("[SSE Handler] Client disconnected for runId: %s", runId))
			return

		case line, ok := <-tailer.Lines:
			if !ok {
				// Channel closed, tailer might
				// have stopped or encountered an error.
				tailErr := tailer.Err()
				if tailErr != nil && tailErr != io.EOF {
					logger.Error(fmt.Sprintf("[SSE Tailing] Error during tailing for runId %s: %v", runId, tailErr))
				} else {
					logger.Info(fmt.Sprintf("[SSE Tailing] Tailer lines channel closed for runId: %s", runId))
				}
				return
			}

			// Debug log.
			fmt.Printf("[SSE Handler] New line for runId %s: %s\n", runId, line.Text)

			if line.Text == "__END__" {
				logger.Info(fmt.Sprintf("[SSE Handler] Detected END marker for run %s. Sending '%s' event and closing stream.", runId, sseDoneEvent))

				// Send SSE event indicating stream end.
				_, writeErr := fmt.Fprintf(w, "event: %s\ndata: {\"message\": \"Stream ended.\"}\n\n", sseDoneEvent)
				if writeErr != nil {
					logger.Warn(fmt.Sprintf("[SSE Handler] Error writing '%s' event for runId %s: %v", sseDoneEvent, runId, writeErr))
				}

				// Attempt to flush the final event.
				if err := rc.Flush(); err != nil {
					logger.Error(fmt.Sprintf("[SSE Handler] Error flushing final event for runId %s: %v", runId, err))
					return
				}
				return
			}

			// Format and send SSE message.
			// SSE format: "data: content\n\n"
			_, writeErr := fmt.Fprintf(w, "data: %s\n\n", line.Text)

			if writeErr != nil {
				logger.Warn(fmt.Sprintf("[SSE Handler] Error writing to client for runId %s: %v", runId, writeErr))
				return
			}

			if err := rc.Flush(); err != nil {
				logger.Error(fmt.Sprintf("[SSE Handler] Error flushing response for runId %s: %v", runId, err))
				return
			}

			// Sleep to prevent overwhelming the client.
			time.Sleep(50 * time.Millisecond)
		}
	}
}
