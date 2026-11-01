{
  "inputs": {
    "iP.x": 0.0,
    "iP.y": 0.0,
    "iP.z": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r9 = iP.z;",
    "float r8 = iP.x;",
    "float r7 = iP.y;",
    "float r6 = max(r8, r7);",
    "float r5 = min(r9, r6);",
    "float r10 = r9 * r5;",
    "float r4 = r7 * r5;",
    "float r3 = r10 - r4;",
    "float r2 = exp(r3);",
    "float r12 = -r3;",
    "float r11 = exp(r12);",
    "float r13 = r2 - r11;",
    "float r1 = r11 + r2;",
    "float r0 = r13 / r1;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}