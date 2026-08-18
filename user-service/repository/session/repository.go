package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	apperror "user-service/common/error"
	"user-service/domain/model"

	redislib "github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "session:"

type repository struct {
	client redislib.Cmdable
}

type Repository interface {
	Set(context.Context, string, model.Session, time.Duration) error
	Get(context.Context, string) (*model.Session, error)
	Delete(context.Context, string) error
}

func NewRepository(client redislib.Cmdable) Repository {
	return &repository{
		client: client,
	}
}

func (r *repository) Set(ctx context.Context, accessToken string, session model.Session, ttl time.Duration) error {
	accessToken = strings.TrimSpace(accessToken)

	if accessToken == "" {
		return fmt.Errorf("%w: access token must not be empty", apperror.ErrInvalidArgument)
	}

	if session.UserID <= 0 {
		return fmt.Errorf("%w: session user_id must be greater than zero", apperror.ErrInvalidArgument)
	}

	if ttl <= 0 {
		return fmt.Errorf("%w: ttl must be greater than zero", apperror.ErrInvalidArgument)
	}

	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}

	if err := r.client.Set(ctx, sessionKey(accessToken), payload, ttl).Err(); err != nil {
		return fmt.Errorf("store session: %w", err)
	}
	return nil
}

func (r *repository) Get(ctx context.Context, accessToken string) (*model.Session, error) {
	accessToken = strings.TrimSpace(accessToken)

	if accessToken == "" {
		return nil, fmt.Errorf("%w: access token must not be empty", apperror.ErrInvalidArgument)
	}

	payload, err := r.client.Get(ctx, sessionKey(accessToken)).Bytes()

	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return nil, fmt.Errorf("get session: %w", err)
		}
	}

	var session model.Session
	if err := json.Unmarshal(payload, &session); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	if session.UserID <= 0 {
		return nil, fmt.Errorf("decode session: invalid user id")
	}

	return &session, nil
}

func (r *repository) Delete(ctx context.Context, accessToken string) error {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return fmt.Errorf("%w: access token must not be empty", apperror.ErrInvalidArgument)
	}

	deleted, err := r.client.Del(ctx, sessionKey(accessToken)).Result()

	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	if deleted == 0 {
		return fmt.Errorf("delete session: %w", apperror.ErrNotFound)
	}

	return nil
}

func sessionKey(accessToken string) string {
	digest := sha256.Sum256([]byte(accessToken))

	return sessionKeyPrefix + hex.EncodeToString(digest[:])
}
