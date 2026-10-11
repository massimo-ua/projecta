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

// ProjectReader is a consumer-owned interface providing read access to projects.
type ProjectReader interface {
	FindOne(ctx context.Context, filter projecta.ProjectFilter) (*projecta.Project, error)
}

type ServiceImpl struct {
	db           core.DbConnection
	assets       Repository
	people       projecta.PeopleService
	projects     ProjectReader
	rateProvider currency.CurrencyRateProvider
	investments  InvestmentSource
}

func (s *ServiceImpl) SetInvestmentSource(is InvestmentSource) {
	s.investments = is
}

func NewService(
	db core.DbConnection,
	assets Repository,
	people projecta.PeopleService,
	projects ProjectReader,
	investmentSource InvestmentSource,
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
		projects:     projects,
		investments:  investmentSource,
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

	anAsset := ReconstituteAsset(
		uuid.New(),
		command.Name,
		command.Description,
		project,
		price,
		startDate,
		owner,
		status,
		startDate,
		command.CompletedDate,
		price,
		command.Tags,
	)

	paymentDescription := command.Description
	if paymentDescription == "" {
		paymentDescription = command.Name
	}

	if command.WithPayment {
		_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
			if err = s.assets.Save(ctx, anAsset); err != nil {
				return nil, exceptions.NewInternalException(failedToCreateAsset, err)
			}

			if s.investments != nil {
				if err = s.investments.CreateInitialInvestment(ctx, InitialInvestment{
					ID:            uuid.New(),
					ProjectID:     command.ProjectID,
					AssetID:       anAsset.ID(),
					ContributorID: owner.PersonID,
					Amount:        price,
					Date:          startDate,
					Description:   paymentDescription,
					Tags:          command.Tags,
				}); err != nil {
					return nil, exceptions.NewInternalException(failedToCreateAsset, err)
				}
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

	if s.investments == nil {
		return nil, exceptions.NewInternalException("investment source not configured", nil)
	}

	invItems, err := s.investments.FindInvestments(ctx, command.ProjectID, command.PaymentIDs)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}
	if len(invItems) != len(command.PaymentIDs) {
		return nil, exceptions.NewValidationException("one or more investments not found in project", nil)
	}

	acquiredAt := command.AcquiredAt
	if acquiredAt.IsZero() {
		for _, inv := range invItems {
			if inv.Date.After(acquiredAt) {
				acquiredAt = inv.Date
			}
		}
		if acquiredAt.IsZero() {
			acquiredAt = time.Now()
		}
	}

	targetCurrency := strings.ToUpper(strings.TrimSpace(command.TargetCurrency))
	if targetCurrency == "" {
		firstCurr := invItems[0].Amount.Currency().Code
		allSameCurr := true
		for _, inv := range invItems[1:] {
			if inv.Amount.Currency().Code != firstCurr {
				allSameCurr = false
				break
			}
		}
		if allSameCurr {
			targetCurrency = firstCurr
		} else {
			targetCurrency = project.MainCurrency()
			if targetCurrency == "" {
				targetCurrency = "UAH"
			}
		}
	}

	var totalAmount int64
	tagSet := make(map[string]struct{})
	for _, inv := range invItems {
		currCode := inv.Amount.Currency().Code
		if currCode == targetCurrency {
			totalAmount += inv.Amount.Amount()
		} else {
			if s.rateProvider == nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("currency rate provider not available to convert %s to %s", currCode, targetCurrency), nil)
			}
			converted, err := s.rateProvider.Convert(
				currency.NewCurrency(inv.Amount.Amount(), currCode),
				currency.NewCurrency(0, targetCurrency),
			)
			if err != nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("failed to convert currency %s to %s: %s", currCode, targetCurrency, err.Error()), err)
			}
			totalAmount += converted.Amount
		}

		for _, t := range inv.Tags {
			if t != "" {
				tagSet[t] = struct{}{}
			}
		}
	}

	var tags []string
	for t := range tagSet {
		tags = append(tags, t)
	}

	if totalAmount <= 0 {
		return nil, exceptions.NewValidationException("total asset price must be greater than 0", nil)
	}

	price := money.New(totalAmount, targetCurrency)
	anAsset := ReconstituteAsset(
		uuid.New(),
		command.Name,
		command.Description,
		project,
		price,
		acquiredAt,
		owner,
		AssetStatusActive,
		acquiredAt,
		nil,
		price,
		tags,
	)
	anAsset.AssignValuation(price, price, nil, nil)

	if err = s.assets.Save(ctx, anAsset); err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	if err = s.investments.AssignToAsset(ctx, anAsset.ID(), command.PaymentIDs); err != nil {
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

	price := command.TargetPrice
	if price == nil {
		price = command.Price
	}

	anAsset.UpdateDetails(command.Name, command.Description, price, command.Tags)
	if command.CompletedDate != nil {
		anAsset.Complete(*command.CompletedDate)
	} else if command.Status == AssetStatusActive {
		anAsset.Reopen()
	} else if command.Status != "" {
		anAsset.SetStatus(command.Status)
	}

	if !command.StartDate.IsZero() {
		anAsset.SetStartDate(command.StartDate)
	} else if !command.AcquiredAt.IsZero() {
		anAsset.SetStartDate(command.AcquiredAt)
	}

	return s.assets.Save(ctx, anAsset)
}

