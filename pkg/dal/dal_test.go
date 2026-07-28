package dal

import (
	"context"
	types "database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type mockRow struct {
	row []any
	err error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	for i, v := range dest {
		if i >= len(m.row) {
			break
		}
		val := m.row[i]
		if val == nil {
			continue
		}
		switch d := v.(type) {
		case *string:
			if str, ok := val.(string); ok {
				*d = str
			} else {
				*d = fmt.Sprintf("%v", val)
			}
		case *int64:
			switch n := val.(type) {
			case int64:
				*d = n
			case int:
				*d = int64(n)
			}
		case *int:
			switch n := val.(type) {
			case int:
				*d = n
			case int64:
				*d = int(n)
			}
		case *bool:
			*d = val.(bool)
		case *time.Time:
			if tVal, ok := val.(time.Time); ok {
				*d = tVal
			} else if str, ok := val.(string); ok {
				if parsed, err := time.Parse(time.RFC3339, str); err == nil {
					*d = parsed
				}
			}
		case *types.NullString:
			d.String = val.(string)
			d.Valid = true
		}
	}
	return nil
}

type mockRows struct {
	data [][]any
	idx  int
	err  error
}

func (m *mockRows) Close()                                      {}
func (m *mockRows) Err() error                                  { return m.err }
func (m *mockRows) CommandTag() pgconn.CommandTag               { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Next() bool {
	if m.idx < len(m.data) {
		m.idx++
		return true
	}
	return false
}
func (m *mockRows) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	row := m.data[m.idx-1]
	for i, d := range dest {
		if i >= len(row) {
			break
		}
		val := row[i]
		if val == nil {
			continue
		}
		switch target := d.(type) {
		case *string:
			if str, ok := val.(string); ok {
				*target = str
			} else {
				*target = fmt.Sprintf("%v", val)
			}
		case *int64:
			switch n := val.(type) {
			case int64:
				*target = n
			case int:
				*target = int64(n)
			}
		case *int:
			switch n := val.(type) {
			case int:
				*target = n
			case int64:
				*target = int(n)
			}
		case *bool:
			*target = val.(bool)
		case *time.Time:
			if tVal, ok := val.(time.Time); ok {
				*target = tVal
			} else if str, ok := val.(string); ok {
				if parsed, err := time.Parse(time.RFC3339, str); err == nil {
					*target = parsed
				}
			}
		case *types.NullString:
			target.String = val.(string)
			target.Valid = true
		}
	}
	return nil
}
func (m *mockRows) Values() ([]any, error) { return nil, nil }
func (m *mockRows) RawValues() [][]byte    { return nil }
func (m *mockRows) Conn() *pgx.Conn        { return nil }

type mockPgDb struct {
	execErr    error
	execTag    pgconn.CommandTag
	queryErr   error
	rowsData   [][]any
	rowVal     []any
	rowErr     error
	countErr   error
	isNotFound bool
	zeroTotal  bool
}

func (m *mockPgDb) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	if m.execErr != nil {
		return pgconn.CommandTag{}, m.execErr
	}
	if m.execTag.String() != "" {
		return m.execTag, nil
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *mockPgDb) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return &mockRows{data: m.rowsData}, nil
}

func (m *mockPgDb) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if m.rowErr != nil {
		return &mockRow{err: m.rowErr}
	}
	if m.isNotFound {
		return &mockRow{err: pgx.ErrNoRows}
	}
	if strings.Contains(sql, "COUNT") || strings.Contains(sql, "count") || strings.Contains(sql, "1 as exists") {
		if m.countErr != nil {
			return &mockRow{err: m.countErr}
		}
		if m.zeroTotal {
			return &mockRow{row: []any{int(0), int64(0)}}
		}
		return &mockRow{row: []any{int(1), int64(1)}}
	}
	return &mockRow{row: m.rowVal}
}

func withMockDb(ctx context.Context, db PgDb) context.Context {
	return context.WithValue(ctx, txKey{}, db)
}

