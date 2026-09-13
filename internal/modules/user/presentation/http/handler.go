package http

import (
	"net/http"

	"github.com/fiqrikm18/quran-app/internal/core/middleware"
	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
	"github.com/fiqrikm18/quran-app/internal/modules/user/application/command"
	"github.com/fiqrikm18/quran-app/internal/modules/user/application/query"

	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
	httpUtils "github.com/fiqrikm18/quran-app/internal/shared/http"
)

type Handler struct {
	registerCmd     *command.RegisterCommandHandler
	loginCmd        *command.LoginCommandHandler
	refreshTokenCmd *command.RefreshTokenCommandHandler
	verifyEmailCmd  *command.VerifyEmailCommandHandler
	logoutCmd       *command.LogoutCommandHandler
	meQuery         *query.MeQueryHandler
}

func NewHandler(
	registerCmd *command.RegisterCommandHandler,
	loginCmd *command.LoginCommandHandler,
	refreshTokenCmd *command.RefreshTokenCommandHandler,
	verifyEmailCmd *command.VerifyEmailCommandHandler,
	logoutCmd *command.LogoutCommandHandler,
	meQuery *query.MeQueryHandler,
) *Handler {
	return &Handler{
		registerCmd:     registerCmd,
		loginCmd:        loginCmd,
		refreshTokenCmd: refreshTokenCmd,
		verifyEmailCmd:  verifyEmailCmd,
		logoutCmd:       logoutCmd,
		meQuery:         meQuery,
	}
}

// Register creates a new user account.
//
//	@Summary		Register a new user
//	@Description	Creates a user account with email, password and display name.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		application.RegisterRequest	true	"Registration payload"
//	@Success		201		{object}	RegisterResponse
//	@Failure		400		{object}	RegisterResponse
//	@Failure		409		{object}	RegisterResponse
//	@Router			/v1/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var req application.RegisterRequest
	if err := httpUtils.Decode(r, &req); err != nil {
		return err
	}

	if err := h.registerCmd.Exec(r.Context(), &command.RegisterCommand{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}); err != nil {
		return err
	}

	httpUtils.Success(w, http.StatusCreated, "user register successful", nil)
	return nil
}

// Login authenticates a user and returns access and refresh tokens.
//
//	@Summary		User login
//	@Description	Authenticates user with email and password, returns access and refresh tokens.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		application.LoginRequest	true	"Login payload"
//	@Success		200		{object}	application.LoginResponse
//	@Failure		400		{object}	application.LoginResponse
//	@Failure		401		{object}	application.LoginResponse
//	@Router			/v1/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req application.LoginRequest
	if err := httpUtils.Decode(r, &req); err != nil {
		return err
	}

	resp, err := h.loginCmd.Exec(r.Context(), &command.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	httpUtils.Success(w, http.StatusOK, "login successful", resp)
	return nil
}

// RefreshToken exchanges a refresh token for a new access and refresh token pair.
//
//	@Summary		Refresh access token
//	@Description	Exchanges a valid refresh token for a new access and refresh token pair.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		application.RefreshTokenRequest	true	"Refresh token payload"
//	@Success		200		{object}	application.RefreshTokenResponse
//	@Failure		400		{object}	application.RefreshTokenResponse
//	@Failure		401		{object}	application.RefreshTokenResponse
//	@Router			/v1/auth/refresh [post]
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	var req application.RefreshTokenRequest
	if err := httpUtils.Decode(r, &req); err != nil {
		return err
	}

	resp, err := h.refreshTokenCmd.Exec(r.Context(), &command.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return err
	}

	httpUtils.Success(w, http.StatusOK, "token refreshed successfully", resp)
	return nil
}

// VerifyEmail verifies a user's email address using a verification token.
//
//	@Summary		Verify email address
//	@Description	Verifies user's email address using the token sent via email.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		application.VerifyEmailRequest	true	"Email verification payload"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Router			/v1/auth/verify-email [post]
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) error {
	var req application.VerifyEmailRequest
	if err := httpUtils.Decode(r, &req); err != nil {
		return err
	}

	if err := h.verifyEmailCmd.Exec(r.Context(), &command.VerifyEmailCommand{
		Token: req.Token,
	}); err != nil {
		return err
	}

	httpUtils.Success(w, http.StatusOK, "email verified successfully", nil)
	return nil
}

// Logout revokes the user's refresh token. Requires authentication.
//
//	@Summary		User logout
//	@Description	Revokes the authenticated user's refresh token, invalidating the session.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		application.LogoutRequest	true	"Logout payload"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Router			/v1/auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	var req application.LogoutRequest
	if err := httpUtils.Decode(r, &req); err != nil {
		return err
	}

	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		return appErrors.UnauthorizedWithDetails("Please log in to continue.", appErrors.ActionDetails(appErrors.ActionLogin))
	}

	if err := h.logoutCmd.Exec(r.Context(), &command.LogoutCommand{
		RefreshToken: req.RefreshToken,
		UserID:       uint64(claims.UserID),
	}); err != nil {
		return err
	}

	httpUtils.Success(w, http.StatusOK, "logout successful", nil)
	return nil
}

// Me returns the authenticated user's profile.
//
//	@Summary		Get current user profile
//	@Description	Returns the authenticated user's profile information.
//	@Tags			Authentication
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	application.MeResponse
//	@Failure		401	{object}	application.MeResponse
//	@Router			/v1/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) error {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		return appErrors.UnauthorizedWithDetails("Please log in to continue.", appErrors.ActionDetails(appErrors.ActionLogin))
	}

	user, err := h.meQuery.Exec(r.Context(), &query.MeQuery{
		UserID: uint64(claims.UserID),
	})
	if err != nil {
		return err
	}

	resp := application.MeResponse{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: !user.EmailVerifiedAt.IsZero(),
		CreatedAt:     user.CreatedAt,
	}

	httpUtils.Success(w, http.StatusOK, "user profile retrieved", resp)
	return nil
}
