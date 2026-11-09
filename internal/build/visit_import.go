package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitImport(imp *program.Import, rootScope *Scope, pkgScope *Scope) error {
	if b.progScope == nil || b.progScope.Program == nil {
		return fmt.Errorf("no program scope or program provided")
	}
	pkgPath := imp.Path
	pkg, err := b.progScope.Program.GetImport(pkgPath)
	if err != nil {
		if buildErr, ok := err.(*program.Error); ok {
			return PkgErrLoc(pkgPath, buildErr)
		}
		return ErrLoc(pkgScope, imp.Start, err)
	}
	if pkg == nil {
		return fmt.Errorf("could not load package %q", pkgPath)
	}

	depScope, err := b.progScope.AddDependency(pkg)
	if err != nil {
		return err
	}

	impScope, alreadyBuilt := b.importedScopes[pkgPath]
	if !alreadyBuilt {
		impScope = NewScope(b, pkgPath, imp.Start, rootScope, depScope)
		b.importedScopes[pkgPath] = impScope

		b.delegate.SetScope(impScope)
		defer b.delegate.SetScope(pkgScope)

		info, err := b.delegate.BuildImport(imp, pkg)
		if err != nil {
			return err
		}
		if info == nil {
			return fmt.Errorf("no import info returned for package %q", pkgPath)
		}

		fileScopes := make([]*Scope, len(pkg.Files))

		for idx, file := range pkg.Files {
			fileScope := NewScope(b, fmt.Sprintf("%s_file_%d", pkgPath, idx), nil, impScope, depScope.Files[idx])
			fileScope.isFileScope = true
			fileScope.fileIdx = idx
			fileScopes[idx] = fileScope

			b.delegate.SetScope(fileScope)
			for _, nestedImp := range file.Imports {
				err = b.visitImport(nestedImp, rootScope, fileScope)
				if err != nil {
					return err
				}
				CheckScope(b, fileScope)
			}

			for _, fnType := range file.FnTypes {
				err := b.visitFnType(fnType, impScope)
				if err != nil {
					return err
				}
				CheckScope(b, fileScope)
			}
		}

		for idx, file := range pkg.Files {
			fileScope := fileScopes[idx]
			b.delegate.SetScope(fileScope)

			for _, assign := range file.Assignments {
				err := b.visitAssignment(assign, fileScope)
				if err != nil {
					return err
				}
				CheckScope(b, fileScope)
			}
		}

		b.delegate.SetScope(impScope)

		err = b.delegate.EndBuildImport(info)
		if err != nil {
			return err
		}
	}

	return pkgScope.addImport(imp, impScope)
}
