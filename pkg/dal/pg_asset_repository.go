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
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

var ErrFailedToSaveAsset = errors.New("failed to save asset")
var ErrAssetNotFound = errors.New("asset not found")

type PgAssetRepository struct {
	db *PgRepository
}

func NewPgAssetRepository(conn *PgDbConnection) *PgAssetRepository {
	return &PgAssetRepository{
		db: &PgRepository{db: conn},
	}
}

func (r *PgAssetRepository) Save(ctx context.Context, anAsset *asset.Asset) error {
	_, err := r.FindOne(ctx, asset.Filter{
		ID: anAsset.ID(),
	})

	if err != nil {
		if errors.Is(err, ErrAssetNotFound) {
			return r.create(ctx, anAsset)
		}

		return err
	}

	return r.update(ctx, anAsset)
}

func (r *PgAssetRepository) create(ctx context.Context, asset *asset.Asset) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()

	qb.InsertInto("projecta_assets")
	qb.Cols(
		"asset_id",
		"name",
		"description",
		"project_id",
		"type_id",
		"price",
		"currency",
		"acquired_at",
		"owner_id")

	qb.Values(
		asset.ID().String(),
		asset.Name(),
		asset.Description(),
		asset.Project().ProjectID.String(),
		asset.Type().ID.String(),
		asset.Price().Amount(),
		asset.Price().Currency().Code,
		asset.AcquiredAt(),
		asset.Owner().PersonID.String())

	sql, args := qb.Build()

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return errors.Join(ErrFailedToSaveAsset, err)
	}

	return nil
}

func (r *PgAssetRepository) update(ctx context.Context, asset *asset.Asset) error {
	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()

	qb.Update("projecta_assets")
	qb.Set(
		qb.Assign("name", asset.Name()),
		qb.Assign("description", asset.Description()),
		qb.Assign("type_id", asset.Type().ID.String()),
		qb.Assign("price", asset.Price().Amount()),
		qb.Assign("currency", asset.Price().Currency().Code),
		qb.Assign("acquired_at", asset.AcquiredAt()),
	)
	qb.Where(qb.Equal("asset_id", asset.ID().String()))
	qb.Where(qb.Equal("owner_id", asset.Owner().PersonID.String()))

	sql, args := qb.Build()

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return errors.Join(ErrFailedToSaveAsset, err)
	}

	return nil
}

func (r *PgAssetRepository) Remove(ctx context.Context, asset *asset.Asset) error {
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_assets")
	qb.Where(qb.Equal("asset_id", asset.ID().String()))
	qb.Where(qb.Equal("owner_id", asset.Owner().PersonID.String()))

	sql, args := qb.Build()

	res, err := r.db.Exec(ctx, sql, args...)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrAssetNotFound
	}

	return nil
}

