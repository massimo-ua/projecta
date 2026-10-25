package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

var ErrFailedToSaveInvestment = errors.New("failed to save investment")
var ErrInvestmentNotFound = errors.Join(exceptions.NotFoundError, errors.New("investment not found"))

type PgInvestmentRepository struct {
	db *PgRepository
}

func NewPgInvestmentRepository(conn *PgDbConnection) *PgInvestmentRepository {
	return &PgInvestmentRepository{
		db: &PgRepository{db: conn},
	}
}

func (r *PgInvestmentRepository) Save(ctx context.Context, inv *investment.Investment) error {
	var hourlyRateVal any
	if inv.TimeHourlyRate() != nil {
		hourlyRateVal = inv.TimeHourlyRate().Amount()
	}

	amountVal := int64(0)
	currVal := "UAH"
	if inv.Amount() != nil {
		amountVal = inv.Amount().Amount()
		if inv.Amount().Currency() != nil {
			currVal = inv.Amount().Currency().Code
		}
	}

	tags := inv.Tags()
	if tags == nil {
		tags = []string{}
	}

	sql := `
		INSERT INTO projecta_investments (
			investment_id, project_id, asset_id, contributor_id,
			resource_type, amount, currency, time_hours,
			time_hourly_rate, goods_quantity, goods_unit,
			goods_item_name, description, date, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (investment_id) DO UPDATE SET
			asset_id = EXCLUDED.asset_id,
			resource_type = EXCLUDED.resource_type,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			time_hours = EXCLUDED.time_hours,
			time_hourly_rate = EXCLUDED.time_hourly_rate,
			goods_quantity = EXCLUDED.goods_quantity,
			goods_unit = EXCLUDED.goods_unit,
			goods_item_name = EXCLUDED.goods_item_name,
			description = EXCLUDED.description,
			date = EXCLUDED.date,
			tags = EXCLUDED.tags
	`

	if _, err := r.db.Exec(
		ctx,
		sql,
		inv.ID().String(),
		inv.ProjectID().String(),
		inv.PrimaryAssetID().String(),
		inv.Contributor().PersonID.String(),
		inv.ResourceType().String(),
		amountVal,
		currVal,
		inv.TimeHours(),
		hourlyRateVal,
		inv.GoodsQuantity(),
		inv.GoodsUnit(),
		inv.GoodsItemName(),
		inv.Description(),
		inv.Date(),
		tags,
	); err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	if err := r.saveAllocations(ctx, inv); err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	return nil
}

func (r *PgInvestmentRepository) saveAllocations(ctx context.Context, inv *investment.Investment) error {
	delSql := "DELETE FROM projecta_investment_assets WHERE investment_id = $1"
	if _, err := r.db.Exec(ctx, delSql, inv.ID().String()); err != nil {
		return err
	}

	allocations := inv.AssetAllocations()
	if len(allocations) == 0 && !inv.PrimaryAssetID().IsNil() {
		allocations = []investment.InvestmentAssetLink{
			{
				AssetID:         inv.PrimaryAssetID(),
				AssetName:       inv.PrimaryAssetName(),
				SharePercentage: 100.0,
			},
		}
	}

	for _, a := range allocations {
		insertSql := `
			INSERT INTO projecta_investment_assets (investment_id, asset_id, share_percentage)
			VALUES ($1, $2, $3)
			ON CONFLICT (investment_id, asset_id) DO UPDATE SET share_percentage = EXCLUDED.share_percentage
		`
		if _, err := r.db.Exec(ctx, insertSql, inv.ID().String(), a.AssetID.String(), a.SharePercentage); err != nil {
			return err
		}
	}
	return nil
}

func (r *PgInvestmentRepository) loadAllocations(ctx context.Context, investmentIDs []string) (map[string][]investment.InvestmentAssetLink, error) {
	result := make(map[string][]investment.InvestmentAssetLink)
	if len(investmentIDs) == 0 {
		return result, nil
	}

	sql := `
		SELECT ia.investment_id, ia.asset_id, a.name, ia.share_percentage
		FROM projecta_investment_assets ia
		JOIN projecta_assets a ON a.asset_id = ia.asset_id
		WHERE ia.investment_id = ANY($1::uuid[])
		ORDER BY ia.share_percentage DESC
	`
	rows, err := r.db.Query(ctx, sql, investmentIDs)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var invID, astID, astName string
		var share float64
		if err := rows.Scan(&invID, &astID, &astName, &share); err == nil {
			result[invID] = append(result[invID], investment.InvestmentAssetLink{
				AssetID:         core.MustParseAssetID(astID),
				AssetName:       astName,
				SharePercentage: share,
			})
		}
	}
	return result, nil
}

func (r *PgInvestmentRepository) Remove(ctx context.Context, inv *investment.Investment) error {
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_investments")
	qb.Where(qb.Equal("investment_id", inv.ID().String()))
	qb.Where(qb.Equal("project_id", inv.ProjectID().String()))

	sql, args := qb.Build()
	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrInvestmentNotFound
	}

	return nil
}

