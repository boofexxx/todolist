package handlers

import "net/http"

func (s *ServerMux) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Printf("handle %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *ServerMux) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || !isValid(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Authenticate", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isValid(username, password string) bool {
	if username != "me" || password != "me" {
		return false
	}
	return true
}
