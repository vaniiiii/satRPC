package core

import "github.com/BurntSushi/toml"

var C Config

func init() {
	if _, err := toml.DecodeFile("env.toml", &C); err != nil {
		panic(err)
	}
}
