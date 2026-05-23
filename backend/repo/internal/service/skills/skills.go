package skills

import (
	"context"
	"errors"
	"regexp"
	"strings"

	types "juancavallotti.com/recipe-types"
)

var (
	ErrInvalidSkillName        = errors.New("service: invalid skill name")
	ErrInvalidSkillDescription = errors.New("service: invalid skill description")
	ErrInvalidSkillContent     = errors.New("service: invalid skill content")
	ErrInvalidSkillID          = errors.New("service: invalid skill id")
)

const (
	maxSkillNameLen        = 64
	maxSkillDescriptionLen = 512
	maxSkillContentLen     = 64 * 1024
)

var skillNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func (s *Service) ListSkills(ctx context.Context) ([]types.Skill, error) {
	return s.store.ListSkills(ctx)
}

func (s *Service) GetSkill(ctx context.Context, id string) (types.Skill, error) {
	if strings.TrimSpace(id) == "" {
		return types.Skill{}, ErrInvalidSkillID
	}
	return s.store.GetSkill(ctx, id)
}

func (s *Service) GetSkillByName(ctx context.Context, name string) (types.Skill, error) {
	if err := validateSkillName(name); err != nil {
		return types.Skill{}, err
	}
	return s.store.GetSkillByName(ctx, name)
}

func (s *Service) CreateSkill(ctx context.Context, name, description, content string) (string, error) {
	if err := validateSkillName(name); err != nil {
		return "", err
	}
	if err := validateSkillDescription(description); err != nil {
		return "", err
	}
	if err := validateSkillContent(content); err != nil {
		return "", err
	}
	return s.store.CreateSkill(ctx, name, description, content)
}

func (s *Service) UpdateSkill(ctx context.Context, id, description, content string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidSkillID
	}
	if err := validateSkillDescription(description); err != nil {
		return err
	}
	if err := validateSkillContent(content); err != nil {
		return err
	}
	return s.store.UpdateSkill(ctx, id, description, content)
}

func (s *Service) DeleteSkill(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidSkillID
	}
	return s.store.DeleteSkill(ctx, id)
}

func validateSkillName(name string) error {
	if name == "" || len(name) > maxSkillNameLen || !skillNameRe.MatchString(name) {
		return ErrInvalidSkillName
	}
	return nil
}

func validateSkillDescription(description string) error {
	if strings.TrimSpace(description) == "" || len(description) > maxSkillDescriptionLen {
		return ErrInvalidSkillDescription
	}
	return nil
}

func validateSkillContent(content string) error {
	if strings.TrimSpace(content) == "" || len(content) > maxSkillContentLen {
		return ErrInvalidSkillContent
	}
	return nil
}
