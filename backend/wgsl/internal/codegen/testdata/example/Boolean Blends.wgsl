fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn gsl_remap(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectSharp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0));
}

fn gsl_remap_2(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_28 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_29 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_30 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectSharp(vec3<f32>(gsl_tmp_28, -gsl_tmp_29, -gsl_tmp_30));
}

fn builtin_intersectLp(fields: vec3<f32>, width: f32, exponent: f32) -> f32 {
  let peak = builtin_intersectSharp(fields);
  let peakInset = max(peak + width, 0.0);
  let normalizer = max(peakInset, 0.000000001);
  let sumOfPowers = 0.0 + pow(max(fields.x + width, 0.0) / normalizer, exponent) + pow(max(fields.y + width, 0.0) / normalizer, exponent) + pow(max(fields.z + width, 0.0) / normalizer, exponent);
  return min(-width, peak) + peakInset * pow(sumOfPowers, 1.0 / exponent);
}

fn gsl_remap_3(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectLp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), 0.25, 1.0);
}

fn gsl_remap_4(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_31 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_32 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_33 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectLp(vec3<f32>(gsl_tmp_31, -gsl_tmp_32, -gsl_tmp_33), 0.25, 1.0);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn gsl_remap_5(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectRound(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), 0.25);
}

fn gsl_remap_6(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_34 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_35 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_36 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectRound(vec3<f32>(gsl_tmp_34, -gsl_tmp_35, -gsl_tmp_36), 0.25);
}

fn gsl_remap_7(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectLp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), 0.25, 3.0);
}

fn gsl_remap_8(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_37 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_38 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_39 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectLp(vec3<f32>(gsl_tmp_37, -gsl_tmp_38, -gsl_tmp_39), 0.25, 3.0);
}

fn builtin_intersectChamferEdges(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp(fields);
  return builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.x + fields.y + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields.x + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.z + setback) / 1.4142135, 0.0 > 0.0)))));
}

fn gsl_remap_9(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectChamferEdges(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), 0.25);
}

fn gsl_remap_10(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_40 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_41 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_42 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectChamferEdges(vec3<f32>(gsl_tmp_40, -gsl_tmp_41, -gsl_tmp_42), 0.25);
}

fn builtin_intersectChamferVertices(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp(fields);
  return builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.z + setback) / 1.7320508, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0)))))));
}

fn gsl_remap_11(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectChamferVertices(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), 0.25);
}

fn gsl_remap_12(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_43 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_44 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_45 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectChamferVertices(vec3<f32>(gsl_tmp_43, -gsl_tmp_44, -gsl_tmp_45), 0.25);
}

fn builtin_firstOf(list: vec2<f32>) -> f32 {
  return list.x;
}

fn builtin_lastOf(list: vec2<f32>) -> f32 {
  return list.y;
}

fn builtin_intersectChamfer(fields: vec3<f32>, setback: vec2<f32>) -> f32 {
  return max(max(builtin_intersectSharp(fields), builtin_intersectChamferEdges(fields, builtin_firstOf(setback))), builtin_intersectChamferVertices(fields, builtin_lastOf(setback)));
}

fn gsl_remap_13(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectChamfer(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), vec2<f32>(0.25, 0.5));
}

fn gsl_remap_14(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_46 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_47 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_48 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectChamfer(vec3<f32>(gsl_tmp_46, -gsl_tmp_47, -gsl_tmp_48), vec2<f32>(0.25, 0.5));
}

fn builtin_gcl_closure_intersectRound(radius: f32, fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, radius);
}

fn builtin_gcl_closure_intersectChamferBlend(radius: f32, fields: vec3<f32>, setback: vec2<f32>) -> f32 {
  return builtin_gcl_closure_intersectRound(radius, vec3<f32>(builtin_intersectSharp(fields), builtin_intersectChamferEdges(fields, builtin_firstOf(setback)), builtin_intersectChamferVertices(fields, builtin_lastOf(setback))));
}

fn builtin_intersectChamferRound(fields: vec3<f32>, setback: vec2<f32>, radius: f32) -> f32 {
  return builtin_gcl_closure_intersectChamferBlend(radius, fields, setback);
}

fn gsl_remap_15(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectChamferRound(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), vec2<f32>(0.25, 0.5), 0.3);
}

