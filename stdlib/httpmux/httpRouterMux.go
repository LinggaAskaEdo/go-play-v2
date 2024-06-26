package httpmux

import (
	"html/template"
	"net/http"

	swagger "github.com/swaggo/http-swagger"
)

func (m *httpRouterMux) Handler() http.Handler {
	if m.cors != nil {
		return m.cors.Handler(m.mux)
	}

	return m.mux
}

func (m *httpRouterMux) HandleFunc(method Method, path string, hf http.HandlerFunc) {
	m.handleFunc(method, path, hf)
}

func (m *httpRouterMux) registerHTTPSwagger() {
	if m.opt.Swagger.Enabled {
		m.mux.Get(m.opt.Swagger.Path,
			basicAuthHandler(
				m.logger,
				m.opt.Swagger.BasicAuth.Username,
				m.opt.Swagger.BasicAuth.Password,
				swagger.Handler(
					swagger.URL(m.opt.Swagger.DocFile),
				)))

		if m.opt.Swagger.SwaggerTemplate.Enabled {
			m.handleFunc(GET, m.opt.Swagger.SwaggerTemplate.Path,
				basicAuthHandlerFunc(
					m.logger,
					m.opt.Swagger.SwaggerTemplate.BasicAuth.Username,
					m.opt.Swagger.SwaggerTemplate.BasicAuth.Password,
					m.swaggerTemplate))
		}
	}
}

func (m *httpRouterMux) handleFunc(method Method, path string, hf http.HandlerFunc) {
	m.mux.Add(string(method), path, hf)
}

func (m *httpRouterMux) swaggerTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(m.opt.Swagger.SwaggerTemplate.TemplateFile)
	if err != nil {
		m.logger.Error().Msg(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.Execute(w, m.opt.Swagger.SwaggerTemplate.GoTemplate); err != nil {
		m.logger.Error().Msg(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
}
