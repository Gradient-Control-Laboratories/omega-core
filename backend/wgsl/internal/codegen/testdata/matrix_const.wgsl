fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let myVec = vec4<f32>(p.x, p.y, p.z, time);
  let rows = mat4x4<f32>(vec4<f32>(1.0, 0.0, 0.0, 0.0) + myVec, vec4<f32>(0.0, 1.0, 0.0, 0.0) + myVec, vec4<f32>(0.0, 0.0, 1.0, 0.0) + myVec, vec4<f32>(10.0, 20.0, 30.0, 1.0) + myVec);
  let row0 = rows[0];
  let row1 = rows[1];
  let gsl_s1 = gsl_shape(p.x + row0.x + row1.y, 1u);
  return gsl_s1;
}
