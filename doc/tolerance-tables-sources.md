# Process-to-tolerance reference tables (sourced)

Compiled 2026-09-30. Every value below was read from the URL cited under its table. Units are exactly as the source gives them. Where a value was *computed* (only the fit clearances in 1.4), that is labelled. Nothing was interpolated or filled from memory.

Source-quality key used below:
- **ISO preview**: the ISO standard's own free preview pages (iTeh / SIS), i.e. the standard's text as published. Primary.
- **Full-text copy**: a complete copy of the standard re-hosted on a third-party site (with a licensee watermark). The content is the standard's own; the hosting is unofficial.
- **Gov handbook**: MIL-HDBK-727, a US DoD public handbook. Primary for its own claims.
- **University**: course material that reproduces a standard's chart.
- **Secondary / vendor**: engineering reference sites or manufacturer pages.

---

## 1. ISO 286 limits and fits

### 1.1 Standard tolerance grades IT5–IT16 (ISO 286-1:2010 Table 1, reproduced as ISO 286-2:2010 Table 1)

IT5–IT11 are in **µm** and IT12–IT16 are in **mm**, as the source splits them. The sizes are "above / up to and including" in mm, and the first row is labelled "— 3" in the source.

| Nominal size (mm) | IT5 µm | IT6 µm | IT7 µm | IT8 µm | IT9 µm | IT10 µm | IT11 µm | IT12 mm | IT13 mm | IT14 mm | IT15 mm | IT16 mm |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| — to 3 | 4 | 6 | 10 | 14 | 25 | 40 | 60 | 0.1 | 0.14 | 0.25 | 0.4 | 0.6 |
| >3 to 6 | 5 | 8 | 12 | 18 | 30 | 48 | 75 | 0.12 | 0.18 | 0.3 | 0.48 | 0.75 |
| >6 to 10 | 6 | 9 | 15 | 22 | 36 | 58 | 90 | 0.15 | 0.22 | 0.36 | 0.58 | 0.9 |
| >10 to 18 | 8 | 11 | 18 | 27 | 43 | 70 | 110 | 0.18 | 0.27 | 0.43 | 0.7 | 1.1 |
| >18 to 30 | 9 | 13 | 21 | 33 | 52 | 84 | 130 | 0.21 | 0.33 | 0.52 | 0.84 | 1.3 |
| >30 to 50 | 11 | 16 | 25 | 39 | 62 | 100 | 160 | 0.25 | 0.39 | 0.62 | 1 | 1.6 |
| >50 to 80 | 13 | 19 | 30 | 46 | 74 | 120 | 190 | 0.3 | 0.46 | 0.74 | 1.2 | 1.9 |
| >80 to 120 | 15 | 22 | 35 | 54 | 87 | 140 | 220 | 0.35 | 0.54 | 0.87 | 1.4 | 2.2 |
| >120 to 180 | 18 | 25 | 40 | 63 | 100 | 160 | 250 | 0.4 | 0.63 | 1 | 1.6 | 2.5 |
| >180 to 250 | 20 | 29 | 46 | 72 | 115 | 185 | 290 | 0.46 | 0.72 | 1.15 | 1.85 | 2.9 |
| >250 to 315 | 23 | 32 | 52 | 81 | 130 | 210 | 320 | 0.52 | 0.81 | 1.3 | 2.1 | 3.2 |

Source (ISO preview, primary): ISO 286-2:2010 preview, Table 1 ("taken from ISO 286-1:2010"): https://cdn.standards.iteh.ai/samples/54915/7d6147c29fed4e31af36b60561e88752/ISO-286-2-2010.pdf
Cross-check (secondary): Wikipedia reproduces ISO 286 Table 1. The rows I spot-checked (up to 30 mm) agree: https://en.wikipedia.org/wiki/IT_Grade

### 1.2 Fundamental deviations: shafts c, f, g, h (upper deviation es, µm) (ISO 286-1:2010 Table 4)

c is included because it is needed for H11/c11.

