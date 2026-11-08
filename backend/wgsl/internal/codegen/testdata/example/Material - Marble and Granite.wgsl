fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.0, noiseTex);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn noise_cellHash3(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_cornerGradient(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return normalize(noise_cellHash3(cell, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLattice_gsl_anon(pi: vec3<f32>, pf: vec3<f32>, u: f32, v: f32, w: f32, acc: f32, o: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let wx = o.x * u + (1.0 - o.x) * (1.0 - u);
  let wy = o.y * v + (1.0 - o.y) * (1.0 - v);
  let wz = o.z * w + (1.0 - o.z) * (1.0 - w);
  let g = noise_cornerGradient(pi + o, noiseTex);
  let d = pf - o;
  return acc + wx * wy * wz * dot(g, d);
}

fn noise_gradientLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let pi = floor(q);
  let pf = q - pi;
  let u = ease_cubic(pf.x, noiseTex);
  let v = ease_cubic(pf.y, noiseTex);
  let w = ease_cubic(pf.z, noiseTex);
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLattice_gsl_anon(pi, pf, u, v, w, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn noise_cellularF1F2PairLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  var gsl_result : gsl_struct;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  gsl_result.v_3 = gsl_cellularScan(q, 1, noiseTex).nearCell;
  gsl_result.v_4 = gsl_cellularScan(q, 1, noiseTex).farCell;
  return gsl_result;
}

fn color_srgbToLinearChannel(s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x, noiseTex), color_srgbToLinearChannel(c.y, noiseTex), color_srgbToLinearChannel(c.z, noiseTex));
}

fn color_cbrt(x: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sign(x) * pow(abs(x), 0.33333334);
}

fn color_rgb2oklab(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let lin = color_linearize(c, noiseTex);
  let l = dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), lin);
  let l_ = color_cbrt(l, noiseTex);
  let m = dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), lin);
  let m_ = color_cbrt(m, noiseTex);
  let s = dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), lin);
  let s_ = color_cbrt(s, noiseTex);
  let lms = vec3<f32>(l_, m_, s_);
  return vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), lms), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), lms), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), lms));
}

fn color_linearToSrgbChannel(l: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(1.055 * pow(l, 0.41666666) - 0.055, l * 12.92, l <= 0.0031308);
}

fn color_gammaEncode(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_linearToSrgbChannel(c.x, noiseTex), color_linearToSrgbChannel(c.y, noiseTex), color_linearToSrgbChannel(c.z, noiseTex));
}

fn color_oklab2rgb(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let l_ = c.x + 0.39633778 * c.y + 0.21580376 * c.z;
  let l = l_ * l_ * l_;
  let m_ = c.x - 0.105561346 * c.y - 0.06385417 * c.z;
  let m = m_ * m_ * m_;
  let s_ = c.x - 0.08948418 * c.y - 1.2914855 * c.z;
  let s = s_ * s_ * s_;
  let lms = vec3<f32>(l, m, s);
  let lin = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), lms), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), lms), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), lms));
  return color_gammaEncode(lin, noiseTex);
}

