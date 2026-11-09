fn matrix_outer(u: vec3<f32>, v: vec3<f32>) -> mat3x3<f32> {
  return mat3x3<f32>(u * v.x, u * v.y, u * v.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let q = mat3x3<f32>(vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) * p;
  let form = dot(p, mat3x3<f32>(vec3<f32>(5.0, 1.0, 2.0), vec3<f32>(1.0, 2.0, 1.0), vec3<f32>(2.0, 1.0, 2.0)) * p);
  let op = matrix_outer(p, q);
  let op0 = op[0];
  let ph = vec4<f32>(p.x, p.y, p.z, 1.0);
  let r4 = mat4x4<f32>(vec4<f32>(1.0, 0.000000000000000062969, 0.0, 0.0), vec4<f32>(-0.000000000000000050672674, 1.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(-0.0000000000000001110223, -0.00000000000000008326673, 0.0, 1.0)) * ph;
  let gsl_s1 = gsl_shape(length(q) + form + op0.x + r4.x + 3.5296843 - 9.0, 1u);
  return gsl_s1;
}
