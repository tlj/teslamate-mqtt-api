package handlers

import "net/http"

type AuthMiddleware struct {
	apiKey string
}

func NewAuthMiddleware(apiKey string) *AuthMiddleware {
	return &AuthMiddleware{
		apiKey: apiKey,
	}
}

func (ah *AuthMiddleware) ApiKeyAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.URL.Query().Get("api_key")
			if apiKey != ah.apiKey {
				w.Write([]byte(`{"response":"invalid or missing api_key"}`))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}