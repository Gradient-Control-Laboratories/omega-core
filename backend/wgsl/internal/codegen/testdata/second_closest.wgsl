fn builtin_intersectSharp(fields: array<f32, 6>) -> f32 {
  return max(max(max(max(max(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]);
}

fn builtin_unionSharp(d1: f32, d2: f32, d3: f32) -> f32 {
  return min(min(d1, d2), d3);
}

fn builtin_linearMap(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: vec3<f32>, outMax: vec3<f32>) -> vec3<f32> {
  let slope = (outMax - outMin) / (inMax - inMin);
  return slope * (gsl_var_value - inMin) + outMin;
}

fn builtin_ramp(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: vec3<f32>, outMax: vec3<f32>) -> vec3<f32> {
  let mapped = builtin_linearMap(gsl_var_value, inMin, inMax, outMin, outMax);
  let low = min(outMin, outMax);
  let high = max(outMin, outMax);
  return clamp(mapped, low, high);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let s = length(p) - 0.8;
  let r = builtin_intersectSharp(array<f32, 6>(0.3 - p.x, p.x - 1.2, 0.3 - p.y, p.y - 1.2, -1.2 - p.z, p.z - 1.2));
  let t = length(vec2<f32>(length(p.xy) - 1.3, p.z)) - 0.2;
  let combined = builtin_unionSharp(s, r, t);
  let gsl_s1 = gsl_shape(combined, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let s = length(p) - 0.8;
    let r = builtin_intersectSharp(array<f32, 6>(0.3 - p.x, p.x - 1.2, 0.3 - p.y, p.y - 1.2, -1.2 - p.z, p.z - 1.2));
    let t = length(vec2<f32>(length(p.xy) - 1.3, p.z)) - 0.2;
    let combined = builtin_unionSharp(s, r, t);
    let gsl_tmp_1 = 0.3 - p.x;
    let gsl_tmp_2 = p.x - 1.2;
    let gsl_tmp_3 = 0.3 - p.y;
    let gsl_tmp_4 = p.y - 1.2;
    let gsl_tmp_5 = -1.2 - p.z;
    let gsl_tmp_6 = max(max(max(max(gsl_tmp_1, gsl_tmp_2), gsl_tmp_3), gsl_tmp_4), gsl_tmp_5);
    let gsl_tmp_7 = p.z - 1.2;
    let edges = length(vec2<f32>(combined, select(min(s, r), min(t, select(min(s, select(gsl_tmp_6, max(gsl_tmp_7, select(max(max(max(gsl_tmp_1, gsl_tmp_2), gsl_tmp_3), gsl_tmp_4), max(gsl_tmp_5, select(max(max(gsl_tmp_1, gsl_tmp_2), gsl_tmp_3), max(gsl_tmp_4, select(max(gsl_tmp_1, gsl_tmp_2), max(gsl_tmp_3, min(gsl_tmp_1, gsl_tmp_2)), max(gsl_tmp_1, gsl_tmp_2) > gsl_tmp_3)), max(max(gsl_tmp_1, gsl_tmp_2), gsl_tmp_3) > gsl_tmp_4)), max(max(max(gsl_tmp_1, gsl_tmp_2), gsl_tmp_3), gsl_tmp_4) > gsl_tmp_5)), gsl_tmp_6 > gsl_tmp_7)), r, s < r)), min(s, r) < t)));
    let color_darkGray = vec3<f32>(0.25, 0.25, 0.25);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    return vec4<f32>(builtin_ramp(edges, 0.005, 0.01, color_darkGray, color_white), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
