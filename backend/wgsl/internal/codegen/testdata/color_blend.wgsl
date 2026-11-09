fn builtin_intersectSharp(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectExponential(fields: vec2<f32>, width: f32) -> f32 {
  let peak = builtin_intersectSharp(fields);
  let scale = width / 2.7182817;
  let total = 0.0 + exp((fields.x - peak) / scale) + exp((fields.y - peak) / scale);
  return peak + scale * log(total);
}

fn builtin_unionExponential(fields: vec2<f32>, width: f32) -> f32 {
  let negated = vec2<f32>(-fields.x, -fields.y);
  return -builtin_intersectExponential(negated, width);
}

fn gsl_remap(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(1.0, 0.0, 0.0);
}

fn gsl_remap_2(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(0.0, 0.0, 1.0);
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let box = length(p - vec3<f32>(1.5, 0.0, 0.0)) - 1.0;
  return box;
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let ball = length(p) - 1.0;
  return ball;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let ball = length(p) - 1.0;
  let box = length(p - vec3<f32>(1.5, 0.0, 0.0)) - 1.0;
  let gsl_s1 = gsl_shape(builtin_unionExponential(vec2<f32>(ball, box), 0.5), 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let ball = length(p) - 1.0;
    let box = length(p - vec3<f32>(1.5, 0.0, 0.0)) - 1.0;
    return vec4<f32>(clamp((gsl_remap(p) * -(select(-0.0, -1.0, -ball > -box) + -0.0 * log(0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973)) + 0.18393973 * (exp((-ball - max(-ball, -box)) / 0.18393973) * ((-1.0 - select(-0.0, -1.0, -ball > -box)) * 0.18393973 - (-ball - max(-ball, -box)) * -0.0) / 0.03383382 + exp((-box - max(-ball, -box)) / 0.18393973) * ((-0.0 - select(-0.0, -1.0, -ball > -box)) * 0.18393973 - (-box - max(-ball, -box)) * -0.0) / 0.03383382) / (0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973))) + gsl_remap_2(p) * -(select(-1.0, -0.0, -ball > -box) + -0.0 * log(0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973)) + 0.18393973 * (exp((-ball - max(-ball, -box)) / 0.18393973) * ((-0.0 - select(-1.0, -0.0, -ball > -box)) * 0.18393973 - (-ball - max(-ball, -box)) * -0.0) / 0.03383382 + exp((-box - max(-ball, -box)) / 0.18393973) * ((-1.0 - select(-1.0, -0.0, -ball > -box)) * 0.18393973 - (-box - max(-ball, -box)) * -0.0) / 0.03383382) / (0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973))) + select(gsl_remap(p), gsl_remap_2(p), gsl_remap_3(p) < gsl_remap_4(p)) * max(1.0 - (-(select(-0.0, -1.0, -ball > -box) + -0.0 * log(0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973)) + 0.18393973 * (exp((-ball - max(-ball, -box)) / 0.18393973) * ((-1.0 - select(-0.0, -1.0, -ball > -box)) * 0.18393973 - (-ball - max(-ball, -box)) * -0.0) / 0.03383382 + exp((-box - max(-ball, -box)) / 0.18393973) * ((-0.0 - select(-0.0, -1.0, -ball > -box)) * 0.18393973 - (-box - max(-ball, -box)) * -0.0) / 0.03383382) / (0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973))) + -(select(-1.0, -0.0, -ball > -box) + -0.0 * log(0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973)) + 0.18393973 * (exp((-ball - max(-ball, -box)) / 0.18393973) * ((-0.0 - select(-1.0, -0.0, -ball > -box)) * 0.18393973 - (-ball - max(-ball, -box)) * -0.0) / 0.03383382 + exp((-box - max(-ball, -box)) / 0.18393973) * ((-1.0 - select(-1.0, -0.0, -ball > -box)) * 0.18393973 - (-box - max(-ball, -box)) * -0.0) / 0.03383382) / (0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973)))), 0.0)) / max(-(select(-0.0, -1.0, -ball > -box) + -0.0 * log(0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973)) + 0.18393973 * (exp((-ball - max(-ball, -box)) / 0.18393973) * ((-1.0 - select(-0.0, -1.0, -ball > -box)) * 0.18393973 - (-ball - max(-ball, -box)) * -0.0) / 0.03383382 + exp((-box - max(-ball, -box)) / 0.18393973) * ((-0.0 - select(-0.0, -1.0, -ball > -box)) * 0.18393973 - (-box - max(-ball, -box)) * -0.0) / 0.03383382) / (0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973))) + -(select(-1.0, -0.0, -ball > -box) + -0.0 * log(0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973)) + 0.18393973 * (exp((-ball - max(-ball, -box)) / 0.18393973) * ((-0.0 - select(-1.0, -0.0, -ball > -box)) * 0.18393973 - (-ball - max(-ball, -box)) * -0.0) / 0.03383382 + exp((-box - max(-ball, -box)) / 0.18393973) * ((-1.0 - select(-1.0, -0.0, -ball > -box)) * 0.18393973 - (-box - max(-ball, -box)) * -0.0) / 0.03383382) / (0.0 + exp((-ball - max(-ball, -box)) / 0.18393973) + exp((-box - max(-ball, -box)) / 0.18393973))), 1.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
