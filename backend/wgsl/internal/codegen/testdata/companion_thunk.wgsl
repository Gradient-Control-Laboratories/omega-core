fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn gcl_closure_unionRoundExact(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  let d1 = length(p - vec3<f32>(-0.5, 0.0, 0.0)) - 1.0;
  let d2 = length(p - vec3<f32>(0.5, 0.0, 0.0)) - 1.0;
  let sumScale = length(normal1 + normal2);
  let sum = (d1 + d2) / sumScale;
  let diffScale = length(normal1 - normal2);
  let diff = (d1 - d2) / diffScale;
  let sharp = min(d1, d2);
  let field1OnSurface2 = builtin_sphere(p - normal2 * d2, vec3<f32>(-0.5, 0.0, 0.0), 1.0);
  let field2OnSurface1 = builtin_sphere(p - normal1 * d1, vec3<f32>(0.5, 0.0, 0.0), 1.0);
  return select(-length(vec2<f32>(sum, diff)), sharp, max(field1OnSurface2, field2OnSurface1) > 0.0);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let blended = gcl_closure_unionRoundExact(p, vec3<f32>(dot(p - vec3<f32>(-0.5, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(-0.5, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(-0.5, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(-0.5, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(-0.5, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(-0.5, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(-0.5, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(-0.5, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(-0.5, 0.0, 0.0)) > 0.0)))), vec3<f32>(dot(p - vec3<f32>(0.5, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(0.5, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.5, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.5, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(0.5, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.5, 0.0, 0.0)) > 0.0))), dot(p - vec3<f32>(0.5, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(0.5, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.5, 0.0, 0.0)) > 0.0)))));
  let sphereNormal = vec3<f32>(dot(p, vec3<f32>(1.0, 0.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), dot(p, vec3<f32>(0.0, 1.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), dot(p, vec3<f32>(0.0, 0.0, 1.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))));
  let gsl_tmp_1 = vec2<f32>(length(p.xy) - 2.0, p.z);
  let gsl_tmp_2 = vec2<f32>(length(p.xy) - 2.0, p.z);
  let gsl_tmp_3 = vec2<f32>(length(p.xy) - 2.0, p.z);
  let torusNormal = vec3<f32>(dot(gsl_tmp_1, vec2<f32>(dot(p.xy, vec2<f32>(1.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_1) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_1) > 0.0))), dot(gsl_tmp_2, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 1.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_2) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_2) > 0.0))), dot(gsl_tmp_3, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 1.0)) / (length(gsl_tmp_3) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_3) > 0.0))));
  let gsl_s1 = gsl_shape(blended + 0.0 * (sphereNormal.x + torusNormal.z), 1u);
  return gsl_s1;
}