func (r *PgInvestmentRepository) FindOne(ctx context.Context, filter investment.Filter) (*investment.Investment, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_investments")
	setupInvestmentSelect(qb)

	qb.Where(qb.Equal("projecta_investments.investment_id", filter.ID.String()))
	if !filter.ProjectID.IsNil() {
		qb.Where(qb.Equal("projecta_investments.project_id", filter.ProjectID.String()))
	}
	if !filter.AssetID.IsNil() {
		qb.Where(fmt.Sprintf("(projecta_investments.asset_id = %s OR projecta_investments.investment_id IN (SELECT investment_id FROM projecta_investment_assets WHERE asset_id = %s))", qb.Var(filter.AssetID.String()), qb.Var(filter.AssetID.String())))
	}

	sql, args := qb.Build()
	row := r.db.QueryRow(ctx, sql, args...)

	rec, err := scanInvestmentRecord(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvestmentNotFound
		}
		return nil, err
	}

	allocsMap, err := r.loadAllocations(ctx, []string{rec.id.String()})
	if err != nil {
		return nil, err
	}
	var allocs []investment.InvestmentAssetLink
	if allocsMap != nil {
		allocs = allocsMap[rec.id.String()]
	}

	inv := investment.ReconstituteInvestment(
		rec.id,
		rec.projectID,
		rec.homeCurrency,
		rec.primaryAssetID,
		rec.primaryAssetName,
		allocs,
		rec.contributor,
		rec.resourceType,
		rec.amount,
		rec.timeHours,
		rec.timeHourlyRate,
		rec.goodsQuantity,
		rec.goodsUnit,
		rec.goodsItemName,
		rec.description,
		rec.date,
		rec.tags,
	)

	return inv, nil
}

func (r *PgInvestmentRepository) Find(ctx context.Context, filter investment.CollectionFilter) (*investment.Collection, error) {
	countQb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countQb.From("projecta_investments")
	applyInvestmentFilters(countQb, filter)
	countQb.Select(countQb.As("COUNT(*)", "total"))

	countSql, countArgs := countQb.Build()

	var total int
	if err := r.db.QueryRow(ctx, countSql, countArgs...).Scan(&total); err != nil {
		return nil, err
	}

	if total == 0 {
		return investment.NewCollection(0), nil
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_investments")
	setupInvestmentSelect(qb)
	applyInvestmentFilters(qb, filter)

	qb.Offset(filter.Offset)
	qb.Limit(filter.Limit)

	validInvestmentOrderBy := map[string]bool{
		"date":          true,
		"amount":        true,
		"resource_type": true,
		"description":   true,
	}
	if filter.OrderBy != "" && filter.Order != "" && validInvestmentOrderBy[filter.OrderBy] {
		qb.OrderBy(fmt.Sprintf("projecta_investments.%s %s", filter.OrderBy, filter.Order.String()))
	} else {
		qb.OrderBy("projecta_investments.date DESC")
	}

	sql, args := qb.Build()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*investmentRecord, 0, total)
	invIDs := make([]string, 0, total)
	for rows.Next() {
		rec, err := scanInvestmentRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
		invIDs = append(invIDs, rec.id.String())
	}

	var allocsMap map[string][]investment.InvestmentAssetLink
	if len(invIDs) > 0 {
		var err error
		allocsMap, err = r.loadAllocations(ctx, invIDs)
		if err != nil {
			return nil, err
		}
	}

	col := investment.NewCollection(total)
	for _, rec := range records {
		var allocs []investment.InvestmentAssetLink
		if allocsMap != nil {
			allocs = allocsMap[rec.id.String()]
		}
		inv := investment.ReconstituteInvestment(
			rec.id,
			rec.projectID,
			rec.homeCurrency,
			rec.primaryAssetID,
			rec.primaryAssetName,
			allocs,
			rec.contributor,
			rec.resourceType,
			rec.amount,
			rec.timeHours,
			rec.timeHourlyRate,
			rec.goodsQuantity,
			rec.goodsUnit,
			rec.goodsItemName,
			rec.description,
			rec.date,
			rec.tags,
		)
		col.Add(inv)
	}

	return col, nil
}

