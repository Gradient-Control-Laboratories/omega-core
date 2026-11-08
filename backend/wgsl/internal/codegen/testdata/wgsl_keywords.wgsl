fn gsl_fn_sampler(p: vec3<f32>, gsl_var_impl: f32) -> vec3<f32> {
  return p + gsl_var_impl;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_var_subgroupInclusiveAdd = p + 1.0;
  let gsl_var_subgroupInclusiveMul = gsl_fn_sampler(p + 7.0, 1.0);
  let gsl_s1 = gsl_shape(gsl_var_subgroupInclusiveAdd, 1u);
  let gsl_s2 = gsl_shape(gsl_var_subgroupInclusiveMul, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
