fn gsl_remap(p: vec3<f32>) -> f32 {
  let a = p.x;
  return a;
}

fn boundaryMap__d(v: f32, v__dp0: f32, v__dp1: f32, v__dp2: f32) -> vec3<f32> {
  return vec3<f32>(v__dp0, v__dp1, v__dp2) * v;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let a = p.x;
  let b = dot(p - vec3<f32>(0.0, 0.0, 0.0), normalize(vec3<f32>(1.0, 1.0, 0.0)));
  let body = abs(max(a, b));
  let ab = abs(gsl_remap(p - boundaryMap__d(b, 0.70710677, 0.70710677, 0.0)));
  let gsl_s1 = gsl_shape(body, 1u);
  let gsl_s2 = gsl_shape(ab, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
