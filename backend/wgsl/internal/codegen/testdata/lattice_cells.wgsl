fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn map_triangleWave(x: f32, period: f32) -> f32 {
  let half = period * 0.5;
  let twoPeriod = period * 2.0;
  let u = gsl_floored((gsl_floored((x + half), twoPeriod) + twoPeriod), twoPeriod);
  return half - abs(u - period);
}

fn map_triangleDistance(x: f32, period: f32) -> f32 {
  return abs(map_triangleWave(x, period));
}

fn map_triangleWaveOddPositive(x: f32, period: f32) -> f32 {
  return map_triangleDistance(x - period * 0.5, period);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_unionRound(fields: vec3<f32>, radius: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectRound(negated, radius);
}

fn builtin_gcl_closure_unionRound(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.5);
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn map_repeatCentered(x: f32, period: f32) -> f32 {
  let half = period * 0.5;
  return gsl_floored((gsl_floored((x + half), period) + period), period) - half;
}

fn builtin_unionSharp_2(fields: vec3<f32>) -> f32 {
  return min(min(fields.x, fields.y), fields.z);
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields)) + sqrt(sumOfSquares);
}

fn builtin_unionRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let negated = vec2<f32>(-fields.x, -fields.y);
  return -builtin_intersectRound_2(negated, radius);
}

fn builtin_gcl_closure_unionRound_2(fields: vec2<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.3);
}

fn map_triangleWaveEvenPositive(x: f32, period: f32) -> f32 {
  return map_triangleDistance(x, period);
}

fn builtin_gcl_closure_unionRound_3(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.3);
}

fn builtin_gcl_closure_unionRound_4(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.4);
}

fn builtin_gcl_closure_unionRound_5(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.2);
}

fn builtin_gcl_closure_unionRound_6(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.2);
}

