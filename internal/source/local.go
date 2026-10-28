package source

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

func loadLocalSources(localPath string) ([]string, []string, error) {

	// If there is a package.yaml file,
	// then use it to determine the source files and asset file names
	pkgYamlFile := filepath.Join(localPath, "package.yaml")
	if stat, err := os.Stat(pkgYamlFile); err == nil && !stat.IsDir() {
		return loadLocalPackageYaml(localPath, pkgYamlFile)
	}

	// If there is a <pkgname>.gsl file in the local path, treat it as the primary source
	// and treat all other non *.gsl files as assets.
	pkgName := filepath.Base(localPath)
	pkgFile := filepath.Join(localPath, pkgName+".gsl")
	if stat, err := os.Stat(pkgFile); err == nil && !stat.IsDir() {
		return loadLocalSingleSource(localPath, pkgFile)
	}

	// If neither of the above, then package not found here
	return nil, nil, nil
}

// Read the package.yaml file and return the source files and asset file names
//
// The package yaml file must contain:
//
//	sources: ["file1", "file2", ...]
//
// The package yaml file may contain:
//
//	assets: ["file1", "file2", ...]
//
// If assets is present, then all files listed are assets.
// If assets is not present, then all non `*.gsl` files are assets.
func loadLocalPackageYaml(localPath string, pkgYamlFile string) ([]string, []string, error) {
	yamlBytes, err := os.ReadFile(pkgYamlFile)
	if err != nil {
		return nil, nil, err
	}
	type packageYaml struct {
		Sources []string  `yaml:"sources"`
		Assets  *[]string `yaml:"assets"`
	}
	var config packageYaml
	if err := yaml.Unmarshal(yamlBytes, &config); err != nil {
		return nil, nil, err
	}

	var contents []string
	for _, filename := range config.Sources {
		filePath := filepath.Join(localPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, nil, err
		}
		contents = append(contents, string(content))
	}

	var assetNames []string
	if config.Assets != nil {
		assetNames = *config.Assets
	} else {
		var err error
		assetNames, err = collectLocalAssetNames(localPath)
		if err != nil {
			return nil, nil, err
		}
	}

	return contents, assetNames, nil
}

// Treat the GSL file as the primary source and all other non `*.gsl` files as assets.
func loadLocalSingleSource(localPath string, pkgFile string) ([]string, []string, error) {
	content, err := os.ReadFile(pkgFile)
	if err != nil {
		return nil, nil, err
	}
	assetNames, err := collectLocalAssetNames(localPath)
	if err != nil {
		return nil, nil, err
	}
	return []string{string(content)}, assetNames, nil
}

func collectLocalAssetNames(localPath string) ([]string, error) {
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return nil, err
	}
	var assetNames []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) != ".gsl" && entry.Name() != "package.yaml" {
			assetNames = append(assetNames, entry.Name())
		}
	}
	sort.Strings(assetNames)
	return assetNames, nil
}
