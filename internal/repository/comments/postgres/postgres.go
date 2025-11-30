package postgres

import (
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type CommentsRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewCommentsRepository(db *dbpg.DB, retries retry.Strategy) *CommentsRepository {
	return &CommentsRepository{
		db:      db,
		retries: retries,
	}
}
