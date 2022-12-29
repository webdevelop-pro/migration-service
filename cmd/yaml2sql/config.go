package main

const pkgName = "migration"

type Config struct {
	Yaml string `required:"true"`
	Sql  string `required:"true"`
}
