fn gsl_remap(p: vec3<f32>) -> f32 {
  return length(p.xy - vec2<f32>(0.0, 0.0)) - 0.5;
}

fn builtin_mix(a: f32, b: f32, t: f32) -> f32 {
  return a + (b - a) * t;
}

fn builtin_intersectSharp_2(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound_2(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp_2(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound_2(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.1);
}

fn builtin_gcl_closure_intersectRound_3(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.1);
}

fn builtin_firstOf(list: f32) -> f32 {
  return list;
}

fn builtin_intersectChamferEdges(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp_2(fields);
  return builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.x + fields.y + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields.x + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.z + setback) / 1.4142135, 0.0 > 0.0)))));
}

fn builtin_lastOf(list: f32) -> f32 {
  return list;
}

fn builtin_intersectChamferVertices(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp_2(fields);
  return builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.z + setback) / 1.7320508, 1.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0)))))));
}

fn builtin_intersectChamfer(fields: vec3<f32>, setback: f32) -> f32 {
  return max(max(builtin_intersectSharp_2(fields), builtin_intersectChamferEdges(fields, builtin_firstOf(setback))), builtin_intersectChamferVertices(fields, builtin_lastOf(setback)));
}

fn builtin_gcl_closure_intersectChamfer(fields: vec3<f32>) -> f32 {
  return builtin_intersectChamfer(fields, 0.1);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let gsl_tmp_34 = vec3<f32>(0.0, 0.0, 0.0) - p;
  let gsl_tmp_35 = 0.25 * (6.5 + dot(gsl_tmp_34, vec3<f32>(0.0, -2.0, 0.0))) / 3.0;
  let gsl_tmp_37 = -0.5 * (0.5 - 3.0 * gsl_tmp_35) + 0.25 * dot(gsl_tmp_34, vec3<f32>(1.5, 1.0, 0.0));
  let gsl_tmp_36 = gsl_tmp_35 - 0.25;
  let gsl_tmp_44 = sqrt(max(-gsl_tmp_36, 0.0));
  let gsl_tmp_45 = acos(clamp(gsl_tmp_37 / (2.0 * gsl_tmp_36 * gsl_tmp_44 - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let gsl_tmp_47 = sin(gsl_tmp_45) * 1.7320508;
  let gsl_tmp_46 = cos(gsl_tmp_45);
  let gsl_tmp_50 = clamp((gsl_tmp_47 - gsl_tmp_46) * gsl_tmp_44 - -0.5, 0.0, 1.0);
  let gsl_tmp_49 = clamp((-gsl_tmp_47 - gsl_tmp_46) * gsl_tmp_44 - -0.5, 0.0, 1.0);
  let gsl_tmp_53 = gsl_tmp_34 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_49) * gsl_tmp_49;
  let gsl_tmp_54 = dot(gsl_tmp_53, gsl_tmp_53);
  let gsl_tmp_55 = gsl_tmp_34 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_50) * gsl_tmp_50;
  let gsl_tmp_56 = dot(gsl_tmp_55, gsl_tmp_55);
  let gsl_tmp_48 = clamp((gsl_tmp_46 + gsl_tmp_46) * gsl_tmp_44 - -0.5, 0.0, 1.0);
  let gsl_tmp_51 = gsl_tmp_34 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_48) * gsl_tmp_48;
  let gsl_tmp_52 = dot(gsl_tmp_51, gsl_tmp_51);
  let gsl_tmp_38 = gsl_tmp_37 * gsl_tmp_37 + 4.0 * gsl_tmp_36 * gsl_tmp_36 * gsl_tmp_36;
  let gsl_tmp_39 = sqrt(max(gsl_tmp_38, 0.0));
  let gsl_tmp_40 = (gsl_tmp_39 - gsl_tmp_37) / 2.0;
  let gsl_tmp_41 = (-gsl_tmp_39 - gsl_tmp_37) / 2.0;
  let gsl_tmp_42 = clamp(sign(gsl_tmp_40) * exp(log(max(abs(gsl_tmp_40), 0.000000000001)) / 3.0) + sign(gsl_tmp_41) * exp(log(max(abs(gsl_tmp_41), 0.000000000001)) / 3.0) - -0.5, 0.0, 1.0);
  let gsl_tmp_57 = select(select(select(gsl_tmp_50, gsl_tmp_49, gsl_tmp_54 <= gsl_tmp_56), select(gsl_tmp_50, gsl_tmp_48, gsl_tmp_52 <= gsl_tmp_56), gsl_tmp_52 <= gsl_tmp_54), gsl_tmp_42, gsl_tmp_38 >= 0.0);
  let gsl_tmp_59 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(3.0, 2.0, 0.0) * gsl_tmp_57 + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_57 * gsl_tmp_57;
  let gsl_tmp_65 = p - gsl_tmp_59;
  let gsl_tmp_60 = vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -4.0, 0.0) * gsl_tmp_57;
  let gsl_tmp_62 = gsl_tmp_60 / (length(gsl_tmp_60) + 0.000000000001);
  let gsl_tmp_63 = vec3<f32>(0.0, 0.0, 1.0) - gsl_tmp_62 * dot(vec3<f32>(0.0, 0.0, 1.0), gsl_tmp_62);
  let gsl_tmp_64 = gsl_tmp_63 / (length(gsl_tmp_63) + 0.000000000001);
  let gsl_tmp_66 = vec3<f32>(dot(gsl_tmp_65, cross(gsl_tmp_62, gsl_tmp_64)), dot(gsl_tmp_65, gsl_tmp_64), 0.0);
  let filleted = builtin_gcl_closure_intersectRound_2(vec3<f32>(builtin_mix(gsl_remap(gsl_tmp_66), builtin_gcl_closure_rectangleXYCentered(gsl_tmp_66, vec2<f32>(0.0, 0.0), vec2<f32>(0.8, 0.8)), gsl_tmp_57), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.8320503, 0.5547002, 0.0)), dot(p - vec3<f32>(3.0, 0.0, 0.0), vec3<f32>(0.8320503, -0.5547002, 0.0))));
  return filleted;
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let gsl_tmp_67 = vec3<f32>(0.0, 0.0, 0.0) - p;
  let gsl_tmp_68 = 0.25 * (6.5 + dot(gsl_tmp_67, vec3<f32>(0.0, -2.0, 0.0))) / 3.0;
  let gsl_tmp_70 = -0.5 * (0.5 - 3.0 * gsl_tmp_68) + 0.25 * dot(gsl_tmp_67, vec3<f32>(1.5, 1.0, 0.0));
  let gsl_tmp_69 = gsl_tmp_68 - 0.25;
  let gsl_tmp_77 = sqrt(max(-gsl_tmp_69, 0.0));
  let gsl_tmp_78 = acos(clamp(gsl_tmp_70 / (2.0 * gsl_tmp_69 * gsl_tmp_77 - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let gsl_tmp_80 = sin(gsl_tmp_78) * 1.7320508;
  let gsl_tmp_79 = cos(gsl_tmp_78);
  let gsl_tmp_83 = clamp((gsl_tmp_80 - gsl_tmp_79) * gsl_tmp_77 - -0.5, 0.0, 1.0);
  let gsl_tmp_82 = clamp((-gsl_tmp_80 - gsl_tmp_79) * gsl_tmp_77 - -0.5, 0.0, 1.0);
  let gsl_tmp_86 = gsl_tmp_67 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_82) * gsl_tmp_82;
  let gsl_tmp_87 = dot(gsl_tmp_86, gsl_tmp_86);
  let gsl_tmp_88 = gsl_tmp_67 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_83) * gsl_tmp_83;
  let gsl_tmp_89 = dot(gsl_tmp_88, gsl_tmp_88);
  let gsl_tmp_81 = clamp((gsl_tmp_79 + gsl_tmp_79) * gsl_tmp_77 - -0.5, 0.0, 1.0);
  let gsl_tmp_84 = gsl_tmp_67 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_81) * gsl_tmp_81;
  let gsl_tmp_85 = dot(gsl_tmp_84, gsl_tmp_84);
  let gsl_tmp_71 = gsl_tmp_70 * gsl_tmp_70 + 4.0 * gsl_tmp_69 * gsl_tmp_69 * gsl_tmp_69;
  let gsl_tmp_72 = sqrt(max(gsl_tmp_71, 0.0));
  let gsl_tmp_73 = (gsl_tmp_72 - gsl_tmp_70) / 2.0;
  let gsl_tmp_74 = (-gsl_tmp_72 - gsl_tmp_70) / 2.0;
  let gsl_tmp_75 = clamp(sign(gsl_tmp_73) * exp(log(max(abs(gsl_tmp_73), 0.000000000001)) / 3.0) + sign(gsl_tmp_74) * exp(log(max(abs(gsl_tmp_74), 0.000000000001)) / 3.0) - -0.5, 0.0, 1.0);
  let gsl_tmp_90 = select(select(select(gsl_tmp_83, gsl_tmp_82, gsl_tmp_87 <= gsl_tmp_89), select(gsl_tmp_83, gsl_tmp_81, gsl_tmp_85 <= gsl_tmp_89), gsl_tmp_85 <= gsl_tmp_87), gsl_tmp_75, gsl_tmp_71 >= 0.0);
  let gsl_tmp_92 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(3.0, 2.0, 0.0) * gsl_tmp_90 + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_90 * gsl_tmp_90;
  let gsl_tmp_98 = p - gsl_tmp_92;
  let gsl_tmp_93 = vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -4.0, 0.0) * gsl_tmp_90;
  let gsl_tmp_95 = gsl_tmp_93 / (length(gsl_tmp_93) + 0.000000000001);
  let gsl_tmp_96 = vec3<f32>(0.0, 0.0, 1.0) - gsl_tmp_95 * dot(vec3<f32>(0.0, 0.0, 1.0), gsl_tmp_95);
  let gsl_tmp_97 = gsl_tmp_96 / (length(gsl_tmp_96) + 0.000000000001);
  let gsl_tmp_99 = dot(gsl_tmp_98, cross(gsl_tmp_95, gsl_tmp_97));
  let gsl_tmp_101 = builtin_mix(1.0, 1.5, gsl_tmp_90);
  let gsl_tmp_100 = dot(gsl_tmp_98, gsl_tmp_97);
  let gsl_tmp_102 = builtin_mix(0.6666667, 1.0, gsl_tmp_90);
  let tapered = builtin_intersectSharp_2(vec3<f32>(builtin_mix(gsl_remap(vec3<f32>(gsl_tmp_99 / gsl_tmp_101, gsl_tmp_100 / gsl_tmp_101, 0.0)) * gsl_tmp_101, builtin_gcl_closure_rectangleXYCentered(vec3<f32>(gsl_tmp_99 / gsl_tmp_102, gsl_tmp_100 / gsl_tmp_102, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(0.8, 0.8)) * gsl_tmp_102, gsl_tmp_90), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.8320503, 0.5547002, 0.0)), dot(p - vec3<f32>(3.0, 0.0, 0.0), vec3<f32>(0.8320503, -0.5547002, 0.0))));
  return tapered;
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let gsl_tmp_103 = vec3<f32>(0.0, 0.0, 0.0) - p;
  let gsl_tmp_104 = 0.25 * (6.5 + dot(gsl_tmp_103, vec3<f32>(0.0, -2.0, 0.0))) / 3.0;
  let gsl_tmp_106 = -0.5 * (0.5 - 3.0 * gsl_tmp_104) + 0.25 * dot(gsl_tmp_103, vec3<f32>(1.5, 1.0, 0.0));
  let gsl_tmp_105 = gsl_tmp_104 - 0.25;
  let gsl_tmp_113 = sqrt(max(-gsl_tmp_105, 0.0));
  let gsl_tmp_114 = acos(clamp(gsl_tmp_106 / (2.0 * gsl_tmp_105 * gsl_tmp_113 - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let gsl_tmp_116 = sin(gsl_tmp_114) * 1.7320508;
  let gsl_tmp_115 = cos(gsl_tmp_114);
  let gsl_tmp_119 = clamp((gsl_tmp_116 - gsl_tmp_115) * gsl_tmp_113 - -0.5, 0.0, 1.0);
  let gsl_tmp_118 = clamp((-gsl_tmp_116 - gsl_tmp_115) * gsl_tmp_113 - -0.5, 0.0, 1.0);
  let gsl_tmp_122 = gsl_tmp_103 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_118) * gsl_tmp_118;
  let gsl_tmp_123 = dot(gsl_tmp_122, gsl_tmp_122);
  let gsl_tmp_124 = gsl_tmp_103 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_119) * gsl_tmp_119;
  let gsl_tmp_125 = dot(gsl_tmp_124, gsl_tmp_124);
  let gsl_tmp_117 = clamp((gsl_tmp_115 + gsl_tmp_115) * gsl_tmp_113 - -0.5, 0.0, 1.0);
  let gsl_tmp_120 = gsl_tmp_103 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_117) * gsl_tmp_117;
  let gsl_tmp_121 = dot(gsl_tmp_120, gsl_tmp_120);
  let gsl_tmp_107 = gsl_tmp_106 * gsl_tmp_106 + 4.0 * gsl_tmp_105 * gsl_tmp_105 * gsl_tmp_105;
  let gsl_tmp_108 = sqrt(max(gsl_tmp_107, 0.0));
  let gsl_tmp_109 = (gsl_tmp_108 - gsl_tmp_106) / 2.0;
  let gsl_tmp_110 = (-gsl_tmp_108 - gsl_tmp_106) / 2.0;
  let gsl_tmp_111 = clamp(sign(gsl_tmp_109) * exp(log(max(abs(gsl_tmp_109), 0.000000000001)) / 3.0) + sign(gsl_tmp_110) * exp(log(max(abs(gsl_tmp_110), 0.000000000001)) / 3.0) - -0.5, 0.0, 1.0);
  let gsl_tmp_126 = select(select(select(gsl_tmp_119, gsl_tmp_118, gsl_tmp_123 <= gsl_tmp_125), select(gsl_tmp_119, gsl_tmp_117, gsl_tmp_121 <= gsl_tmp_125), gsl_tmp_121 <= gsl_tmp_123), gsl_tmp_111, gsl_tmp_107 >= 0.0);
  let gsl_tmp_128 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(3.0, 2.0, 0.0) * gsl_tmp_126 + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_126 * gsl_tmp_126;
  let gsl_tmp_134 = p - gsl_tmp_128;
  let gsl_tmp_129 = vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -4.0, 0.0) * gsl_tmp_126;
  let gsl_tmp_131 = gsl_tmp_129 / (length(gsl_tmp_129) + 0.000000000001);
  let gsl_tmp_132 = vec3<f32>(0.0, 0.0, 1.0) - gsl_tmp_131 * dot(vec3<f32>(0.0, 0.0, 1.0), gsl_tmp_131);
  let gsl_tmp_133 = gsl_tmp_132 / (length(gsl_tmp_132) + 0.000000000001);
  let gsl_tmp_135 = dot(gsl_tmp_134, cross(gsl_tmp_131, gsl_tmp_133));
  let gsl_tmp_137 = builtin_mix(1.0, 1.5, gsl_tmp_126);
  let gsl_tmp_136 = dot(gsl_tmp_134, gsl_tmp_133);
  let gsl_tmp_138 = builtin_mix(0.6666667, 1.0, gsl_tmp_126);
  let taperedFillet = builtin_gcl_closure_intersectRound_3(vec3<f32>(builtin_mix(gsl_remap(vec3<f32>(gsl_tmp_135 / gsl_tmp_137, gsl_tmp_136 / gsl_tmp_137, 0.0)) * gsl_tmp_137, builtin_gcl_closure_rectangleXYCentered(vec3<f32>(gsl_tmp_135 / gsl_tmp_138, gsl_tmp_136 / gsl_tmp_138, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(0.8, 0.8)) * gsl_tmp_138, gsl_tmp_126), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.8320503, 0.5547002, 0.0)), dot(p - vec3<f32>(3.0, 0.0, 0.0), vec3<f32>(0.8320503, -0.5547002, 0.0))));
  return taperedFillet;
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let gsl_tmp_139 = vec3<f32>(0.0, 0.0, 0.0) - p;
  let gsl_tmp_140 = 0.25 * (6.5 + dot(gsl_tmp_139, vec3<f32>(0.0, -2.0, 0.0))) / 3.0;
  let gsl_tmp_142 = -0.5 * (0.5 - 3.0 * gsl_tmp_140) + 0.25 * dot(gsl_tmp_139, vec3<f32>(1.5, 1.0, 0.0));
  let gsl_tmp_141 = gsl_tmp_140 - 0.25;
  let gsl_tmp_149 = sqrt(max(-gsl_tmp_141, 0.0));
  let gsl_tmp_150 = acos(clamp(gsl_tmp_142 / (2.0 * gsl_tmp_141 * gsl_tmp_149 - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let gsl_tmp_152 = sin(gsl_tmp_150) * 1.7320508;
  let gsl_tmp_151 = cos(gsl_tmp_150);
  let gsl_tmp_155 = clamp((gsl_tmp_152 - gsl_tmp_151) * gsl_tmp_149 - -0.5, 0.0, 1.0);
  let gsl_tmp_154 = clamp((-gsl_tmp_152 - gsl_tmp_151) * gsl_tmp_149 - -0.5, 0.0, 1.0);
  let gsl_tmp_158 = gsl_tmp_139 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_154) * gsl_tmp_154;
  let gsl_tmp_159 = dot(gsl_tmp_158, gsl_tmp_158);
  let gsl_tmp_160 = gsl_tmp_139 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_155) * gsl_tmp_155;
  let gsl_tmp_161 = dot(gsl_tmp_160, gsl_tmp_160);
  let gsl_tmp_153 = clamp((gsl_tmp_151 + gsl_tmp_151) * gsl_tmp_149 - -0.5, 0.0, 1.0);
  let gsl_tmp_156 = gsl_tmp_139 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_153) * gsl_tmp_153;
  let gsl_tmp_157 = dot(gsl_tmp_156, gsl_tmp_156);
  let gsl_tmp_143 = gsl_tmp_142 * gsl_tmp_142 + 4.0 * gsl_tmp_141 * gsl_tmp_141 * gsl_tmp_141;
  let gsl_tmp_144 = sqrt(max(gsl_tmp_143, 0.0));
  let gsl_tmp_145 = (gsl_tmp_144 - gsl_tmp_142) / 2.0;
  let gsl_tmp_146 = (-gsl_tmp_144 - gsl_tmp_142) / 2.0;
  let gsl_tmp_147 = clamp(sign(gsl_tmp_145) * exp(log(max(abs(gsl_tmp_145), 0.000000000001)) / 3.0) + sign(gsl_tmp_146) * exp(log(max(abs(gsl_tmp_146), 0.000000000001)) / 3.0) - -0.5, 0.0, 1.0);
  let gsl_tmp_162 = select(select(select(gsl_tmp_155, gsl_tmp_154, gsl_tmp_159 <= gsl_tmp_161), select(gsl_tmp_155, gsl_tmp_153, gsl_tmp_157 <= gsl_tmp_161), gsl_tmp_157 <= gsl_tmp_159), gsl_tmp_147, gsl_tmp_143 >= 0.0);
  let gsl_tmp_164 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(3.0, 2.0, 0.0) * gsl_tmp_162 + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_162 * gsl_tmp_162;
  let gsl_tmp_170 = p - gsl_tmp_164;
  let gsl_tmp_165 = vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -4.0, 0.0) * gsl_tmp_162;
  let gsl_tmp_167 = gsl_tmp_165 / (length(gsl_tmp_165) + 0.000000000001);
  let gsl_tmp_168 = vec3<f32>(0.0, 0.0, 1.0) - gsl_tmp_167 * dot(vec3<f32>(0.0, 0.0, 1.0), gsl_tmp_167);
  let gsl_tmp_169 = gsl_tmp_168 / (length(gsl_tmp_168) + 0.000000000001);
  let gsl_tmp_171 = dot(gsl_tmp_170, cross(gsl_tmp_167, gsl_tmp_169));
  let gsl_tmp_173 = clamp(gsl_tmp_162, 0.0, 1.0);
  let gsl_tmp_174 = gsl_tmp_173 * gsl_tmp_173 * (3.0 - 2.0 * gsl_tmp_173);
  let gsl_tmp_175 = builtin_mix(1.0, 1.5, gsl_tmp_174);
  let gsl_tmp_172 = dot(gsl_tmp_170, gsl_tmp_169);
  let gsl_tmp_176 = builtin_mix(0.6666667, 1.0, gsl_tmp_174);
  let eased = builtin_intersectSharp_2(vec3<f32>(builtin_mix(gsl_remap(vec3<f32>(gsl_tmp_171 / gsl_tmp_175, gsl_tmp_172 / gsl_tmp_175, 0.0)) * gsl_tmp_175, builtin_gcl_closure_rectangleXYCentered(vec3<f32>(gsl_tmp_171 / gsl_tmp_176, gsl_tmp_172 / gsl_tmp_176, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(0.8, 0.8)) * gsl_tmp_176, gsl_tmp_174), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.8320503, 0.5547002, 0.0)), dot(p - vec3<f32>(3.0, 0.0, 0.0), vec3<f32>(0.8320503, -0.5547002, 0.0))));
  return eased;
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  let gsl_tmp_177 = vec3<f32>(0.0, 0.0, 0.0) - p;
  let gsl_tmp_178 = 0.25 * (6.5 + dot(gsl_tmp_177, vec3<f32>(0.0, -2.0, 0.0))) / 3.0;
  let gsl_tmp_180 = -0.5 * (0.5 - 3.0 * gsl_tmp_178) + 0.25 * dot(gsl_tmp_177, vec3<f32>(1.5, 1.0, 0.0));
  let gsl_tmp_179 = gsl_tmp_178 - 0.25;
  let gsl_tmp_187 = sqrt(max(-gsl_tmp_179, 0.0));
  let gsl_tmp_188 = acos(clamp(gsl_tmp_180 / (2.0 * gsl_tmp_179 * gsl_tmp_187 - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let gsl_tmp_190 = sin(gsl_tmp_188) * 1.7320508;
  let gsl_tmp_189 = cos(gsl_tmp_188);
  let gsl_tmp_193 = clamp((gsl_tmp_190 - gsl_tmp_189) * gsl_tmp_187 - -0.5, 0.0, 1.0);
  let gsl_tmp_192 = clamp((-gsl_tmp_190 - gsl_tmp_189) * gsl_tmp_187 - -0.5, 0.0, 1.0);
  let gsl_tmp_196 = gsl_tmp_177 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_192) * gsl_tmp_192;
  let gsl_tmp_197 = dot(gsl_tmp_196, gsl_tmp_196);
  let gsl_tmp_198 = gsl_tmp_177 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_193) * gsl_tmp_193;
  let gsl_tmp_199 = dot(gsl_tmp_198, gsl_tmp_198);
  let gsl_tmp_191 = clamp((gsl_tmp_189 + gsl_tmp_189) * gsl_tmp_187 - -0.5, 0.0, 1.0);
  let gsl_tmp_194 = gsl_tmp_177 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_191) * gsl_tmp_191;
  let gsl_tmp_195 = dot(gsl_tmp_194, gsl_tmp_194);
  let gsl_tmp_181 = gsl_tmp_180 * gsl_tmp_180 + 4.0 * gsl_tmp_179 * gsl_tmp_179 * gsl_tmp_179;
  let gsl_tmp_182 = sqrt(max(gsl_tmp_181, 0.0));
  let gsl_tmp_183 = (gsl_tmp_182 - gsl_tmp_180) / 2.0;
  let gsl_tmp_184 = (-gsl_tmp_182 - gsl_tmp_180) / 2.0;
  let gsl_tmp_185 = clamp(sign(gsl_tmp_183) * exp(log(max(abs(gsl_tmp_183), 0.000000000001)) / 3.0) + sign(gsl_tmp_184) * exp(log(max(abs(gsl_tmp_184), 0.000000000001)) / 3.0) - -0.5, 0.0, 1.0);
  let gsl_tmp_200 = select(select(select(gsl_tmp_193, gsl_tmp_192, gsl_tmp_197 <= gsl_tmp_199), select(gsl_tmp_193, gsl_tmp_191, gsl_tmp_195 <= gsl_tmp_199), gsl_tmp_195 <= gsl_tmp_197), gsl_tmp_185, gsl_tmp_181 >= 0.0);
  let gsl_tmp_202 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(3.0, 2.0, 0.0) * gsl_tmp_200 + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_200 * gsl_tmp_200;
  let gsl_tmp_208 = p - gsl_tmp_202;
  let gsl_tmp_203 = vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -4.0, 0.0) * gsl_tmp_200;
  let gsl_tmp_205 = gsl_tmp_203 / (length(gsl_tmp_203) + 0.000000000001);
  let gsl_tmp_206 = vec3<f32>(0.0, 0.0, 1.0) - gsl_tmp_205 * dot(vec3<f32>(0.0, 0.0, 1.0), gsl_tmp_205);
  let gsl_tmp_207 = gsl_tmp_206 / (length(gsl_tmp_206) + 0.000000000001);
  let gsl_tmp_209 = dot(gsl_tmp_208, cross(gsl_tmp_205, gsl_tmp_207));
  let gsl_tmp_211 = clamp(gsl_tmp_200, 0.0, 1.0);
  let gsl_tmp_212 = gsl_tmp_211 * gsl_tmp_211 * (3.0 - 2.0 * gsl_tmp_211);
  let gsl_tmp_213 = builtin_mix(1.0, 1.5, gsl_tmp_212);
  let gsl_tmp_210 = dot(gsl_tmp_208, gsl_tmp_207);
  let gsl_tmp_214 = builtin_mix(0.6666667, 1.0, gsl_tmp_212);
  let easedChamfer = builtin_gcl_closure_intersectChamfer(vec3<f32>(builtin_mix(gsl_remap(vec3<f32>(gsl_tmp_209 / gsl_tmp_213, gsl_tmp_210 / gsl_tmp_213, 0.0)) * gsl_tmp_213, builtin_gcl_closure_rectangleXYCentered(vec3<f32>(gsl_tmp_209 / gsl_tmp_214, gsl_tmp_210 / gsl_tmp_214, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(0.8, 0.8)) * gsl_tmp_214, gsl_tmp_212), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.8320503, 0.5547002, 0.0)), dot(p - vec3<f32>(3.0, 0.0, 0.0), vec3<f32>(0.8320503, -0.5547002, 0.0))));
  return easedChamfer;
}

