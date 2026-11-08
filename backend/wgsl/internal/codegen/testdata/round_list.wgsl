fn intersectRound(list: vec3<f32>, k: f32) -> f32 {
  let maxAll = max(max(max(list.x, list.x), list.y), list.z);
  let sumSquares = 0.0 + max(list.x + k, 0.0) * max(list.x + k, 0.0) + max(list.y + k, 0.0) * max(list.y + k, 0.0) + max(list.z + k, 0.0) * max(list.z + k, 0.0);
  return min(-k, maxAll) + sqrt(sumSquares);
}

fn intersectRound_2(d1: f32, d2: f32, d3: f32, k: f32) -> f32 {
  let c1 = max(d1 + k, 0.0);
  let c2 = max(d2 + k, 0.0);
  let c3 = max(d3 + k, 0.0);
  return min(-k, max(max(d1, d2), d3)) + length(vec3<f32>(c1, c2, c3));
}

fn unionRound(list: vec3<f32>, k: f32) -> f32 {
  let negated = vec3<f32>(-list.x, -list.y, -list.z);
  return -intersectRound(negated, k);
}

fn unionRound_2(d1: f32, d2: f32, d3: f32, k: f32) -> f32 {
  return -intersectRound_2(-d1, -d2, -d3, k);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let viaList = intersectRound(vec3<f32>(p.x, p.y, p.z), 0.2);
  let viaArgs = intersectRound_2(p.x, p.y, p.z, 0.2);
  let uViaList = unionRound(vec3<f32>(p.x, p.y, p.z), 0.2);
  let uViaArgs = unionRound_2(p.x, p.y, p.z, 0.2);
  let gsl_s1 = gsl_shape(viaList, 1u);
  let gsl_s2 = gsl_shape(viaArgs, 2u);
  let gsl_s3 = gsl_shape(uViaList, 3u);
  let gsl_s4 = gsl_shape(uViaArgs, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}