func TestPgDbConnectionAndRepo(t *testing.T) {
	t.Run("NewPgDbConnection success and error", func(t *testing.T) {
		_, err := NewPgDbConnection("invalid connection string")
		if err == nil {
			t.Errorf("expected error for invalid config")
		}

		conn, err := NewPgDbConnection("postgres://user:pass@localhost:5432/dbname")
		if err != nil || conn == nil {
			t.Fatalf("unexpected error creating PgDbConnection: %v", err)
		}
		conn.Close()
	})

	t.Run("errorRow Scan", func(t *testing.T) {
		e := &errorRow{err: errors.New("scan error")}
		if err := e.Scan(); err == nil {
			t.Errorf("expected scan error")
		}
	})

	t.Run("PgDbConnection Tx, Close, Ping", func(t *testing.T) {
		emptyConn := &PgDbConnection{}

		// Tx when GetConnection returns error (nil pool)
		_, err := emptyConn.Tx(context.Background(), func(ctx context.Context) (any, error) { return nil, nil })
		if err == nil {
			t.Errorf("expected Tx error when pool is nil")
		}

		// Tx with mockDb in context
		mockDb := &mockPgDb{}
		ctxTx := context.WithValue(context.Background(), txKey{}, mockDb)
		executed := false
		res, err := emptyConn.Tx(ctxTx, func(ctx context.Context) (any, error) {
			executed = true
			return "success", nil
		})
		if err != nil || !executed || res != "success" {
			t.Errorf("expected Tx execution for mock db in context")
		}

		// Tx returning function error
		txErr := errors.New("tx error")
		_, err = emptyConn.Tx(ctxTx, func(ctx context.Context) (any, error) {
			return nil, txErr
		})
		if !errors.Is(err, txErr) {
			t.Errorf("expected txErr to be returned")
		}

		emptyConn.Close()
		_ = emptyConn.Ping(context.Background())

		// PgRepository with nil db
		repo := &PgRepository{db: nil}
		_, err = repo.Exec(context.Background(), "SELECT 1")
		if err == nil {
			t.Errorf("expected error on nil db Exec")
		}
		_, err = repo.Query(context.Background(), "SELECT 1")
		if err == nil {
			t.Errorf("expected error on nil db Query")
		}
		row := repo.QueryRow(context.Background(), "SELECT 1")
		if err = row.Scan(); err == nil {
			t.Errorf("expected error on nil db QueryRow Scan")
		}
	})
}

func TestPgPeopleRepository(t *testing.T) {
	dbConn := &PgDbConnection{}
	repo := NewPgPeopleRepository(dbConn)
	pID := uuid.New()
	cred, _ := people.NewCredentials("LOCAL", "john@example.com", "secret")
	person, _ := people.NewPerson(pID, "John", "Doe", "J.D.", []people.Credentials{cred})

	t.Run("Register success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{}
		ctx := withMockDb(context.Background(), mockDb)

		err := repo.Register(ctx, person)
		if err != nil {
			t.Errorf("unexpected Register error: %v", err)
		}

		mockDbErr := &mockPgDb{execErr: errors.New("exec error")}
		ctxErr := withMockDb(context.Background(), mockDbErr)
		err = repo.Register(ctxErr, person)
		if err == nil {
			t.Errorf("expected exec error")
		}
	})

	t.Run("FindCredentials success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowVal: []any{pID.String(), "john@example.com"}}
		ctx := withMockDb(context.Background(), mockDb)

		gotID, identity, err := repo.FindCredentials(ctx, "LOCAL", "john@example.com")
		if err != nil || gotID != pID || identity != "john@example.com" {
			t.Errorf("FindCredentials error: %v", err)
		}

		mockDbErr := &mockPgDb{rowErr: pgx.ErrNoRows}
		ctxErr := withMockDb(context.Background(), mockDbErr)
		_, _, err = repo.FindCredentials(ctxErr, "LOCAL", "bad")
		if err == nil {
			t.Errorf("expected not found error")
		}

		mockDbBadUUID := &mockPgDb{rowVal: []any{"not-uuid", "john@example.com"}}
		ctxBadUUID := withMockDb(context.Background(), mockDbBadUUID)
		_, _, err = repo.FindCredentials(ctxBadUUID, "LOCAL", "john@example.com")
		if err == nil {
			t.Errorf("expected bad UUID error")
		}
	})

	t.Run("FindByID success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowVal: []any{"John", "Doe", "J.D."}}
		ctx := withMockDb(context.Background(), mockDb)

		p, err := repo.FindByID(ctx, pID)
		if err != nil || p == nil {
			t.Errorf("FindByID error: %v", err)
		}

		mockDbErr := &mockPgDb{rowErr: pgx.ErrNoRows}
		ctxErr := withMockDb(context.Background(), mockDbErr)
		_, err = repo.FindByID(ctxErr, pID)
		if err == nil {
			t.Errorf("expected not found error")
		}

		mockDbOtherErr := &mockPgDb{rowErr: errors.New("db failure")}
		ctxOther := withMockDb(context.Background(), mockDbOtherErr)
		_, err = repo.FindByID(ctxOther, pID)
		if err == nil {
			t.Errorf("expected internal error on db failure")
		}
	})

	t.Run("toPersonFromPg test", func(t *testing.T) {
		p, err := toPersonFromPg(pID.String(), "John", "Doe", "J.D.")
		if err != nil || p.FirstName() != "John" {
			t.Errorf("toPersonFromPg error: %v", err)
		}
	})
}