fn builtin_unionSharp(fields: array<f32, 6>) -> f32 {
  return min(min(min(min(min(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]);
}

fn builtin_intersectSharp(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound(fields: vec2<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_rectangleXYCentered(p: vec3<f32>, center: vec2<f32>, size: vec2<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p.xy - center) - halfSize;
  return builtin_gcl_closure_intersectRound(vec2<f32>(d.x, d.y));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = vec3<f32>(0.0, 0.0, 0.0) - p;
  let gsl_tmp_2 = 0.25 * (6.5 + dot(gsl_tmp_1, vec3<f32>(0.0, -2.0, 0.0))) / 3.0;
  let gsl_tmp_4 = -0.5 * (0.5 - 3.0 * gsl_tmp_2) + 0.25 * dot(gsl_tmp_1, vec3<f32>(1.5, 1.0, 0.0));
  let gsl_tmp_3 = gsl_tmp_2 - 0.25;
  let gsl_tmp_11 = sqrt(max(-gsl_tmp_3, 0.0));
  let gsl_tmp_12 = acos(clamp(gsl_tmp_4 / (2.0 * gsl_tmp_3 * gsl_tmp_11 - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let gsl_tmp_14 = sin(gsl_tmp_12) * 1.7320508;
  let gsl_tmp_13 = cos(gsl_tmp_12);
  let gsl_tmp_17 = clamp((gsl_tmp_14 - gsl_tmp_13) * gsl_tmp_11 - -0.5, 0.0, 1.0);
  let gsl_tmp_16 = clamp((-gsl_tmp_14 - gsl_tmp_13) * gsl_tmp_11 - -0.5, 0.0, 1.0);
  let gsl_tmp_20 = gsl_tmp_1 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_16) * gsl_tmp_16;
  let gsl_tmp_21 = dot(gsl_tmp_20, gsl_tmp_20);
  let gsl_tmp_22 = gsl_tmp_1 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_17) * gsl_tmp_17;
  let gsl_tmp_23 = dot(gsl_tmp_22, gsl_tmp_22);
  let gsl_tmp_15 = clamp((gsl_tmp_13 + gsl_tmp_13) * gsl_tmp_11 - -0.5, 0.0, 1.0);
  let gsl_tmp_18 = gsl_tmp_1 + (vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_15) * gsl_tmp_15;
  let gsl_tmp_19 = dot(gsl_tmp_18, gsl_tmp_18);
  let gsl_tmp_5 = gsl_tmp_4 * gsl_tmp_4 + 4.0 * gsl_tmp_3 * gsl_tmp_3 * gsl_tmp_3;
  let gsl_tmp_6 = sqrt(max(gsl_tmp_5, 0.0));
  let gsl_tmp_7 = (gsl_tmp_6 - gsl_tmp_4) / 2.0;
  let gsl_tmp_8 = (-gsl_tmp_6 - gsl_tmp_4) / 2.0;
  let gsl_tmp_9 = clamp(sign(gsl_tmp_7) * exp(log(max(abs(gsl_tmp_7), 0.000000000001)) / 3.0) + sign(gsl_tmp_8) * exp(log(max(abs(gsl_tmp_8), 0.000000000001)) / 3.0) - -0.5, 0.0, 1.0);
  let gsl_tmp_24 = select(select(select(gsl_tmp_17, gsl_tmp_16, gsl_tmp_21 <= gsl_tmp_23), select(gsl_tmp_17, gsl_tmp_15, gsl_tmp_19 <= gsl_tmp_23), gsl_tmp_19 <= gsl_tmp_21), gsl_tmp_9, gsl_tmp_5 >= 0.0);
  let gsl_tmp_26 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(3.0, 2.0, 0.0) * gsl_tmp_24 + vec3<f32>(0.0, -2.0, 0.0) * gsl_tmp_24 * gsl_tmp_24;
  let gsl_tmp_32 = p - gsl_tmp_26;
  let gsl_tmp_27 = vec3<f32>(3.0, 2.0, 0.0) + vec3<f32>(0.0, -4.0, 0.0) * gsl_tmp_24;
  let gsl_tmp_29 = gsl_tmp_27 / (length(gsl_tmp_27) + 0.000000000001);
  let gsl_tmp_30 = vec3<f32>(0.0, 0.0, 1.0) - gsl_tmp_29 * dot(vec3<f32>(0.0, 0.0, 1.0), gsl_tmp_29);
  let gsl_tmp_31 = gsl_tmp_30 / (length(gsl_tmp_30) + 0.000000000001);
  let gsl_tmp_33 = vec3<f32>(dot(gsl_tmp_32, cross(gsl_tmp_29, gsl_tmp_31)), dot(gsl_tmp_32, gsl_tmp_31), 0.0);
  let plain = builtin_intersectSharp_2(vec3<f32>(builtin_mix(gsl_remap(gsl_tmp_33), builtin_gcl_closure_rectangleXYCentered(gsl_tmp_33, vec2<f32>(0.0, 0.0), vec2<f32>(0.8, 0.8)), gsl_tmp_24), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.8320503, 0.5547002, 0.0)), dot(p - vec3<f32>(3.0, 0.0, 0.0), vec3<f32>(0.8320503, -0.5547002, 0.0))));
  let gsl_s1 = gsl_shape(builtin_unionSharp(array<f32, 6>(plain, gsl_remap_2(p - vec3<f32>(0.0, 3.0, 0.0)), gsl_remap_3(p - vec3<f32>(0.0, 6.0, 0.0)), gsl_remap_4(p - vec3<f32>(0.0, 9.0, 0.0)), gsl_remap_5(p - vec3<f32>(0.0, 12.0, 0.0)), gsl_remap_6(p - vec3<f32>(0.0, 15.0, 0.0)))), 1u);
  return gsl_s1;
}
