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
	router.HandlerFunc(http.MethodPost, "/api/v1/books", a.AddBookHandler)          //add a book
	router.HandlerFunc(http.MethodGet, "/api/v1/books/search", a.SearchFunction)    //search for book based on title/author/genre
	router.HandlerFunc(http.MethodPut, "/api/v1/books/:id", a.UpdateBookHandler)    //Update a book
	router.HandlerFunc(http.MethodDelete, "/api/v1/books/:id", a.DeleteBookHandler) //Delete a book
	router.HandlerFunc(http.MethodGet, "/api/v1/book/:id", a.ListBookHandler)       //list a single book
	router.HandlerFunc(http.MethodGet, "/api/v1/books", a.ListAllHandler)           //list all books

	return a.recoverPanic(a.rateLimit(router))
}