func (r *PgInvestmentRepository) FindTags(ctx context.Context, projectID core.ProjectID) ([]string, error) {
	sql := `
		SELECT DISTINCT tag FROM (
			SELECT unnest(tags) AS tag 
			FROM projecta_investments 
			WHERE project_id = $1 
			UNION 
			SELECT unnest(tags) AS tag 
			FROM projecta_assets 
			WHERE project_id = $1
			UNION
			SELECT tag AS tag
			FROM projecta_project_tag_budgets
			WHERE project_id = $1
		) t 
		WHERE tag <> '' 
		ORDER BY tag ASC
	`
	rows, err := r.db.Query(ctx, sql, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err == nil && tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags, nil
}

func setupInvestmentSelect(qb *sqlbuilder.SelectBuilder) {
	qb.Select(
		"projecta_investments.investment_id",
		"projecta_investments.project_id",
		"projecta_investments.asset_id",
		"projecta_assets.name",
		"projecta_investments.contributor_id",
		qb.As("COALESCE(people.display_name, people.first_name, '')", "display_name"),
		"projecta_investments.resource_type",
		"projecta_investments.amount",
		"projecta_investments.currency",
		"projecta_investments.time_hours",
		"projecta_investments.time_hourly_rate",
		"projecta_investments.goods_quantity",
		"projecta_investments.goods_unit",
		"projecta_investments.goods_item_name",
		"projecta_investments.description",
		"projecta_investments.date",
		"projecta_investments.tags",
	)

	qb.Join("projecta_assets", "projecta_assets.asset_id = projecta_investments.asset_id")
	qb.Join("people", "people.person_id = projecta_investments.contributor_id")
}

func applyInvestmentFilters(qb *sqlbuilder.SelectBuilder, filter investment.CollectionFilter) {
	if len(filter.IDs) > 0 {
		vals := make([]interface{}, len(filter.IDs))
		for i, id := range filter.IDs {
			vals[i] = id.String()
		}
		qb.Where(qb.In("projecta_investments.investment_id", vals...))
	}
	if !filter.ProjectID.IsNil() {
		qb.Where(qb.Equal("projecta_investments.project_id", filter.ProjectID.String()))
	}
	if !filter.AssetID.IsNil() {
		qb.Where(fmt.Sprintf("(projecta_investments.asset_id = %s OR projecta_investments.investment_id IN (SELECT investment_id FROM projecta_investment_assets WHERE asset_id = %s))", qb.Var(filter.AssetID.String()), qb.Var(filter.AssetID.String())))
	}
	if !filter.OwnerID.IsNil() {
		qb.Where(qb.Equal("projecta_investments.contributor_id", filter.OwnerID.String()))
	}
	if filter.ResourceType != "" {
		qb.Where(qb.Equal("projecta_investments.resource_type", filter.ResourceType.String()))
	}
	if filter.Tag != "" {
		qb.Where(fmt.Sprintf("%s = ANY(projecta_investments.tags)", qb.Var(filter.Tag)))
	}
	if filter.StartDate != nil {
		qb.Where(qb.GreaterEqualThan("projecta_investments.date", *filter.StartDate))
	}
	if filter.EndDate != nil {
		qb.Where(qb.LessEqualThan("projecta_investments.date", *filter.EndDate))
	}
}

type scannable interface {
	Scan(dest ...any) error
}

type investmentRecord struct {
	id               core.InvestmentID
	projectID        core.ProjectID
	homeCurrency     string
	primaryAssetID   core.AssetID
	primaryAssetName string
	contributor      investment.ContributorRef
	resourceType     investment.ResourceType
	amount           *money.Money
	timeHours        *float64
	timeHourlyRate   *money.Money
	goodsQuantity    *float64
	goodsUnit        string
	goodsItemName    string
	description      string
	date             time.Time
	tags             []string
}

func scanInvestmentRecord(s scannable) (*investmentRecord, error) {
	var (
		invID         string
		projID        string
		astID         string
		astName       string
		contribID     string
		contribDisp   string
		resType       string
		amount        int64
		currencyCode  string
		timeHours     *float64
		hourlyRate    *int64
		goodsQty      *float64
		goodsUnit     *string
		goodsItemName *string
		description   string
		date          time.Time
		tags          []string
	)

	err := s.Scan(
		&invID,
		&projID,
		&astID,
		&astName,
		&contribID,
		&contribDisp,
		&resType,
		&amount,
		&currencyCode,
		&timeHours,
		&hourlyRate,
		&goodsQty,
		&goodsUnit,
		&goodsItemName,
		&description,
		&date,
		&tags,
	)
	if err != nil {
		return nil, err
	}

	contrib := investment.ContributorRef{
		PersonID:    core.MustParsePersonID(contribID),
		DisplayName: contribDisp,
	}

	var hourlyRateMoney *money.Money
	if hourlyRate != nil {
		hourlyRateMoney = money.New(*hourlyRate, currencyCode)
	}

	var gUnit string
	if goodsUnit != nil {
		gUnit = *goodsUnit
	}
	var gItem string
	if goodsItemName != nil {
		gItem = *goodsItemName
	}

	rType, _ := investment.ToResourceType(resType)

	var parsedAstID core.AssetID
	if astID != "" {
		parsedAstID = core.MustParseAssetID(astID)
	}

	return &investmentRecord{
		id:               core.MustParseInvestmentID(invID),
		projectID:        core.MustParseProjectID(projID),
		homeCurrency:     currencyCode,
		primaryAssetID:   parsedAstID,
		primaryAssetName: astName,
		contributor:      contrib,
		resourceType:     rType,
		amount:           money.New(amount, currencyCode),
		timeHours:        timeHours,
		timeHourlyRate:   hourlyRateMoney,
		goodsQuantity:    goodsQty,
		goodsUnit:        gUnit,
		goodsItemName:    gItem,
		description:      description,
		date:             date,
		tags:             tags,
	}, nil
}
