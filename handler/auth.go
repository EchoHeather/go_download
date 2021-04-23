package handler

import "net/http"

//HTTPinterceptor http拦截器
func HTTPinterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")
			if len(username) < 3 || !IsTokenValid(username, token) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h(w, r)
		},

	)
}
