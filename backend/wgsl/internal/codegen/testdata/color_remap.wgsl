fn gsl_remap(p: vec3<f32>) -> f32 {
  let ball = length(p) - 1.0;
  return ball;
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn gsl_remap_2(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(0.0, 0.0, 1.0);
}

fn gsl_remap_3(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(1.0, 0.0, 0.0);
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let moved = gsl_remap(p - vec3<f32>(2.0, 0.0, 0.0));
  return moved;
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let box = length(p) - 1.0;
  return box;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let moved = gsl_remap(p - vec3<f32>(2.0, 0.0, 0.0));
  let box = length(p) - 1.0;
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(moved, box)), 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let moved = gsl_remap(p - vec3<f32>(2.0, 0.0, 0.0));
    let box = length(p) - 1.0;
    return vec4<f32>(clamp((gsl_remap_2(p) * select(1.0, 0.0, moved < box) + clamp(gsl_remap_3(p - vec3<f32>(2.0, 0.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)) * select(0.0, 1.0, moved < box) + select(gsl_remap_2(p), clamp(gsl_remap_3(p - vec3<f32>(2.0, 0.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_remap_4(p) < gsl_remap_5(p)) * max(1.0 - (select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box)), 0.0)) / max(select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box), 1.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
