package dbops

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	types "juancavallotti.com/recipe-types"
)

// ErrSkillNotFound is returned when no skill exists for the given id or name.
var ErrSkillNotFound = errors.New("dbops: skill not found")

// ErrSkillNameTaken is returned when a skill with the given name already exists.
var ErrSkillNameTaken = errors.New("dbops: skill name already exists")

const skillColumns = "id::text, name, description, content, created_at, updated_at"

// ListSkills returns all skills ordered by name.
func (s *Store) ListSkills(ctx context.Context) ([]types.Skill, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT `+skillColumns+`
FROM skills
ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]types.Skill, 0)
	for rows.Next() {
		sk, err := scanSkill(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sk)
	}
	return out, rows.Err()
}

// GetSkill returns a skill by id.
func (s *Store) GetSkill(ctx context.Context, id string) (types.Skill, error) {
	if s.db == nil {
		return types.Skill{}, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return types.Skill{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT `+skillColumns+`
FROM skills
WHERE id = $1`, id)
	sk, err := scanSkill(row)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Skill{}, ErrSkillNotFound
	}
	return sk, err
}

// GetSkillByName returns a skill by name.
func (s *Store) GetSkillByName(ctx context.Context, name string) (types.Skill, error) {
	if s.db == nil {
		return types.Skill{}, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return types.Skill{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT `+skillColumns+`
FROM skills
WHERE name = $1`, name)
	sk, err := scanSkill(row)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Skill{}, ErrSkillNotFound
	}
	return sk, err
}

// CreateSkill inserts a new skill and returns its id. If a skill with the same
// name already exists, returns ErrSkillNameTaken.
func (s *Store) CreateSkill(ctx context.Context, name, description, content string) (string, error) {
	if s.db == nil {
		return "", errNilDB
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var id string
	err := s.db.QueryRowContext(ctx, `
INSERT INTO skills (name, description, content)
VALUES ($1, $2, $3)
RETURNING id::text`, name, description, content).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "skills_name_unique") {
			return "", ErrSkillNameTaken
		}
		return "", err
	}
	return id, nil
}

// UpdateSkill replaces the description and content of a skill by id.
func (s *Store) UpdateSkill(ctx context.Context, id, description, content string) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE skills
SET description = $2, content = $3
WHERE id = $1`, id, description, content)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSkillNotFound
	}
	return nil
}

// DeleteSkill removes a skill by id.
func (s *Store) DeleteSkill(ctx context.Context, id string) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM skills WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSkillNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSkill(r rowScanner) (types.Skill, error) {
	var sk types.Skill
	if err := r.Scan(&sk.ID, &sk.Name, &sk.Description, &sk.Content, &sk.CreatedAt, &sk.UpdatedAt); err != nil {
		return types.Skill{}, err
	}
	return sk, nil
}

// isUniqueViolation reports whether err is a Postgres unique_violation
// (SQLSTATE 23505) whose error message names the given constraint. We avoid a
// hard dependency on the lib/pq error type to keep this package import-light.
func isUniqueViolation(err error, constraint string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, constraint)
}
