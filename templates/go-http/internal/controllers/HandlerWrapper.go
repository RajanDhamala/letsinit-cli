package controller

import (
	sqlc "http-server/db/sqlc"
)

type Controller struct {
	queries sqlc.Querier
}

func NewController(queries sqlc.Querier) *Controller {
	return &Controller{
		queries: queries,
	}
}
