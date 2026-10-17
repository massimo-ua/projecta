package main

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/project"
	"gitlab.com/massimo-ua/projecta/internal/statement"
)

type projectValidator struct {
	svc project.ProjectService
}

func (v *projectValidator) ValidateProject(ctx context.Context, projectID uuid.UUID) error {
	_, err := v.svc.FindOne(ctx, project.ProjectFilter{ProjectID: projectID})
	return err
}

type investmentTransactionFinder struct {
	svc investment.Service
}

func (f *investmentTransactionFinder) FindExisting(ctx context.Context, projectID uuid.UUID) ([]statement.ExistingTransaction, error) {
	col, err := f.svc.Find(ctx, investment.CollectionFilter{
		ProjectID:  projectID,
		Pagination: core.Pagination{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	res := make([]statement.ExistingTransaction, 0, len(col.Elements()))
	for _, inv := range col.Elements() {
		res = append(res, statement.ExistingTransaction{
			Date:        inv.Date(),
			Amount:      inv.Amount().Amount(),
			Description: inv.Description(),
		})
	}
	return res, nil
}

type investmentSourceAdapter struct {
	svc investment.Service
}

func (a *investmentSourceAdapter) FindInvestments(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) ([]asset.InvestmentItem, error) {
	items := make([]asset.InvestmentItem, 0, len(ids))
	for _, id := range ids {
		inv, err := a.svc.FindOne(ctx, investment.Filter{ID: id, ProjectID: projectID})
		if err != nil {
			continue
		}
		items = append(items, asset.InvestmentItem{
			ID:     inv.ID(),
			Amount: inv.Amount(),
			Date:   inv.Date(),
			Tags:   inv.Tags(),
		})
	}
	return items, nil
}

func (a *investmentSourceAdapter) AssignToAsset(ctx context.Context, assetID uuid.UUID, investmentIDs []uuid.UUID) error {
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
