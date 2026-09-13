package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/config"
	"github.com/golang-jwt/jwt/v5"
)

// Token types bound into the "token_type" claim so a long-lived refresh
// token can never be used where an access token is expected and vice versa.
const (
	TokenTypeAccess      = "access"
	TokenTypeRefresh     = "refresh"
	TokenTypeVerifyEmail = "verify_email"
)

type JwtClaim struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role,omitempty"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type JwtProvider struct {
	config *config.Config
}

// TokenGate is the single gate for everything token-related: callers depend
// on this interface (handy for DI and mocks) instead of reaching for
// separate generate/validate helpers.
type TokenGate interface {
	// Issue mints an access + refresh pair together.
	Issue(userID uint, email, role string) (TokenPair, error)
	// ValidateAccessToken accepts only token_type=access tokens.
	ValidateAccessToken(tokenStr string) (*JwtClaim, error)
	// ValidateRefreshToken accepts only token_type=refresh tokens.
	ValidateRefreshToken(tokenStr string) (*JwtClaim, error)
	// ValidateVerifyEmailToken accepts only token_type=verify_email tokens.
	ValidateVerifyEmailToken(tokenStr string) (*JwtClaim, error)
	// GenerateVerifyEmailToken mints a verification token.
	GenerateVerifyEmailToken(userID uint, email string) (string, error)
}

// Compile-time guarantee that the provider satisfies the gate.
var _ TokenGate = (*JwtProvider)(nil)

// TokenPair is the complete credential set handed to a client at login.
// Expiries are stamped at issuance so callers can report them (e.g. in the
// login response) without parsing TTL config or the tokens themselves.
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

func NewJwtProvider(cfg *config.Config) *JwtProvider {
	return &JwtProvider{
		config: cfg,
	}
}

// Issue mints both tokens through the same signer so the pair always shares
// one secret, one algorithm, and one claims shape. Either half failing fails
// the whole issuance — never hand out a lone access token.
func (provider *JwtProvider) Issue(userID uint, email, role string) (TokenPair, error) {
	accessTTL, err := provider.config.JWT.AccessDuration()
	if err != nil {
		return TokenPair{}, err
	}
	refreshTTL, err := provider.config.JWT.RefreshDuration()
	if err != nil {
		return TokenPair{}, err
	}

	now := time.Now()
	access, err := provider.GenerateToken(userID, email, role)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := provider.GenerateRefreshToken(userID)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresAt:  now.Add(accessTTL),
		RefreshExpiresAt: now.Add(refreshTTL),
	}, nil
}

func (provider *JwtProvider) GenerateToken(userID uint, email, role string) (string, error) {
	accessTTL, err := provider.config.JWT.AccessDuration()
	if err != nil {
		return "", err
	}
	return provider.sign(userID, email, role, TokenTypeAccess, accessTTL, "")
}

// GenerateRefreshToken mints a long-lived refresh token. It carries only the
// user ID plus a random jti (for future rotation/revocation lists) — never
// email/role authorizations, which are re-resolved when the token is redeemed.
func (provider *JwtProvider) GenerateRefreshToken(userID uint) (string, error) {
	refreshTTL, err := provider.config.JWT.RefreshDuration()
	if err != nil {
		return "", err
	}

	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return provider.sign(userID, "", "", TokenTypeRefresh, refreshTTL, hex.EncodeToString(raw[:]))
}

// GenerateVerifyEmailToken mints a one-time email verification token.
func (provider *JwtProvider) GenerateVerifyEmailToken(userID uint, email string) (string, error) {
	verifyTTL, err := provider.config.JWT.AccessDuration()
	if err != nil {
		return "", err
	}
	return provider.sign(userID, email, "", TokenTypeVerifyEmail, verifyTTL, "")
}

// sign is the single choke point for minting tokens: one secret source,
// one algorithm, one claims shape.
func (provider *JwtProvider) sign(userID uint, email, role, tokenType string, ttl time.Duration, jti string) (string, error) {
	secret := provider.config.JWT.Secret
	if secret == "" {
		return "", errors.New("jwt secret is not configured")
	}

	now := time.Now()
	claims := JwtClaim{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    provider.config.JWT.Issuer,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Sentinel errors so callers can branch (e.g. expired → refresh flow)
// without string-matching. All other failures collapse to
// ErrTokenInvalid to avoid leaking why a token was rejected.
var (
	ErrTokenExpired = errors.New("token is expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// ValidateToken parses and verifies a token string. It rejects:
//   - empty input and unconfigured secrets
//   - "none" and any non-HS256 algorithm (algorithm-confusion defense,
//     checked both in the keyfunc and via WithValidMethods)
//   - bad signatures, malformed tokens, wrong issuer, missing/future expiry
//     (zero leeway)
func (provider *JwtProvider) ValidateToken(tokenStr string) (*JwtClaim, error) {
	jwtCfg := provider.config.JWT
	if jwtCfg.Secret == "" {
		return nil, errors.New("jwt secret is not configured")
	}
	if strings.TrimSpace(tokenStr) == "" {
		return nil, ErrTokenInvalid
	}

	claims := &JwtClaim{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(jwtCfg.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(0),
	)

	token, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrTokenInvalid
		}
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(jwtCfg.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// ValidateAccessToken is ValidateToken plus a binding to token_type=access,
// so refresh tokens are unusable as access tokens.
func (provider *JwtProvider) ValidateAccessToken(tokenStr string) (*JwtClaim, error) {
	return provider.validateTyped(tokenStr, TokenTypeAccess)
}

// ValidateRefreshToken is ValidateToken plus a binding to token_type=refresh.
func (provider *JwtProvider) ValidateRefreshToken(tokenStr string) (*JwtClaim, error) {
	return provider.validateTyped(tokenStr, TokenTypeRefresh)
}

// ValidateVerifyEmailToken is ValidateToken plus a binding to token_type=verify_email.
func (provider *JwtProvider) ValidateVerifyEmailToken(tokenStr string) (*JwtClaim, error) {
	return provider.validateTyped(tokenStr, TokenTypeVerifyEmail)
}

func (provider *JwtProvider) validateTyped(tokenStr, wantType string) (*JwtClaim, error) {
	claims, err := provider.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != wantType {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
