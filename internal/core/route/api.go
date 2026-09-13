package route

import (
	"context"
	"net/http"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/di"
	"github.com/fiqrikm18/quran-app/internal/core/middleware"
	userApplication "github.com/fiqrikm18/quran-app/internal/modules/user/application"
	userCmd "github.com/fiqrikm18/quran-app/internal/modules/user/application/command"
	userQuery "github.com/fiqrikm18/quran-app/internal/modules/user/application/query"
	userInfra "github.com/fiqrikm18/quran-app/internal/modules/user/infrastructure/database"
	userPresentation "github.com/fiqrikm18/quran-app/internal/modules/user/presentation/http"
	"github.com/fiqrikm18/quran-app/internal/shared/errors"
	httpResponse "github.com/fiqrikm18/quran-app/internal/shared/http"
	"github.com/go-chi/chi/v5"
)

// HealthzError mirrors the error object inside the shared response envelope.
type HealthzError struct {
	Code    string `json:"code" example:"SERVICE_UNAVAILABLE"`
	Details any    `json:"details,omitempty"`
}

// HealthzResponse mirrors the shared response envelope for API docs.
type HealthzResponse struct {
	Success bool              `json:"success" example:"true"`
	Message string            `json:"message" example:"ok"`
	Data    map[string]string `json:"data,omitempty"`
	Error   *HealthzError     `json:"error,omitempty"`
}

// Healthz reports liveness and Postgres reachability.
//
//	@Summary		Liveness + DB readiness probe
//	@Description	Returns 200 when the API is up and Postgres is reachable, 503 otherwise.
//	@Tags			Healthz
//	@Produce		json
//	@Success		200	{object}	HealthzResponse
//	@Failure		503	{object}	HealthzResponse
//	@Router			/healthz [get]
func HealthzHandler(container *di.Container) middleware.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := container.DB.DB()
		if err == nil {
			err = sqlDB.PingContext(ctx)
		}
		if err != nil {
			container.Logger.Error().
				Err(err).
				Msg("health check: database unreachable")
			return errors.New(
				http.StatusServiceUnavailable,
				"SERVICE_UNAVAILABLE",
				"Service temporarily unavailable. Please try again in a moment.",
				errors.ActionDetails(errors.ActionRetryLater),
			)
		}

		httpResponse.Success(w, http.StatusOK, "ok", map[string]string{
			"status":   "up",
			"database": "up",
		})
		return nil
	}
}

var (
	userRepository    userApplication.IUserRepository
	sessionRepository userApplication.IUserSessionRepository
	passwordHasher    userApplication.IPasswordHasher

	registerCommand     *userCmd.RegisterCommandHandler
	loginCommand        *userCmd.LoginCommandHandler
	refreshTokenCommand *userCmd.RefreshTokenCommandHandler
	verifyEmailCommand  *userCmd.VerifyEmailCommandHandler
	logoutCommand       *userCmd.LogoutCommandHandler
	meQuery             *userQuery.MeQueryHandler

	userHandler *userPresentation.Handler
)

func RegisterAPI(container *di.Container) {
	userRepository = userInfra.NewUserRepository(container.DB)
	sessionRepository = userInfra.NewUserSessionRepository(container.DB)
	passwordHasher = userInfra.NewPasswordHasher()

	registerCommand = userCmd.NewRegisterCommandHandler(userRepository, passwordHasher)
	loginCommand = userCmd.NewLoginCommandHandler(userRepository, sessionRepository, passwordHasher, container.Tokens)
	refreshTokenCommand = userCmd.NewRefreshTokenCommandHandler(userRepository, sessionRepository, container.Tokens)
	verifyEmailCommand = userCmd.NewVerifyEmailCommandHandler(userRepository, container.Tokens)
	logoutCommand = userCmd.NewLogoutCommandHandler(sessionRepository, container.Tokens)
	meQuery = userQuery.NewMeQueryHandler(userRepository)

	userHandler = userPresentation.NewHandler(
		registerCommand,
		loginCommand,
		refreshTokenCommand,
		verifyEmailCommand,
		logoutCommand,
		meQuery,
	)

	container.Router.Route("/api", func(r chi.Router) {
		r.Get("/healthz", middleware.Handle(*container.Logger, HealthzHandler(container)).ServeHTTP)

		r.Route("/v1", func(v1 chi.Router) {
			v1.Route("/auth", func(authRoute chi.Router) {
				// Public auth routes
				authRoute.Post("/register", middleware.Handle(*container.Logger, userHandler.Register).ServeHTTP)
				authRoute.Post("/login", middleware.Handle(*container.Logger, userHandler.Login).ServeHTTP)
				authRoute.Post("/refresh", middleware.Handle(*container.Logger, userHandler.RefreshToken).ServeHTTP)
				authRoute.Post("/verify-email", middleware.Handle(*container.Logger, userHandler.VerifyEmail).ServeHTTP)

				// Protected auth routes (require valid access token)
				authRoute.Group(func(protected chi.Router) {
					protected.Use(middleware.RequireAuthMiddleware(container.Tokens))
					protected.Post("/logout", middleware.Handle(*container.Logger, userHandler.Logout).ServeHTTP)
					protected.Get("/me", middleware.Handle(*container.Logger, userHandler.Me).ServeHTTP)
				})
			})
		})
	})
}
