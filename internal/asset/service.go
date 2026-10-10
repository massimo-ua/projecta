package asset

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

const (
	failedToCreateAsset = "failed to create asset"
	failedToFindAsset   = "failed to find asset"
	failedToUpdateAsset = "failed to update asset"
)

type ServiceImpl struct {
	db           core.DbConnection
	assets       Repository
	people       projecta.PeopleService
	types        projecta.TypeRepository
	projects     projecta.ProjectRepository
	payments     projecta.PaymentRepository
	rateProvider currency.CurrencyRateProvider
}

func NewService(
	db core.DbConnection,
	assets Repository,
	people projecta.PeopleService,
	types projecta.TypeRepository,
	projects projecta.ProjectRepository,
	payments projecta.PaymentRepository,
	rateProvider ...currency.CurrencyRateProvider,
) *ServiceImpl {
	var rp currency.CurrencyRateProvider
	if len(rateProvider) > 0 {
		rp = rateProvider[0]
	}
	return &ServiceImpl{
		db:           db,
		assets:       assets,
		people:       people,
		types:        types,
		projects:     projects,
		payments:     payments,
		rateProvider: rp,
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

	anAsset, err := s.assets.FindOne(ctx, filter)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToFindAsset, err)
	}

	return anAsset, nil
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

	var costType *projecta.CostType
	if s.types != nil {
		var err error
		costType, err = s.types.FindOne(ctx, projecta.TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})
		if err != nil {
			return nil, exceptions.NewInternalException(failedToCreateAsset, err)
		}
	}

	price := command.TargetPrice
	if price == nil {
		price = command.Price
	}

	startDate := command.StartDate
	if startDate.IsZero() {
		startDate = core.DateOrNow(command.AcquiredAt)
	}

	status := command.Status
	if status == "" {
		status = AssetStatusActive
	}

	anAsset := NewAsset(
		uuid.New(),
		command.Name,
		command.Description,
		project,
		costType,
		price,
		startDate,
		owner,
	)
	anAsset.SetStatus(status)
	anAsset.SetStartDate(startDate)
	anAsset.SetCompletedDate(command.CompletedDate)
	anAsset.SetTargetPrice(price)
	anAsset.SetTags(command.Tags)

	paymentDescription := command.Description
	if paymentDescription == "" {
		paymentDescription = command.Name
	}

	if command.WithPayment && s.payments != nil {
		payment := projecta.NewPayment(
			uuid.New(),
			project,
			owner,
			costType,
			paymentDescription,
			price,
			startDate,
			projecta.UponCompletionPayment,
		)

		_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
			if err = s.payments.Save(ctx, payment); err != nil {
				return nil, exceptions.NewInternalException(failedToCreateAsset, err)
			}

			if err = s.assets.Save(ctx, anAsset); err != nil {
				return nil, exceptions.NewInternalException(failedToCreateAsset, err)
			}

			return nil, nil
		})

		if err != nil {
			return nil, err
		}

		return anAsset, nil
	}

	err = s.assets.Save(ctx, anAsset)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	return anAsset, nil
}

func (s *ServiceImpl) CreateFromPayments(ctx context.Context, command CreateAssetFromPaymentsCommand) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToCreateAsset, err)
	}

	if command.ProjectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	if len(command.PaymentIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one payment must be selected", nil)
	}

	if strings.TrimSpace(command.Name) == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	project, err := s.projects.FindOne(ctx, projecta.ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	var payments []*projecta.Payment
	for _, pid := range command.PaymentIDs {
		payment, err := s.payments.FindOne(ctx, projecta.PaymentFilter{
			PaymentID: pid,
			ProjectID: command.ProjectID,
		})
		if err != nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("payment %s not found in project", pid), err)
		}
		payments = append(payments, payment)
	}

	var costType *projecta.CostType
	if command.TypeID != uuid.Nil {
		costType, err = s.types.FindOne(ctx, projecta.TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})
		if err != nil {
			return nil, exceptions.NewValidationException("invalid cost type", err)
		}
	} else {
		firstTypeID := payments[0].Type.ID
		allSame := true
		for _, p := range payments[1:] {
			if p.Type == nil || p.Type.ID != firstTypeID {
				allSame = false
				break
			}
		}
		if !allSame {
			return nil, exceptions.NewValidationException("cost type is required when selected payments have different types", nil)
		}
		costType = payments[0].Type
	}

	acquiredAt := command.AcquiredAt
	if acquiredAt.IsZero() {
		for _, p := range payments {
			if p.Date.After(acquiredAt) {
				acquiredAt = p.Date
			}
		}
		if acquiredAt.IsZero() {
			acquiredAt = time.Now()
		}
	}

	targetCurrency := strings.ToUpper(strings.TrimSpace(command.TargetCurrency))
	if targetCurrency == "" {
		firstCurr := payments[0].Amount.Currency().Code
		allSameCurr := true
		for _, p := range payments[1:] {
			if p.Amount.Currency().Code != firstCurr {
				allSameCurr = false
				break
			}
		}
		if allSameCurr {
			targetCurrency = firstCurr
		} else {
			targetCurrency = project.MainCurrency
			if targetCurrency == "" {
				targetCurrency = "UAH"
			}
		}
	}

	var totalAmount int64
	for _, p := range payments {
		currCode := p.Amount.Currency().Code
		if currCode == targetCurrency {
			totalAmount += p.Amount.Amount()
		} else {
			if s.rateProvider == nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("currency rate provider not available to convert %s to %s", currCode, targetCurrency), nil)
			}
			converted, err := s.rateProvider.Convert(
				currency.NewCurrency(p.Amount.Amount(), currCode),
				currency.NewCurrency(0, targetCurrency),
			)
			if err != nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("failed to convert currency %s to %s: %s", currCode, targetCurrency, err.Error()), err)
			}
			totalAmount += converted.Amount
		}
	}

	if totalAmount <= 0 {
		return nil, exceptions.NewValidationException("total asset price must be greater than 0", nil)
	}

	price := money.New(totalAmount, targetCurrency)

	anAsset := NewAsset(
		uuid.New(),
		command.Name,
		command.Description,
		project,
		costType,
		price,
		acquiredAt,
		owner,
	)
	anAsset.SetTargetPrice(price)

	if err = s.assets.Save(ctx, anAsset); err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	return anAsset, nil
}