| Nominal size (mm) | c | f | g | h |
|---|---|---|---|---|
| — to 3 | −60 | −6 | −2 | 0 |
| >3 to 6 | −70 | −10 | −4 | 0 |
| >6 to 10 | −80 | −13 | −5 | 0 |
| >10 to 18 (10–14, 14–18) | −95 | −16 | −6 | 0 |
| >18 to 30 (18–24, 24–30) | −110 | −20 | −7 | 0 |
| >30 to 40 | −120 | −25 | −9 | 0 |
| >40 to 50 | −130 | −25 | −9 | 0 |
| >50 to 65 | −140 | −30 | −10 | 0 |
| >65 to 80 | −150 | −30 | −10 | 0 |
| >80 to 100 | −170 | −36 | −12 | 0 |
| >100 to 120 | −180 | −36 | −12 | 0 |
| >120 to 140 | −200 | −43 | −14 | 0 |
| >140 to 160 | −210 | −43 | −14 | 0 |
| >160 to 180 | −230 | −43 | −14 | 0 |
| >180 to 200 | −240 | −50 | −15 | 0 |
| >200 to 225 | −260 | −50 | −15 | 0 |
| >225 to 250 | −280 | −50 | −15 | 0 |
| >250 to 280 | −300 | −56 | −17 | 0 |
| >280 to 315 | −330 | −56 | −17 | 0 |

In the source, f and g share one value across the split sub-rows (for example 30–40 and 40–50). Only c changes per sub-row.

**Hole H: lower deviation EI = 0.** This is ISO 286-1:2010 Table 2, H column (I checked the rows up to 18 mm). ISO 286-2 Table 6 also gives EI = 0 for H at 3–6 and 6–10.

Source (full-text copy of ISO 286-1:2010, IHS-licensed watermark, re-hosted): https://www.alekvs.com/wp-content/uploads/2025/06/ISO-286-1-2010en-Geometrical-product-specifications-GPS-%E2%80%94-ISO-code-system-for-tolerances-on-linear-sizes-%E2%80%94-Part-1-Basis-of-tolerances-deviations-and-fits.pdf

### 1.3 Limit deviations at ⌀6 and ⌀10 (ISO 286-2:2010 Tables 6, 17, 20, 21, 22)

⌀6 falls in the row ">3 to 6". ⌀10 falls in the row ">6 to 10".

| Class | ⌀6 upper / lower (µm) | ⌀10 upper / lower (µm) | ISO 286-2 table |
|---|---|---|---|
| H7 | +12 / 0 | +15 / 0 | Table 6 |
| H8 | +18 / 0 | +22 / 0 | Table 6 |
| H11 | +75 / 0 | +90 / 0 | Table 6 |
| g6 | −4 / −12 | −5 / −14 | Table 21 |
| h6 | 0 / −8 | 0 / −9 | Table 22 |
| f7 | −10 / −22 | −13 / −28 | Table 20 |
| c11 | −70 / −145 | −80 / −170 | Table 17 |

Source (full-text copy of ISO 286-2:2010, re-hosted): https://www.alekvs.com/wp-content/uploads/2025/06/ISO-286-2-2010en-Geometrical-product-specifications-GPS-%E2%80%94-ISO-code-system-for-tolerances-on-linear-sizes.pdf
Internal consistency: every row equals the 286-1 fundamental deviation plus or minus the IT value in 1.1 and 1.2. For example, c11 at ⌀6 is −70 − 75 = −145.

### 1.4 Fit clearances (COMPUTED from 1.3, not read from a source)

Clearance = hole limit − shaft limit, in µm.

| Fit | ⌀6 min / max clearance | ⌀10 min / max clearance |
|---|---|---|
| H7/g6 | 4 / 24 | 5 / 29 |
| H7/h6 | 0 / 20 | 0 / 24 |
| H8/f7 | 10 / 40 | 13 / 50 |
| H11/c11 | 70 / 220 | 80 / 260 |

