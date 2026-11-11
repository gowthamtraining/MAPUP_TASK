package handler

import (
	"mapup-geofence/internal/websocket"
	"net/http"
)

type WebSocketHandler struct {
	hub *websocket.Hub
}

func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// ServeWs handles GET /ws/alerts
func (h *WebSocketHandler) ServeWs(w http.ResponseWriter, r *http.Request) {
	websocket.ServeWs(h.hub, w, r)
}
