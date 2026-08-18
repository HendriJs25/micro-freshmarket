package jwt

import (
	"fmt"
	"strings"
	"time"
	apperror "user-service/common/error"
	"user-service/config"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	UserID    int64
	Issuer    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type AccessToken struct {
	Value     string
	ExpiresAt time.Time
}

type claims struct {
	UserID int64 `json:"user_id"`

	jwtlib.RegisteredClaims
}

type service struct {
	secretKey      []byte
	issuer         string
	accessTokenTTL time.Duration
}

type Service interface {
	GenerateAccessToken(int64) (*AccessToken, error)
	ValidateAccessToken(string) (*AccessTokenClaims, error)
}

func NewService(cfg config.JWT) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate jwt configuration: %w", err)
	}

	return &service{
		secretKey:      []byte(cfg.SecretKey),
		issuer:         cfg.Issuer,
		accessTokenTTL: cfg.AccessTokenTTL,
	}, nil
}

func (s *service) GenerateAccessToken(userID int64) (*AccessToken, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w:user id must be greater than zero", apperror.ErrInvalidArgument)
	}

	now := time.Now().UTC()

	expiresAt := now.Add(s.accessTokenTTL)

	tokenClaims := claims{
		UserID: userID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, tokenClaims)

	tokenString, err := token.SignedString(s.secretKey)

	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	return &AccessToken{
		Value:     tokenString,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *service) ValidateAccessToken(accessToken string) (*AccessTokenClaims, error) {
	tokenString := strings.TrimSpace(accessToken)

	if tokenString == "" {
		return nil, fmt.Errorf("%w: access token must not be empty", apperror.ErrInvalidToken)
	}

	tokenClaims := &claims{}

	token, err := jwtlib.ParseWithClaims(
		tokenString,
		tokenClaims,
		func(token *jwtlib.Token) (any, error) { return s.secretKey, nil },
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithIssuer(s.issuer),
		jwtlib.WithExpirationRequired(),
		jwtlib.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: access token validation failed: %v", apperror.ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("%w: access token is not valid", apperror.ErrInvalidToken)
	}

	if tokenClaims.UserID <= 0 {
		return nil, fmt.Errorf("%w: access token user id is invalid", apperror.ErrInvalidToken)
	}

	if tokenClaims.IssuedAt == nil {
		return nil, fmt.Errorf("%w: access token issued-at is missing", apperror.ErrInvalidToken)
	}

	if tokenClaims.ExpiresAt == nil {
		return nil, fmt.Errorf("%w: access token expiration is missing", apperror.ErrInvalidToken)
	}

	return &AccessTokenClaims{
		UserID:    tokenClaims.UserID,
		Issuer:    tokenClaims.Issuer,
		IssuedAt:  tokenClaims.IssuedAt.Time,
		ExpiresAt: tokenClaims.ExpiresAt.Time,
	}, nil
}
