package httpmux

import (
	"errors"
	"net/http"
	"sync"

	"github.com/bmizerany/pat"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

type (
	Method string
)

var (
	once = &sync.Once{}
)

const (
	POST    Method = http.MethodPost
	GET     Method = http.MethodGet
	PUT     Method = http.MethodPut
	DELETE  Method = http.MethodDelete
	OPTIONS Method = http.MethodOptions
	CONNECT Method = http.MethodConnect
	HEAD    Method = http.MethodHead
	PATCH   Method = http.MethodPatch
)

type HttpMux interface {
	HandleFunc(method Method, path string, hf http.HandlerFunc)
	Handler() http.Handler
}

type httpRouterMux struct {
	logger zerolog.Logger
	mux    *pat.PatternServeMux
	opt    Options
	cors   *cors.Cors
}

type Options struct {
	Swagger SwaggerOptions
	Cors    CorsOptions
}

type SwaggerOptions struct {
	Enabled         bool
	Path            string
	DocFile         string
	BasicAuth       BasicAuthOptions
	SwaggerTemplate SwaggerTemplateOptions
}

type BasicAuthOptions struct {
	Username, Password string
}

type SwaggerTemplateOptions struct {
	BasicAuth    BasicAuthOptions
	Enabled      bool
	TemplateFile string
	Path         string
	GoTemplate   GoTemplateOptions
}

type GoTemplateOptions struct {
	Description string
	Title       string
	Version     string
	Schemes     string
	Host        string
	BasePath    string
}

type CorsOptions struct {
	Enabled            bool
	Mode               string
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	MaxAge             int
	AllowCredentials   bool
	OptionsPassthrough bool
	Debug              bool
}

func Init(logger zerolog.Logger, opt Options) HttpMux {
	var mux HttpMux

	once.Do(func() {
		var c *cors.Cors
		if opt.Cors.Enabled {
			switch opt.Cors.Mode {
			case "custom":
				c = cors.New(cors.Options{
					AllowedOrigins:     opt.Cors.AllowedOrigins,
					AllowedMethods:     opt.Cors.AllowedMethods,
					AllowedHeaders:     opt.Cors.AllowedHeaders,
					ExposedHeaders:     opt.Cors.ExposedHeaders,
					MaxAge:             opt.Cors.MaxAge,
					AllowCredentials:   opt.Cors.AllowCredentials,
					OptionsPassthrough: opt.Cors.OptionsPassthrough,
					Debug:              opt.Cors.Debug,
				})
			case "allowall":
				c = cors.AllowAll()
			case "default":
				c = cors.Default()
			default:
				c = nil
			}
		}

		httpRouterMux := &httpRouterMux{
			logger: logger,
			mux:    pat.New(),
			opt:    opt,
			cors:   c,
		}

		httpRouterMux.registerHTTPSwagger()
		mux = httpRouterMux
	})

	return mux
}

func basicAuthHandler(logger zerolog.Logger, authUser, authPassword string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		if !validateBasicAuth(authUser, "", authPassword, "") {
			username, password, ok := r.BasicAuth()
			if !ok {
				err := errors.New("Swagger Invalid Basic Auth")
				logger.Error().Err(err).Send()
				http.Error(w, http.StatusText(401), 401)
				return
			}

			if !validateBasicAuth(authUser, username, authPassword, password) {
				err := errors.New("Swagger Invalid Basic Auth")
				logger.Error().Err(err).Send()
				http.Error(w, http.StatusText(401), 401)
				return
			}
		}

		h.ServeHTTP(w, r)
	}
}

func validateBasicAuth(username, authUser, password, authPassword string) bool {
	if username != authUser || password != authPassword {
		return false
	}

	return true
}

func basicAuthHandlerFunc(logger zerolog.Logger, authUser, authPassword string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		if !validateBasicAuth(authUser, "", authPassword, "") {
			username, password, ok := r.BasicAuth()
			if !ok {
				err := errors.New("Swagger Invalid Basic Auth")
				logger.Error().Err(err).Send()
				http.Error(w, http.StatusText(401), 401)
				return
			}

			if !validateBasicAuth(authUser, username, authPassword, password) {
				err := errors.New("Swagger Invalid Basic Auth")
				logger.Error().Err(err).Send()
				http.Error(w, http.StatusText(401), 401)
				return
			}
		}
		h.ServeHTTP(w, r)
	}
}