The fit descriptions ("H11/c11 loose running", "H8/f7 close running", "H7/g6 sliding", "H7/h6 locational clearance") are listed in the preferred-fits table of the Benha University course notes: https://beng.stafpu.bu.edu.eg/Mechanical%20Engineering/4120/reports/Fit%20and%20Tolerance.pdf

---

## 2. Machining-process capability

### 2.1 IT grade by process: ANSI B4.2-1978 chart (read from figure)

These bars were read from the figure "International tolerance grades related to machining processes [ANSI B4.2-1978]", as reproduced in Benha University course notes (Fig-9). The chart's columns run from IT4 to IT11. Bars that reach the right edge are marked "11+" because the reproduction is cropped at IT11.

| Process | IT range |
|---|---|
| Lapping & honing | 4–5 |
| Cylindrical grinding | 5–7 |
| Surface grinding | 5–8 |
| Diamond turning | 5–7 |
| Diamond boring | 5–7 |
| Broaching | 5–8 |
| Powder metal, sized | 6–8 |
| Reaming | 6–10 |
| Turning | 7–11+ |
| Powder metal, sintered | 7–10 |
| Boring | 7–11+ |
| Milling | 10–11+ |
| Planing & shaping | 10–11+ |
| Drilling | 10–11+ |
| Punching | 10–11+ |

Source (university reproduction of an ANSI chart; values read visually from the image): https://beng.stafpu.bu.edu.eg/Mechanical%20Engineering/4120/reports/Fit%20and%20Tolerance.pdf (p. 7)

### 2.2 IT grade by process: RoyMech chart (secondary; disagrees with 2.1)

| Process | IT range | | Process | IT range |
|---|---|---|---|---|
| Lapping | 2–5 | | Sawing | 9–11 |
| Honing | 3–5 | | Milling | 9–13 |
| Superfinishing | 4–6 | | Planing, shaping | 10–14 |
| Cylindrical grinding | 4–7 | | Extruding | 10–13 |
| Diamond turning | 4–7 | | Cold rolling, drawing | 10–14 |
| Plan grinding | 5–9 | | Drilling | 11–14 |
| Broaching | 5–9 | | | |
| Reaming | 5–9 | | Forging | 13–16 |
| Boring, turning | 6–12 | | | |
| | | | Hot rolling, flame cutting | 14–16 |

Source (secondary UK engineering reference; I read the raw HTML table cell by cell): https://www.roymech.co.uk/Useful_Tables/ISO_Tolerances/
**Disagreements with 2.1:** RoyMech is one to two grades finer for lapping and honing (2–5 / 3–5 vs 4–5), and wider for reaming (5–9 vs 6–10). For drilling, RoyMech starts at IT11 and ANSI at IT10. RoyMech gives no source for its chart.

### 2.3 Absolute tolerances by process: MIL-HDBK-727 (Design Guidance for Producibility, 1984)

| Process | Tolerance as stated | Section |
|---|---|---|
| Reaming | holes <12.7 mm: ±0.025 mm. 12.7–25.4 mm: ±0.038 mm. >25.4 mm: ±0.051 mm. "Holes may be out of round by as much as size tolerance." | 4-3.2.1.5.4 |
| Broaching | round and square holes ±0.013 to ±0.025 mm | 4-3.2.1.5.5 |
| Boring, large machines | ±0.025 mm on holes up to 152.4 mm. **Hole location to ±0.013 mm** | 4-3.2.1.5.6 |
| Boring, special production machines | small holes ±0.003 to ±0.005 mm | 4-3.2.1.5.6 |
| Jig bores | **hole location to 0.003 mm** | 4-3.2.1.5.6 |
| Milling | "Tolerances of ±0.05 mm (0.002 in.) are not unusual" | 4-3.2.1.5.2 |
| Cylindrical grinding | ±0.003 to 0.013 mm on diameters | 4-3.2.1.5.10 |
| Centerless grinding | 0.0010 to 0.13 mm | 4-3.2.1.5.10 |
| Turning (Table 4-46) | engine lathe ±0.005 mm. Turret lathe ±0.127. Single-spindle auto ±0.076. Six-spindle auto ±0.127 | Table 4-46 |
| NC drilling (Table 4-50) | standard ±0.05 mm. Special, at added cost, ±0.025 mm | Table 4-50 |
| NC milling (Table 4-50) | standard ±0.05. Special ±0.013 | Table 4-50 |
| NC turning (Table 4-50) | standard ±0.08. Special ±0.025 | Table 4-50 |
| NC boring (Table 4-50) | standard ±0.03. Special ±0.008 | Table 4-50 |

