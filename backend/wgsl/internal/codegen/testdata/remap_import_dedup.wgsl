fn gsl_remap(p: vec3<f32>, pkg1_somePi: f32) -> f32 {
  return length(p) - pkg1_somePi;
}

fn gsl_remap_2(p: vec3<f32>, pkg1_somePi: f32) -> f32 {
  return gsl_remap(vec3<f32>(p.x, p.y, p.z), pkg1_somePi) + 0.0;
}

fn gsl_remap_3(p: vec3<f32>, pkg1_somePi: f32) -> f32 {
  let side = gsl_remap_2(vec3<f32>(p.x - 1.0, p.y, p.z), pkg1_somePi);
  return side;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let pkg1_somePi = 3.14159;
  let s1 = gsl_remap_3(vec3<f32>(p.x, p.y, p.z), pkg1_somePi);
  let s2 = gsl_remap_3(vec3<f32>(p.x + 0.5, p.y, p.z), pkg1_somePi);
  let gsl_s1 = gsl_shape(min(s1, s2), 1u);
  return gsl_s1;
}
