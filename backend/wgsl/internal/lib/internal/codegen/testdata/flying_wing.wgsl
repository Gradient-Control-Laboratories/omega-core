fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_3 = abs(p) - vec3<f32>(4.5, 2.0, 0.2);
  return length(max(gsl_tmp_3, vec3<f32>(0.0, 0.0, 0.0))) + min(max(gsl_tmp_3.x, max(gsl_tmp_3.y, gsl_tmp_3.z)), 0.0);
}

fn propeller(p: vec3<f32>, majorRadius: f32, minorRadius: f32) -> f32 {
  let q = vec2<f32>(length(p.xz) - majorRadius, p.y);
  return length(q) - minorRadius;
}

fn gsl_landing_gear_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_2 = vec2<f32>(length(p.yz) - 0.4, abs(p.x) - 0.1);
  return min(max(gsl_tmp_2.x, gsl_tmp_2.y), 0.0) + length(max(gsl_tmp_2, vec2<f32>(0.0)));
}

fn landing_gear(p: vec3<f32>) -> gsl_struct {
  let gsl_tmp_1 = vec2<f32>(length(p.xy) - 0.1, abs(p.z) - 0.6);
  let strut = min(max(gsl_tmp_1.x, gsl_tmp_1.y), 0.0) + length(max(gsl_tmp_1, vec2<f32>(0.0)));
  let wheel = gsl_landing_gear_remap(p - vec3<f32>(0.0, 0.0, -0.6));
  var gsl_result : gsl_struct;
  gsl_result.v_1 = strut;
  gsl_result.v_2 = wheel;
  return gsl_result;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let wing = gsl_remap(p - vec3<f32>(0.0, 0.0, 1.2));
  let prop1 = propeller(p - vec3<f32>(1.5, 2.0, 1.45), 0.7, 0.05);
  let prop2 = propeller(p - vec3<f32>(-1.5, 2.0, 1.45), 0.7, 0.05);
  let prop3 = propeller(p - vec3<f32>(3.5, 2.0, 1.45), 0.7, 0.05);
  let prop4 = propeller(p - vec3<f32>(-3.5, 2.0, 1.45), 0.7, 0.05);
  let gsl_result = landing_gear(p - vec3<f32>(0.0, 1.5, 0.0));
  let front_strut = gsl_result.v_1;
  let front_gear = gsl_result.v_2;
  let gsl_result_2 = landing_gear(p - vec3<f32>(-3.0, -1.0, 0.0));
  let rear_strut_L = gsl_result_2.v_1;
  let rear_gear_L = gsl_result_2.v_2;
  let gsl_result_3 = landing_gear(p - vec3<f32>(3.0, -1.0, 0.0));
  let rear_strut_R = gsl_result_3.v_1;
  let rear_gear_R = gsl_result_3.v_2;
  let gsl_s1 = gsl_shape(wing, 1u);
  let gsl_s2 = gsl_shape(prop1, 2u);
  let gsl_s3 = gsl_shape(prop2, 3u);
  let gsl_s4 = gsl_shape(prop3, 4u);
  let gsl_s5 = gsl_shape(prop4, 5u);
  let gsl_s6 = gsl_shape(front_strut, 6u);
  let gsl_s7 = gsl_shape(front_gear, 7u);
  let gsl_s8 = gsl_shape(rear_strut_L, 8u);
  let gsl_s9 = gsl_shape(rear_gear_L, 9u);
  let gsl_s10 = gsl_shape(rear_strut_R, 10u);
  let gsl_s11 = gsl_shape(rear_gear_R, 11u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
}
