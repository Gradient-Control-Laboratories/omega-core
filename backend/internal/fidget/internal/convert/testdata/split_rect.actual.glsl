{
  "inputs": {
    "iP.z": 0.0,
    "iP.x": 0.0,
    "iP.y": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r15 = iP.z;",
    "float r6 = iP.x;",
    "float r14 = sign(r6 - 0.0);",
    "float r13 = r14 * 1.5;",
    "float r12 = r15 - r13;",
    "float r16 = r12 - 1.0;",
    "float r11 = -1.0 - r12;",
    "float r9 = iP.y;",
    "float r10 = r9 - 1.0;",
    "float r8 = -1.0 - r9;",
    "float r7 = r6 - 2.0;",
    "float r5 = -2.0 - r6;",
    "float r4 = max(r7, r5);",
    "float r3 = max(r8, r4);",
    "float r2 = max(r10, r3);",
    "float r1 = max(r11, r2);",
    "float r0 = max(r16, r1);",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}