func (s *ServiceImpl) LinkChild(ctx context.Context, command LinkChildCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException("failed to link child asset", err)
	}

	if command.ParentID == command.ChildID {
		return ErrSelfReference
	}

	if command.SharePercentage <= 0 || command.SharePercentage > 100 {
		return ErrInvalidShare
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
			return ErrCircularDependency
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

func (s *ServiceImpl) AssignInvestments(ctx context.Context, command AssignInvestmentsCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException("failed to assign investments to asset", err)
	}

	if command.AssetID == uuid.Nil {
		return exceptions.NewValidationException("asset id is required", nil)
	}
	if command.ProjectID == uuid.Nil {
		return exceptions.NewValidationException("project id is required", nil)
	}
	if len(command.InvestmentIDs) == 0 {
		return exceptions.NewValidationException("at least one investment must be selected", nil)
	}

	_, err = s.assets.FindOne(ctx, Filter{
		ID:        command.AssetID,
		ProjectID: command.ProjectID,
		OwnerID:   personID,
	})
	if err != nil {
		return exceptions.NewValidationException("asset not found in project", err)
	}

	if s.investments != nil {
		items, err := s.investments.FindInvestments(ctx, command.ProjectID, command.InvestmentIDs)
		if err != nil || len(items) != len(command.InvestmentIDs) {
			return exceptions.NewValidationException("one or more investments not found in project", err)
		}

		if err = s.investments.AssignToAsset(ctx, command.AssetID, command.InvestmentIDs); err != nil {
			return exceptions.NewInternalException("failed to assign investments to asset", err)
		}
	}

	return nil
}

func (s *ServiceImpl) Group(ctx context.Context, command GroupAssetsCommand) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to group assets", err)
	}

	if command.ProjectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	if strings.TrimSpace(command.Name) == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	if len(command.ChildAssetIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one asset must be selected to group", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to group assets", err)
	}

	project, err := s.projects.FindOne(ctx, projecta.ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return nil, exceptions.NewInternalException("failed to group assets", err)
	}

	// Verify all child assets exist and belong to the project
	for _, childID := range command.ChildAssetIDs {
		child, err := s.assets.FindOne(ctx, Filter{ID: childID, OwnerID: personID})
		if err != nil || child == nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("child asset %s not found in project", childID), err)
		}
	}

	now := time.Now()
	parentID := uuid.New()
	tags := command.Tags
	if tags == nil {
		tags = []string{}
	}
	parentAsset := ReconstituteAsset(
		parentID,
		command.Name,
		command.Description,
		project,
		money.New(0, "UAH"),
		now,
		owner,
		AssetStatusActive,
		now,
		nil,
		money.New(0, "UAH"),
		tags,
	)

	_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
		if err := s.assets.Save(ctx, parentAsset); err != nil {
			return nil, exceptions.NewInternalException("failed to group assets", err)
		}

		for _, childID := range command.ChildAssetIDs {
			if err := s.assets.AddChild(ctx, parentID, childID, 100.0); err != nil {
				return nil, exceptions.NewInternalException("failed to link child asset", err)
			}
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return s.assets.FindOne(ctx, Filter{ID: parentID, OwnerID: personID})
}

func (s *ServiceImpl) LinkChildren(ctx context.Context, command LinkChildrenCommand) error {
	for _, childID := range command.ChildAssetIDs {
		if err := s.LinkChild(ctx, LinkChildCommand{
			ParentID:        command.ParentID,
			ChildID:         childID,
			ProjectID:       command.ProjectID,
			SharePercentage: 100.0,
		}); err != nil {
			return err
		}
	}
	return nil
}