func (s *ServiceImpl) Remove(ctx context.Context, command RemoveAssetCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException(failedToFindAsset, err)
	}

	anAsset, err := s.assets.FindOne(ctx, Filter{ID: command.AssetID, OwnerID: personID})
	if err != nil {
		return exceptions.NewInternalException(failedToFindAsset, err)
	}

	return s.assets.Remove(ctx, anAsset)
}

func (s *ServiceImpl) Update(ctx context.Context, command UpdateAssetCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException(failedToUpdateAsset, err)
	}

	_, err = s.projects.FindOne(ctx, projecta.ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return exceptions.NewInternalException(failedToUpdateAsset, err)
	}

	anAsset, err := s.assets.FindOne(ctx, Filter{ID: command.AssetID, OwnerID: personID})
	if err != nil {
		return exceptions.NewInternalException(failedToUpdateAsset, err)
	}

	if command.TypeID != uuid.Nil && s.types != nil {
		costType, err := s.types.FindOne(ctx, projecta.TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})
		if err != nil {
			return exceptions.NewInternalException(failedToUpdateAsset, err)
		}
		anAsset.SetType(costType)
	}

	anAsset.SetName(command.Name)
	anAsset.SetDescription(command.Description)

	if command.TargetPrice != nil {
		anAsset.SetTargetPrice(command.TargetPrice)
	} else if command.Price != nil {
		anAsset.SetTargetPrice(command.Price)
	}

	if command.Status != "" {
		anAsset.SetStatus(command.Status)
	}

	if !command.StartDate.IsZero() {
		anAsset.SetStartDate(command.StartDate)
	} else if !command.AcquiredAt.IsZero() {
		anAsset.SetStartDate(command.AcquiredAt)
	}

	if command.CompletedDate != nil {
		anAsset.SetCompletedDate(command.CompletedDate)
	}

	if command.Tags != nil {
		anAsset.SetTags(command.Tags)
	}

	return s.assets.Save(ctx, anAsset)
}

func (s *ServiceImpl) LinkChild(ctx context.Context, command LinkChildCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException("failed to link child asset", err)
	}

	if command.ParentID == command.ChildID {
		return exceptions.NewValidationException("an asset cannot be a child of itself", nil)
	}

	if command.SharePercentage <= 0 || command.SharePercentage > 100 {
		return exceptions.NewValidationException("share percentage must be between 0 and 100", nil)
	}

	_, err = s.assets.FindOne(ctx, Filter{ID: command.ParentID, OwnerID: personID})
	if err != nil {
		return exceptions.NewValidationException("parent asset not found", err)
	}

	_, err = s.assets.FindOne(ctx, Filter{ID: command.ChildID, OwnerID: personID})
	if err != nil {
		return exceptions.NewValidationException("child asset not found", err)
	}

	ancestors, err := s.assets.FindAncestors(ctx, command.ParentID)
	if err != nil {
		return exceptions.NewInternalException("failed to check hierarchy cycles", err)
	}
	for _, aID := range ancestors {
		if aID == command.ChildID {
			return exceptions.NewValidationException("cannot link child asset: circular dependency detected", nil)
		}
	}

	return s.assets.AddChild(ctx, command.ParentID, command.ChildID, command.SharePercentage)
}

func (s *ServiceImpl) UnlinkChild(ctx context.Context, command UnlinkChildCommand) error {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException("failed to unlink child asset", err)
	}

	return s.assets.RemoveChild(ctx, command.ParentID, command.ChildID)
}
