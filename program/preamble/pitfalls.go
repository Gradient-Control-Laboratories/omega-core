package preamble

// CommonPitfalls targets failure modes we've seen models trip on repeatedly.
// It closes the GSL author skill (GSLSkillIndex, GSLAuthorSkill) and is
// appended to the chat system message ahead of the user prompt. Kept here —
// not in GSLDescription — so it ships only to the model and not to users via
// the in-app Language Guide (which renders GSLDescription).
//
// Style notes: no exclamation noise — LLMs read tokens; the surrounding
// punctuation just dilutes the signal. The trailing "Re-check…" line is the
// repetition trick (recency bias works on LLMs too).
const CommonPitfalls = `
# CRITICAL — common failures, verify against each before responding

**HARD CONSTRAINT: name variables for what they REPRESENT, not for their type.** GSL has no type annotations. ` + "`f32`" + `, ` + "`u8`" + `, ` + "`i32`" + `, ` + "`u32`" + `, ` + "`i64`" + ` are reserved type tags from other languages and **MUST NOT** appear as identifiers — not as bare names, not as the LHS of an assignment, never. If you're tempted to write ` + "`f32 = ...`" + `, ask what the value MEANS (radius? height? frequency? count?) and use that word.

* Nested ternaries group to the right: ` + "`a ? x : b ? y : z`" + ` is ` + "`a ? x : (b ? y : z)`" + `. For sign-based dispatch on a single value, GSL has a built-in three-way form: ` + "`value ? posVal : zeroVal : negVal`" + `.
* Don't use reserved words for variable names. The most common offenders, with concrete alternatives:
  - ` + "`shape`" + ` → ` + "`myShape`" + ` or ` + "`unionShape`" + `
  - ` + "`out`" + ` → ` + "`result`" + ` or ` + "`output`" + `
  - ` + "`mesh`" + ` → ` + "`meshOut`" + `
  - any builtin function name like ` + "`length`" + `, ` + "`distance`" + `, ` + "`dot`" + `, ` + "`cross`" + `, ` + "`clamp`" + `, ` + "`min`" + `, ` + "`max`" + `, ` + "`abs`" + `, ` + "`sign`" + ` → suffix with ` + "`Val`" + ` (` + "`lenVal`" + `, ` + "`distVal`" + `, ` + "`signVal`" + `…)
  - the globals ` + "`time`" + ` and ` + "`p`" + ` (the implicit point) → ` + "`t`" + `, ` + "`pt`" + `
  - Type tags (` + "`f32`" + `, ` + "`u8`" + `, ` + "`i32`" + `, etc.): see the hard constraint above. Name by purpose.
  Variables that hold shapes are nouns describing the role (` + "`bodyBox`" + `, ` + "`eyeSphere`" + `, ` + "`armCylinder`" + `), never the bare word ` + "`shape`" + `.
* All variables are immutable (SSA). You cannot reassign a variable after it has been declared.
* Collections hold any number of elements, but swizzles (` + "`.x`" + ` to ` + "`.w`" + `) name only the first four:
index further elements with constant integers or iterate with ` + "`map`" + ` and ` + "`reduce`" + `.
A ` + "`map`" + ` whose function returns a vector produces a list of vectors.
* ` + "`atan2(y, x)`" + ` is not supported. Use ` + "`atan(y, x)`" + ` (two arguments) instead.
* Use ` + "`pow(base, exponent)`" + `. The ` + "`**`" + ` and ` + "`^`" + ` operators do not exist in GSL.
* No ` + "`if`" + `, no ` + "`switch`" + `, no loops. Use conditional expressions and recursion.
* When passing an anonymous function as a value to be invoked, include the call parentheses, e.g. ` + "`result = remap(() { return length(p); }(), -p);`" + `
* You can return multiple shapes; their variable names appear in a panel for the user.

**RE-READ THE HARD CONSTRAINT BEFORE RESPONDING.** The most frequent failure is using a reserved word — especially a type tag like ` + "`f32`" + ` or ` + "`u8`" + ` — as a variable name.
`
