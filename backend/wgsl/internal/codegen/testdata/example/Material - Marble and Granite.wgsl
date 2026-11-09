fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sharp = builtin_intersectSharp(fields, noiseTex);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
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

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
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

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
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

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn noise_cellHash3(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_cornerGradient(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return normalize(noise_cellHash3(cell, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLatticeVector_gsl_anon(pi: vec3<f32>, pf: vec3<f32>, e: vec3<f32>, acc: vec3<f32>, o: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let weight = (o.x * e.x + (1.0 - o.x) * (1.0 - e.x)) * (o.y * e.y + (1.0 - o.y) * (1.0 - e.y)) * (o.z * e.z + (1.0 - o.z) * (1.0 - e.z));
  let g = noise_cornerGradient(pi + o, noiseTex);
  let d = pf - o;
  return acc + weight * vec3<f32>(dot(g, d), dot(vec3<f32>(g.y, g.z, g.x), d), dot(vec3<f32>(g.z, g.x, g.y), d));
}

fn noise_gradientLatticeVector(q: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let pi = floor(q);
  let pf = q - pi;
  let e = vec3<f32>(ease_cubic(pf.x, noiseTex), ease_cubic(pf.y, noiseTex), ease_cubic(pf.z, noiseTex));
  var gsl_fold = vec3<f32>(0.0, 0.0, 0.0);
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLatticeVector_gsl_anon(pi, pf, e, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn stone_grainOf(c: vec3<f32>, size: f32, s: f32, zone: f32, zoneWeight: f32, clump: f32, seed: vec3<f32>, scatter: f32, level: f32, noiseTex: texture_3d<f32>) -> f32 {
  let at = (c + fract(textureLoad(noiseTex, vec3<i32>(floor(c)) & vec3<i32>(63), 0).rgb)) * size;
  let cluster = zoneWeight * noise_gradientLattice(at / (zone * s) + seed, noiseTex) + 0.6 * noise_gradientLattice(at / (clump * s) + seed * 1.7, noiseTex);
  return select(0.0, 1.0, cluster - scatter * fract(textureLoad(noiseTex, vec3<i32>(floor(c + seed)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((c + seed) / 64.0)) & vec3<i32>(63), 0).rgb).x > level);
}

fn stone_quartzOf(c: vec3<f32>, g: f32, s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return stone_grainOf(c, g, s, 4.0, 0.75, 1.2, vec3<f32>(7.3, 3.1, 11.7), 0.8, -0.31, noiseTex);
}

fn stone_maficOf(c: vec3<f32>, t: f32, s: f32, nearQuartz: f32, noiseTex: texture_3d<f32>) -> f32 {
  return stone_grainOf(c, t, s, 4.0, 0.25, 0.6, vec3<f32>(13.1, 7.7, 2.3), 0.25, 0.17 - 0.15 * nearQuartz, noiseTex);
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

fn stone_veinFrame(p: vec3<f32>, direction: vec3<f32>, alongStretch: f32, sheetStretch: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let along = normalize(direction);
  let up = select(vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 0.0), abs(along.z) > 0.9);
  let across = normalize(cross(along, up));
  let sheet = cross(across, along);
  return vec3<f32>(dot(p, across), dot(p, sheet) / sheetStretch, dot(p, along) / alongStretch);
}

fn gsl_stone_statuarioNetwork_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return p + (noise_gradientLatticeVector(p / 16.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 16.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 3.0;
}

fn gsl_stone_statuarioNetwork_remap_2(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return p + (noise_gradientLatticeVector(p / 3.5, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 3.5 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.65625) / 0.21875 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.08;
}

fn gsl_stone_statuarioNetwork_remap_3(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return p + (noise_gradientLatticeVector(p / 1.3, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 1.3 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.24375) / 0.08125 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 0.3;
}

fn stone_sheetPoint(c: vec2<f32>, zBand: f32, t: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let lo = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(c.x, c.y, zBand))) & vec3<i32>(63), 0).rgb);
  let hi = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(c.x, c.y, zBand + 1.0))) & vec3<i32>(63), 0).rgb);
  let h = lo + (hi - lo) * t;
  return c + vec2<f32>(h.x, h.y);
}

fn stone_sheetEdge_gsl_anon(base: vec2<f32>, zBand: f32, t: f32, uv: vec2<f32>, a: vec3<f32>, o: vec2<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let v = stone_sheetPoint(base + o, zBand, t, noiseTex) - uv;
  let d = dot(v, v);
  return select(a, vec3<f32>(d, v.x, v.y), d < a.x);
}

fn stone_sheetEdge_gsl_anon_2(home: vec2<f32>, zBand: f32, t: f32, uv: vec2<f32>, mr: vec2<f32>, a: vec3<f32>, o: vec2<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let r = stone_sheetPoint(home + o, zBand, t, noiseTex) - uv;
  let gap = r - mr;
  let toWall = dot(0.5 * (mr + r), normalize(gap + vec2<f32>(0.000001, 0.0)));
  return select(a, vec3<f32>(toWall, r.x, r.y), dot(gap, gap) > 0.00001 && toWall < a.x);
}

fn stone_sheetEdge(q: vec3<f32>, size: f32, depth: f32, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  let uv = vec2<f32>(q.x, q.z) / size;
  let base = floor(uv);
  let band = q.y / depth;
  let zBand = floor(band);
  let t = ease_cubic(fract(band), noiseTex);
  var gsl_fold = vec3<f32>(64.0, 0.0, 0.0);
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = stone_sheetEdge_gsl_anon(base, zBand, t, uv, gsl_fold, gsl_foldK_2[gsl_i], noiseTex);
  }
  let nearest = gsl_fold;
  let mr = vec2<f32>(nearest.y, nearest.z);
  let home = floor(uv + mr);
  var gsl_fold_2 = vec3<f32>(64.0, 0.0, 0.0);
  for (var gsl_i_2 = 0; gsl_i_2 < 25; gsl_i_2 = gsl_i_2 + 1) {
    gsl_fold_2 = stone_sheetEdge_gsl_anon_2(home, zBand, t, uv, mr, gsl_fold_2, gsl_foldK_3[gsl_i_2], noiseTex);
  }
  let edge = gsl_fold_2;
  let pair = home + floor(uv + vec2<f32>(edge.y, edge.z));
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = edge.x * size;
  gsl_result.v_2 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(pair.x, pair.y, 17.0))) & vec3<i32>(63), 0).rgb);
  return gsl_result;
}

