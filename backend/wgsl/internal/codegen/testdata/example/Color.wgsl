fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn gsl_remap(p: vec3<f32>, gsl_param_4_10: f32, gsl_param_8_11: f32, time: f32, noiseTex: texture_3d<f32>) -> f32 {
  let ball = length(p) - 5.0;
  let gsl_tmp_1 = p / 1.0;
  let cellular = (gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1) * 1.0;
  let ballSpike = ball + cellular * 0.2;
  return ballSpike;
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_4_10: f32, gsl_param_8_11: f32, time: f32, vector_dirZ: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let ballRot = gsl_remap(map_rotateZX(p, time * 0.5, noiseTex), gsl_param_4_10, gsl_param_8_11, time, noiseTex);
  return ballRot;
}

fn builtin_intersectSharp(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound(fields: vec2<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_unionRound(fields: vec2<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let negated = vec2<f32>(-fields.x, -fields.y);
  return -builtin_intersectRound(negated, radius, noiseTex);
}

fn gsl_remap_3(p: vec3<f32>, gsl_param_4_10: f32, gsl_param_8_11: f32, time: f32, vector_dirZ: vec3<f32>, color_viridisData: array<vec3<f32>, 11>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_1 = p / 1.0;
  let cellular = (gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1) * 1.0;
  return color_sampleMap(color_viridisData, cellular * 1.2, noiseTex);
}

fn gsl_remap_4(p: vec3<f32>, gsl_param_4_10: f32, gsl_param_8_11: f32, time: f32, vector_dirZ: vec3<f32>, color_viridisData: array<vec3<f32>, 11>, color_gray: vec3<f32>, color_lightGray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let checkerboard = map_triangleWave(p.x, 2.0, noiseTex) * map_triangleWave(p.y, 2.0, noiseTex);
  return select(color_lightGray, color_gray, checkerboard > 0.0);
}

fn gsl_remap_5(p: vec3<f32>, gsl_param_4_10: f32, gsl_param_8_11: f32, time: f32, vector_dirZ: vec3<f32>, color_viridisData: array<vec3<f32>, 11>, color_gray: vec3<f32>, color_lightGray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let flr = p.z;
  return flr;
}

fn gsl_remap_6(p: vec3<f32>, gsl_param_4_10: f32, gsl_param_8_11: f32, time: f32, vector_dirZ: vec3<f32>, color_viridisData: array<vec3<f32>, 11>, color_gray: vec3<f32>, color_lightGray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let height = gsl_param_4_10;
  let ballTrans = gsl_remap_2(p - vector_dirZ * height, gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, noiseTex);
  let ball2 = length(p - vec3<f32>(6.0, 0.0, 2.0)) - 2.0;
  let ballrad = gsl_param_8_11;
  let balls = builtin_unionRound(vec2<f32>(ballTrans, ball2), ballrad, noiseTex);
  return balls;
}

fn gsl_remap_7(p: vec3<f32>, gsl_param_4_10: f32, gsl_param_8_11: f32, time: f32, vector_dirZ: vec3<f32>, color_viridisData: array<vec3<f32>, 11>, color_gray: vec3<f32>, color_lightGray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_1 = p / 1.0;
  let cellular = (gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1) * 1.0;
  return color_sampleMap(color_viridisData, cellular * 1.2, noiseTex);
}

fn color_sampleMap(palette: array<vec3<f32>, 11>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let tt = clamp(t, 0.0, 1.0);
  let weighted = array<vec3<f32>, 11>(palette[0] * max(0.0, 1.0 - abs(tt * 10.0 - 0.0)), palette[1] * max(0.0, 1.0 - abs(tt * 10.0 - 1.0)), palette[2] * max(0.0, 1.0 - abs(tt * 10.0 - 2.0)), palette[3] * max(0.0, 1.0 - abs(tt * 10.0 - 3.0)), palette[4] * max(0.0, 1.0 - abs(tt * 10.0 - 4.0)), palette[5] * max(0.0, 1.0 - abs(tt * 10.0 - 5.0)), palette[6] * max(0.0, 1.0 - abs(tt * 10.0 - 6.0)), palette[7] * max(0.0, 1.0 - abs(tt * 10.0 - 7.0)), palette[8] * max(0.0, 1.0 - abs(tt * 10.0 - 8.0)), palette[9] * max(0.0, 1.0 - abs(tt * 10.0 - 9.0)), palette[10] * max(0.0, 1.0 - abs(tt * 10.0 - 10.0)));
  return weighted[0] + weighted[1] + weighted[2] + weighted[3] + weighted[4] + weighted[5] + weighted[6] + weighted[7] + weighted[8] + weighted[9] + weighted[10];
}

fn map_triangleWave(x: f32, period: f32, noiseTex: texture_3d<f32>) -> f32 {
  let half = period * 0.5;
  let twoPeriod = period * 2.0;
  let u = gsl_floored((gsl_floored((x + half), twoPeriod) + twoPeriod), twoPeriod);
  return half - abs(u - period);
}

fn map_rotateZX(v: vec3<f32>, ang: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let ca = cos(ang);
  let sa = sin(ang);
  return vec3<f32>(v.x * ca - v.z * sa, v.y, v.z * ca + v.x * sa);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_4_10: f32, gsl_param_8_11: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let height = gsl_param_4_10;
  let ballTrans = gsl_remap_2(p - vector_dirZ * height, gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, noiseTex);
  let ball2 = length(p - vec3<f32>(6.0, 0.0, 2.0)) - 2.0;
  let ballrad = gsl_param_8_11;
  let balls = builtin_unionRound(vec2<f32>(ballTrans, ball2), ballrad, noiseTex);
  let flr = p.z;
  let body = builtin_unionRound(vec2<f32>(balls, flr), ballrad, noiseTex);
  let gsl_s1 = gsl_shape(body, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_4_10: f32, gsl_param_8_11: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let height = gsl_param_4_10;
    let color_viridisData = array<vec3<f32>, 11>(vec3<f32>(0.267, 0.005, 0.329), vec3<f32>(0.283, 0.141, 0.458), vec3<f32>(0.254, 0.265, 0.53), vec3<f32>(0.207, 0.372, 0.553), vec3<f32>(0.164, 0.471, 0.558), vec3<f32>(0.128, 0.567, 0.551), vec3<f32>(0.135, 0.659, 0.518), vec3<f32>(0.267, 0.749, 0.441), vec3<f32>(0.478, 0.821, 0.318), vec3<f32>(0.741, 0.873, 0.15), vec3<f32>(0.993, 0.906, 0.144));
    let ballTrans = gsl_remap_2(p - vector_dirZ * height, gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, noiseTex);
    let ball2 = length(p - vec3<f32>(6.0, 0.0, 2.0)) - 2.0;
    let ballrad = gsl_param_8_11;
    let balls = builtin_unionRound(vec2<f32>(ballTrans, ball2), ballrad, noiseTex);
    let flr = p.z;
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_lightGray = vec3<f32>(0.75, 0.75, 0.75);
    return vec4<f32>(clamp((clamp(clamp(gsl_remap_3(map_rotateZX(p - vector_dirZ * height, time * 0.5, noiseTex), gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, color_viridisData, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)) * -(select(select(-0.0, -1.0, -balls > -flr), -0.0, -ballrad < max(-balls, -flr)) + (select(0.0, -1.0, -balls + ballrad > 0.0) * max(-balls + ballrad, 0.0) + max(-balls + ballrad, 0.0) * select(0.0, -1.0, -balls + ballrad > 0.0) + select(0.0, -0.0, -flr + ballrad > 0.0) * max(-flr + ballrad, 0.0) + max(-flr + ballrad, 0.0) * select(0.0, -0.0, -flr + ballrad > 0.0)) / (2.0 * (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) + 1.0 - select(0.0, 1.0, (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) > 0.0))))) + gsl_remap_4(p, gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, color_viridisData, color_gray, color_lightGray, noiseTex) * -(select(select(-1.0, -0.0, -balls > -flr), -0.0, -ballrad < max(-balls, -flr)) + (select(0.0, -0.0, -balls + ballrad > 0.0) * max(-balls + ballrad, 0.0) + max(-balls + ballrad, 0.0) * select(0.0, -0.0, -balls + ballrad > 0.0) + select(0.0, -1.0, -flr + ballrad > 0.0) * max(-flr + ballrad, 0.0) + max(-flr + ballrad, 0.0) * select(0.0, -1.0, -flr + ballrad > 0.0)) / (2.0 * (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) + 1.0 - select(0.0, 1.0, (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) > 0.0))))) + select(clamp(clamp(gsl_remap_7(map_rotateZX(p - vector_dirZ * height, time * 0.5, noiseTex), gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, color_viridisData, color_gray, color_lightGray, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_remap_4(p, gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, color_viridisData, color_gray, color_lightGray, noiseTex), gsl_remap_5(p, gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, color_viridisData, color_gray, color_lightGray, noiseTex) < gsl_remap_6(p, gsl_param_4_10, gsl_param_8_11, time, vector_dirZ, color_viridisData, color_gray, color_lightGray, noiseTex)) * max(1.0 - (-(select(select(-0.0, -1.0, -balls > -flr), -0.0, -ballrad < max(-balls, -flr)) + (select(0.0, -1.0, -balls + ballrad > 0.0) * max(-balls + ballrad, 0.0) + max(-balls + ballrad, 0.0) * select(0.0, -1.0, -balls + ballrad > 0.0) + select(0.0, -0.0, -flr + ballrad > 0.0) * max(-flr + ballrad, 0.0) + max(-flr + ballrad, 0.0) * select(0.0, -0.0, -flr + ballrad > 0.0)) / (2.0 * (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) + 1.0 - select(0.0, 1.0, (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) > 0.0))))) + -(select(select(-1.0, -0.0, -balls > -flr), -0.0, -ballrad < max(-balls, -flr)) + (select(0.0, -0.0, -balls + ballrad > 0.0) * max(-balls + ballrad, 0.0) + max(-balls + ballrad, 0.0) * select(0.0, -0.0, -balls + ballrad > 0.0) + select(0.0, -1.0, -flr + ballrad > 0.0) * max(-flr + ballrad, 0.0) + max(-flr + ballrad, 0.0) * select(0.0, -1.0, -flr + ballrad > 0.0)) / (2.0 * (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) + 1.0 - select(0.0, 1.0, (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) > 0.0)))))), 0.0)) / max(-(select(select(-0.0, -1.0, -balls > -flr), -0.0, -ballrad < max(-balls, -flr)) + (select(0.0, -1.0, -balls + ballrad > 0.0) * max(-balls + ballrad, 0.0) + max(-balls + ballrad, 0.0) * select(0.0, -1.0, -balls + ballrad > 0.0) + select(0.0, -0.0, -flr + ballrad > 0.0) * max(-flr + ballrad, 0.0) + max(-flr + ballrad, 0.0) * select(0.0, -0.0, -flr + ballrad > 0.0)) / (2.0 * (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) + 1.0 - select(0.0, 1.0, (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) > 0.0))))) + -(select(select(-1.0, -0.0, -balls > -flr), -0.0, -ballrad < max(-balls, -flr)) + (select(0.0, -0.0, -balls + ballrad > 0.0) * max(-balls + ballrad, 0.0) + max(-balls + ballrad, 0.0) * select(0.0, -0.0, -balls + ballrad > 0.0) + select(0.0, -1.0, -flr + ballrad > 0.0) * max(-flr + ballrad, 0.0) + max(-flr + ballrad, 0.0) * select(0.0, -1.0, -flr + ballrad > 0.0)) / (2.0 * (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) + 1.0 - select(0.0, 1.0, (sqrt(0.0 + max(-balls + ballrad, 0.0) * max(-balls + ballrad, 0.0) + max(-flr + ballrad, 0.0) * max(-flr + ballrad, 0.0)) > 0.0))))), 1.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
