package preamble

// GSLAuthorSkill returns the content for the GSL author skill. The domains
// argument is ignored and will be removed: no library sets a domain. It leaves out
// CommonPitfalls: a chat that sends it appends the pitfalls itself, last,
// where they weigh most. GSLSkillIndex, which a model reads on its own,
// carries them.
func GSLAuthorSkill(domains ...string) string {
	return GSLDescription +
		"\n\n" + PrimitiveAndBuiltinDescriptiveText() +
		"\n\n" + LibrariesDescriptiveText() +
		"\n\n" + ReservedWordsDescriptiveText()
}
