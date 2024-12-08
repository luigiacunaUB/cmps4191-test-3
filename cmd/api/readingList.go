package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/luigiacunaUB/cmps4191-test-3/internal/data"
	"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
)

func (a *applicationDependencies) AddReadingList(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside AddReadingList")
	//accept the incoming data
	var incomingData struct {
		ReadListName string `json:"name"`
		Books        []int  `json:"book_name"`
		CreatedBy    int    `json:"createdby"`
		Description  string `json:"description"`
		Status       string `json:"status"`
	}
	//check if the data is correctly formed
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
	}
	//assign the incoming data to a ReadingListModel
	list := &data.ReadingList{
		ReadListName: incomingData.ReadListName,
		Books:        incomingData.Books,
		Description:  incomingData.Description,
		Status:       incomingData.Status,
		CreatedBy:    incomingData.CreatedBy,
	}

	//do validation
	v := validator.New()
	data.ValidateReadingList(v, a.ReadingListModel, list)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	logger.Info("Error Here 1")
	//once the validation passes do the push tp DB execution
	ans, err := a.ReadingListModel.AddReadingListToDatabase(*list)
	if err != nil {
		logger.Info("Error Here 2")
		a.serverErrorResponse(w, r, err)
	}
	logger.Info("Error Here 3")

	data := envelope{
		"Reading List ID": ans.ID,
		"Books":           ans.Books,
		"Created By User": ans.CreatedBy,
		"Description":     ans.Description,
		"Status":          ans.Status,
	}

	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}
