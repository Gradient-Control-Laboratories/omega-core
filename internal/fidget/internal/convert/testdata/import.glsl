{
  "inputs": {
    "iP.x": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r1 = iP.x;",
    "float r0 = r1 + 3.1415927;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}