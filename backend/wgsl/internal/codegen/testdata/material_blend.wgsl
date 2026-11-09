fn builtin_sphere(p: vec3<f32>, radius: f32) -> f32 {
  return length(p) - radius;
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let ball = length(p) - 1.0;
  let box = builtin_sphere(p - vec3<f32>(3.0, 0.0, 0.0), 1.0);
  let plain = builtin_sphere(p - vec3<f32>(0.0, 3.0, 0.0), 1.0);
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(ball, box)), 1u);
  let gsl_s2 = gsl_shape(plain, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_brushedMetal; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
