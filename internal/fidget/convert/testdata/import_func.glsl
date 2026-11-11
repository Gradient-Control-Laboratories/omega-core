{
  "inputs": {
    "iP.z": 0.0,
    "iP.x": 0.0,
    "iP.y": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r10 = iP.z;",
    "float r9 = r10 * r10;",
    "float r8 = iP.y;",
    "float r7 = r8 * r8;",
    "float r6 = iP.x;",
    "float r5 = r6 * r6;",
    "float r4 = r7 + r5;",
    "float r3 = r9 + r4;",
    "float r2 = max(r3, 0.0);",
    "float r1 = sqrt(r2);",
    "float r0 = r1 - 2.5;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}