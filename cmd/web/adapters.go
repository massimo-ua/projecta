package main

import (
	"context"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/project"
	"gitlab.com/massimo-ua/projecta/internal/statement"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type currencyConverterAdapter struct {
	provider currency.CurrencyRateProvider
}

func (a *currencyConverterAdapter) ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error) {
	if amount == nil {
		return money.New(0, targetCurrency), nil
	}
	currCode := amount.Currency().Code
	if currCode == targetCurrency {
		return amount, nil
	}
	if a.provider == nil {
		return nil, exceptions.NewValidationException("currency rate provider not configured", nil)
	}
	converted, err := a.provider.Convert(
		currency.NewCurrency(amount.Amount(), currCode),
		currency.NewCurrency(0, targetCurrency),
	)
	if err != nil {
		return nil, err
	}
	return money.New(converted.Amount, converted.Code), nil
}

type personReader interface {
	FindByID(ctx context.Context, personID core.PersonID) (*people.Person, error)
}

type projectPeopleAdapter struct {
	repo personReader
}

func (a *projectPeopleAdapter) FindOwner(ctx context.Context, personID core.PersonID) (*project.Owner, error) {
	person, err := a.repo.FindByID(ctx, personID)
	if err != nil {
		return nil, err
	}
	return &project.Owner{
		PersonID:        person.ID(),
		DisplayName:     person.DisplayName(),
		FirstName:       person.FirstName(),
		CanHaveProjects: person.CanHaveProjects(),
	}, nil
}

type assetProjectReaderAdapter struct {
	projects project.ProjectRepository
}

func (a *assetProjectReaderAdapter) FindProject(ctx context.Context, projectID core.ProjectID) (asset.ProjectDescriptor, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return asset.ProjectDescriptor{}, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}
	proj, err := a.projects.FindOne(ctx, project.ProjectFilter{
		ProjectID:   projectID,
		RequesterID: personID,
	})
	if err != nil {
		return asset.ProjectDescriptor{}, err
	}
	return asset.ProjectDescriptor{
		ID:           proj.ID(),
		MainCurrency: proj.MainCurrency(),
	}, nil
}

type assetOwnerReaderAdapter struct {
	people project.PeopleService
}

func (a *assetOwnerReaderAdapter) FindOwner(ctx context.Context, personID core.PersonID) (asset.OwnerRef, error) {
	owner, err := a.people.FindOwner(ctx, personID)
	if err != nil {
		return asset.OwnerRef{}, err
	}
	return asset.OwnerRef{
		PersonID:    owner.PersonID,
		DisplayName: owner.DisplayName,
	}, nil
}

type investmentAssetReaderAdapter struct {
	assets asset.Repository
}

func (a *investmentAssetReaderAdapter) FindAsset(ctx context.Context, projectID core.ProjectID, assetID core.AssetID) (investment.AssetDescriptor, error) {
	ast, err := a.assets.FindOne(ctx, asset.Filter{ID: assetID, ProjectID: projectID})
	if err != nil {
		return investment.AssetDescriptor{}, err
	}
	return investment.AssetDescriptor{
		ID:   ast.ID(),
		Name: ast.Name(),
	}, nil
}

type investmentProjectReaderAdapter struct {
	projects project.ProjectRepository
}

func (a *investmentProjectReaderAdapter) FindProject(ctx context.Context, projectID core.ProjectID) (investment.ProjectDescriptor, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return investment.ProjectDescriptor{}, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}
	proj, err := a.projects.FindOne(ctx, project.ProjectFilter{
		ProjectID:   projectID,
		RequesterID: personID,
	})
	if err != nil {
		return investment.ProjectDescriptor{}, err
	}
	return investment.ProjectDescriptor{
		ID:           proj.ID(),
		MainCurrency: proj.MainCurrency(),
	}, nil
}

type investmentContributorReaderAdapter struct {
	people project.PeopleService
}

func (a *investmentContributorReaderAdapter) FindContributor(ctx context.Context, personID core.PersonID) (investment.ContributorRef, error) {
	owner, err := a.people.FindOwner(ctx, personID)
	if err != nil {
		return investment.ContributorRef{}, err
	}
	return investment.ContributorRef{
		PersonID:    owner.PersonID,
		DisplayName: owner.DisplayName,
	}, nil
}

type projectValidator struct {
	svc project.ProjectService
}

func (v *projectValidator) ValidateProject(ctx context.Context, projectID core.ProjectID) error {
	_, err := v.svc.FindOne(ctx, project.ProjectFilter{ProjectID: projectID})
	return err
}

type investmentTransactionFinder struct {
	svc investment.Service
}

