fn builtin_unionSharp(fields: vec3<f32>) -> f32 {
  return min(min(fields.x, fields.y), fields.z);
}

fn gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(p: vec3<f32>) -> f32 {
  return length(p.xy) - 0.2;
}

fn gcl_closure_remapEdges(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(p: vec3<f32>) -> f32 {
  return max(abs(p.x), abs(p.y)) - 0.2;
}

fn gcl_closure_remapEdges_2(p: vec3<f32>, fields: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec3<f32>(builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.y, 0.0)), 1.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.x, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.y, fields.z, 0.0)), 1.0 > 0.0))), builtin_unionSharp(vec3<f32>(select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.x, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.y, 0.0)), 0.0 > 0.0), select(100000.0, gsl_gcl_closure_remapEdges_2_gsl_anon_gsl_anon_remap(vec3<f32>(fields.z, fields.z, 0.0)), 0.0 > 0.0)))));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let circle = gcl_closure_remapEdges(p, vec3<f32>(length(p) - 1.0, length(p) - 2.0, length(p) - 3.0));
  let diamond = gcl_closure_remapEdges_2(p, vec3<f32>(length(p) - 1.0, length(p) - 2.0, length(p) - 3.0));
  let gsl_s1 = gsl_shape(min(circle, diamond), 1u);
  return gsl_s1;
}