struct gsl_CellularScan {
  f1: f32,
  f2: f32,
  feature: vec3<f32>,
  farFeature: vec3<f32>,
  nearCell: vec3<f32>,
  farCell: vec3<f32>,
}

fn gsl_cellularHash(cell: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let fine = fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
  if (oct <= 1) { return fine; }
  // fract(fract(a) + b) == fract(a + b): the inner fract only removes an
  // integer, so folding the supercell on top of the already-fracted fine
  // lookup is the same value noiseHash3's two-lookup form produces.
  return fract(fine + textureLoad(noiseTex, vec3<i32>(floor(cell / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn gsl_cellularScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularScan {
  let pi = floor(q);
  let pf = q - pi;
  // Seeded with an unreachable distance, squared, because the scan compares
  // squared distances and roots the two winners at the end.
  var best = 64.000000;
  var second = 64.000000;
  var bestCell = pi;
  var bestH = vec3<f32>(0.0);
  var secondCell = pi;
  var secondH = vec3<f32>(0.0);
  for (var i = 0; i < 27; i = i + 1) {
    // The same offset order cellularOffsets() lists: x fastest, then y, then z.
    let o = vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
    let h = gsl_cellularHash(pi + o, oct, noiseTex);
    let v = pf - (o + h);
    let d2 = dot(v, v);
    if (d2 < best) {
      second = best;
      secondCell = bestCell;
      secondH = bestH;
      best = d2;
      bestCell = pi + o;
      bestH = h;
    } else if (d2 < second) {
      second = d2;
      secondCell = pi + o;
      secondH = h;
    }
  }
  var out: gsl_CellularScan;
  out.f1 = sqrt(best);
  out.f2 = sqrt(second);
  // Both closest points, so both gradients are available without a second hash.
  out.feature = bestCell + bestH;
  out.farFeature = secondCell + secondH;
  out.nearCell = bestCell;
  out.farCell = secondCell;
  return out;
}
