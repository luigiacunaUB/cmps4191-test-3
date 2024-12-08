package main

import (
	"net/http"

	"github.com/luigiacunaUB/cmps4191-test-3/internal/data"
	"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
)

func (a *applicationDependencies) AddReadingList(w http.ResponseWriter, r *http.Request) {
	//accept the incoming data
	var incomingData struct {
		ReadListName string `json:"name"`
		Books        []int  `json:"book_name"`
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
	}

	//do validation
	v := validator.New()
	data.ValidateReadingList(v, a.ReadingListModel, list)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	//once the validation passes do the push tp DB execution

}
