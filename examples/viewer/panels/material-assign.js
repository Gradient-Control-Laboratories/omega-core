// Assigning a surface to a body, as a source edit.
//
// The document is the truth: picking a swatch writes the same GSL an author
// would write by hand, so the choice survives a reload, shows up in the diff,
// and composes with everything else the material libraries do. Nothing here
// touches the renderer.
//
// The two annotation kinds are written differently, and the difference is the
// one common/program/annotation draws. An @material returns a record AND the
// albedo that goes with it, so it is destructured and sets both properties. An
// @finish carries no colour on purpose, so it sets .material alone and leaves
// whatever colour the body already had — which is the whole point of a finish:
// it is meant to be paired with a colour from somewhere else.
//
// This module is a pure string -> string function, the way mesh-import.js's
// placeMeshInSource is, so the whole contract is testable without a browser.

import { topLevelReturnSpan } from "../render/gsl-return-span.js";

// isAssignableBody reports whether `name` is something a surface can be put on.
//
// A shape's name comes from program.OperandName (the wgsl agent's), which is a
// label rather than a binding: a document returning `sphere(3)` names that
// shape "sphere", and one returning `min(a, b)` names it "min(a, b)". Neither
// is a value the document binds, so assigning to it would write GSL that means
// nothing — `sphere.material = …` sets a property on a library function.
//
// A body is assignable when it is an identifier the document binds at the top
// level, which is where the panel's own lines go. Column zero is what makes it
// top level, the same convention the libraries' doc extraction reads.
//
// A binding can name more than one body at once. A multi-return destructure —
//
//	g1, g2 = gears.gearPair(…);
//	return g1, g2;
//
// binds g1 and g2 just as separate lines would, and `g1.material = …` is
// ordinary GSL. So the whole left-hand side is read, not just a first name.
export function isAssignableBody(source, name) {
    if (!name || !/^[A-Za-z_]\w*$/.test(name)) return false;
    for (const line of source.split("\n")) {
        const bound = boundNames(line);
        if (bound && bound.includes(name)) return true;
    }
    return false;
}

// boundNames returns the names a top-level binding line declares, or null when
// the line is not one. The left-hand side is identifiers and commas only, which
// is what separates a binding from a property assignment (`bolt.material = …`,
// which has a dot), a function declaration (which has parentheses) and a
// comparison (`=`, never `==`).
function boundNames(line) {
    const m = line.match(/^([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)*)\s*=(?!=)/);
    return m ? m[1].split(",").map((n) => n.trim()) : null;
}

// propertyStartRE matches the first line of `body.prop = …`. The body name is
// anchored so `bolt` does not match `boltHead`.
function propertyStartRE(body, prop) {
    return new RegExp(`^\\s*${escapeName(body)}\\.${prop}\\s*=`);
}

// destructureRE matches a binding of two or more names from one call: the
// pair a material assignment writes, `a, b = pkg.sym();`, or the fuller
// destructure Adjust writes, `a, b, c, d = pkg.ctor(x, y, z);`.
const destructureRE = /^\s*([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)+)\s*=\s*([^;]*);\s*$/;

// destructureOf reads such a line: the names it binds, in order, and the text
// it binds them from; null for any other line. The first name is the surface
// and the second its colour, which is how both the libraries' pairs and the
// Adjust templates are laid out.
function destructureOf(line) {
    const m = line.match(destructureRE);
    return m ? { names: m[1].split(",").map((n) => n.trim()), value: m[2].trim() } : null;
}

// callName is the name a destructure's call is known by: `metal.brass` for
// `metal.brass()`, `coating.powderCoat` for `coating.powderCoat(a, b, c)`, and
// the text itself when it is not a call.
function callName(value) {
    const open = value.indexOf("(");
    return open >= 0 && value.endsWith(")") ? value.slice(0, open).trim() : value;
}

