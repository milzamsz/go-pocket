package middleware

import (
	"errors"
	"strings"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/pocketbase/pocketbase/core"
)

type ActorContext struct {
	UserID      string
	Email       string
	IsSuperuser bool
}

const actorContextStoreKey = "actorContext"

func BindAuthContext() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := resolveAuthRecord(e)
		if err == nil && record != nil {
			e.Auth = record
			SetActorContext(e, ActorContext{
				UserID:      record.Id,
				Email:       record.Email(),
				IsSuperuser: record.IsSuperuser(),
			})
		}
		return e.Next()
	}
}

func RequireAuthenticated() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := CurrentActor(e); err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		return e.Next()
	}
}

func RequireSuperuser() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := CurrentActor(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		if !actor.IsSuperuser {
			return e.ForbiddenError("superuser access required", domain.ErrForbidden)
		}
		return e.Next()
	}
}

func SetActorContext(e *core.RequestEvent, actor ActorContext) {
	e.Set(actorContextStoreKey, actor)
}

func CurrentActor(e *core.RequestEvent) (ActorContext, error) {
	if actor, ok := e.Get(actorContextStoreKey).(ActorContext); ok && actor.UserID != "" {
		return actor, nil
	}
	if e.Auth == nil {
		return ActorContext{}, domain.ErrUnauthenticated
	}

	actor := ActorContext{
		UserID:      e.Auth.Id,
		Email:       e.Auth.Email(),
		IsSuperuser: e.Auth.IsSuperuser(),
	}
	SetActorContext(e, actor)
	return actor, nil
}

func resolveAuthRecord(e *core.RequestEvent) (*core.Record, error) {
	if e.Auth != nil {
		return e.Auth, nil
	}

	token := strings.TrimSpace(readAuthTokenFromRequest(e))
	if token == "" {
		return nil, domain.ErrUnauthenticated
	}

	record, err := e.App.FindAuthRecordByToken(token)
	if err != nil {
		return nil, errors.Join(domain.ErrUnauthenticated, err)
	}

	return record, nil
}

func readAuthTokenFromRequest(e *core.RequestEvent) string {
	header := strings.TrimSpace(e.Request.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[len("Bearer "):])
	}
	cookie, err := e.Request.Cookie("pb_auth")
	if err == nil && cookie != nil {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}
