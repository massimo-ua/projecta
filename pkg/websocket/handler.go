package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
)

type originValidator = func(origin string) bool

func createOriginValidator(allowedOrigins []string) func(origin string) bool {
	return func(origin string) bool {
		for _, allowedOrigin := range allowedOrigins {
			if strings.HasSuffix(origin, allowedOrigin) {
				return true
			}
		}
		return false
	}
}

func createWsUpgradeHandler(isAllowedOrigin originValidator) websocket.Upgrader {
	return websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return isAllowedOrigin(origin)
		},
	}
}

func makeWsHandler(a *AppAdapter, allowedOrigins []string) http.HandlerFunc {
	u := createWsUpgradeHandler(createOriginValidator(allowedOrigins))
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := u.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}

			response, err := a.Handle(IncomingMessage{
				MessageType: mt,
				Payload:     message,
			})

			if err != nil {
				continue
			}

			err = c.WriteMessage(mt, response)
			if err != nil {
				break
			}
		}
	}
}

func CreateWsHandler(allowedOrigins []string) http.HandlerFunc {
	a := NewAppAdapter()
	return makeWsHandler(a, allowedOrigins)
}
