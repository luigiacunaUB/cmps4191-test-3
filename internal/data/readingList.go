package data

import (
	"database/sql"

	"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
)

type ReadingListModel struct {
	DB *sql.DB
}

type ReadingList struct {
	ID           int64  `json:"id"`
	ReadListName string `json:"name"`
	Books        []int  `json:"book_name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
}

// --------------------------------------------------------------------------------------------------------------------
func ValidateReadingList(v *validator.Validator, rl ReadingListModel, list *ReadingList) {
	//Check the name of the Reading List
	v.Check(list.ReadListName != "", "Reading List:", "The name must not be empty")
	v.Check(len(list.ReadListName) <= 25, "Title", "Must not be more than 25 bytes long")
	//Book Validation Check
	v.Check(len(list.Books) > 0, "Books", "At least one book ID must be provided")
	for _, bookID := range list.Books {
		v.Check(bookID > 0, "Books", "Book ID must be a positive integer")
	}
	//Descritption Check
	v.Check(list.Description != "", "Description", "Must not be Empty")
	v.Check(len(list.Description) <= 100, "Description", "Must not be more than 100 bytes long")
	//Status Check
	validStatuses := []string{"currently reading", "completed"}
	v.Check(stringInSlice(list.Status, validStatuses), "Status", "Status must be either 'currently reading' or 'completed'")

}

func stringInSlice(value string, list []string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}

func (rl ReadingListModel) AddReadingListToDatabase(list ReadingList) error {
	// Start a database transaction
	tx, err := rl.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			tx.Rollback() // rollback transaction on error
		} else {
			err = tx.Commit() // commit transaction
		}
	}()

	// Insert into `reading_lists` table
	var readingListID int
	query := `
			INSERT INTO reading_lists (name, description, created_by, status)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`
	err = tx.QueryRow(query, list.ReadListName, list.Description, list.Status).Scan(&readingListID)
	if err != nil {
		return err
	}

	// Insert into `reading_list_books` table
	booksQuery := `
			INSERT INTO reading_list_books (reading_list_id, book_id)
			VALUES ($1, $2)
		`
	for _, bookID := range list.Books {
		_, err = tx.Exec(booksQuery, readingListID, bookID)
		if err != nil {
			return err
		}
	}
	return nil
}
