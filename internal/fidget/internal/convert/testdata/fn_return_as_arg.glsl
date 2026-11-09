{
  "inputs": {
    "iP.y": 0.0,
    "iP.z": 0.0,
    "iP.x": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r10 = iP.z;",
    "float r9 = abs(r10);",
    "float r8 = r9 - 1.0;",
    "float r21 = iP.y;",
    "float r20 = abs(r21);",
    "float r19 = r20 - 1.0;",
    "float r24 = iP.x;",
    "float r23 = abs(r24);",
    "float r22 = r23 - 1.0;",
    "float r25 = min(r19, r22);",
    "float r18 = r22 - r19;",
    "float r17 = abs(r18);",
    "float r16 = 0.25 - r17;",
    "float r15 = max(r16, 0.0);",
    "float r14 = r15 / 0.25;",
    "float r13 = r14 * r14;",
    "float r12 = r13 * 0.0625;",
    "float r11 = r25 - r12;",
    "float r26 = min(r8, r11);",
    "float r7 = r11 - r8;",
    "float r6 = abs(r7);",
    "float r5 = 0.25 - r6;",
    "float r4 = max(r5, 0.0);",
    "float r3 = r4 / 0.25;",
    "float r2 = r3 * r3;",
    "float r1 = r2 * 0.0625;",
    "float r0 = r26 - r1;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}