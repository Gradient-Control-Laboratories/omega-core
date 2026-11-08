fn gsl_remap(p: vec3<f32>) -> f32 {
  let ball = length(p) - 1.0;
  return ball;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let reflected = gsl_remap(vec3<f32>(-p.x, p.y, p.z));
  return reflected;
}

fn gsl_remap_3(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(clamp(p.x, 0.0, 1.0), 0.0, 0.0);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let moved = gsl_remap_2(p - vec3<f32>(4.0, 0.0, 0.0));
  let gsl_s1 = gsl_shape(moved, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(clamp(clamp(gsl_remap_3(vec3<f32>(-(p - vec3<f32>(4.0, 0.0, 0.0)).x, (p - vec3<f32>(4.0, 0.0, 0.0)).y, (p - vec3<f32>(4.0, 0.0, 0.0)).z)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
