{
  "inputs": {
    "iP.x": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r6 = iP.x;",
    "float r5 = abs(r6);",
    "float r4 = max(r5, 1e-12);",
    "float r3 = mod(3.0, r4);",
    "float r2 = min(r3, 2.0);",
    "float r1 = min(r2, 3.0);",
    "float r0 = min(r1, 4.0);",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}