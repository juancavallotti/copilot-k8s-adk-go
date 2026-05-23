package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	repo "juancavallotti.com/recipes-repo"
)

func (r Runner) cmdListSkills(ctx context.Context, rp SkillRepo) error {
	skills, err := rp.ListSkills(ctx)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(r.stdout)
	for _, sk := range skills {
		row := struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}{Name: sk.Name, Description: sk.Description}
		if err := enc.Encode(row); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) cmdLoadSkill(ctx context.Context, rp SkillRepo, name string) error {
	sk, err := rp.GetSkillByName(ctx, name)
	if err != nil {
		if errors.Is(err, repo.ErrSkillNotFound) {
			fmt.Fprintf(r.stderr, "load-skill: no skill named %q\n", name)
			return ErrUsage
		}
		return err
	}
	if _, err := fmt.Fprint(r.stdout, sk.Content); err != nil {
		return err
	}
	return nil
}