fn gsl_remap_16(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_49 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_50 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_51 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectChamferRound(vec3<f32>(gsl_tmp_49, -gsl_tmp_50, -gsl_tmp_51), vec2<f32>(0.25, 0.5), 0.3);
}

fn builtin_intersectExponential(fields: vec3<f32>, width: f32) -> f32 {
  let peak = builtin_intersectSharp(fields);
  let scale = width / 2.7182817;
  let total = 0.0 + exp((fields.x - peak) / scale) + exp((fields.y - peak) / scale) + exp((fields.z - peak) / scale);
  return peak + scale * log(total);
}

fn gsl_remap_17(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_intersectExponential(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0), 0.25);
}

fn gsl_remap_18(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_52 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_53 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_54 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  return builtin_intersectExponential(vec3<f32>(gsl_tmp_52, -gsl_tmp_53, -gsl_tmp_54), 0.25);
}

fn builtin_unionSharp(fields: vec3<f32>) -> f32 {
  return min(min(fields.x, fields.y), fields.z);
}

fn gsl_remap_19(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return max(builtin_intersectSharp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), -gcl_closure_remapEdges(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)));
}

fn gsl_remap_20(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_55 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_56 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_57 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  let gsl_tmp_25 = builtin_intersectSharp(vec3<f32>(gsl_tmp_55, -gsl_tmp_56, -gsl_tmp_57));
  return min(gsl_tmp_25, max(gcl_closure_remapEdges_3(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), gsl_tmp_25 - 0.5));
}

fn gsl_remap_21(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return max(builtin_intersectSharp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), -gcl_closure_remapEdges_4(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)));
}

fn gsl_remap_22(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_58 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_59 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_60 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  let gsl_tmp_26 = builtin_intersectSharp(vec3<f32>(gsl_tmp_58, -gsl_tmp_59, -gsl_tmp_60));
  return min(gsl_tmp_26, max(gcl_closure_remapEdges_6(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), gsl_tmp_26 - 0.5));
}

fn gsl_remap_23(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return max(builtin_intersectSharp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), -gcl_closure_remapEdges_7(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)));
}

fn gsl_remap_24(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_61 = length(p + vector_dirX) - 2.0;
  let gsl_tmp_62 = length(p - vector_dirX) - 2.0;
  let gsl_tmp_63 = length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0;
  let gsl_tmp_27 = builtin_intersectSharp(vec3<f32>(gsl_tmp_61, -gsl_tmp_62, -gsl_tmp_63));
  return min(gsl_tmp_27, max(gcl_closure_remapEdges_9(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), gsl_tmp_27 - 0.5));
}

