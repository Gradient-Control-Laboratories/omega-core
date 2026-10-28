{
  "inputs": {
    "iP.z": 0.0,
    "iP.x": 0.0,
    "iP.y": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r13 = iP.z;",
    "float r12 = r13 * r13;",
    "float r11 = iP.y;",
    "float r10 = r11 * r11;",
    "float r9 = iP.x;",
    "float r8 = r9 * r9;",
    "float r7 = r10 + r8;",
    "float r6 = r12 + r7;",
    "float r5 = max(r6, 0.0);",
    "float r4 = sqrt(r5);",
    "float r14 = r4 - 9.5;",
    "float r3 = r4 + 8.0;",
    "float r2 = r14 + r3;",
    "float r1 = r2 * 2.0;",
    "float r0 = r2 + r1;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}