package main

import (
	"context"
	"errors"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/crypto"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
	"gitlab.com/massimo-ua/projecta/pkg/dal"
	"gitlab.com/massimo-ua/projecta/pkg/logger"
	"gitlab.com/massimo-ua/projecta/pkg/web"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

func createErrorHandler(logger core.Logger) func(err error) {
	return func(err error) {
		logger.Error("failed to start the application", err, nil)
		os.Exit(1)
	}
}

func setupAppHandler(config *core.AppConfig, db *dal.PgDbConnection) (http.Handler, error) {
	peopleRepository := dal.NewPgPeopleRepository(db)
	hasher := crypto.NewBcryptHasher(0)
	googleAuth, err := crypto.NewGoogleAuthProvider(crypto.GoogleAuthConfig{
		ClientID:      config.GoogleClientID,
		ClientSecret:  config.GoogleClientSecret,
		CertsCacheTTL: config.GoogleCertTTL,
	})

	if err != nil {
		return nil, err
	}

	tokenProvider := crypto.NewJwtTokenProvider(
		config.JwtSecret,
		config.TokenTTL,
		hasher,
	)
	invitationRepository := dal.NewPgInvitationRepository(db)
	authService := people.NewAuthService(
		db,
		peopleRepository,
		tokenProvider,
		hasher,
		googleAuth,
		invitationRepository,
	)
	invitationService := people.NewInvitationService(db, invitationRepository, peopleRepository)

	customerService := people.NewCustomerService(db, peopleRepository, hasher)
	projectRepository := dal.NewPgProjectRepository(db)
	assetRepository := dal.NewPgAssetRepository(db)
	peopleService := projecta.NewPeopleService(peopleRepository)
	projectService := projecta.NewProjectService(projectRepository, peopleService)
	rateProvider := currency.NewNBUCurrencyRateProvider(currency.NBUCurrencyRateProviderOptions{
		SupportedCurrencies: []string{"UAH", "USD", "EUR", "PLN"},
		CacheTTL:            30 * time.Minute,
	})

	investmentRepository := dal.NewPgInvestmentRepository(db)
	assetService := asset.NewService(
		db,
		assetRepository,
		peopleService,
		projectRepository,
		investmentRepository,
		rateProvider,
	)

	investmentService := investment.NewService(
		db,
		investmentRepository,
		assetRepository,
		projectRepository,
		peopleService,
	)

	statementService := projecta.NewStatementService(
		projectService,
		&investmentTransactionFinder{svc: investmentService},
	)

	return web.MakeHTTPHandler(
		customerService,
		tokenProvider,
		authService,
		invitationService,
		projectService,
		statementService,
		assetService,
		rateProvider,
		investmentService,
	)
}

type investmentTransactionFinder struct {
	svc investment.Service
}

func (f *investmentTransactionFinder) FindExisting(ctx context.Context, projectID uuid.UUID) ([]projecta.ExistingTransaction, error) {
	col, err := f.svc.Find(ctx, investment.CollectionFilter{
		ProjectID: projectID,
		Pagination: core.Pagination{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	res := make([]projecta.ExistingTransaction, 0, len(col.Elements()))
	for _, inv := range col.Elements() {
		res = append(res, projecta.ExistingTransaction{
			Date:        inv.Date(),
			Amount:      inv.Amount().Amount(),
			Description: inv.Description(),
		})
	}
	return res, nil
}

func main() {
	log := logger.New(dbUri, jwtSecret)
	handleError := createErrorHandler(log)

	config, err := loadConfig()

	if err != nil {
		handleError(err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := dal.NewPgDbConnection(config.DbUri)

	if err != nil {
		handleError(exceptions.NewInternalException("failed to connect to the database", err))
		return
	}

	defer db.Close()

	startTime := time.Now()
	if err = db.Ping(ctx); err != nil {
		handleError(exceptions.NewInternalException("database ping failed", err))
	}

	log.Info("Database Connection Established", map[string]any{"connection_time": time.Since(startTime)})

	webAPI, err := setupAppHandler(config, db)

	if err != nil {
		handleError(err)
	}

	server := &http.Server{
		Addr:    config.HttpUri,
		Handler: webAPI,
		// Add timeouts for security
		ReadTimeout:  config.HttpReadTimeout,
		WriteTimeout: config.HttpWriteTimeout,
	}

	go func() {
		log.Info("Starting HTTP Server", map[string]any{"uri": config.HttpUri})
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP Server Failed", err, nil)
		}
	}()

	// Waiting for interrupt signal
	<-ctx.Done()

	// Graceful Shutdown
	startTime = time.Now()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server shutdown failed", err, map[string]any{
			"shutdown_time": time.Since(startTime),
		})
	}

	log.Info("application shutdown completed", map[string]any{
		"shutdown_time": time.Since(startTime),
	})
}
