fn sdShape(p: vec3<f32>, time: f32, gsl_param_2_9: f32, gsl_param_6_13: f32, gsl_param_10_19: f32) -> gsl_shape {
  let scale = gsl_param_2_9;
  let sp = p * scale;
  let gyroid_val = sin(sp.x) * cos(sp.y) + sin(sp.y) * cos(sp.z) + sin(sp.z) * cos(sp.x);
  let thickness = gsl_param_6_13;
  let d_lattice = (abs(gyroid_val) - thickness) / scale;
  let r = sqrt(p.x * p.x + p.y * p.y + p.z * p.z);
  let boundary_radius = gsl_param_10_19;
  let d_sphere = r - boundary_radius;
  let gsl_s1 = gsl_shape(max(d_lattice, d_sphere), 1u);
  return gsl_s1;
}