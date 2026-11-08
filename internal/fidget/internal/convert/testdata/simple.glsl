{
  "inputs": {
    "iP.x": 0.0
  },
  "helpers": [
    "float fidget_not(float x) { return (x == 0.0) ? 1.0 : 0.0; }"
  ],
  "glsl": [
    "float r21 = iP.x;",
    "float r22 = 6.0 - r21;",
    "float r20 = min(r21, 3.0);",
    "float r19 = min(r22, r20);",
    "float r33 = r19 - r22;",
    "float r25 = sign(7.0 - r22);",
    "float r24 = max(r25, 0.0);",
    "float r23 = -r21;",
    "float r18 = fidget_not(r19);",
    "float r17 = r23 + r18;",
    "float r16 = r24 + r17;",
    "float r15 = r16 * r16;",
    "float r14 = r15 - 7.0;",
    "float r13 = r14 * r14;",
    "float r12 = abs(r13);",
    "float r26 = sign(r12 - r19);",
    "float r11 = sign(r26 - r12);",
    "float r10 = max(r11, 0.0);",
    "float r28 = 1.0 - r10;",
    "float r27 = r22 * r28;",
    "float r9 = r21 * r10;",
    "float r8 = r27 + r9;",
    "float r32 = sign(0.0 - r8);",
    "float r31 = max(r32, 0.0);",
    "float r30 = r33 * r31;",
    "float r29 = r21 - r22;",
    "float r7 = sign(r8 - 0.0);",
    "float r6 = max(r7, 0.0);",
    "float r5 = r29 * r6;",
    "float r4 = r30 + r5;",
    "float r3 = r22 + r4;",
    "float r2 = min(r8, r3);",
    "float r1 = min(r19, r2);",
    "float r0 = min(r19, r1);",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}