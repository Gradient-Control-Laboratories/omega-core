fn gsl_remap(p: vec3<f32>) -> f32 {
  let g1 = gsl_sphere_remap_2(vec3<f32>(p.x, p.y + 12.0, p.z));
  return g1;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let g2 = gsl_sphere_remap_3(vec3<f32>(p.x, p.y + 12.0, p.z)) + 12.0;
  return g2;
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let g1 = gsl_sphere_remap_2(vec3<f32>(p.x, p.y + 12.0, p.z));
  let g2 = gsl_sphere_remap_3(vec3<f32>(p.x, p.y + 12.0, p.z)) + 12.0;
  return max(g1, g2);
}

fn dist(pt: vec3<f32>) -> f32 {
  return sqrt(pt.x * pt.x + pt.y * pt.y + pt.z * pt.z);
}

fn gsl_sphere_remap(p: vec3<f32>, radius: f32) -> f32 {
  return dist(p) - radius;
}

fn sphere(p: vec3<f32>, radius: f32) -> f32 {
  return gsl_sphere_remap(vec3<f32>(p.x, p.y + 12.0, p.z), radius);
}

fn gsl_sphere_remap_2(p: vec3<f32>) -> f32 {
  return dist(p) - 7.5;
}

fn gsl_sphere_remap_3(p: vec3<f32>) -> f32 {
  return dist(p) - 12.5;
}

fn gcl_closure_translate(p: vec3<f32>, displacement: vec3<f32>) -> f32 {
  return sphere(p - displacement, 2.5);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let s0 = sphere(vec3<f32>(p.x - 7.0, p.y, p.z), 2.5);
  let s1 = gsl_remap(vec3<f32>(p.x - 10.0, p.y, p.z));
  let s2 = gsl_remap_2(vec3<f32>(p.x - 15.0, p.y, p.z));
  let s3 = gsl_remap_3(vec3<f32>(p.x - 20.0, p.y, p.z));
  let s4 = sphere(vec3<f32>(p.x - 25.0, p.y, p.z), 20.5);
  let s5 = sphere(vec3<f32>(p.x - 35.0, p.y, p.z), 30.0);
  let s6 = gcl_closure_translate(p, vec3<f32>(p.x - 40.0, p.y, p.z));
  let gsl_s1 = gsl_shape(s0, 1u);
  let gsl_s2 = gsl_shape(s1, 2u);
  let gsl_s3 = gsl_shape(s2, 3u);
  let gsl_s4 = gsl_shape(s3, 4u);
  let gsl_s5 = gsl_shape(s4, 5u);
  let gsl_s6 = gsl_shape(s5, 6u);
  let gsl_s7 = gsl_shape(s6, 7u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7);
}
