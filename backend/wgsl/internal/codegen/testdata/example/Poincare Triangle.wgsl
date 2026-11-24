fn rotationalGeodesic(p: vec3<f32>, count: f32, halfShift: f32) -> f32 {
  let r = length(p.xy);
  let theta = atan2(p.y, p.x);
  let sectorWidth = 6.2831855 / count;
  let halfSector = sectorWidth / 2.0;
  let thetaAdjusted = theta - halfShift * halfSector;
  let thetaShifted = thetaAdjusted + halfSector;
  let thetaNorm = thetaShifted - floor(thetaShifted / sectorWidth) * sectorWidth;
  let thetaFolded = thetaNorm - halfSector;
  let thetaFinal = abs(thetaFolded);
  let px = r * cos(thetaFinal);
  let py = r * sin(thetaFinal);
  let sectorAngle = 3.1415927 / count;
  let geoDistance = 1.0 / cos(sectorAngle);
  let geoRadius = tan(sectorAngle);
  let geodesicDisc = length(vec2<f32>(px, py) - vec2<f32>(geoDistance, 0.0)) - geoRadius;
  let insideDisc = length(p.xy) - 1.0;
  return max(max(geodesicDisc, insideDisc), abs(p.z) - 0.05);
}

fn poincareTiling(p: vec3<f32>, n: f32) -> gsl_struct {
  let unitCircle = length(p.xy) - 1.0;
  let slab = abs(p.z) - 0.05;
  let circleResult = max(unitCircle, slab);
  let g1 = rotationalGeodesic(p, n, 0.0);
  let g2 = rotationalGeodesic(p, 2.0 * n, 1.0);
  let g3 = rotationalGeodesic(p, 4.0 * n, 1.0);
  let g4 = rotationalGeodesic(p, 8.0 * n, 1.0);
  let g5 = rotationalGeodesic(p, 16.0 * n, 1.0);
  let g6 = rotationalGeodesic(p, 32.0 * n, 1.0);
  let g7 = rotationalGeodesic(p, 64.0 * n, 1.0);
  let g8 = rotationalGeodesic(p, 128.0 * n, 1.0);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = circleResult;
  gsl_result.v_2 = g1;
  gsl_result.v_3 = g2;
  gsl_result.v_4 = g3;
  gsl_result.v_5 = g4;
  gsl_result.v_6 = g5;
  gsl_result.v_7 = g6;
  gsl_result.v_8 = g7;
  gsl_result.v_9 = g8;
  return gsl_result;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let t = sin(time) * 4.0;
  let nr = 2.0 * p.x + t * p.y;
  let dr = -t * p.y + 2.0;
  let ni = 2.0 * p.y - t * p.x - t;
  let di = t * p.x + t;
  let denom2 = dr * dr + di * di;
  let wx = (nr * dr + ni * di) / denom2;
  let wy = (ni * dr - nr * di) / denom2;
  let gsl_result = poincareTiling(vec3<f32>(wx, wy, p.z), 3.0);
  let t0 = gsl_result.v_1;
  let t1 = gsl_result.v_2;
  let t2 = gsl_result.v_3;
  let t3 = gsl_result.v_4;
  let t4 = gsl_result.v_5;
  let t5 = gsl_result.v_6;
  let t6 = gsl_result.v_7;
  let t7 = gsl_result.v_8;
  let t8 = gsl_result.v_9;
  let gsl_s1 = gsl_shape(t0, 1u);
  let gsl_s2 = gsl_shape(t1, 2u);
  let gsl_s3 = gsl_shape(t2, 3u);
  let gsl_s4 = gsl_shape(t3, 4u);
  let gsl_s5 = gsl_shape(t4, 5u);
  let gsl_s6 = gsl_shape(t5, 6u);
  let gsl_s7 = gsl_shape(t6, 7u);
  let gsl_s8 = gsl_shape(t7, 8u);
  let gsl_s9 = gsl_shape(t8, 9u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : f32,
  v_4 : f32,
  v_5 : f32,
  v_6 : f32,
  v_7 : f32,
  v_8 : f32,
  v_9 : f32,
}
