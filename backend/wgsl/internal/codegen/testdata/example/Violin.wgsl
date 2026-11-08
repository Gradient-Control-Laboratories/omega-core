fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_41 = abs(p) - vec3<f32>(2.5, 0.3, 0.3);
  return length(max(gsl_tmp_41, vec3<f32>(0.0, 0.0, 0.0))) + min(max(gsl_tmp_41.x, max(gsl_tmp_41.y, gsl_tmp_41.z)), 0.0);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let gsl_tmp_44 = abs(p) - vec3<f32>(2.6, 0.4, 0.1);
  return length(max(gsl_tmp_44, vec3<f32>(0.0, 0.0, 0.0))) + min(max(gsl_tmp_44.x, max(gsl_tmp_44.y, gsl_tmp_44.z)), 0.0);
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let gsl_tmp_46 = abs(p) - vec3<f32>(0.5, 0.35, 0.6);
  return length(max(gsl_tmp_46, vec3<f32>(0.0, 0.0, 0.0))) + min(max(gsl_tmp_46.x, max(gsl_tmp_46.y, gsl_tmp_46.z)), 0.0);
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return length(vec2<f32>(length(p.xy) - 0.5, p.z)) - 0.15;
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let gsl_tmp_50 = abs(p) - vec3<f32>(0.4, 0.8, 0.1);
  return length(max(gsl_tmp_50, vec3<f32>(0.0, 0.0, 0.0))) + min(max(gsl_tmp_50.x, max(gsl_tmp_50.y, gsl_tmp_50.z)), 0.0);
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  let gsl_tmp_52 = abs(p) - vec3<f32>(0.06, 0.4, 0.3);
  return length(max(gsl_tmp_52, vec3<f32>(0.0, 0.0, 0.0))) + min(max(gsl_tmp_52.x, max(gsl_tmp_52.y, gsl_tmp_52.z)), 0.0);
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  let gsl_tmp_51 = abs(p) - vec3<f32>(0.05, 1.0, 0.5);
  return max(length(max(gsl_tmp_51, vec3<f32>(0.0, 0.0, 0.0))) + min(max(gsl_tmp_51.x, max(gsl_tmp_51.y, gsl_tmp_51.z)), 0.0), -gsl_remap_6(p - vec3<f32>(0.0, 0.0, -0.1)));
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  let gsl_tmp_35 = vec2<f32>(p.x - 1.0, abs(p.y) - 0.9);
  let gsl_tmp_36 = abs(gsl_tmp_35) - vec2<f32>(0.07, 0.7);
  let gsl_tmp_40 = p - vec3<f32>(-5.0, 0.0, 0.2);
  let gsl_tmp_37 = clamp((-p.x - 2.5) / 5.0, 0.0, 1.0);
  let gsl_tmp_38 = 1.0 - 0.2 * gsl_tmp_37;
  let gsl_tmp_39 = 1.0 + 0.5 * gsl_tmp_37;
  let gsl_tmp_43 = p - vec3<f32>(-4.9, 0.0, 0.75);
  let gsl_tmp_42 = 1.0 - 0.3 * clamp((-p.x - 2.3) / 5.2, 0.0, 1.0);
  let gsl_tmp_45 = p - vec3<f32>(-8.0, 0.0, 0.5);
  let gsl_tmp_47 = p - vec3<f32>(-9.0, 0.0, 0.5);
  let gsl_tmp_48 = p - vec3<f32>(3.5, 0.0, 0.9);
  let gsl_tmp_49 = 1.0 - 0.5 * clamp((gsl_tmp_48.x + 0.4) / 0.8, 0.0, 1.0);
  return min(smin(max(max(max(smin(distance(p.xy, vec2<f32>(1.75, 0.0)) - 2.1, distance(p.xy, vec2<f32>(-1.25, 0.0)) - 1.7, 0.7), -(distance(vec2<f32>(p.x, abs(p.y)), vec2<f32>(0.2, 2.2)) - 1.1)), abs(p.z) - (0.6 + 0.3 * (1.0 - clamp(pow(p.x / 4.0, 2.0) + pow(p.y / 2.1, 2.0), 0.0, 1.0)))), -max(min(min(distance(gsl_tmp_35, vec2<f32>(0.0, 0.6)) - 0.2, distance(gsl_tmp_35, vec2<f32>(0.0, -0.6)) - 0.2), length(max(gsl_tmp_36, vec2<f32>(0.0, 0.0))) + min(max(gsl_tmp_36.x, gsl_tmp_36.y), 0.0)), abs(p.z - 0.7) - 0.2)), min(min(min(gsl_remap(vec3<f32>(gsl_tmp_40.x, gsl_tmp_40.y / gsl_tmp_38, gsl_tmp_40.z / gsl_tmp_39)) * min(gsl_tmp_38, gsl_tmp_39), gsl_remap_2(vec3<f32>(gsl_tmp_43.x, gsl_tmp_43.y / gsl_tmp_42, gsl_tmp_43.z)) * gsl_tmp_42), gsl_remap_3(vec3<f32>(gsl_tmp_45.x * 0.9800666 - gsl_tmp_45.z * 0.19866933, gsl_tmp_45.y, gsl_tmp_45.x * 0.19866933 + gsl_tmp_45.z * 0.9800666))), gsl_remap_4(vec3<f32>(gsl_tmp_47.z, gsl_tmp_47.y, gsl_tmp_47.x))), 0.3), min(gsl_remap_5(vec3<f32>(gsl_tmp_48.x, gsl_tmp_48.y / gsl_tmp_49, gsl_tmp_48.z)) * gsl_tmp_49, gsl_remap_7(p - vec3<f32>(1.0, 0.0, 0.95))));
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  return min(min(min(min(capsule(p, vec3<f32>(3.3, -0.25, 0.9), vec3<f32>(1.0, -0.3, 1.45), 0.03), capsule(p, vec3<f32>(1.0, -0.3, 1.45), vec3<f32>(-7.5, -0.15, 0.85), 0.03)), min(capsule(p, vec3<f32>(3.3, -0.08, 0.9), vec3<f32>(1.0, -0.1, 1.45), 0.025), capsule(p, vec3<f32>(1.0, -0.1, 1.45), vec3<f32>(-7.5, -0.05, 0.85), 0.025))), min(capsule(p, vec3<f32>(3.3, 0.08, 0.9), vec3<f32>(1.0, 0.1, 1.45), 0.02), capsule(p, vec3<f32>(1.0, 0.1, 1.45), vec3<f32>(-7.5, 0.05, 0.85), 0.02))), min(capsule(p, vec3<f32>(3.3, 0.25, 0.9), vec3<f32>(1.0, 0.3, 1.45), 0.015), capsule(p, vec3<f32>(1.0, 0.3, 1.45), vec3<f32>(-7.5, 0.15, 0.85), 0.015)));
}

fn smin(a: f32, b: f32, k: f32) -> f32 {
  let h = max(k - abs(a - b), 0.0);
  return min(a, b) - pow(h, 2.0) / (4.0 * k);
}

fn capsule(p: vec3<f32>, a: vec3<f32>, b: vec3<f32>, r: f32) -> f32 {
  let pa = p - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h) - r;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let p_transformed = vec3<f32>((p.x - 1.6928838) / 0.5992509, p.y / 0.5992509, p.z / 0.5992509);
  let gsl_s1 = gsl_shape(gsl_remap_8(p_transformed), 1u);
  let gsl_s2 = gsl_shape(gsl_remap_9(p_transformed), 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
