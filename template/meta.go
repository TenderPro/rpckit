package template

import (
	"html/template"
	"net/http"

	"github.com/pkg/errors"

	base "github.com/apisite/apitpl/samplemeta"
)

// ErrRedirect is an error returned when page needs to be redirected
var ErrRedirect = errors.New("Abort with redirect")

// Meta holds template metadata
type Meta struct {
	base.Meta
	location string
	// store original message because error stack may change it
	errorMessage string
	JS           []string
	CSS          []string
}

// NewMeta returns new initialised Meta struct
func NewMeta(status int, ctype string) *Meta {
	m := base.NewMeta(status, ctype)
	return &Meta{Meta: *m}
}

// Raise interrupts template processing (if given) and raise error
func (p *Meta) Raise(status int, abort bool, message string) (string, error) {
	p.SetStatus(status)
	p.errorMessage = message // TODO: pass it via error only
	if abort {
		return "", errors.New(message)
	}
	return "", nil
}

// RedirectFound interrupts template processing and return redirect with StatusFound status
func (p *Meta) RedirectFound(uri string) (string, error) {
	p.SetStatus(http.StatusFound)
	p.location = uri
	return "", ErrRedirect // TODO: Is there a way to pass status & title via error?
}

// ErrorMessage returns internal or template error message
func (p Meta) ErrorMessage() string {
	if p.errorMessage != "" {
		return p.errorMessage
	}
	if p.Error() == nil {
		return ""
	}
	return p.Error().Error()
}

// Location returns redirect location
func (p Meta) Location() string { return p.location }

// AddJS - add .js file to scripts list
func (p *Meta) AddJS(file string) (string, error) {
	p.JS = append(p.JS, file)
	return "", nil
}

// AddCSS - add .css file to styles list
func (p *Meta) AddCSS(file string) (string, error) {
	p.JS = append(p.CSS, file)
	return "", nil
}

// SetProtoFuncs appends function templates and not related to request functions to funcs
// This func is not for use in templates
func SetProtoFuncs(funcs template.FuncMap) {
	funcs["request"] = func() interface{} { return nil }
	funcs["param"] = func(key string) string { return "" }
}

// SetRequestFuncs appends funcs which return real data inside request processing
// This func is not for use in templates
/*
TODO
func SetRequestFuncs(funcs template.FuncMap, ctx *gin.Context) {
	funcs["request"] = func() interface{} { return ctx.Request }
	funcs["param"] = func(key string) string { return ctx.Param(key) }
}
*/
