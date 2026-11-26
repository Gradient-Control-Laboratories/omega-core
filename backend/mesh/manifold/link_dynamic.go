//go:build !omega_static

package manifold

/*
// Link against the release build of the glue dylib (tool/build-fidget.sh builds it into target/release).
#cgo LDFLAGS: -L../../../target/release -lmanifold_glue
#cgo linux LDFLAGS: -Wl,-rpath,${SRCDIR}/../../../target/release
*/
import "C"