func TestPgProjectRepository(t *testing.T) {
	dbConn := &PgDbConnection{}
	repo := NewPgProjectRepository(dbConn)

	pID := uuid.New()
	ownerID := uuid.New()
	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "John Doe"}
	now := time.Now()
	proj, _ := projecta.NewProject(pID, "Project A", "Desc", owner, now, now)

	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, ownerID)

	t.Run("FindOne unauthenticated", func(t *testing.T) {
		_, err := repo.FindOne(context.Background(), projecta.ProjectFilter{})
		if err == nil {
			t.Errorf("expected unauthenticated error")
		}
	})

	t.Run("FindOne success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowVal: []any{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "J.D."}}
		ctx := withMockDb(authedCtx, mockDb)

		p, err := repo.FindOne(ctx, projecta.ProjectFilter{ProjectID: pID, Name: "Project A"})
		if err != nil || p == nil {
			t.Errorf("FindOne error: %v", err)
		}

		mockDbErr := &mockPgDb{rowErr: pgx.ErrNoRows}
		ctxErr := withMockDb(authedCtx, mockDbErr)
		_, err = repo.FindOne(ctxErr, projecta.ProjectFilter{})
		if err == nil {
			t.Errorf("expected not found error")
		}
	})

	t.Run("Create, Update, Remove", func(t *testing.T) {
		mockDb := &mockPgDb{}
		ctx := withMockDb(authedCtx, mockDb)

		err := repo.Create(ctx, proj)
		if err != nil {
			t.Errorf("Create error: %v", err)
		}

		err = repo.Update(ctx, proj)
		if err != nil {
			t.Errorf("Update error: %v", err)
		}

		err = repo.Remove(ctx, proj)
		if err != nil {
			t.Errorf("Remove error: %v", err)
		}
	})

	t.Run("Find collection success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowsData: [][]any{{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "J.D."}}}
		ctx := withMockDb(authedCtx, mockDb)

		projects, err := repo.Find(ctx, projecta.ProjectCollectionFilter{Name: "Project A", Pagination: core.Pagination{Limit: 10, Offset: 0}})
		if err != nil || len(projects) != 1 {
			t.Errorf("Find error: %v", err)
		}

		// Unauthenticated
		_, err = repo.Find(context.Background(), projecta.ProjectCollectionFilter{})
		if err == nil {
			t.Errorf("expected unauthenticated error")
		}

		// Query error
		mockDbErr := &mockPgDb{queryErr: errors.New("query error")}
		ctxErr := withMockDb(authedCtx, mockDbErr)
		_, err = repo.Find(ctxErr, projecta.ProjectCollectionFilter{})
		if err == nil {
			t.Errorf("expected query error")
		}
	})

	t.Run("toProject test", func(t *testing.T) {
		p, err := toProject(pID.String(), "Name", "Desc", ownerID.String(), "John", "Doe", "J.D.", now, now)
		if err != nil || p == nil {
			t.Errorf("toProject error: %v", err)
		}
	})
}

