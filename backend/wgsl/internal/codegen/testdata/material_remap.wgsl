fn gsl_remap(p: vec3<f32>) -> f32 {
  let ball = length(p) - 1.0;
  return ball;
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn gsl_remap_2(p: vec3<f32>, material_matte: mat4x4<f32>) -> vec4<f32> {
  return material_matte[0];
}

fn gsl_remap_3(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> vec4<f32> {
  return material_chrome[0];
}

fn gsl_remap_4(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> vec4<f32> {
  return material_chrome[1];
}

fn gsl_remap_5(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> vec4<f32> {
  return material_chrome[2];
}

fn gsl_remap_6(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> f32 {
  let moved = gsl_remap(p - vec3<f32>(2.0, 0.0, 0.0));
  return moved;
}

fn gsl_remap_7(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> f32 {
  let box = length(p) - 1.0;
  return box;
}

fn gsl_remap_8(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> vec4<f32> {
  return material_matte[0];
}

fn gsl_remap_9(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> vec4<f32> {
  return material_matte[1];
}

fn gsl_remap_10(p: vec3<f32>, material_matte: mat4x4<f32>, material_chrome: mat4x4<f32>) -> vec4<f32> {
  return material_matte[2];
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let moved = gsl_remap(p - vec3<f32>(2.0, 0.0, 0.0));
  let box = length(p) - 1.0;
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(moved, box)), 1u);
  return gsl_s1;
}

fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let moved = gsl_remap(p - vec3<f32>(2.0, 0.0, 0.0));
    let box = length(p) - 1.0;
    let material_chrome = mat4x4<f32>(vec4<f32>(0.04, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = mat4x4<f32>(clamp((gsl_remap_2(p, material_matte) * select(1.0, 0.0, moved < box) + clamp(gsl_remap_3(p - vec3<f32>(2.0, 0.0, 0.0), material_matte, material_chrome) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)) * select(0.0, 1.0, moved < box) + select(gsl_remap_8(p, material_matte, material_chrome), clamp(gsl_remap_3(p - vec3<f32>(2.0, 0.0, 0.0), material_matte, material_chrome) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), gsl_remap_6(p, material_matte, material_chrome) < gsl_remap_7(p, material_matte, material_chrome)) * max(1.0 - (select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box)), 0.0)) / max(select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box), 1.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp((gsl_remap_9(p, material_matte, material_chrome) * select(1.0, 0.0, moved < box) + clamp(gsl_remap_4(p - vec3<f32>(2.0, 0.0, 0.0), material_matte, material_chrome) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)) * select(0.0, 1.0, moved < box) + select(gsl_remap_9(p, material_matte, material_chrome), clamp(gsl_remap_4(p - vec3<f32>(2.0, 0.0, 0.0), material_matte, material_chrome) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), gsl_remap_6(p, material_matte, material_chrome) < gsl_remap_7(p, material_matte, material_chrome)) * max(1.0 - (select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box)), 0.0)) / max(select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box), 1.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp((gsl_remap_10(p, material_matte, material_chrome) * select(1.0, 0.0, moved < box) + clamp(gsl_remap_5(p - vec3<f32>(2.0, 0.0, 0.0), material_matte, material_chrome) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)) * select(0.0, 1.0, moved < box) + select(gsl_remap_10(p, material_matte, material_chrome), clamp(gsl_remap_5(p - vec3<f32>(2.0, 0.0, 0.0), material_matte, material_chrome) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), gsl_remap_6(p, material_matte, material_chrome) < gsl_remap_7(p, material_matte, material_chrome)) * max(1.0 - (select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box)), 0.0)) / max(select(1.0, 0.0, moved < box) + select(0.0, 1.0, moved < box), 1.0), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
