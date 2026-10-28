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

fn noise_gradientBump(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / size, noiseTex) * size / 4.0;
}

fn noise_fbmBump4(p: vec3<f32>, size: f32, gain: f32, lacunarity: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientBump(p, size, noiseTex) + gain * noise_gradientBump(p, size / lacunarity, noiseTex) + gain * gain * noise_gradientBump(p, size / (lacunarity * lacunarity), noiseTex) + gain * gain * gain * noise_gradientBump(p, size / (lacunarity * lacunarity * lacunarity), noiseTex);
}

fn noise_turbulenceBump2(p: vec3<f32>, size: f32, gain: f32, lacunarity: f32, noiseTex: texture_3d<f32>) -> f32 {
  return abs(noise_gradientBump(p, size, noiseTex)) + gain * abs(noise_gradientBump(p, size / lacunarity, noiseTex));
}

fn noise_gradient(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / size, noiseTex);
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn noise_cellularF1F2CellLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  var gsl_result : gsl_struct;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  gsl_result.v_3 = gsl_cellularScan(q, 1, noiseTex).nearCell;
  return gsl_result;
}

fn noise_cellularCell(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let gsl_result = noise_cellularF1F2CellLattice(p / size, noiseTex);
  let f1 = gsl_result.v_1;
  let f2 = gsl_result.v_2;
  let cell = gsl_result.v_3;
  var gsl_result_2 : gsl_struct;
  gsl_result_2.v_1 = f1 * size;
  gsl_result_2.v_2 = f2 * size;
  gsl_result_2.v_3 = cell;
  return gsl_result_2;
}

fn color_srgbToLinearChannel(s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x, noiseTex), color_srgbToLinearChannel(c.y, noiseTex), color_srgbToLinearChannel(c.z, noiseTex));
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

