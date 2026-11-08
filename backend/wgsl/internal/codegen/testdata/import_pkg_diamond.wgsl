fn pkg1_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  return max(max(d.x, d.y), d.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let s = pkg1_boxCentered(p, vec3<f32>(3.14159, 4.0, 5.0), vec3<f32>(6.0, 7.0, 8.0));
  let pkg2_partSize = vec3<f32>(1.0, 2.0, 5.14159);
  let t = pkg1_boxCentered(p, vec3<f32>(1.0, 2.0, 3.0), pkg2_partSize);
  let gsl_s1 = gsl_shape(s, 1u);
  let gsl_s2 = gsl_shape(t, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
