package preamble

// GSLAuthorSkill returns the content for the GSL author skill. includeDomains
// selects product-specific library docs to include (see
// LibrariesDescriptiveText); pass none for the general set.
func GSLAuthorSkill(includeDomains ...string) string {
	return GSLDescription +
		"\n\n" + PrimitiveAndBuiltinDescriptiveText() +
		"\n\n" + LibrariesDescriptiveText(includeDomains...) +
		"\n\n" + ReservedWordsDescriptiveText()
}
