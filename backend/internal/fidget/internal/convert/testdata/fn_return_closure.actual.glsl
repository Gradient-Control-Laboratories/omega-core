{
  "inputs": {
    "iP.y": 0.0,
    "iP.x": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r9 = iP.x;",
    "float r8 = iP.y;",
    "float r10 = min(r9, r8);",
    "float r7 = r9 - r8;",
    "float r6 = abs(r7);",
    "float r5 = 0.1 - r6;",
    "float r4 = max(r5, 0.0);",
    "float r3 = r4 / 0.1;",
    "float r2 = r3 * r3;",
    "float r1 = r2 * 0.025;",
    "float r0 = r10 - r1;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}