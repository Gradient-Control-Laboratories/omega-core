fn gsl_remap(p: vec3<f32>) -> f32 {
  return length(p);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return dot(p - vec3<f32>(0.5, 0.0, 0.0), vec3<f32>(0.70710677, 0.0, 0.70710677));
}

fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn builtin_sphere(p: vec3<f32>, radius: f32) -> f32 {
  return length(p) - radius;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let viaBuiltin = gsl_remap(p - vec3<f32>(2.0, 0.0, 0.0));
  let viaExpr = gsl_remap_2(vec3<f32>(abs(p.x), p.y, p.z));
  let viaFn = builtin_plane(vec3<f32>(abs(p.x), p.y, p.z), vec3<f32>(0.5, 0.0, 0.0), vec3<f32>(0.70710677, 0.0, 0.70710677));
  let viaLib = builtin_sphere(p - vec3<f32>(2.0, 0.0, 0.0), 0.0);
  let gsl_s1 = gsl_shape(viaBuiltin, 1u);
  let gsl_s2 = gsl_shape(viaExpr, 2u);
  let gsl_s3 = gsl_shape(viaFn, 3u);
  let gsl_s4 = gsl_shape(viaLib, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}
