fn gsl_compare2(x : f32, y : f32) -> f32 {
  if x > y { return 1.0; }
  if x < y { return -1.0; }
  return 0.0;
}

fn gsl_select3(v : f32, a : f32, b : f32, c : f32) -> f32 {
  if v > 0.0 { return a; }
  if v < 0.0 { return c; }
  return b;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let b = p.x;
  let c = 6.0 - b;
  let d = min(min(3.0, b), c);
  let e = select(0.0, 1.0, !(d > 0.0)) + -b + select(0.0, 1.0, (c < 7.0));
  let f = pow(e, 2.0);
  let g = pow(f - 7.0, 2.0);
  let h = abs(g);
  let i = gsl_compare2(h, d);
  let j = select(c, b, i > h);
  let k = gsl_select3(j, b, c, d);
  let a2 = vec2<f32>(j, k);
  let a3 = vec3<f32>(a2.x, a2.y, d);
  let a4 = vec4<f32>(a3.x, a3.y, a3.z, d);
  let a5 = array<f32, 5>(a4.x, a4.y, a4.z, a4.w, h);
  let a6 = array<f32, 6>(a5[0], a5[1], a5[2], a5[3], a5[3], j);
  let gsl_s1 = gsl_shape(a6, 1u);
  return gsl_s1;
}
