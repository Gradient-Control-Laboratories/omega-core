fn gsl_floored_vec3(a: vec3<f32>, n: vec3<f32>) -> vec3<f32> {
  return a - n * floor(a / n);
}

fn gsl_floored_vec2(a: vec2<f32>, n: vec2<f32>) -> vec2<f32> {
  return a - n * floor(a / n);
}

fn gsl_floored_vec4(a: vec4<f32>, n: vec4<f32>) -> vec4<f32> {
  return a - n * floor(a / n);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let b1 = p.x;
  let c1 = gsl_floored_vec3(vec3<f32>(2.0, 3.0, 4.0), vec3<f32>(b1));
  let a2 = p.y;
  let c2 = gsl_floored_vec3(vec3<f32>(a2), vec3<f32>(5.0, 6.0, 7.0));
  let b3 = p.z;
  let c3 = gsl_floored_vec2(vec2<f32>(2.0, 3.0), vec2<f32>(b3));
  let a4 = p.x;
  let c4 = gsl_floored_vec4(vec4<f32>(a4), vec4<f32>(2.0, 3.0, 4.0, 5.0));
  let gsl_s1 = gsl_shape(c1, 1u);
  let gsl_s2 = gsl_shape(c2, 2u);
  let gsl_s3 = gsl_shape(c3, 3u);
  let gsl_s4 = gsl_shape(c4, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}
