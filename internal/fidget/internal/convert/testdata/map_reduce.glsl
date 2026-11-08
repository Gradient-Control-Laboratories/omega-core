{
  "inputs": {
    "iP.y": 0.0,
    "iP.x": 0.0,
    "iP.z": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r11 = iP.z;",
    "float r10 = r11 * r11;",
    "float r9 = iP.y;",
    "float r8 = r9 * r9;",
    "float r7 = iP.x;",
    "float r6 = r7 - 0.5;",
    "float r5 = r6 * r6;",
    "float r4 = r8 + r5;",
    "float r3 = r10 + r4;",
    "float r2 = max(r3, 0.0);",
    "float r1 = sqrt(r2);",
    "float r0 = r1 - 1.0;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}