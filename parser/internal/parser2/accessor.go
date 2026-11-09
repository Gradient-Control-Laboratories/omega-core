package parser2

import "fmt"

func IsAccessor(accessor string) (bool, error) {
	for _, c := range accessor {
		if c != 'x' && c != 'y' && c != 'z' && c != 'w' {
			return false, nil
		}
	}
	if len(accessor) > 4 {
		return false, fmt.Errorf("accessor must be 1 to 4 characters: %s", accessor)
	}
	return true, nil
}
