package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fitraditya/useria/internal/utils"
)

const platformCompanySlug = "useria-platform"

type planSeed struct {
	Code       string
	Name       string
	PriceCents int
}

// defaultPlans mirrors the free/pro/enterprise tiers already used as plain
// strings on companies.plan. Pricing is a placeholder — no billing logic
// reads these yet.
var defaultPlans = []planSeed{
	{"free", "Free", 0},
	{"pro", "Pro", 2900},
	{"enterprise", "Enterprise", 9900},
}

type permissionSeed struct {
	Code     string
	Name     string
	Category string
}

var defaultPermissions = []permissionSeed{
	{"users:read", "View users", "users"},
	{"users:create", "Create users", "users"},
	{"users:write", "Edit users", "users"},
	{"users:delete", "Delete users", "users"},
	{"companies:read", "View companies", "companies"},
	{"companies:create", "Create companies", "companies"},
	{"companies:write", "Edit companies", "companies"},
	{"companies:delete", "Delete companies", "companies"},
	{"members:read", "View team members", "members"},
	{"members:create", "Invite team members", "members"},
	{"members:write", "Edit team member roles", "members"},
	{"members:delete", "Remove team members", "members"},
	{"settings:read", "View settings", "settings"},
	{"settings:write", "Edit settings", "settings"},
	{"audit:read", "View audit logs", "audit"},
}

type roleSeed struct {
	Name        string
	Description string
	Permissions []string
}

var defaultRoles = []roleSeed{
	{
		Name:        "SuperAdmin",
		Description: "Full access across all companies (cannot invite team members)",
		Permissions: []string{
			"users:read", "users:create", "users:write", "users:delete",
			"companies:read", "companies:create", "companies:write", "companies:delete",
			"members:read", "members:write", "members:delete",
			"settings:read", "settings:write",
			"audit:read",
		},
	},
	{
		Name:        "Admin",
		Description: "Manage team, roles, and settings for own company",
		Permissions: []string{
			"members:read", "members:create", "members:write", "members:delete",
			"settings:read", "settings:write",
		},
	},
	{
		Name:        "Manager",
		Description: "View team and invite members",
		Permissions: []string{"members:read", "members:create", "settings:read"},
	},
	{
		Name:        "Member",
		Description: "View own profile and basic data",
		Permissions: []string{"members:read"},
	},
	{
		Name:        "Consultant",
		Description: "Read-only access",
		Permissions: []string{"members:read", "settings:read"},
	},
}

