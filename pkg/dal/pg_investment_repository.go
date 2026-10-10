package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

var ErrFailedToSaveInvestment = errors.New("failed to save investment")
var ErrInvestmentNotFound = errors.New("investment not found")

type PgInvestmentRepository struct {
	db *PgRepository
}

func NewPgInvestmentRepository(conn *PgDbConnection) *PgInvestmentRepository {
	return &PgInvestmentRepository{
		db: &PgRepository{db: conn},
	}
}

func (r *PgInvestmentRepository) Save(ctx context.Context, inv *investment.Investment) error {
	_, err := r.FindOne(ctx, investment.Filter{
		ID:        inv.ID,
		ProjectID: inv.Project.ProjectID,
	})

	if err != nil {
		if errors.Is(err, ErrInvestmentNotFound) {
			return r.create(ctx, inv)
		}
		return err
	}

	return r.update(ctx, inv)
}

func (r *PgInvestmentRepository) create(ctx context.Context, inv *investment.Investment) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_investments")
	qb.Cols(
		"investment_id",
		"project_id",
		"asset_id",
		"contributor_id",
		"resource_type",
		"amount",
		"currency",
		"time_hours",
		"time_hourly_rate",
		"goods_quantity",
		"goods_unit",
		"goods_item_name",
		"description",
		"date",
		"tags",
	)

	var hourlyRateVal any
	if inv.TimeHourlyRate != nil {
		hourlyRateVal = inv.TimeHourlyRate.Amount()
	}

	qb.Values(
		inv.ID.String(),
		inv.Project.ProjectID.String(),
		inv.Asset.ID().String(),
		inv.Contributor.PersonID.String(),
		inv.ResourceType.String(),
		inv.Amount.Amount(),
		inv.Amount.Currency().Code,
		inv.TimeHours,
		hourlyRateVal,
		inv.GoodsQuantity,
		inv.GoodsUnit,
		inv.GoodsItemName,
		inv.Description,
		inv.Date,
		inv.Tags,
	)

	sql, args := qb.Build()
	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	if err := r.saveAllocations(ctx, inv); err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	return nil
}

func (r *PgInvestmentRepository) update(ctx context.Context, inv *investment.Investment) error {
	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	qb.Update("projecta_investments")

	var hourlyRateVal any
	if inv.TimeHourlyRate != nil {
		hourlyRateVal = inv.TimeHourlyRate.Amount()
	}

	qb.Set(
		qb.Assign("asset_id", inv.Asset.ID().String()),
		qb.Assign("resource_type", inv.ResourceType.String()),
		qb.Assign("amount", inv.Amount.Amount()),
		qb.Assign("currency", inv.Amount.Currency().Code),
		qb.Assign("time_hours", inv.TimeHours),
		qb.Assign("time_hourly_rate", hourlyRateVal),
		qb.Assign("goods_quantity", inv.GoodsQuantity),
		qb.Assign("goods_unit", inv.GoodsUnit),
		qb.Assign("goods_item_name", inv.GoodsItemName),
		qb.Assign("description", inv.Description),
		qb.Assign("date", inv.Date),
		qb.Assign("tags", inv.Tags),
	)
	qb.Where(qb.Equal("investment_id", inv.ID.String()))
	qb.Where(qb.Equal("project_id", inv.Project.ProjectID.String()))

	sql, args := qb.Build()
	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	if err := r.saveAllocations(ctx, inv); err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	return nil
}

