package migration

import (
	"strings"
)

// Migration is a single migration.
type Migration struct {
	AllowError bool
	NoAuto     bool
	Path       string
	Queries    []string
}

func NewMigration(queries []string, path string) Migration {
	mig := Migration{
		AllowError: false,
		Queries:    queries,
		Path:       path,
	}

	lines := strings.Split(queries[0], "\n")
	for _, line := range lines {
		line = strings.Replace(line, "\t", "", -1)
		// we don't have any comments at all
		if len(line) < 2 || line[0:2] != "--" {
			break
		}
		if len(line) > 15 && line[0:2] == "--" {
			comment := line[2:]
			comment = strings.Replace(comment, " ", "", -1)
			comment = strings.Replace(comment, "-", "", -1)
			pairs := strings.Split(comment, ",")
			for _, pair := range pairs {
				vals := strings.Split(pair, ":")
				if vals[0] == "allow_error" {
					if vals[1] == "true" || vals[1] == "1" {
						mig.AllowError = true
						break
					}
				}
			}
		}
	}
	return mig
}
