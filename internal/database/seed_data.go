package database

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SeedData is the shape of seed.yaml, consumed by `seed-admin` and
// `seed-companies`. Copy seed.yaml.example to seed.yaml and edit it —
// seed.yaml itself is gitignored since it ends up holding real credentials.
type SeedData struct {
	SuperAdmin *SeedUser     `yaml:"super_admin"`
	Companies  []SeedCompany `yaml:"companies"`
}

type SeedUser struct {
	Email     string `yaml:"email"`
	Password  string `yaml:"password"`
	FirstName string `yaml:"first_name"`
	LastName  string `yaml:"last_name"`
}

type SeedCompany struct {
	Name    string       `yaml:"name"`
	Admin   SeedUser     `yaml:"admin"`
	Members []SeedMember `yaml:"members"`
}

type SeedMember struct {
	SeedUser `yaml:",inline"`
	Role     string `yaml:"role"`
}

func LoadSeedData(path string) (*SeedData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read seed data %s (copy seed.yaml.example to seed.yaml first): %w", path, err)
	}
	var data SeedData
	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("parse seed data %s: %w", path, err)
	}
	return &data, nil
}
