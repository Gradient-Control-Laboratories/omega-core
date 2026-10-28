fn sdShape(p: vec3<f32>, time: f32, gsl_param_2_9: f32, gsl_param_6_17: f32, gsl_param_10_12: f32) -> gsl_shape {
  let box_size = gsl_param_10_12;
  let d_box = max(max(abs(p.x) - box_size, abs(p.y) - box_size), abs(p.z) - box_size);
  let r_sphere = sqrt(p.x * p.x + p.y * p.y + p.z * p.z);
  let sphere_radius = gsl_param_6_17;
  let d_sphere = r_sphere - sphere_radius;
  let blend = gsl_param_2_9;
  let h = clamp(0.5 + 0.5 * (d_box - d_sphere) / blend, 0.0, 1.0);
  let d_smooth = (1.0 - h) * d_box + h * d_sphere - blend * h * (1.0 - h);
  let d_cyl_x = sqrt(p.y * p.y + p.z * p.z) - 0.55;
  let d_cyl_y = sqrt(p.x * p.x + p.z * p.z) - 0.55;
  let d_cyl_z = sqrt(p.x * p.x + p.y * p.y) - 0.55;
  let d_cross = min(min(d_cyl_x, d_cyl_y), d_cyl_z);
  let gsl_s1 = gsl_shape(max(d_smooth, -d_cross), 1u);
  return gsl_s1;
}