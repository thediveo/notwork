package main

import (
	"fmt"

	"github.com/thediveo/notwork/netns"
)

func init() {}

var et = netns.EnterTransient
var prtf = fmt.Printf
var _ = prtf

func somethingelse() {}

func bar() (int, int) { return 42, 666 }

func Foo() {
	a, b := bar()
	_ = a
	_ = b

	defer somethingelse()

	defer netns.EnterTransient()   // want "incorrect defer of EnterTransient.*"
	defer netns.EnterTransient()() // that's correct usage

	ET := et
	defer ET()   // want "incorrect defer of EnterTransient.*"
	defer ET()() // that's again correct usage
}

func main() {}