Source (gov handbook, primary): https://everyspec.com/MIL-HDBK/MIL-HDBK-0700-0799/MIL_HDBK_727_1853/ (PDF: https://everyspec.com/MIL-HDBK/MIL-HDBK-0700-0799/download.php?spec=MIL_HDBK_727.1853.pdf)
Table 4-50 does not say whether its NC drilling tolerance is size or location.

### 2.4 Hole positional accuracy: no jig vs jig vs CNC

- **Primary figure:** MIL-HDBK-727 Figure 4-38 "Cost vs True Position Tolerance" (Ref. 5, SME). The x-axis is true-position tolerance ±mm (in.), with ticks at 0.013 (±0.0005), 0.13 (0.005), 0.25 (0.010), 0.38 (0.015) and 0.51 (0.020). The figure shows three regions:
  - **"Drill Jig"** runs from **±0.25 mm (0.010 in)** out to the end of the axis at **±0.51 mm (0.020 in)**.
  - **"Jig Bore"** runs from about ±0.07 mm (read visually; the boundary sits between the 0.013 and 0.13 ticks) up to **±0.25 mm**.
  - **"Special Equipment"** covers anything finer than the jig-bore range.
  - Relative cost rises steeply below ±0.13 mm.

  Source: same MIL-HDBK-727 PDF as 2.3, p. 4-97 (read from the scanned image, so boundaries are approximate except the 0.25 mm Jig Bore/Drill Jig boundary, which falls on a tick).
