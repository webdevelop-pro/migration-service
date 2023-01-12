package migration

import (
	"strings"
)

// Migration is a single migration.
type Migration struct {
	AllowError bool
	NoAuto     bool
	Queries    []string
}

func NewMigration(queries []string) Migration {
	mig := Migration{
		AllowError: false,
		Queries:    queries,
	}

	lines := strings.Split(queries[0], "\n")
	if len(lines[0]) > 15 && lines[0][0:3] == "---" {
		comment := lines[0][3:len(lines[0])]
		pairs := strings.Split(comment, ",")
		for _, pair := range pairs {
			pair = strings.Replace(pair, " ", "", -1)
			vals := strings.Split(pair, ":")
			if vals[0] == "allow_error" {
				if vals[1] == "true" || vals[1] == "1" {
					mig.AllowError = true
				}
			}
		}
	}
	return mig
}
