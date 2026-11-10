//go:build !omega_static

package fidget

/*
// Link against the release build of the rust wrapper. The bvhSdf oracle's
// BSP traversal + Fast GWN is a performance-critical inner loop, and the
// debug profile is ~10x slower. The launcher script does `cargo build
// --release` to match.
#cgo LDFLAGS: -L../../../../target/release -lrender
*/
import "C"
