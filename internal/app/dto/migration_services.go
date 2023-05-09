package dto

type MigrationServicesLog struct {
	MigrationServiceName string
	Priority             int
	Version              int
	FileName             string
	SQL                  string
	Hash                 string
}
