fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_2 = sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
  let gsl_tmp_1 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8);
  let gsl_tmp_3 = select(0.0, gsl_tmp_2.x, gsl_tmp_1.x + 0.0 > 0.0);
  let gsl_tmp_4 = select(0.0, gsl_tmp_2.y, gsl_tmp_1.y + 0.0 > 0.0);
  let gsl_tmp_5 = select(0.0, gsl_tmp_2.z, gsl_tmp_1.z + 0.0 > 0.0);
  let gsl_tmp_6 = sqrt(0.0 + max(gsl_tmp_1.x + 0.0, 0.0) * max(gsl_tmp_1.x + 0.0, 0.0) + max(gsl_tmp_1.y + 0.0, 0.0) * max(gsl_tmp_1.y + 0.0, 0.0) + max(gsl_tmp_1.z + 0.0, 0.0) * max(gsl_tmp_1.z + 0.0, 0.0));
  let gsl_tmp_8 = sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
  let gsl_tmp_7 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8);
  let gsl_tmp_9 = select(0.0, gsl_tmp_8.x, gsl_tmp_7.x + 0.0 > 0.0);
  let gsl_tmp_10 = select(0.0, gsl_tmp_8.y, gsl_tmp_7.y + 0.0 > 0.0);
  let gsl_tmp_11 = select(0.0, gsl_tmp_8.z, gsl_tmp_7.z + 0.0 > 0.0);
  let gsl_tmp_12 = sqrt(0.0 + max(gsl_tmp_7.x + 0.0, 0.0) * max(gsl_tmp_7.x + 0.0, 0.0) + max(gsl_tmp_7.y + 0.0, 0.0) * max(gsl_tmp_7.y + 0.0, 0.0) + max(gsl_tmp_7.z + 0.0, 0.0) * max(gsl_tmp_7.z + 0.0, 0.0));
  let gsl_tmp_14 = sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
  let gsl_tmp_13 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8);
  let gsl_tmp_15 = select(0.0, gsl_tmp_14.x, gsl_tmp_13.x + 0.0 > 0.0);
  let gsl_tmp_16 = select(0.0, gsl_tmp_14.y, gsl_tmp_13.y + 0.0 > 0.0);
  let gsl_tmp_17 = select(0.0, gsl_tmp_14.z, gsl_tmp_13.z + 0.0 > 0.0);
  let gsl_tmp_18 = sqrt(0.0 + max(gsl_tmp_13.x + 0.0, 0.0) * max(gsl_tmp_13.x + 0.0, 0.0) + max(gsl_tmp_13.y + 0.0, 0.0) * max(gsl_tmp_13.y + 0.0, 0.0) + max(gsl_tmp_13.z + 0.0, 0.0) * max(gsl_tmp_13.z + 0.0, 0.0));
  let notch = gcl_closure_intersectRoundExact_2(p, vec3<f32>(select(select(gsl_tmp_2.z, select(gsl_tmp_2.y, gsl_tmp_2.x, gsl_tmp_1.x > gsl_tmp_1.y), max(gsl_tmp_1.x, gsl_tmp_1.y) > gsl_tmp_1.z), -0.0, -0.0 < max(max(gsl_tmp_1.x, gsl_tmp_1.y), gsl_tmp_1.z)) + (gsl_tmp_3 * max(gsl_tmp_1.x + 0.0, 0.0) + max(gsl_tmp_1.x + 0.0, 0.0) * gsl_tmp_3 + gsl_tmp_4 * max(gsl_tmp_1.y + 0.0, 0.0) + max(gsl_tmp_1.y + 0.0, 0.0) * gsl_tmp_4 + gsl_tmp_5 * max(gsl_tmp_1.z + 0.0, 0.0) + max(gsl_tmp_1.z + 0.0, 0.0) * gsl_tmp_5) / (2.0 * (gsl_tmp_6 + 1.0 - select(0.0, 1.0, (gsl_tmp_6 > 0.0)))), select(select(gsl_tmp_8.z, select(gsl_tmp_8.y, gsl_tmp_8.x, gsl_tmp_7.x > gsl_tmp_7.y), max(gsl_tmp_7.x, gsl_tmp_7.y) > gsl_tmp_7.z), -0.0, -0.0 < max(max(gsl_tmp_7.x, gsl_tmp_7.y), gsl_tmp_7.z)) + (gsl_tmp_9 * max(gsl_tmp_7.x + 0.0, 0.0) + max(gsl_tmp_7.x + 0.0, 0.0) * gsl_tmp_9 + gsl_tmp_10 * max(gsl_tmp_7.y + 0.0, 0.0) + max(gsl_tmp_7.y + 0.0, 0.0) * gsl_tmp_10 + gsl_tmp_11 * max(gsl_tmp_7.z + 0.0, 0.0) + max(gsl_tmp_7.z + 0.0, 0.0) * gsl_tmp_11) / (2.0 * (gsl_tmp_12 + 1.0 - select(0.0, 1.0, (gsl_tmp_12 > 0.0)))), select(select(gsl_tmp_14.z, select(gsl_tmp_14.y, gsl_tmp_14.x, gsl_tmp_13.x > gsl_tmp_13.y), max(gsl_tmp_13.x, gsl_tmp_13.y) > gsl_tmp_13.z), -0.0, -0.0 < max(max(gsl_tmp_13.x, gsl_tmp_13.y), gsl_tmp_13.z)) + (gsl_tmp_15 * max(gsl_tmp_13.x + 0.0, 0.0) + max(gsl_tmp_13.x + 0.0, 0.0) * gsl_tmp_15 + gsl_tmp_16 * max(gsl_tmp_13.y + 0.0, 0.0) + max(gsl_tmp_13.y + 0.0, 0.0) * gsl_tmp_16 + gsl_tmp_17 * max(gsl_tmp_13.z + 0.0, 0.0) + max(gsl_tmp_13.z + 0.0, 0.0) * gsl_tmp_17) / (2.0 * (gsl_tmp_18 + 1.0 - select(0.0, 1.0, (gsl_tmp_18 > 0.0))))), vec3<f32>(dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0)))));
  return notch;
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
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
