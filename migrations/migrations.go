package migrations

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
)

//go:embed *.sql
var embedded embed.FS

var migrationFilePattern = regexp.MustCompile(`^([0-9]{6})_([a-z0-9_]+)\.(up|down)\.sql$`)

type Migration struct {
	ID      string
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

func Embedded() ([]Migration, error) {
	return Load(embedded)
}

func Load(source fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	byID := make(map[string]*Migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", entry.Name(), err)
		}

		id := matches[1] + "_" + matches[2]
		migration := byID[id]
		if migration == nil {
			migration = &Migration{
				ID:      id,
				Version: version,
				Name:    matches[2],
			}
			byID[id] = migration
		}

		sqlBytes, err := fs.ReadFile(source, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}

		switch matches[3] {
		case "up":
			migration.UpSQL = string(sqlBytes)
		case "down":
			migration.DownSQL = string(sqlBytes)
		}
	}

	result := make([]Migration, 0, len(byID))
	for _, migration := range byID {
		if migration.UpSQL == "" {
			return nil, fmt.Errorf("migration %s is missing up sql", migration.ID)
		}
		if migration.DownSQL == "" {
			return nil, fmt.Errorf("migration %s is missing down sql", migration.ID)
		}
		result = append(result, *migration)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	for i := 1; i < len(result); i++ {
		if result[i-1].Version == result[i].Version {
			return nil, fmt.Errorf("duplicate migration version %06d", result[i].Version)
		}
	}

	return result, nil
}
