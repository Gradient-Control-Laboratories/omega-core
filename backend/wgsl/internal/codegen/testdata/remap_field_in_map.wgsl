fn gsl_remap(p: vec3<f32>) -> f32 {
  let side = max(abs(p.x) - 2.0, abs(p.y) - 0.5);
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
  let sides = vec4<f32>(gsl_remap(map_rotateXY(p, -1.0471976)), gsl_remap(map_rotateXY(p, -2.0943952)), gsl_remap(map_rotateXY(p, 1.0471976)), gsl_remap(map_rotateXY(p, 2.0943952)));
  let gsl_s1 = gsl_shape(builtin_unionSharp(sides), 1u);
  return gsl_s1;
}
