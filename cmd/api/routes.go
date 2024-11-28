package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	//setup a new router
	router := httprouter.New()

	//errors
	//404
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	//405
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	//routes
	//basic
	router.HandlerFunc(http.MethodGet, "/", a.Index)                            //root page
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthCheckHandler) //healthcheck
	//-------------------------------------BOOKS------------------------------------------------
	router.HandlerFunc(http.MethodPost, "/api/v1/books", a.AddBookHandler)

	return a.recoverPanic(a.rateLimit(router))
}
