//go:build omega_static

package fidget

/*
// Release binaries (the omega-gsl wheel, tool/build-wheel.sh) link the rust
// wrapper as a static library, so a program that uses it is one file that
// loads only the system's libraries. tool/build-wheel.sh builds it with
//   cargo rustc -p render --release --crate-type staticlib --target-dir target/static
// which bundles OpenVDB, TBB, Imath, blosc and zlib into librender.a. What
// else it needs is the system's: rustc --print native-static-libs. The C++
// runtime is the system's (libc++ on macOS, libstdc++ on Linux, where the
// manylinux image's toolchain keeps compile within manylinux_2_28's symbol
// versions, which tool/build-wheel.sh checks); libgcc is linked in.
#cgo LDFLAGS: ${SRCDIR}/../../../../target/static/release/librender.a
#cgo darwin LDFLAGS: -lc++ -liconv
#cgo linux LDFLAGS: -static-libgcc -lstdc++ -ldl -lpthread -lm
*/
import "C"
