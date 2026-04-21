package server

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

var subscriberCounter atomic.Int64

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := strconv.FormatInt(subscriberCounter.Add(1), 10)
	ch := s.bus.Subscribe(id)
	defer s.bus.Unsubscribe(id)

	snap := s.bus.Snapshot()
	for _, data := range snap.Metrics {
		fmt.Fprintf(w, "event: metric\ndata: %s\n\n", data)
	}
	flusher.Flush()

	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, evt.Data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
