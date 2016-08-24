package goutils

import (
	"github.com/vharitonsky/iniflags"
)

var (
	pkgInitFn = []func(){}
)

// PkgInit is a deferred helper to initialize the package enviornment varible AFTER
// flags are set.
// Example usage:
// >package.go
// func init() {
//     utils.PkgInit(func() {
//         // Your initialization codes here... All flags are available.
//     })
// }
// >main.go
// func main() {
//     utils.Init()
//     // Other codes ...
// }
func PkgInit(f func()) {
	pkgInitFn = append(pkgInitFn, f)
}

// Init need to be execute in the beginning of main() to get PkgInit() to work.
// NOTE: It already called flag.Parse() alternative method. No need to call flag.Parse() any more.
func Init() {
	iniflags.Parse()
	for _, fn := range pkgInitFn {
		fn()
	}
}
