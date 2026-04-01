package main

import "github.com/thediveo/notwork/netns"

var (
	_ = netns.EnterTransient // want "deprecated usage of github.com/thediveo/notwork/netns.EnterTransient"
	_ = netns.NsID[string]   // want "deprecated usage of github.com/thediveo/notwork/netns.NsID"
	_ = EnterTransient       // not a deprecated fn
)

func EnterTransient() {}

func Foo() {
	defer netns.EnterTransient()() // want "deprecated usage of github.com/thediveo/notwork/netns.EnterTransient"
}
