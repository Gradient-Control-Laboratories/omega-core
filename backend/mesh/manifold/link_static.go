//go:build omega_static

package manifold

/*
// Release binaries (the omega-core wheel, tool/build-wheel.sh) link the glue as a static library, as
// internal/fidget does the render library, so a program that uses it loads only the system's libraries.
// tool/build-wheel.sh builds it with
//   cargo rustc -p manifold_glue --release --crate-type staticlib --target-dir target/static
// which bundles the vendored manifold. The C++ runtime is the system's.
#cgo LDFLAGS: ${SRCDIR}/../../../target/static/release/libmanifold_glue.a
#cgo darwin LDFLAGS: -lc++
#cgo linux LDFLAGS: -lstdc++ -lm
*/
import "C"
