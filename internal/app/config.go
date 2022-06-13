package app

type Config struct {
	Dir        string `required:"true"`
	ForceApply bool   `required:"true" split_words:"true"`
	ApplyOnly  bool   `split_words:"true" default:"false"`
}
