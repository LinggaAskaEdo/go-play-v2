package parser

import (
	"errors"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/xeipuuv/gojsonschema"
)

type jsonConfig string

const (
	// JSONConfigDefault set :
	//	EscapeHTML :					true
	JSONConfigDefault jsonConfig = `default`

	// JSONConfigCompatibleWithStdLibrary
	//  EscapeHTML:             		true
	//  SortMapKeys:					true
	//  ValidateJsonRawMessage:			true
	JSONConfigCompatibleWithStdLibrary jsonConfig = `standard`

	// JSONConfigFastest
	//  EscapeHTML:                    	false
	//  MarshalFloatWith6Digits:       	true
	//  ObjectFieldMustBeSimpleString: 	true
	JSONConfigFastest jsonConfig = `fastest`

	// JSONConfigCustom
	//	Custom Configuration which is set in JSONOptions
	JSONConfigCustom jsonConfig = `custom`
)

type JSONParser interface {
	Marshal(orig interface{}) ([]byte, error)
	MarshalWithSchemaValidation(sch string, orig interface{}) ([]byte, error)
	Unmarshal(blob []byte, dest interface{}) error
	UnmarshalWithSchemaValidation(sch string, blob []byte, dest interface{}) error
}

type jsonparser struct {
	logger zerolog.Logger
	schema map[string]*gojsonschema.Schema
	API    jsoniter.API
	opt    JSONOptions
}

type JSONOptions struct {
	Config                        jsonConfig
	IndentionStep                 int
	MarshalFloatWith6Digits       bool
	EscapeHTML                    bool
	SortMapKeys                   bool
	UseNumber                     bool
	DisallowUnknownFields         bool
	TagKey                        string
	OnlyTaggedField               bool
	ValidateJSONRawMessage        bool
	ObjectFieldMustBeSimpleString bool
	CaseSensitive                 bool
	Schema                        map[string]string
}

func initJSONP(logger zerolog.Logger, opt JSONOptions) JSONParser {
	var jsonAPI jsoniter.API
	switch opt.Config {

	case JSONConfigDefault:
		jsonAPI = jsoniter.ConfigDefault

	case JSONConfigFastest:
		jsonAPI = jsoniter.ConfigFastest

	case JSONConfigCompatibleWithStdLibrary:
		jsonAPI = jsoniter.ConfigCompatibleWithStandardLibrary

	case JSONConfigCustom:
		jsonAPI = jsoniter.Config{
			IndentionStep:                 opt.IndentionStep,
			MarshalFloatWith6Digits:       opt.MarshalFloatWith6Digits,
			EscapeHTML:                    opt.EscapeHTML,
			SortMapKeys:                   opt.SortMapKeys,
			UseNumber:                     opt.UseNumber,
			DisallowUnknownFields:         opt.DisallowUnknownFields,
			TagKey:                        opt.TagKey,
			OnlyTaggedField:               opt.OnlyTaggedField,
			ValidateJsonRawMessage:        opt.ValidateJSONRawMessage,
			ObjectFieldMustBeSimpleString: opt.ObjectFieldMustBeSimpleString,
			CaseSensitive:                 opt.CaseSensitive,
		}.Froze()

	default:
		jsonAPI = jsoniter.ConfigCompatibleWithStandardLibrary
	}
	p := &jsonparser{
		logger: logger,
		API:    jsonAPI,
		opt:    opt,
		schema: make(map[string]*gojsonschema.Schema),
	}

	p.initSchema(opt.Schema)

	return p
}

func (p *jsonparser) initSchema(sources map[string]string) {
	for sch, src := range sources {
		schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(src))
		if err != nil {
			p.logger.Panic().Err(err).Msg(fmt.Sprintf("error on load : %s", sch))
			return
		}

		p.schema[sch] = schema
	}
}

func (p *jsonparser) Marshal(orig interface{}) ([]byte, error) {
	stream := p.API.BorrowStream(nil)
	defer p.API.ReturnStream(stream)

	stream.WriteVal(orig)
	result := make([]byte, stream.Buffered())
	if stream.Error != nil {
		return nil, stream.Error
	}

	copy(result, stream.Buffer())

	return result, nil
}

func (p *jsonparser) MarshalWithSchemaValidation(sch string, orig interface{}) ([]byte, error) {
	blob, err := p.Marshal(orig)
	if err != nil {
		return nil, err
	}

	s, ok := p.schema[sch]
	if !ok {
		return nil, errors.New(fmt.Sprintf("schema not found : %s", sch))
	}

	blobLoader := gojsonschema.NewBytesLoader(blob)
	res, err := s.Validate(blobLoader)
	if err != nil {
		return nil, err
	}

	if !res.Valid() {
		var errString []string
		for _, desc := range res.Errors() {
			errString = append(errString, fmt.Sprintf("- %s", desc))
		}

		return nil, errors.New(strings.Join(errString, "\n"))
	}

	return blob, nil
}

func (p *jsonparser) Unmarshal(blob []byte, dest interface{}) error {
	iter := p.API.BorrowIterator(blob)
	defer p.API.ReturnIterator(iter)

	iter.ReadVal(dest)
	if iter.Error != nil {
		return iter.Error
	}

	return nil
}

func (p *jsonparser) UnmarshalWithSchemaValidation(sch string, blob []byte, dest interface{}) error {
	s, ok := p.schema[sch]
	if !ok {
		return errors.New(fmt.Sprintf("schema not found : %s", sch))
	}

	blobLoader := gojsonschema.NewBytesLoader(blob)
	res, err := s.Validate(blobLoader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		var errString []string
		for _, desc := range res.Errors() {
			errString = append(errString, fmt.Sprintf("- %s", desc))
		}

		return errors.New(strings.Join(errString, "\n"))
	}

	return p.Unmarshal(blob, dest)
}
