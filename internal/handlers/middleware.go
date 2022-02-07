package handlers

import "net/http"

func (s *ServerMux) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Printf("handle %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
