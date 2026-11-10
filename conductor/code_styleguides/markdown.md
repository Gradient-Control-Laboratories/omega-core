# Markdown Style Guide

This document outlines guidelines and principles for writing Markdown documentation in this repository.

## Path & Link Formatting

- Markdown files should always contain relative paths and never contain absolute paths.
- Never use machine-specific absolute paths (such as `/Users/...` or `file:///...`) in markdown files.

## Line Length & Wrapping

- Prefer lines in markdown documents to be at most 100 characters.
- Keep each sentence in a paragraph fully intact on a single line if it is under 100 characters.
- Do not wrap in the middle of a sentence unless the sentence itself exceeds 120 characters.
- If a paragraph contains multiple sentences, prefer wrapping exactly at sentence boundaries (after
  the period).
- If wrapping within a sentence is necessary, wrap at a logical boundary
  such as before "then", "also", "along with", or "such as".

## Whitespace & Blank Lines

- Never remove a single blank line at the start of a markdown file, in between non-blank lines, or
  at the end of a markdown file.
- Always replace multiple blank lines with a single blank line in markdown files.
- Always ensure that there is exactly one blank line at the end of a markdown document.

## Diagrams (ASCII Art vs. Mermaid)

- Prefer compact ASCII art flowcharts over Mermaid for simple, linear decision flows,
  state transitions, and directory/file trees.
- Compact ASCII diagrams MUST NOT have rectangular boxes (`+---+`, `|   |`);
  use clean text labels, branch tags (`[Yes]`, `[No]`), and directional arrows (`-->`, `|`, `v`).
- Reserve Mermaid diagrams (rendered to PNG via `mmdc` in a relative `images/` directory
  with collapsible `<details>` source below) for complicated, involved diagrams
  (e.g., multi-service topologies, complex concurrent pipelines, large class/ER diagrams).
- When creating Mermaid diagrams, embed the rendered PNG image using standard Markdown syntax:
  `![Diagram Title](images/diagram.png)`.
- Include the original Mermaid diagram source in a collapsible `<details>` and `<summary>` block
  immediately below the image embed.
- Do NOT use HTML comments (`<!-- ... -->`) around Mermaid code blocks to avoid syntax conflicts
  with Mermaid's `-->` directed arrow operator.

## Tables

- Format tables with aligned columns padded with spaces so that vertical pipes (`|`)
  line up vertically across header rows, separator rows, and data rows.
- Aligning columns ensures tables are easily readable as plain text (ASCII) in diffs
  and code editors, as well as when rendered in Markdown viewers.
- Separator rows should match column widths using dashes (`:---...--- |`).
- Table rows are exempt from the standard 100-character line length limit when column
  content requires additional width to maintain proper alignment.
