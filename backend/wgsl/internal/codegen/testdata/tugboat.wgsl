fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, -1.0);
  return min(sdBox(p - vec3<f32>(0.0, 0.0, -0.2), vec3<f32>(0.4, 0.25, 0.8)), sdCylinder(vec3<f32>(gsl_tmp_1.z, gsl_tmp_1.y, gsl_tmp_1.x), 0.25, 0.4));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return min(min(sdBox(p - vec3<f32>(0.0, 0.45, -0.1), vec3<f32>(0.35, 0.2, 0.4)), sdBox(p - vec3<f32>(0.0, 0.8, -0.1), vec3<f32>(0.3, 0.15, 0.3))), sdCylinder(p - vec3<f32>(0.0, 0.85, 0.2), 0.2, 0.1));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(sdTorus((p - vec3<f32>(0.45, 0.1, -0.6)).xzy, 0.3, 0.1), sdTorus((p - vec3<f32>(0.45, 0.1, 0.0)).xzy, 0.3, 0.1)), sdTorus((p - vec3<f32>(0.45, 0.1, 0.6)).xzy, 0.3, 0.1)), sdTorus((p - vec3<f32>(-0.45, 0.1, -0.6)).xzy, 0.3, 0.1)), sdTorus((p - vec3<f32>(-0.45, 0.1, 0.0)).xzy, 0.3, 0.1)), sdTorus((p - vec3<f32>(-0.45, 0.1, 0.6)).xzy, 0.3, 0.1));
}

fn sdBox(pt: vec3<f32>, b: vec3<f32>) -> f32 {
  let q = abs(pt) - b;
  let len_q_pos = sqrt(pow(max(q.x, 0.0), 2.0) + pow(max(q.y, 0.0), 2.0) + pow(max(q.z, 0.0), 2.0));
  let max_comp = max(max(q.x, q.y), q.z);
  return len_q_pos + min(max_comp, 0.0);
}

fn sdCylinder(pt: vec3<f32>, h: f32, r: f32) -> f32 {
  let d_xz = sqrt(pow(pt.x, 2.0) + pow(pt.z, 2.0)) - r;
  let d_y = abs(pt.y) - h;
  let len_d_pos = sqrt(pow(max(d_xz, 0.0), 2.0) + pow(max(d_y, 0.0), 2.0));
  let max_comp = max(d_xz, d_y);
  return len_d_pos + min(max_comp, 0.0);
}

fn sdTorus(pt: vec3<f32>, major: f32, minor: f32) -> f32 {
  let q_xz = sqrt(pow(pt.x, 2.0) + pow(pt.z, 2.0)) - major;
  let q_y = pt.y;
  return sqrt(pow(q_xz, 2.0) + pow(q_y, 2.0)) - minor;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let p_rotated = vec3<f32>(p.y, p.z, -p.x);
  let h = gsl_remap(p_rotated);
  let s = gsl_remap_2(p_rotated);
  let b = gsl_remap_3(p_rotated);
  let gsl_s1 = gsl_shape(h, 1u);
  let gsl_s2 = gsl_shape(s, 2u);
  let gsl_s3 = gsl_shape(b, 3u);
  return gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3);
}
