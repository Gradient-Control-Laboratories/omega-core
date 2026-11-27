fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_intersectRound(a: f32, b: f32, radius: f32) -> f32 {
  let inset1 = max(a + radius, 0.0);
  let inset2 = max(b + radius, 0.0);
  return min(-radius, max(a, b)) + length(vec2<f32>(inset1, inset2));
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

fn domain_rotateXY(v: vec3<f32>, angle: f32) -> vec3<f32> {
  let ca = cos(angle);
  let sa = sin(angle);
  return vec3<f32>(v.x * ca + v.y * sa, -v.x * sa + v.y * ca, v.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let sides = vec4<f32>(gsl_remap_2(domain_rotateXY(p, -1.0471976)), gsl_remap_2(domain_rotateXY(p, -2.0943952)), gsl_remap_2(domain_rotateXY(p, 1.0471976)), gsl_remap_2(domain_rotateXY(p, 2.0943952)));
  let gsl_s1 = gsl_shape(builtin_unionSharp(sides), 1u);
  return gsl_s1;
}
