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
	//-------------------------------------BOOKS--------------------------------------------------------------------------------------------------------------------------------
	router.HandlerFunc(http.MethodPost, "/api/v1/books", a.requirePermission("books:write", a.AddBookHandler))          //add a book
	router.HandlerFunc(http.MethodGet, "/api/v1/books/search", a.requirePermission("books:read", a.SearchFunction))     //search for book based on title/author/genre
	router.HandlerFunc(http.MethodPut, "/api/v1/books/:id", a.requirePermission("books:write", a.UpdateBookHandler))    //Update a book
	router.HandlerFunc(http.MethodDelete, "/api/v1/books/:id", a.requirePermission("books:write", a.DeleteBookHandler)) //Delete a book
	router.HandlerFunc(http.MethodGet, "/api/v1/book/:id", a.requirePermission("books:read", a.ListBookHandler))        //list a single book
	router.HandlerFunc(http.MethodGet, "/api/v1/books", a.requirePermission("books:read", a.ListAllHandler))            //list all books
	//--------------------------------------REVIEWS-----------------------------------------------------------------------------------------------------------------------------
	router.HandlerFunc(http.MethodPost, "/api/v1/books/:id/reviews", a.requirePermission("books:write", a.AddBookReviewHandler)) //add a review
	router.HandlerFunc(http.MethodPut, "/api//v1/reviews/:id", a.requirePermission("books:write", a.UpdateBookReviewHandler))    //update review
	//--------------------------------------USERS-------------------------------------------------------------------------------------------------------------------------------
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)                              //register a user
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)                     //activate a user
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler) //authenticate token
	router.HandlerFunc(http.MethodPut, "/api/v1/reviews/:id", a.UpdateBookReviewHandler)                 //update a review
	return a.recoverPanic(a.rateLimit(a.authenticate(router)))
}