func (r *PgAssetRepository) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	setupSelectQueryBuilder(qb)

	qb.Where(qb.Equal("projecta_assets.asset_id", filter.ID.String()))

	if filter.OwnerID != uuid.Nil {
		qb.Where(qb.Equal("projecta_assets.owner_id", filter.OwnerID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.ILike("projecta_assets.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	sql, args := qb.Build()

	var (
		assetID             string
		name                string
		description         string
		projectID           string
		projectName         string
		projectDescription  string
		typeID              string
		typeName            string
		typeDescription     string
		price               int64
		currencyCode        string
		acquiredAt          time.Time
		ownerID             string
		ownerFirstName      string
		ownerDisplayName    string
		categoryID          string
		categoryName        string
		categoryDescription string
	)

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&assetID,
		&name,
		&description,
		&projectID,
		&projectName,
		&projectDescription,
		&typeID,
		&typeName,
		&typeDescription,
		&price,
		&currencyCode,
		&acquiredAt,
		&ownerID,
		&ownerFirstName,
		&ownerDisplayName,
		&categoryID,
		&categoryName,
		&categoryDescription,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}

		return nil, err
	}

	a, err := toAssetFromPg(
		assetID,
		name,
		description,
		projectID,
		projectName,
		projectDescription,
		typeID,
		typeName,
		typeDescription,
		price,
		currencyCode,
		acquiredAt,
		ownerID,
		ownerFirstName,
		ownerDisplayName,
		categoryID,
		categoryName,
		categoryDescription,
	)

	if err != nil {
		return nil, errors.Join(ErrAssetNotFound, err)
	}

	return a, nil
}

func (r *PgAssetRepository) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	if filter.OwnerID != uuid.Nil {
		qb.Where(qb.Equal("projecta_assets.owner_id", filter.OwnerID.String()))
	}

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_assets.project_id", filter.ProjectID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.ILike("projecta_assets.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	if filter.TypeID != uuid.Nil {
		qb.Where(qb.Equal("projecta_assets.type_id", filter.TypeID.String()))
	}

	qb.Select(qb.As("COUNT(*)", "total"))

	sql, args := qb.Build()

	var total int

	if err := r.db.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return nil, err
	}

	qb.Select() // reset select

	setupSelectQueryBuilder(qb)

	qb.Offset(filter.Offset)
	qb.Limit(filter.Limit)

	if filter.OrderBy != "" && filter.Order != "" {
		qb.OrderBy(fmt.Sprintf("projecta_assets.%s %s", filter.OrderBy, filter.Order.String()))
	} else {
		qb.OrderBy("projecta_assets.acquired_at DESC")
	}

	sql, args = qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	collection := asset.NewCollection(total)

	for rows.Next() {
		var (
			assetID             string
			name                string
			description         string
			projectID           string
			projectName         string
			projectDescription  string
			typeID              string
			typeName            string
			typeDescription     string
			price               int64
			currencyCode        string
			acquiredAt          time.Time
			ownerID             string
			ownerFirstName      string
			ownerDisplayName    string
			categoryID          string
			categoryName        string
			categoryDescription string
		)

		if err = rows.Scan(
			&assetID,
			&name,
			&description,
			&projectID,
			&projectName,
			&projectDescription,
			&typeID,
			&typeName,
			&typeDescription,
			&price,
			&currencyCode,
			&acquiredAt,
			&ownerID,
			&ownerFirstName,
			&ownerDisplayName,
			&categoryID,
			&categoryName,
			&categoryDescription,
		); err != nil {
			return nil, err
		}

		a, err := toAssetFromPg(
			assetID,
			name,
			description,
			projectID,
			projectName,
			projectDescription,
			typeID,
			typeName,
			typeDescription,
			price,
			currencyCode,
			acquiredAt,
			ownerID,
			ownerFirstName,
			ownerDisplayName,
			categoryID,
			categoryName,
			categoryDescription,
		)

		if err != nil {
			return nil, err
		}

		collection.Add(a)
	}

	return collection, nil
}

func setupSelectQueryBuilder(qb *sqlbuilder.SelectBuilder) {
	qb.Select(
		"asset_id",
		"projecta_assets.name",
		"projecta_assets.description",
		"projecta_projects.project_id",
		qb.As("projecta_projects.name", "project_name"),
		qb.As("projecta_projects.description", "project_description"),
		"projecta_assets.type_id",
		qb.As("projecta_cost_types.name", "type_name"),
		qb.As("projecta_cost_types.description", "type_description"),
		"projecta_assets.price",
		"projecta_assets.currency",
		"projecta_assets.acquired_at",
		"projecta_assets.owner_id",
		"people.first_name",
		qb.As("COALESCE(people.display_name, '')", "display_name"),
		"projecta_cost_categories.category_id",
		qb.As("projecta_cost_categories.name", "category_name"),
		qb.As("projecta_cost_categories.description", "category_description"),
	)

	qb.Join("people", "people.person_id = projecta_assets.owner_id")
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_assets.project_id")
	qb.Join("projecta_cost_types", "projecta_cost_types.type_id = projecta_assets.type_id")
	qb.Join("projecta_cost_categories", "projecta_cost_categories.category_id = projecta_cost_types.category_id")
}

func toAssetFromPg(
	assetID string,
	name string,
	description string,
	projectID string,
	projectName string,
	projectDescription string,
	typeID string,
	typeName string,
	typeDescription string,
	price int64,
	currencyCode string,
	acquiredAt time.Time,
	ownerID string,
	ownerFirstName string,
	ownerDisplayName string,
	categoryID string,
	categoryName string,
	categoryDescription string,
) (*asset.Asset, error) {
	owner := &projecta.Owner{
		PersonID:    uuid.MustParse(ownerID),
		FirstName:   ownerFirstName,
		DisplayName: ownerDisplayName,
	}

	project := &projecta.Project{
		ProjectID:   uuid.MustParse(projectID),
		Name:        projectName,
		Description: projectDescription,
	}

	category := &projecta.CostCategory{
		ID:          uuid.MustParse(categoryID),
		Name:        categoryName,
		Description: categoryDescription,
	}

	costType := &projecta.CostType{
		ID:          uuid.MustParse(typeID),
		Name:        typeName,
		Description: typeDescription,
		Category:    category,
	}

	priceMoney := money.New(price, currencyCode)

	return asset.NewAsset(
		uuid.MustParse(assetID),
		name,
		description,
		project,
		costType,
		priceMoney,
		acquiredAt,
		owner,
	), nil
}