func (r *PgInvestmentRepository) saveAllocations(ctx context.Context, inv *investment.Investment) error {
	delSql := "DELETE FROM projecta_investment_assets WHERE investment_id = $1"
	if _, err := r.db.Exec(ctx, delSql, inv.ID.String()); err != nil {
		return err
	}

	allocations := inv.AssetAllocations
	if len(allocations) == 0 && inv.Asset != nil {
		allocations = []investment.InvestmentAssetLink{
			{
				AssetID:         inv.Asset.ID(),
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
		if _, err := r.db.Exec(ctx, insertSql, inv.ID.String(), a.AssetID.String(), a.SharePercentage); err != nil {
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
				AssetID:         uuid.MustParse(astID),
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
	qb.Where(qb.Equal("investment_id", inv.ID.String()))
	qb.Where(qb.Equal("project_id", inv.Project.ProjectID.String()))

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
	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_investments.project_id", filter.ProjectID.String()))
	}
	if filter.AssetID != uuid.Nil {
		qb.Where(fmt.Sprintf("(projecta_investments.asset_id = %s OR projecta_investments.investment_id IN (SELECT investment_id FROM projecta_investment_assets WHERE asset_id = %s))", qb.Var(filter.AssetID.String()), qb.Var(filter.AssetID.String())))
	}

	sql, args := qb.Build()
	row := r.db.QueryRow(ctx, sql, args...)

	inv, err := scanInvestment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvestmentNotFound
		}
		return nil, err
	}

	allocsMap, _ := r.loadAllocations(ctx, []string{inv.ID.String()})
	if allocs, ok := allocsMap[inv.ID.String()]; ok && len(allocs) > 0 {
		inv.SetAssetAllocations(allocs)
	}

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

	if filter.OrderBy != "" && filter.Order != "" {
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

	col := investment.NewCollection(total)
	invIDs := make([]string, 0, total)
	for rows.Next() {
		inv, err := scanInvestment(rows)
		if err != nil {
			return nil, err
		}
		col.Add(inv)
		invIDs = append(invIDs, inv.ID.String())
	}

	if len(invIDs) > 0 {
		allocsMap, _ := r.loadAllocations(ctx, invIDs)
		for _, inv := range col.Elements() {
			if allocs, ok := allocsMap[inv.ID.String()]; ok && len(allocs) > 0 {
				inv.SetAssetAllocations(allocs)
			}
		}
	}

	return col, nil
}

func (r *PgInvestmentRepository) FindTags(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	sql := `
		SELECT DISTINCT tag FROM (
			SELECT unnest(tags) AS tag 
			FROM projecta_investments 
			WHERE project_id = $1 
			UNION 
			SELECT unnest(tags) AS tag 
			FROM projecta_assets 
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
		"projecta_projects.name",
		"projecta_projects.description",
		"projecta_investments.asset_id",
		"projecta_assets.name",
		"projecta_assets.description",
		"projecta_investments.contributor_id",
		"people.first_name",
		qb.As("COALESCE(people.display_name, '')", "display_name"),
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

	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_investments.project_id")
	qb.Join("projecta_assets", "projecta_assets.asset_id = projecta_investments.asset_id")
	qb.Join("people", "people.person_id = projecta_investments.contributor_id")
}

func applyInvestmentFilters(qb *sqlbuilder.SelectBuilder, filter investment.CollectionFilter) {
	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_investments.project_id", filter.ProjectID.String()))
	}
	if filter.AssetID != uuid.Nil {
		qb.Where(fmt.Sprintf("(projecta_investments.asset_id = %s OR projecta_investments.investment_id IN (SELECT investment_id FROM projecta_investment_assets WHERE asset_id = %s))", qb.Var(filter.AssetID.String()), qb.Var(filter.AssetID.String())))
	}
	if filter.OwnerID != uuid.Nil {
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

func scanInvestment(s scannable) (*investment.Investment, error) {
	var (
		invID         string
		projID        string
		projName      string
		projDesc      string
		astID         string
		astName       string
		astDesc       string
		contribID     string
		contribFirst  string
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
		&projName,
		&projDesc,
		&astID,
		&astName,
		&astDesc,
		&contribID,
		&contribFirst,
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

	proj := &projecta.Project{
		ProjectID:   uuid.MustParse(projID),
		Name:        projName,
		Description: projDesc,
	}

	ast := asset.NewAsset(
		uuid.MustParse(astID),
		astName,
		astDesc,
		proj,
		nil,
		nil,
		date,
		nil,
	)

	contrib := &projecta.Owner{
		PersonID:    uuid.MustParse(contribID),
		FirstName:   contribFirst,
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

	return investment.NewInvestment(
		uuid.MustParse(invID),
		proj,
		ast,
		contrib,
		rType,
		money.New(amount, currencyCode),
		timeHours,
		hourlyRateMoney,
		goodsQty,
		gUnit,
		gItem,
		description,
		date,
		tags,
	), nil
}

func (r *PgInvestmentRepository) FindInvestments(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) ([]asset.InvestmentItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = id.String()
	}
	sql := `
		SELECT investment_id, amount, currency, date, tags
		FROM projecta_investments
		WHERE project_id = $1 AND investment_id = ANY($2::uuid[])
	`
	rows, err := r.db.Query(ctx, sql, projectID.String(), idStrs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []asset.InvestmentItem
	for rows.Next() {
		var (
			invIDStr string
			amount   int64
			curr     string
			date     time.Time
			tags     []string
		)
		if err := rows.Scan(&invIDStr, &amount, &curr, &date, &tags); err != nil {
			return nil, err
		}
		items = append(items, asset.InvestmentItem{
			ID:     uuid.MustParse(invIDStr),
			Amount: money.New(amount, curr),
			Date:   date,
			Tags:   tags,
		})
	}
	return items, nil
}

func (r *PgInvestmentRepository) AssignToAsset(ctx context.Context, assetID uuid.UUID, investmentIDs []uuid.UUID) error {
	if len(investmentIDs) == 0 {
		return nil
	}
	idStrs := make([]string, len(investmentIDs))
	for i, id := range investmentIDs {
		idStrs[i] = id.String()
	}

	updateSql := `
		UPDATE projecta_investments
		SET asset_id = $1, updated_at = current_timestamp
		WHERE investment_id = ANY($2::uuid[])
	`
	if _, err := r.db.Exec(ctx, updateSql, assetID.String(), idStrs); err != nil {
		return err
	}

	delSql := `DELETE FROM projecta_investment_assets WHERE investment_id = ANY($1::uuid[])`
	if _, err := r.db.Exec(ctx, delSql, idStrs); err != nil {
		return err
	}

	insertSql := `
		INSERT INTO projecta_investment_assets (investment_id, asset_id, share_percentage)
		SELECT unnest($1::uuid[]), $2, 100.00
	`
	if _, err := r.db.Exec(ctx, insertSql, idStrs, assetID.String()); err != nil {
		return err
	}

	return nil
}

func (r *PgInvestmentRepository) CreateInitialInvestment(ctx context.Context, item asset.InitialInvestment) error {
	amountVal := int64(0)
	currVal := "UAH"
	if item.Amount != nil {
		amountVal = item.Amount.Amount()
		currVal = item.Amount.Currency().Code
	}
	dateVal := item.Date
	if dateVal.IsZero() {
		dateVal = time.Now()
	}
	tagsVal := item.Tags
	if tagsVal == nil {
		tagsVal = []string{}
	}
	invID := item.ID
	if invID == uuid.Nil {
		invID = uuid.New()
	}

	insertInvSql := `
		INSERT INTO projecta_investments (
			investment_id, project_id, asset_id, contributor_id, resource_type,
			amount, currency, description, date, tags, created_at
		) VALUES (
			$1, $2, $3, $4, 'MONEY',
			$5, $6, $7, $8, $9, current_timestamp
		)
	`
	if _, err := r.db.Exec(ctx, insertInvSql,
		invID.String(),
		item.ProjectID.String(),
		item.AssetID.String(),
		item.ContributorID.String(),
		amountVal,
		currVal,
		item.Description,
		dateVal,
		tagsVal,
	); err != nil {
		return err
	}

	insertLinkSql := `
		INSERT INTO projecta_investment_assets (investment_id, asset_id, share_percentage)
		VALUES ($1, $2, 100.00)
		ON CONFLICT (investment_id, asset_id) DO UPDATE SET share_percentage = EXCLUDED.share_percentage
	`
	if _, err := r.db.Exec(ctx, insertLinkSql, invID.String(), item.AssetID.String()); err != nil {
		return err
	}

	return nil
}