- **CNC:** the only primary figure is MIL-HDBK-727 Table 4-50 in 2.3: NC drilling ±0.05 mm standard, ±0.025 mm special. It is not stated whether this is positional.
- **Drilled without a jig (layout / hand-positioned):** no citable primary figure found. Vendor pages give CNC drilling true position ±0.02–0.05 mm (https://daheprecision.com/cnc-drilling-tolerances/), but I did not verify that page's text directly, so treat it as uncited.

---

## 3. ISO 965-1 metric thread tolerances

All values are in µm. Grades that the standard leaves blank are shown as "—".

### 3.1 Fundamental deviations (ISO 965-1:2013 Table 1)

| Pitch P (mm) | G EI | H EI | e es | f es | g es | h es |
|---|---|---|---|---|---|---|
| 0.5 | +20 | 0 | −50 | −36 | −20 | 0 |
| 0.6 | +21 | 0 | −53 | −36 | −21 | 0 |
| 0.7 | +22 | 0 | −56 | −38 | −22 | 0 |
| 0.75 | +22 | 0 | −56 | −38 | −22 | 0 |
| 0.8 | +24 | 0 | −60 | −38 | −24 | 0 |
| 1 | +26 | 0 | −60 | −40 | −26 | 0 |
| 1.25 | +28 | 0 | −63 | −42 | −28 | 0 |
| 1.5 | +32 | 0 | −67 | −45 | −32 | 0 |
| 1.75 | +34 | 0 | −71 | −48 | −34 | 0 |
| 2 | +38 | 0 | −71 | −52 | −38 | 0 |
| 2.5 | +42 | 0 | −80 | −58 | −42 | 0 |

Source (ISO preview, primary): https://cdn.standards.iteh.ai/samples/57778/765c17fc0f7946ff850db22b43920f09/ISO-965-1-2013.pdf. This matches the 1998 edition Table 1 in the copy cited in 3.3.

### 3.2 Crest-diameter tolerances (ISO 965-1:2013 Tables 2 and 3; same values as 1998 Tables 3 and 4)

| P (mm) | Td g4 | Td g6 | Td g8 | TD1 g4 | TD1 g5 | TD1 g6 | TD1 g7 | TD1 g8 |
|---|---|---|---|---|---|---|---|---|
| 0.5 | 67 | 106 | — | 90 | 112 | 140 | 180 | — |
| 0.6 | 80 | 125 | — | 100 | 125 | 160 | 200 | — |
| 0.7 | 90 | 140 | — | 112 | 140 | 180 | 224 | — |
| 0.75 | 90 | 140 | — | 118 | 150 | 190 | 236 | — |
| 0.8 | 95 | 150 | 236 | 125 | 160 | 200 | 250 | 315 |
| 1 | 112 | 180 | 280 | 150 | 190 | 236 | 300 | 375 |
| 1.25 | 132 | 212 | 335 | 170 | 212 | 265 | 335 | 425 |
| 1.5 | 150 | 236 | 375 | 190 | 236 | 300 | 375 | 475 |
| 1.75 | 170 | 265 | 425 | 212 | 265 | 335 | 425 | 530 |
| 2 | 180 | 280 | 450 | 236 | 300 | 375 | 475 | 600 |
| 2.5 | 212 | 335 | 530 | 280 | 355 | 450 | 560 | 710 |

Td is the major diameter of the external thread; grades 5 and 7 do not exist for it. TD1 is the minor diameter of the internal thread.
Source (ISO preview, primary): https://cdn.standards.iteh.ai/samples/57778/765c17fc0f7946ff850db22b43920f09/ISO-965-1-2013.pdf

### 3.3 Pitch-diameter tolerances (ISO 965-1:1998 Tables 5 and 6), for basic major diameters covering M3–M20

**TD2 (internal), grades 4–8**

| d range (mm) | P | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|
| >2.8–5.6 | 0.35 | 56 | 71 | 90 | — | — |
| | 0.5 | 63 | 80 | 100 | 125 | — |
| | 0.6 | 71 | 90 | 112 | 140 | — |
| | 0.7 | 75 | 95 | 118 | 150 | — |
| | 0.75 | 75 | 95 | 118 | 150 | — |
| | 0.8 | 80 | 100 | 125 | 160 | 200 |
| >5.6–11.2 | 0.75 | 85 | 106 | 132 | 170 | — |
| | 1 | 95 | 118 | 150 | 190 | 236 |
| | 1.25 | 100 | 125 | 160 | 200 | 250 |
| | 1.5 | 112 | 140 | 180 | 224 | 280 |
| >11.2–22.4 | 1 | 100 | 125 | 160 | 200 | 250 |
| | 1.25 | 112 | 140 | 180 | 224 | 280 |
| | 1.5 | 118 | 150 | 190 | 236 | 300 |
| | 1.75 | 125 | 160 | 200 | 250 | 315 |
| | 2 | 132 | 170 | 212 | 265 | 335 |
| | 2.5 | 140 | 180 | 224 | 280 | 355 |

**Td2 (external), grades 3–9**

| d range (mm) | P | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
|---|---|---|---|---|---|---|---|---|
| >2.8–5.6 | 0.35 | 34 | 42 | 53 | 67 | 85 | — | — |
| | 0.5 | 38 | 48 | 60 | 75 | 95 | — | — |
| | 0.6 | 42 | 53 | 67 | 85 | 106 | — | — |
| | 0.7 | 45 | 56 | 71 | 90 | 112 | — | — |
| | 0.75 | 45 | 56 | 71 | 90 | 112 | — | — |
| | 0.8 | 48 | 60 | 75 | 95 | 118 | 150 | 190 |
| >5.6–11.2 | 0.75 | 50 | 63 | 80 | 100 | 125 | — | — |
| | 1 | 56 | 71 | 90 | 112 | 140 | 180 | 224 |
| | 1.25 | 60 | 75 | 95 | 118 | 150 | 190 | 236 |
| | 1.5 | 67 | 85 | 106 | 132 | 170 | 212 | 265 |
| >11.2–22.4 | 1 | 60 | 75 | 95 | 118 | 150 | 190 | 236 |
| | 1.25 | 67 | 85 | 106 | 132 | 170 | 212 | 265 |
| | 1.5 | 71 | 90 | 112 | 140 | 180 | 224 | 280 |
| | 1.75 | 75 | 95 | 118 | 150 | 190 | 236 | 300 |
| | 2 | 80 | 100 | 125 | 160 | 200 | 250 | 315 |
| | 2.5 | 85 | 106 | 132 | 170 | 212 | 265 | 335 |

Source (full-text copy of BS ISO 965-1:1998, a BSI-licensed copy watermarked "Istanbul Teknik Üniversitesi", re-hosted): https://regbar.com/wp-content/uploads/2019/09/ISO-965-1-1998-%C3%96NEML%C4%B0.pdf
These tables come from the **1998** edition. In the 2013 ISO preview (3.1), Table 4 (TD2) is cut off after its first row, and that row matches 1998. I could not check the rest of the 2013 edition's TD2/Td2 values.
Grade ratios given by the standard (clause 13): Td2 grades 3/4/5/6/7/8/9 = 0.5/0.63/0.8/1/1.25/1.6/2 × Td2(6). TD2 grades 4/5/6/7/8 = 0.85/1.06/1.32/1.7/2.12 × Td2(6). Td2(6) = 90·P^0.4·d^0.1.

### 3.4 Default and recommended classes (ISO 965-1:1998 clauses 5.2, 6 and 12, Tables 8 and 9)

- If no tolerance class is given, the default is quality **medium**. Internal threads: **5H up to and including M1.4, 6H for M1.6 and larger**. External threads: **6h up to and including M1.4, 6g for M1.6 and larger**. If no engagement group is given, the default is **N** (normal).
- "Grade 6 shall be used for tolerance quality medium and normal length of thread engagement." Fine = precision threads. **Medium = "for general use."** Coarse = where manufacturing is difficult (for example hot-rolled bar or long blind holes).
- Preferred fits are H/g, H/h or G/h.
- Table 8 (internal), by engagement group S / N / L. Fine: H 4H 5H 6H. Medium: G (5G) 6G (7G); H 5H 6H 7H. Coarse: G – (7G) (8G); H – 7H 8H.
- Table 9 (external), by engagement group S / N / L. Fine: g – (4g) (5g4g); h (3h4h) 4h (5h4h). Medium: e – 6e (7e6e); f – 6f –; g (5g6g) 6g (7g6g); h (5h6h) 6h (7h6h). Coarse: e – (8e) (9e8e); g – 8g (9g8g).
- In the source, bold marks first choice and parentheses mark third choice. The bold formatting and the "commercial thread" frames were lost in text extraction.

Source: same BS ISO 965-1:1998 copy, https://regbar.com/wp-content/uploads/2019/09/ISO-965-1-1998-%C3%96NEML%C4%B0.pdf

---

## Could not source (or sourced only weakly)

1. **Positional accuracy of drilled holes without a jig** (layout / hand-positioned drilling): no citable figure. MIL-HDBK-727 Fig. 4-38 gives only drill jig (±0.25–0.51 mm) and jig bore (≈±0.07–0.25 mm) regions. Its Table 4-50 gives NC drilling ±0.05 / ±0.025 mm without saying whether that is size or location.
2. **A primary ISO source for process → IT grade.** ISO 286 does not contain one. The two charts found (ANSI B4.2-1978 via a university reproduction, and RoyMech) disagree by 1–2 grades at the fine end and for reaming and drilling. The ANSI reproduction is also cropped at IT11, so the coarse end of turning, boring, milling and drilling is cut off.
3. **ISO 965-1:2013 TD2/Td2 tables:** only the 1998-edition values could be read (2013 preview truncated). I did not verify whether 2013 changed any pitch-diameter values.
4. **MIL-HDBK-727 Fig. 4-16** (tolerance range vs size for general machining processes) exists in the source but is a badly interlaced scan. I did not transcribe it.