fn stone_softLine(gsl_var_distance: f32, halfWidth: f32, soft: f32, noiseTex: texture_3d<f32>) -> f32 {
  let drawn = max(halfWidth, 0.5 * gsl_pw);
  let blur = max(soft, 0.5 * gsl_pw);
  return (1.0 - texture_mask(gsl_var_distance, drawn, blur, noiseTex)) * halfWidth / drawn;
}

fn stone_statuarioNetwork(p: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_3 {
  let folded = gsl_stone_statuarioNetwork_remap(p + (noise_gradientLatticeVector(p / 60.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 60.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 10.8, noiseTex);
  let jagged = gsl_stone_statuarioNetwork_remap_3(gsl_stone_statuarioNetwork_remap_2(folded, noiseTex), noiseTex);
  let gsl_result_3 = stone_sheetEdge(jagged, 36.0, 30.0, noiseTex);
  let wallHash = gsl_result_3.v_2;
  let cracked = select(0.15, 1.0, wallHash.x < 0.7);
  let tone = 0.6 + 0.4 * texture_mask(noise_gradientLattice(p / 18.0, noiseTex), -0.1, 0.25, noiseTex);
  let wall = gsl_result_3.v_1;
  let width = 0.12 + 1.0 * texture_mask(noise_gradientLattice(p / 25.0, noiseTex) + 0.5 * noise_gradientLattice(p / 9.0, noiseTex), 0.05, 0.3, noiseTex);
  let fringe = tone * texture_mask(noise_gradientLattice(p / 2.5, noiseTex), -0.1, 0.25, noiseTex);
  let vein = max(cracked * tone * stone_softLine(wall, width, 0.45 + 0.3 * width, noiseTex), 0.35 * fringe * stone_softLine(wall, 1.5 * width, 0.3 + 0.5 * width, noiseTex));
  let gsl_result_4 = stone_sheetEdge(jagged + vec3<f32>(31.0, 17.0, 5.0), 15.0, 15.0, noiseTex);
  let midHash = gsl_result_4.v_2;
  let midTone = texture_mask(noise_gradientLattice(p / 9.0, noiseTex), -0.15, 0.2, noiseTex);
  let midWall = gsl_result_4.v_1;
  let mid = (0.3 + 0.35 * midHash.z) * select(0.0, 1.0, midHash.x < 0.7) * midTone * stone_softLine(midWall, 0.15 + 0.3 * midHash.y, 0.6, noiseTex);
  let smoke = texture_mask(noise_gradientLattice(p / 7.0, noiseTex), 0.05, 0.2, noiseTex);
  let halo = tone * (0.15 + 0.85 * smoke) * (1.0 - texture_mask(wall, 2.0 * width + 1.0, 1.0 + 1.5 * width, noiseTex)) + 0.5 * tone * smoke * (1.0 - texture_mask(wall, 3.0, 3.0, noiseTex));
  let midHalo = 0.5 * midTone * smoke * (1.0 - texture_mask(midWall, 1.2, 1.2, noiseTex));
  var gsl_result : gsl_struct_3;
  gsl_result.v_1 = max(vein, mid);
  gsl_result.v_2 = max(halo, 0.6 * midHalo);
  return gsl_result;
}

fn stone_statuario(p: vec3<f32>, direction: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  var gsl_simpleValue = vec3<f32>(0.8546301, 0.8554181, 0.86065423);
  if (gsl_simpleDetail > 0.0) {
    let gsl_tmp_145 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), 1.0);
    let gsl_tmp_146 = vec2<f32>(gsl_tmp_145.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_145.y + 0.5);
    let gsl_tmp_147 = vec2<f32>(gsl_tmp_146.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_146.y + 0.25);
    let gsl_tmp_148 = vec2<f32>(gsl_tmp_147.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_147.y + 0.0);
    let gsl_tmp_149 = vec2<f32>(gsl_tmp_148.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_148.y + 0.0);
    let gsl_tmp_150 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), 1.0);
    let gsl_tmp_151 = vec2<f32>(gsl_tmp_150.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_150.y + 0.5);
    let gsl_tmp_152 = vec2<f32>(gsl_tmp_151.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_151.y + 0.25);
    let gsl_tmp_153 = vec2<f32>(gsl_tmp_152.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_152.y + 0.0);
    let gsl_tmp_154 = vec2<f32>(gsl_tmp_153.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_153.y + 0.0);
    let cloud = texture_mask(2.0 * (gsl_tmp_149.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.3515625) / 0.1171875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.9375, noiseTex)) / max(gsl_tmp_154.y + 0.0, 0.000001), -0.05, 0.35, noiseTex);
    let gsl_result_5 = stone_statuarioNetwork(stone_veinFrame(p, direction, 1.1, 1.0, noiseTex), noiseTex);
    let halo = gsl_result_5.v_2;
    let gsl_tmp_155 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.3) / 0.1 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.8, noiseTex), 1.0);
    let gsl_tmp_156 = vec2<f32>(gsl_tmp_155.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.15) / 0.05 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.4, noiseTex), gsl_tmp_155.y + 0.5);
    let gsl_tmp_157 = vec2<f32>(gsl_tmp_156.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.075) / 0.025 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.2, noiseTex), gsl_tmp_156.y + 0.0);
    let gsl_tmp_158 = vec2<f32>(gsl_tmp_157.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.0375) / 0.0125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.1, noiseTex), gsl_tmp_157.y + 0.0);
    let gsl_tmp_159 = vec2<f32>(gsl_tmp_158.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.01875) / 0.00625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.05, noiseTex), gsl_tmp_158.y + 0.0);
    let gsl_tmp_160 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.3) / 0.1 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.8, noiseTex), 1.0);
    let gsl_tmp_161 = vec2<f32>(gsl_tmp_160.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.15) / 0.05 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.4, noiseTex), gsl_tmp_160.y + 0.5);
    let gsl_tmp_162 = vec2<f32>(gsl_tmp_161.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.075) / 0.025 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.2, noiseTex), gsl_tmp_161.y + 0.0);
    let gsl_tmp_163 = vec2<f32>(gsl_tmp_162.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.0375) / 0.0125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.1, noiseTex), gsl_tmp_162.y + 0.0);
    let gsl_tmp_164 = vec2<f32>(gsl_tmp_163.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.01875) / 0.00625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.05, noiseTex), gsl_tmp_163.y + 0.0);
    let grain = 2.0 * (gsl_tmp_159.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.009375) / 0.003125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.025, noiseTex)) / max(gsl_tmp_164.y + 0.0, 0.000001) + (fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.25) + vec3<f32>(9.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.25) + vec3<f32>(9.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * (1.0 - texture_mask(gsl_pw, 0.09375, 0.03125, noiseTex));
    let ground = vec3<f32>(0.9737285, 0.00047567784, -0.0016194935) + vec3<f32>(-0.06 * cloud - 0.11 * halo + 0.012 * grain, 0.0, 0.0);
    let vein = gsl_result_5.v_1;
    gsl_simpleValue = mix(gsl_simpleValue, color_oklab2rgb(ground + (vec3<f32>(0.329599, -0.0021425667, -0.0053033656) - ground) * vein, noiseTex), gsl_simpleDetail);
  }
  let slab = gsl_simpleValue;
  return slab;
}