function escapeName(s) {
    return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

// surfaceNames are the two values a material assignment binds, named after the
// body rather than after the material, so assigning a second material to the
// same body rewrites the same lines instead of leaving the first one's names
// behind as litter.
function surfaceNames(source, body) {
    let base = `${body}Surface`;
    let suffix = 2;
    while (new RegExp(`\\b${escapeName(base)}\\b`).test(source)) base = `${body}Surface${suffix++}`;
    return [base, `${base}Color`];
}

// codeOf strips a line comment and the contents of string literals, so that a
// `;` or a bracket inside either is not mistaken for syntax.
function codeOf(line) {
    let out = "";
    let quote = false;
    for (let i = 0; i < line.length; i++) {
        const c = line[i];
        if (quote) {
            if (c === "\\") i++;
            else if (c === '"') quote = false;
            continue;
        }
        if (c === '"') { quote = true; continue; }
        if (c === "/" && line[i + 1] === "/") break;
        out += c;
    }
    return out;
}

// findProperty locates an existing `body.prop = …;`, returning the inclusive
// line range it occupies and the text it assigns, or null.
//
// It is a statement scan rather than a line match, because an existing
// assignment is not reliably one tidy line. Both of these are in the examples:
//
//	plates.material = material.withRoughness(aluminumMat, 0.5);  // anodized
//	screwPH.material = material.withDisplacement(zincMat,
//	  remap(thread.maleRelief(…), …));
//
// Matching only `… = …;` to end of line misses both — and missing one is not
// harmless: the caller then believes the property is unset and writes a second
// assignment, which the compiler rejects with "property already assigned".
function findProperty(lines, body, prop) {
    const re = propertyStartRE(body, prop);
    const from = lines.findIndex((line) => re.test(line));
    if (from < 0) return null;

    let depth = 0;
    let value = "";
    for (let i = from; i < lines.length; i++) {
        let code = codeOf(lines[i]);
        // A continuation line's indentation is layout, not part of the value.
        code = i === from ? code.slice(code.indexOf("=") + 1) : code.trimStart();
        for (const ch of code) {
            if ("([{".includes(ch)) depth++;
            else if (")]}".includes(ch)) depth--;
            else if (ch === ";" && depth <= 0) return { from, to: i, value: value.trim() };
            value += ch;
        }
        value += " ";
    }
    return null; // unterminated; treat as absent rather than guess at its extent
}

// returnLine is the line the document's top-level return starts on, or the end
// of the file when it has none. Assignments go above it, because a value has to
// be bound before the return reads it.
function returnLine(source, lines) {
    const span = topLevelReturnSpan(source);
    if (!span) return lines.length;
    return source.slice(0, span.start).split("\n").length - 1;
}

// assignSurface writes `entry` onto the body named `body`, returning the new
// source. entry is one catalog row: { kind, package, symbol }.
//
// Re-assigning a body the panel has already written rewrites the call in place,
// so a user trying six materials in a row leaves one set of lines behind, not
// six.
//
// options.keepColor applies a material's surface alone, as a finish is applied:
// the material half of its pair goes on .material and .color is left as it is.
// options.colorLiteral is the body's colour as numbers when the panel knows it,
// which keepColor needs when that colour is the half of a pair it is about to
// rewrite (see assignMaterialOnly).
//
// options.keepDisplacement and options.keepLay carry the body's relief and its
// lay over to the new surface (see bindKept and wrapKept): a thread's relief
// survives a change of metal, a turned part keeps its grooves' direction.
export function assignSurface(source, body, entry, options = {}) {
    if (!isAssignableBody(source, body)) {
        throw new Error(`${body} is not a named body this document binds`);
    }
    const call = entry.kind === "finish"
        ? `${entry.package}.${entry.symbol}`
        : `${entry.package}.${entry.symbol}()`;
    let kept = null;
    if (options.keepDisplacement || options.keepLay) ({ source, kept } = bindKept(source, body));
    let next = entry.kind === "material" && options.keepColor
        ? assignMaterialOnly(source, body, call, options.colorLiteral)
        : assignWhole(source, body, entry, call);
    if (kept) next = wrapKept(next, body, kept, options);
    return pruneOrphanPairs(next, body);
}

// assignWhole is a finish on .material, or a material on both properties.
function assignWhole(source, body, entry, call) {
    const lines = source.split("\n");

    if (entry.kind === "finish") {
        const line = `${body}.material = ${call};`;
        const found = findProperty(lines, body, "material");
        if (found) {
            replaceStatement(lines, found, line);
            return lines.join("\n");
        }
        return insertBlock(source, lines, [line]);
    }

    // A material sets both properties through a destructure. If the body's
    // .material and .color already read the two names of one pair line of the
    // panel's own naming, that line is a previous assignment of ours: rewrite
    // its call and stop. A pair of any other name — an author's, or a kept
    // surface — is not the panel's to rewrite.
    const mat = findProperty(lines, body, "material");
    const col = findProperty(lines, body, "color");
    if (mat && col) {
        const pairIdx = lines.findIndex((line) => {
            const d = destructureOf(line);
            return d && d.names[0] === mat.value && d.names[1] === col.value && oursRE(body).test(d.names[0]);
        });
        if (pairIdx >= 0) return rewriteDestructure(lines, pairIdx, body, call);
    }

    // A colour picked since the panel wrote the pair leaves .material reading
    // the pair while .color holds the pick. A material brings its own colour,
    // so the pair's call is rewritten and .color goes back to reading it —
    // rather than writing a second pair beside the first. Only a pair of the
    // panel's own naming is taken over: a hand-written `steelMat, steelColor`
    // may be read by other bodies, which rewriting its call would change too.
    if (mat) {
        const pairIdx = lines.findIndex((line) => {
            const d = destructureOf(line);
            return d && d.names[0] === mat.value && oursRE(body).test(d.names[0]);
        });
        if (pairIdx >= 0) {
            const colName = destructureOf(lines[pairIdx]).names[1];
            const colLine = `${body}.color = ${colName};`;
            if (col) replaceStatement(lines, col, colLine);
            else lines.splice(mat.to + 1, 0, colLine);
            return rewriteDestructure(lines, pairIdx, body, call);
        }
    }

    // Otherwise write a fresh set: the pair, then the two properties. Whatever
    // is already there is replaced where it stands, and only what is missing is
    // inserted — in ONE block, so the three lines stay together. Inserting them
    // one at a time would pad between them and split the block up.
    const [matName, colName] = surfaceNames(source, body);
    const pairLine = `${matName}, ${colName} = ${call};`;
    const matLine = `${body}.material = ${matName};`;
    const colLine = `${body}.color = ${colName};`;

    // Replace what is already there, the LATER statement first: a replacement
    // collapses a range of lines to one and so moves everything below it.
    const present = [
        mat && { range: mat, line: matLine },
        col && { range: col, line: colLine },
    ].filter(Boolean).sort((a, b) => b.range.from - a.range.from);
    for (const { range, line } of present) replaceStatement(lines, range, line);

    const block = [pairLine];
    if (!mat) block.push(matLine);
    if (!col) block.push(colLine);

    if (present.length === 0) return insertBlock(source, lines, block);
    // The destructure has to be bound before anything reads it, so it goes
    // above the first of the statements that survived. That one's position is
    // still good: only lines below it were rewritten.
    lines.splice(Math.min(...present.map((p) => p.range.from)), 0, ...block);
    return lines.join("\n");
}

// assignColor writes `rgb` as the body's colour, returning the new source.
//
// A colour is sRGB in [0, 1], the hex code over 255, which is what the colour
// library's constants are and what the viewer decodes to linear light. Three
// decimals round-trip every 8-bit channel. The colour replaces whatever .color
// was — a literal, a library constant, or the colour half of a material's
// pair — and leaves .material alone, so a material keeps its roughness and
// metalness under a colour chosen for it.
export function assignColor(source, body, rgb) {
    if (!isAssignableBody(source, body)) {
        throw new Error(`${body} is not a named body this document binds`);
    }
    const line = `${body}.color = ${colorLiteralText(rgb)};`;
    const lines = source.split("\n");
    const found = findProperty(lines, body, "color");
    if (found) {
        replaceStatement(lines, found, line);
        return lines.join("\n");
    }
    return insertBlock(source, lines, [line]);
}

function colorLiteralText(rgb) {
    const channel = (v) => String(Number(Math.max(0, Math.min(1, v)).toFixed(3)));
    return `[${rgb.map(channel).join(", ")}]`;
}

// rewriteDestructure points the panel's destructure at line idx at a new
// call, returning the new source. A pair stays a pair with its call swapped.
// The fuller destructure Adjust wrote collapses back to the pair, since a
// preset returns two values, and the parameter variables only its call read
// go with it, so a preset tried after an adjustment leaves no sliders behind.
function rewriteDestructure(lines, idx, body, call) {
    const d = destructureOf(lines[idx]);
    const lifted = liftedNames(d.value, body);
    lines[idx] = `${d.names[0]}, ${d.names[1]} = ${call};`;
    return removeUnreadDeclarations(lines.join("\n"), lifted);
}

// expandSurface opens the preset on `body` up into its parameters, returning
// the new source. entry is the preset's catalog row and adjust its adjust
// record: the constructors it calls, their annotated parameters with the
// preset's own arguments as defaults, and the preset's body as a template
// (tool/materialswatch, adjust.go).
//
// One top-level variable is written per parameter, named `<body><Param>`
// (boltCoatColor, boltGlossUnits), each with the control its kind declares:
// a scalar with .min and .max for the slider, @color for the chip row, @point,
// @direction and @transform for the gizmos. The template follows, filled in
// for the body, in place of the one-line form the panel wrote. From here the
// Parameters panel and the viewer take over; this edit is made once.
export function expandSurface(source, body, entry, adjust) {
    if (!isAssignableBody(source, body)) {
        throw new Error(`${body} is not a named body this document binds`);
    }
    if (!adjust?.parameters?.length || !adjust.template) {
        throw new Error(`${entry.package}.${entry.symbol} has nothing to adjust`);
    }
    const lines = source.split("\n");

    // Each parameter's variable, fresh against the document and each other.
    const names = {};
    let taken = source;
    for (const p of adjust.parameters) {
        names[p.name] = freshName(taken, `${body}${p.name[0].toUpperCase()}${p.name.slice(1)}`);
        taken += ` ${names[p.name]}`;
    }
    const block = [
        ...adjust.parameters.flatMap((p) => declarationLines(p, names)),
        ...adjust.template.split("\n").map((line) =>
            line.replace(/\$\{(\w+)\}/g, (m, key) => (key === "body" ? body : names[key] ?? m))),
    ];

    // The lines the block replaces: the panel's pair and the two properties.
    // What is there is removed where it stands and the block goes in at the
    // first of those places, so it lands where the surface was written.
    const mat = findProperty(lines, body, "material");
    const col = findProperty(lines, body, "color");
    const inner = mat ? unwrapKept(mat.value, body).inner : null;
    const pairIdx = inner === null ? -1 : lines.findIndex((line) => {
        const d = destructureOf(line);
        return d && d.names[0] === inner && oursRE(body).test(d.names[0]);
    });
    const ranges = [];
    if (pairIdx >= 0) ranges.push({ from: pairIdx, to: pairIdx });
    if (mat) ranges.push(mat);
    if (col) ranges.push(col);
    if (ranges.length === 0) return insertBlock(source, lines, block);
    ranges.sort((a, b) => b.from - a.from);
    const at = ranges[ranges.length - 1].from;
    for (const r of ranges) lines.splice(r.from, r.to - r.from + 1);
    lines.splice(at, 0, ...block);
    return pruneOrphanPairs(lines.join("\n"), body);
}

// declarationLines writes one parameter's variable with the control its kind
// declares. The value is the preset's own argument, so the document draws
// exactly as it did before the expansion.
function declarationLines(p, names) {
    const v = names[p.name];
    const vec = (xs) => `[${xs.map(numberText).join(", ")}]`;
    switch (p.kind) {
        case "range":
            return [`${v} = ${numberText(p.default[0])};`, `${v}.min = ${numberText(p.min)};`, `${v}.max = ${numberText(p.max)};`];
        case "color":
            return [`@color ${v} = ${vec(p.default)};`];
        case "point":
            return [`@point ${v} = ${vec(p.default)};`];
        case "direction": {
            // Drawn from the point parameter the constructor anchors it to,
            // when that was lifted too; from the origin otherwise.
            const anchor = p.anchor && names[p.anchor] ? `(${names[p.anchor]})` : "";
            return [`@direction${anchor} ${v} = ${vec(p.default)};`];
        }
        case "transform":
            // A matrix is a list of columns.
            return [`@transform ${v} = [${[0, 4, 8, 12].map((i) => vec(p.default.slice(i, i + 4))).join(", ")}];`];
        default:
            return [`${v} = ${p.default.length === 1 ? numberText(p.default[0]) : vec(p.default)};`];
    }
}

// numberText writes a number as GSL: six decimals at most, no trailing zeros.
function numberText(v) {
    return String(Number(Number(v).toFixed(6)));
}

// liftedNames are the variables Adjust wrote that a destructure's call reads:
// the identifiers of the panel's `<body><Param>` naming among its arguments.
function liftedNames(value, body) {
    const open = value.indexOf("(");
    if (open < 0) return [];
    const re = new RegExp(`\\b${escapeName(body)}[A-Z]\\w*\\b`, "g");
    return [...new Set(value.slice(open).match(re) || [])];
}

// removeUnreadDeclarations drops the top-level declarations of `names` that
// nothing reads any more, with their .min and .max lines, until none is left
// to drop: a direction's declaration reads its anchor's, so the anchor goes
// on the pass after the direction.
function removeUnreadDeclarations(source, names) {
    let lines = source.split("\n");
    let pending = new Set(names);
    for (let changed = true; changed && pending.size;) {
        changed = false;
        for (const name of [...pending]) {
            const own = (line) => declaresOrSets(line, name);
            const readElsewhere = lines.some((line) => !own(line) && new RegExp(`\\b${escapeName(name)}\\b`).test(codeOf(line)));
            if (readElsewhere) continue;
            lines = lines.filter((line) => !own(line));
            pending.delete(name);
            changed = true;
        }
    }
    return lines.join("\n");
}

// declaresOrSets: the line is `name = …`, with or without an annotation in
// front, or `name.min = …` / `name.max = …`.
function declaresOrSets(line, name) {
    const n = escapeName(name);
    return new RegExp(`^(@\\w+(\\([^)]*\\))?\\s+)?${n}\\s*=(?!=)`).test(line) ||
        new RegExp(`^\\s*${n}\\.(min|max)\\s*=`).test(line);
}

// declarationRE matches the top-level declaration of `name`, with or without
// an annotation in front, capturing the annotation prefix and the value.
function declarationRE(name) {
    return new RegExp(`^(@\\w+(?:\\([^)]*\\))?\\s+)?${escapeName(name)}\\s*=\\s*([^;]*);\\s*$`);
}

// callArguments splits a call's arguments at its top-level commas: the text
// inside the outermost parentheses, each argument trimmed.
function callArguments(value) {
    const open = value.indexOf("(");
    if (open < 0 || !value.endsWith(")")) return [];
    const inner = value.slice(open + 1, -1);
    const out = [];
    let depth = 0;
    let start = 0;
    for (let i = 0; i < inner.length; i++) {
        const c = inner[i];
        if ("([{".includes(c)) depth++;
        else if (")]}".includes(c)) depth--;
        else if (c === "," && depth === 0) {
            out.push(inner.slice(start, i).trim());
            start = i + 1;
        }
    }
    if (inner.trim() !== "") out.push(inner.slice(start).trim());
    return out;
}

// adjustedColorOf reads the coat colour of an adjusted body: the literal the
// document declares for the one @color parameter of a preset's adjust record,
// found through the constructor call the body's surface reads. Returns the
// variable's name and its colour as numbers, or null: the record has no
// single colour parameter, the body does not read that constructor, or the
// parameter's argument is not a variable declared as a literal.
export function adjustedColorOf(source, body, adjust) {
    const colours = (adjust?.parameters || []).filter((p) => p.kind === "color");
    if (colours.length !== 1) return null;
    const [param] = colours;
    const lines = source.split("\n");

    // The calls the body's surface and colour read: a destructure's, or the
    // property's own text.
    const calls = [];
    for (const prop of ["material", "color"]) {
        const found = findProperty(lines, body, prop);
        if (!found) continue;
        const inner = prop === "material" ? unwrapKept(found.value, body).inner : found.value;
        let value = inner;
        for (const line of lines) {
            const d = destructureOf(line);
            if (d && d.names.includes(inner)) { value = d.value; break; }
        }
        calls.push(value);
    }
    const value = calls.find((v) => callName(v) === param.call);
    if (!value) return null;
    const arg = callArguments(value)[param.arg];
    if (!arg || !/^[A-Za-z_]\w*$/.test(arg)) return null;
    const decl = lines.map((line) => line.match(declarationRE(arg))).find(Boolean);
    const rgb = decl ? parseColorLiteral(decl[2].trim()) : null;
    return rgb ? { name: arg, rgb } : null;
}

// assignAdjustedColor writes `rgb` as the literal of the colour variable
// `name` declares, keeping its annotation, so an adjusted body's coat colour
// can be picked from the Bodies chip as well as the Parameters panel.
export function assignAdjustedColor(source, name, rgb) {
    const lines = source.split("\n");
    const at = lines.findIndex((line) => declarationRE(name).test(line));
    if (at < 0) throw new Error(`${name} is not a variable this document declares`);
    const prefix = lines[at].match(declarationRE(name))[1] || "";
    lines[at] = `${prefix}${name} = ${colorLiteralText(rgb)};`;
    return lines.join("\n");
}

// ── Operators: a stack of layers on a body's surface ──
//
// An operator takes a surface and returns it modified, so in the document it
// is a wrapping edit, not the replacing edit a swatch makes: a layer line binds
// the operator's result and the body's properties read the layer instead of
// what it wraps. The catalog's operators (tool/materialswatch, operators.go)
// are @surface presets, overloads of each operator taking only the slots a
// document supplies: the substrate material, its colour, the body's frame
// (doc/material_composition.md, "Design: operators as presets").

// The material a body has when its document sets none; what a wrap wraps then.
const DEFAULT_MATERIAL = "material.standard";

// LAYER_SUFFIXES are the names a layer binds after its base, in the order the
// operators return their values.
const LAYER_SUFFIXES = ["", "Color", "Height", "Slope"];

// layerBase is the name the panel binds a layer to, <body><Op>: boltTurned.
function layerBase(body, opCall) {
    const op = opCall.slice(opCall.indexOf(".") + 1);
    return `${body}${op[0].toUpperCase()}${op.slice(1)}`;
}

// slotCount is how many arguments the preset form of an operator takes.
function slotCount(op) {
    return 1 + (op.slots.color >= 0 ? 1 : 0) + (op.slots.frame >= 0 ? 1 : 0);
}

// bindingOf reads a line binding one or more names from one value: the single
// binding an arity-1 layer writes, `a = material.withRoughness(b);`, as well as
// a destructure. Null for any other line.
function bindingOf(line) {
    const m = line.match(/^\s*([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)*)\s*=(?!=)\s*([^;]*);\s*$/);
    return m ? { names: m[1].split(",").map((n) => n.trim()), value: m[2].trim() } : null;
}

// bindingLineOf finds the line binding `name` first, with what it binds.
function bindingLineOf(lines, name) {
    const at = lines.findIndex((line) => {
        const b = bindingOf(line);
        return b && b.names[0] === name;
    });
    return at < 0 ? null : { at, ...bindingOf(lines[at]) };
}

// stackOf reads the body's stack of operators: from .material, each layer whose
// call is a catalog operator, by preset or by its full form, down to the base
// that is not one. Layers come innermost first. Each has the name it binds,
// the operator's catalog key, its arguments, and whether it is still the
// preset form (one argument per slot), which is what Adjust can open.
// baseSurface is the base named as surfaceOf names a surface: the call behind a
// pair, else the text itself; null when the body sets no material.
export function stackOf(source, body, operators) {
    const lines = source.split("\n");
    const mat = findProperty(lines, body, "material");
    if (!mat) return { base: null, baseSurface: null, layers: [] };
    let current = unwrapKept(mat.value, body).inner;
    const layers = [];
    const seen = new Set();
    while (/^[A-Za-z_]\w*$/.test(current) && !seen.has(current)) {
        seen.add(current);
        const bind = bindingLineOf(lines, current);
        const op = bind ? operators?.[callName(bind.value)] : null;
        if (!op) break;
        const args = callArguments(bind.value);
        layers.unshift({ name: current, names: bind.names, op: callName(bind.value), call: bind.value, args, preset: args.length === slotCount(op) });
        current = args[op.slots.material] ?? "";
    }
    const pair = bindingLineOf(lines, current);
    return { base: current, baseSurface: pair ? callName(pair.value) : current, layers };
}

// wrapSurface wraps the body's surface with the operator preset op, returning
// the new source: one layer line binding what the preset returns, and the
// body's properties repointed at it. The material slot is whatever .material
// read, as text, or the default material when it read nothing; the colour
// slot is .color's text, or options.colorLiteral (the colour the body is drawn
// in) when the body sets none; the frame slot is the body's own frame. An
// operator returning the material alone leaves .color as it is.
export function wrapSurface(source, body, op, options = {}) {
    if (!isAssignableBody(source, body)) {
        throw new Error(`${body} is not a named body this document binds`);
    }
    const lines = source.split("\n");
    const mat = findProperty(lines, body, "material");
    const col = findProperty(lines, body, "color");

    const args = new Array(slotCount(op));
    args[op.slots.material] = mat ? mat.value : DEFAULT_MATERIAL;
    if (op.slots.color >= 0) {
        args[op.slots.color] = col ? col.value : colorLiteralText(options.colorLiteral || [0.5, 0.5, 0.5]);
    }
    if (op.slots.frame >= 0) args[op.slots.frame] = `[${body}.u, ${body}.v]`;

    const base = freshName(source, layerBase(body, op.call));
    const names = LAYER_SUFFIXES.slice(0, op.arity).map((s) => `${base}${s}`);
    const layerLine = `${names.join(", ")} = ${op.call}(${args.join(", ")});`;
    const matLine = `${body}.material = ${names[0]};`;
    const colLine = op.arity >= 2 ? `${body}.color = ${names[1]};` : null;

    // Replace the properties where they stand, the later first, and put the
    // layer line above the first of them; with neither present the block goes
    // above the return, as an assignment does.
    const present = [
        mat && { range: mat, line: matLine },
        colLine && col && { range: col, line: colLine },
    ].filter(Boolean).sort((a, b) => b.range.from - a.range.from);
    for (const { range, line } of present) replaceStatement(lines, range, line);
    const block = [layerLine];
    if (!mat) block.push(matLine);
    if (colLine && !col) block.push(colLine);
    if (present.length === 0) return insertBlock(source, lines, block);
    lines.splice(Math.min(...present.map((p) => p.range.from)), 0, ...block);
    return lines.join("\n");
}

// expandLayer opens the layer bound to layerName up into the operator's
// parameters, as expandSurface opens a preset: one variable per parameter
// above the layer, and the layer's line rewritten as the operator's full call
// from the catalog template, its slots carried over from the preset call.
export function expandLayer(source, body, layerName, op) {
    const lines = source.split("\n");
    const bind = bindingLineOf(lines, layerName);
    if (!bind || callName(bind.value) !== op.call) {
        throw new Error(`${layerName} is not a layer of ${op.call} this document binds`);
    }
    const args = callArguments(bind.value);
    if (args.length !== slotCount(op)) throw new Error(`${layerName} is already opened up`);
    if (!op.adjust?.parameters?.length || !op.adjust.template) {
        throw new Error(`${op.call} has nothing to adjust`);
    }

    const names = {};
    let taken = source;
    for (const p of op.adjust.parameters) {
        names[p.name] = freshName(taken, `${body}${p.name[0].toUpperCase()}${p.name.slice(1)}`);
        taken += ` ${names[p.name]}`;
    }
    const slots = {
        material: args[op.slots.material],
        color: op.slots.color >= 0 ? args[op.slots.color] : "",
        frame: op.slots.frame >= 0 ? args[op.slots.frame] : "",
    };
    const fill = (line) => line.replace(/\$\{(\w+)\}/g, (m, key) =>
        key === "body" ? body : key === "layer" ? layerName : key in slots ? slots[key] : names[key] ?? m);
    // The template's property lines already stand in the document, reading
    // the layer; only its binding lines replace the preset's one.
    const bindings = op.adjust.template.split("\n").filter((line) => !line.startsWith("${body}."));
    const block = [...op.adjust.parameters.flatMap((p) => declarationLines(p, names)), ...bindings.map(fill)];
    lines.splice(bind.at, 1, ...block);
    return lines.join("\n");
}

// removeLayer takes the layer bound to layerName out of the body's stack,
// returning the new source. Whatever read the layer, the body's properties or
// the layer above, reads the layer's own slots instead: its material argument
// and its colour argument. A coat has no colour slot, so the colour it
// replaced is given back as the substrate pair's when the substrate is a pair
// and dropped otherwise. The layer's line goes, and with it the parameter
// variables only its call read.
export function removeLayer(source, body, layerName, op) {
    let lines = source.split("\n");
    const bind = bindingLineOf(lines, layerName);
    if (!bind || callName(bind.value) !== op.call) {
        throw new Error(`${layerName} is not a layer of ${op.call} this document binds`);
    }
    const args = callArguments(bind.value);
    const materialText = args[op.slots.material];
    let colorText = op.slots.color >= 0 ? args[op.slots.color] : null;
    if (colorText === null) {
        const pair = bindingLineOf(lines, materialText);
        colorText = pair && pair.names.length >= 2 ? pair.names[1] : null;
    }
    const lifted = liftedNames(bind.value, body);
    lines.splice(bind.at, 1);

    const word = (name) => new RegExp(`\\b${escapeName(name)}\\b`, "g");
    const [matName, colName] = bind.names;
    lines = lines.flatMap((line) => {
        let next = line;
        if (word(matName).test(codeOf(next))) next = next.replace(word(matName), materialText);
        if (colName && word(colName).test(codeOf(next))) {
            if (colorText) next = next.replace(word(colName), colorText);
            else if (propertyStartRE(body, "color").test(next)) return [];
            else next = next.replace(word(colName), colorLiteralText([0.5, 0.5, 0.5]));
        }
        return [next];
    });
    return removeUnreadDeclarations(lines.join("\n"), lifted);
}

// surfaceOf reads what the document gives the body: the surface it is
// assigned, as the call that made it when that was a material's pair
// (`metal.brass`) and as the assigned text otherwise (`material.matte`), and
// its colour, as text and, when the colour is a literal, as numbers. Either is
// null when the document does not set it. colorIsSurface is true when the
// colour is the material's own, the colour half of the same pair.
//
// colorSource is the call whose colour the body wears, whenever .color reads
// the colour half of a pair: the same pair as the material's, or another —
// the one a finish was put over, say, which still supplies the colour after
// .material has moved on. The panel looks that call's colours up, so a body
// in brass's colour under a matte finish shows brass's colour, not a guess.
export function surfaceOf(source, body) {
    const lines = source.split("\n");
    const mat = findProperty(lines, body, "material");
    const col = findProperty(lines, body, "color");

    // Seen through the panel's own wrapping of a kept surface, the surface is
    // the one inside it, and what it keeps from the kept one is listed.
    const { inner, keeps } = mat ? unwrapKept(mat.value, body) : { inner: null, keeps: [] };
    let surface = inner;
    let colorIsSurface = false;
    if (mat) {
        for (const line of lines) {
            const d = destructureOf(line);
            if (d && d.names[0] === inner) {
                surface = callName(d.value);
                colorIsSurface = !!col && col.value === d.names[1];
                break;
            }
        }
    }
    let colorSource = null;
    if (col) {
        for (const line of lines) {
            const d = destructureOf(line);
            if (d && d.names[1] === col.value) {
                colorSource = callName(d.value);
                break;
            }
        }
    }
    return { surface, color: col ? col.value : null, rgb: col ? parseColorLiteral(col.value) : null, colorIsSurface, colorSource, keeps };
}

function parseColorLiteral(text) {
    const num = "\\s*(-?\\d*\\.?\\d+(?:[eE][-+]?\\d+)?)\\s*";
    const m = text.match(new RegExp(`^\\[${num},${num},${num}\\]$`));
    return m ? [Number(m[1]), Number(m[2]), Number(m[3])] : null;
}

// oursRE matches the names the panel binds a body's pairs to, boltSurface,
// boltSurface2, …: the only pairs it may rewrite or remove, since one an author
// wrote may be read by other bodies.
function oursRE(body) {
    return new RegExp(`^${escapeName(body)}Surface\\d*$`);
}

// keptRE matches the names the panel binds a body's kept surface to: boltKept,
// boltKept2, ….
function keptRE(body) {
    return new RegExp(`^${escapeName(body)}Kept\\d*$`);
}

function freshName(source, base) {
    let name = base;
    for (let n = 2; new RegExp(`\\b${escapeName(name)}\\b`).test(source); n++) name = `${base}${n}`;
    return name;
}

// The wrappers wrapKept writes, and unwraps when it meets its own work again.
function wrapDisplacement(inner, kept) {
    return `material.withDisplacement(${inner}, material.displacement(${kept}))`;
}
function wrapLay(inner, kept) {
    return `material.withLay(${inner}, material.lay(${kept}), material.anisotropy(${kept}))`;
}

// bindKept names the surface whose displacement and lay are to be kept, so the
// assignment that follows can replace .material without losing them. It
// returns the new source and the name, or a null name when the body has no
// surface to keep.
//
//   - .material is the panel's own wrapping of a kept surface: the same kept
//     surface stays, and .material is unwrapped to the surface inside, which
//     the assignment then replaces as it would any.
//   - .material reads one of the panel's pairs: the pair is renamed to the
//     kept name, so the assignment writes a new pair rather than rewriting
//     the call being kept — and a colour reading that pair goes on reading it.
//   - Anything else — an author's own expression, a thread's relief, a
//     finish: it is bound to the kept name, `boltKept = <expression>;`. That
//     binding is the author's work, so it is never removed.
function bindKept(source, body) {
    const lines = source.split("\n");
    const mat = findProperty(lines, body, "material");
    if (!mat) return { source, kept: null };

    const unwrapped = unwrapKept(mat.value, body);
    if (unwrapped.kept) {
        replaceStatement(lines, mat, `${body}.material = ${unwrapped.inner};`);
        return { source: lines.join("\n"), kept: unwrapped.kept };
    }

    const kept = freshName(source, `${body}Kept`);
    if (oursRE(body).test(mat.value)) {
        const pairIdx = lines.findIndex((line) => {
            const d = destructureOf(line);
            return d && d.names[0] === mat.value;
        });
        if (pairIdx >= 0) {
            const colName = destructureOf(lines[pairIdx]).names[1];
            const renamed = lines.map((line) => line
                .replace(new RegExp(`\\b${escapeName(mat.value)}\\b`, "g"), kept)
                .replace(new RegExp(`\\b${escapeName(colName)}\\b`, "g"), `${kept}Color`));
            return { source: renamed.join("\n"), kept };
        }
    }
    replaceStatement(lines, mat, `${body}.material = ${kept};`);
    lines.splice(mat.from, 0, `${kept} = ${mat.value};`);
    return { source: lines.join("\n"), kept };
}

// unwrapKept undoes wrapKept: the surface inside the panel's wrapping, the kept
// surface's name, and what is kept from it. A value the panel did not wrap
// comes back as it is, with no kept name.
function unwrapKept(value, body) {
    const named = value.match(new RegExp(`\\b(${escapeName(body)}Kept\\d*)\\b`));
    if (!named) return { inner: value, kept: null, keeps: [] };
    const kept = named[1];
    let inner = value;
    const keeps = [];
    for (let changed = true; changed;) {
        changed = false;
        for (const [wrap, what] of [[wrapLay, "lay"], [wrapDisplacement, "displacement"]]) {
            const [pre, post] = wrap("\u0000", kept).split("\u0000");
            if (inner.startsWith(pre) && inner.endsWith(post)) {
                inner = inner.slice(pre.length, inner.length - post.length);
                keeps.unshift(what);
                changed = true;
            }
        }
    }
    // A kept name that is not the panel's wrapping is an author's expression
    // that happens to read it; it is left whole.
    return keeps.length ? { inner, kept, keeps } : { inner: value, kept: null, keeps: [] };
}

// wrapKept carries the kept surface's displacement and lay over to whatever
// .material now is.
function wrapKept(source, body, kept, options) {
    const lines = source.split("\n");
    const mat = findProperty(lines, body, "material");
    if (!mat) return source;
    let value = mat.value;
    if (options.keepDisplacement) value = wrapDisplacement(value, kept);
    if (options.keepLay) value = wrapLay(value, kept);
    replaceStatement(lines, mat, `${body}.material = ${value};`);
    return lines.join("\n");
}

// assignMaterialOnly puts a material's surface on the body and keeps its
// colour.
//
// The one case that needs care is a colour that is itself the colour half of
// the pair the panel is about to rewrite — `bolt.color = boltSurfaceColor` after
// a material was applied whole. Rewriting that pair's call would change the
// colour too. So the colour is pinned first: to colorLiteral, when the panel
// knows it as numbers (a uniform catalog material); otherwise — a patterned
// colour, which numbers cannot hold — the old pair is left to go on supplying
// the colour, and the new surface gets a pair of its own.
function assignMaterialOnly(source, body, call, colorLiteral) {
    const lines = source.split("\n");
    const ours = oursRE(body);
    const pairOf = (name) => lines.findIndex((line) => {
        const d = destructureOf(line);
        return d && d.names[0] === name && ours.test(d.names[0]);
    });

    let mat = findProperty(lines, body, "material");
    if (mat && pairOf(mat.value) >= 0) {
        const colName = destructureOf(lines[pairOf(mat.value)]).names[1];
        const col = findProperty(lines, body, "color");
        const colourIsThePair = col && col.value === colName;
        if (colourIsThePair && colorLiteral) {
            replaceStatement(lines, col, `${body}.color = ${colorLiteralText(colorLiteral)};`);
        }
        if (!colourIsThePair || colorLiteral) {
            const idx = pairOf(findProperty(lines, body, "material").value);
            return rewriteDestructure(lines, idx, body, call);
        }
    }

    // A pair of its own, of which only the material half is read.
    const [matName, colName] = surfaceNames(source, body);
    const pairLine = `${matName}, ${colName} = ${call};`;
    const matLine = `${body}.material = ${matName};`;
    mat = findProperty(lines, body, "material");
    if (!mat) return insertBlock(source, lines, [pairLine, matLine]);
    replaceStatement(lines, mat, matLine);
    lines.splice(mat.from, 0, pairLine);
    return lines.join("\n");
}

// pruneOrphanPairs removes the body's panel-written pair lines that nothing
// reads any more — the line a pinned colour or a second pair left behind, a
// kept catalog surface no longer kept — so trying surfaces leaves no litter. A
// single kept binding, `boltKept = <expression>;`, is an author's own work and
// stays.
function pruneOrphanPairs(source, body) {
    const lines = source.split("\n");
    const ours = oursRE(body);
    const readElsewhere = (name, at) => lines.some((line, i) =>
        i !== at && new RegExp(`\\b${escapeName(name)}\\b`).test(codeOf(line)));
    const keptPair = keptRE(body);
    // The parameter variables a dropped destructure read go with it, when
    // nothing else reads them.
    let lifted = [];
    const remaining = lines.filter((line, i) => {
        const d = destructureOf(line);
        if (!d || !(ours.test(d.names[0]) || keptPair.test(d.names[0]))) return true;
        if (d.names.some((name) => readElsewhere(name, i))) return true;
        lifted = lifted.concat(liftedNames(d.value, body));
        return false;
    });
    return removeUnreadDeclarations(remaining.join("\n"), lifted);
}

// replaceStatement swaps the lines a statement occupies for a single line.
function replaceStatement(lines, range, line) {
    lines.splice(range.from, range.to - range.from + 1, line);
}

// insertBlock puts `block` above the document's top-level return, separated
// from what comes before by a single blank line.
function insertBlock(source, lines, block) {
    const next = lines.slice();
    const at = returnLine(source, next);
    const pad = at > 0 && next[at - 1].trim() !== "" ? ["", ...block] : block;
    next.splice(at, 0, ...pad);
    return next.join("\n");
}
