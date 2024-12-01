package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/luigiacunaUB/cmps4191-test-3/internal/validator"
)

type ReviewModel struct {
	DB *sql.DB
}

type Review struct {
	ID     int64  `json:"reviewid"`
	BookID int64  `json:"bookid"`
	UserID int64  `json:"userid"`
	Review string `json:"review"`
	Rating int64  `json:"rating"`
}

func ValidateReview(v *validator.Validator, b BookModel, r ReviewModel, review *Review) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside ValidateReview")
	logger.Info("BookID being sent: ", review.BookID)
	//check firstly if the book even exist
	v.Check(review.BookID >= 1, "BookID", "BookID cannot be less than 1 this one")
	ans, err := b.SearchBookByID(review.BookID)
	if err != nil {
		return
	}
	logger.Info("Answer: ", ans)
	v.Check(ans, "BookID", "BookID does not exist")
	//check if the review is less than 100 bytes long
	v.Check(review.Review != "", "Review", "Must not be Empty")
	v.Check(len(review.Review) <= 100, "Review", "Must not be more than 100 bytes long")
	//Check Ratings is between 1 and 5
	v.Check(review.Rating >= 1 && review.Rating <= 5, "Ratings", "Ratings must between 1 and 5")
}

func (r ReviewModel) AddBookReview(review Review) (Review, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside AddBooReviewHandler")
	logger.Info("BookID sent SQL:", review.BookID)
	logger.Info("UserID sent SQL:", review.UserID)
	logger.Info("Review sent SQL:", review.Review)
	logger.Info("Rating sent SQL:", review.Rating)

	// SQL query to insert the review
	query := `
		INSERT INTO reviews (book_id, user_id, rating, review)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	// Arguments for the query
	args := []any{
		review.BookID,
		review.UserID,
		review.Rating,
		review.Review,
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query
	var createdReview Review
	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&createdReview.ID)
	if err != nil {
		logger.Error("Error inserting review", "error", err)
		return Review{}, err
	}

	// Populate the rest of the returned review object
	createdReview.BookID = review.BookID
	createdReview.UserID = review.UserID
	createdReview.Rating = review.Rating
	createdReview.Review = review.Review

	logger.Info("Successfully added review", "reviewID", createdReview.ID)
	return createdReview, nil

}

// --------------------------------------------------------------------------------------------------------------------------------------
func (r ReviewModel) UpdateReview(review Review) (Review, error) {
	//Check if the user actually made the review
	rcheck := r.CheckIfReviewExistForUser(review.BookID, review.UserID)
	if rcheck {
		return review, fmt.Errorf("user has already reviewed this book")
	}

	//get the review id number to modify
	var reviewID int64
	query := `SELECT id FROM reviews WHERE user_id=$1 AND book_id=$2`
	err := r.DB.QueryRow(query, review.UserID, review.BookID).Scan(&reviewID)
	if err != nil {
		if err == sql.ErrNoRows {
			return review, fmt.Errorf("review not found")
		}
		return review, fmt.Errorf("failed to retrieve review ID: %v", err)
	}

	// Update the review in the database
	updateQuery := `UPDATE reviews SET rating = $1, review = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err = r.DB.Exec(updateQuery, review.Rating, review.Review, reviewID)
	if err != nil {
		return review, fmt.Errorf("failed to update review: %v", err)
	}

	// Return the updated review
	return review, nil
}

func (r ReviewModel) CheckIfReviewExistForUser(bookid int64, userid int64) bool {
	query := `SELECT book_id,user_id FROM reviews WHERE book_id=$1 AND user_id=$2`

	var existingReview struct {
		BookID int64
		UserID int64
	}

	// Execute the query to check if a review exists for the given book and user
	err := r.DB.QueryRow(query, bookid, userid).Scan(&existingReview.BookID, &existingReview.UserID)

	if err != nil {
		if err == sql.ErrNoRows {
			// No review found, return false
			return false
		}
		// Handle any other errors
		log.Println("Error checking for existing review:", err)
		return false
	}

	// A review exists
	return true
}
