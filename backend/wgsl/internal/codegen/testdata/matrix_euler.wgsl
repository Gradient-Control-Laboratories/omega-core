fn vector_multiplyVector(m: mat4x4<f32>, dir: vec3<f32>) -> vec3<f32> {
  let h = m * vec4<f32>(dir.x, dir.y, dir.z, 0.0);
  return vec3<f32>(h.x, h.y, h.z);
}

fn domain_rotateYZ(v: vec3<f32>, angle: f32) -> vec3<f32> {
  let ca = cos(angle);
  let sa = sin(angle);
  return vec3<f32>(v.x, v.y * ca + v.z * sa, -v.y * sa + v.z * ca);
}

fn domain_rotateZX(v: vec3<f32>, angle: f32) -> vec3<f32> {
  let ca = cos(angle);
  let sa = sin(angle);
  return vec3<f32>(v.x * ca - v.z * sa, v.y, v.z * ca + v.x * sa);
}

fn domain_rotateXY(v: vec3<f32>, angle: f32) -> vec3<f32> {
  let ca = cos(angle);
  let sa = sin(angle);
  return vec3<f32>(v.x * ca + v.y * sa, -v.x * sa + v.y * ca, v.z);
}

fn domain_eulerAngles(v: vec3<f32>, angles: vec3<f32>) -> vec3<f32> {
  return domain_rotateXY(domain_rotateZX(domain_rotateYZ(v, angles.x), angles.y), angles.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let viaMatrix = vector_multiplyVector(mat4x4<f32>(vec4<f32>(0.34692946, -0.681633, -0.64421767, 0.0), vec4<f32>(0.76504755, 0.6030044, -0.22602633, 0.0), vec4<f32>(0.5425331, -0.414442, 0.73068166, 0.0), vec4<f32>(0.0, 0.0, 0.0, 1.0)), p);
  let viaEuler = domain_eulerAngles(p, vec3<f32>(0.3, -0.7, 1.1));
  let diff = viaMatrix - viaEuler;
  let gsl_s1 = gsl_shape(1e30 + length(diff), 1u);
  return gsl_s1;
}