fn gsl_stone_carraraWisps_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_165 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 6.75) / 2.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 18.0, noiseTex), 1.0);
  let gsl_tmp_166 = vec2<f32>(gsl_tmp_165.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.375) / 1.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 9.0, noiseTex), gsl_tmp_165.y + 0.6);
  let gsl_tmp_167 = vec2<f32>(gsl_tmp_166.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.6875) / 0.5625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 4.5, noiseTex), gsl_tmp_166.y + 0.36);
  let gsl_tmp_168 = vec2<f32>(gsl_tmp_167.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.84375) / 0.28125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.25, noiseTex), gsl_tmp_167.y + 0.0);
  let gsl_tmp_169 = vec2<f32>(gsl_tmp_168.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.421875) / 0.140625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.125, noiseTex), gsl_tmp_168.y + 0.0);
  let gsl_tmp_170 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 6.75) / 2.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 18.0, noiseTex), 1.0);
  let gsl_tmp_171 = vec2<f32>(gsl_tmp_170.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.375) / 1.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 9.0, noiseTex), gsl_tmp_170.y + 0.6);
  let gsl_tmp_172 = vec2<f32>(gsl_tmp_171.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.6875) / 0.5625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 4.5, noiseTex), gsl_tmp_171.y + 0.36);
  let gsl_tmp_173 = vec2<f32>(gsl_tmp_172.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.84375) / 0.28125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.25, noiseTex), gsl_tmp_172.y + 0.0);
  let gsl_tmp_174 = vec2<f32>(gsl_tmp_173.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.421875) / 0.140625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.125, noiseTex), gsl_tmp_173.y + 0.0);
  let gsl_tmp_175 = clamp((2.0 * (gsl_tmp_169.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.2109375) / 0.0703125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.5625, noiseTex)) / max(gsl_tmp_174.y + 0.0, 0.000001) - 0.1) / 0.35 * 0.5 + 0.5, 0.0, 1.0);
  return gsl_tmp_175 * gsl_tmp_175 * (3.0 - 2.0 * gsl_tmp_175);
}

