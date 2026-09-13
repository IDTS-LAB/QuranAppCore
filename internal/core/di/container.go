package di

import (
	"github.com/fiqrikm18/quran-app/internal/core/infrastructure/storage"
	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Container struct {
	Router  *chi.Mux
	DB      *gorm.DB
	Logger  *zerolog.Logger
	Tokens  security.TokenGate
	Storage *storage.S3
}

func NewContainer(r *chi.Mux, db *gorm.DB, logger zerolog.Logger, tokens security.TokenGate, store *storage.S3) *Container {
	return &Container{
		Router:  r,
		DB:      db,
		Logger:  &logger,
		Tokens:  tokens,
		Storage: store,
	}
}
