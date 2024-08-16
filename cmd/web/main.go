package main

import (
	"fmt"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/crypto"
	"gitlab.com/massimo-ua/projecta/pkg/dal"
	"gitlab.com/massimo-ua/projecta/pkg/web"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	TokenTTL                  = 300
	GoogleCertCacheSecondsTTL = 24 * 60 * 60
)

func main() {
	pool, err := dal.Connect(os.Getenv("DB_URI"))

	if err != nil {
		log.Fatalln(err)
	}

	os.Stdout.Write([]byte("Connected to the database\n"))

	//brk, err := broker.NewAMQPBroker(os.Getenv("AMQP_URI"))
	//
	//if err != nil {
	//	log.Fatal(err)
	//}

	defer func() {
		//brk.Close()
		pool.Close()
	}()

	peopleRepository := dal.NewPgPeopleRepository(pool)
	hasher := crypto.NewBcryptHasher(0)
	googleAuth := crypto.NewGoogleAuthProvider(
		os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleCertCacheSecondsTTL,
	)
	tokenProvider := crypto.NewJwtTokenProvider(
		os.Getenv("JWT_SECRET"),
		TokenTTL,
		hasher,
	)
	authService := people.NewAuthService(
		peopleRepository,
		tokenProvider,
		hasher,
		googleAuth,
	)

	customerService := people.NewCustomerService(
		peopleRepository,
		hasher)

	projectRepository := dal.NewPgProjectaProjectRepository(pool)
	categoryRepository := dal.NewPgProjectaCategoryRepository(pool)
	typeRepository := dal.NewPgProjectaCostTypeRepository(pool)
	expenseRepository := dal.NewPgProjectaPaymentRepository(pool)
	peopleService := projecta.NewPeopleService(peopleRepository)
	projectService := projecta.NewProjectService(projectRepository, peopleService)
	categoryService := projecta.NewCategoryService(categoryRepository, projectService)
	typeService := projecta.NewTypeService(typeRepository, categoryRepository, projectRepository)
	expenseService := projecta.NewPaymentService(
		expenseRepository,
		typeRepository,
		projectRepository,
		peopleService,
	)

	webAPI, err := web.MakeHTTPHandler(
		customerService,
		tokenProvider,
		authService,
		projectService,
		categoryService,
		typeService,
		expenseService,
	)

	if err != nil {
		log.Fatal(err)
	}

	uri := os.Getenv("HTTP_URI")
	httpListener, err := net.Listen("tcp", uri)
	if err != nil {
		os.Stderr.Write([]byte(fmt.Sprintf("failed to initialize HTTP listen: %s", err.Error())))
	}

	os.Stdout.Write([]byte(fmt.Sprintf("HTTP server is listening on %s\n", uri)))
	err = http.Serve(httpListener, webAPI)

	if err != nil {
		os.Stderr.Write([]byte(fmt.Sprintf("failed to serve HTTP: %s", err.Error())))
	}
}