func TestPgCategoryTypePaymentAssetRepositories(t *testing.T) {
	dbConn := &PgDbConnection{}
	projRepo := NewPgProjectRepository(dbConn)
	catRepo := NewPgCategoryRepository(dbConn)
	typeRepo := NewPgCostTypeRepository(dbConn)
	payRepo := NewPgPaymentRepository(dbConn)
	astRepo := NewPgAssetRepository(dbConn)

	pID := uuid.New()
	catID := uuid.New()
	typeID := uuid.New()
	payID := uuid.New()
	astID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()

	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "John"}
	proj, _ := projecta.NewProject(pID, "Project", "Desc", owner, now, now)
	cat, _ := projecta.NewCostCategory(catID, pID, "Category", "Desc")
	costType, _ := projecta.NewCostType(pID, cat, "Type", "Desc")
	pay := projecta.NewPayment(payID, proj, owner, costType, "Payment", money.New(100, money.USD), now, projecta.DownPayment)
	ast := asset.NewAsset(astID, "Laptop", "Desc", proj, costType, money.New(1000, money.USD), now, owner)

	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, ownerID)

	t.Run("PgCostCategoryRepository methods and create branch", func(t *testing.T) {
		mockDb := &mockPgDb{
			rowVal:   []any{catID.String(), pID.String(), "Category", "Desc"},
			rowsData: [][]any{{catID.String(), pID.String(), "Category", "Desc"}},
		}
		ctx := withMockDb(authedCtx, mockDb)

		err := catRepo.Save(ctx, cat)
		if err != nil {
			t.Errorf("Save category error: %v", err)
		}

		// Save calling create when not found
		mockDbCreate := &mockPgDb{isNotFound: true}
		ctxCreate := withMockDb(authedCtx, mockDbCreate)
		err = catRepo.Save(ctxCreate, cat)
		if err != nil {
			t.Errorf("Save category create branch error: %v", err)
		}

		// Save update error
		mockDbUpdErr := &mockPgDb{execTag: pgconn.NewCommandTag("UPDATE 0")}
		ctxUpdErr := withMockDb(authedCtx, mockDbUpdErr)
		err = catRepo.Save(ctxUpdErr, cat)
		if err == nil {
			t.Errorf("expected error on category update failure")
		}

		c, err := catRepo.FindOne(ctx, projecta.CategoryFilter{CategoryID: catID, ProjectID: pID, Name: "Category"})
		if err != nil || c == nil {
			t.Errorf("FindOne category error: %v", err)
		}

		// FindOne other db error
		mockDbOtherErr := &mockPgDb{rowErr: errors.New("db error")}
		ctxOtherErr := withMockDb(authedCtx, mockDbOtherErr)
		_, err = catRepo.FindOne(ctxOtherErr, projecta.CategoryFilter{CategoryID: catID, ProjectID: pID})
		if err == nil {
			t.Errorf("expected internal error on category FindOne db error")
		}

		cols, err := catRepo.Find(ctx, projecta.CategoryCollectionFilter{ProjectID: pID, Name: "Cat", Pagination: core.Pagination{Limit: 10}})
		if err != nil || cols == nil {
			t.Errorf("Find category collection error: %v", err)
		}

		err = catRepo.Remove(ctx, cat)
		if err != nil {
			t.Errorf("Remove category error: %v", err)
		}

		// Remove zero rows affected error
		mockDbZero := &mockPgDb{execTag: pgconn.NewCommandTag("DELETE 0")}
		ctxZero := withMockDb(authedCtx, mockDbZero)
		err = catRepo.Remove(ctxZero, cat)
		if err == nil {
			t.Errorf("expected remove error on 0 rows affected")
		}
	})

	t.Run("toCostCategory test", func(t *testing.T) {
		c, err := toCostCategory(catID.String(), pID.String(), "Cat", "Desc")
		if err != nil || c == nil {
			t.Errorf("toCostCategory error: %v", err)
		}
	})

	t.Run("PgCostTypeRepository methods and zero total branch", func(t *testing.T) {
		mockDb := &mockPgDb{
			rowVal:   []any{typeID.String(), pID.String(), catID.String(), "Category", "Cat Desc", "Type", "Type Desc"},
			rowsData: [][]any{{typeID.String(), pID.String(), catID.String(), "Category", "Cat Desc", "Type", "Type Desc"}},
		}
		ctx := withMockDb(authedCtx, mockDb)

		err := typeRepo.Save(ctx, costType)
		if err != nil {
			t.Errorf("Save cost type error: %v", err)
		}

		tp, err := typeRepo.FindOne(ctx, projecta.TypeFilter{TypeID: typeID, ProjectID: pID, Name: "Type"})
		if err != nil || tp == nil {
			t.Errorf("FindOne cost type error: %v", err)
		}

		// FindOne other db error
		mockDbOtherErr := &mockPgDb{rowErr: errors.New("db error")}
		ctxOtherErr := withMockDb(authedCtx, mockDbOtherErr)
		_, err = typeRepo.FindOne(ctxOtherErr, projecta.TypeFilter{TypeID: typeID, ProjectID: pID})
		if err == nil {
			t.Errorf("expected internal error on cost type FindOne db error")
		}

		cols, err := typeRepo.Find(ctx, projecta.TypeCollectionFilter{ProjectID: pID, Name: "Type", Pagination: core.Pagination{Limit: 10}})
		if err != nil || cols == nil {
			t.Errorf("Find cost type collection error: %v", err)
		}

		// Find with 0 total
		mockDbZero := &mockPgDb{zeroTotal: true}
		ctxZero := withMockDb(authedCtx, mockDbZero)
		zeroCols, err := typeRepo.Find(ctxZero, projecta.TypeCollectionFilter{ProjectID: pID})
		if err != nil || zeroCols.Total() != 0 {
			t.Errorf("expected 0 total cost types")
		}

		err = typeRepo.Remove(ctx, costType)
		if err != nil {
			t.Errorf("Remove cost type error: %v", err)
		}
	})

	t.Run("toCostType test", func(t *testing.T) {
		tp, err := toCostType(typeID.String(), pID.String(), "Type", "Desc", catID.String(), "Cat")
		if err != nil || tp == nil {
			t.Errorf("toCostType error: %v", err)
		}
	})

	t.Run("PgPaymentRepository methods and zero total branch", func(t *testing.T) {
		mockDb := &mockPgDb{
			rowVal: []any{
				payID.String(), pID.String(), "Project", catID.String(), "Cat",
				typeID.String(), "Type", int64(100), "USD", "Payment",
				ownerID.String(), "John", "J.D.", now, "DOWN_PAYMENT",
			},
			rowsData: [][]any{{
				payID.String(), pID.String(), "Project", catID.String(), "Cat",
				typeID.String(), "Type", int64(100), "USD", "Payment",
				ownerID.String(), "John", "J.D.", now, "DOWN_PAYMENT",
			}},
		}
		ctx := withMockDb(authedCtx, mockDb)

		err := payRepo.Save(ctx, pay)
		if err != nil {
			t.Errorf("Save payment error: %v", err)
		}

		// Save calling create when not found (0 rows affected on select/update)
		mockDbCreate := &mockPgDb{execTag: pgconn.NewCommandTag("UPDATE 0")}
		ctxCreate := withMockDb(authedCtx, mockDbCreate)
		err = payRepo.Save(ctxCreate, pay)
		if err != nil {
			t.Errorf("Save payment create branch error: %v", err)
		}

		p, err := payRepo.FindOne(ctx, projecta.PaymentFilter{PaymentID: payID, ProjectID: pID})
		if err != nil || p == nil {
			t.Errorf("FindOne payment error: %v", err)
		}

		// FindOne other db error
		mockDbOtherErr := &mockPgDb{rowErr: errors.New("db error")}
		ctxOtherErr := withMockDb(authedCtx, mockDbOtherErr)
		_, err = payRepo.FindOne(ctxOtherErr, projecta.PaymentFilter{PaymentID: payID, ProjectID: pID})
		if err == nil {
			t.Errorf("expected internal error on payment FindOne db error")
		}

		cols, err := payRepo.Find(ctx, projecta.PaymentCollectionFilter{
			ProjectID:  pID,
			CategoryID: catID,
			TypeID:     typeID,
			Sorting:    core.Sorting{OrderBy: "date", Order: core.DESC},
			Pagination: core.Pagination{Limit: 10, Offset: 0},
		})
		if err != nil || cols == nil {
			t.Errorf("Find payment collection error: %v", err)
		}

		// Find zero total
		mockDbZero := &mockPgDb{zeroTotal: true}
		ctxZero := withMockDb(authedCtx, mockDbZero)
		zeroCols, err := payRepo.Find(ctxZero, projecta.PaymentCollectionFilter{ProjectID: pID})
		if err != nil || zeroCols.Total() != 0 {
			t.Errorf("expected 0 total payments")
		}

		err = payRepo.Remove(ctx, pay)
		if err != nil {
			t.Errorf("Remove payment error: %v", err)
		}

		// Remove exec error
		mockDbRemErr := &mockPgDb{execErr: errors.New("exec error")}
		ctxRemErr := withMockDb(authedCtx, mockDbRemErr)
		err = payRepo.Remove(ctxRemErr, pay)
		if err == nil {
			t.Errorf("expected exec error on payment Remove")
		}
	})

	t.Run("toExpense test", func(t *testing.T) {
		p := toExpense(payID.String(), pID.String(), "Project", catID.String(), "Cat", typeID.String(), "Type", 100, "USD", "Desc", ownerID.String(), "John", "J.D.", now, "DOWN_PAYMENT")
		if p == nil {
			t.Errorf("expected non-nil payment")
		}
	})

	t.Run("PgAssetRepository methods and zero total branch", func(t *testing.T) {
		mockDb := &mockPgDb{
			rowVal: []any{
				astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc",
				typeID.String(), "Type", "Type Desc", int64(1000), "USD", now,
				ownerID.String(), "John", "J.D.", catID.String(), "Cat", "Cat Desc",
			},
			rowsData: [][]any{{
				astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc",
				typeID.String(), "Type", "Type Desc", int64(1000), "USD", now,
				ownerID.String(), "John", "J.D.", catID.String(), "Cat", "Cat Desc",
			}},
		}
		ctx := withMockDb(authedCtx, mockDb)

		err := astRepo.Save(ctx, ast)
		if err != nil {
			t.Errorf("Save asset error: %v", err)
		}

		// Save calling create when ErrAssetNotFound
		mockDbCreate := &mockPgDb{rowErr: ErrAssetNotFound}
		ctxCreate := withMockDb(authedCtx, mockDbCreate)
		err = astRepo.Save(ctxCreate, ast)
		if err != nil {
			t.Errorf("Save asset create branch error: %v", err)
		}

		// Save update failure (exec error)
		mockDbUpdErr := &mockPgDb{
			rowVal: []any{
				astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc",
				typeID.String(), "Type", "Type Desc", int64(1000), "USD", now,
				ownerID.String(), "John", "J.D.", catID.String(), "Cat", "Cat Desc",
			},
			execErr: errors.New("exec update error"),
		}
		ctxUpdErr := withMockDb(authedCtx, mockDbUpdErr)
		err = astRepo.Save(ctxUpdErr, ast)
		if err == nil {
			t.Errorf("expected error on asset update failure")
		}

		a, err := astRepo.FindOne(ctx, asset.Filter{ID: astID, ProjectID: pID})
		if err != nil || a == nil {
			t.Errorf("FindOne asset error: %v", err)
		}

		// FindOne non-ErrAssetNotFound error
		mockDbOtherErr := &mockPgDb{rowErr: errors.New("db failure")}
		ctxOtherErr := withMockDb(authedCtx, mockDbOtherErr)
		_, err = astRepo.FindOne(ctxOtherErr, asset.Filter{ID: astID, ProjectID: pID})
		if err == nil {
			t.Errorf("expected error on asset FindOne db failure")
		}

		cols, err := astRepo.Find(ctx, asset.CollectionFilter{
			ProjectID:  pID,
			TypeID:     typeID,
			Name:       "Laptop",
			Sorting:    core.Sorting{OrderBy: "price", Order: core.ASC},
			Pagination: core.Pagination{Limit: 10, Offset: 0},
		})
		if err != nil || cols == nil {
			t.Errorf("Find asset collection error: %v", err)
		}

		// Find zero total
		mockDbZero := &mockPgDb{zeroTotal: true}
		ctxZero := withMockDb(authedCtx, mockDbZero)
		zeroCols, err := astRepo.Find(ctxZero, asset.CollectionFilter{ProjectID: pID})
		if err != nil || zeroCols.Total() != 0 {
			t.Errorf("expected 0 total assets")
		}

		err = astRepo.Remove(ctx, ast)
		if err != nil {
			t.Errorf("Remove asset error: %v", err)
		}

		// Remove zero rows affected error
		mockDbRemZero := &mockPgDb{execTag: pgconn.NewCommandTag("DELETE 0")}
		ctxRemZero := withMockDb(authedCtx, mockDbRemZero)
		err = astRepo.Remove(ctxRemZero, ast)
		if err == nil {
			t.Errorf("expected error on 0 rows deleted for asset")
		}
	})

	t.Run("toAssetFromPg test", func(t *testing.T) {
		a, err := toAssetFromPg(astID.String(), "Laptop", "Desc", pID.String(), "Proj", "Desc", typeID.String(), "Type", "Desc", 1000, "USD", now, ownerID.String(), "John", "J.D.", catID.String(), "Cat", "Desc")
		if err != nil || a == nil {
			t.Errorf("toAssetFromPg error: %v", err)
		}
	})

	t.Run("Repository additional FindOne and error branches", func(t *testing.T) {
		// toPersonFromPg with empty fields
		_, err := toPersonFromPg(pID.String(), "", "", "")
		if err == nil {
			t.Errorf("expected error for empty person fields")
		}

		// toProject with valid fields
		p1, err := toProject(pID.String(), "Project 1", "Description", ownerID.String(), "John", "Doe", "J.D.", now, now)
		if err != nil || p1 == nil {
			t.Errorf("toProject error: %v", err)
		}

		// toCostCategory with valid fields
		c1, err := toCostCategory(catID.String(), pID.String(), "Category 1", "Description")
		if err != nil || c1 == nil {
			t.Errorf("toCostCategory error: %v", err)
		}

		mockDbProj := &mockPgDb{rowVal: []any{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "J.D."}}
		ctxProj := withMockDb(authedCtx, mockDbProj)
		p, err := projRepo.FindOne(ctxProj, projecta.ProjectFilter{ProjectID: pID, Name: "Project A"})
		if err != nil || p == nil {
			t.Errorf("FindOne project error: %v", err)
		}

		mockDbCat := &mockPgDb{rowVal: []any{catID.String(), pID.String(), "Category", "Desc"}}
		ctxCat := withMockDb(authedCtx, mockDbCat)
		c, err := catRepo.FindOne(ctxCat, projecta.CategoryFilter{ProjectID: pID})
		if err != nil || c == nil {
			t.Errorf("FindOne category by ProjectID error: %v", err)
		}

		mockDbType := &mockPgDb{rowVal: []any{typeID.String(), pID.String(), catID.String(), "Category", "Cat Desc", "Type", "Type Desc"}}
		ctxType := withMockDb(authedCtx, mockDbType)
		tp, err := typeRepo.FindOne(ctxType, projecta.TypeFilter{ProjectID: pID})
		if err != nil || tp == nil {
			t.Errorf("FindOne cost type by ProjectID error: %v", err)
		}

		mockDbPay := &mockPgDb{rowVal: []any{
			payID.String(), pID.String(), "Project", catID.String(), "Cat",
			typeID.String(), "Type", int64(100), "USD", "Payment",
			ownerID.String(), "John", "J.D.", now, "DOWN_PAYMENT",
		}}
		ctxPay := withMockDb(authedCtx, mockDbPay)
		payOne, err := payRepo.FindOne(ctxPay, projecta.PaymentFilter{ProjectID: pID})
		if err != nil || payOne == nil {
			t.Errorf("FindOne payment by ProjectID error: %v", err)
		}

		mockDbAst := &mockPgDb{rowVal: []any{
			astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc",
			typeID.String(), "Type", "Type Desc", int64(1000), "USD", now,
			ownerID.String(), "John", "J.D.", catID.String(), "Cat", "Cat Desc",
		}}
		ctxAst := withMockDb(authedCtx, mockDbAst)
		astOne, err := astRepo.FindOne(ctxAst, asset.Filter{ID: astID, ProjectID: pID, Name: "Laptop"})
		if err != nil || astOne == nil {
			t.Errorf("FindOne asset by all filters error: %v", err)
		}
	})

	t.Run("Repository Save row error branches", func(t *testing.T) {
		mockDbRowErr := &mockPgDb{rowErr: errors.New("query row error")}
		ctxRowErr := withMockDb(authedCtx, mockDbRowErr)

		err := catRepo.Save(ctxRowErr, cat)
		if err == nil {
			t.Errorf("expected Save category row error")
		}
	})

	t.Run("Repository Find count error branches", func(t *testing.T) {
		mockDbCountErr := &mockPgDb{countErr: errors.New("count query error")}
		ctxCountErr := withMockDb(authedCtx, mockDbCountErr)

		_, err := catRepo.Find(ctxCountErr, projecta.CategoryCollectionFilter{ProjectID: pID})
		if err == nil {
			t.Errorf("expected Find category count error")
		}

		_, err = typeRepo.Find(ctxCountErr, projecta.TypeCollectionFilter{ProjectID: pID})
		if err == nil {
			t.Errorf("expected Find type count error")
		}

		_, err = payRepo.Find(ctxCountErr, projecta.PaymentCollectionFilter{ProjectID: pID})
		if err == nil {
			t.Errorf("expected Find payment count error")
		}

		_, err = astRepo.Find(ctxCountErr, asset.CollectionFilter{ProjectID: pID})
		if err == nil {
			t.Errorf("expected Find asset count error")
		}
	})
}
