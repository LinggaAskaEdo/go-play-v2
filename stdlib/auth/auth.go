package auth

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog"

	"github.com/linggaaskaedo/go-play-v2/stdlib/constanta/header"
	"github.com/linggaaskaedo/go-play-v2/stdlib/parser"
)

const (
	ErrLoadPubKey  = `Error load public key`
	ErrLoadPrivKey = `Error load private key`
)

type Auth interface {
	// Create()
	Validate(fn func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request)
}

type auth struct {
	logger     zerolog.Logger
	opt        Options
	json       parser.JSONParser
	localc     gcache.Cache
	verifyKey  *rsa.PublicKey
	decryptKey *rsa.PrivateKey
}

type Options struct {
	PrivateKeyPath      string
	PublicKeyPath       string
	CacheBin            int
	CacheExpirationTime time.Duration
}

func Init(logger zerolog.Logger, opt Options, parse parser.Parser) Auth {
	a := &auth{
		logger: logger,
		opt:    opt,
		json:   parse.JSONParser(),
		localc: gcache.New(opt.CacheBin).LFU().Expiration(opt.CacheExpirationTime).Build(),
	}

	verifyBytes, err := os.ReadFile(a.opt.PublicKeyPath)
	if err != nil {
		a.logger.Fatal().AnErr(ErrLoadPubKey, err)
	}

	a.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		a.logger.Fatal().AnErr(ErrLoadPubKey, err)
	}

	decryptBytes, err := os.ReadFile(a.opt.PrivateKeyPath)
	if err != nil {
		a.logger.Fatal().AnErr(ErrLoadPrivKey, err)
	}

	a.decryptKey, err = jwt.ParseRSAPrivateKeyFromPEM(decryptBytes)
	if err != nil {
		a.logger.Fatal().AnErr(ErrLoadPrivKey, err)
	}

	return a
}

func (a *auth) Validate(fn func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorizationHeader := r.Header.Get("Authorization")

		if !strings.Contains(authorizationHeader, "Bearer") {
			a.httpRespError(w, r, x.NewWithCode(svcerr.CodeHTTPUnauthorized, "Unauthorized"))
			return
		}

		tokenStr := strings.Replace(authorizationHeader, "Bearer ", "", -1)
	}
}

func (a *auth) httpRespError(w http.ResponseWriter, r *http.Request, err error) {
	debugMode := false
	lang := header.LangID
	if r.Header.Get(header.AppDebug) == "true" {
		debugMode = true
	}
	
	if r.Header.Get(header.AppLang) == header.LangEN {
		lang = header.LangEN
	}

	statusCode, displayError := apperr.Compile(apperr.COMMON, err, lang, debugMode)

	jsonErrResp := &HTTPErrResp{
		Meta: Meta{
			Path:       r.URL.String(),
			StatusCode: statusCode,
			Status:     http.StatusText(statusCode),
			Message:    fmt.Sprintf("%s %s [%d] %s", r.Method, r.URL.RequestURI(), statusCode, http.StatusText(statusCode)),
			Error:      &displayError,
			Timestamp:  time.Now().Format(time.RFC3339),
		},
	}

	a.logger.ErrorWithContext(r.Context(), displayError.Error())

	raw, err := a.json.Marshal(jsonErrResp)
	if err != nil {
		statusCode = http.StatusInternalServerError
		a.logger.ErrorWithContext(r.Context(), err)
	}

	w.Header().Set(header.ContentType, header.ContentJSON)
	w.WriteHeader(statusCode)
	w.Write(raw)
}