fn stone_granite(p: vec3<f32>, cellSize: f32, baseColor: vec3<f32>, fleckColorA: vec3<f32>, fleckColorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let quartzLab = color_rgb2oklab(fleckColorB, noiseTex);
  let quartzMean = quartzLab + (vec3<f32>(0.5298521, 0.0021907585, -0.0075764153) - quartzLab) * 0.45 + vec3<f32>(-0.03096, -0.000832, -0.002496);
  let average = color_rgb2oklab(baseColor, noiseTex) * 0.55 + quartzMean * 0.3 + color_rgb2oklab(fleckColorA, noiseTex) * 0.15 - vec3<f32>(0.0252, 0.0, 0.0);
  var gsl_detailValue = color_oklab2rgb(average, noiseTex);
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (0.65 * cellSize)) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let gsl_result_2 = noise_cellularCell(p, cellSize, noiseTex);
    let cell = gsl_result_2.v_3;
    let pick = fract(textureLoad(noiseTex, vec3<i32>(floor(cell + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let tone = (pick.y - 0.5) * 0.14;
    let feldspar = color_rgb2oklab(baseColor, noiseTex) + vec3<f32>(tone, 0.0, 0.0);
    let f1 = gsl_result_2.v_1;
    let core = 1.0 - texture_mask(f1 / cellSize, 0.35, 0.25, noiseTex);
    let inclusion = select(0.0, 1.0, fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / (0.06 * cellSize)) + vec3<f32>(313.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / (0.06 * cellSize)) + vec3<f32>(313.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x > 0.96) * texture_fadeDetail(0.06 * cellSize, noiseTex);
    let quartz = quartzLab + (vec3<f32>(0.5298521, 0.0021907585, -0.0075764153) - quartzLab) * 0.45 + vec3<f32>(tone * 0.5 - 0.12 * core - 0.15 * inclusion, -0.004 * core, -0.012 * core);
    let dark = color_rgb2oklab(fleckColorA, noiseTex) + vec3<f32>(tone * 0.4, 0.0, 0.0);
    let crystal = select(select(feldspar, quartz, pick.x > 0.55), dark, pick.x > 0.85);
    let resolved = average + (crystal - average) * texture_fadeDetail(0.65 * cellSize, noiseTex);
    let f2 = gsl_result_2.v_2;
    let seam = (1.0 - texture_mask(f2 - f1, 0.05 * cellSize, 0.03 * cellSize, noiseTex)) * texture_fadeDetail(0.1 * cellSize, noiseTex);
    gsl_detailValue = color_oklab2rgb(resolved - vec3<f32>(0.14 * seam, 0.0, 0.0), noiseTex);
    gsl_pw = gsl_detailOuterPw;
  }
  var gsl_simpleValue = color_oklab2rgb(average, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue = mix(gsl_simpleValue, gsl_detailValue, gsl_simpleDetail);
  }
  let crystals = gsl_simpleValue;
  return crystals;
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn texture_paintMaterial(field: f32, edge: f32, width: f32, matA: mat4x4<f32>, matB: mat4x4<f32>, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_blend(matA, matB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn color_cbrt(x: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sign(x) * pow(abs(x), 0.33333334);
}

fn natural_terrain(p: vec3<f32>, scale: f32, snowBias: f32, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  let unit = scale / 12.0;
  let continents = noise_fbmBump4(p, scale, 0.5, 2.0, noiseTex) - 0.25 * unit;
  let ridges = noise_turbulenceBump2(p, scale / 4.0, 0.5, 2.0, noiseTex);
  let height = 1.2 * continents + 2.2 * (0.45 * unit - ridges) * clamp(continents / unit, 0.0, 1.0);
  let band = (height + 0.25 * unit * noise_gradient(p, scale / 4.0, noiseTex)) / unit;
  let seaLevel = height / unit;
  let sandMat = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let sandy = texture_paintMaterial(seaLevel, 0.0, 0.02, mat4x4<f32>(vec4<f32>(0.08, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), sandMat, noiseTex);
  let grassMat = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let grassy = texture_paintMaterial(band, 0.2, 0.12, sandy, grassMat, noiseTex);
  let stony = texture_paintMaterial(band, 1.2, 0.25, grassy, mat4x4<f32>(vec4<f32>(0.8, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
  let snowMat = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.3098, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let sea = vec3<f32>(0.49309835, -0.060483743, -0.05850419) + vec3<f32>(-0.2985687, 0.053410567, -0.009546458) * clamp(-seaLevel / 1.2, 0.0, 1.0);
  let coast = sea + (vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.44005498, 0.05374613, 0.0 > 0.0), select(0.38603505, 0.050665636, 0.0 > 0.0), select(0.23069455, 0.04006192, 0.0 > 0.0))), noiseTex)))) - sea) * texture_mask(seaLevel, 0.0, 0.02, noiseTex);
  let grassLab = vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex))));
  let meadow = grassLab + (vec3<f32>(0.34344932, -0.06014222, 0.05241708) - grassLab) * texture_mask(noise_gradient(p, scale / 3.0, noiseTex), 0.0, 0.3, noiseTex);
  let green = coast + (meadow - coast) * texture_mask(band, 0.2, 0.12, noiseTex);
  let rock = color_rgb2oklab(stone_granite(p, 0.6 * unit, vec3<f32>(0.42, 0.4, 0.38), vec3<f32>(0.22, 0.21, 0.2), vec3<f32>(0.62, 0.6, 0.58), noiseTex), noiseTex);
  let rocky = green + (rock - green) * texture_mask(band, 1.2, 0.25, noiseTex);
  let land = rocky + (vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex)))) - rocky) * texture_mask(band + snowBias, 2.0, 0.2, noiseTex);
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = texture_paintMaterial(band + snowBias, 2.0, 0.2, stony, snowMat, noiseTex);
  gsl_result.v_2 = color_oklab2rgb(land, noiseTex);
  gsl_result.v_3 = height;
  return gsl_result;
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let latitude = abs(p.z) / 25.0;
  let polar = 3.0 * clamp((latitude - 0.72) / 0.28, 0.0, 1.0);
  let gsl_result_3 = natural_terrain(p, 12.0, polar, noiseTex);
  let height = gsl_result_3.v_3;
  let planet = length(p) - 25.0 - max(height, 0.0);
  let gsl_s1 = gsl_shape(planet, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let latitude = abs(p.z) / 25.0;
    let polar = 3.0 * clamp((latitude - 0.72) / 0.28, 0.0, 1.0);
    let gsl_result_3 = natural_terrain(p, 12.0, polar, noiseTex);
    let planetColor = gsl_result_3.v_2;
    return vec4<f32>(planetColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let latitude = abs(p.z) / 25.0;
    let polar = 3.0 * clamp((latitude - 0.72) / 0.28, 0.0, 1.0);
    let gsl_result_3 = natural_terrain(p, 12.0, polar, noiseTex);
    let planetMaterial = gsl_result_3.v_1;
    let gsl_m = planetMaterial; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_detailPw: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
}
struct gsl_struct_2 {
  v_1 : mat4x4<f32>,
  v_2 : vec3<f32>,
  v_3 : f32,
}
