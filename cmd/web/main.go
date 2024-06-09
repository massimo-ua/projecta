package main

import (
    "gitlab.com/massimo-ua/projecta/internal/people"
    "gitlab.com/massimo-ua/projecta/internal/projecta"
    "gitlab.com/massimo-ua/projecta/pkg/broker"
    "gitlab.com/massimo-ua/projecta/pkg/crypto"
    "gitlab.com/massimo-ua/projecta/pkg/dal"
    "gitlab.com/massimo-ua/projecta/pkg/web"
    "log"
    "net"
    "net/http"
    "os"
)

const TokenTTL = 3600

func main() {
    pool, err := dal.Connect(os.Getenv("DB_URI"))

    if err != nil {
        log.Fatalln(err)
    }

    brk, err := broker.NewAMQPBroker(os.Getenv("AMQP_URI"))

    if err != nil {
        log.Fatal(err)
    }

    defer func() {
        brk.Close()
        pool.Close()
    }()

    peopleRepository := dal.NewPgPeopleRepository(pool)
    hasher := crypto.NewBcryptHasher(0)
    tokenProvider := crypto.NewJwtTokenProvider(
        os.Getenv("JWT_SECRET"),
        TokenTTL,
        hasher,
    )
    authService := people.NewAuthService(
        peopleRepository,
        tokenProvider,
        hasher,
    )

    customerService := people.NewCustomerService(
        peopleRepository,
        hasher)

    projectRepository := dal.NewPgProjectaProjectRepository(pool)
    categoryRepository := dal.NewPgProjectaCategoryRepository(pool)
    typeRepository := dal.NewPgProjectaCostTypeRepository(pool)
    expenseRepository := dal.NewPgProjectaExpenseRepository(pool)
    peopleService := projecta.NewPeopleService(peopleRepository)
    projectService := projecta.NewProjectService(projectRepository, peopleService)
    categoryService := projecta.NewCategoryService(categoryRepository, projectService)
    typeService := projecta.NewTypeService(typeRepository)
    expenseService := projecta.NewExpenseService(
        expenseRepository,
        categoryRepository,
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

    httpListener, err := net.Listen("tcp", os.Getenv("HTTP_URI"))
    if err != nil {
        log.Fatalf("failed to initialize HTTP listen: %s", err.Error())
    }

    err = http.Serve(httpListener, webAPI)

    if err != nil {
        log.Fatalf("failed to serve HTTP: %s", err.Error())
    }
}
