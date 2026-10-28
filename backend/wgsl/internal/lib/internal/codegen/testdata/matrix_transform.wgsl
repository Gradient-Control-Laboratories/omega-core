fn vector_multiplyPoint(m: mat4x4<f32>, point: vec3<f32>) -> vec3<f32> {
  let h = m * vec4<f32>(point.x, point.y, point.z, 1.0);
  return vec3<f32>(h.x, h.y, h.z);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return length(p) - 1.0;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let local = vector_multiplyPoint(mat4x4<f32>(vec4<f32>(0.5, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.5, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.5, 0.0), vec4<f32>(-1.5, -0.0, -0.0, 1.0)), p);
  let d = gsl_remap(local) * 2.0;
  let gsl_s1 = gsl_shape(d + 0.0, 1u);
  return gsl_s1;
}
