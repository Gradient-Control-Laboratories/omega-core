fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_circleXY(p: vec3<f32>, center: vec2<f32>, radius: f32) -> f32 {
  return length(p.xy - center) - radius;
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_1 = length(p.xy);
  let gsl_tmp_2 = gsl_floored((atan2(p.y, p.x) + 0.3926991), 0.7853982) - 0.3926991;
  let rosette = motif(vec3<f32>(gsl_tmp_1 * cos(gsl_tmp_2), gsl_tmp_1 * sin(gsl_tmp_2), p.z));
  let rosetteDisk = max(rosette, length(p.xy) - 11.0);
  return rosetteDisk;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let tiled = motif(vec3<f32>(map_triangleWave(p.x, 6.0), map_triangleWave(p.y, 6.0), p.z));
  let tiledBlock = max(tiled, builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(22.0, 22.0, 1.2)));
  return tiledBlock;
}

fn builtin_intersectSharp(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound(fields: vec2<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_rectangleXYCentered(p: vec3<f32>, center: vec2<f32>, size: vec2<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p.xy - center) - halfSize;
  return builtin_gcl_closure_intersectRound(vec2<f32>(d.x, d.y));
}

fn builtin_intersectSharp_2(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound_2(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound_2(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_2(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn map_triangleWave(x: f32, period: f32) -> f32 {
  let half = period * 0.5;
  let twoPeriod = period * 2.0;
  let u = gsl_floored((gsl_floored((x + half), twoPeriod) + twoPeriod), twoPeriod);
  return half - abs(u - period);
}

fn motif(p: vec3<f32>) -> f32 {
  let bar = builtin_gcl_closure_rectangleXYCentered(p, vec2<f32>(1.6, 0.0), vec2<f32>(3.0, 0.7));
  let blob = builtin_circleXY(p, vec2<f32>(0.0, 2.2), 0.7);
  let flat = min(bar, blob);
  return max(flat, abs(p.z) - 0.5);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let rosettePlaced = gsl_remap(vec3<f32>(p.x + 13.0, p.y, p.z));
  let tiledPlaced = gsl_remap_2(vec3<f32>(p.x - 13.0, p.y, p.z));
  let gsl_s1 = gsl_shape(rosettePlaced, 1u);
  let gsl_s2 = gsl_shape(tiledPlaced, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