fn gsl_remap_25(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let sharp = min(min(gsl_remap(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -sharpKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_2(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return sharp;
}

fn gsl_remap_26(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let lpThin = min(min(gsl_remap_3(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -lpThinKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_4(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return lpThin;
}

fn gsl_remap_27(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let rounded = min(min(gsl_remap_5(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -roundKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_6(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return rounded;
}

fn gsl_remap_28(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let lpFat = min(min(gsl_remap_7(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -lpFatKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_8(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return lpFat;
}

fn gsl_remap_29(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let chamferEdges = min(min(gsl_remap_9(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -chamferEdgesKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_10(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return chamferEdges;
}

fn gsl_remap_30(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let chamferVertices = min(min(gsl_remap_11(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -chamferVerticesKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_12(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return chamferVertices;
}

fn gsl_remap_31(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let chamfer = min(min(gsl_remap_13(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -chamferKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_14(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return chamfer;
}

fn gsl_remap_32(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let chamferBlend = min(min(gsl_remap_15(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -chamferBlendKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_16(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return chamferBlend;
}

fn gsl_remap_33(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let edgesCircle = min(min(gsl_remap_19(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), min(builtin_unionSharp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), gcl_closure_remapEdges_2(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_20(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return edgesCircle;
}

fn gsl_remap_34(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let edgesDiamond = min(min(gsl_remap_21(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), min(builtin_unionSharp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), gcl_closure_remapEdges_5(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_22(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return edgesDiamond;
}

fn gsl_remap_35(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let edgesSquare = min(min(gsl_remap_23(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), min(builtin_unionSharp(vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)), gcl_closure_remapEdges_8(p, vec3<f32>(length(p + vector_dirX) - 2.0, length(p - vector_dirX) - 2.0, length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_24(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return edgesSquare;
}

fn gsl_remap_36(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let exponential = min(min(gsl_remap_17(p - vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ), -expKernel(vec3<f32>(-(length(p + vector_dirX) - 2.0), -(length(p - vector_dirX) - 2.0), -(length(p - vec3<f32>(0.0, 0.0, 1.7320508)) - 2.0)))), gsl_remap_18(p + vec3<f32>(0.0, 0.0, 7.5), vector_dirX, vector_dirZ));
  return exponential;
}

fn sharpKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectSharp(list);
}

fn lpThinKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectLp(list, 0.25, 1.0);
}

fn roundKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectRound(list, 0.25);
}

fn lpFatKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectLp(list, 0.25, 3.0);
}

fn chamferEdgesKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectChamferEdges(list, 0.25);
}

fn chamferVerticesKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectChamferVertices(list, 0.25);
}

fn chamferKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectChamfer(list, vec2<f32>(0.25, 0.5));
}

fn chamferBlendKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectChamferRound(list, vec2<f32>(0.25, 0.5), 0.3);
}

fn expKernel(list: vec3<f32>) -> f32 {
  return builtin_intersectExponential(list, 0.25);
}

fn gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(p: vec3<f32>) -> f32 {
  return length(p.xy) - 0.25;
}

fn gcl_closure_remapEdges(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gcl_closure_remapEdges_2(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gcl_closure_remapEdges_3(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(p: vec3<f32>) -> f32 {
  return max(abs(p.x), abs(p.y)) - 0.25;
}

fn gcl_closure_remapEdges_4(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gcl_closure_remapEdges_5(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gcl_closure_remapEdges_6(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_4_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(p: vec3<f32>) -> f32 {
  return max(abs(p.x + p.y) * 0.5, abs(p.x - p.y) * 2.0) / 1.4142135 - 0.25;
}

fn gcl_closure_remapEdges_7(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gcl_closure_remapEdges_8(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gcl_closure_remapEdges_9(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_7_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let vector_dirX = vec3<f32>(1.0, 0.0, 0.0);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let gsl_s1 = gsl_shape(gsl_remap_25(p - vec3<f32>(0.0, -7.5, 0.0), vector_dirX, vector_dirZ), 1u);
  let gsl_s2 = gsl_shape(gsl_remap_26(p - vec3<f32>(0.0, -2.5, 0.0), vector_dirX, vector_dirZ), 2u);
  let gsl_s3 = gsl_shape(gsl_remap_27(p - vec3<f32>(0.0, 2.5, 0.0), vector_dirX, vector_dirZ), 3u);
  let gsl_s4 = gsl_shape(gsl_remap_28(p - vec3<f32>(0.0, 7.5, 0.0), vector_dirX, vector_dirZ), 4u);
  let gsl_s5 = gsl_shape(gsl_remap_29(p - vec3<f32>(-5.0, -7.5, 0.0), vector_dirX, vector_dirZ), 5u);
  let gsl_s6 = gsl_shape(gsl_remap_30(p - vec3<f32>(-5.0, -2.5, 0.0), vector_dirX, vector_dirZ), 6u);
  let gsl_s7 = gsl_shape(gsl_remap_31(p - vec3<f32>(-5.0, 2.5, 0.0), vector_dirX, vector_dirZ), 7u);
  let gsl_s8 = gsl_shape(gsl_remap_32(p - vec3<f32>(-5.0, 7.5, 0.0), vector_dirX, vector_dirZ), 8u);
  let gsl_s9 = gsl_shape(gsl_remap_33(p - vec3<f32>(5.0, -7.5, 0.0), vector_dirX, vector_dirZ), 9u);
  let gsl_s10 = gsl_shape(gsl_remap_34(p - vec3<f32>(5.0, -2.5, 0.0), vector_dirX, vector_dirZ), 10u);
  let gsl_s11 = gsl_shape(gsl_remap_35(p - vec3<f32>(5.0, 2.5, 0.0), vector_dirX, vector_dirZ), 11u);
  let gsl_s12 = gsl_shape(gsl_remap_36(p - vec3<f32>(5.0, 7.5, 0.0), vector_dirX, vector_dirZ), 12u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12);
}
