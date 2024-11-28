package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
)

type BookModel struct {
	DB *sql.DB
}

type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Authors         []string  `json:"authors"`
	ISBN            string    `json:"isbn"`
	PublicationDate time.Time `json:"publication_date"`
	Genre           string    `json:"genre"`
	Description     string    `json:"description"`
	AverageRating   float64   `json:"average_rating"`
}

// -----------------------------------------------------------------------------------------------------------------
func ValidateBook(v *validator.Validator, b BookModel, book *Book) {
	//Title Checks
	v.Check(book.Title != "", "Title", "Must not be Empty")
	v.Check(len(book.Title) <= 25, "Title", "Must not be more than 25 bytes long")

	//Author Checks
	v.Check(len(book.Authors) > 0, "Authors", "Atleast one author must be provided")
	for _, author := range book.Authors {
		v.Check(author != "", "Authors", "Author's name must not be empty")
		v.Check(len(author) <= 25, "Authors", "Author's name must not be more than 25 bytes")
	}

	//ISBN length -> following ISBN system 10 digits before 1 Jan 2007, 13 digits after
	v.Check(book.ISBN != "", "ISBN", "Cannot be empty")
	v.Check(len(book.ISBN) == 10 || len(book.ISBN) == 13, "ISBN", "Must be a 10 or 13 Digit Code")

	//Publication Date Checks, how does it parse time??
	v.Check(!book.PublicationDate.IsZero(), "Publication Date", "Must be a valid date")
	v.Check(book.PublicationDate.Before(time.Now()), "Publication Date", "Must not be set in the future")

	//Genre Checks
	v.Check(book.Genre != "", "Genre", "Must not be Empty")
	v.Check(len(book.Genre) <= 25, "Genre", "Must not be more than 25 bytes long")

	//Description Checks
	v.Check(book.Description != "", "Description", "Must not be Empty")
	v.Check(len(book.Genre) <= 100, "Description", "Must not be more than 100 bytes long")

	//Rating Check
	v.Check(book.AverageRating >= 1 && book.AverageRating <= 5, "Ratings", "Ratings must between 1 and 5")
}

// ---------------------------------------------------------------------------------------------------------------------
func (b BookModel) AddBookToDatabase(book Book) error {

	// Begin a transaction
	tx, err := b.DB.Begin()
	if err != nil {
		return err
	}

	// Insert the book into the books table
	var bookID int
	err = tx.QueryRow(
		`INSERT INTO books (title, isbn, publication_date, genre, description, average_rating) 
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		book.Title, book.ISBN, book.PublicationDate, book.Genre, book.Description, book.AverageRating,
	).Scan(&bookID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert authors and the book-author relationship
	for _, author := range book.Authors {
		var authorID int
		// Check if the author already exists
		err = tx.QueryRow(`SELECT id FROM authors WHERE name = $1`, author).Scan(&authorID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Insert the author if they don't exist
				err = tx.QueryRow(
					`INSERT INTO authors (name) VALUES ($1) RETURNING id`,
					author,
				).Scan(&authorID)
				if err != nil {
					tx.Rollback()
					return err
				}
			} else {
				tx.Rollback()
				return err
			}
		}

		// Insert the relationship into the book_authors table
		_, err = tx.Exec(
			`INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`,
			bookID, authorID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// --------------------------------------------------------------------------------------------------------------------
func (b BookModel) SearchDatabase(search Book) ([]Book, error) {
	query := `
		SELECT id, title, authors, isbn, publication_date, genre, description, average_rating
		FROM books
		WHERE 
			($1 = '' OR title ILIKE '%' || $1 || '%') AND
			($2 = '{}' OR authors && $2) AND
			($3 = '' OR genre ILIKE '%' || $3 || '%')
		`

	// Execute the query with parameters
	rows, err := b.DB.Query(query, search.Title, pq.Array(search.Authors), search.Genre)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Prepare to collect the results
	var books []Book

	for rows.Next() {
		var book Book
		var authors []string

		// Scan the row into the Book struct
		err := rows.Scan(
			&book.Title,
			pq.Array(&authors),
			&book.ISBN,
			&book.PublicationDate,
			&book.Genre,
			&book.Description,
			&book.AverageRating,
		)
		if err != nil {
			return nil, err
		}

		book.Authors = authors // Assign authors to the struct
		books = append(books, book)
	}

	// Check for any errors that occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
