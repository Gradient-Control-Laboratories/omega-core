package main

import (
	_ "embed"
	"encoding/json"
)

// sections.json regroups the database's materials: it lists, by the section a
// material goes in, the ones whose database category is too broad to choose
// from. It is data, kept beside the snapshot; anything it does not list keeps
// its category.
//
//go:embed sections.json
var sectionsJSON []byte

// sectionOf is sections.json by material name.
var sectionOf = func() map[string]string {
	var sections map[string][]string
	if err := json.Unmarshal(sectionsJSON, &sections); err != nil {
		panic("sections.json: " + err.Error())
	}
	out := map[string]string{}
	for section, names := range sections {
		for _, name := range names {
			out[name] = section
		}
	}
	return out
}()
