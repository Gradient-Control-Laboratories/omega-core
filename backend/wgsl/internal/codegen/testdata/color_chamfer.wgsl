fn builtin_intersectChamfer(d1: f32, d2: f32, setback: f32) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2 + setback) / 1.4142135;
  return max(face, edge);
}

fn builtin_unionChamfer(d1: f32, d2: f32, setback: f32) -> f32 {
  return -builtin_intersectChamfer(-d1, -d2, setback);
}

fn gsl_remap(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(1.0, 0.0, 0.0);
}

fn gsl_remap_2(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(0.0, 0.0, 1.0);
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let box = length(p - vec3<f32>(1.4, 0.0, 0.0)) - 1.0;
  return box;
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let ball = length(p) - 1.0;
  return ball;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let ball = length(p) - 1.0;
  let box = length(p - vec3<f32>(1.4, 0.0, 0.0)) - 1.0;
  let gsl_s1 = gsl_shape(builtin_unionChamfer(ball, box, 0.5), 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let ball = length(p) - 1.0;
    let box = length(p - vec3<f32>(1.4, 0.0, 0.0)) - 1.0;
    return vec4<f32>(clamp((gsl_remap(p) * -select((-1.4142135 - (-ball + -box + 0.5) * 0.0) / 2.0, select(-0.0, -1.0, -ball > -box), max(-ball, -box) > (-ball + -box + 0.5) / 1.4142135) + gsl_remap_2(p) * -select((-1.4142135 - (-ball + -box + 0.5) * 0.0) / 2.0, select(-1.0, -0.0, -ball > -box), max(-ball, -box) > (-ball + -box + 0.5) / 1.4142135) + select(gsl_remap(p), gsl_remap_2(p), gsl_remap_3(p) < gsl_remap_4(p)) * max(1.0 - (-select((-1.4142135 - (-ball + -box + 0.5) * 0.0) / 2.0, select(-0.0, -1.0, -ball > -box), max(-ball, -box) > (-ball + -box + 0.5) / 1.4142135) + -select((-1.4142135 - (-ball + -box + 0.5) * 0.0) / 2.0, select(-1.0, -0.0, -ball > -box), max(-ball, -box) > (-ball + -box + 0.5) / 1.4142135)), 0.0)) / max(-select((-1.4142135 - (-ball + -box + 0.5) * 0.0) / 2.0, select(-0.0, -1.0, -ball > -box), max(-ball, -box) > (-ball + -box + 0.5) / 1.4142135) + -select((-1.4142135 - (-ball + -box + 0.5) * 0.0) / 2.0, select(-1.0, -0.0, -ball > -box), max(-ball, -box) > (-ball + -box + 0.5) / 1.4142135), 1.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
