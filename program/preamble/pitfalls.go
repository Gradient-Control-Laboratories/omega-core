package preamble

// CommonPitfalls is appended to the chat system message ahead of the user
// prompt. It targets failure modes we've seen the LLM trip on repeatedly.
// Kept here — not in GSLDescription — so it ships only to the model and not
// to users via the in-app Language Guide (which renders GSLDescription).
//
// Style notes: no exclamation noise — LLMs read tokens; the surrounding
// punctuation just dilutes the signal. The trailing "Re-check…" line is the
// repetition trick (recency bias works on LLMs too).
const CommonPitfalls = `
# CRITICAL — common failures, verify against each before responding

**HARD CONSTRAINT 1: name variables for what they REPRESENT, not for their type.** GSL has no type annotations. ` + "`f32`" + `, ` + "`u8`" + `, ` + "`i32`" + `, ` + "`u32`" + `, ` + "`i64`" + ` are reserved type tags from other languages and **MUST NOT** appear as identifiers — not as bare names, not as the LHS of an assignment, never. If you're tempted to write ` + "`f32 = ...`" + `, ask what the value MEANS (radius? height? frequency? count?) and use that word.

**HARD CONSTRAINT 2: decompose every nested ternary.** GSL's ternary branches don't recursively parse another ternary by default — unparenthesized nesting like ` + "`a ? b : c ? d : e`" + ` is rejected. Parenthesized inline nesting like ` + "`a ? b : (c ? d : e)`" + ` is technically accepted, but easy to get wrong (you keep seeing this fail). **Always decompose by assigning the inner conditional to a variable first.** This is the most reliable pattern:

` + "```" + `
// WRONG (you keep doing this):
result = (a > 0) ? x : (b > 0) ? y : z;

// RIGHT:
inner = (b > 0) ? y : z;
result = (a > 0) ? x : inner;
` + "```" + `

For sign-based dispatch on a single value, GSL has a built-in three-way form: ` + "`value ? posVal : zeroVal : negVal`" + ` (no nesting required).

* Don't use reserved words for variable names. The most common LLM offenders, with concrete alternatives:
  - ` + "`shape`" + ` → ` + "`myShape`" + ` or ` + "`unionShape`" + `
  - ` + "`out`" + ` → ` + "`result`" + ` or ` + "`output`" + `
  - ` + "`mesh`" + ` → ` + "`meshOut`" + `
  - any builtin function name like ` + "`length`" + `, ` + "`distance`" + `, ` + "`dot`" + `, ` + "`cross`" + `, ` + "`clamp`" + `, ` + "`min`" + `, ` + "`max`" + `, ` + "`abs`" + `, ` + "`sign`" + ` → suffix with ` + "`Val`" + ` (` + "`lenVal`" + `, ` + "`distVal`" + `, ` + "`signVal`" + `…)
  - the globals ` + "`time`" + ` and ` + "`p`" + ` (the implicit point) → ` + "`t`" + `, ` + "`pt`" + `
  - Type tags (` + "`f32`" + `, ` + "`u8`" + `, ` + "`i32`" + `, etc.): see the hard constraint above. Name by purpose.
  Variables that hold shapes are nouns describing the role (` + "`bodyBox`" + `, ` + "`eyeSphere`" + `, ` + "`armCylinder`" + `), never the bare word ` + "`shape`" + `.
* All variables are immutable (SSA). You cannot reassign a variable after it has been declared.
* Vectors can only be up to 4 components (swizzling stops at ` + "`.w`" + `). Longer collections are lists:
index them with constant integers or iterate them with ` + "`map`" + ` and ` + "`reduce`" + `.
A ` + "`map`" + ` whose function returns a vector produces a list of vectors.
* ` + "`atan2(y, x)`" + ` is not supported. Use ` + "`atan(y, x)`" + ` (two arguments) instead.
* Use ` + "`pow(base, exponent)`" + `. The ` + "`**`" + ` and ` + "`^`" + ` operators do not exist in GSL.
* No ` + "`if`" + `, no ` + "`switch`" + `, no loops. Use conditional expressions and recursion.
* When passing an anonymous function as a value to be invoked, include the call parentheses, e.g. ` + "`result = remap(() { return length(p); }(), -p);`" + `
* You can return multiple shapes; their variable names appear in a panel for the user (Gemini, this means you).

**RE-READ HARD CONSTRAINTS 1 AND 2 BEFORE RESPONDING.** The two most frequent failures are (1) using a reserved word — especially a type tag like ` + "`f32`" + ` or ` + "`u8`" + ` — as a variable name, and (2) inlining a nested ternary instead of decomposing it.
`
