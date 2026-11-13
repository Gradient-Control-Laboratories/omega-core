{
  "inputs": {
    "iP.x": 0.0
  },
  "helpers": [
    "float fidget_not(float x) { return (x == 0.0) ? 1.0 : 0.0; }",
    "float fidget_and(float x, float y) { return (x != 0.0) && (y != 0.0) ? 1.0 : 0.0; }"
  ],
  "glsl": [
    "float r6 = iP.x;",
    "float r5 = 6.0 - r6;",
    "float r26 = sign(7.0 - r5);",
    "float r25 = max(r26, 0.0);",
    "float r24 = -r6;",
    "float r23 = min(r6, 3.0);",
    "float r22 = min(r5, r23);",
    "float r21 = fidget_not(r22);",
    "float r20 = r24 + r21;",
    "float r19 = r25 + r20;",
    "float r18 = r19 * r19;",
    "float r17 = r18 - 7.0;",
    "float r16 = r17 * r17;",
    "float r15 = abs(r16);",
    "float r27 = sign(r15 - r22);",
    "float r14 = sign(r27 - r15);",
    "float r13 = max(r14, 0.0);",
    "float r29 = fidget_not(r13);",
    "float r28 = fidget_and(r29, r5);",
    "float r12 = fidget_and(r13, r6);",
    "float r11 = r28 + r12;",
    "float r31 = sign(0.0 - r11);",
    "float r30 = max(r31, 0.0);",
    "float r34 = fidget_and(r30, r22);",
    "float r10 = sign(r11 - 0.0);",
    "float r9 = max(r10, 0.0);",
    "float r33 = fidget_and(r9, r6);",
    "float r32 = r34 + r33;",
    "float r8 = r30 + r9;",
    "float r7 = fidget_not(r8);",
    "float r4 = fidget_and(r7, r5);",
    "float r3 = r32 + r4;",
    "float r2 = min(r11, r3);",
    "float r1 = min(r22, r2);",
    "float r0 = min(r22, r1);",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}