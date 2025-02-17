// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: suggest_feature.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createSuggestFeature = `-- name: CreateSuggestFeature :exec
INSERT INTO suggest_feature (id , title, description , username) VALUES ($1, $2 , $3 , $4) RETURNING id, title, description, username, upvote
`

type CreateSuggestFeatureParams struct {
	ID          uuid.UUID
	Title       string
	Description string
	Username    string
}

func (q *Queries) CreateSuggestFeature(ctx context.Context, arg CreateSuggestFeatureParams) error {
	_, err := q.db.ExecContext(ctx, createSuggestFeature,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.Username,
	)
	return err
}

const getSuggestFeature = `-- name: GetSuggestFeature :many
SELECT id, title, description, username, upvote FROM suggest_feature ORDER BY upvote DESC
`

func (q *Queries) GetSuggestFeature(ctx context.Context) ([]SuggestFeature, error) {
	rows, err := q.db.QueryContext(ctx, getSuggestFeature)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SuggestFeature
	for rows.Next() {
		var i SuggestFeature
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Username,
			&i.Upvote,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const suggestFeatureDownvote = `-- name: SuggestFeatureDownvote :exec
UPDATE suggest_feature SET upvote = upvote - 1 WHERE id = $1 RETURNING id, title, description, username, upvote
`

func (q *Queries) SuggestFeatureDownvote(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, suggestFeatureDownvote, id)
	return err
}

const suggestFeatureUpvote = `-- name: SuggestFeatureUpvote :exec
UPDATE suggest_feature SET upvote = upvote + 1 WHERE id = $1 RETURNING id, title, description, username, upvote
`

func (q *Queries) SuggestFeatureUpvote(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, suggestFeatureUpvote, id)
	return err
}
