package migration

// Migration is a single migration.
type Migration struct {
	AllowError bool
	NoAuto     bool
	Queries    []string
}