// Seed inserts default system permissions and roles if they do not already exist.
func Seed(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	permIDs := make(map[string]string, len(defaultPermissions))
	for _, p := range defaultPermissions {
		var id string
		err := tx.QueryRow(`SELECT id FROM permissions WHERE code = ?`, p.Code).Scan(&id)
		if err == sql.ErrNoRows {
			id = uuid.NewString()
			if _, err := tx.Exec(
				`INSERT INTO permissions (id, code, name, category) VALUES (?, ?, ?, ?)`,
				id, p.Code, p.Name, p.Category,
			); err != nil {
				return fmt.Errorf("insert permission %s: %w", p.Code, err)
			}
		} else if err != nil {
			return fmt.Errorf("lookup permission %s: %w", p.Code, err)
		}
		permIDs[p.Code] = id
	}

	for _, r := range defaultRoles {
		var roleID string
		err := tx.QueryRow(
			`SELECT id FROM roles WHERE company_id IS NULL AND name = ?`, r.Name,
		).Scan(&roleID)
		if err == sql.ErrNoRows {
			roleID = uuid.NewString()
			if _, err := tx.Exec(
				`INSERT INTO roles (id, company_id, name, description, is_system) VALUES (?, NULL, ?, ?, TRUE)`,
				roleID, r.Name, r.Description,
			); err != nil {
				return fmt.Errorf("insert role %s: %w", r.Name, err)
			}
		} else if err != nil {
			return fmt.Errorf("lookup role %s: %w", r.Name, err)
		}

		for _, code := range r.Permissions {
			permID, ok := permIDs[code]
			if !ok {
				return fmt.Errorf("unknown permission code %s for role %s", code, r.Name)
			}
			var exists string
			err := tx.QueryRow(
				`SELECT id FROM role_permissions WHERE role_id = ? AND permission_id = ?`,
				roleID, permID,
			).Scan(&exists)
			if err == sql.ErrNoRows {
				if _, err := tx.Exec(
					`INSERT INTO role_permissions (id, role_id, permission_id) VALUES (?, ?, ?)`,
					uuid.NewString(), roleID, permID,
				); err != nil {
					return fmt.Errorf("link permission %s to role %s: %w", code, r.Name, err)
				}
			} else if err != nil {
				return fmt.Errorf("lookup role_permission %s/%s: %w", roleID, permID, err)
			}
		}

		// Prune links for permissions no longer in this role's list, so
		// changes to defaultRoles take effect on re-seed, not just inserts.
		desired := make(map[string]bool, len(r.Permissions))
		for _, code := range r.Permissions {
			desired[permIDs[code]] = true
		}
		linked, err := tx.Query(`SELECT permission_id FROM role_permissions WHERE role_id = ?`, roleID)
		if err != nil {
			return fmt.Errorf("list linked permissions for role %s: %w", r.Name, err)
		}
		var stale []string
		for linked.Next() {
			var pid string
			if err := linked.Scan(&pid); err != nil {
				linked.Close()
				return fmt.Errorf("scan linked permission for role %s: %w", r.Name, err)
			}
			if !desired[pid] {
				stale = append(stale, pid)
			}
		}
		linked.Close()
		for _, pid := range stale {
			if _, err := tx.Exec(`DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?`, roleID, pid); err != nil {
				return fmt.Errorf("prune stale permission from role %s: %w", r.Name, err)
			}
		}
	}

	for _, p := range defaultPlans {
		var exists string
		err := tx.QueryRow(`SELECT id FROM plans WHERE code = ?`, p.Code).Scan(&exists)
		if err == sql.ErrNoRows {
			if _, err := tx.Exec(
				`INSERT INTO plans (id, code, name, price_cents) VALUES (?, ?, ?, ?)`,
				uuid.NewString(), p.Code, p.Name, p.PriceCents,
			); err != nil {
				return fmt.Errorf("insert plan %s: %w", p.Code, err)
			}
		} else if err != nil {
			return fmt.Errorf("lookup plan %s: %w", p.Code, err)
		}
	}

	return tx.Commit()
}