fn material_make(rough: f32, metal: f32, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(vec4<f32>(rough, metal, index, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
}

fn stone_gradientAt(q: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(q / size, noiseTex);
}

fn stone_warpAt(q: vec3<f32>, size: f32, strength: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return q + vec3<f32>(stone_gradientAt(q, size, noiseTex), stone_gradientAt(q + vec3<f32>(37.0, 0.0, 0.0), size, noiseTex), stone_gradientAt(q + vec3<f32>(0.0, 59.0, 0.0), size, noiseTex)) * strength;
}

fn stone_fbm3At(q: vec3<f32>, size: f32, gain: f32, lacunarity: f32, noiseTex: texture_3d<f32>) -> f32 {
  return stone_gradientAt(q, size, noiseTex) + gain * stone_gradientAt(q, size / lacunarity, noiseTex) + gain * gain * stone_gradientAt(q, size / (lacunarity * lacunarity), noiseTex);
}

fn stone_veinedMarble(p: vec3<f32>, q: vec3<f32>, sheets: f32, foldSize: f32, foldDepth: f32, veinScale: f32, veinWidth: f32, sharpness: f32, crackShare: f32, facing: f32, clusterSize: f32, clusterLevel: f32, haloStrength: f32, fineStrength: f32, groundColor: vec3<f32>, cloudColor: vec3<f32>, veinColor: vec3<f32>, accentColor: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let groundLab = color_rgb2oklab(groundColor, noiseTex);
  let folded = stone_warpAt(q, foldSize, foldDepth, noiseTex);
  let cloud = texture_mask(stone_gradientAt(folded + vec3<f32>(19.0, 3.0, 47.0), veinScale * 1.5, noiseTex) + 0.5 * stone_gradientAt(folded + vec3<f32>(19.0, 3.0, 47.0), veinScale * 0.75, noiseTex), 0.15, 0.35, noiseTex);
  let zone = texture_mask(stone_gradientAt(q + vec3<f32>(5.0, 101.0, 37.0), clusterSize, noiseTex), clusterLevel, 0.06, noiseTex);
  let jagged = stone_warpAt(folded, veinScale * 0.06, veinScale * 0.02, noiseTex);
  let gsl_result = noise_cellularF1F2PairLattice(jagged / veinScale, noiseTex);
  let farCell = gsl_result.v_4;
  let farFeature = farCell + fract(textureLoad(noiseTex, vec3<i32>(floor(farCell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((farCell) / 64.0)) & vec3<i32>(63), 0).rgb);
  let nearCell = gsl_result.v_3;
  let nearFeature = nearCell + fract(textureLoad(noiseTex, vec3<i32>(floor(nearCell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((nearCell) / 64.0)) & vec3<i32>(63), 0).rgb);
  let wallNormal = normalize(farFeature - nearFeature);
  let sheetLike = texture_mask(abs(wallNormal.x), facing, 0.1, noiseTex);
  let crackNoise = stone_gradientAt((nearFeature + farFeature) * 0.5 * veinScale + vec3<f32>(3.0, 11.0, 7.0), veinScale * 1.5, noiseTex);
  let crackGate = sheetLike * texture_mask(crackNoise, (0.5 - crackShare) * 0.6, 0.04, noiseTex);
  let cracked = select(crackGate, 1.0, sheets > 0.5);
  let f2 = gsl_result.v_2;
  let f1 = gsl_result.v_1;
  let crackWall = (f2 - f1) * veinScale * 0.5;
  let sheetWall = abs(stone_fbm3At(jagged, veinScale, 0.4, 2.1, noiseTex)) * veinScale / 1.4;
  let wall = select(crackWall, sheetWall, sheets > 0.5);
  let along = stone_gradientAt(q + vec3<f32>(71.0, 13.0, 29.0), veinScale * 0.7, noiseTex);
  let halfWidth = 0.5 * veinWidth * (0.2 + 1.3 * texture_mask(along, -0.1, 0.5, noiseTex));
  let halo = haloStrength * zone * cracked * (1.0 - texture_mask(wall, 2.5 * halfWidth, 1.5 * halfWidth, noiseTex));
  let grain = (fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.35) + vec3<f32>(9.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.35) + vec3<f32>(9.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.025 * (1.0 - texture_mask(gsl_pw, 0.13125, 0.04375, noiseTex));
  let ground = groundLab + (color_rgb2oklab(cloudColor, noiseTex) - groundLab) * clamp(0.5 * cloud + halo, 0.0, 1.0) + vec3<f32>(grain, 0.0, 0.0);
  let veinLab = color_rgb2oklab(veinColor, noiseTex);
  let tinted = veinLab + (color_rgb2oklab(accentColor, noiseTex) - veinLab) * texture_mask(along, 0.3, 0.25, noiseTex);
  let vein = zone * cracked * (1.0 - texture_mask(wall, halfWidth, max(halfWidth / sharpness, gsl_pw), noiseTex));
  let gsl_result_2 = noise_cellularF1F2PairLattice((jagged + vec3<f32>(31.0, 17.0, 5.0)) / (veinScale / 3.0), noiseTex);
  let farFine = gsl_result_2.v_4;
  let nearFine = gsl_result_2.v_3;
  let fineNormal = normalize(farFine + fract(textureLoad(noiseTex, vec3<i32>(floor(farFine)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((farFine) / 64.0)) & vec3<i32>(63), 0).rgb) - nearFine - fract(textureLoad(noiseTex, vec3<i32>(floor(nearFine)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((nearFine) / 64.0)) & vec3<i32>(63), 0).rgb));
  let fineCracked = (0.3 + 0.7 * zone) * texture_mask(abs(fineNormal.x), facing * 0.7, 0.1, noiseTex) * select(0.0, 1.0, fract(textureLoad(noiseTex, vec3<i32>(floor(nearFine + farFine + vec3<f32>(5.0, 2.0, 13.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((nearFine + farFine + vec3<f32>(5.0, 2.0, 13.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x < 0.5);
  let g2 = gsl_result_2.v_2;
  let g1 = gsl_result_2.v_1;
  let fineWall = (g2 - g1) * veinScale / 6.0;
  let fineHalf = 0.3 * halfWidth;
  let fine = fineStrength * fineCracked * (1.0 - texture_mask(fineWall, fineHalf, max(fineHalf / sharpness, gsl_pw), noiseTex));
  return color_oklab2rgb(ground + (tinted - ground) * max(vein, 0.7 * fine), noiseTex);
}

fn stone_stoneFinish(rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_make(rough, 0.0, 1.55, noiseTex);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-420.01, 19.99, -0.01);
  let gsl_partHi = vec3<f32>(-19.99, 420.01, 20.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered(p, vec3<f32>(-220.0, 220.0, 10.0), vec3<f32>(400.0, 400.0, 20.0), noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let calacattaSlab = gsl_part;
  let gsl_partLo_2 = vec3<f32>(19.99, 19.99, -0.01);
  let gsl_partHi_2 = vec3<f32>(420.01, 420.01, 20.01);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_gcl_closure_boxCentered(p, vec3<f32>(220.0, 220.0, 10.0), vec3<f32>(400.0, 400.0, 20.0), noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let statuarioSlab = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-420.01, -420.01, -0.01);
  let gsl_partHi_3 = vec3<f32>(-19.99, -19.99, 20.01);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-220.0, -220.0, 10.0), vec3<f32>(400.0, 400.0, 20.0), noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let carraraSlab = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(19.99, -420.01, -0.01);
  let gsl_partHi_4 = vec3<f32>(420.01, -19.99, 20.01);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_gcl_closure_boxCentered(p, vec3<f32>(220.0, -220.0, 10.0), vec3<f32>(400.0, 400.0, 20.0), noiseTex) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let graniteSlab = gsl_part_4;
  let gsl_s1 = gsl_shape(calacattaSlab, 1u);
  let gsl_s2 = gsl_shape(statuarioSlab, 2u);
  let gsl_s3 = gsl_shape(carraraSlab, 3u);
  let gsl_s4 = gsl_shape(graniteSlab, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    var gsl_simpleValue = vec3<f32>(0.9128024, 0.9032628, 0.8845272);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue = mix(gsl_simpleValue, stone_veinedMarble(p, vec3<f32>(dot(p, vec3<f32>(0.51449573, -0.8574929, 0.0)), dot(p, vec3<f32>(-0.0, 0.0, 1.0)) / 4.0, dot(p, vec3<f32>(0.8574929, 0.51449573, 0.0)) / 4.0), 1.0, 400.0, 35.0, 220.0, 9.0, 0.7, 0.6, 0.6, 300.0, -0.2, 0.7, 0.15, vec3<f32>(0.94, 0.93, 0.91), vec3<f32>(0.84, 0.83, 0.81), vec3<f32>(0.45, 0.45, 0.46), vec3<f32>(0.7, 0.58, 0.4), noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    var gsl_simpleValue_2 = vec3<f32>(0.89379543, 0.8958852, 0.907837);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, stone_veinedMarble(p, vec3<f32>(dot(p, vec3<f32>(0.51449573, -0.8574929, 0.0)), dot(p, vec3<f32>(-0.0, 0.0, 1.0)) / 1.5, dot(p, vec3<f32>(0.8574929, 0.51449573, 0.0)) / 1.5), 0.0, 150.0, 25.0, 45.0, 0.9, 1.0, 0.45, 0.5, 200.0, -0.3, 0.12, 0.8, vec3<f32>(0.93, 0.93, 0.94), vec3<f32>(0.86, 0.87, 0.89), vec3<f32>(0.22, 0.23, 0.25), vec3<f32>(0.35, 0.36, 0.38), noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_2, 1.0);
  }
  if (gsl_colorIdx == 3u) {
    var gsl_simpleValue_3 = vec3<f32>(0.81244713, 0.82391816, 0.8399287);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, stone_veinedMarble(p, vec3<f32>(dot(p, vec3<f32>(0.51449573, -0.8574929, 0.0)), dot(p, vec3<f32>(-0.0, 0.0, 1.0)) / 3.0, dot(p, vec3<f32>(0.8574929, 0.51449573, 0.0)) / 3.0), 1.0, 200.0, 40.0, 90.0, 6.0, 0.5, 0.5, 0.6, 300.0, -0.3, 1.0, 0.5, vec3<f32>(0.89, 0.9, 0.91), vec3<f32>(0.78, 0.79, 0.81), vec3<f32>(0.58, 0.6, 0.63), vec3<f32>(0.54, 0.56, 0.6), noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_3, 1.0);
  }
  if (gsl_colorIdx == 4u) {
    var gsl_detailValue = vec3<f32>(0.4635686, 0.45140734, 0.44797245);
    if (max(gsl_pw, gsl_detailPw) < 1.3) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_38 = p / 4.0;
      let gsl_tmp_40 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_cellularScan(gsl_tmp_38, 1, noiseTex).nearCell + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_cellularScan(gsl_tmp_38, 1, noiseTex).nearCell + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_41 = (gsl_tmp_40.y - 0.5) * 0.14;
      let gsl_tmp_39 = gsl_cellularScan(gsl_tmp_38, 1, noiseTex).f1 * 4.0;
      let gsl_tmp_42 = 1.0 - texture_mask(gsl_tmp_39 / 4.0, 0.35, 0.25, noiseTex);
      gsl_detailValue = color_oklab2rgb(vec3<f32>(0.5583731, 0.0032741998, 0.0024098433) + (select(select(vec3<f32>(0.62754637, 0.0047757193, 0.0054191384) + vec3<f32>(gsl_tmp_41, 0.0, 0.0), vec3<f32>(0.717357, 0.002310573, 0.002968035) + vec3<f32>(gsl_tmp_41 * 0.5 - 0.12 * gsl_tmp_42 - 0.15 * select(0.0, 1.0, fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.24) + vec3<f32>(313.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.24) + vec3<f32>(313.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x > 0.96) * (1.0 - texture_mask(gsl_pw, 0.09, 0.03, noiseTex)), -0.004 * gsl_tmp_42, -0.012 * gsl_tmp_42), gsl_tmp_40.x > 0.55), vec3<f32>(0.21668997, 0.0013598828, -0.004748624) + vec3<f32>(gsl_tmp_41 * 0.4, 0.0, 0.0), gsl_tmp_40.x > 0.85) - vec3<f32>(0.5583731, 0.0032741998, 0.0024098433)) * (1.0 - texture_mask(gsl_pw, 0.975, 0.325, noiseTex)) - vec3<f32>(0.14 * (1.0 - texture_mask(gsl_cellularScan(gsl_tmp_38, 1, noiseTex).f2 * 4.0 - gsl_tmp_39, 0.2, 0.12, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.15, 0.05, noiseTex)), 0.0, 0.0), noiseTex);
      gsl_pw = gsl_detailOuterPw;
    }
    var gsl_simpleValue_4 = vec3<f32>(0.4635686, 0.45140734, 0.44797245);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_4 = mix(gsl_simpleValue_4, gsl_detailValue, gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_4, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) {
    var gsl_simpleValue_5 = mat4x4<f32>(vec4<f32>(0.1086, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_43 = p / 4.0;
      let gsl_tmp_44 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_cellularScan(gsl_tmp_43, 1, noiseTex).nearCell + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_cellularScan(gsl_tmp_43, 1, noiseTex).nearCell + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_simpleFull = stone_stoneFinish(select(select(0.15, 0.0495, gsl_tmp_44.x > 0.55), 0.075, gsl_tmp_44.x > 0.85), noiseTex);
      gsl_simpleValue_5 = mat4x4<f32>(mix(gsl_simpleValue_5[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_5[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_5[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_5[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let gsl_m = gsl_simpleValue_5; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
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

var<private> gsl_vanishEye: vec4<f32>;
var<private> gsl_vanishView: vec4<f32>;
var<private> gsl_partNear: vec3<f32>;

fn gsl_vanish(lo: vec3<f32>, hi: vec3<f32>) -> vec4<f32> {
  if (gsl_vanishEye.w <= 0.0) { return vec4<f32>(1.0, 0.0, 1e10, 1e10); }
  let size = hi - lo;
  let dist = length(gsl_vanishEye.xyz - clamp(gsl_vanishEye.xyz, lo, hi));
  let pixel = gsl_vanishEye.w * mix(max(dist, 1.0), 1.0, gsl_vanishView.x);
  let edge = min(size.x, min(size.y, size.z));
  let px = edge / pixel;
  let w = smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), px);
  return vec4<f32>(w, (1.0 - w) * 0.5 * edge, px, edge);
}

fn gsl_boxDistance(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  return length(max(max(lo - q, q - hi), vec3<f32>(0.0)));
}

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_detailPw: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
  v_4 : vec3<f32>,
}
