package skills

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	skillops "juancavallotti.com/recipes-repo/internal/dbops/skills"
)

const testUUID = "550e8400-e29b-41d4-a716-446655440000"

func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return NewService(skillops.NewStore(db)), mock, func() { db.Close() }
}

func skillRows() *sqlmock.Rows {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	return sqlmock.NewRows([]string{"id", "name", "description", "content", "created_at", "updated_at"}).
		AddRow(testUUID, "prep", "Prepare ingredients", "Instructions", now, now)
}

func TestService_ListSkills_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectQuery("FROM skills").WillReturnRows(skillRows())

	got, err := s.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "prep" {
		t.Fatalf("got = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_GetSkill_ValidatesID(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	_, err := s.GetSkill(context.Background(), " ")
	if !errors.Is(err, ErrInvalidSkillID) {
		t.Fatalf("err = %v, want ErrInvalidSkillID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_GetSkill_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectQuery("FROM skills").WithArgs(testUUID).WillReturnRows(skillRows())

	got, err := s.GetSkill(context.Background(), testUUID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != testUUID {
		t.Fatalf("got = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_GetSkillByName_ValidatesName(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"", "Upper", "-bad", "bad-", "has spaces"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s, mock, cleanup := newMockService(t)
			defer cleanup()

			_, err := s.GetSkillByName(context.Background(), name)
			if !errors.Is(err, ErrInvalidSkillName) {
				t.Fatalf("err = %v, want ErrInvalidSkillName", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestService_GetSkillByName_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectQuery("FROM skills").WithArgs("prep").WillReturnRows(skillRows())

	got, err := s.GetSkillByName(context.Background(), "prep")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "prep" {
		t.Fatalf("got = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_CreateSkill_ValidatesInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		skillName   string
		description string
		content     string
		wantErr     error
	}{
		{"bad name", "Bad Name", "description", "content", ErrInvalidSkillName},
		{"empty description", "prep", " ", "content", ErrInvalidSkillDescription},
		{"empty content", "prep", "description", " ", ErrInvalidSkillContent},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, mock, cleanup := newMockService(t)
			defer cleanup()

			_, err := s.CreateSkill(context.Background(), tt.skillName, tt.description, tt.content)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestService_CreateSkill_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectQuery("INSERT INTO skills").
		WithArgs("prep", "Prepare ingredients", "Do the prep").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUUID))

	id, err := s.CreateSkill(context.Background(), "prep", "Prepare ingredients", "Do the prep")
	if err != nil {
		t.Fatal(err)
	}
	if id != testUUID {
		t.Fatalf("id = %q, want %q", id, testUUID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_UpdateSkill_ValidatesInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		id          string
		description string
		content     string
		wantErr     error
	}{
		{"empty id", " ", "description", "content", ErrInvalidSkillID},
		{"empty description", testUUID, " ", "content", ErrInvalidSkillDescription},
		{"empty content", testUUID, "description", " ", ErrInvalidSkillContent},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, mock, cleanup := newMockService(t)
			defer cleanup()

			err := s.UpdateSkill(context.Background(), tt.id, tt.description, tt.content)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestService_UpdateSkill_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectExec("UPDATE skills").
		WithArgs(testUUID, "Updated", "Updated content").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.UpdateSkill(context.Background(), testUUID, "Updated", "Updated content"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_DeleteSkill_ValidatesID(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	if err := s.DeleteSkill(context.Background(), ""); !errors.Is(err, ErrInvalidSkillID) {
		t.Fatalf("err = %v, want ErrInvalidSkillID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_DeleteSkill_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM skills").
		WithArgs(testUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.DeleteSkill(context.Background(), testUUID); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_CreateSkill_StoreErrorPropagates(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	boom := errors.New("db down")
	mock.ExpectQuery("INSERT INTO skills").
		WithArgs("prep", "Prepare ingredients", "Do the prep").
		WillReturnError(boom)

	_, err := s.CreateSkill(context.Background(), "prep", "Prepare ingredients", "Do the prep")
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
