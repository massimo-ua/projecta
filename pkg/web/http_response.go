package web

import (
	"context"
	"encoding/json"
	ht "github.com/go-kit/kit/transport/http"
	"net/http"
)

func encodeJSON(statusCode int) ht.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, response any) error {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		w.WriteHeader(statusCode)

		if statusCode == http.StatusNoContent {
			return nil
		}

		return json.NewEncoder(w).Encode(response)
	}
}
