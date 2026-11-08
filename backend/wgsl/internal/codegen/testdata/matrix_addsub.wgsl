fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let A = mat3x3<f32>(vec3<f32>(p.x, 0.0, 0.0), vec3<f32>(0.0, p.y, 0.0), vec3<f32>(0.0, 0.0, p.z));
  let S = A + mat3x3<f32>(vec3<f32>(1.0, 2.0, 3.0), vec3<f32>(4.0, 5.0, 6.0), vec3<f32>(7.0, 8.0, 9.0));
  let s0 = S[0];
  let D = A - mat3x3<f32>(vec3<f32>(1.0, 2.0, 3.0), vec3<f32>(4.0, 5.0, 6.0), vec3<f32>(7.0, 8.0, 9.0));
  let d0 = D[0];
  let N = -A;
  let n0 = N[0];
  let gsl_s1 = gsl_shape(s0.x + d0.x + n0.x - 1.0, 1u);
  return gsl_s1;
}
