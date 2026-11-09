fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_2 = sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
  let gsl_tmp_1 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8);
  let gsl_tmp_3 = select(gsl_tmp_2.z, select(gsl_tmp_2.y, gsl_tmp_2.x, gsl_tmp_1.x > gsl_tmp_1.y), max(gsl_tmp_1.x, gsl_tmp_1.y) > gsl_tmp_1.z);
  let gsl_tmp_5 = select(0.0, gsl_tmp_2.x, gsl_tmp_1.x + 0.0 > 0.0);
  let gsl_tmp_6 = select(0.0, gsl_tmp_2.y, gsl_tmp_1.y + 0.0 > 0.0);
  let gsl_tmp_7 = select(0.0, gsl_tmp_2.z, gsl_tmp_1.z + 0.0 > 0.0);
  let gsl_tmp_4 = 0.0 + max(gsl_tmp_1.x + 0.0, 0.0) * max(gsl_tmp_1.x + 0.0, 0.0) + max(gsl_tmp_1.y + 0.0, 0.0) * max(gsl_tmp_1.y + 0.0, 0.0) + max(gsl_tmp_1.z + 0.0, 0.0) * max(gsl_tmp_1.z + 0.0, 0.0);
  let gsl_tmp_9 = sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
  let gsl_tmp_8 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8);
  let gsl_tmp_10 = select(gsl_tmp_9.z, select(gsl_tmp_9.y, gsl_tmp_9.x, gsl_tmp_8.x > gsl_tmp_8.y), max(gsl_tmp_8.x, gsl_tmp_8.y) > gsl_tmp_8.z);
  let gsl_tmp_12 = select(0.0, gsl_tmp_9.x, gsl_tmp_8.x + 0.0 > 0.0);
  let gsl_tmp_13 = select(0.0, gsl_tmp_9.y, gsl_tmp_8.y + 0.0 > 0.0);
  let gsl_tmp_14 = select(0.0, gsl_tmp_9.z, gsl_tmp_8.z + 0.0 > 0.0);
  let gsl_tmp_11 = 0.0 + max(gsl_tmp_8.x + 0.0, 0.0) * max(gsl_tmp_8.x + 0.0, 0.0) + max(gsl_tmp_8.y + 0.0, 0.0) * max(gsl_tmp_8.y + 0.0, 0.0) + max(gsl_tmp_8.z + 0.0, 0.0) * max(gsl_tmp_8.z + 0.0, 0.0);
  let gsl_tmp_16 = sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
  let gsl_tmp_15 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8);
  let gsl_tmp_17 = select(gsl_tmp_16.z, select(gsl_tmp_16.y, gsl_tmp_16.x, gsl_tmp_15.x > gsl_tmp_15.y), max(gsl_tmp_15.x, gsl_tmp_15.y) > gsl_tmp_15.z);
  let gsl_tmp_19 = select(0.0, gsl_tmp_16.x, gsl_tmp_15.x + 0.0 > 0.0);
  let gsl_tmp_20 = select(0.0, gsl_tmp_16.y, gsl_tmp_15.y + 0.0 > 0.0);
  let gsl_tmp_21 = select(0.0, gsl_tmp_16.z, gsl_tmp_15.z + 0.0 > 0.0);
  let gsl_tmp_18 = 0.0 + max(gsl_tmp_15.x + 0.0, 0.0) * max(gsl_tmp_15.x + 0.0, 0.0) + max(gsl_tmp_15.y + 0.0, 0.0) * max(gsl_tmp_15.y + 0.0, 0.0) + max(gsl_tmp_15.z + 0.0, 0.0) * max(gsl_tmp_15.z + 0.0, 0.0);
  let notch = gcl_closure_intersectRoundExact_2(p, vec3<f32>(gsl_tmp_3 - select(0.0, gsl_tmp_3, max(max(gsl_tmp_1.x, gsl_tmp_1.y), gsl_tmp_1.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_5 * max(gsl_tmp_1.x + 0.0, 0.0) + max(gsl_tmp_1.x + 0.0, 0.0) * gsl_tmp_5 + gsl_tmp_6 * max(gsl_tmp_1.y + 0.0, 0.0) + max(gsl_tmp_1.y + 0.0, 0.0) * gsl_tmp_6 + gsl_tmp_7 * max(gsl_tmp_1.z + 0.0, 0.0) + max(gsl_tmp_1.z + 0.0, 0.0) * gsl_tmp_7, gsl_tmp_4 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_4, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_4, 0.000000000000000000000000000001)) > 0.0)))), gsl_tmp_10 - select(0.0, gsl_tmp_10, max(max(gsl_tmp_8.x, gsl_tmp_8.y), gsl_tmp_8.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_12 * max(gsl_tmp_8.x + 0.0, 0.0) + max(gsl_tmp_8.x + 0.0, 0.0) * gsl_tmp_12 + gsl_tmp_13 * max(gsl_tmp_8.y + 0.0, 0.0) + max(gsl_tmp_8.y + 0.0, 0.0) * gsl_tmp_13 + gsl_tmp_14 * max(gsl_tmp_8.z + 0.0, 0.0) + max(gsl_tmp_8.z + 0.0, 0.0) * gsl_tmp_14, gsl_tmp_11 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_11, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_11, 0.000000000000000000000000000001)) > 0.0)))), gsl_tmp_17 - select(0.0, gsl_tmp_17, max(max(gsl_tmp_15.x, gsl_tmp_15.y), gsl_tmp_15.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_19 * max(gsl_tmp_15.x + 0.0, 0.0) + max(gsl_tmp_15.x + 0.0, 0.0) * gsl_tmp_19 + gsl_tmp_20 * max(gsl_tmp_15.y + 0.0, 0.0) + max(gsl_tmp_15.y + 0.0, 0.0) * gsl_tmp_20 + gsl_tmp_21 * max(gsl_tmp_15.z + 0.0, 0.0) + max(gsl_tmp_15.z + 0.0, 0.0) * gsl_tmp_21, gsl_tmp_18 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_18, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_18, 0.000000000000000000000000000001)) > 0.0))))), vec3<f32>(dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0)))));
  return notch;
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn gsl_gcl_closure_unionRoundExact_2_remap(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0;
}