fn gsl_stone_carraraWisps_remap_2(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return p + (noise_gradientLatticeVector(p / 2.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 2.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 0.72;
}

fn gsl_stone_carraraWisps_remap_3(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_176 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.3125) / 0.4375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.5, noiseTex), 1.0);
  let gsl_tmp_177 = vec2<f32>(gsl_tmp_176.x + 0.55 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.65625) / 0.21875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.75, noiseTex), gsl_tmp_176.y + 0.55);
  let gsl_tmp_178 = vec2<f32>(gsl_tmp_177.x + 0.3025 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.328125) / 0.109375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.875, noiseTex), gsl_tmp_177.y + 0.3025);
  let gsl_tmp_179 = vec2<f32>(gsl_tmp_178.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1640625) / 0.0546875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.4375, noiseTex), gsl_tmp_178.y + 0.0);
  let gsl_tmp_180 = vec2<f32>(gsl_tmp_179.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.08203125) / 0.02734375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.21875, noiseTex), gsl_tmp_179.y + 0.0);
  let gsl_tmp_181 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.3125) / 0.4375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.5, noiseTex), 1.0);
  let gsl_tmp_182 = vec2<f32>(gsl_tmp_181.x + 0.55 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.65625) / 0.21875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.75, noiseTex), gsl_tmp_181.y + 0.55);
  let gsl_tmp_183 = vec2<f32>(gsl_tmp_182.x + 0.3025 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.328125) / 0.109375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.875, noiseTex), gsl_tmp_182.y + 0.3025);
  let gsl_tmp_184 = vec2<f32>(gsl_tmp_183.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1640625) / 0.0546875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.4375, noiseTex), gsl_tmp_183.y + 0.0);
  let gsl_tmp_185 = vec2<f32>(gsl_tmp_184.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.08203125) / 0.02734375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.21875, noiseTex), gsl_tmp_184.y + 0.0);
  let gsl_tmp_186 = clamp((2.0 * (gsl_tmp_180.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.041015625) / 0.013671875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.109375, noiseTex)) / max(gsl_tmp_185.y + 0.0, 0.000001) - 0.34) / 0.25 * 0.5 + 0.5, 0.0, 1.0);
  return gsl_tmp_186 * gsl_tmp_186 * (3.0 - 2.0 * gsl_tmp_186);
}

