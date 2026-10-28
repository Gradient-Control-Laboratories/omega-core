fn vector_multiplyPoint(m: mat4x4<f32>, point: vec3<f32>) -> vec3<f32> {
  let h = m * vec4<f32>(point.x, point.y, point.z, 1.0);
  return vec3<f32>(h.x, h.y, h.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let pt = vector_multiplyPoint(mat4x4<f32>(vec4<f32>(1.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 1.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(-10.0, -20.0, -30.0, 1.0)), p);
  let pkg1_somePi = 3.14159;
  let branch_pkg_shaftRadius = pkg1_somePi;
  let gsl_s1 = gsl_shape(length(pt) - branch_pkg_shaftRadius, 1u);
  return gsl_s1;
}
