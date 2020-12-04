module github.com/pion/mediadevices/examples

go 1.14

// Please don't commit require entries of examples.
// `git checkout master examples/go.mod` to revert this file.
require github.com/pion/mediadevices v0.0.0

replace github.com/pion/mediadevices v0.0.0 => ../
