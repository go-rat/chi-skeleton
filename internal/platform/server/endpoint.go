// Package server assembles the HTTP layer from the modules' route contributions.
package server

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/contrib/openapi"

	"github.com/libtnb/chi-skeleton/internal/shared/registry"
)

// Version is the build version, injected by main; the OpenAPI document carries it.
type Version string

// NewVersion converts the build-time injector input to its distinct graph type.
func NewVersion(version string) Version {
	return Version(version)
}

// HTTP registers every route contribution on r.
func HTTP(groups registry.Routes, r chi.Router) {
	for _, endpoints := range groups {
		for _, e := range endpoints {
			r.Method(e.Method, e.Path, e.Handler)
		}
	}
}

// SpecJSON assembles the OpenAPI 3.1 document from every documented endpoint.
func SpecJSON(
	title string,
	version Version,
	validate *validator.Validator,
	groups registry.Routes,
) ([]byte, error) {
	if version == "" {
		version = "dev"
	}
	g, err := openapi.New(title, string(version),
		openapi.WithValidator(validate),
		openapi.WithSchema[time.Time](&openapi.Schema{Type: "string", Format: "date-time"}),
	)
	if err != nil {
		return nil, err
	}
	for _, endpoints := range groups {
		for _, e := range endpoints {
			if e.Document == nil {
				continue
			}
			if err := e.Document(g, e.Method, e.Path, e.Summary, e.Tags); err != nil {
				return nil, err
			}
		}
	}

	return g.JSON()
}
