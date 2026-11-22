package app

import (
	"context"
	"errors"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

// OwnerReader is a narrow consumer-owned port for resolving project owners from the people context.
type OwnerReader interface {
	FindOwner(ctx context.Context, personID core.PersonID) (*project.Owner, error)
}

// TotalsReadModel is a narrow read-model port for computing aggregated project financial totals.
type TotalsReadModel interface {
	CalculateTotals(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID, homeCurrency string) (project.ProjectTotals, error)
}

func preserveOrWrapError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}
	var exc exceptions.Exception
	if errors.As(err, &exc) {
		return err
	}
	if errors.Is(err, exceptions.NotFoundError) {
		return exceptions.NewNotFoundException(fallbackMsg, err)
	}
	return exceptions.NewInternalException(fallbackMsg, err)
}

func logExecution(logger core.Logger, kind, name string, start time.Time, err error) {
	if logger == nil {
		return
	}
	fields := map[string]any{
		"kind":     kind,
		"handler":  name,
		"duration": time.Since(start).String(),
	}
	if err != nil {
		logger.Error(kind+" execution failed", err, fields)
		return
	}
	logger.Info(kind+" executed", fields)
}