func (f *investmentTransactionFinder) FindExisting(ctx context.Context, projectID core.ProjectID) ([]statement.ExistingTransaction, error) {
	pageSize := 500
	offset := 0
	var res []statement.ExistingTransaction

	for {
		col, err := f.svc.Find(ctx, investment.CollectionFilter{
			ProjectID: projectID,
			Pagination: core.Pagination{
				Limit:  pageSize,
				Offset: offset,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, inv := range col.Elements() {
			if inv == nil || inv.Amount() == nil {
				continue
			}
			res = append(res, statement.ExistingTransaction{
				Date:        inv.Date(),
				Amount:      inv.Amount().Amount(),
				Description: inv.Description(),
			})
		}
		if len(col.Elements()) < pageSize || len(res) >= col.Total() {
			break
		}
		offset += pageSize
	}
	return res, nil
}

type investmentSourceAdapter struct {
	svc investment.Service
}

func (a *investmentSourceAdapter) FindInvestments(ctx context.Context, projectID core.ProjectID, ids []core.InvestmentID) ([]asset.InvestmentItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	col, err := a.svc.Find(ctx, investment.CollectionFilter{
		ProjectID:  projectID,
		IDs:        ids,
		Pagination: core.Pagination{Limit: len(ids)},
	})
	if err != nil {
		return nil, err
	}
	items := make([]asset.InvestmentItem, 0, len(col.Elements()))
	for _, inv := range col.Elements() {
		items = append(items, asset.InvestmentItem{
			ID:     inv.ID(),
			Amount: inv.Amount(),
			Date:   inv.Date(),
			Tags:   inv.Tags(),
		})
	}
	return items, nil
}

func (a *investmentSourceAdapter) AssignToAsset(ctx context.Context, assetID core.AssetID, investmentIDs []core.InvestmentID) error {
	return a.svc.AssignToAsset(ctx, assetID, investmentIDs)
}

func (a *investmentSourceAdapter) CreateInitialInvestment(ctx context.Context, item asset.InitialInvestment) error {
	_, err := a.svc.Create(ctx, investment.CreateInvestmentCommand{
		ProjectID:    item.ProjectID,
		AssetID:      item.AssetID,
		ResourceType: investment.ResourceTypeMoney,
		Amount:       item.Amount,
		Description:  item.Description,
		Date:         item.Date,
		Tags:         item.Tags,
	})
	return err
}

type projectTotalsCalculatorAdapter struct {
	assets       asset.Service
	investments  investment.Service
	rateProvider currency.CurrencyRateProvider
}

func (a *projectTotalsCalculatorAdapter) CalculateTotals(ctx context.Context, projectID core.ProjectID, homeCurrency string) (project.ProjectTotals, error) {
	var invAmounts []*money.Money
	var rootAssets []*asset.Asset
	hasInvestments := false
	hasAssets := false

	const pageLimit = 100

	if a.investments != nil {
		for offset := 0; ; offset += pageLimit {
			page, err := a.investments.Find(ctx, investment.CollectionFilter{
				ProjectID: projectID,
				Pagination: core.Pagination{
					Limit:  pageLimit,
					Offset: offset,
				},
			})
			if err != nil {
				return project.ProjectTotals{}, err
			}
			if page.Total() == 0 {
				break
			}
			for _, e := range page.Elements() {
				hasInvestments = true
				invAmounts = append(invAmounts, e.Amount())
			}
			if len(page.Elements()) < pageLimit {
				break
			}
		}
	}

	if a.assets != nil {
		for offset := 0; ; offset += pageLimit {
			page, err := a.assets.Find(ctx, asset.CollectionFilter{
				ProjectID: projectID,
				Pagination: core.Pagination{
					Limit:  pageLimit,
					Offset: offset,
				},
			})
			if err != nil {
				return project.ProjectTotals{}, err
			}
			if page.Total() == 0 {
				break
			}
			for _, e := range page.Elements() {
				if len(e.Parents()) > 0 {
					continue
				}
				hasAssets = true
				rootAssets = append(rootAssets, e)
			}
			if len(page.Elements()) < pageLimit {
				break
			}
		}
	}

	var converter asset.CurrencyConverter
	if a.rateProvider != nil {
		converter = &currencyConverterAdapter{provider: a.rateProvider}
	}

	totals, err := asset.CalculateProjectTotals(invAmounts, rootAssets, homeCurrency, converter)
	if err != nil {
		return project.ProjectTotals{}, err
	}

	return project.ProjectTotals{
		TotalInvested: totals.TotalInvested,
		Balance:       totals.Balance(),
		HasActivity:   hasAssets || hasInvestments,
		Currency:      totals.Currency,
	}, nil
}
