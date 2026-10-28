fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = abs(p - vec3<f32>(1.0, 2.0, 3.0)) - vec3<f32>(0.5, 1.0, 2.570795);
  let t = max(max(gsl_tmp_1.x, gsl_tmp_1.y), gsl_tmp_1.z);
  let gsl_s1 = gsl_shape(t, 1u);
  return gsl_s1;
}
