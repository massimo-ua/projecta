package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"net/http"
	"strings"
)

const (
	authQueryParam = "token"
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

func makeWsHandler(a *AppAdapterImpl, allowedOrigins []string, authorizer authorizer) http.HandlerFunc {
	u := createWsUpgradeHandler(createOriginValidator(allowedOrigins))
	return func(w http.ResponseWriter, r *http.Request) {
		aToken := r.URL.Query().Get(authQueryParam)
		_, err := authorizer(aToken)

		// TODO: review this solution
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		c, err := u.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				continue
			}

			var payload PayloadDTO

			err = json.Unmarshal(message, &payload)

			if err != nil {
				continue
			}

			requesterID, err := authorizer(payload.Token)

			ctx := createAuthorizedContext(requesterID)

			response, err := a.Handle(ctx, payload.Type, payload.Data)

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

func CreateWsHandler(allowedOrigins []string, provider core.AuthTokenProvider) http.HandlerFunc {
	a := NewAppAdapter()
	authorizer := createJwtAuthorizer(provider)
	return makeWsHandler(a, allowedOrigins, authorizer)
}
