package asset

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

const (
	failedToCreateAsset = "failed to create asset"
	failedToFindAsset   = "failed to find asset"
)

type ServiceImpl struct {
	assets   Repository
	people   projecta.PeopleService
	types    projecta.TypeRepository
	projects projecta.ProjectRepository
	payments projecta.PaymentRepository
}

func NewService(
	assets Repository,
	people projecta.PeopleService,
	types projecta.TypeRepository,
	projects projecta.ProjectRepository,
	payments projecta.PaymentRepository,
) *ServiceImpl {
	return &ServiceImpl{
		assets:   assets,
		people:   people,
		types:    types,
		projects: projects,
		payments: payments,
	}
}

func (s *ServiceImpl) Find(ctx context.Context, filter CollectionFilter) (*Collection, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindAsset, err)
	}

	filter.OwnerID = personID

	collection, err := s.assets.Find(ctx, filter)

	if err != nil {
		return nil, exceptions.NewInternalException(failedToFindAsset, err)
	}

	return collection, nil
}

func (s *ServiceImpl) FindOne(ctx context.Context, filter Filter) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindAsset, err)
	}

	filter.OwnerID = personID

	asset, err := s.assets.FindOne(ctx, filter)

	if err != nil {
		return nil, exceptions.NewInternalException(failedToFindAsset, err)
	}

	return asset, nil
}

func (s *ServiceImpl) Create(ctx context.Context, command CreateAssetCommand) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToCreateAsset, err)
	}

	owner, err := s.people.FindOwner(ctx, personID)

	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	project, err := s.projects.FindOne(ctx, projecta.ProjectFilter{ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	costType, err := s.types.FindOne(ctx, projecta.TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	acquiredAt := core.DateOrNow(command.AcquiredAt)

	asset := NewAsset(
		uuid.New(),
		command.Name,
		command.Description,
		project,
		costType,
		command.Price,
		acquiredAt,
		owner,
	)

	if command.WithPayment {
		payment := projecta.NewPayment(
			uuid.New(),
			project,
			owner,
			costType,
			command.Description,
			command.Price,
			acquiredAt,
			projecta.UponCompletionPayment,
		)

		txCtx, err := s.assets.TxCtx(ctx)

		defer s.assets.RollbackTxFromCtx(txCtx)

		if err != nil {
			return nil, exceptions.NewInternalException(failedToCreateAsset, err)
		}

		if err = s.payments.Save(txCtx, payment); err != nil {
			return nil, exceptions.NewInternalException(failedToCreateAsset, err)
		}

		if err = s.assets.Save(txCtx, asset); err != nil {
			return nil, exceptions.NewInternalException(failedToCreateAsset, err)
		}

		if err = s.assets.CommitTxFromCtx(txCtx); err != nil {
			return nil, exceptions.NewInternalException(failedToCreateAsset, err)
		}

		return asset, nil
	}

	err = s.assets.Save(ctx, asset)

	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	return asset, nil
}

func (s *ServiceImpl) Remove(ctx context.Context, command RemoveAssetCommand) error {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return exceptions.NewUnauthorizedException(failedToFindAsset, err)
	}

	asset, err := s.assets.FindOne(ctx, Filter{ID: command.AssetID, OwnerID: personID})

	if err != nil {
		return exceptions.NewInternalException(failedToFindAsset, err)
	}

	return s.assets.Remove(ctx, asset)
}
