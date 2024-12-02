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

func ValidateReview(v *validator.Validator, r ReviewModel, review *Review) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside ValidateReview")
	//check if the review is less than 100 bytes long
	v.Check(review.Review != "", "Review", "Must not be Empty")
	v.Check(len(review.Review) <= 100, "Review", "Must not be more than 100 bytes long")
	//Check Ratings is between 1 and 5
	v.Check(review.Rating >= 1 && review.Rating <= 5, "Ratings", "Ratings must between 1 and 5")
}

func ValidateReviewIDOnly(v *validator.Validator, r ReviewModel, review *Review) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("ReviewID being sent: ", review.ID)
	v.Check(review.ID >= 1, "ReviewID", "ReviewID cannot be less than 1 this one")
}

func (r ReviewModel) AddBookReview(review Review) (Review, error) {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Inside AddBooReviewHandler")
	logger.Info("UserID sent SQL:", review.UserID)
	logger.Info("BookID sent SQL: ", review.BookID)
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
	//workflow
	//parameters recieved from review: reviewID, updated review, updated rating
	//update the review

	//EXECUTION
	updateQuery := `
		UPDATE reviews 
		SET rating = $1, review = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3 
		RETURNING id, book_id, rating, review
	`
	// Prepare variables to store the updated data
	var updatedReview Review

	// Execute the query and scan the updated values
	err := r.DB.QueryRow(updateQuery, review.Rating, review.Review, review.ID).
		Scan(&updatedReview.ID, &updatedReview.BookID, &updatedReview.Rating, &updatedReview.Review)

	if err != nil {
		return Review{}, fmt.Errorf("failed to update review: %w", err)
	}

	// Return the updated review
	return updatedReview, nil
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

func (b BookModel) SearchReviewByID(id int64) (bool, error) {
	query := `SELECT id FROM review WHERE id=$1;`
	var foundID int
	err := b.DB.QueryRow(query, id).Scan(&foundID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Review with the given ID does not exist
			return false, nil
		}
		// An unexpected error occurred
		return false, err
	}

	// Review exists
	return true, nil
}
