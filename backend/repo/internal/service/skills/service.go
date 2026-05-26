package skills

import (
	skillops "juancavallotti.com/recipes-repo/internal/dbops/skills"
)

type Service struct {
	store *skillops.Store
}

// NewService wires a skill store into the skill service layer.
func NewService(store *skillops.Store) *Service {
	return &Service{store: store}
}
