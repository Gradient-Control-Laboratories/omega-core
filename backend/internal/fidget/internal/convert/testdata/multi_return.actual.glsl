{
  "inputs": {
    "iP.x": 0.0,
    "iP.y": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r3 = iP.x;",
    "float r2 = iP.y;",
    "float r1 = r2 + 7.0;",
    "float r0 = min(r3, r1);",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}