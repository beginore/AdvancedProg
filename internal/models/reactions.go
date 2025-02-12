package models

import (
	"database/sql"
)

// ReactionModel wrapper for database connection
type ReactionModel struct {
	DB *sql.DB
}

// Check if a user has liked a post
func (m *ReactionModel) IsLiked(postID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)`
	err := m.DB.QueryRow(stmt, postID, userID).Scan(&exists)
	return exists, err
}

// Check if a user has disliked a post
func (m *ReactionModel) IsDisliked(postID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM post_dislikes WHERE post_id = $1 AND user_id = $2)`
	err := m.DB.QueryRow(stmt, postID, userID).Scan(&exists)
	return exists, err
}

// Like a post
func (m *ReactionModel) LikePost(postID, userID int) error {
	liked, err := m.IsLiked(postID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.IsDisliked(postID, userID)
	if err != nil {
		return err
	}

	if liked {
		return m.RemoveLikePost(postID, userID)
	}
	if disliked {
		if err := m.RemoveDislikePost(postID, userID); err != nil {
			return err
		}
	}

	// Insert the like
	stmt := `INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2)`
	_, err = m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	// Update the likes counter
	stmt2 := `UPDATE posts SET likes = likes + 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

// Dislike a post
func (m *ReactionModel) DislikePost(postID, userID int) error {
	liked, err := m.IsLiked(postID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.IsDisliked(postID, userID)
	if err != nil {
		return err
	}

	if disliked {
		return m.RemoveDislikePost(postID, userID)
	}
	if liked {
		if err := m.RemoveLikePost(postID, userID); err != nil {
			return err
		}
	}

	// Insert the dislike
	stmt := `INSERT INTO post_dislikes (post_id, user_id) VALUES ($1, $2)`
	_, err = m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	// Update the dislikes counter
	stmt2 := `UPDATE posts SET dislikes = dislikes + 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

// Remove a like from a post
func (m *ReactionModel) RemoveLikePost(postID, userID int) error {
	// Delete the like
	stmt := `DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2`
	_, err := m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	// Decrement the likes counter
	stmt2 := `UPDATE posts SET likes = likes - 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

// Remove a dislike from a post
func (m *ReactionModel) RemoveDislikePost(postID, userID int) error {
	// Delete the dislike
	stmt := `DELETE FROM post_dislikes WHERE post_id = $1 AND user_id = $2`
	_, err := m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	// Decrement the dislikes counter
	stmt2 := `UPDATE posts SET dislikes = dislikes - 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

// Check if a user has liked a comment
func (m *ReactionModel) IsCommentLiked(commentID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM comment_likes WHERE comment_id = $1 AND user_id = $2)`
	err := m.DB.QueryRow(stmt, commentID, userID).Scan(&exists)
	return exists, err
}

// Check if a user has disliked a comment
func (m *ReactionModel) IsCommentDisliked(commentID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM comment_dislikes WHERE comment_id = $1 AND user_id = $2)`
	err := m.DB.QueryRow(stmt, commentID, userID).Scan(&exists)
	return exists, err
}

// Like a comment
func (m *ReactionModel) LikeComment(commentID, userID int) error {
	liked, err := m.IsCommentLiked(commentID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.IsCommentDisliked(commentID, userID)
	if err != nil {
		return err
	}

	if liked {
		return m.RemoveLikeComment(commentID, userID)
	}
	if disliked {
		if err := m.RemoveDislikeComment(commentID, userID); err != nil {
			return err
		}
	}

	// Insert the like
	stmt := `INSERT INTO comment_likes (comment_id, user_id) VALUES ($1, $2)`
	_, err = m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	// Update the likes counter
	stmt2 := `UPDATE comments SET likes = likes + 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}

// Dislike a comment
func (m *ReactionModel) DislikeComment(commentID, userID int) error {
	liked, err := m.IsCommentLiked(commentID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.IsCommentDisliked(commentID, userID)
	if err != nil {
		return err
	}

	if disliked {
		return m.RemoveDislikeComment(commentID, userID)
	}
	if liked {
		if err := m.RemoveLikeComment(commentID, userID); err != nil {
			return err
		}
	}

	// Insert the dislike
	stmt := `INSERT INTO comment_dislikes (comment_id, user_id) VALUES ($1, $2)`
	_, err = m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	// Update the dislikes counter
	stmt2 := `UPDATE comments SET dislikes = dislikes + 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}

// Remove a like from a comment
func (m *ReactionModel) RemoveLikeComment(commentID, userID int) error {
	// Delete the like
	stmt := `DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2`
	_, err := m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	// Decrement the likes counter
	stmt2 := `UPDATE comments SET likes = likes - 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}

// Remove a dislike from a comment
func (m *ReactionModel) RemoveDislikeComment(commentID, userID int) error {
	// Delete the dislike
	stmt := `DELETE FROM comment_dislikes WHERE comment_id = $1 AND user_id = $2`
	_, err := m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	// Decrement the dislikes counter
	stmt2 := `UPDATE comments SET dislikes = dislikes - 1 WHERE id = $1`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}
