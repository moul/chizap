package chizap_test

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"moul.io/chizap"
)

func Example() {
	logger := zap.NewExample()
	r := chi.NewRouter()
	r.Use(chizap.New(logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))
}
