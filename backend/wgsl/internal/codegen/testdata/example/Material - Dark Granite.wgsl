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

fn builtin_sphere(p: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - radius;
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

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn noise_octaveFade(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - ease_quadratic(clamp((gsl_pw - 0.375 * size) / (0.125 * size) * 0.5 + 0.5, 0.0, 1.0), noiseTex);
}

fn noise_warp(p: vec3<f32>, size: f32, amount: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let coarse = noise_gradientLatticeVector(p / size, noiseTex);
  let fine = noise_gradientLatticeVector(p * 2.03 / size + vec3<f32>(17.3, 5.1, 9.7), noiseTex);
  return p + (coarse + fine * 0.5 * noise_octaveFade(0.5 * size, noiseTex)) * amount * 1.2;
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

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
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

fn texture_reliefRoughness(roughness: f32, slope: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sqrt(roughness * roughness + (1.0 - texture_fadeDetail(size, noiseTex) * gsl_dispDrawn) * slope * slope);
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn stone_grainAt(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  let f1 = gsl_cellularScan(q, 1, noiseTex).f1;
  let f2 = gsl_cellularScan(q, 1, noiseTex).f2;
  let far = gsl_cellularScan(q, 1, noiseTex).farCell;
  let near = gsl_cellularScan(q, 1, noiseTex).nearCell;
  let gap = length(far + fract(textureLoad(noiseTex, vec3<i32>(floor(far)) & vec3<i32>(63), 0).rgb) - near - fract(textureLoad(noiseTex, vec3<i32>(floor(near)) & vec3<i32>(63), 0).rgb));
  var gsl_result : gsl_struct;
  gsl_result.v_1 = f1;
  gsl_result.v_2 = (f2 * f2 - f1 * f1) / (2.0 * max(gap, 0.001));
  gsl_result.v_3 = near;
  gsl_result.v_4 = far;
  return gsl_result;
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

fn stone_graniteGrains(p: vec3<f32>, s: f32, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  let w = noise_warp(p, 1.3 * s, 0.22 * s, noiseTex);
  let g = 0.45 * s;
  let sq = (w + noise_warp(p, 0.5 * s, 0.08 * s, noiseTex) - p + vec3<f32>(37.1, 11.3, 5.9) * s) / g;
  let gsl_result_2 = stone_grainAt(sq, noiseTex);
  let small = gsl_result_2.v_3;
  let sf1 = gsl_result_2.v_1;
  let swall = gsl_result_2.v_2;
  let sh = fract(textureLoad(noiseTex, vec3<i32>(floor(small + vec3<f32>(7.0, 29.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((small + vec3<f32>(7.0, 29.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let quartzIn = stone_quartzOf(small, g, s, noiseTex);
  let sAcross = 2.0 * texture_mask(swall * g, 0.0, 0.05 * s, noiseTex) - 1.0;
  let smallFar = gsl_result_2.v_4;
  let inQuartz = quartzIn * (0.5 + 0.5 * sAcross) + stone_quartzOf(smallFar, g, s, noiseTex) * (0.5 - 0.5 * sAcross);
  let t = 0.15 * s;
  let gsl_result_3 = stone_grainAt((w + vec3<f32>(5.3, 61.7, 17.1) * s) / t, noiseTex);
  let tiny = gsl_result_3.v_3;
  let twall = gsl_result_3.v_2;
  let tAcross = 2.0 * texture_mask(twall * t, 0.0, 0.05 * s, noiseTex) - 1.0;
  let tinyFar = gsl_result_3.v_4;
  let mafic = stone_maficOf(tiny, t, s, quartzIn, noiseTex) * (0.5 + 0.5 * tAcross) + stone_maficOf(tinyFar, t, s, quartzIn, noiseTex) * (0.5 - 0.5 * tAcross);
  let plag = select(0.0, 1.0 - quartzIn, sh.x < 0.15);
  let gsl_result = stone_grainAt(w / s, noiseTex);
  let bwall = gsl_result.v_2;
  let big = gsl_result.v_3;
  let bh = fract(textureLoad(noiseTex, vec3<i32>(floor(big + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((big + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let th = fract(textureLoad(noiseTex, vec3<i32>(floor(tiny + vec3<f32>(3.0, 0.0, 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((tiny + vec3<f32>(3.0, 0.0, 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  var gsl_result_4 : gsl_struct_2;
  gsl_result_4.v_1 = sq;
  gsl_result_4.v_2 = small;
  gsl_result_4.v_3 = sf1;
  gsl_result_4.v_4 = swall * g;
  gsl_result_4.v_5 = sh;
  gsl_result_4.v_6 = inQuartz;
  gsl_result_4.v_7 = mafic;
  gsl_result_4.v_8 = plag;
  gsl_result_4.v_9 = bwall * s;
  gsl_result_4.v_10 = bh;
  gsl_result_4.v_11 = th;
  return gsl_result_4;
}

fn stone_darkFloor(L: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 0.5 * (L + 0.3 + sqrt((L - 0.3) * (L - 0.3) + 0.0064));
}

fn stone_graniteMosaic(p: vec3<f32>, s: f32, baseColor: vec3<f32>, fleckColorA: vec3<f32>, fleckColorB: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_3 {
  let feldLab = color_rgb2oklab(baseColor, noiseTex);
  let gsl_result_4 = stone_graniteGrains(p, s, noiseTex);
  let bh = gsl_result_4.v_10;
  let cloud = noise_gradientLattice(p / (0.3 * s) + vec3<f32>(1.7, 9.1, 4.4), noiseTex) + 0.15 * texture_fadeDetail(0.12 * s, noiseTex) * noise_gradientLattice(p / (0.12 * s) + vec3<f32>(8.2, 2.6, 7.3), noiseTex) + 0.4 * noise_gradientLattice(p / (0.9 * s) + vec3<f32>(4.2, 6.6, 1.3), noiseTex);
  let dir = normalize(bh - 0.5 + vec3<f32>(0.0001, 0.0, 0.0));
  let along = dot(p, dir);
  let streak = noise_gradientLattice((p - dir * along) / (0.08 * s) + dir * along / (0.8 * s), noiseTex) * texture_fadeDetail(0.08 * s, noiseTex);
  let bwall = gsl_result_4.v_9;
  let rim = 1.0 - texture_mask(bwall, 0.06 * s, 0.05 * s, noiseTex);
  let frost = texture_fadeTo(abs(noise_gradientLattice(p / (0.15 * s) + vec3<f32>(6.1, 4.9, 2.2), noiseTex)), 0.157, 0.15 * s, noiseTex) + 0.5 * texture_fadeTo(abs(noise_gradientLattice(p / (0.06 * s) + vec3<f32>(1.1, 3.5, 8.8), noiseTex)), 0.157, 0.06 * s, noiseTex);
  let feldspar = feldLab + vec3<f32>((bh.y - 0.5) * 0.1 + 0.1 * cloud + 0.015 * streak + 0.04 * rim + 0.12 * (frost - 0.235) * (0.5 + rim), 0.0, 0.0) - vec3<f32>(0.0, feldLab.y, feldLab.z) * 0.3 * rim;
  let sh = gsl_result_4.v_5;
  let plagLab = feldLab + vec3<f32>(0.06 + (sh.y - 0.5) * 0.06 + 0.04 * cloud + 0.12 * (frost - 0.235), -0.5 * feldLab.y, -0.5 * feldLab.z);
  let plag = gsl_result_4.v_8;
  let ground = feldspar + (plagLab - feldspar) * plag;
  let quartzLab = color_rgb2oklab(fleckColorB, noiseTex);
  let glass = vec3<f32>(quartzLab.x + (0.5 - quartzLab.x) * 0.45, quartzLab.y * 0.8, quartzLab.z * 0.8);
  let sf1 = gsl_result_4.v_3;
  let core = 1.0 - texture_mask(sf1, 0.3, 0.25, noiseTex);
  let sq = gsl_result_4.v_1;
  let small = gsl_result_4.v_2;
  let lean = dot(sq - small - fract(textureLoad(noiseTex, vec3<i32>(floor(small)) & vec3<i32>(63), 0).rgb), normalize(sh - 0.5 + vec3<f32>(0.0, 0.0001, 0.0)));
  let swall = gsl_result_4.v_4;
  let crack = (1.0 - texture_mask(swall, 0.02 * s, 0.015 * s, noiseTex)) * texture_fadeDetail(0.05 * s, noiseTex);
  let quartz = glass + vec3<f32>((sh.z - 0.5) * 0.12 - 0.1 * core + 0.05 * lean + 0.01 * crack + 0.02 * cloud + 0.06 * (frost - 0.235), -0.003 * core, -0.01 * core + (sh.y - 0.5) * 0.02);
  let inQuartz = gsl_result_4.v_6;
  let broad = 0.15 * noise_gradientLattice(p / (3.0 * s) + vec3<f32>(2.9, 8.3, 6.1), noiseTex) + 0.08 * noise_gradientLattice(p / (1.4 * s) + vec3<f32>(9.4, 1.2, 3.8), noiseTex);
  let felsic = ground + (quartz - ground) * inQuartz + vec3<f32>(broad, 0.0, 0.0);
  let darkLab = color_rgb2oklab(fleckColorA, noiseTex);
  let th = gsl_result_4.v_11;
  let dark = vec3<f32>(stone_darkFloor(darkLab.x, noiseTex), darkLab.y, darkLab.z) + vec3<f32>((th.y - 0.5) * 0.06 + select(0.0, 0.12, th.z > 0.96), 0.0, 0.0) + select(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.03, 0.01, 0.02), th.z < 0.08);
  let mafic = gsl_result_4.v_7;
  let crystal = felsic + (dark - felsic) * mafic;
  let wPlag = (1.0 - mafic) * (1.0 - inQuartz) * plag;
  let wQuartz = (1.0 - mafic) * inQuartz;
  var gsl_result : gsl_struct_3;
  gsl_result.v_1 = crystal;
  gsl_result.v_2 = (1.0 - mafic) * (1.0 - inQuartz) * (1.0 - plag);
  gsl_result.v_3 = wPlag;
  gsl_result.v_4 = wQuartz;
  gsl_result.v_5 = mafic;
  return gsl_result;
}

fn stone_graniteAverage(baseColor: vec3<f32>, fleckColorA: vec3<f32>, fleckColorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let feldLab = color_rgb2oklab(baseColor, noiseTex);
  let quartzLab = color_rgb2oklab(fleckColorB, noiseTex);
  let glass = vec3<f32>(quartzLab.x + (0.5 - quartzLab.x) * 0.45, quartzLab.y * 0.8, quartzLab.z * 0.8);
  let darkLab = color_rgb2oklab(fleckColorA, noiseTex);
  return feldLab * 0.507 + (feldLab - vec3<f32>(0.0, feldLab.y, feldLab.z) * 0.5) * 0.087 + glass * 0.33 + vec3<f32>(stone_darkFloor(darkLab.x, noiseTex), darkLab.y, darkLab.z) * 0.076 - vec3<f32>(0.0, feldLab.y, feldLab.z) * 0.046 + vec3<f32>(0.0076, -0.0001, -0.0004);
}

fn stone_granite(p: vec3<f32>, cellSize: f32, baseColor: vec3<f32>, fleckColorA: vec3<f32>, fleckColorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let average = stone_graniteAverage(baseColor, fleckColorA, fleckColorB, noiseTex);
  var gsl_detailValue = color_oklab2rgb(average, noiseTex);
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (0.6 * cellSize)) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let gsl_result_5 = stone_graniteMosaic(p, cellSize, baseColor, fleckColorA, fleckColorB, noiseTex);
    let crystal = gsl_result_5.v_1;
    gsl_detailValue = color_oklab2rgb(average + (crystal - average) * texture_fadeDetail(0.6 * cellSize, noiseTex), noiseTex);
    gsl_pw = gsl_detailOuterPw;
  }
  var gsl_simpleValue = color_oklab2rgb(average, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue = mix(gsl_simpleValue, gsl_detailValue, gsl_simpleDetail);
  }
  let crystals = gsl_simpleValue;
  return crystals;
}

fn stone_stoneFinish(rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_make(rough, 0.0, 1.55, noiseTex);
}

fn stone_graniteFacetFinish(p: vec3<f32>, cellSize: f32, rough: f32, tilt: f32, gleam: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  var gsl_simpleValue = material_make(rough * 0.632 + 0.33 * sqrt(sqrt(0.0016 * rough * rough * rough * rough + 0.666 * tilt * tilt)), gleam * 0.33, 1.55, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let gsl_result_6 = stone_graniteGrains(p, cellSize, noiseTex);
    let mafic = gsl_result_6.v_7;
    let base = stone_stoneFinish(rough * (1.0 - 0.5 * mafic), noiseTex);
    let g = 0.45 * cellSize;
    let glassy = sqrt(texture_reliefRoughness(rough * rough * 0.04, 0.816 * tilt, g, noiseTex));
    let inQuartz = gsl_result_6.v_6;
    let quartz = (1.0 - mafic) * inQuartz;
    let sq = gsl_result_6.v_1;
    let small = gsl_result_6.v_2;
    let facet = (fract(textureLoad(noiseTex, vec3<i32>(floor(small + vec3<f32>(17.0, 53.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((small + vec3<f32>(17.0, 53.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb) - 0.5) * 2.0;
    let lean = dot(sq - small - fract(textureLoad(noiseTex, vec3<i32>(floor(small)) & vec3<i32>(63), 0).rgb), facet) * g * tilt;
    let gsl_simpleFull = material_withDisplacement(material_blend(base, material_make(glassy, gleam, 1.55, noiseTex), quartz, noiseTex), quartz * lean * texture_fadeDetail(g, noiseTex), noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let facets = gsl_simpleValue;
  return facets;
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-320.01, -200.01, -0.01);
  let gsl_partHi = vec3<f32>(80.01, 200.01, 20.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered(p, vec3<f32>(-120.0, 0.0, 10.0), vec3<f32>(400.0, 400.0, 20.0), noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let graniteSlab = gsl_part;
  let gsl_partLo_2 = vec3<f32>(125.0, -75.0, 0.0);
  let gsl_partHi_2 = vec3<f32>(275.0, 75.0, 150.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_sphere(p - vec3<f32>(200.0, 0.0, 75.0), 75.0, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let graniteBall = gsl_part_2;
  let gsl_s1 = gsl_shape(graniteSlab, 1u);
  let gsl_s2 = gsl_shape(graniteBall, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(stone_granite(p, 7.0, vec3<f32>(0.2, 0.19, 0.19), vec3<f32>(0.04, 0.04, 0.05), vec3<f32>(0.66, 0.67, 0.7), noiseTex), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(stone_granite(p, 7.0, vec3<f32>(0.2, 0.19, 0.19), vec3<f32>(0.04, 0.04, 0.05), vec3<f32>(0.66, 0.67, 0.7), noiseTex), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = stone_graniteFacetFinish(p, 7.0, 0.12, 0.05, 0.35, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = stone_graniteFacetFinish(p, 7.0, 0.12, 0.05, 0.35, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
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

var<private> gsl_dispDrawn: f32 = 0.0;

var<private> gsl_detailPw: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
  v_4 : vec3<f32>,
}
struct gsl_struct_2 {
  v_1 : vec3<f32>,
  v_2 : vec3<f32>,
  v_3 : f32,
  v_4 : f32,
  v_5 : vec3<f32>,
  v_6 : f32,
  v_7 : f32,
  v_8 : f32,
  v_9 : f32,
  v_10 : vec3<f32>,
  v_11 : vec3<f32>,
}
struct gsl_struct_3 {
  v_1 : vec3<f32>,
  v_2 : f32,
  v_3 : f32,
  v_4 : f32,
  v_5 : f32,
}
