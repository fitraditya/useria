package repository

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/fitraditya/useria/internal/database"
	"github.com/fitraditya/useria/internal/models"
	"github.com/fitraditya/useria/internal/utils"
)

// newTestDB spins up a throwaway SQLite database with the dev schema applied
// and the default roles/permissions seeded. Repository tests then exercise
// the real SQL through database.DB's placeholder rewriting, without needing
// an external server.
func newTestDB(t *testing.T) *database.DB {
	t.Helper()

	db, err := database.Connect("sqlite", filepath.Join(t.TempDir(), "useria-test.db"))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db, "../database/migrations/001_init_sqlite.sql"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

func createUser(t *testing.T, db *database.DB, email string) *models.User {
	t.Helper()

	u := &models.User{ID: utils.NewUUID(), Email: email, Status: models.UserStatusActive}
	if err := NewUserRepository(db).Create(context.Background(), u); err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	return u
}

func TestUserRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := NewUserRepository(db)

	u := createUser(t, db, "ada@example.test")

	got, err := repo.GetByEmail(ctx, "ada@example.test")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("GetByEmail returned id %q, want %q", got.ID, u.ID)
	}

	if _, err := repo.GetByID(ctx, u.ID); err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if err := repo.UpdateProfile(ctx, u.ID, "Ada", "Lovelace"); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	updated, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if updated.FirstName == nil || *updated.FirstName != "Ada" {
		t.Fatalf("first name = %v, want Ada", updated.FirstName)
	}

	n, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 1 {
		t.Fatalf("Count = %d, want 1", n)
	}

	if _, err := repo.GetByEmail(ctx, "nobody@example.test"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByEmail(missing) err = %v, want ErrNotFound", err)
	}
}

func TestCompanyRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner := createUser(t, db, "owner@example.test")
	repo := NewCompanyRepository(db)

	c := &models.Company{
		ID:        utils.NewUUID(),
		Name:      "Acme Widgets",
		Slug:      utils.Slugify("Acme Widgets"),
		Plan:      "free",
		Status:    models.CompanyStatusActive,
		CreatedBy: owner.ID,
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	byID, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID.Slug != c.Slug {
		t.Fatalf("slug = %q, want %q", byID.Slug, c.Slug)
	}

	if _, err := repo.GetBySlug(ctx, c.Slug); err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}

	list, err := repo.List(ctx, "Acme", "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List returned %d rows, want 1", len(list))
	}

	if err := repo.Delete(ctx, c.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, c.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID after soft-delete err = %v, want ErrNotFound", err)
	}
}

func TestMemberRepositoryInviteAcceptRemove(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	admin := createUser(t, db, "admin@example.test")
	member := createUser(t, db, "member@example.test")

	role, err := NewRoleRepository(db).GetSystemRoleByName(ctx, "Admin")
	if err != nil {
		t.Fatalf("GetSystemRoleByName(Admin): %v", err)
	}

	company := &models.Company{
		ID:        utils.NewUUID(),
		Name:      "Acme",
		Slug:      utils.Slugify("Acme"),
		Plan:      "free",
		Status:    models.CompanyStatusActive,
		CreatedBy: admin.ID,
	}
	if err := NewCompanyRepository(db).Create(ctx, company); err != nil {
		t.Fatalf("create company: %v", err)
	}

	repo := NewMemberRepository(db)
	m := &models.CompanyMember{
		ID:        utils.NewUUID(),
		CompanyID: company.ID,
		UserID:    member.ID,
		RoleID:    role.ID,
		Status:    models.MemberStatusInvited,
	}
	if err := repo.Create(ctx, m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := repo.ListByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListByCompanyID: %v", err)
	}
	if len(list) != 1 || list[0].RoleName != "Admin" || list[0].Email != "member@example.test" {
		t.Fatalf("unexpected member row: %+v", list)
	}

	if err := repo.Accept(ctx, m.ID); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	active, err := repo.ListByUserID(ctx, member.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("active memberships = %d, want 1", len(active))
	}

	if err := repo.Delete(ctx, m.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByCompanyAndUser(ctx, company.ID, member.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByCompanyAndUser after delete err = %v, want ErrNotFound", err)
	}
	any, err := repo.GetByCompanyAndUserAny(ctx, company.ID, member.ID)
	if err != nil {
		t.Fatalf("GetByCompanyAndUserAny: %v", err)
	}
	if any.DeletedAt == nil {
		t.Fatal("expected soft-deleted row to have DeletedAt set")
	}

	if err := repo.Revive(ctx, m.ID, role.ID); err != nil {
		t.Fatalf("Revive: %v", err)
	}
	revived, err := repo.GetByCompanyAndUser(ctx, company.ID, member.ID)
	if err != nil {
		t.Fatalf("GetByCompanyAndUser after revive: %v", err)
	}
	if revived.Status != models.MemberStatusInvited {
		t.Fatalf("status after revive = %q, want invited", revived.Status)
	}
}
