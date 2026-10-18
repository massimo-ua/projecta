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

func AuthGuard(ctx context.Context) (PersonID, error) {
    if personID, ok := ctx.Value(RequesterIDContextKey).(PersonID); ok {
        return personID, nil
    }
    if u, ok := ctx.Value(RequesterIDContextKey).(uuid.UUID); ok {
        return PersonID(u), nil
    }

    return NilPersonID, FailedToIdentifyRequester
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

