package models

import (
	"database/sql"
	"errors"
	"time"
)

// Allowed values for the Status field, per spec.
const (
	StatusPublish = "publish"
	StatusDraft   = "draft"
	StatusThrash  = "thrash"
)

// ErrNotFound is returned when a post with the given id doesn't exist.
var ErrNotFound = errors.New("post not found")

// Post mirrors the `posts` table.
type Post struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

// PostInput is the shape accepted from the request body when creating
// or updating a post.
type PostInput struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

// PostRepository talks to the `posts` table.
type PostRepository struct {
	DB *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{DB: db}
}

// Create inserts a new post and returns it with its generated id and timestamps.
func (r *PostRepository) Create(in PostInput) (*Post, error) {
	res, err := r.DB.Exec(
		`INSERT INTO posts (title, content, category, status) VALUES (?, ?, ?, ?)`,
		in.Title, in.Content, in.Category, in.Status,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(id))
}

// List returns a paginated slice of posts ordered by newest first, along
// with the total number of posts in the table (for building pagination meta).
func (r *PostRepository) List(limit, offset int) ([]Post, int, error) {
	rows, err := r.DB.Query(
		`SELECT id, title, content, category, status, created_date, updated_date
		 FROM posts
		 ORDER BY id DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Category, &p.Status, &p.CreatedDate, &p.UpdatedDate); err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&total); err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetByID fetches a single post. Returns ErrNotFound if it doesn't exist.
func (r *PostRepository) GetByID(id int) (*Post, error) {
	var p Post
	err := r.DB.QueryRow(
		`SELECT id, title, content, category, status, created_date, updated_date
		 FROM posts WHERE id = ?`,
		id,
	).Scan(&p.ID, &p.Title, &p.Content, &p.Category, &p.Status, &p.CreatedDate, &p.UpdatedDate)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update overwrites title/content/category/status for an existing post.
func (r *PostRepository) Update(id int, in PostInput) (*Post, error) {
	res, err := r.DB.Exec(
		`UPDATE posts SET title = ?, content = ?, category = ?, status = ? WHERE id = ?`,
		in.Title, in.Content, in.Category, in.Status, id,
	)
	if err != nil {
		return nil, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		// Row might exist but be identical to the update (0 rows affected in
		// MySQL for a no-op update) — check existence explicitly before
		// declaring not found.
		if _, err := r.GetByID(id); err != nil {
			return nil, err
		}
	}

	return r.GetByID(id)
}

// Delete removes a post by id. Returns ErrNotFound if it doesn't exist.
func (r *PostRepository) Delete(id int) error {
	res, err := r.DB.Exec(`DELETE FROM posts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
