package core

import (
    "context"
    "errors"
    "github.com/google/uuid"
)

type requesterIDContextKey string
type requesterRolesContextKey string

var RequesterIDContextKey = requesterIDContextKey("requesterId")
var RequesterRolesContextKey = requesterRolesContextKey("requesterRoles")

var FailedToIdentifyRequester = errors.New("failed to identify requester")

func AuthGuard(ctx context.Context) (uuid.UUID, error) {
    personID, ok := ctx.Value(RequesterIDContextKey).(uuid.UUID)

    if !ok {
        return uuid.Nil, FailedToIdentifyRequester
    }

    return personID, nil
}

func RequesterRoles(ctx context.Context) []string {
    if roles, ok := ctx.Value(RequesterRolesContextKey).([]string); ok {
        return roles
    }
    return nil
}

func HasRole(ctx context.Context, role string) bool {
    for _, r := range RequesterRoles(ctx) {
        if r == role {
            return true
        }
    }
    return false
}

