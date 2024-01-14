package version

import "net/http"

// Semver is the latest semver version of powergrid.
var Semver = "v0.2.2"

// GitHash is the latest git commit, and is replaced at build time by CI.
var GitHash = "dev"
var Debug = GitHash == "dev"

var String = Semver + "+" + GitHash

func Middleware(prefix string, next http.Handler) http.Handler {
	serverString := prefix + "/" + String
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", serverString)
		next.ServeHTTP(w, r)
	})
}
