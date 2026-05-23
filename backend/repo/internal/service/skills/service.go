package skills

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

type store interface {
	ListSkills(ctx context.Context) ([]types.Skill, error)
	GetSkill(ctx context.Context, id string) (types.Skill, error)
	GetSkillByName(ctx context.Context, name string) (types.Skill, error)
	CreateSkill(ctx context.Context, name, description, content string) (string, error)
	UpdateSkill(ctx context.Context, id, description, content string) error
	DeleteSkill(ctx context.Context, id string) error
}

type Service struct {
	store store
}

// NewService wires a skill store into the skill service layer.
func NewService(store store) *Service {
	return &Service{store: store}
}
