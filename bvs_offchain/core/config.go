package core

import (
	"fmt"

	"github.com/satlayer/satlayer-api/logger"

	"github.com/BurntSushi/toml"
)

var C Config
var L logger.Logger

// init initializes the package by loading configuration from env.toml and setting up the logger.
//
// No parameters.
// No return values.
func init() {
	// load env.toml file
	if _, err := toml.DecodeFile("env.toml", &C); err != nil {
		panic(err)
	}
	fmt.Println("C: ", C)
	L = logger.NewELKLogger(C.Chain.BvsHash)
}
