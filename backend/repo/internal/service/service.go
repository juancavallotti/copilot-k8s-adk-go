package service

import "juancavallotti.com/recipes-repo/internal/db"

type Service struct {
	store *db.Store
}

func NewService(store *db.Store) *Service {
	return &Service{store: store}
}
