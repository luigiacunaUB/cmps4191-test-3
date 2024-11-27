package main

import (
	"net/http"
	"time"

	"github.com/luigiacunaUB/cmps4191-test-3/internal/data"
	"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
	//"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
)

// ------------------------------------------------------------------------------------------
func (a *applicationDependencies) AddBookHandler(w http.ResponseWriter, r *http.Request) {
	//set the data coming from the curl command
	var incomingData struct {
		Title           string    `json:"title"`
		Authors         []string  `json:"authors"`
		ISBN            string    `json:"isbn"`
		PublicationDate time.Time `json:"publication_date"`
		Genre           string    `json:"genre"`
		Description     string    `json:"description"`
		AverageRating   float64   `json:"average_rating"`
	}

	//read the data to see JSON is properly formed
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
	}
	book := &data.Book{
		Title:           incomingData.Title,
		Authors:         incomingData.Authors,
		ISBN:            incomingData.ISBN,
		PublicationDate: incomingData.PublicationDate,
		Genre:           incomingData.Genre,
		Description:     incomingData.Description,
		AverageRating:   incomingData.AverageRating,
	}
	//call the validator to verify all fields match their specs
	v := validator.New()
	data.ValidateBook(v, a.BookModel, book)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	//search for title if it exist to prevent duplication
	//need to implement search function and use to check title, if empty proceed, if not stop add

}

// ----------------------------------------------------------------------------------------------------
func (a *applicationDependencies) SearchFunction(w http.ResponseWriter, r *http.Request) {
	//search for title,author,genre
	//setting the incoming data for required search
	var incomingData struct {
		Title   string   `json:"title"`
		Authors []string `json:"authors"`
		Genre   string   `json:"genre"`
	}
	//read the data to see JSON is properly formed
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
	}

	search := &data.Book{
		Title:   incomingData.Title,
		Authors: incomingData.Authors,
		Genre:   incomingData.Genre,
	}
	//pass search to SearchDatabase
	err = a.BookModel.SearchDatabase(*search)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
