{
  "inputs": {
    "iP.y": 0.0,
    "iP.x": 0.0,
    "iP.z": 0.0
  },
  "helpers": [],
  "glsl": [
    "float r9 = iP.x;",
    "float r24 = r9 * 2.0;",
    "float r13 = iP.y;",
    "float r23 = r9 + r13;",
    "float r22 = r23 * 2.0;",
    "float r21 = r24 + r22;",
    "float r20 = r21 * 0.05;",
    "float r19 = iP.z;",
    "float r18 = r9 * 0.1;",
    "float r17 = r19 + r18;",
    "float r16 = r20 + r17;",
    "float r15 = r16 + 0.06;",
    "float r14 = r15 * r15;",
    "float r12 = r13 + 0.09459442;",
    "float r11 = r12 + 0.05;",
    "float r10 = r11 * r11;",
    "float r8 = r9 + 0.1;",
    "float r7 = r8 + 0.006708149;",
    "float r6 = r7 + 0.04;",
    "float r5 = r6 * r6;",
    "float r4 = r10 + r5;",
    "float r3 = r14 + r4;",
    "float r2 = max(r3, 0.0);",
    "float r1 = sqrt(r2);",
    "float r0 = r1 - 2.0;",
    "return ColorImplicit(r0, iP / length(iP), vec4(color_primary, 1.0));"
  ]
}