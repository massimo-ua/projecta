package main

import (
	"context"
	"errors"
	assetapp "gitlab.com/massimo-ua/projecta/internal/asset/app"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	investmentapp "gitlab.com/massimo-ua/projecta/internal/investment/app"
	"gitlab.com/massimo-ua/projecta/internal/people"
	projectapp "gitlab.com/massimo-ua/projecta/internal/project/app"
	"gitlab.com/massimo-ua/projecta/internal/statement"
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
)

var Version = "dev"

func createErrorHandler(logger core.Logger) func(err error) {
	return func(err error) {
		logger.Error("failed to start the application", err, nil)
		os.Exit(1)
	}
}

type appHandlerDeps struct {
	rateProvider currency.CurrencyRateProvider
	googleAuth   core.ThirdPartyAuth
	logger       core.Logger
}

func setupAppHandler(config *core.AppConfig, db *dal.PgDbConnection, log ...core.Logger) (http.Handler, error) {
	var l core.Logger
	if len(log) > 0 {
		l = log[0]
	}
	return setupAppHandlerWithDeps(config, db, appHandlerDeps{logger: l})
}

func setupAppHandlerWithDeps(config *core.AppConfig, db *dal.PgDbConnection, deps appHandlerDeps) (http.Handler, error) {
	if config == nil {
		return nil, exceptions.NewInternalException("missing app config", nil)
	}
	if db == nil {
		return nil, exceptions.NewInternalException("missing db connection", nil)
	}

	peopleRepository := dal.NewPgPeopleRepository(db)
	hasher := crypto.NewBcryptHasher(0)
	googleAuth := deps.googleAuth
	if googleAuth == nil {
		var err error
		googleAuth, err = crypto.NewGoogleAuthProvider(crypto.GoogleAuthConfig{
			ClientID:      config.GoogleClientID,
			ClientSecret:  config.GoogleClientSecret,
			CertsCacheTTL: config.GoogleCertTTL,
		})
		if err != nil {
			return nil, err
		}
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

	userService := people.NewUserService(db, peopleRepository, hasher)
	projectRepository := dal.NewPgProjectRepository(db)
	assetRepository := dal.NewPgAssetRepository(db)
	peopleService := &projectPeopleAdapter{repo: peopleRepository}

	rateProvider := deps.rateProvider
	if rateProvider == nil {
		rateProvider = currency.NewNBUCurrencyRateProvider(currency.NBUCurrencyRateProviderOptions{
			SupportedCurrencies: []string{"UAH", "USD", "EUR", "PLN"},
			CacheTTL:            30 * time.Minute,
		})
	}

	moneyConverter := &currencyConverterAdapter{provider: rateProvider}

	investmentRepository := dal.NewPgInvestmentRepository(db)
	investmentApp := investmentapp.NewApplication(
		db,
		investmentRepository,
		investmentRepository,
		&investmentAssetReaderAdapter{assets: assetRepository},
		&investmentProjectReaderAdapter{projects: projectRepository},
		&investmentContributorReaderAdapter{people: peopleService},
		moneyConverter,
		deps.logger,
	)

	assetApp := assetapp.NewApplication(
		db,
		assetRepository,
		&assetOwnerReaderAdapter{people: peopleService},
		&assetProjectReaderAdapter{projects: projectRepository, rateProvider: rateProvider},
		&investmentSourceAdapter{app: investmentApp},
		moneyConverter,
		deps.logger,
	)

	totalsCalculator := &projectTotalsCalculatorAdapter{
		assets:       assetApp,
		investments:  investmentApp,
		rateProvider: rateProvider,
	}
	projectApp := projectapp.NewApplication(projectRepository, peopleService, totalsCalculator, moneyConverter, deps.logger)

	statementService := statement.NewService(
		&projectValidator{app: projectApp},
		&investmentTransactionFinder{app: investmentApp},
	)

	return web.MakeHTTPHandler(
		userService,
		tokenProvider,
		authService,
		invitationService,
		projectApp,
		statementService,
		assetApp,
		investmentApp,
	)
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

	webAPI, err := setupAppHandler(config, db, log)

	if err != nil {
		handleError(err)
	}

	server := &http.Server{
		Addr:    config.HttpUri,
		Handler: webAPI,
		// Add timeouts for security
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       config.HttpReadTimeout,
		WriteTimeout:      config.HttpWriteTimeout,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Info("Starting HTTP Server", map[string]any{"uri": config.HttpUri, "version": Version})
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
