fn gsl_builtin_unionRoundExact_remap(p: vec3<f32>, a: f32, b: f32, normalA: vec3<f32>, normalB: vec3<f32>, radius: f32) -> f32 {
  return a;
}

fn gsl_builtin_unionRoundExact_remap_2(p: vec3<f32>, a: f32, b: f32, normalA: vec3<f32>, normalB: vec3<f32>, radius: f32) -> f32 {
  return b;
}

fn builtin_unionRoundExact(p: vec3<f32>, a: f32, b: f32, radius: f32, normalA: vec3<f32>, normalB: vec3<f32>) -> f32 {
  let d1 = a - radius;
  let d2 = b - radius;
  let sumScale = length(normalA + normalB);
  let sum = (d1 + d2) / sumScale;
  let diffScale = length(normalA - normalB);
  let diff = (d1 - d2) / diffScale;
  let sharp = min(d1, d2);
  let field1OnSurface2 = gsl_builtin_unionRoundExact_remap(p - normalB * d2, a, b, normalA, normalB, radius) - radius;
  let field2OnSurface1 = gsl_builtin_unionRoundExact_remap_2(p - normalA * d1, a, b, normalA, normalB, radius) - radius;
  return select(-length(vec2<f32>(sum, diff)), sharp, max(field1OnSurface2, field2OnSurface1) > 0.0) + radius;
}

fn unionRoundExactAuto__d(p: vec3<f32>, sdf1: f32, sdf2: f32, sdf1__dp0: f32, sdf1__dp1: f32, sdf1__dp2: f32, sdf2__dp0: f32, sdf2__dp1: f32, sdf2__dp2: f32) -> f32 {
  return builtin_unionRoundExact(p, sdf1, sdf2, 0.0, vec3<f32>(sdf1__dp0, sdf1__dp1, sdf1__dp2), vec3<f32>(sdf2__dp0, sdf2__dp1, sdf2__dp2));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let blended = unionRoundExactAuto__d(p, length(p - vec3<f32>(-0.5, 0.0, 0.0)) - 1.0, length(p - vec3<f32>(0.5, 0.0, 0.0)) - 1.0, dot(clamp((p - vec3<f32>(-0.5, 0.0, 0.0)) / max(length(p - vec3<f32>(-0.5, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(1.0, 0.0, 0.0)), dot(clamp((p - vec3<f32>(-0.5, 0.0, 0.0)) / max(length(p - vec3<f32>(-0.5, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 1.0, 0.0)), dot(clamp((p - vec3<f32>(-0.5, 0.0, 0.0)) / max(length(p - vec3<f32>(-0.5, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 0.0, 1.0)), dot(clamp((p - vec3<f32>(0.5, 0.0, 0.0)) / max(length(p - vec3<f32>(0.5, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(1.0, 0.0, 0.0)), dot(clamp((p - vec3<f32>(0.5, 0.0, 0.0)) / max(length(p - vec3<f32>(0.5, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 1.0, 0.0)), dot(clamp((p - vec3<f32>(0.5, 0.0, 0.0)) / max(length(p - vec3<f32>(0.5, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 0.0, 1.0)));
  let sphereNormal = vec3<f32>(dot(clamp(p / max(length(p), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(1.0, 0.0, 0.0)), dot(clamp(p / max(length(p), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 1.0, 0.0)), dot(clamp(p / max(length(p), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 0.0, 1.0)));
  let gsl_tmp_1 = vec2<f32>(length(p.xy) - 2.0, p.z);
  let gsl_tmp_2 = vec2<f32>(length(p.xy) - 2.0, p.z);
  let gsl_tmp_3 = vec2<f32>(length(p.xy) - 2.0, p.z);
  let torusNormal = vec3<f32>(dot(clamp(gsl_tmp_1 / max(length(gsl_tmp_1), 0.00000000000000000001), vec2<f32>(-1.0), vec2<f32>(1.0)), vec2<f32>(dot(clamp(p.xy / max(length(p.xy), 0.00000000000000000001), vec2<f32>(-1.0), vec2<f32>(1.0)), vec2<f32>(1.0, 0.0)), 0.0)), dot(clamp(gsl_tmp_2 / max(length(gsl_tmp_2), 0.00000000000000000001), vec2<f32>(-1.0), vec2<f32>(1.0)), vec2<f32>(dot(clamp(p.xy / max(length(p.xy), 0.00000000000000000001), vec2<f32>(-1.0), vec2<f32>(1.0)), vec2<f32>(0.0, 1.0)), 0.0)), dot(clamp(gsl_tmp_3 / max(length(gsl_tmp_3), 0.00000000000000000001), vec2<f32>(-1.0), vec2<f32>(1.0)), vec2<f32>(dot(clamp(p.xy / max(length(p.xy), 0.00000000000000000001), vec2<f32>(-1.0), vec2<f32>(1.0)), vec2<f32>(0.0, 0.0)), 1.0)));
  let gsl_s1 = gsl_shape(blended + 0.0 * (sphereNormal.x + torusNormal.z), 1u);
  return gsl_s1;
}