fn honeycomb_segmentDistance(point: vec2<f32>, a: vec2<f32>, b: vec2<f32>) -> f32 {
  let pa = point - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn honeycomb_gcl_closure_hexStar(point: vec2<f32>, starCenter: vec2<f32>, transition: vec2<f32>, topY: f32) -> f32 {
  let lineV = honeycomb_segmentDistance(point, starCenter, vec2<f32>(0.0, topY));
  let lineR = honeycomb_segmentDistance(point, starCenter, transition);
  let lineL = honeycomb_segmentDistance(point, starCenter, vec2<f32>(-transition.x, transition.y));
  return builtin_gcl_closure_unionRound_4(vec3<f32>(lineV, lineR, lineL));
}

fn honeycomb_gcl_closure_hexStar_2(point: vec2<f32>, starCenter: vec2<f32>, transition: vec2<f32>, topY: f32) -> f32 {
  let lineV = honeycomb_segmentDistance(point, starCenter, vec2<f32>(0.0, topY));
  let lineR = honeycomb_segmentDistance(point, starCenter, transition);
  let lineL = honeycomb_segmentDistance(point, starCenter, vec2<f32>(-transition.x, transition.y));
  return builtin_gcl_closure_unionRound_4(vec3<f32>(lineV, lineR, lineL));
}

fn honeycomb_gcl_closure_hexStar_3(point: vec2<f32>, starCenter: vec2<f32>, transition: vec2<f32>, topY: f32) -> f32 {
  let lineV = honeycomb_segmentDistance(point, starCenter, vec2<f32>(0.0, topY));
  let lineR = honeycomb_segmentDistance(point, starCenter, transition);
  let lineL = honeycomb_segmentDistance(point, starCenter, vec2<f32>(-transition.x, transition.y));
  return builtin_gcl_closure_unionRound_5(vec3<f32>(lineV, lineR, lineL));
}

fn honeycomb_gcl_closure_hexStar_4(point: vec2<f32>, starCenter: vec2<f32>, transition: vec2<f32>, topY: f32) -> f32 {
  let lineV = honeycomb_segmentDistance(point, starCenter, vec2<f32>(0.0, topY));
  let lineR = honeycomb_segmentDistance(point, starCenter, transition);
  let lineL = honeycomb_segmentDistance(point, starCenter, vec2<f32>(-transition.x, transition.y));
  return builtin_gcl_closure_unionRound_5(vec3<f32>(lineV, lineR, lineL));
}

fn honeycomb_gcl_closure_triStar(point: vec2<f32>, sx: f32, sy: f32) -> f32 {
  let distH = abs(point.y);
  let distP60 = abs(dot(point, normalize(vec2<f32>(sy, sx))));
  let distN60 = abs(dot(point, normalize(vec2<f32>(sy, -sx))));
  return builtin_gcl_closure_unionRound_6(vec3<f32>(distH, distP60, distN60));
}

fn honeycomb_gcl_closure_triStar_2(point: vec2<f32>, sx: f32, sy: f32) -> f32 {
  let distH = abs(point.y);
  let distP60 = abs(dot(point, normalize(vec2<f32>(sy, sx))));
  let distN60 = abs(dot(point, normalize(vec2<f32>(sy, -sx))));
  return builtin_gcl_closure_unionRound_6(vec3<f32>(distH, distP60, distN60));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let cub = builtin_gcl_closure_unionRound(vec3<f32>(map_triangleWaveOddPositive(p.x, 6.0), map_triangleWaveOddPositive(p.y, 6.0), map_triangleWaveOddPositive(p.z, 6.0))) - 0.5;
  let sqr = builtin_unionSharp(vec2<f32>(map_triangleWaveOddPositive(p.x, 5.0), map_triangleWaveOddPositive(p.y, 5.0))) - 0.4;
  let dia = builtin_unionSharp(vec2<f32>(map_triangleDistance((p.x + p.y) / 1.4142135, 5.0), map_triangleDistance((p.x - p.y) / 1.4142135, 7.0))) - 0.4;
  let rf1 = builtin_unionSharp(vec2<f32>(builtin_unionSharp(vec2<f32>(map_triangleWaveOddPositive(p.x, 5.0), map_triangleWaveOddPositive(p.y, 5.0))), abs(dot(vec2<f32>(map_repeatCentered(p.x, 5.0), map_repeatCentered(p.y, 5.0)), vec2<f32>(0.70710677, 0.70710677))))) - 0.4;
  let gsl_tmp_1 = vec2<f32>(map_repeatCentered(p.x, 5.0), map_repeatCentered(p.y, 5.0));
  let rf2 = builtin_unionSharp_2(vec3<f32>(builtin_unionSharp(vec2<f32>(map_triangleWaveOddPositive(p.x, 5.0), map_triangleWaveOddPositive(p.y, 5.0))), abs(dot(gsl_tmp_1, vec2<f32>(0.70710677, 0.70710677))), abs(dot(gsl_tmp_1, vec2<f32>(-0.70710677, 0.70710677))))) - 0.4;
  let oct = builtin_gcl_closure_unionRound_3(vec3<f32>(builtin_gcl_closure_unionRound_2(vec2<f32>(map_triangleWaveOddPositive(p.x, 6.0), map_triangleWaveOddPositive(p.y, 6.0))), builtin_gcl_closure_unionRound_2(vec2<f32>(map_triangleWaveEvenPositive(p.x, 6.0), map_triangleWaveEvenPositive(p.y, 6.0))), builtin_gcl_closure_unionRound_2(vec2<f32>(map_triangleWaveEvenPositive((p.x + p.y) / 1.4142135, 4.2426405), map_triangleWaveEvenPositive((p.x - p.y) / 1.4142135, 4.2426405))))) - 0.4;
  let gsl_tmp_2 = abs(map_repeatCentered(p.x, 4.0));
  let gsl_tmp_3 = abs(map_repeatCentered(p.y, 6.928203));
  let hex = select(honeycomb_gcl_closure_hexStar_2(vec2<f32>(gsl_tmp_2 - 2.0, 3.4641016 - gsl_tmp_3), vec2<f32>(0.0, 2.309401), vec2<f32>(2.0, 1.1547005), 6.928203), honeycomb_gcl_closure_hexStar(vec2<f32>(gsl_tmp_2, gsl_tmp_3), vec2<f32>(0.0, 2.309401), vec2<f32>(2.0, 1.1547005), 6.928203), gsl_tmp_2 < 1.0) - 0.3;
  let gsl_tmp_4 = abs(map_repeatCentered(p.x, 4.0));
  let gsl_tmp_5 = abs(map_repeatCentered(p.y, 6.928203));
  let hxs = select(honeycomb_gcl_closure_hexStar_4(vec2<f32>(gsl_tmp_4 - 2.0, 3.4641016 - gsl_tmp_5), vec2<f32>(0.0, 2.3105557), vec2<f32>(2.0, 1.1535459), 6.928203), honeycomb_gcl_closure_hexStar_3(vec2<f32>(gsl_tmp_4, gsl_tmp_5), vec2<f32>(0.0, 2.3105557), vec2<f32>(2.0, 1.1535459), 6.928203), gsl_tmp_4 < 1.0) - 0.3;
  let gsl_tmp_6 = abs(map_repeatCentered(p.x, 4.0));
  let gsl_tmp_7 = abs(map_repeatCentered(p.y, 6.928203));
  let tri = select(honeycomb_gcl_closure_triStar_2(vec2<f32>(gsl_tmp_6 - 2.0, gsl_tmp_7 - 3.4641016), 4.0, 6.928203), honeycomb_gcl_closure_triStar(vec2<f32>(gsl_tmp_6, gsl_tmp_7), 4.0, 6.928203), gsl_tmp_7 < 1.7320508) - 0.3;
  let gsl_s1 = gsl_shape(min(min(min(cub, sqr), min(dia, rf1)), min(min(rf2, oct), min(hex, min(hxs, tri)))), 1u);
  return gsl_s1;
}
