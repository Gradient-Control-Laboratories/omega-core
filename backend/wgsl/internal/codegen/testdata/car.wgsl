fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_2 = abs(p) - vec3<f32>(1.0, 0.81, 0.5);
  return length(max(gsl_tmp_2, vec3<f32>(0.0))) + min(max(gsl_tmp_2.x, max(gsl_tmp_2.y, gsl_tmp_2.z)), 0.0);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let gsl_tmp_3 = abs(vec2<f32>(length(p.xz), p.y)) - vec2<f32>(0.6, 0.3);
  return min(max(gsl_tmp_3.x, gsl_tmp_3.y), 0.0) + length(max(gsl_tmp_3, vec2<f32>(0.0)));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let gsl_tmp_4 = abs(vec2<f32>(length(p.xz), p.y)) - vec2<f32>(0.6, 0.3);
  return min(max(gsl_tmp_4.x, gsl_tmp_4.y), 0.0) + length(max(gsl_tmp_4, vec2<f32>(0.0)));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let gsl_tmp_5 = abs(vec2<f32>(length(p.xz), p.y)) - vec2<f32>(0.6, 0.3);
  return min(max(gsl_tmp_5.x, gsl_tmp_5.y), 0.0) + length(max(gsl_tmp_5, vec2<f32>(0.0)));
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let gsl_tmp_6 = abs(vec2<f32>(length(p.xz), p.y)) - vec2<f32>(0.6, 0.3);
  return min(max(gsl_tmp_6.x, gsl_tmp_6.y), 0.0) + length(max(gsl_tmp_6, vec2<f32>(0.0)));
}

fn sdRoundedBox(p: vec3<f32>, size: vec3<f32>, radius: f32) -> f32 {
  return length(max(abs(p) - size, vec3<f32>(0.0))) - radius;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let body = length(max(abs(p) - vec3<f32>(2.0, 0.9, 0.5), vec3<f32>(0.0))) - 0.2;
  let cabin = sdRoundedBox(p - vec3<f32>(-0.3, 0.0, 0.8), vec3<f32>(1.2, 0.8, 0.6), 0.2);
  let carChassis = min(body, cabin);
  let windowCutout = gsl_remap(p - vec3<f32>(-0.3, 0.0, 0.9));
  let carWithWindows = max(carChassis, -windowCutout);
  let wheel_fr = gsl_remap_2(p - vec3<f32>(1.5, 0.9, -0.1));
  let wheel_fl = gsl_remap_3(p - vec3<f32>(1.5, -0.9, -0.1));
  let wheel_rr = gsl_remap_4(p - vec3<f32>(-1.5, 0.9, -0.1));
  let wheel_rl = gsl_remap_5(p - vec3<f32>(-1.5, -0.9, -0.1));
  let allWheels = min(min(min(wheel_fr, wheel_fl), wheel_rr), wheel_rl);
  let finalCar = min(carWithWindows, allWheels);
  let gsl_s1 = gsl_shape(finalCar, 1u);
  return gsl_s1;
}
