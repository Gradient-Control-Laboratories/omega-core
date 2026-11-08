fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_intersectRound(d1: f32, d2: f32, radius: f32) -> f32 {
  let inset1 = max(d1 + radius, 0.0);
  let inset2 = max(d2 + radius, 0.0);
  return min(-radius, max(d1, d2)) + length(vec2<f32>(inset1, inset2));
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return builtin_intersectRound(abs(length(p.xy) - 1.29) - 0.075, abs(length(p.xy) * sin(gsl_floored((atan2(p.x, p.y) + 0.5235988), 1.0471976) - 0.5235988)) - 0.25, 0.0);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let side = gsl_remap(vec3<f32>(p.x - 4.0, p.y, p.z));
  return side;
}

fn builtin_unionSharp(fields: vec4<f32>) -> f32 {
  return min(min(min(fields.x, fields.y), fields.z), fields.w);
}

fn map_rotateXY(v: vec3<f32>, ang: f32) -> vec3<f32> {
  let ca = cos(ang);
  let sa = sin(ang);
  return vec3<f32>(v.x * ca + v.y * sa, -v.x * sa + v.y * ca, v.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let sides = vec4<f32>(gsl_remap_2(map_rotateXY(p, -1.0471976)), gsl_remap_2(map_rotateXY(p, -2.0943952)), gsl_remap_2(map_rotateXY(p, 1.0471976)), gsl_remap_2(map_rotateXY(p, 2.0943952)));
  let gsl_s1 = gsl_shape(builtin_unionSharp(sides), 1u);
  return gsl_s1;
}