fn stone_carraraWisps(p: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_3 {
  let w = p + (noise_gradientLatticeVector(p / 30.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 30.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 7.2;
  let zone = gsl_stone_carraraWisps_remap(w, noiseTex);
  let feathered = gsl_stone_carraraWisps_remap_2(w, noiseTex);
  let dash = gsl_stone_carraraWisps_remap_3(vec3<f32>(feathered.x, feathered.y, feathered.z / 2.5), noiseTex);
  var gsl_result : gsl_struct_3;
  gsl_result.v_1 = zone;
  gsl_result.v_2 = dash;
  return gsl_result;
}

fn stone_carrara(p: vec3<f32>, direction: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  var gsl_simpleValue = vec3<f32>(0.93127376, 0.9332723, 0.9376854);
  if (gsl_simpleDetail > 0.0) {
    let gsl_tmp_187 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 6.0) / 2.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 16.0, noiseTex), 1.0);
    let gsl_tmp_188 = vec2<f32>(gsl_tmp_187.x + 0.45 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 8.0, noiseTex), gsl_tmp_187.y + 0.45);
    let gsl_tmp_189 = vec2<f32>(gsl_tmp_188.x + 0.2025 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.5) / 0.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 4.0, noiseTex), gsl_tmp_188.y + 0.2025);
    let gsl_tmp_190 = vec2<f32>(gsl_tmp_189.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.75) / 0.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.0, noiseTex), gsl_tmp_189.y + 0.0);
    let gsl_tmp_191 = vec2<f32>(gsl_tmp_190.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.0, noiseTex), gsl_tmp_190.y + 0.0);
    let gsl_tmp_192 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 6.0) / 2.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 16.0, noiseTex), 1.0);
    let gsl_tmp_193 = vec2<f32>(gsl_tmp_192.x + 0.45 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 8.0, noiseTex), gsl_tmp_192.y + 0.45);
    let gsl_tmp_194 = vec2<f32>(gsl_tmp_193.x + 0.2025 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.5) / 0.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 4.0, noiseTex), gsl_tmp_193.y + 0.2025);
    let gsl_tmp_195 = vec2<f32>(gsl_tmp_194.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.75) / 0.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.0, noiseTex), gsl_tmp_194.y + 0.0);
    let gsl_tmp_196 = vec2<f32>(gsl_tmp_195.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.0, noiseTex), gsl_tmp_195.y + 0.0);
    let cloud = 2.0 * (gsl_tmp_191.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1875) / 0.0625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.5, noiseTex)) / max(gsl_tmp_196.y + 0.0, 0.000001);
    let gsl_tmp_197 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.0, noiseTex), 1.0);
    let gsl_tmp_198 = vec2<f32>(gsl_tmp_197.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1875) / 0.0625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.5, noiseTex), gsl_tmp_197.y + 0.6);
    let gsl_tmp_199 = vec2<f32>(gsl_tmp_198.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.09375) / 0.03125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.25, noiseTex), gsl_tmp_198.y + 0.0);
    let gsl_tmp_200 = vec2<f32>(gsl_tmp_199.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.046875) / 0.015625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.125, noiseTex), gsl_tmp_199.y + 0.0);
    let gsl_tmp_201 = vec2<f32>(gsl_tmp_200.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.0234375) / 0.0078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.0625, noiseTex), gsl_tmp_200.y + 0.0);
    let gsl_tmp_202 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.0, noiseTex), 1.0);
    let gsl_tmp_203 = vec2<f32>(gsl_tmp_202.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1875) / 0.0625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.5, noiseTex), gsl_tmp_202.y + 0.6);
    let gsl_tmp_204 = vec2<f32>(gsl_tmp_203.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.09375) / 0.03125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.25, noiseTex), gsl_tmp_203.y + 0.0);
    let gsl_tmp_205 = vec2<f32>(gsl_tmp_204.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.046875) / 0.015625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.125, noiseTex), gsl_tmp_204.y + 0.0);
    let gsl_tmp_206 = vec2<f32>(gsl_tmp_205.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.0234375) / 0.0078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.0625, noiseTex), gsl_tmp_205.y + 0.0);
    let grain = 0.3 * 2.0 * (gsl_tmp_201.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.01171875) / 0.00390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.03125, noiseTex)) / max(gsl_tmp_206.y + 0.0, 0.000001) + (fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.3) + vec3<f32>(9.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.3) + vec3<f32>(9.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 1.3 * (1.0 - texture_mask(gsl_pw, 0.1125, 0.0375, noiseTex));
    let gsl_result_6 = stone_carraraWisps(stone_veinFrame(p, direction, 1.4, 4.0, noiseTex), noiseTex);
    let zone = gsl_result_6.v_1;
    let dash = gsl_result_6.v_2;
    let ground = vec3<f32>(0.96706367, -0.00026034258, -0.00008032342) + vec3<f32>(0.035 * cloud + 0.03 * grain - 0.04 * zone - 0.13 * zone * zone * dash, 0.0005 * zone, -0.003 * zone - 0.004 * dash);
    gsl_simpleValue = mix(gsl_simpleValue, color_oklab2rgb(ground, noiseTex), gsl_simpleDetail);
  }
  let slab = gsl_simpleValue;
  return slab;
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
  if (gsl_colorIdx == 2u) { return vec4<f32>(stone_statuario(p, vec3<f32>(1.0, 0.6, 0.0), noiseTex), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(stone_carrara(p, vec3<f32>(1.0, 0.6, 0.0), noiseTex), 1.0); }
  if (gsl_colorIdx == 4u) {
    var gsl_detailValue = vec3<f32>(0.5565874, 0.542274, 0.53184134);
    if (max(gsl_pw, gsl_detailPw) < 1.2) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_70 = p + (noise_gradientLatticeVector(p / 5.2, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 5.2 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.975) / 0.325 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.056;
      let gsl_tmp_71 = gsl_tmp_70 / 4.0;
      let gsl_tmp_74 = gsl_cellularScan(gsl_tmp_71, 1, noiseTex).nearCell;
      let gsl_tmp_94 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_74 + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_74 + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_98 = noise_gradientLattice(p / 1.2 + vec3<f32>(1.7, 9.1, 4.4), noiseTex) + 0.15 * (1.0 - texture_mask(gsl_pw, 0.18, 0.06, noiseTex)) * noise_gradientLattice(p / 0.48 + vec3<f32>(8.2, 2.6, 7.3), noiseTex) + 0.4 * noise_gradientLattice(p / 3.6 + vec3<f32>(4.2, 6.6, 1.3), noiseTex);
      let gsl_tmp_96 = normalize(gsl_tmp_94 - 0.5 + vec3<f32>(0.0001, 0.0, 0.0));
      let gsl_tmp_97 = dot(p, gsl_tmp_96);
      let gsl_tmp_73 = gsl_cellularScan(gsl_tmp_71, 1, noiseTex).f2;
      let gsl_tmp_72 = gsl_cellularScan(gsl_tmp_71, 1, noiseTex).f1;
      let gsl_tmp_75 = gsl_cellularScan(gsl_tmp_71, 1, noiseTex).farCell;
      let gsl_tmp_100 = 1.0 - texture_mask((gsl_tmp_73 * gsl_tmp_73 - gsl_tmp_72 * gsl_tmp_72) / (2.0 * max(length(gsl_tmp_75 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_75)) & vec3<i32>(63), 0).rgb) - gsl_tmp_74 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_74)) & vec3<i32>(63), 0).rgb)), 0.001)) * 4.0, 0.24, 0.2, noiseTex);
      let gsl_tmp_99 = texture_fadeTo(abs(noise_gradientLattice(p / 0.6 + vec3<f32>(6.1, 4.9, 2.2), noiseTex)), 0.157, 0.6, noiseTex) + 0.5 * texture_fadeTo(abs(noise_gradientLattice(p / 0.24 + vec3<f32>(1.1, 3.5, 8.8), noiseTex)), 0.157, 0.24, noiseTex);
      let gsl_tmp_101 = vec3<f32>(0.62754637, 0.0047757193, 0.0054191384) + vec3<f32>((gsl_tmp_94.y - 0.5) * 0.1 + 0.1 * gsl_tmp_98 + 0.015 * noise_gradientLattice((p - gsl_tmp_96 * gsl_tmp_97) / 0.32 + gsl_tmp_96 * gsl_tmp_97 / 3.2, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.12, 0.04, noiseTex)) + 0.04 * gsl_tmp_100 + 0.12 * (gsl_tmp_99 - 0.235) * (0.5 + gsl_tmp_100), 0.0, 0.0) - vec3<f32>(0.0, 0.0047757193, 0.0054191384) * 0.3 * gsl_tmp_100;
      let gsl_tmp_76 = (gsl_tmp_70 + p + (noise_gradientLatticeVector(p / 2.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 2.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 0.384 - p + vec3<f32>(148.4, 45.2, 23.6)) / 1.8;
      let gsl_tmp_79 = gsl_cellularScan(gsl_tmp_76, 1, noiseTex).nearCell;
      let gsl_tmp_82 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_79 + vec3<f32>(7.0, 29.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_79 + vec3<f32>(7.0, 29.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_84 = stone_quartzOf(gsl_tmp_79, 1.8, 4.0, noiseTex);
      let gsl_tmp_93 = select(0.0, 1.0 - gsl_tmp_84, gsl_tmp_82.x < 0.15);
      let gsl_tmp_103 = gsl_tmp_101 + (vec3<f32>(0.62754637, 0.0047757193, 0.0054191384) + vec3<f32>(0.06 + (gsl_tmp_82.y - 0.5) * 0.06 + 0.04 * gsl_tmp_98 + 0.12 * (gsl_tmp_99 - 0.235), -0.0023878596, -0.0027095692) - gsl_tmp_101) * gsl_tmp_93;
      let gsl_tmp_77 = gsl_cellularScan(gsl_tmp_76, 1, noiseTex).f1;
      let gsl_tmp_102 = 1.0 - texture_mask(gsl_tmp_77, 0.3, 0.25, noiseTex);
      let gsl_tmp_78 = gsl_cellularScan(gsl_tmp_76, 1, noiseTex).f2;
      let gsl_tmp_80 = gsl_cellularScan(gsl_tmp_76, 1, noiseTex).farCell;
      let gsl_tmp_81 = (gsl_tmp_78 * gsl_tmp_78 - gsl_tmp_77 * gsl_tmp_77) / (2.0 * max(length(gsl_tmp_80 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_80)) & vec3<i32>(63), 0).rgb) - gsl_tmp_79 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_79)) & vec3<i32>(63), 0).rgb)), 0.001));
      let gsl_tmp_83 = 2.0 * texture_mask(gsl_tmp_81 * 1.8, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_tmp_91 = gsl_tmp_84 * (0.5 + 0.5 * gsl_tmp_83) + stone_quartzOf(gsl_tmp_80, 1.8, 4.0, noiseTex) * (0.5 - 0.5 * gsl_tmp_83);
      let gsl_tmp_104 = gsl_tmp_103 + (vec3<f32>(0.7039235, 0.0019268822, 0.00927625) + vec3<f32>((gsl_tmp_82.z - 0.5) * 0.12 - 0.1 * gsl_tmp_102 + 0.05 * dot(gsl_tmp_76 - gsl_tmp_79 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_79)) & vec3<i32>(63), 0).rgb), normalize(gsl_tmp_82 - 0.5 + vec3<f32>(0.0, 0.0001, 0.0))) + 0.01 * (1.0 - texture_mask(gsl_tmp_81 * 1.8, 0.08, 0.06, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.075, 0.025, noiseTex)) + 0.02 * gsl_tmp_98 + 0.06 * (gsl_tmp_99 - 0.235), -0.003 * gsl_tmp_102, -0.01 * gsl_tmp_102 + (gsl_tmp_82.y - 0.5) * 0.02) - gsl_tmp_103) * gsl_tmp_91 + vec3<f32>(0.15 * noise_gradientLattice(p / 12.0 + vec3<f32>(2.9, 8.3, 6.1), noiseTex) + 0.08 * noise_gradientLattice(p / 5.6 + vec3<f32>(9.4, 1.2, 3.8), noiseTex), 0.0, 0.0);
      let gsl_tmp_85 = (gsl_tmp_70 + vec3<f32>(21.2, 246.8, 68.4)) / 0.6;
      let gsl_tmp_88 = gsl_cellularScan(gsl_tmp_85, 1, noiseTex).nearCell;
      let gsl_tmp_95 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_88 + vec3<f32>(3.0, 0.0, 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_88 + vec3<f32>(3.0, 0.0, 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_87 = gsl_cellularScan(gsl_tmp_85, 1, noiseTex).f2;
      let gsl_tmp_86 = gsl_cellularScan(gsl_tmp_85, 1, noiseTex).f1;
      let gsl_tmp_89 = gsl_cellularScan(gsl_tmp_85, 1, noiseTex).farCell;
      let gsl_tmp_90 = 2.0 * texture_mask((gsl_tmp_87 * gsl_tmp_87 - gsl_tmp_86 * gsl_tmp_86) / (2.0 * max(length(gsl_tmp_89 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_89)) & vec3<i32>(63), 0).rgb) - gsl_tmp_88 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_88)) & vec3<i32>(63), 0).rgb)), 0.001)) * 0.6, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_tmp_92 = stone_maficOf(gsl_tmp_88, 0.6, 4.0, gsl_tmp_84, noiseTex) * (0.5 + 0.5 * gsl_tmp_90) + stone_maficOf(gsl_tmp_89, 0.6, 4.0, gsl_tmp_84, noiseTex) * (0.5 - 0.5 * gsl_tmp_90);
      gsl_detailValue = color_oklab2rgb(vec3<f32>(0.63668054, 0.0030485727, 0.0050342223) + (gsl_tmp_104 + (vec3<f32>(0.31609565, 0.0013598828, -0.004748624) + vec3<f32>((gsl_tmp_95.y - 0.5) * 0.06 + select(0.0, 0.12, gsl_tmp_95.z > 0.96), 0.0, 0.0) + select(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.03, 0.01, 0.02), gsl_tmp_95.z < 0.08) - gsl_tmp_104) * gsl_tmp_92 - vec3<f32>(0.63668054, 0.0030485727, 0.0050342223)) * (1.0 - texture_mask(gsl_pw, 0.9, 0.3, noiseTex)), noiseTex);
      gsl_pw = gsl_detailOuterPw;
    }
    var gsl_simpleValue_2 = vec3<f32>(0.5565874, 0.542274, 0.53184134);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, gsl_detailValue, gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_2, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) {
    var gsl_simpleValue_3 = mat4x4<f32>(vec4<f32>(0.111135, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_105 = p + (noise_gradientLatticeVector(p / 5.2, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 5.2 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.975) / 0.325 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.056;
      let gsl_tmp_120 = (gsl_tmp_105 + vec3<f32>(21.2, 246.8, 68.4)) / 0.6;
      let gsl_tmp_123 = gsl_cellularScan(gsl_tmp_120, 1, noiseTex).nearCell;
      let gsl_tmp_111 = (gsl_tmp_105 + p + (noise_gradientLatticeVector(p / 2.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 2.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 0.384 - p + vec3<f32>(148.4, 45.2, 23.6)) / 1.8;
      let gsl_tmp_114 = gsl_cellularScan(gsl_tmp_111, 1, noiseTex).nearCell;
      let gsl_tmp_119 = stone_quartzOf(gsl_tmp_114, 1.8, 4.0, noiseTex);
      let gsl_tmp_122 = gsl_cellularScan(gsl_tmp_120, 1, noiseTex).f2;
      let gsl_tmp_121 = gsl_cellularScan(gsl_tmp_120, 1, noiseTex).f1;
      let gsl_tmp_124 = gsl_cellularScan(gsl_tmp_120, 1, noiseTex).farCell;
      let gsl_tmp_125 = 2.0 * texture_mask((gsl_tmp_122 * gsl_tmp_122 - gsl_tmp_121 * gsl_tmp_121) / (2.0 * max(length(gsl_tmp_124 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_124)) & vec3<i32>(63), 0).rgb) - gsl_tmp_123 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_123)) & vec3<i32>(63), 0).rgb)), 0.001)) * 0.6, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_tmp_126 = stone_maficOf(gsl_tmp_123, 0.6, 4.0, gsl_tmp_119, noiseTex) * (0.5 + 0.5 * gsl_tmp_125) + stone_maficOf(gsl_tmp_124, 0.6, 4.0, gsl_tmp_119, noiseTex) * (0.5 - 0.5 * gsl_tmp_125);
      let gsl_tmp_113 = gsl_cellularScan(gsl_tmp_111, 1, noiseTex).f2;
      let gsl_tmp_112 = gsl_cellularScan(gsl_tmp_111, 1, noiseTex).f1;
      let gsl_tmp_115 = gsl_cellularScan(gsl_tmp_111, 1, noiseTex).farCell;
      let gsl_tmp_116 = (gsl_tmp_113 * gsl_tmp_113 - gsl_tmp_112 * gsl_tmp_112) / (2.0 * max(length(gsl_tmp_115 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_115)) & vec3<i32>(63), 0).rgb) - gsl_tmp_114 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_114)) & vec3<i32>(63), 0).rgb)), 0.001));
      let gsl_tmp_118 = 2.0 * texture_mask(gsl_tmp_116 * 1.8, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_simpleFull = stone_stoneFinish(0.15 * (1.0 - 0.67 * (1.0 - gsl_tmp_126) * (gsl_tmp_119 * (0.5 + 0.5 * gsl_tmp_118) + stone_quartzOf(gsl_tmp_115, 1.8, 4.0, noiseTex) * (0.5 - 0.5 * gsl_tmp_118)) - 0.5 * gsl_tmp_126), noiseTex);
      gsl_simpleValue_3 = mat4x4<f32>(mix(gsl_simpleValue_3[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_3[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_3[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_3[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let gsl_m = gsl_simpleValue_3; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

const gsl_foldK_2 = array<vec2<f32>, 9>(vec2<f32>(-1.0, -1.0), vec2<f32>(0.0, -1.0), vec2<f32>(1.0, -1.0), vec2<f32>(-1.0, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(1.0, 0.0), vec2<f32>(-1.0, 1.0), vec2<f32>(0.0, 1.0), vec2<f32>(1.0, 1.0));

const gsl_foldK_3 = array<vec2<f32>, 25>(vec2<f32>(-2.0, -2.0), vec2<f32>(-1.0, -2.0), vec2<f32>(0.0, -2.0), vec2<f32>(1.0, -2.0), vec2<f32>(2.0, -2.0), vec2<f32>(-2.0, -1.0), vec2<f32>(-1.0, -1.0), vec2<f32>(0.0, -1.0), vec2<f32>(1.0, -1.0), vec2<f32>(2.0, -1.0), vec2<f32>(-2.0, 0.0), vec2<f32>(-1.0, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(1.0, 0.0), vec2<f32>(2.0, 0.0), vec2<f32>(-2.0, 1.0), vec2<f32>(-1.0, 1.0), vec2<f32>(0.0, 1.0), vec2<f32>(1.0, 1.0), vec2<f32>(2.0, 1.0), vec2<f32>(-2.0, 2.0), vec2<f32>(-1.0, 2.0), vec2<f32>(0.0, 2.0), vec2<f32>(1.0, 2.0), vec2<f32>(2.0, 2.0));

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
  v_4 : vec3<f32>,
}
struct gsl_struct_2 {
  v_1 : f32,
  v_2 : vec3<f32>,
}
struct gsl_struct_3 {
  v_1 : f32,
  v_2 : f32,
}
