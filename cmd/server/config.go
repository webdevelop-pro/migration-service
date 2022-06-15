package main

type Config struct {
	ApplyOnly bool `split_words:"true" default:"false"`
}
