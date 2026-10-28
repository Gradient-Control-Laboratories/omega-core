{
  "inputs": {
    "iP.z": 0.0,
    "iP.y": 0.0,
    "iP.x": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r9 = iP.z;",
    "float r8 = r9 * r9;",
    "float r7 = iP.y;",
    "float r6 = r7 * r7;",
    "float r5 = iP.x;",
    "float r4 = r5 * r5;",
    "float r3 = r6 + r4;",
    "float r2 = r8 + r3;",
    "float r1 = sqrt(r2);",
    "float r0 = r1 - 1.0;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}