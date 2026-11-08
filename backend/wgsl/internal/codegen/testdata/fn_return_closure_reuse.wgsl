fn gcl_closure_scaleBy(a: f32) -> f32 {
  return a * 0.5;
}

fn gcl_closure_apply(x: f32) -> f32 {
  return gcl_closure_scaleBy(gcl_closure_scaleBy(x));
}

fn gcl_closure_scaleBy_2(a: f32) -> f32 {
  return a * 3.0;
}

fn gcl_closure_apply_2(x: f32) -> f32 {
  return gcl_closure_scaleBy_2(gcl_closure_scaleBy_2(x));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(gcl_closure_apply(p.x) + gcl_closure_apply_2(p.y), 1u);
  return gsl_s1;
}
