package router

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	openapi "github.com/go-openapi/runtime/middleware"
	"github.com/sirupsen/logrus"
	"gitlab.com/mooncascade/event-timing-server/athletes"
)

// New initializes router with provided service and logger
func New(logger *logrus.Logger, service *athletes.Service) *chi.Mux {
	r := chi.NewRouter()
	r.Use(loggerMiddleware(logger))
	r.Post("/update", service.ReceiveTimingEventHandler())
	r.Get("/leaderboard", service.LeaderboardHandler())
	r.Get("/ws", service.WSHandler())
	r.Get("/openapi", func(w http.ResponseWriter, r *http.Request) {
		openapi.Redoc(openapi.RedocOpts{Title: "Event timing server API", SpecURL: "docs/openapi.json", Path: "openapi"}, nil).ServeHTTP(w, r)
	})
	fileServer(r)
	return r
}

// fileServer for openapi docs
func fileServer(r chi.Router) {
	path := "/docs"
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(http.Dir("./docs")))
		fs.ServeHTTP(w, r)
	})
}

func loggerMiddleware(logger *logrus.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					remoteIP = r.RemoteAddr
				}
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				fields := logrus.Fields{
					"status":    ww.Status(),
					"duration":  time.Since(t1).String(),
					"remote_ip": remoteIP,
					"proto":     r.Proto,
					"method":    r.Method,
				}
				logger.WithFields(fields).Infof("%s://%s%s", scheme, r.Host, r.RequestURI)
			}()
			h.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
