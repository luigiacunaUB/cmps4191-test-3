package data

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
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
func (b BookModel) AddBookToDatabase(book Book) (int64, error) {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside AddBookToDatabase")

	// Begin a transaction
	tx, err := b.DB.Begin()
	if err != nil {
		return 0, err
	}

	// Insert the book into the books table
	var bookID int64
	err = tx.QueryRow(
		`INSERT INTO books (title, isbn, publication_date, genre, description, average_rating) 
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		book.Title, book.ISBN, book.PublicationDate, book.Genre, book.Description, book.AverageRating,
	).Scan(&bookID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	logger.Info("Finished Adding book pushing into authors")
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
					return 0, err
				}
			} else {
				tx.Rollback()
				return 0, err
			}
		}
		logger.Info("Finished insert authors")
		// Insert the relationship into the book_authors table
		_, err = tx.Exec(
			`INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`,
			bookID, authorID,
		)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return bookID, nil
}

// --------------------------------------------------------------------------------------------------------------------
func (b BookModel) SearchDatabase(search Book) ([]Book, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside Search Database, starting search")
	//checking if results passed are ok
	logger.Info("Title to be Searched", search.Title)
	logger.Info("Author to be Searched", search.Authors)
	logger.Info("Genre to be Searched", search.Genre)

	query := `
		SELECT 
    b.id, 
    b.title, 
    ARRAY_AGG(a.name) AS authors, 
    b.isbn, 
    b.publication_date, 
    b.genre, 
    b.description, 
    b.average_rating
FROM 
    books b
LEFT JOIN 
    book_authors ba ON b.id = ba.book_id
LEFT JOIN 
    authors a ON ba.author_id = a.id
WHERE 
    ($1 = '' OR b.title ILIKE '%' || $1 || '%') AND
    ($2 = '{}' OR EXISTS (
        SELECT 1 
        FROM authors a_sub 
        WHERE a_sub.id = ba.author_id AND a_sub.name = ANY(COALESCE($2::TEXT[], ARRAY[]::TEXT[]))
    )) AND
    ($3 = '' OR b.genre ILIKE '%' || $3 || '%')
GROUP BY 
    b.id;

		`

	// Execute the query with parameters
	rows, err := b.DB.Query(query, search.Title, pq.Array(search.Authors), search.Genre)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logger.Info("Finished doing queries pushing data to slice")

	// Prepare to collect the results
	var books []Book

	for rows.Next() {
		var book Book
		var authors []string

		// Scan the row into the Book struct
		err := rows.Scan(
			&book.ID,
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
		logger.Info("It Reaches here")
		book.Authors = authors // Assign authors to the struct
		books = append(books, book)
	}

	// Check for any errors that occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
