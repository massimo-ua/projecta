package core

import (
    "context"
    "errors"
    "github.com/google/uuid"
)

type requesterIDContextKey string

var RequesterIDContextKey = requesterIDContextKey("requesterId")

var FailedToIdentifyRequester = errors.New("failed to identify requester")

func AuthGuard(ctx context.Context) (uuid.UUID, error) {
    personID, ok := ctx.Value(RequesterIDContextKey).(uuid.UUID)

    if !ok {
        return uuid.Nil, FailedToIdentifyRequester
    }

    return personID, nil
}
