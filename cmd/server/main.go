package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fitraditya/useria/internal/config"
	"github.com/fitraditya/useria/internal/database"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage: server [command]

commands:
  (none)   run the customer-facing app — same as "run" (SERVER_PORT, default 8080)
  run      run the customer-facing app (SERVER_PORT, default 8080)
  admin    run the SuperAdmin app (ADMIN_PORT, default 9080)
  migrate  apply the database schema
  seed     seed default roles/permissions/plans, then bootstrap the
           SuperAdmin and demo companies from seed.yaml if present
           (override path with SEED_FILE)

Copy seed.yaml.example to seed.yaml and edit it — that file is gitignored
since it ends up holding real credentials. seed is idempotent and safe to
rerun after editing it; without a seed.yaml it still seeds the base
roles/permissions/plans, just skips the admin/company bootstrap.`)
}

func main() {
	cmd := "run"
	if len(os.Args) >= 2 {
		cmd = os.Args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	db, err := database.Connect(cfg.DBDriver, cfg.DSN())
	if err != nil {
		log.Fatalf("database connect: %v", err)
	}
	defer db.Close()

	switch cmd {
	case "migrate":
		runMigration(db, cfg.MigrationPath())
	case "seed":
		if err := database.Seed(db); err != nil {
			log.Fatalf("seed: %v", err)
		}
		log.Println("seed complete")

		data, err := database.LoadSeedData(seedFilePath())
		if err != nil {
			log.Printf("skipping admin/company bootstrap: %v", err)
			return
		}
		seedAdminFromData(db, data)
		seedCompaniesFromData(db, data)
	case "run":
		d := buildDeps(cfg, db)
		serve("run", cfg.ServerPort, newTenantRouter(d), cfg.ServerEnv)
	case "admin":
		d := buildDeps(cfg, db)
		serve("admin", cfg.AdminPort, newAdminRouter(d), cfg.ServerEnv)
	default:
		usage()
		os.Exit(1)
	}
}

func seedFilePath() string {
	if path := os.Getenv("SEED_FILE"); path != "" {
		return path
	}
	return "seed.yaml"
}

func seedAdminFromData(db *sql.DB, data *database.SeedData) {
	if data.SuperAdmin == nil {
		return
	}
	sa := data.SuperAdmin
	if sa.Email == "" || sa.Password == "" {
		log.Fatal("seed.yaml super_admin needs email and password")
	}
	created, err := database.SeedAdmin(db, sa.Email, sa.Password, sa.FirstName, sa.LastName)
	if err != nil {
		log.Fatalf("seed-admin: %v", err)
	}
	if created {
		log.Printf("SuperAdmin created: %s", sa.Email)
	} else {
		log.Printf("SuperAdmin already exists: %s (membership ensured)", sa.Email)
	}
}

func seedCompaniesFromData(db *sql.DB, data *database.SeedData) {
	for _, c := range data.Companies {
		if c.Name == "" || c.Admin.Email == "" || c.Admin.Password == "" {
			log.Fatalf("company %q needs a name and an admin email/password", c.Name)
		}

		adminCreated, err := database.SeedCompanyUser(db, c.Name, c.Admin.Email, c.Admin.Password, c.Admin.FirstName, c.Admin.LastName, "Admin")
		if err != nil {
			log.Fatalf("seed-companies (%s admin): %v", c.Name, err)
		}
		if adminCreated {
			log.Printf("Admin created: %s (%s)", c.Admin.Email, c.Name)
		} else {
			log.Printf("Admin already exists: %s (membership ensured in %s)", c.Admin.Email, c.Name)
		}

		for _, m := range c.Members {
			if m.Email == "" || m.Password == "" || m.Role == "" {
				log.Fatalf("company %q has a member missing email/password/role", c.Name)
			}
			memberCreated, err := database.SeedCompanyUser(db, c.Name, m.Email, m.Password, m.FirstName, m.LastName, m.Role)
			if err != nil {
				log.Fatalf("seed-companies (%s member %s): %v", c.Name, m.Email, err)
			}
			if memberCreated {
				log.Printf("%s created: %s (%s)", m.Role, m.Email, c.Name)
			} else {
				log.Printf("%s already exists: %s (membership ensured in %s)", m.Role, m.Email, c.Name)
			}
		}
	}
}

func serve(name, port string, handler http.Handler, env string) {
	addr := ":" + port
	log.Printf("useria %s listening on %s (env=%s)", name, addr, env)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("%s server: %v", name, err)
	}
}

func runMigration(db *sql.DB, path string) {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read migration %s: %v", path, err)
	}

	var cleaned []string
	for _, line := range strings.Split(string(sqlBytes), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		cleaned = append(cleaned, line)
	}

	statements := strings.Split(strings.Join(cleaned, "\n"), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			log.Fatalf("exec migration statement: %v\n%s", err, stmt)
		}
	}
	log.Println("migration complete")
}
