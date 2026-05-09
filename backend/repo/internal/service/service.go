package service

import "juancavallotti.com/recipes-repo/internal/dbops"

type Service struct {
	store *dbops.Store
}

func NewService(store *dbops.Store) *Service {
	return &Service{store: store}
}
