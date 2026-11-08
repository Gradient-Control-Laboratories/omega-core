{
  "inputs": {
    "iP.y": 0.0,
    "iP.x": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r5 = iP.y;",
    "float r4 = r5 + 7.0;",
    "float r3 = r4 + 1.0;",
    "float r2 = iP.x;",
    "float r1 = r2 + 1.0;",
    "float r0 = min(r3, r1);",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}