// SeedAdmin ensures a platform SuperAdmin user exists, creating it (and the
// bootstrap "Useria Platform" company that holds the membership) if needed.
// Run Seed first so the SuperAdmin system role is present.
// Returns true if a new admin user was created.
func SeedAdmin(db *sql.DB, email, password, firstName, lastName string) (bool, error) {
	if firstName == "" {
		firstName = "Super"
	}
	if lastName == "" {
		lastName = "Admin"
	}

	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var superAdminRoleID string
	err = tx.QueryRow(
		`SELECT id FROM roles WHERE company_id IS NULL AND name = 'SuperAdmin'`,
	).Scan(&superAdminRoleID)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("SuperAdmin role not found, run seed first")
	} else if err != nil {
		return false, fmt.Errorf("lookup SuperAdmin role: %w", err)
	}

	var userID string
	created := false
	err = tx.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&userID)
	if err == sql.ErrNoRows {
		hash, err := utils.HashPassword(password)
		if err != nil {
			return false, fmt.Errorf("hash admin password: %w", err)
		}
		userID = uuid.NewString()
		if _, err := tx.Exec(
			`INSERT INTO users (id, email, password_hash, first_name, last_name, oauth_provider, status)
			 VALUES (?, ?, ?, ?, ?, 'local', 'active')`,
			userID, email, hash, firstName, lastName,
		); err != nil {
			return false, fmt.Errorf("insert admin user: %w", err)
		}
		created = true
	} else if err != nil {
		return false, fmt.Errorf("lookup admin user: %w", err)
	}

	var companyID string
	err = tx.QueryRow(`SELECT id FROM companies WHERE slug = ?`, platformCompanySlug).Scan(&companyID)
	if err == sql.ErrNoRows {
		companyID = uuid.NewString()
		if _, err := tx.Exec(
			`INSERT INTO companies (id, name, slug, plan, status, created_by) VALUES (?, ?, ?, 'enterprise', 'active', ?)`,
			companyID, "Useria Platform", platformCompanySlug, userID,
		); err != nil {
			return false, fmt.Errorf("insert platform company: %w", err)
		}
	} else if err != nil {
		return false, fmt.Errorf("lookup platform company: %w", err)
	}

	var memberID string
	err = tx.QueryRow(
		`SELECT id FROM company_members WHERE company_id = ? AND user_id = ?`, companyID, userID,
	).Scan(&memberID)
	if err == sql.ErrNoRows {
		now := time.Now()
		if _, err := tx.Exec(
			`INSERT INTO company_members (id, company_id, user_id, role_id, status, joined_at) VALUES (?, ?, ?, ?, 'active', ?)`,
			uuid.NewString(), companyID, userID, superAdminRoleID, now,
		); err != nil {
			return false, fmt.Errorf("insert admin membership: %w", err)
		}
	} else if err != nil {
		return false, fmt.Errorf("lookup admin membership: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return created, nil
}

// SeedCompanyUser ensures a company exists (found or created by the slug
// derived from companyName) and that the given user is an active member of
// it holding roleName — creating the user if needed, or updating their role
// if they're already a member. Run Seed first so the role exists. Returns
// true if a new user was created.
func SeedCompanyUser(db *sql.DB, companyName, userEmail, userPassword, firstName, lastName, roleName string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var roleID string
	err = tx.QueryRow(
		`SELECT id FROM roles WHERE company_id IS NULL AND name = ?`, roleName,
	).Scan(&roleID)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("role %s not found, run seed first", roleName)
	} else if err != nil {
		return false, fmt.Errorf("lookup role %s: %w", roleName, err)
	}

	var userID string
	created := false
	err = tx.QueryRow(`SELECT id FROM users WHERE email = ?`, userEmail).Scan(&userID)
	if err == sql.ErrNoRows {
		hash, err := utils.HashPassword(userPassword)
		if err != nil {
			return false, fmt.Errorf("hash password: %w", err)
		}
		userID = uuid.NewString()
		if _, err := tx.Exec(
			`INSERT INTO users (id, email, password_hash, first_name, last_name, oauth_provider, status)
			 VALUES (?, ?, ?, ?, ?, 'local', 'active')`,
			userID, userEmail, hash, firstName, lastName,
		); err != nil {
			return false, fmt.Errorf("insert user: %w", err)
		}
		created = true
	} else if err != nil {
		return false, fmt.Errorf("lookup user %s: %w", userEmail, err)
	}

	slug := utils.Slugify(companyName)
	var companyID string
	err = tx.QueryRow(`SELECT id FROM companies WHERE slug = ?`, slug).Scan(&companyID)
	if err == sql.ErrNoRows {
		companyID = uuid.NewString()
		if _, err := tx.Exec(
			`INSERT INTO companies (id, name, slug, plan, status, created_by) VALUES (?, ?, ?, 'free', 'active', ?)`,
			companyID, companyName, slug, userID,
		); err != nil {
			return false, fmt.Errorf("insert company: %w", err)
		}
	} else if err != nil {
		return false, fmt.Errorf("lookup company %s: %w", slug, err)
	}

	var memberID string
	err = tx.QueryRow(
		`SELECT id FROM company_members WHERE company_id = ? AND user_id = ?`, companyID, userID,
	).Scan(&memberID)
	switch {
	case err == sql.ErrNoRows:
		now := time.Now()
		if _, err := tx.Exec(
			`INSERT INTO company_members (id, company_id, user_id, role_id, status, joined_at) VALUES (?, ?, ?, ?, 'active', ?)`,
			uuid.NewString(), companyID, userID, roleID, now,
		); err != nil {
			return false, fmt.Errorf("insert membership: %w", err)
		}
	case err != nil:
		return false, fmt.Errorf("lookup membership: %w", err)
	default:
		if _, err := tx.Exec(`UPDATE company_members SET role_id = ? WHERE id = ?`, roleID, memberID); err != nil {
			return false, fmt.Errorf("update membership role: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return created, nil
}