fn gcl_closure_unionRoundExact_2(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  let d1 = length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0;
  let d2 = length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0;
  let sumScale = length(normal1 + normal2);
  let sum = (d1 + d2) / sumScale;
  let diffScale = length(normal1 - normal2);
  let diff = (d1 - d2) / diffScale;
  let sharp = min(d1, d2);
  let field1OnSurface2 = gsl_gcl_closure_unionRoundExact_2_remap(p - normal2 * d2, normal1, normal2);
  let field2OnSurface1 = builtin_sphere(p - normal1 * d1, vec3<f32>(1.2, 0.0, 0.0), 1.0);
  return select(-length(vec2<f32>(sum, diff)), sharp, max(field1OnSurface2, field2OnSurface1) > 0.0);
}

fn gsl_gcl_closure_unionRoundExact_5_remap(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6));
}

fn gsl_gcl_closure_unionRoundExact_5_remap_2(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -(length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0);
}

fn gcl_closure_unionRoundExact_5(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  let d1 = -builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6));
  let d2 = -(length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0);
  let sumScale = length(normal1 + normal2);
  let sum = (d1 + d2) / sumScale;
  let diffScale = length(normal1 - normal2);
  let diff = (d1 - d2) / diffScale;
  let sharp = min(d1, d2);
  let field1OnSurface2 = gsl_gcl_closure_unionRoundExact_5_remap(p - normal2 * d2, normal1, normal2);
  let field2OnSurface1 = gsl_gcl_closure_unionRoundExact_5_remap_2(p - normal1 * d1, normal1, normal2);
  return select(-length(vec2<f32>(sum, diff)), sharp, max(field1OnSurface2, field2OnSurface1) > 0.0);
}

fn gcl_closure_intersectRoundExact_2(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -gcl_closure_unionRoundExact_5(p, -normal1, -normal2);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let lens = gcl_closure_unionRoundExact_2(p, vec3<f32>(dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(0.0, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.0, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(0.0, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.0, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(0.0, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.0, 0.0, 0.0)) > 0.0)))), vec3<f32>(dot(p - vec3<f32>(1.2, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(1.2, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(1.2, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(1.2, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(1.2, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(1.2, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(1.2, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(1.2, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(1.2, 0.0, 0.0)) > 0.0)))));
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(lens, gsl_remap(p - vec3<f32>(4.0, 0.0, 0.0)))), 1u);
  return gsl_s1;
}
