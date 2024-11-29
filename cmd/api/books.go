package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/luigiacunaUB/cmps4191-test-3/internal/data"
	"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
)

// ------------------------------------------------------------------------------------------
func (a *applicationDependencies) AddBookHandler(w http.ResponseWriter, r *http.Request) {
	//set the data coming from the curl command
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside AddBookHandler 1")
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
	var preResults []data.Book
	titleSearch := data.Book{Title: book.Title, Authors: book.Authors, Genre: book.Genre}
	logger.Info("Just before SearchDatabase Title Search")
	preResults, err = a.BookModel.SearchDatabase(titleSearch)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//extracting only the title portion of preResults
	titles := make([]string, len(preResults))
	for i, book := range preResults {
		titles[i] = book.Title
	}
	logger.Info("Results Titles", slog.Any("titles", titles))

	//If title is found, return error
	for _, title := range titles {
		if title == incomingData.Title {
			a.errorResponseJSON(w, r, http.StatusConflict, "A book with this title already exists")
			return
		}
	}
	//if no book is found go ahead with addition
	logger.Info("Just Before AddBookToDatabase")
	bookID, err := a.BookModel.AddBookToDatabase(*book)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
	book.ID = bookID

	//create the headers
	fmt.Fprintf(w, "%+v\n", incomingData)
	headers := make(http.Header)
	//making the apporiate header for GET /api/v1/books/api
	headers.Set("Location", fmt.Sprintf("/api/v1/books/%d", book.ID))

	data := envelope{
		"Location: /api/v1/books/": book.ID,
		"Title":                    book.Title,
		"Authors":                  book.Authors,
		"ISBN":                     book.ISBN,
		"Publication Date":         book.PublicationDate,
		"Genre":                    book.Genre,
		"Description":              book.Description,
		"Average Rating":           book.AverageRating,
	}

	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

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
		return
	}

	search := &data.Book{
		Title:   incomingData.Title,
		Authors: incomingData.Authors,
		Genre:   incomingData.Genre,
	}

	var results []data.Book
	//pass search to SearchDatabase
	results, err = a.BookModel.SearchDatabase(*search)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Return the search results
	err = a.writeJSON(w, http.StatusOK, envelope{"results": results}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
