fn gsl_remap(p: vec3<f32>) -> f32 {
  return length(max(abs(p) - vec3<f32>(0.9, 4.5, 0.1), vec3<f32>(0.0))) + min(max(max((abs(p) - vec3<f32>(0.9, 4.5, 0.1)).x, (abs(p) - vec3<f32>(0.9, 4.5, 0.1)).y), (abs(p) - vec3<f32>(0.9, 4.5, 0.1)).z), 0.0);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return length(max(abs(p) - vec3<f32>(0.75, 4.0, 0.1), vec3<f32>(0.0))) + min(max(max((abs(p) - vec3<f32>(0.75, 4.0, 0.1)).x, (abs(p) - vec3<f32>(0.75, 4.0, 0.1)).y), (abs(p) - vec3<f32>(0.75, 4.0, 0.1)).z), 0.0);
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return min(max((abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.08, 0.75)).x, (abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.08, 0.75)).y), 0.0) + length(max(abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.08, 0.75), vec2<f32>(0.0)));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return length(max(abs(p) - vec3<f32>(0.6, 0.075, 0.9), vec3<f32>(0.0))) + min(max(max((abs(p) - vec3<f32>(0.6, 0.075, 0.9)).x, (abs(p) - vec3<f32>(0.6, 0.075, 0.9)).y), (abs(p) - vec3<f32>(0.6, 0.075, 0.9)).z), 0.0);
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  return length(max(abs(p) - vec3<f32>(0.75, 1.75, 0.075), vec3<f32>(0.0))) + min(max(max((abs(p) - vec3<f32>(0.75, 1.75, 0.075)).x, (abs(p) - vec3<f32>(0.75, 1.75, 0.075)).y), (abs(p) - vec3<f32>(0.75, 1.75, 0.075)).z), 0.0);
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  return min(length(max(abs(p) - vec3<f32>(0.05, 1.5, 0.15), vec3<f32>(0.0))) + min(max(max((abs(p) - vec3<f32>(0.05, 1.5, 0.15)).x, (abs(p) - vec3<f32>(0.05, 1.5, 0.15)).y), (abs(p) - vec3<f32>(0.05, 1.5, 0.15)).z), 0.0), length(max(abs(p) - vec3<f32>(0.05, 0.15, 1.5), vec3<f32>(0.0))) + min(max(max((abs(p) - vec3<f32>(0.05, 0.15, 1.5)).x, (abs(p) - vec3<f32>(0.05, 0.15, 1.5)).y), (abs(p) - vec3<f32>(0.05, 0.15, 1.5)).z), 0.0));
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  return min(max((abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.1, 0.35)).x, (abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.1, 0.35)).y), 0.0) + length(max(abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.1, 0.35), vec2<f32>(0.0)));
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  return min(max((abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.05, 0.4)).x, (abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.05, 0.4)).y), 0.0) + length(max(abs(vec2<f32>(length(p.xy), p.z)) - vec2<f32>(0.05, 0.4), vec2<f32>(0.0)));
}

fn sdSphere(p: vec3<f32>, r: f32) -> f32 {
  return length(p) - r;
}

fn opSubtract(a: f32, b: f32) -> f32 {
  return max(a, -b);
}

fn sdTorusY(p: vec3<f32>, R: f32, r: f32) -> f32 {
  let q = vec2<f32>(length(p.xz) - R, p.y);
  return length(q) - r;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let fuselage_body = length(max(abs(p) - vec3<f32>(3.5, 0.6, 0.5), vec3<f32>(0.0))) + min(max(max((abs(p) - vec3<f32>(3.5, 0.6, 0.5)).x, (abs(p) - vec3<f32>(3.5, 0.6, 0.5)).y), (abs(p) - vec3<f32>(3.5, 0.6, 0.5)).z), 0.0);
  let cockpit_hole = sdSphere(p - vec3<f32>(-0.5, 0.0, 0.3), 0.5);
  let fuselage = opSubtract(fuselage_body, cockpit_hole);
  let wing_upper = gsl_remap(p - vec3<f32>(0.0, 0.0, 1.0));
  let wing_lower = gsl_remap_2(p - vec3<f32>(0.0, 0.0, -0.5));
  let strut1 = gsl_remap_3(p - vec3<f32>(0.3, 3.0, 0.25));
  let strut2 = gsl_remap_3(p - vec3<f32>(0.3, -3.0, 0.25));
  let strut3 = gsl_remap_3(p - vec3<f32>(-0.3, 3.0, 0.25));
  let strut4 = gsl_remap_3(p - vec3<f32>(-0.3, -3.0, 0.25));
  let tail_v = gsl_remap_4(p - vec3<f32>(3.0, 0.0, 0.6));
  let tail_h = gsl_remap_5(p - vec3<f32>(3.0, 0.0, -0.1));
  let prop_hub = sdSphere(p - vec3<f32>(-3.6, 0.0, 0.0), 0.2);
  let prop_blades = gsl_remap_6(p - vec3<f32>(-3.6, 0.0, 0.0));
  let wheel_left = sdTorusY(p - vec3<f32>(0.0, -1.0, -1.2), 0.4, 0.1);
  let wheel_right = sdTorusY(p - vec3<f32>(0.0, 1.0, -1.2), 0.4, 0.1);
  let gear_strut_left = gsl_remap_7(p - vec3<f32>(0.0, -1.0, -0.85));
  let gear_strut_right = gsl_remap_7(p - vec3<f32>(0.0, 1.0, -0.85));
  let tail_skid = gsl_remap_8(p - vec3<f32>(3.2, 0.0, -0.9));
  let gsl_s1 = gsl_shape(fuselage, 1u);
  let gsl_s2 = gsl_shape(wing_upper, 2u);
  let gsl_s3 = gsl_shape(wing_lower, 3u);
  let gsl_s4 = gsl_shape(strut1, 4u);
  let gsl_s5 = gsl_shape(strut2, 5u);
  let gsl_s6 = gsl_shape(strut3, 6u);
  let gsl_s7 = gsl_shape(strut4, 7u);
  let gsl_s8 = gsl_shape(tail_v, 8u);
  let gsl_s9 = gsl_shape(tail_h, 9u);
  let gsl_s10 = gsl_shape(prop_hub, 10u);
  let gsl_s11 = gsl_shape(prop_blades, 11u);
  let gsl_s12 = gsl_shape(wheel_left, 12u);
  let gsl_s13 = gsl_shape(wheel_right, 13u);
  let gsl_s14 = gsl_shape(gear_strut_left, 14u);
  let gsl_s15 = gsl_shape(gear_strut_right, 15u);
  let gsl_s16 = gsl_shape(tail_skid, 16u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16);
}
