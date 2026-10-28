package pkg

import "gradientcontrol.com/omega/core/program"

type Manager interface {
	Package(path string) (*program.Package, error)
	Source(path string) (string, error)
}
