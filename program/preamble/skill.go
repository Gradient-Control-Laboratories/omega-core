package preamble

// GSLAuthorSkill returns the content for the GSL author skill. includeDomains
// selects product-specific library docs to include (see
// LibrariesDescriptiveText); pass none for the general set. It leaves out
// CommonPitfalls: a chat that sends it appends the pitfalls itself, last,
// where they weigh most. GSLSkillIndex, which a model reads on its own,
// carries them.
func GSLAuthorSkill(includeDomains ...string) string {
	return GSLDescription +
		"\n\n" + PrimitiveAndBuiltinDescriptiveText() +
		"\n\n" + LibrariesDescriptiveText(includeDomains...) +
		"\n\n" + ReservedWordsDescriptiveText()
}
