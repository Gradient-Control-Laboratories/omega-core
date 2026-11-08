fn sdBox(p: vec3<f32>, b: vec3<f32>) -> f32 {
  let q = abs(p) - b;
  let max_q = vec3<f32>(max(q.x, 0.0), max(q.y, 0.0), max(q.z, 0.0));
  return length(max_q) + min(max(max(q.x, q.y), q.z), 0.0);
}

fn sdTorus(p: vec3<f32>, t: vec2<f32>) -> f32 {
  let q = vec2<f32>(length(p.xy) - t.x, p.z);
  return length(q) - t.y;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let q_hull = p / vec3<f32>(4.8, 1.2, 1.2);
  let hull = (length(q_hull) - 1.0) * 1.2;
  let conningTower = sdBox(p - vec3<f32>(0.8, 0.0, 1.2), vec3<f32>(1.2, 0.56, 0.8));
  let fin_h = sdBox(p - vec3<f32>(-4.4, 0.0, 0.0), vec3<f32>(0.8, 2.4, 0.08));
  let fin_v = sdBox(p - vec3<f32>(-4.4, 0.0, 0.0), vec3<f32>(0.8, 0.08, 1.44));
  let tail = min(fin_h, fin_v);
  let p1_prop = p - vec3<f32>(-4.96, 0.64, 0.0);
  let prop1 = sdTorus(p1_prop.yzx, vec2<f32>(0.48, 0.08));
  let p2_prop = p - vec3<f32>(-4.96, -0.64, 0.0);
  let prop2 = sdTorus(p2_prop.yzx, vec2<f32>(0.48, 0.08));
  let props = min(prop1, prop2);
  let submarine = min(min(min(hull, conningTower), tail), props);
  let gsl_s1 = gsl_shape(submarine, 1u);
  return gsl_s1;
}
