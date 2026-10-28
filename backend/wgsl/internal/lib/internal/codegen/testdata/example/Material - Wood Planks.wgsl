fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

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

fn builtin_gcl_closure_intersectRound_2(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_2(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_2(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn map_axisAt(p: vec3<f32>, point: vec3<f32>, axis: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let d = p - point;
  let n = normalize(axis);
  let sz = sign(n.z);
  let s = sz + 1.0 - abs(sz);
  let a = -1.0 / (s + n.z);
  let b = n.x * n.y * a;
  let t1 = vec3<f32>(1.0 + s * n.x * n.x * a, s * b, -(s * n.x));
  let t2 = vec3<f32>(b, s + n.y * n.y * a, -n.y);
  return vec3<f32>(dot(d, t1), dot(d, t2), dot(d, n));
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

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn color_srgbToLinearChannel(s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x, noiseTex), color_srgbToLinearChannel(c.y, noiseTex), color_srgbToLinearChannel(c.z, noiseTex));
}

fn color_linearToSrgbChannel(l: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(1.055 * pow(l, 0.41666666) - 0.055, l * 12.92, l <= 0.0031308);
}

fn color_gammaEncode(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_linearToSrgbChannel(c.x, noiseTex), color_linearToSrgbChannel(c.y, noiseTex), color_linearToSrgbChannel(c.z, noiseTex));
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
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

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let la = color_rgb2oklab(a, noiseTex);
  let lb = color_rgb2oklab(b, noiseTex);
  return color_oklab2rgb(la + (lb - la) * t, noiseTex);
}

fn material_roughnessFromGloss(gloss: f32, noiseTex: texture_3d<f32>) -> f32 {
  let g = clamp(gloss, 0.5, 90.0);
  return clamp(0.647 - 0.1243 * log(g) + 0.008 * log(1.0 - g / 91.5), 0.045, 1.0);
}

fn material_filmKeeps(wavelength: f32, build: f32, noiseTex: texture_3d<f32>) -> f32 {
  let x = wavelength / (15.0 * max(build, 0.0001));
  return 1.0 - exp(-x * x);
}

fn material_hiddenShare(fade: f32, standIn: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(1.0 - fade * gsl_dispDrawn, 1.0, standIn > 0.5);
}

fn material_make(rough: f32, metal: f32, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(vec4<f32>(rough, metal, index, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
}

fn material_withLay(m: mat4x4<f32>, direction: vec3<f32>, anisotropy: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], m[1], m[2], vec4<f32>(direction.x, direction.y, direction.z, anisotropy));
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn wood_pickSpecies(number: f32, whiteOak: vec4<f32>, redOak: vec4<f32>, liveOak: vec4<f32>, japaneseElm: vec4<f32>, walnut: vec4<f32>, cherry: vec4<f32>, sugarMaple: vec4<f32>, norwayMaple: vec4<f32>, beech: vec4<f32>, rosewood: vec4<f32>, macassarEbony: vec4<f32>, ebony: vec4<f32>, yew: vec4<f32>, douglasFir: vec4<f32>, pine: vec4<f32>, chestnut: vec4<f32>, japaneseWalnut: vec4<f32>, keyaki: vec4<f32>, cherryBirch: vec4<f32>, sycamore: vec4<f32>, sakura: vec4<f32>, horseChestnut: vec4<f32>, hinoki: vec4<f32>, sugi: vec4<f32>, ginkgo: vec4<f32>, poplar: vec4<f32>, darkPoplar: vec4<f32>, butternut: vec4<f32>, englishWalnut: vec4<f32>, honeyLocust: vec4<f32>, hickory: vec4<f32>, teak: vec4<f32>, padauk: vec4<f32>, mahogany: vec4<f32>, americanElm: vec4<f32>, dutchElm: vec4<f32>, chineseElm: vec4<f32>, whiteAsh: vec4<f32>, europeanAsh: vec4<f32>, oliveAsh: vec4<f32>, osageOrange: vec4<f32>, europeanSycamore: vec4<f32>, purpleheart: vec4<f32>, zebrawood: vec4<f32>, lacewood: vec4<f32>, spruce: vec4<f32>, westernRedCedar: vec4<f32>, aromaticCedar: vec4<f32>, holly: vec4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(select(holly, aromaticCedar, number < 47.5), westernRedCedar, number < 46.5), spruce, number < 45.5), lacewood, number < 44.5), zebrawood, number < 43.5), purpleheart, number < 42.5), europeanSycamore, number < 41.5), osageOrange, number < 40.5), oliveAsh, number < 39.5), europeanAsh, number < 38.5), whiteAsh, number < 37.5), chineseElm, number < 36.5), dutchElm, number < 35.5), americanElm, number < 34.5), mahogany, number < 33.5), padauk, number < 32.5), teak, number < 31.5), hickory, number < 30.5), honeyLocust, number < 29.5), englishWalnut, number < 28.5), butternut, number < 27.5), darkPoplar, number < 26.5), poplar, number < 25.5), ginkgo, number < 24.5), sugi, number < 23.5), hinoki, number < 22.5), horseChestnut, number < 21.5), sakura, number < 20.5), sycamore, number < 19.5), cherryBirch, number < 18.5), keyaki, number < 17.5), japaneseWalnut, number < 16.5), chestnut, number < 15.5), pine, number < 14.5), douglasFir, number < 13.5), yew, number < 12.5), ebony, number < 11.5), macassarEbony, number < 10.5), rosewood, number < 9.5), beech, number < 8.5), norwayMaple, number < 7.5), sugarMaple, number < 6.5), cherry, number < 5.5), walnut, number < 4.5), japaneseElm, number < 3.5), liveOak, number < 2.5), redOak, number < 1.5), whiteOak, number < 0.5);
}

fn wood_speciesRings(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(1.5, 0.6, 0.3, 0.09), vec4<f32>(1.9, 0.6, 0.3, 0.09), vec4<f32>(2.0, 0.45, 0.45, 0.06), vec4<f32>(4.5, 0.6, 0.3, 0.175), vec4<f32>(3.3, 0.45, 0.45, 0.15), vec4<f32>(2.8, 0.45, 0.45, 0.125), vec4<f32>(2.0, 0.45, 0.45, 0.125), vec4<f32>(2.0, 0.45, 0.45, 0.125), vec4<f32>(2.7, 0.45, 0.45, 0.125), vec4<f32>(3.0, 0.45, 0.45, 0.1), vec4<f32>(2.5, 0.45, 0.45, 0.075), vec4<f32>(2.0, 0.45, 0.45, 0.05), vec4<f32>(1.5, 0.25, 0.1, 0.2), vec4<f32>(3.0, 0.35, 0.05, 0.225), vec4<f32>(4.0, 0.4, 0.3, 0.2), vec4<f32>(3.5, 0.6, 0.3, 0.09), vec4<f32>(8.8, 0.45, 0.45, 0.15), vec4<f32>(3.85, 0.6, 0.3, 0.09), vec4<f32>(4.4, 0.45, 0.45, 0.125), vec4<f32>(3.6, 0.45, 0.45, 0.125), vec4<f32>(4.9, 0.45, 0.45, 0.125), vec4<f32>(1.8, 0.45, 0.45, 0.125), vec4<f32>(2.2, 0.2, 0.08, 0.2), vec4<f32>(5.9, 0.35, 0.15, 0.2), vec4<f32>(5.5, 0.3, 0.2, 0.2), vec4<f32>(4.0, 0.45, 0.45, 0.1), vec4<f32>(4.0, 0.45, 0.45, 0.1), vec4<f32>(4.0, 0.45, 0.45, 0.15), vec4<f32>(3.5, 0.45, 0.45, 0.15), vec4<f32>(3.5, 0.6, 0.3, 0.09), vec4<f32>(3.0, 0.6, 0.3, 0.09), vec4<f32>(4.0, 0.45, 0.45, 0.15), vec4<f32>(3.0, 0.45, 0.45, 0.1), vec4<f32>(4.0, 0.45, 0.45, 0.1), vec4<f32>(3.5, 0.7, 0.3, 0.12), vec4<f32>(3.5, 0.7, 0.3, 0.12), vec4<f32>(3.0, 0.7, 0.3, 0.12), vec4<f32>(3.5, 0.7, 0.3, 0.1), vec4<f32>(3.5, 0.7, 0.3, 0.1), vec4<f32>(3.5, 0.7, 0.3, 0.1), vec4<f32>(3.0, 0.7, 0.3, 0.12), vec4<f32>(2.0, 0.45, 0.45, 0.125), vec4<f32>(3.0, 0.45, 0.45, 0.1), vec4<f32>(3.0, 0.25, 0.1, 0.3), vec4<f32>(4.0, 0.45, 0.45, 0.1), vec4<f32>(1.2, 0.2, 0.3, 0.15), vec4<f32>(2.8, 0.15, 0.1, 0.2), vec4<f32>(1.6, 0.15, 0.3, 0.25), vec4<f32>(1.4, 0.45, 0.45, 0.125), noiseTex);
}

fn wood_speciesEarlywood(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.63, 0.51, 0.41, 0.5), vec4<f32>(0.68, 0.53, 0.44, 0.5), vec4<f32>(0.49, 0.38, 0.24, 0.3), vec4<f32>(0.64, 0.53, 0.44, 0.45), vec4<f32>(0.46, 0.39, 0.35, 0.29), vec4<f32>(0.66, 0.52, 0.43, 0.14), vec4<f32>(0.79, 0.67, 0.53, 0.09), vec4<f32>(0.72, 0.59, 0.52, 0.09), vec4<f32>(0.75, 0.61, 0.53, 0.1), vec4<f32>(0.33, 0.25, 0.21, 0.35), vec4<f32>(0.37, 0.26, 0.15, 0.15), vec4<f32>(0.11, 0.11, 0.09, 0.15), vec4<f32>(0.72, 0.53, 0.28, 0.0), vec4<f32>(0.76, 0.63, 0.5, 0.0), vec4<f32>(0.93, 0.8, 0.67, 0.0), vec4<f32>(0.68, 0.58, 0.51, 0.5), vec4<f32>(0.6, 0.52, 0.46, 0.29), vec4<f32>(0.62, 0.47, 0.35, 0.5), vec4<f32>(0.7, 0.58, 0.53, 0.15), vec4<f32>(0.76, 0.63, 0.56, 0.09), vec4<f32>(0.67, 0.53, 0.44, 0.14), vec4<f32>(0.76, 0.64, 0.57, 0.08), vec4<f32>(0.84, 0.73, 0.66, 0.0), vec4<f32>(0.7, 0.57, 0.49, 0.0), vec4<f32>(0.82, 0.68, 0.55, 0.0), vec4<f32>(0.97, 0.89, 0.71, 0.12), vec4<f32>(0.78, 0.78, 0.6, 0.12), vec4<f32>(0.81, 0.72, 0.6, 0.35), vec4<f32>(0.55, 0.46, 0.37, 0.29), vec4<f32>(0.79, 0.62, 0.48, 0.5), vec4<f32>(0.62, 0.47, 0.33, 0.45), vec4<f32>(0.41, 0.29, 0.14, 0.35), vec4<f32>(0.45, 0.2, 0.05, 0.35), vec4<f32>(0.58, 0.41, 0.22, 0.3), vec4<f32>(0.81, 0.7, 0.55, 0.4), vec4<f32>(0.62, 0.49, 0.43, 0.4), vec4<f32>(0.74, 0.52, 0.34, 0.3), vec4<f32>(0.82, 0.71, 0.57, 0.45), vec4<f32>(0.9, 0.77, 0.68, 0.45), vec4<f32>(0.83, 0.67, 0.58, 0.45), vec4<f32>(0.84, 0.7, 0.42, 0.4), vec4<f32>(0.69, 0.6, 0.44, 0.09), vec4<f32>(0.55, 0.36, 0.38, 0.2), vec4<f32>(0.72, 0.57, 0.4, 0.25), vec4<f32>(0.7, 0.56, 0.45, 0.15), vec4<f32>(0.98, 0.86, 0.71, 0.0), vec4<f32>(0.46, 0.34, 0.19, 0.0), vec4<f32>(0.68, 0.49, 0.45, 0.0), vec4<f32>(0.98, 0.9, 0.74, 0.08), noiseTex);
}

fn wood_speciesLatewood(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.57, 0.46, 0.37, 1.0), vec4<f32>(0.62, 0.47, 0.39, 1.0), vec4<f32>(0.45, 0.34, 0.21, 0.2), vec4<f32>(0.58, 0.47, 0.39, 1.0), vec4<f32>(0.42, 0.35, 0.31, 0.5), vec4<f32>(0.64, 0.5, 0.41, 0.2), vec4<f32>(0.72, 0.61, 0.47, 0.0), vec4<f32>(0.66, 0.53, 0.47, 0.0), vec4<f32>(0.69, 0.55, 0.47, 0.0), vec4<f32>(0.3, 0.23, 0.19, 0.3), vec4<f32>(0.34, 0.24, 0.13, 0.0), vec4<f32>(0.06, 0.05, 0.04, 0.0), vec4<f32>(0.42, 0.24, 0.05, 0.0), vec4<f32>(0.55, 0.36, 0.18, 0.0), vec4<f32>(0.79, 0.65, 0.5, 0.0), vec4<f32>(0.62, 0.52, 0.46, 1.0), vec4<f32>(0.55, 0.47, 0.41, 0.5), vec4<f32>(0.57, 0.42, 0.31, 1.0), vec4<f32>(0.64, 0.52, 0.48, 0.0), vec4<f32>(0.69, 0.57, 0.5, 0.0), vec4<f32>(0.61, 0.48, 0.39, 0.2), vec4<f32>(0.69, 0.58, 0.51, 0.0), vec4<f32>(0.59, 0.44, 0.34, 0.0), vec4<f32>(0.53, 0.38, 0.31, 0.0), vec4<f32>(0.72, 0.55, 0.4, 0.0), vec4<f32>(0.88, 0.81, 0.64, 0.0), vec4<f32>(0.71, 0.71, 0.54, 0.0), vec4<f32>(0.74, 0.65, 0.54, 0.5), vec4<f32>(0.5, 0.42, 0.33, 0.5), vec4<f32>(0.72, 0.56, 0.43, 1.0), vec4<f32>(0.57, 0.42, 0.29, 1.0), vec4<f32>(0.37, 0.26, 0.12, 0.5), vec4<f32>(0.41, 0.17, 0.02, 0.1), vec4<f32>(0.53, 0.37, 0.19, 0.0), vec4<f32>(0.74, 0.64, 0.5, 1.0), vec4<f32>(0.56, 0.45, 0.39, 1.0), vec4<f32>(0.67, 0.47, 0.31, 0.8), vec4<f32>(0.75, 0.65, 0.52, 1.0), vec4<f32>(0.82, 0.7, 0.62, 1.0), vec4<f32>(0.76, 0.61, 0.53, 1.0), vec4<f32>(0.76, 0.64, 0.38, 1.0), vec4<f32>(0.63, 0.55, 0.4, 0.0), vec4<f32>(0.5, 0.33, 0.35, 0.0), vec4<f32>(0.4, 0.3, 0.2, 0.2), vec4<f32>(0.64, 0.51, 0.41, 0.0), vec4<f32>(0.84, 0.71, 0.54, 0.0), vec4<f32>(0.19, 0.12, 0.03, 0.0), vec4<f32>(0.61, 0.42, 0.39, 0.0), vec4<f32>(0.91, 0.83, 0.67, 0.15), noiseTex);
}

fn wood_speciesPoreRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.45, 0.3, 0.2, 1.0), vec4<f32>(0.41, 0.23, 0.15, 1.0), vec4<f32>(0.3, 0.22, 0.15, 1.0), vec4<f32>(0.45, 0.33, 0.25, 0.0), vec4<f32>(0.29, 0.22, 0.18, 0.2), vec4<f32>(0.43, 0.28, 0.21, 0.3), vec4<f32>(0.6, 0.47, 0.4, 0.3), vec4<f32>(0.6, 0.47, 0.4, 0.3), vec4<f32>(0.6, 0.46, 0.38, 0.7), vec4<f32>(0.15, 0.11, 0.1, 0.0), vec4<f32>(0.12, 0.1, 0.09, 0.0), vec4<f32>(0.05, 0.05, 0.05, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.54, 0.39, 0.25, 0.1), vec4<f32>(0.47, 0.34, 0.26, 0.2), vec4<f32>(0.39, 0.23, 0.15, 0.4), vec4<f32>(0.53, 0.44, 0.4, 0.3), vec4<f32>(0.58, 0.48, 0.42, 1.0), vec4<f32>(0.51, 0.4, 0.33, 0.3), vec4<f32>(0.58, 0.49, 0.43, 0.1), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.74, 0.68, 0.54, 0.2), vec4<f32>(0.6, 0.6, 0.46, 0.2), vec4<f32>(0.55, 0.45, 0.35, 0.2), vec4<f32>(0.3, 0.22, 0.17, 0.2), vec4<f32>(0.5, 0.33, 0.22, 0.3), vec4<f32>(0.45, 0.33, 0.22, 0.3), vec4<f32>(0.25, 0.17, 0.08, 0.1), vec4<f32>(0.25, 0.1, 0.03, 0.1), vec4<f32>(0.33, 0.21, 0.1, 0.3), vec4<f32>(0.52, 0.42, 0.3, 0.3), vec4<f32>(0.4, 0.3, 0.25, 0.3), vec4<f32>(0.48, 0.32, 0.2, 0.3), vec4<f32>(0.55, 0.45, 0.34, 0.2), vec4<f32>(0.6, 0.48, 0.4, 0.2), vec4<f32>(0.56, 0.42, 0.34, 0.2), vec4<f32>(0.72, 0.6, 0.34, 0.2), vec4<f32>(0.55, 0.47, 0.36, 0.35), vec4<f32>(0.38, 0.22, 0.24, 0.3), vec4<f32>(0.45, 0.35, 0.24, 0.2), vec4<f32>(0.5, 0.38, 0.3, 1.2), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.74, 0.63, 0.56, 0.5), noiseTex);
}

fn wood_speciesKnotRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.35, 0.28, 0.22, 1.0), vec4<f32>(0.38, 0.28, 0.23, 0.8), vec4<f32>(0.27, 0.2, 0.12, 1.2), vec4<f32>(0.36, 0.28, 0.23, 1.0), vec4<f32>(0.25, 0.21, 0.18, 1.0), vec4<f32>(0.39, 0.3, 0.25, 0.3), vec4<f32>(0.45, 0.37, 0.28, 0.25), vec4<f32>(0.41, 0.32, 0.28, 0.25), vec4<f32>(0.43, 0.34, 0.28, 0.45), vec4<f32>(0.17, 0.13, 0.1, 1.0), vec4<f32>(0.2, 0.14, 0.06, 1.0), vec4<f32>(0.02, 0.02, 0.01, 1.0), vec4<f32>(0.45, 0.28, 0.2, 1.0), vec4<f32>(0.34, 0.21, 0.1, 1.0), vec4<f32>(0.561, 0.387, 0.262, 1.0), vec4<f32>(0.38, 0.32, 0.28, 0.6), vec4<f32>(0.34, 0.28, 0.25, 1.0), vec4<f32>(0.35, 0.25, 0.18, 0.6), vec4<f32>(0.39, 0.32, 0.29, 0.3), vec4<f32>(0.43, 0.35, 0.3, 0.7), vec4<f32>(0.37, 0.29, 0.23, 0.3), vec4<f32>(0.43, 0.36, 0.31, 0.2), vec4<f32>(0.5, 0.32, 0.22, 1.0), vec4<f32>(0.45, 0.28, 0.2, 1.0), vec4<f32>(0.45, 0.34, 0.24, 1.0), vec4<f32>(0.55, 0.5, 0.39, 0.3), vec4<f32>(0.44, 0.44, 0.33, 0.3), vec4<f32>(0.46, 0.4, 0.33, 1.0), vec4<f32>(0.3, 0.25, 0.19, 1.0), vec4<f32>(0.45, 0.34, 0.26, 0.6), vec4<f32>(0.35, 0.25, 0.17, 0.5), vec4<f32>(0.22, 0.15, 0.06, 0.5), vec4<f32>(0.25, 0.09, 0.01, 0.5), vec4<f32>(0.32, 0.22, 0.1, 0.4), vec4<f32>(0.42, 0.33, 0.24, 0.4), vec4<f32>(0.33, 0.25, 0.21, 0.4), vec4<f32>(0.38, 0.26, 0.17, 0.4), vec4<f32>(0.45, 0.36, 0.27, 0.3), vec4<f32>(0.49, 0.39, 0.32, 0.3), vec4<f32>(0.45, 0.34, 0.27, 0.3), vec4<f32>(0.55, 0.42, 0.2, 0.4), vec4<f32>(0.4, 0.33, 0.25, 0.3), vec4<f32>(0.3, 0.18, 0.18, 0.4), vec4<f32>(0.35, 0.26, 0.16, 0.4), vec4<f32>(0.4, 0.3, 0.24, 2.5), vec4<f32>(0.59, 0.42, 0.28, 1.0), vec4<f32>(0.36, 0.25, 0.09, 1.0), vec4<f32>(0.49, 0.36, 0.33, 1.0), vec4<f32>(0.56, 0.5, 0.4, 0.5), noiseTex);
}

fn wood_speciesSapwood(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.68, 0.56, 0.47, 20.0), vec4<f32>(0.83, 0.7, 0.56, 8.0), vec4<f32>(0.54, 0.43, 0.29, 12.0), vec4<f32>(0.74, 0.63, 0.54, 3.0), vec4<f32>(0.73, 0.64, 0.53, 13.0), vec4<f32>(0.75, 0.66, 0.6, 10.0), vec4<f32>(0.92, 0.86, 0.78, 50.0), vec4<f32>(0.9, 0.84, 0.76, 40.0), vec4<f32>(0.79, 0.65, 0.56, 30.0), vec4<f32>(0.78, 0.71, 0.61, 12.0), vec4<f32>(0.8, 0.66, 0.42, 24.0), vec4<f32>(0.78, 0.68, 0.55, 42.0), vec4<f32>(0.76, 0.64, 0.47, 12.0), vec4<f32>(0.76, 0.67, 0.55, 13.0), vec4<f32>(0.94, 0.86, 0.75, 15.0), vec4<f32>(0.84, 0.8, 0.72, 5.0), vec4<f32>(0.74, 0.68, 0.63, 2.0), vec4<f32>(0.65, 0.59, 0.51, 8.0), vec4<f32>(0.88, 0.82, 0.68, 15.0), vec4<f32>(0.85, 0.76, 0.74, 17.0), vec4<f32>(0.86, 0.75, 0.69, 8.0), vec4<f32>(0.76, 0.64, 0.57, 40.0), vec4<f32>(0.87, 0.79, 0.74, 16.0), vec4<f32>(0.92, 0.84, 0.74, 15.0), vec4<f32>(0.82, 0.68, 0.55, 20.0), vec4<f32>(0.95, 0.92, 0.84, 10.0), vec4<f32>(0.95, 0.92, 0.84, 10.0), vec4<f32>(0.98, 0.92, 0.8, 6.0), vec4<f32>(0.71, 0.61, 0.52, 12.0), vec4<f32>(0.92, 0.84, 0.71, 10.0), vec4<f32>(0.88, 0.78, 0.61, 17.0), vec4<f32>(0.51, 0.43, 0.26, 9.0), vec4<f32>(0.57, 0.48, 0.35, 27.0), vec4<f32>(0.68, 0.62, 0.49, 9.0), vec4<f32>(0.93, 0.83, 0.68, 3.0), vec4<f32>(0.71, 0.58, 0.53, 3.0), vec4<f32>(0.9, 0.84, 0.72, 4.0), vec4<f32>(0.87, 0.79, 0.66, 25.0), vec4<f32>(0.9, 0.77, 0.68, 22.0), vec4<f32>(0.95, 0.78, 0.69, 22.0), vec4<f32>(0.88, 0.78, 0.6, 2.0), vec4<f32>(0.69, 0.6, 0.44, 40.0), vec4<f32>(0.83, 0.74, 0.65, 15.0), vec4<f32>(0.83, 0.73, 0.59, 33.0), vec4<f32>(0.82, 0.72, 0.61, 8.0), vec4<f32>(0.98, 0.89, 0.73, 60.0), vec4<f32>(0.9, 0.84, 0.72, 10.0), vec4<f32>(0.86, 0.7, 0.55, 12.0), vec4<f32>(0.97, 0.91, 0.78, 40.0), noiseTex);
}

fn wood_speciesStreakRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.48, 0.3, 0.22, 0.35), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.2, 0.15, 0.15, 0.7), vec4<f32>(0.1, 0.09, 0.07, 1.0), vec4<f32>(0.14, 0.11, 0.09, 0.25), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.72, 0.72, 0.55, 0.25), vec4<f32>(0.4, 0.34, 0.42, 0.8), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.4, 0.3, 0.22, 0.4), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.3, 0.2, 0.1, 0.5), vec4<f32>(0.36, 0.14, 0.04, 0.3), vec4<f32>(0.5, 0.33, 0.17, 0.35), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.55, 0.42, 0.33, 0.7), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.4, 0.29, 0.15, 0.6), vec4<f32>(0.86, 0.7, 0.55, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.0), noiseTex);
}

fn wood_speciesSurfaceRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.5, 0.3, 0.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), vec4<f32>(0.3, 0.2, 0.0, 0.0), vec4<f32>(0.25, 0.15, 0.0, 0.0), vec4<f32>(0.35, 0.6, 120.0, 0.0), vec4<f32>(0.45, 0.8, 0.0, 0.3), vec4<f32>(0.5, 0.8, 450.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.45, 0.4, 0.0, 0.0), vec4<f32>(0.4, 0.2, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.4, 0.2, 0.0, 0.0), vec4<f32>(0.45, 0.6, 250.0, 0.0), vec4<f32>(0.45, 0.8, 300.0, 0.0), vec4<f32>(0.45, 0.4, 0.0, 0.0), vec4<f32>(0.45, 0.25, 0.0, 0.0), vec4<f32>(0.45, 0.25, 0.0, 0.0), vec4<f32>(0.4, 0.3, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.45, 0.4, 0.0, 0.0), vec4<f32>(0.45, 0.35, 0.0, 0.0), vec4<f32>(0.35, 0.25, 0.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.45, 0.3, 0.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.5, 0.4, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), vec4<f32>(0.4, 0.25, 0.0, 0.0), vec4<f32>(0.45, 0.3, 0.0, 0.0), vec4<f32>(0.45, 0.8, 550.0, 0.2), vec4<f32>(0.5, 0.8, 350.0, 0.0), vec4<f32>(0.4, 0.5, 150.0, 0.0), vec4<f32>(0.35, 0.2, 0.0, 0.0), noiseTex);
}

fn wood_speciesAgeRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.704, 0.592, 0.211, 0.0), vec4<f32>(0.858, 0.739, 0.553, 0.0), vec4<f32>(0.804, 0.681, 0.351, 0.0), vec4<f32>(0.932, 0.869, 0.777, 0.0), vec4<f32>(1.668, 1.777, 1.554, 0.0), vec4<f32>(0.923, 0.801, 0.974, 0.0), vec4<f32>(0.74, 0.631, 0.476, 0.0), vec4<f32>(0.726, 0.599, 0.476, 0.0), vec4<f32>(0.947, 0.774, 0.635, 0.0), vec4<f32>(0.725, 0.612, 0.645, 0.0), vec4<f32>(0.843, 0.74, 0.578, 0.0), vec4<f32>(0.768, 0.585, 0.375, 0.0), vec4<f32>(0.876, 0.734, 0.484, 0.0), vec4<f32>(0.842, 0.633, 0.359, 0.0), vec4<f32>(0.823, 0.581, 0.303, 0.0), vec4<f32>(0.87, 0.755, 0.595, 0.0), vec4<f32>(1.534, 1.61, 1.445, 0.0), vec4<f32>(0.839, 0.719, 0.482, 0.0), vec4<f32>(0.724, 0.594, 0.483, 0.0), vec4<f32>(0.737, 0.615, 0.498, 0.0), vec4<f32>(0.706, 0.579, 0.502, 0.0), vec4<f32>(0.739, 0.619, 0.503, 0.0), vec4<f32>(0.877, 0.666, 0.459, 0.0), vec4<f32>(0.837, 0.607, 0.36, 0.0), vec4<f32>(0.853, 0.65, 0.392, 0.0), vec4<f32>(0.895, 0.77, 0.538, 0.0), vec4<f32>(0.911, 0.688, 0.537, 0.0), vec4<f32>(1.196, 1.217, 1.17, 0.0), vec4<f32>(1.573, 1.677, 1.527, 0.0), vec4<f32>(0.869, 0.766, 0.57, 0.0), vec4<f32>(0.836, 0.719, 0.459, 0.0), vec4<f32>(0.696, 0.538, 0.267, 0.0), vec4<f32>(1.164, 1.328, 1.163, 0.0), vec4<f32>(0.758, 0.62, 0.405, 0.0), vec4<f32>(0.932, 0.869, 0.777, 0.0), vec4<f32>(0.932, 0.869, 0.777, 0.0), vec4<f32>(0.932, 0.869, 0.777, 0.0), vec4<f32>(0.85, 0.74, 0.52, 0.0), vec4<f32>(0.85, 0.74, 0.52, 0.0), vec4<f32>(0.85, 0.74, 0.52, 0.0), vec4<f32>(0.2, 0.065, 0.15, 0.0), vec4<f32>(0.726, 0.599, 0.476, 0.0), vec4<f32>(1.186, 1.153, 0.463, 0.0), vec4<f32>(0.85, 0.75, 0.6, 0.0), vec4<f32>(0.85, 0.75, 0.6, 0.0), vec4<f32>(0.876, 0.604, 0.288, 0.0), vec4<f32>(0.832, 0.731, 0.597, 0.0), vec4<f32>(0.948, 0.977, 0.706, 0.0), vec4<f32>(0.937, 0.863, 0.707, 0.0), noiseTex);
}

fn wood_speciesBandRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.0, 0.0, 0.0, 1.5), vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.25, 0.3, 0.6, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 1.0), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 1.0), vec4<f32>(0.0, 0.0, 0.0, 3.0), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 1.0), vec4<f32>(0.0, 0.0, 0.0, 0.7), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 1.4), vec4<f32>(0.0, 0.0, 0.0, 1.0), vec4<f32>(0.0, 0.0, 0.0, 3.0), vec4<f32>(0.0, 0.0, 0.0, 4.0), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 3.0), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.25, 0.3, 0.5, 0.5), vec4<f32>(0.25, 0.3, 0.5, 0.5), vec4<f32>(0.25, 0.25, 0.4, 0.5), vec4<f32>(0.05, 0.4, 0.7, 0.5), vec4<f32>(0.05, 0.4, 0.7, 4.0), vec4<f32>(0.05, 0.4, 0.7, 0.5), vec4<f32>(0.3, 0.35, 0.5, 0.3), vec4<f32>(0.0, 0.0, 0.0, 1.0), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.0, 0.0, 0.0, 6.0), vec4<f32>(0.0, 0.0, 0.0, 2.0), vec4<f32>(0.0, 0.0, 0.0, 0.5), vec4<f32>(0.0, 0.0, 0.0, 3.0), noiseTex);
}

fn wood_speciesBoundaryRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.4, 0.32, 0.26, 0.5), vec4<f32>(0.3, 0.24, 0.2, 0.4), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.6, 0.45, 0.32, 0.6), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.72, 0.58, 0.36, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.82, 0.62, 0.55, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.55, 0.45, 0.3, 0.6), vec4<f32>(0.55, 0.45, 0.3, 0.6), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.38, 0.31, 0.24, 0.4), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.3, 0.22, 0.12, 0.4), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.4, 0.32, 0.26, 0.5), vec4<f32>(0.4, 0.32, 0.26, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.55, 0.45, 0.35, 0.3), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), noiseTex);
}

fn wood_speciesWeeksRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(0.78, 0.69, 0.6, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(0.61, 1.31, 1.6, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(0.718, 0.852, 1.959, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(0.948, 0.977, 0.706, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), noiseTex);
}

fn wood_speciesStainRow(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(0.8, 0.8, 1.6, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.05, 0.95, 2.5, 0.1), vec4<f32>(1.0, 1.0, 1.3, 0.6), vec4<f32>(1.0, 1.0, 1.3, 0.5), vec4<f32>(0.8, 0.8, 1.04, 0.5), vec4<f32>(1.0, 1.0, 1.3, 0.3), vec4<f32>(0.63, 0.57, 1.5, 0.1), vec4<f32>(0.4, 0.4, 0.52, 0.1), vec4<f32>(0.3, 0.3, 0.39, 0.1), vec4<f32>(1.3, 0.5, 1.0, 0.2), vec4<f32>(1.3, 0.5, 1.0, 0.3), vec4<f32>(1.3, 0.5, 1.0, 0.6), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.05, 0.95, 2.5, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.0, 1.0, 1.3, 0.7), vec4<f32>(1.0, 1.0, 1.3, 0.4), vec4<f32>(1.0, 1.0, 1.3, 0.5), vec4<f32>(1.0, 1.0, 1.3, 0.3), vec4<f32>(1.3, 0.5, 1.0, 0.3), vec4<f32>(1.3, 0.5, 1.0, 0.3), vec4<f32>(1.3, 0.5, 1.0, 0.3), vec4<f32>(1.0, 1.0, 1.3, 0.5), vec4<f32>(1.0, 1.0, 1.3, 0.5), vec4<f32>(1.05, 0.95, 2.5, 0.1), vec4<f32>(1.05, 0.95, 2.5, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(0.99, 0.81, 2.7, 0.1), vec4<f32>(0.525, 0.475, 1.25, 0.1), vec4<f32>(0.6, 0.6, 1.2, 0.1), vec4<f32>(1.0, 1.0, 2.0, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(1.1, 0.9, 3.0, 0.1), vec4<f32>(0.66, 0.54, 1.8, 0.1), vec4<f32>(0.8, 0.8, 1.04, 0.5), vec4<f32>(0.6, 0.6, 0.78, 0.1), vec4<f32>(0.8, 0.8, 1.6, 0.1), vec4<f32>(1.0, 1.0, 1.3, 0.1), vec4<f32>(1.3, 0.5, 1.0, 0.6), vec4<f32>(1.3, 0.5, 1.0, 0.3), vec4<f32>(1.04, 0.4, 0.8, 0.2), vec4<f32>(1.0, 1.0, 1.3, 0.2), noiseTex);
}

fn wood_speciesKnotStyle(number: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return wood_pickSpecies(number, vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(5.0, 2.0, 0.3, 0.5), vec4<f32>(6.0, 2.0, 0.4, 0.0), vec4<f32>(6.0, 0.0, 0.7, 0.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(6.0, 0.0, 0.4, 1.0), vec4<f32>(6.0, 0.0, 0.4, 0.5), vec4<f32>(4.0, 0.0, 0.55, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), vec4<f32>(5.0, 2.0, 0.31, 0.0), vec4<f32>(4.0, 0.0, 0.4, 1.0), vec4<f32>(4.0, 0.0, 0.5, 1.0), vec4<f32>(2.0, 0.0, 0.7, 1.0), noiseTex);
}

fn wood_prepCoarseness(prep: mat3x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let head = prep[0];
  let kind = head.x;
  let sandedCoarseness = pow(180.0 / max(head.y, 24.0), 0.15);
  return select(select(select(select(1.5, 0.95, kind < 4.5), 1.5, kind < 3.5), 1.9, kind < 2.5), sandedCoarseness, kind < 1.5);
}

fn wood_rawChannel(v: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(pow((v + 0.055) / 1.055, 2.4), v / 12.92, v <= 0.04045);
}

fn wood_charredColor(c: vec3<f32>, char: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let browned = c * vec3<f32>(pow(0.52, char / 0.3), pow(0.38, char / 0.3), pow(0.28, char / 0.3));
  return browned + (vec3<f32>(0.02, 0.018, 0.016) - browned) * (1.0 - exp(-(char / 0.6) * char / 0.6));
}

fn wood_finishedColor(raw: vec3<f32>, aging: vec3<f32>, char: f32, stainGain: vec3<f32>, uptake: f32, wetting: f32, coatTint: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let c = wood_charredColor(vec3<f32>(wood_rawChannel(raw.x, noiseTex), wood_rawChannel(raw.y, noiseTex), wood_rawChannel(raw.z, noiseTex)) * aging, char, noiseTex) * vec3<f32>(pow(stainGain.x, uptake), pow(stainGain.y, uptake), pow(stainGain.z, uptake));
  let wet = min(vec3<f32>(pow(c.x, 1.403), pow(c.y, 1.403), pow(c.z, 1.403)) * 1.042, vec3<f32>(1.0, 1.0, 1.0));
  return color_gammaEncode(min((c + (wet - c) * wetting) * coatTint, vec3<f32>(1.0, 1.0, 1.0)), noiseTex);
}

fn gsl_wood_stemGrowthAt_remap(p: vec3<f32>, axisDir: vec3<f32>, axisOrigin: vec3<f32>, ringWidth: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 60.0, noiseTex) + 0.5 * noise_gradientLattice(p / 30.0, noiseTex);
}

fn gsl_wood_stemGrowthAt_remap_2(p: vec3<f32>, axisDir: vec3<f32>, axisOrigin: vec3<f32>, ringWidth: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 25.0, noiseTex);
}

fn wood_stemGrowthAt(p: vec3<f32>, axisOrigin: vec3<f32>, axisDir: vec3<f32>, ringWidth: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let local = map_axisAt(p, axisOrigin, axisDir, noiseTex);
  let radius = length(local.xy);
  let around = local.xy / max(radius, 0.001) * 40.0;
  let wobble = gsl_wood_stemGrowthAt_remap(vec3<f32>(around.x, around.y, local.z * 0.25), axisDir, axisOrigin, ringWidth, noiseTex);
  let drift = gsl_wood_stemGrowthAt_remap_2(vec3<f32>(local.x, local.y, local.z * 0.2), axisDir, axisOrigin, ringWidth, noiseTex);
  return vec2<f32>(radius * (1.0 + 0.08 * wobble) / ringWidth + 0.3 * drift, 0.0);
}

fn wood_whorlTime(local: vec3<f32>, whorl: f32, ringWidth: f32, spacing: f32, branchRate: f32, seed: f32, style: vec4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_49 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 0.0 + seed, 2.0 * whorl + 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 0.0 + seed, 2.0 * whorl + 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_54 = local - vec3<f32>(0.0, 0.0, (whorl + style.w * (gsl_tmp_49.y - 0.5)) * spacing);
  let gsl_tmp_52 = (0.0 + 0.6 * gsl_tmp_49.x) * 6.2831855 / max(style.x, 1.0);
  let gsl_tmp_53 = vec2<f32>(cos(gsl_tmp_52), sin(gsl_tmp_52));
  let gsl_tmp_55 = max(dot(gsl_tmp_54.xy, gsl_tmp_53), 0.0);
  let gsl_tmp_56 = length(gsl_tmp_54.xy - gsl_tmp_53 * gsl_tmp_55);
  let gsl_tmp_50 = 1.0 + 0.6 * gsl_tmp_49.z * gsl_tmp_49.z;
  let gsl_tmp_51 = style.z * (0.34 + 1.2 * (min(gsl_tmp_50, 1.0) - 0.45));
  let gsl_tmp_57 = gsl_tmp_51 / (1.0 + gsl_tmp_55 / 60.0);
  let gsl_tmp_58 = (gsl_tmp_54.z - gsl_tmp_51 * 60.0 * log(1.0 + gsl_tmp_55 / 60.0)) / sqrt(1.0 + gsl_tmp_57 * gsl_tmp_57);
  let gsl_tmp_59 = sqrt(gsl_tmp_56 * gsl_tmp_56 + gsl_tmp_58 * gsl_tmp_58);
  let gsl_tmp_60 = select(1000.0, min(1000.0, (gsl_tmp_59 + 4.0 * max(gsl_tmp_59 - 12.5 * gsl_tmp_50, 0.0)) / (ringWidth * branchRate)), 0.0 < style.x);
  let gsl_tmp_61 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 7.0 + seed, 2.0 * whorl + 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 7.0 + seed, 2.0 * whorl + 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_66 = local - vec3<f32>(0.0, 0.0, (whorl + style.w * (gsl_tmp_61.y - 0.5)) * spacing);
  let gsl_tmp_64 = (1.0 + 0.6 * gsl_tmp_61.x) * 6.2831855 / max(style.x, 1.0);
  let gsl_tmp_65 = vec2<f32>(cos(gsl_tmp_64), sin(gsl_tmp_64));
  let gsl_tmp_67 = max(dot(gsl_tmp_66.xy, gsl_tmp_65), 0.0);
  let gsl_tmp_68 = length(gsl_tmp_66.xy - gsl_tmp_65 * gsl_tmp_67);
  let gsl_tmp_62 = 0.45 + 0.4 * gsl_tmp_61.z;
  let gsl_tmp_63 = style.z * (0.34 + 1.2 * (min(gsl_tmp_62, 1.0) - 0.45));
  let gsl_tmp_69 = gsl_tmp_63 / (1.0 + gsl_tmp_67 / 60.0);
  let gsl_tmp_70 = (gsl_tmp_66.z - gsl_tmp_63 * 60.0 * log(1.0 + gsl_tmp_67 / 60.0)) / sqrt(1.0 + gsl_tmp_69 * gsl_tmp_69);
  let gsl_tmp_71 = sqrt(gsl_tmp_68 * gsl_tmp_68 + gsl_tmp_70 * gsl_tmp_70);
  let gsl_tmp_72 = select(gsl_tmp_60, min(gsl_tmp_60, (gsl_tmp_71 + 4.0 * max(gsl_tmp_71 - 12.5 * gsl_tmp_62, 0.0)) / (ringWidth * branchRate)), 1.0 < style.x);
  let gsl_tmp_73 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 14.0 + seed, 2.0 * whorl + 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 14.0 + seed, 2.0 * whorl + 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_78 = local - vec3<f32>(0.0, 0.0, (whorl + style.w * (gsl_tmp_73.y - 0.5)) * spacing);
  let gsl_tmp_76 = (2.0 + 0.6 * gsl_tmp_73.x) * 6.2831855 / max(style.x, 1.0);
  let gsl_tmp_77 = vec2<f32>(cos(gsl_tmp_76), sin(gsl_tmp_76));
  let gsl_tmp_79 = max(dot(gsl_tmp_78.xy, gsl_tmp_77), 0.0);
  let gsl_tmp_80 = length(gsl_tmp_78.xy - gsl_tmp_77 * gsl_tmp_79);
  let gsl_tmp_74 = 0.45 + 0.4 * gsl_tmp_73.z;
  let gsl_tmp_75 = style.z * (0.34 + 1.2 * (min(gsl_tmp_74, 1.0) - 0.45));
  let gsl_tmp_81 = gsl_tmp_75 / (1.0 + gsl_tmp_79 / 60.0);
  let gsl_tmp_82 = (gsl_tmp_78.z - gsl_tmp_75 * 60.0 * log(1.0 + gsl_tmp_79 / 60.0)) / sqrt(1.0 + gsl_tmp_81 * gsl_tmp_81);
  let gsl_tmp_83 = sqrt(gsl_tmp_80 * gsl_tmp_80 + gsl_tmp_82 * gsl_tmp_82);
  let gsl_tmp_84 = select(gsl_tmp_72, min(gsl_tmp_72, (gsl_tmp_83 + 4.0 * max(gsl_tmp_83 - 12.5 * gsl_tmp_74, 0.0)) / (ringWidth * branchRate)), 2.0 < style.x);
  let gsl_tmp_85 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 21.0 + seed, 2.0 * whorl + 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 21.0 + seed, 2.0 * whorl + 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_90 = local - vec3<f32>(0.0, 0.0, (whorl + style.w * (gsl_tmp_85.y - 0.5)) * spacing);
  let gsl_tmp_88 = (3.0 + 0.6 * gsl_tmp_85.x) * 6.2831855 / max(style.x, 1.0);
  let gsl_tmp_89 = vec2<f32>(cos(gsl_tmp_88), sin(gsl_tmp_88));
  let gsl_tmp_91 = max(dot(gsl_tmp_90.xy, gsl_tmp_89), 0.0);
  let gsl_tmp_92 = length(gsl_tmp_90.xy - gsl_tmp_89 * gsl_tmp_91);
  let gsl_tmp_86 = 0.45 + 0.4 * gsl_tmp_85.z;
  let gsl_tmp_87 = style.z * (0.34 + 1.2 * (min(gsl_tmp_86, 1.0) - 0.45));
  let gsl_tmp_93 = gsl_tmp_87 / (1.0 + gsl_tmp_91 / 60.0);
  let gsl_tmp_94 = (gsl_tmp_90.z - gsl_tmp_87 * 60.0 * log(1.0 + gsl_tmp_91 / 60.0)) / sqrt(1.0 + gsl_tmp_93 * gsl_tmp_93);
  let gsl_tmp_95 = sqrt(gsl_tmp_92 * gsl_tmp_92 + gsl_tmp_94 * gsl_tmp_94);
  let gsl_tmp_96 = select(gsl_tmp_84, min(gsl_tmp_84, (gsl_tmp_95 + 4.0 * max(gsl_tmp_95 - 12.5 * gsl_tmp_86, 0.0)) / (ringWidth * branchRate)), 3.0 < style.x);
  let gsl_tmp_97 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 28.0 + seed, 2.0 * whorl + 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 28.0 + seed, 2.0 * whorl + 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_102 = local - vec3<f32>(0.0, 0.0, (whorl + style.w * (gsl_tmp_97.y - 0.5)) * spacing);
  let gsl_tmp_100 = (4.0 + 0.6 * gsl_tmp_97.x) * 6.2831855 / max(style.x, 1.0);
  let gsl_tmp_101 = vec2<f32>(cos(gsl_tmp_100), sin(gsl_tmp_100));
  let gsl_tmp_103 = max(dot(gsl_tmp_102.xy, gsl_tmp_101), 0.0);
  let gsl_tmp_104 = length(gsl_tmp_102.xy - gsl_tmp_101 * gsl_tmp_103);
  let gsl_tmp_98 = 0.45 + 0.4 * gsl_tmp_97.z;
  let gsl_tmp_99 = style.z * (0.34 + 1.2 * (min(gsl_tmp_98, 1.0) - 0.45));
  let gsl_tmp_105 = gsl_tmp_99 / (1.0 + gsl_tmp_103 / 60.0);
  let gsl_tmp_106 = (gsl_tmp_102.z - gsl_tmp_99 * 60.0 * log(1.0 + gsl_tmp_103 / 60.0)) / sqrt(1.0 + gsl_tmp_105 * gsl_tmp_105);
  let gsl_tmp_107 = sqrt(gsl_tmp_104 * gsl_tmp_104 + gsl_tmp_106 * gsl_tmp_106);
  let gsl_tmp_108 = select(gsl_tmp_96, min(gsl_tmp_96, (gsl_tmp_107 + 4.0 * max(gsl_tmp_107 - 12.5 * gsl_tmp_98, 0.0)) / (ringWidth * branchRate)), 4.0 < style.x);
  let gsl_tmp_109 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 35.0 + seed, 2.0 * whorl + 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 35.0 + seed, 2.0 * whorl + 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_114 = local - vec3<f32>(0.0, 0.0, (whorl + style.w * (gsl_tmp_109.y - 0.5)) * spacing);
  let gsl_tmp_112 = (5.0 + 0.6 * gsl_tmp_109.x) * 6.2831855 / max(style.x, 1.0);
  let gsl_tmp_113 = vec2<f32>(cos(gsl_tmp_112), sin(gsl_tmp_112));
  let gsl_tmp_115 = max(dot(gsl_tmp_114.xy, gsl_tmp_113), 0.0);
  let gsl_tmp_116 = length(gsl_tmp_114.xy - gsl_tmp_113 * gsl_tmp_115);
  let gsl_tmp_110 = 0.45 + 0.4 * gsl_tmp_109.z;
  let gsl_tmp_111 = style.z * (0.34 + 1.2 * (min(gsl_tmp_110, 1.0) - 0.45));
  let gsl_tmp_117 = gsl_tmp_111 / (1.0 + gsl_tmp_115 / 60.0);
  let gsl_tmp_118 = (gsl_tmp_114.z - gsl_tmp_111 * 60.0 * log(1.0 + gsl_tmp_115 / 60.0)) / sqrt(1.0 + gsl_tmp_117 * gsl_tmp_117);
  let gsl_tmp_119 = sqrt(gsl_tmp_116 * gsl_tmp_116 + gsl_tmp_118 * gsl_tmp_118);
  return select(gsl_tmp_108, min(gsl_tmp_108, (gsl_tmp_119 + 4.0 * max(gsl_tmp_119 - 12.5 * gsl_tmp_110, 0.0)) / (ringWidth * branchRate)), 5.0 < style.x);
}

fn wood_interTime(local: vec3<f32>, whorl: f32, ringWidth: f32, spacing: f32, branchRate: f32, seed: f32, count: f32, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_120 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 101.0 + seed, 3.0 * whorl + 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 101.0 + seed, 3.0 * whorl + 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_122 = local - vec3<f32>(0.0, 0.0, (whorl + 0.2 + 0.6 * gsl_tmp_120.y) * spacing);
  let gsl_tmp_121 = vec2<f32>(cos(6.2831855 * gsl_tmp_120.x), sin(6.2831855 * gsl_tmp_120.x));
  let gsl_tmp_123 = max(dot(gsl_tmp_122.xy, gsl_tmp_121), 0.0);
  let gsl_tmp_124 = length(gsl_tmp_122.xy - gsl_tmp_121 * gsl_tmp_123);
  let gsl_tmp_125 = (gsl_tmp_122.z - 0.1 * gsl_tmp_123) / 1.005;
  let gsl_tmp_126 = sqrt(gsl_tmp_124 * gsl_tmp_124 + gsl_tmp_125 * gsl_tmp_125);
  let gsl_tmp_127 = select(1000.0, min(1000.0, (gsl_tmp_126 + 4.0 * max(gsl_tmp_126 - 3.5 - 2.0 * gsl_tmp_120.z, 0.0)) / (ringWidth * branchRate)), 0.0 < count);
  let gsl_tmp_128 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(whorl, whorl + 114.0 + seed, 3.0 * whorl + 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(whorl, whorl + 114.0 + seed, 3.0 * whorl + 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_130 = local - vec3<f32>(0.0, 0.0, (whorl + 0.2 + 0.6 * gsl_tmp_128.y) * spacing);
  let gsl_tmp_129 = vec2<f32>(cos(6.2831855 * gsl_tmp_128.x), sin(6.2831855 * gsl_tmp_128.x));
  let gsl_tmp_131 = max(dot(gsl_tmp_130.xy, gsl_tmp_129), 0.0);
  let gsl_tmp_132 = length(gsl_tmp_130.xy - gsl_tmp_129 * gsl_tmp_131);
  let gsl_tmp_133 = (gsl_tmp_130.z - 0.1 * gsl_tmp_131) / 1.005;
  let gsl_tmp_134 = sqrt(gsl_tmp_132 * gsl_tmp_132 + gsl_tmp_133 * gsl_tmp_133);
  return select(gsl_tmp_127, min(gsl_tmp_127, (gsl_tmp_134 + 4.0 * max(gsl_tmp_134 - 3.5 - 2.0 * gsl_tmp_128.z, 0.0)) / (ringWidth * branchRate)), 1.0 < count);
}

fn wood_knotTime(local: vec3<f32>, ringWidth: f32, spacing: f32, branchRate: f32, seed: f32, style: vec4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let radius = length(local.xy);
  let middle = floor((local.z - 0.4 * style.z * radius) / spacing + 0.5);
  let lower = min(wood_whorlTime(local, middle - 1.0, ringWidth, spacing, branchRate, seed, style, noiseTex), wood_whorlTime(local, middle, ringWidth, spacing, branchRate, seed, style, noiseTex));
  let whorls = min(lower, wood_whorlTime(local, middle + 1.0, ringWidth, spacing, branchRate, seed, style, noiseTex));
  let between = select(1000.0, wood_interTime(local, floor((local.z - 0.1 * radius) / spacing), ringWidth, spacing, branchRate, seed, style.y, noiseTex), style.y > 0.0);
  return min(whorls, between);
}

fn wood_powerSmoothMin(a: f32, b: f32, k: f32, noiseTex: texture_3d<f32>) -> f32 {
  return a * b / pow(pow(a, k) + pow(b, k), 1.0 / k);
}

fn wood_knottedGrowthAt(p: vec3<f32>, axisOrigin: vec3<f32>, axisDir: vec3<f32>, ringWidth: f32, knotSpacing: f32, knotStyle: vec4<f32>, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let stemTime = wood_stemGrowthAt(p, axisOrigin, axisDir, ringWidth, noiseTex).x;
  let branchTime = wood_knotTime(map_axisAt(p, axisOrigin, axisDir, noiseTex), ringWidth, max(knotSpacing, 1.0), 0.45, 7.0, knotStyle, noiseTex);
  return vec2<f32>(wood_powerSmoothMin(max(stemTime, 0.001), branchTime, 4.0, noiseTex), texture_mask(stemTime - branchTime * 1.15, 0.0, 0.6, noiseTex));
}

fn wood_ringBoundary(k: f32, seed: f32, noiseTex: texture_3d<f32>) -> f32 {
  return k + (fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(k, k + seed, 2.0 * k + 17.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(k, k + seed, 2.0 * k + 17.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.5;
}

fn wood_ringAt(u: f32, seed: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let k = floor(u);
  let b1 = wood_ringBoundary(k + 1.0, seed, noiseTex);
  let above = u >= b1;
  let b0 = wood_ringBoundary(k, seed, noiseTex);
  let below = u < b0;
  let ringIndex = select(select(k, k + 1.0, above), k - 1.0, below);
  let bLow = wood_ringBoundary(k - 1.0, seed, noiseTex);
  let inner = select(select(b0, b1, above), bLow, below);
  let bHigh = wood_ringBoundary(k + 2.0, seed, noiseTex);
  let outer = select(select(b1, bHigh, above), b0, below);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = ringIndex;
  gsl_result.v_2 = (u - inner) / (outer - inner);
  return gsl_result;
}

fn wood_fibreShade(column: vec2<f32>, z: f32, noiseTex: texture_3d<f32>) -> f32 {
  let run = z / 8.0 + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(column.x, column.y, column.x - column.y + 7919.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(column.x, column.y, column.x - column.y + 7919.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let index = floor(run);
  let here = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(column.x, column.y, index))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(column.x, column.y, index)) / 64.0)) & vec3<i32>(63), 0).rgb).x;
  let next = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(column.x, column.y, index + 1.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(column.x, column.y, index + 1.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x;
  return here + (next - here) * texture_mask(run - index, 0.925, 0.075, noiseTex);
}

fn wood_fibreFile(file: f32, radius: f32, z: f32, noiseTex: texture_3d<f32>) -> f32 {
  let q = radius / 0.45 + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(file, file + 3.0, 2.0 * file + 7919.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(file, file + 3.0, 2.0 * file + 7919.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5;
  let cell = floor(q);
  let here = wood_fibreShade(vec2<f32>(file, cell), z, noiseTex);
  let t = q - cell;
  return here + (wood_fibreShade(vec2<f32>(file, cell + 1.0), z, noiseTex) - here) * t * t * (3.0 - 2.0 * t);
}

fn wood_poreZoneShare(poreZone: f32, ringWidth: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(0.3, min(poreZone / ringWidth, 0.3), poreZone > 0.0);
}

fn wood_poreSizeShare(progress: f32, ringPorosity: f32, poreZone: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let semiSize = 1.0 - 0.5 * progress;
  let earlywoodRamp = 1.0 - texture_mask(progress, poreZone, poreZone, noiseTex);
  let ringSize = 0.35 + 0.65 * earlywoodRamp;
  let diffuseSize = 1.0 - 0.15 * progress;
  let size = select(semiSize + (ringSize - semiSize) * (2.0 * ringPorosity - 1.0), diffuseSize + (semiSize - diffuseSize) * 2.0 * ringPorosity, ringPorosity < 0.5);
  return vec2<f32>(size, 0.35 + 0.45 * max(2.0 * ringPorosity - 1.0, 0.0));
}

fn wood_poreDistanceAt(local: vec3<f32>, progress: f32, poreSize: f32, ringPorosity: f32, poreZone: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let poreCellSize = max(poreSize, 0.01);
  let poreQ = local.xy / poreCellSize;
  let poreBase = floor(poreQ - 0.5);
  let gsl_tmp_135 = poreBase + vec2<f32>(0.0, 0.0);
  let gsl_tmp_136 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_135.x, gsl_tmp_135.y, gsl_tmp_135.x - gsl_tmp_135.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_135.x, gsl_tmp_135.y, gsl_tmp_135.x - gsl_tmp_135.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_137 = floor(gsl_tmp_136);
  let gsl_tmp_138 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_135.x, gsl_tmp_135.y, gsl_tmp_137))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_135.x, gsl_tmp_135.y, gsl_tmp_137)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let sizeShare = wood_poreSizeShare(progress, ringPorosity, poreZone, noiseTex);
  let size = sizeShare.x;
  let poreShare = sizeShare.y;
  let gsl_tmp_139 = select(10.0, min(10.0, length(vec2<f32>(length(poreQ - (gsl_tmp_135 + gsl_tmp_138.xy)), max(abs(gsl_tmp_136 - gsl_tmp_137 - 0.5) - 0.2 - 0.25 * gsl_tmp_138.x, 0.0) * 8.0)) / size), gsl_tmp_138.z < poreShare);
  let gsl_tmp_140 = poreBase + vec2<f32>(1.0, 0.0);
  let gsl_tmp_141 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_140.x, gsl_tmp_140.y, gsl_tmp_140.x - gsl_tmp_140.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_140.x, gsl_tmp_140.y, gsl_tmp_140.x - gsl_tmp_140.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_142 = floor(gsl_tmp_141);
  let gsl_tmp_143 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_140.x, gsl_tmp_140.y, gsl_tmp_142))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_140.x, gsl_tmp_140.y, gsl_tmp_142)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_144 = select(gsl_tmp_139, min(gsl_tmp_139, length(vec2<f32>(length(poreQ - (gsl_tmp_140 + gsl_tmp_143.xy)), max(abs(gsl_tmp_141 - gsl_tmp_142 - 0.5) - 0.2 - 0.25 * gsl_tmp_143.x, 0.0) * 8.0)) / size), gsl_tmp_143.z < poreShare);
  let gsl_tmp_145 = poreBase + vec2<f32>(0.0, 1.0);
  let gsl_tmp_146 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_145.x, gsl_tmp_145.y, gsl_tmp_145.x - gsl_tmp_145.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_145.x, gsl_tmp_145.y, gsl_tmp_145.x - gsl_tmp_145.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_147 = floor(gsl_tmp_146);
  let gsl_tmp_148 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_145.x, gsl_tmp_145.y, gsl_tmp_147))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_145.x, gsl_tmp_145.y, gsl_tmp_147)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_149 = select(gsl_tmp_144, min(gsl_tmp_144, length(vec2<f32>(length(poreQ - (gsl_tmp_145 + gsl_tmp_148.xy)), max(abs(gsl_tmp_146 - gsl_tmp_147 - 0.5) - 0.2 - 0.25 * gsl_tmp_148.x, 0.0) * 8.0)) / size), gsl_tmp_148.z < poreShare);
  let gsl_tmp_150 = poreBase + vec2<f32>(1.0, 1.0);
  let gsl_tmp_151 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_150.x, gsl_tmp_150.y, gsl_tmp_150.x - gsl_tmp_150.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_150.x, gsl_tmp_150.y, gsl_tmp_150.x - gsl_tmp_150.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_152 = floor(gsl_tmp_151);
  let gsl_tmp_153 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_150.x, gsl_tmp_150.y, gsl_tmp_152))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_150.x, gsl_tmp_150.y, gsl_tmp_152)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let nearestPore = select(gsl_tmp_149, min(gsl_tmp_149, length(vec2<f32>(length(poreQ - (gsl_tmp_150 + gsl_tmp_153.xy)), max(abs(gsl_tmp_151 - gsl_tmp_152 - 0.5) - 0.2 - 0.25 * gsl_tmp_153.x, 0.0) * 8.0)) / size), gsl_tmp_153.z < poreShare);
  return vec2<f32>(nearestPore, size);
}

fn wood_resinCanal(local: vec3<f32>, share: f32, noiseTex: texture_3d<f32>) -> f32 {
  let q = local.xy / 1.2 + 0.382;
  let cell = floor(q);
  let h = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cell.x, cell.y, cell.x - cell.y + 9173.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(cell.x, cell.y, cell.x - cell.y + 9173.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let run = floor(local.z / 30.0 + h.z);
  let held = select(0.0, 1.0, fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cell.x, cell.y, run))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(cell.x, cell.y, run)) / 64.0)) & vec3<i32>(63), 0).rgb).x < share);
  let across = length(q - cell - 0.2 - h.xy * 0.6) * 1.2;
  return held * (1.0 - texture_mask(across, 0.06, 0.02, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.045, 0.015, noiseTex));
}

fn wood_solidWood(p: vec3<f32>, axisOrigin: vec3<f32>, axisDir: vec3<f32>, ringWidth: f32, latewoodFraction: f32, latewoodSoftness: f32, earlyColor: vec3<f32>, lateColor: vec3<f32>, poreSize: f32, ringPorosity: f32, poreColor: vec3<f32>, rayStrength: f32, raySize: f32, knotSpacing: f32, knotStyle: vec4<f32>, knotColor: vec3<f32>, heartRings: f32, sapColor: vec3<f32>, resinCanals: f32, streak: f32, streakColor: vec3<f32>, ringVariation: f32, bandStrength: f32, bandSpacing: f32, poreZone: f32, sapTransition: f32, lineColor: vec3<f32>, lineStrength: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let earlyLab = color_rgb2oklab(earlyColor, noiseTex);
  let grown = select(wood_stemGrowthAt(p, axisOrigin, axisDir, ringWidth, noiseTex), wood_knottedGrowthAt(p, axisOrigin, axisDir, ringWidth, knotSpacing, knotStyle, noiseTex), knotSpacing > 0.0);
  let gsl_result = wood_ringAt(grown.x, 3.0, noiseTex);
  let progress = gsl_result.v_2;
  let ringIndex = gsl_result.v_1;
  let ringHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(ringIndex, ringIndex + 11.0, 3.0 * ringIndex + 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(ringIndex, ringIndex + 11.0, 3.0 * ringIndex + 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let ringShift = (ringHash.x - 0.5) * 0.3 * ringVariation;
  let ringContrast = 1.0 + (ringHash.y - 0.5) * 2.0 * ringVariation;
  let latewood = texture_mask(progress, 1.0 - latewoodFraction + ringShift, latewoodSoftness, noiseTex) * ringContrast;
  let ringTint = (fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(ringIndex, ringIndex + 5.0, 2.0 * ringIndex + 29.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(ringIndex, ringIndex + 5.0, 2.0 * ringIndex + 29.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.08;
  let local = map_axisAt(p, axisOrigin, axisDir, noiseTex);
  let radius = length(local.xy);
  var gsl_detailValue = 0.0;
  if (max(gsl_pw, gsl_detailPw) < 0.075) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let fileTurn = (atan2(local.y, local.x) + 3.1415927) / 6.2831855 * 3351.0 - 0.5;
    let fileBase = floor(fileTurn);
    let fileShadeA = wood_fibreFile(gsl_floored(fileBase, 3351.0), radius, local.z, noiseTex);
    let fileShadeB = wood_fibreFile(gsl_floored((fileBase + 1.0), 3351.0), radius, local.z, noiseTex);
    let fileT = fileTurn - fileBase;
    let fibreBlend = fileShadeA + (fileShadeB - fileShadeA) * fileT * fileT * (3.0 - 2.0 * fileT);
    gsl_detailValue = (fibreBlend - 0.5) * 0.08 * (1.0 - 0.6 * latewood) * texture_fadeDetail(min(0.15, 6.2831855 * radius / 3351.0), noiseTex);
    gsl_pw = gsl_detailOuterPw;
  }
  let fibre = gsl_detailValue;
  let rayCells = floor(250.0 / max(raySize, 0.05) + 0.001);
  let turn = (atan2(local.y, local.x) + 3.1415927) / 6.2831855 * rayCells;
  let rayIndex = gsl_floored(floor(turn), rayCells);
  let rayZ = local.z / (12.0 * raySize) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(rayIndex, rayIndex + 17.0, 3.0 * rayIndex + 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(rayIndex, rayIndex + 17.0, 3.0 * rayIndex + 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let rayHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(rayIndex, floor(rayZ), rayIndex + 401.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(rayIndex, floor(rayZ), rayIndex + 401.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let rayAcross = abs(turn - floor(turn) - 0.2 - 0.6 * rayHash.x) * 6.2831855 * radius / rayCells;
  let rayAlong = abs(rayZ - floor(rayZ) - 0.5) * 2.0 / (0.35 + 0.65 * rayHash.y);
  let ray = rayStrength * select(0.0, 1.0, rayHash.z < 0.5) * (1.0 - texture_mask(rayAcross, 0.15 * raySize, 0.05 * raySize, noiseTex)) * (1.0 - texture_mask(rayAlong, 1.0, 0.1, noiseTex)) * texture_fadeDetail(0.3 * raySize, noiseTex);
  let bandPhase = radius / max(bandSpacing, 0.01) + noise_gradientLattice(vec3<f32>(local.x / 0.22, local.y / 0.22, local.z / 6.0), noiseTex);
  let bandLine = 1.0 - texture_mask(abs(bandPhase - floor(bandPhase) - 0.5), 0.2, 0.08, noiseTex);
  let bandBreak = texture_mask(noise_gradientLattice(vec3<f32>(local.x / 0.6 + 31.0, local.y / 0.6 + 7.0, local.z / 8.0), noiseTex), -0.3, 0.2, noiseTex);
  let bandFine = texture_fadeDetail(0.5 * max(bandSpacing, 0.01), noiseTex);
  let bandShade = (bandLine * bandBreak - 0.25) * bandFine + (0.8 * bandBreak - 0.5) * (1.0 - bandFine) * (1.0 - texture_mask(gsl_pw, 0.225, 0.075, noiseTex));
  let band = select(0.0, bandStrength * latewood * bandShade, bandStrength > 0.0);
  let wood = earlyLab + (color_rgb2oklab(lateColor, noiseTex) - earlyLab) * latewood + vec3<f32>(ringTint + fibre + 0.1 * ray + band, 0.0, 0.0);
  let sapwood = texture_mask(grown.x, heartRings, 0.5 * sapTransition + 0.05, noiseTex);
  let sapped = wood + (color_rgb2oklab(sapColor, noiseTex) - earlyLab) * sapwood;
  let boundaryLine = lineStrength * (1.0 - texture_mask(abs(grown.x - heartRings + 0.3), 0.3, 0.1, noiseTex));
  let lined = sapped + (color_rgb2oklab(lineColor, noiseTex) - sapped) * boundaryLine;
  let streakNoise = select(0.0, noise_gradientLattice(vec3<f32>(grown.x * 0.45, local.z / 120.0 + local.x * 0.01, local.y * 0.01 + local.z * 0.002), noiseTex), streak > 0.0);
  let streaks = streak * texture_mask(streakNoise, 0.15, 0.12, noiseTex);
  let streaked = lined + (color_rgb2oklab(streakColor, noiseTex) - lined) * streaks;
  let poreCellSize = max(poreSize, 0.01);
  let poreScale = wood_poreSizeShare(progress, ringPorosity, wood_poreZoneShare(poreZone, ringWidth, noiseTex), noiseTex).x;
  var gsl_detailValue_2 = select(0.0, 0.05 * poreScale * poreScale, poreSize > 0.0);
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (0.56 * poreCellSize)) { let gsl_detailOuterPw_2 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let poreHere = wood_poreDistanceAt(local, progress, poreSize, ringPorosity, wood_poreZoneShare(poreZone, ringWidth, noiseTex), noiseTex);
    let poreDistance = poreHere.x;
    let poreFade = texture_fadeDetail(poreCellSize * 0.56 * poreHere.y, noiseTex);
    gsl_detailValue_2 = select(0.0, (1.0 - texture_mask(poreDistance, 0.28, 0.06, noiseTex)) * poreFade + 0.05 * poreHere.y * poreHere.y * (1.0 - poreFade), poreSize > 0.0);
    gsl_pw = gsl_detailOuterPw_2;
  }
  let pore = gsl_detailValue_2;
  let resin = select(0.0, wood_resinCanal(local, resinCanals, noiseTex) * latewood, resinCanals > 0.0);
  let withPores = streaked + (color_rgb2oklab(poreColor, noiseTex) - streaked) * 0.85 * pore - vec3<f32>(0.2 * resin, 0.0, 0.0);
  let inKnot = grown.y;
  return color_oklab2rgb(withPores + (color_rgb2oklab(knotColor, noiseTex) - withPores) * 0.8 * inKnot, noiseTex);
}

fn wood_woodStandIn(earlyColor: vec3<f32>, lateColor: vec3<f32>, latewoodFraction: f32, weather: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let ringsLab = color_rgb2oklab(color_mixOklab(earlyColor, lateColor, latewoodFraction, noiseTex), noiseTex);
  return color_oklab2rgb(ringsLab + (vec3<f32>(0.6675783, 0.0015017033, 0.01586319) - ringsLab) * 0.375 * max(weather, 0.0), noiseTex);
}

fn wood_woodChecks(local: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let fissure = abs(noise_gradientLattice(vec3<f32>(local.x / 6.0, local.y / 6.0, local.z / 60.0), noiseTex));
  let patchy = texture_mask(noise_gradientLattice(vec3<f32>(local.x / 30.0 + 17.0, local.y / 30.0, local.z / 150.0), noiseTex), 0.25, 0.08, noiseTex);
  return (1.0 - texture_mask(fissure, 0.08, 0.025, noiseTex)) * patchy;
}

fn wood_weatheredColor(p: vec3<f32>, axisOrigin: vec3<f32>, axisDir: vec3<f32>, freshColor: vec3<f32>, amount: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let freshLab = color_rgb2oklab(freshColor, noiseTex);
  let checks = wood_woodChecks(map_axisAt(p, axisOrigin, axisDir, noiseTex), noiseTex) * amount;
  return color_oklab2rgb(freshLab + (vec3<f32>(0.6675783, 0.0015017033, 0.01586319) - freshLab) * 0.375 * amount - vec3<f32>(0.25 * checks, 0.0, 0.0), noiseTex);
}

fn wood_colorFromRowsFinished(p: vec3<f32>, rings: vec4<f32>, early: vec4<f32>, late: vec4<f32>, pore: vec4<f32>, knot: vec4<f32>, sap: vec4<f32>, streak: vec4<f32>, surface: vec4<f32>, age: vec4<f32>, bands: vec4<f32>, boundary: vec4<f32>, weeks: vec4<f32>, uptake: vec4<f32>, axisOrigin: vec3<f32>, axisDir: vec3<f32>, look: vec3<f32>, knotSpacing: f32, knotStyle: vec4<f32>, barkRadius: f32, normal: vec3<f32>, prep: mat3x4<f32>, stain: mat2x4<f32>, topcoat: mat2x4<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let aged = clamp(look.y, 0.0, 1.0);
  let aging = (vec3<f32>(1.0, 1.0, 1.0) + (weeks.xyz - vec3<f32>(1.0, 1.0, 1.0)) * (1.0 - pow(1.0 - aged, 37.0))) * (vec3<f32>(1.0, 1.0, 1.0) + (age.xyz - vec3<f32>(1.0, 1.0, 1.0)) * aged);
  let prepHead = prep[0];
  let charDepth = select(0.0, prepHead.y, prepHead.x > 4.5);
  let earlyChar = 1.4 * charDepth;
  let stainLook = stain[0];
  let chip = color_linearize(stainLook.xyz, noiseTex) / 1.042;
  let stainGain = min(vec3<f32>(pow(chip.x, 0.71275836), pow(chip.y, 0.71275836), pow(chip.z, 0.71275836)) / vec3<f32>(0.58701634, 0.4064483, 0.24286681), vec3<f32>(1.0, 1.0, 1.0));
  let n = normal / max(length(normal), 0.000001);
  let endGrain = dot(n, axisDir) / max(length(axisDir), 0.000001);
  let local = map_axisAt(p, axisOrigin, axisDir, noiseTex);
  let blotch = texture_mask(noise_gradientLattice(vec3<f32>(local.x / 6.0, local.y / 6.0, local.z / 40.0), noiseTex), 0.1, 0.25, noiseTex);
  let stainKind = stain[1];
  let pigment = stainKind.x;
  let taken = stainLook.w * (1.0 + 1.5 * endGrain * endGrain) * (1.0 + 0.6 * uptake.w * blotch * (1.0 - endGrain * endGrain)) * select(1.0, wood_prepCoarseness(prep, noiseTex), pigment > 0.5);
  let earlyTaken = taken * uptake.x;
  let coatFilm = topcoat[1];
  let wetting = coatFilm.z;
  let coatLook = topcoat[0];
  let tint = color_linearize(coatLook.xyz, noiseTex) * (vec3<f32>(1.0, 1.0, 1.0) + vec3<f32>(0.0, -0.1, -0.3) * coatFilm.w * aged);
  let lateChar = 0.6 * charDepth;
  let stand = wood_woodStandIn(wood_finishedColor(early.xyz, aging, earlyChar, stainGain, earlyTaken, wetting, tint, noiseTex), wood_finishedColor(late.xyz, aging, lateChar, stainGain, taken * uptake.y, wetting, tint, noiseTex), rings.y, look.z, noiseTex);
  var gsl_simpleValue = stand;
  if (gsl_simpleDetail > 0.0) {
    let poreTaken = taken * select(uptake.x, uptake.z, pigment > 0.5);
    let knots = select(knotSpacing, surface.z, knotSpacing < 0.0);
    let heartRings = barkRadius / rings.x - sap.w + select(1000000.0, 0.0, barkRadius > 0.0);
    let sapAging = vec3<f32>(1.0, 1.0, 1.0) + (age.xyz - vec3<f32>(1.0, 1.0, 1.0)) * aged;
    let fresh = wood_solidWood(p, axisOrigin, axisDir, rings.x, rings.y, rings.z, wood_finishedColor(early.xyz, aging, earlyChar, stainGain, earlyTaken, wetting, tint, noiseTex), wood_finishedColor(late.xyz, aging, lateChar, stainGain, taken * uptake.y, wetting, tint, noiseTex), early.w, late.w, wood_finishedColor(pore.xyz, aging, earlyChar, stainGain, poreTaken, wetting, tint, noiseTex), pore.w, knot.w, knots, knotStyle, wood_finishedColor(knot.xyz, aging, lateChar, stainGain, earlyTaken, wetting, tint, noiseTex), heartRings, wood_finishedColor(sap.xyz, sapAging, earlyChar, stainGain, earlyTaken, wetting, tint, noiseTex), surface.w, streak.w, wood_finishedColor(streak.xyz, aging, earlyChar, stainGain, earlyTaken, wetting, tint, noiseTex), rings.w, bands.x, bands.y, bands.z, bands.w, wood_finishedColor(boundary.xyz, aging, earlyChar, stainGain, earlyTaken, wetting, tint, noiseTex), boundary.w, noiseTex);
    gsl_simpleValue = mix(gsl_simpleValue, select(fresh, wood_weatheredColor(p, axisOrigin, axisDir, fresh, look.z, noiseTex), look.z > 0.0), gsl_simpleDetail);
  }
  let wood = gsl_simpleValue;
  return wood;
}

fn wood_speciesColor(p: vec3<f32>, number: f32, axisOrigin: vec3<f32>, axisDir: vec3<f32>, look: vec3<f32>, knotSpacing: f32, barkRadius: f32, normal: vec3<f32>, prep: mat3x4<f32>, stain: mat2x4<f32>, topcoat: mat2x4<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return wood_colorFromRowsFinished(p, wood_speciesRings(number, noiseTex), wood_speciesEarlywood(number, noiseTex), wood_speciesLatewood(number, noiseTex), wood_speciesPoreRow(number, noiseTex), wood_speciesKnotRow(number, noiseTex), wood_speciesSapwood(number, noiseTex), wood_speciesStreakRow(number, noiseTex), wood_speciesSurfaceRow(number, noiseTex), wood_speciesAgeRow(number, noiseTex), wood_speciesBandRow(number, noiseTex), wood_speciesBoundaryRow(number, noiseTex), wood_speciesWeeksRow(number, noiseTex), wood_speciesStainRow(number, noiseTex), axisOrigin, axisDir, look, knotSpacing, wood_speciesKnotStyle(number, noiseTex), barkRadius, normal, prep, stain, topcoat, noiseTex);
}

fn wood_poreSurfaceAt(local: vec3<f32>, progress: f32, ringWidth: f32, poreSize: f32, ringPorosity: f32, poreZone: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let poreHere = wood_poreDistanceAt(local, progress, poreSize, ringPorosity, wood_poreZoneShare(poreZone, ringWidth, noiseTex), noiseTex);
  let u = min(poreHere.x / 0.28, 1.0);
  let fade = texture_fadeDetail(poreSize * 0.56 * poreHere.y, noiseTex);
  return vec2<f32>(-0.28 * poreSize * poreHere.y * (1.0 - u * u) * fade, (1.0 - texture_mask(poreHere.x, 0.28, 0.06, noiseTex)) * fade);
}

fn wood_crackPoints(local: vec3<f32>, depth: f32, noiseTex: texture_3d<f32>) -> mat2x4<f32> {
  let block = 3.0 + 3.0 * depth;
  return mat2x4<f32>(vec4<f32>(local.x / (3.0 * block), local.y / (3.0 * block), local.z / (0.6 * block), 0.0), vec4<f32>(local.x / (0.6 * block) + 17.0, local.y / (0.6 * block) + 5.0, local.z / (3.0 * block), 0.0));
}

fn wood_gritSize(grit: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 16.0 / max(grit, 24.0);
}

fn wood_sandingPoint(local: vec3<f32>, grit: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(local.x, local.y, local.z / 40.0) / (1.6 * wood_gritSize(grit, noiseTex));
}

fn wood_fibrePoint(local: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(local.x / 0.15, local.y / 0.15, local.z / 2.0);
}

fn wood_sandedHeight(scratches: f32, grit: f32, noiseTex: texture_3d<f32>) -> f32 {
  let d = wood_gritSize(grit, noiseTex);
  let groove = max(1.0 - abs(scratches) / 0.15, 0.0);
  return -0.12 * d * groove * groove * texture_fadeDetail(d, noiseTex);
}

fn wood_sawnHeight(s: f32, noisiness: f32, wander: f32, washboard: f32, fibres: f32, noiseTex: texture_3d<f32>) -> f32 {
  let depth = 0.18 + 0.62 * noisiness;
  let pitch = 2.0 + 2.0 * noisiness;
  let q = s / pitch + 2.0 * noisiness * wander;
  let k = floor(q);
  let uneven = 1.0 + (fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(k, k + 29.0, 2.0 * k + 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(k, k + 29.0, 2.0 * k + 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * (0.4 + 1.2 * noisiness);
  let t = q - k;
  let marks = depth * uneven * (abs(2.0 * t - 1.0) - 1.0) * texture_fadeDetail(pitch, noiseTex);
  let flutter = (0.1 + 0.4 * noisiness) * washboard;
  let fuzz = (0.02 + 0.08 * noisiness) * fibres * (1.0 - texture_mask(gsl_pw, 0.05625, 0.01875, noiseTex));
  let slope = 2.0 * depth * (1.2 + 0.6 * noisiness) / pitch;
  return (marks + flutter + fuzz) / sqrt(1.0 + slope * slope);
}

fn wood_tableSawnHeight(sArbor: f32, bArbor: f32, jitter: f32, feedRate: f32, noiseTex: texture_3d<f32>) -> f32 {
  let turn = (sArbor + sqrt(max(16129.0 - bArbor * bArbor, 0.0))) / 1.2 + 0.3 * feedRate;
  return 0.5 * jitter * cos(6.2831855 * turn) * (1.0 - texture_mask(gsl_pw, 0.45, 0.15, noiseTex));
}

fn wood_planedHeight(s: f32, scallop: f32, vibration: f32, noiseTex: texture_3d<f32>) -> f32 {
  let depth = scallop * scallop / 304.0;
  let q = s / max(scallop, 0.05);
  let k = floor(q);
  let uneven = 1.0 + (fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(k, k + 41.0, 2.0 * k + 11.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(k, k + 41.0, 2.0 * k + 11.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.6;
  let t = q - k;
  let wave = 0.012 * vibration;
  return depth * uneven * (4.0 * (t - 0.5) * (t - 0.5) - 1.0) * texture_fadeDetail(max(scallop, 0.05), noiseTex) + wave;
}

fn wood_wornProfile(progress: f32, noiseTex: texture_3d<f32>) -> f32 {
  return min(progress / 0.88, (1.0 - progress) / 0.12);
}

fn wood_torchedHeight(progress: f32, ringWidth: f32, depth: f32, acrossCrack: f32, alongCrack: f32, noiseTex: texture_3d<f32>) -> f32 {
  let wear = 0.3 * min(depth, 2.0);
  let slope = wear / (0.12 * ringWidth);
  let block = 3.0 + 3.0 * depth;
  let cracks = (1.0 - texture_mask(min(abs(acrossCrack), abs(alongCrack)), 0.05, 0.025, noiseTex)) * texture_mask(depth, 1.0, 0.4, noiseTex) * texture_fadeDetail(0.1 * block, noiseTex);
  return wear / sqrt(1.0 + slope * slope) * (wood_wornProfile(progress, noiseTex) - 0.5) - 0.4 * depth * cracks;
}

fn wood_prepHeight(p: vec3<f32>, prep: mat3x4<f32>, normal: vec3<f32>, local: vec3<f32>, progress: f32, ringWidth: f32, build: f32, noiseTex: texture_3d<f32>) -> f32 {
  let head = prep[0];
  let size = head.y;
  let cracks = wood_crackPoints(local, size, noiseTex);
  let crackAcross = cracks[0];
  let kind = head.x;
  let firstAt = select(select(select(select(crackAcross.xyz, p / 12.0 + 29.0, kind < 4.5), p / 40.0, kind < 3.5), wood_fibrePoint(local, noiseTex), kind < 2.5), wood_sandingPoint(local, size, noiseTex), kind < 1.5);
  let first = noise_gradientLattice(firstAt, noiseTex);
  let sanding = select(0.0, wood_sandedHeight(first, size, noiseTex) * material_filmKeeps(wood_gritSize(size, noiseTex), build, noiseTex), kind < 1.5);
  let feed = prep[1];
  let along = feed.xyz / max(length(feed.xyz), 0.000001);
  let s = dot(p, along);
  let amount = head.z;
  let crackAlong = cracks[1];
  let secondAt = select(crackAlong.xyz, p / 30.0, kind < 4.5);
  let second = noise_gradientLattice(secondAt, noiseTex);
  let third = noise_gradientLattice(p / 40.0 + 13.0, noiseTex);
  let sawing = select(0.0, select(0.0, wood_sawnHeight(s, amount, second, third, first, noiseTex) * material_filmKeeps(1.0 + 3.0 * amount, build, noiseTex), kind < 2.5), kind > 1.5);
  let blade = prep[2];
  let rise = blade.xyz - along * dot(blade.xyz, along);
  let aboveCentre = blade.w + dot(p, rise / max(length(rise), 0.000001));
  let ripping = select(0.0, select(0.0, wood_tableSawnHeight(s, aboveCentre, amount, first, noiseTex) * material_filmKeeps(1.2, build, noiseTex), kind < 3.5), kind > 2.5);
  let planing = select(0.0, select(0.0, wood_planedHeight(s, size, first, noiseTex) * material_filmKeeps(size, build, noiseTex), kind < 4.5), kind > 3.5);
  let n = normal / max(length(normal), 0.000001);
  let inSurface = texture_mask(length(feed.xyz - n * dot(feed.xyz, n)) / max(length(feed.xyz), 0.000001), 0.5, 0.2, noiseTex);
  let burning = select(0.0, wood_torchedHeight(progress, ringWidth, size, first, second, noiseTex) * material_filmKeeps(0.3 * size, build, noiseTex), kind > 4.5);
  return sanding + (sawing + ripping + planing) * inSurface + burning;
}

fn wood_weatheredSurfaceAt(local: vec3<f32>, progress: f32, ringWidth: f32, relief: f32, rough: f32, amount: f32, poreHeight: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let wear = relief * amount;
  let slope = wear / (0.12 * ringWidth);
  let checks = wood_woodChecks(local, noiseTex) * amount;
  return vec2<f32>(rough + (0.9 - rough) * amount, wear / sqrt(1.0 + slope * slope) * (wood_wornProfile(progress, noiseTex) - 0.5) - 1.2 * checks + poreHeight);
}

fn wood_hiddenShare(size: f32, standIn: f32, noiseTex: texture_3d<f32>) -> f32 {
  return material_hiddenShare(texture_fadeDetail(size, noiseTex), standIn, noiseTex);
}

fn wood_prepSlopes(prep: mat3x4<f32>, normal: vec3<f32>, grainIn: f32, ringWidth: f32, build: f32, standIn: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let head = prep[0];
  let size = head.y;
  let feed = prep[1];
  let n = normal / max(length(normal), 0.000001);
  let inSurface = texture_mask(length(feed.xyz - n * dot(feed.xyz, n)) / max(length(feed.xyz), 0.000001), 0.5, 0.2, noiseTex);
  let planeKeeps = material_filmKeeps(size, build, noiseTex) * inSurface;
  let planing = 0.000119 * size * size * planeKeeps * planeKeeps * wood_hiddenShare(max(size, 0.05), standIn, noiseTex);
  let kind = head.x;
  let amount = head.z;
  let ripKeeps = material_filmKeeps(1.2, build, noiseTex) * inSurface;
  let ripping = 6.85 * amount * amount * ripKeeps * ripKeeps * wood_hiddenShare(1.2, standIn, noiseTex);
  let depth = 0.18 + 0.62 * amount;
  let pitch = 2.0 + 2.0 * amount;
  let spread = 0.4 + 1.2 * amount;
  let sawKeeps = material_filmKeeps(1.0 + 3.0 * amount, build, noiseTex) * inSurface;
  let steep = 2.0 * depth * (1.2 + 0.6 * amount) / pitch;
  let sawScale = sawKeeps * sawKeeps / (1.0 + steep * steep);
  let marks = 8.0 * depth / pitch * depth / pitch * (1.0 + spread * spread / 12.0) * sawScale * wood_hiddenShare(pitch, standIn, noiseTex);
  let alongFeed = select(select(select(select(0.0, planing, kind < 4.5), ripping, kind < 3.5), marks, kind < 2.5), 0.0, kind < 1.5);
  let wear = 0.3 * min(size, 2.0);
  let wearSlope = wear / (0.12 * ringWidth);
  let burnKeeps = material_filmKeeps(0.3 * size, build, noiseTex);
  let ringWear = 9.47 * wear * wear / (1.0 + wearSlope * wearSlope) / (ringWidth * ringWidth) * burnKeeps * burnKeeps * wood_hiddenShare(ringWidth, standIn, noiseTex);
  let fuzz = 22.2 * (0.02 + 0.08 * amount) * (0.02 + 0.08 * amount) * sawScale * wood_hiddenShare(0.15, standIn, noiseTex);
  let grit = wood_gritSize(size, noiseTex);
  let sandKeeps = material_filmKeeps(grit, build, noiseTex);
  let sanding = sandKeeps * sandKeeps * wood_hiddenShare(grit, standIn, noiseTex);
  let acrossGrain = select(select(select(0.0, grainIn * ringWear, kind > 4.5), fuzz, kind < 2.5), 0.106 * grainIn * sanding, kind < 1.5);
  let crackDepth = 0.4 * size * texture_mask(size, 1.0, 0.4, noiseTex);
  let block = 3.0 + 3.0 * size;
  let cracking = 118.0 * crackDepth / block * crackDepth / block * burnKeeps * burnKeeps * wood_hiddenShare(0.1 * block, standIn, noiseTex);
  let everyWay = select(select(0.0, (1.0 - grainIn) * ringWear + cracking, kind > 4.5), 0.101 * (1.0 - grainIn) * sanding, kind < 1.5);
  return vec3<f32>(alongFeed, acrossGrain, everyWay);
}

fn wood_poreSlopes__spec1(poreSize: f32, sizeShare: vec2<f32>, grainIn: f32, build: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let y = sizeShare.x;
  let cover = sizeShare.y * y * y;
  let keeps = material_filmKeeps(max(poreSize, 0.01), build, noiseTex);
  let hidden = select(0.0, cover * keeps * keeps * wood_hiddenShare(max(poreSize, 0.25) * 0.56 * y, 0.0, noiseTex), poreSize >= 0.25);
  return vec3<f32>(0.0, 0.38 * grainIn * hidden, 0.33 * (1.0 - grainIn) * hidden);
}

fn wood_poreSlopes__spec2(poreSize: f32, sizeShare: vec2<f32>, grainIn: f32, build: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let keeps = material_filmKeeps(max(poreSize, 0.01), build, noiseTex);
  let y = sizeShare.x;
  let hidden = select(0.0, 0.25 * keeps * keeps * wood_hiddenShare(max(poreSize, 0.25) * 0.56 * y, 1.0, noiseTex), poreSize >= 0.25);
  return vec3<f32>(0.0, 0.38 * grainIn * hidden, 0.33 * (1.0 - grainIn) * hidden);
}

fn wood_woodFinish(rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_make(rough, 0.0, 1.5, noiseTex);
}

fn wood_hiddenSlopeFinish(rough: f32, normal: vec3<f32>, grain: vec3<f32>, feed: vec3<f32>, slopes: vec3<f32>, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let base = rough * rough * rough * rough + slopes.z;
  let n = normal / max(length(normal), 0.000001);
  let feedIn = feed - n * dot(feed, n);
  let inPlane = grain - n * dot(grain, n);
  let inLength = length(inPlane);
  let u = select(normalize(cross(n, vec3<f32>(0.577, 0.577, 0.577)) + vec3<f32>(0.000001, 0.0, 0.0)), inPlane / max(inLength, 0.000001), inLength > 0.000001);
  let feedLength = max(length(feedIn), 0.000001);
  let fu = dot(feedIn, u) / feedLength;
  let tuu = base + slopes.x * fu * fu;
  let c = cross(n, u);
  let fc = dot(feedIn, c) / feedLength;
  let tcc = base + slopes.y + slopes.x * fc * fc;
  let mid = 0.5 * (tuu + tcc);
  let tuc = slopes.x * fu * fc;
  let half = sqrt(0.25 * (tuu - tcc) * (tuu - tcc) + tuc * tuc);
  let narrow = sqrt(max(mid - half, 0.000000000001));
  let wide = sqrt(mid + half);
  let angle = 0.5 * atan2(2.0 * tuc, tuu - tcc + 0.000000000001);
  let au = -sin(angle);
  let ac = cos(angle);
  let flip = select(select(select(1.0, -1.0, ac < 0.0), -1.0, au < -0.000001), 1.0, au > 0.000001);
  let stretch = (1.0 - clamp(narrow / wide, 0.1, 1.0)) / 0.9;
  return material_withLay(wood_woodFinish(sqrt(sqrt(narrow * wide)), noiseTex), (u * au + c * ac) * flip * min(stretch * 1000.0, 1.0), stretch, noiseTex);
}

fn wood_speciesMaterial(p: vec3<f32>, number: f32, axisOrigin: vec3<f32>, axisDir: vec3<f32>, look: vec3<f32>, knotSpacing: f32, normal: vec3<f32>, prep: mat3x4<f32>, topcoat: mat2x4<f32>, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let coatLook = topcoat[0];
  let coatSheen = material_roughnessFromGloss(coatLook.w, noiseTex);
  let prepHead = prep[0];
  let coatFilm = topcoat[1];
  let kind = coatFilm.x;
  let charred = select(0.0, select(0.0, min(prepHead.y, 1.0), prepHead.x > 4.5), kind < 0.5);
  let sheen = coatSheen + (0.85 - coatSheen) * charred;
  let grainDir = normalize(axisDir);
  let feed = prep[1];
  let unit = normal / max(length(normal), 0.000001);
  let grainIn = 1.0 - dot(grainDir, unit) * dot(grainDir, unit);
  let rings = wood_speciesRings(number, noiseTex);
  let build = coatFilm.y;
  let early = wood_speciesEarlywood(number, noiseTex);
  let poreSize = early.w;
  let stand = wood_hiddenSlopeFinish(sheen, normal, grainDir, feed.xyz, wood_prepSlopes(prep, normal, grainIn, rings.x, build, 1.0, noiseTex) + wood_poreSlopes__spec2(poreSize, vec2<f32>(1.0, 1.0), grainIn, build, noiseTex), noiseTex);
  let local = map_axisAt(p, axisOrigin, axisDir, noiseTex);
  let surface = wood_speciesSurfaceRow(number, noiseTex);
  let knots = select(knotSpacing, surface.z, knotSpacing < 0.0);
  let knotStyle = wood_speciesKnotStyle(number, noiseTex);
  let gsl_result_2 = wood_ringAt(select(wood_stemGrowthAt(p, axisOrigin, axisDir, rings.x, noiseTex), wood_knottedGrowthAt(p, axisOrigin, axisDir, rings.x, knots, knotStyle, noiseTex), knots > 0.0).x, 3.0, noiseTex);
  let progress = gsl_result_2.v_2;
  let late = wood_speciesLatewood(number, noiseTex);
  let bands = wood_speciesBandRow(number, noiseTex);
  var gsl_detailValue = vec2<f32>(0.0, 0.0);
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (0.56 * max(poreSize, 0.25))) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue = select(vec2<f32>(0.0, 0.0), wood_poreSurfaceAt(local, progress, rings.x, max(poreSize, 0.25), late.w, bands.z, noiseTex), poreSize >= 0.25);
    gsl_pw = gsl_detailOuterPw;
  }
  var gsl_simpleValue = stand;
  if (gsl_simpleDetail > 0.0) {
    let poreRelief = gsl_detailValue;
    let poreMatte = select(select(0.0, 0.5, kind < 1.5), 1.0, kind < 0.5);
    let poreRough = sheen + (0.9 - sheen) * poreRelief.y * poreMatte;
    let relief = poreRelief.x * material_filmKeeps(max(poreSize, 0.01), build, noiseTex) + wood_prepHeight(p, prep, normal, local, progress, rings.x, build, noiseTex);
    let finished = select(vec2<f32>(poreRough, relief), wood_weatheredSurfaceAt(local, progress, rings.x, surface.y, poreRough, look.z, relief, noiseTex), look.z > 0.0);
    let sizeShare = wood_poreSizeShare(progress, late.w, wood_poreZoneShare(bands.z, rings.x, noiseTex), noiseTex);
    let slopes = wood_prepSlopes(prep, normal, grainIn, rings.x, build, 0.0, noiseTex) + wood_poreSlopes__spec1(poreSize, sizeShare, grainIn, build, noiseTex);
    let gsl_simpleFull = material_withDisplacement(wood_hiddenSlopeFinish(finished.x, normal, grainDir, feed.xyz, slopes, noiseTex), finished.y, noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let surfaceMat = gsl_simpleValue;
  return surfaceMat;
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-480.01, -275.01, -0.01);
  let gsl_partHi = vec3<f32>(480.01, 275.01, 70.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let gsl_partLo_2 = vec3<f32>(-480.01, -275.01, -0.01);
  let gsl_partHi_2 = vec3<f32>(480.01, -174.99, 20.01);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = min(min(builtin_gcl_closure_boxCentered(p, vec3<f32>(-330.0, -225.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex), builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, -225.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex)), builtin_gcl_closure_boxCentered(p, vec3<f32>(330.0, -225.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex)) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let flat = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-480.01, -125.01, -0.01);
  let gsl_partHi_3 = vec3<f32>(480.01, -24.99, 20.01);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = min(min(builtin_gcl_closure_boxCentered(p, vec3<f32>(-330.0, -75.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex), builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, -75.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex)), builtin_gcl_closure_boxCentered(p, vec3<f32>(330.0, -75.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex)) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let quarter = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-380.01, 24.99, -0.01);
  let gsl_partHi_4 = vec3<f32>(380.01, 125.01, 70.01);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = min(min(builtin_gcl_closure_boxCentered_2(p, vec3<f32>(-330.0, 75.0, 35.0), vec3<f32>(100.0, 100.0, 70.0), noiseTex), builtin_gcl_closure_boxCentered_2(p, vec3<f32>(0.0, 75.0, 35.0), vec3<f32>(100.0, 100.0, 70.0), noiseTex)), builtin_gcl_closure_boxCentered_2(p, vec3<f32>(330.0, 75.0, 35.0), vec3<f32>(100.0, 100.0, 70.0), noiseTex)) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let ends = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-480.01, 174.99, -0.01);
  let gsl_partHi_5 = vec3<f32>(480.01, 275.01, 20.01);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = min(min(builtin_gcl_closure_boxCentered(p, vec3<f32>(-330.0, 225.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex), builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 225.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex)), builtin_gcl_closure_boxCentered(p, vec3<f32>(330.0, 225.0, 10.0), vec3<f32>(300.0, 100.0, 20.0), noiseTex)) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let old = gsl_part_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = min(min(flat, quarter), min(ends, old)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let planks = gsl_part;
  let gsl_s1 = gsl_shape(planks, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let material_wood_pineNumber = 14.0;
    let material_wood_walnutNumber = 4.0;
    let column = clamp(floor(p.x / 330.0 + 1.5), 0.0, 2.0);
    let material_wood_whiteOakNumber = 0.0;
    let number = select(select(material_wood_pineNumber, material_wood_walnutNumber, column < 1.5), material_wood_whiteOakNumber, column < 0.5);
    let x = (column - 1.0) * 330.0;
    let row = clamp(floor((p.y + 300.0) / 150.0), 0.0, 3.0);
    let endGrain = select(0.0, 1.0, row > 1.5 && row < 2.5);
    let origin = vec3<f32>(x - 70.0 * endGrain, select(select(225.0, 5.0, row < 2.5), -225.0, row < 1.5), select(select(select(-80.0, 0.0, row < 2.5), 10.0, row < 1.5), -80.0, row < 0.5));
    let axis = select(vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), endGrain > 0.0);
    let look = vec3<f32>(0.0, 0.0, select(0.0, 1.0, row > 2.5));
    let gsl_tmp_22 = p - vec3<f32>(x, -225.0 + 150.0 * row, 10.0);
    let gsl_tmp_24 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_25 = floor((p.y + 300.0) / 150.0);
    let gsl_tmp_26 = sign(gsl_tmp_22) * (vec3<f32>(1.0, 0.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_24 > 2.0), 0.0, gsl_tmp_24 < 0.0) * 330.0, 150.0 * select(select(0.0, 0.0, gsl_tmp_25 > 3.0), 0.0, gsl_tmp_25 < 0.0), 0.0));
    let gsl_tmp_23 = abs(gsl_tmp_22) - vec3<f32>(150.0, 50.0, 10.0);
    let gsl_tmp_27 = select(0.0, gsl_tmp_26.x, gsl_tmp_23.x + 1.0 > 0.0);
    let gsl_tmp_28 = select(0.0, gsl_tmp_26.y, gsl_tmp_23.y + 1.0 > 0.0);
    let gsl_tmp_29 = select(0.0, gsl_tmp_26.z, gsl_tmp_23.z + 1.0 > 0.0);
    let gsl_tmp_30 = sqrt(0.0 + max(gsl_tmp_23.x + 1.0, 0.0) * max(gsl_tmp_23.x + 1.0, 0.0) + max(gsl_tmp_23.y + 1.0, 0.0) * max(gsl_tmp_23.y + 1.0, 0.0) + max(gsl_tmp_23.z + 1.0, 0.0) * max(gsl_tmp_23.z + 1.0, 0.0));
    let gsl_tmp_31 = p - vec3<f32>(x, -225.0 + 150.0 * row, 10.0);
    let gsl_tmp_33 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_34 = floor((p.y + 300.0) / 150.0);
    let gsl_tmp_35 = sign(gsl_tmp_31) * (vec3<f32>(0.0, 1.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_33 > 2.0), 0.0, gsl_tmp_33 < 0.0) * 330.0, 150.0 * select(select(0.0, 0.0, gsl_tmp_34 > 3.0), 0.0, gsl_tmp_34 < 0.0), 0.0));
    let gsl_tmp_32 = abs(gsl_tmp_31) - vec3<f32>(150.0, 50.0, 10.0);
    let gsl_tmp_36 = select(0.0, gsl_tmp_35.x, gsl_tmp_32.x + 1.0 > 0.0);
    let gsl_tmp_37 = select(0.0, gsl_tmp_35.y, gsl_tmp_32.y + 1.0 > 0.0);
    let gsl_tmp_38 = select(0.0, gsl_tmp_35.z, gsl_tmp_32.z + 1.0 > 0.0);
    let gsl_tmp_39 = sqrt(0.0 + max(gsl_tmp_32.x + 1.0, 0.0) * max(gsl_tmp_32.x + 1.0, 0.0) + max(gsl_tmp_32.y + 1.0, 0.0) * max(gsl_tmp_32.y + 1.0, 0.0) + max(gsl_tmp_32.z + 1.0, 0.0) * max(gsl_tmp_32.z + 1.0, 0.0));
    let gsl_tmp_40 = p - vec3<f32>(x, -225.0 + 150.0 * row, 10.0);
    let gsl_tmp_42 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_43 = floor((p.y + 300.0) / 150.0);
    let gsl_tmp_44 = sign(gsl_tmp_40) * (vec3<f32>(0.0, 0.0, 1.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_42 > 2.0), 0.0, gsl_tmp_42 < 0.0) * 330.0, 150.0 * select(select(0.0, 0.0, gsl_tmp_43 > 3.0), 0.0, gsl_tmp_43 < 0.0), 0.0));
    let gsl_tmp_41 = abs(gsl_tmp_40) - vec3<f32>(150.0, 50.0, 10.0);
    let gsl_tmp_45 = select(0.0, gsl_tmp_44.x, gsl_tmp_41.x + 1.0 > 0.0);
    let gsl_tmp_46 = select(0.0, gsl_tmp_44.y, gsl_tmp_41.y + 1.0 > 0.0);
    let gsl_tmp_47 = select(0.0, gsl_tmp_44.z, gsl_tmp_41.z + 1.0 > 0.0);
    let gsl_tmp_48 = sqrt(0.0 + max(gsl_tmp_41.x + 1.0, 0.0) * max(gsl_tmp_41.x + 1.0, 0.0) + max(gsl_tmp_41.y + 1.0, 0.0) * max(gsl_tmp_41.y + 1.0, 0.0) + max(gsl_tmp_41.z + 1.0, 0.0) * max(gsl_tmp_41.z + 1.0, 0.0));
    let gsl_tmp_2 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_3 = sign(p - vec3<f32>(x, 75.0, 35.0)) * (vec3<f32>(1.0, 0.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_2 > 2.0), 0.0, gsl_tmp_2 < 0.0) * 330.0, 0.0, 0.0));
    let gsl_tmp_1 = abs(p - vec3<f32>(x, 75.0, 35.0)) - vec3<f32>(50.0, 50.0, 35.0);
    let gsl_tmp_4 = select(0.0, gsl_tmp_3.x, gsl_tmp_1.x + 1.5 > 0.0);
    let gsl_tmp_5 = select(0.0, gsl_tmp_3.y, gsl_tmp_1.y + 1.5 > 0.0);
    let gsl_tmp_6 = select(0.0, gsl_tmp_3.z, gsl_tmp_1.z + 1.5 > 0.0);
    let gsl_tmp_7 = sqrt(0.0 + max(gsl_tmp_1.x + 1.5, 0.0) * max(gsl_tmp_1.x + 1.5, 0.0) + max(gsl_tmp_1.y + 1.5, 0.0) * max(gsl_tmp_1.y + 1.5, 0.0) + max(gsl_tmp_1.z + 1.5, 0.0) * max(gsl_tmp_1.z + 1.5, 0.0));
    let gsl_tmp_9 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_10 = sign(p - vec3<f32>(x, 75.0, 35.0)) * (vec3<f32>(0.0, 1.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_9 > 2.0), 0.0, gsl_tmp_9 < 0.0) * 330.0, 0.0, 0.0));
    let gsl_tmp_8 = abs(p - vec3<f32>(x, 75.0, 35.0)) - vec3<f32>(50.0, 50.0, 35.0);
    let gsl_tmp_11 = select(0.0, gsl_tmp_10.x, gsl_tmp_8.x + 1.5 > 0.0);
    let gsl_tmp_12 = select(0.0, gsl_tmp_10.y, gsl_tmp_8.y + 1.5 > 0.0);
    let gsl_tmp_13 = select(0.0, gsl_tmp_10.z, gsl_tmp_8.z + 1.5 > 0.0);
    let gsl_tmp_14 = sqrt(0.0 + max(gsl_tmp_8.x + 1.5, 0.0) * max(gsl_tmp_8.x + 1.5, 0.0) + max(gsl_tmp_8.y + 1.5, 0.0) * max(gsl_tmp_8.y + 1.5, 0.0) + max(gsl_tmp_8.z + 1.5, 0.0) * max(gsl_tmp_8.z + 1.5, 0.0));
    let gsl_tmp_16 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_17 = sign(p - vec3<f32>(x, 75.0, 35.0)) * (vec3<f32>(0.0, 0.0, 1.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_16 > 2.0), 0.0, gsl_tmp_16 < 0.0) * 330.0, 0.0, 0.0));
    let gsl_tmp_15 = abs(p - vec3<f32>(x, 75.0, 35.0)) - vec3<f32>(50.0, 50.0, 35.0);
    let gsl_tmp_18 = select(0.0, gsl_tmp_17.x, gsl_tmp_15.x + 1.5 > 0.0);
    let gsl_tmp_19 = select(0.0, gsl_tmp_17.y, gsl_tmp_15.y + 1.5 > 0.0);
    let gsl_tmp_20 = select(0.0, gsl_tmp_17.z, gsl_tmp_15.z + 1.5 > 0.0);
    let gsl_tmp_21 = sqrt(0.0 + max(gsl_tmp_15.x + 1.5, 0.0) * max(gsl_tmp_15.x + 1.5, 0.0) + max(gsl_tmp_15.y + 1.5, 0.0) * max(gsl_tmp_15.y + 1.5, 0.0) + max(gsl_tmp_15.z + 1.5, 0.0) * max(gsl_tmp_15.z + 1.5, 0.0));
    let normal = select(vec3<f32>(select(select(gsl_tmp_26.z, select(gsl_tmp_26.y, gsl_tmp_26.x, gsl_tmp_23.x > gsl_tmp_23.y), max(gsl_tmp_23.x, gsl_tmp_23.y) > gsl_tmp_23.z), -0.0, -1.0 < max(max(gsl_tmp_23.x, gsl_tmp_23.y), gsl_tmp_23.z)) + (gsl_tmp_27 * max(gsl_tmp_23.x + 1.0, 0.0) + max(gsl_tmp_23.x + 1.0, 0.0) * gsl_tmp_27 + gsl_tmp_28 * max(gsl_tmp_23.y + 1.0, 0.0) + max(gsl_tmp_23.y + 1.0, 0.0) * gsl_tmp_28 + gsl_tmp_29 * max(gsl_tmp_23.z + 1.0, 0.0) + max(gsl_tmp_23.z + 1.0, 0.0) * gsl_tmp_29) / (2.0 * (gsl_tmp_30 + 1.0 - select(0.0, 1.0, (gsl_tmp_30 > 0.0)))), select(select(gsl_tmp_35.z, select(gsl_tmp_35.y, gsl_tmp_35.x, gsl_tmp_32.x > gsl_tmp_32.y), max(gsl_tmp_32.x, gsl_tmp_32.y) > gsl_tmp_32.z), -0.0, -1.0 < max(max(gsl_tmp_32.x, gsl_tmp_32.y), gsl_tmp_32.z)) + (gsl_tmp_36 * max(gsl_tmp_32.x + 1.0, 0.0) + max(gsl_tmp_32.x + 1.0, 0.0) * gsl_tmp_36 + gsl_tmp_37 * max(gsl_tmp_32.y + 1.0, 0.0) + max(gsl_tmp_32.y + 1.0, 0.0) * gsl_tmp_37 + gsl_tmp_38 * max(gsl_tmp_32.z + 1.0, 0.0) + max(gsl_tmp_32.z + 1.0, 0.0) * gsl_tmp_38) / (2.0 * (gsl_tmp_39 + 1.0 - select(0.0, 1.0, (gsl_tmp_39 > 0.0)))), select(select(gsl_tmp_44.z, select(gsl_tmp_44.y, gsl_tmp_44.x, gsl_tmp_41.x > gsl_tmp_41.y), max(gsl_tmp_41.x, gsl_tmp_41.y) > gsl_tmp_41.z), -0.0, -1.0 < max(max(gsl_tmp_41.x, gsl_tmp_41.y), gsl_tmp_41.z)) + (gsl_tmp_45 * max(gsl_tmp_41.x + 1.0, 0.0) + max(gsl_tmp_41.x + 1.0, 0.0) * gsl_tmp_45 + gsl_tmp_46 * max(gsl_tmp_41.y + 1.0, 0.0) + max(gsl_tmp_41.y + 1.0, 0.0) * gsl_tmp_46 + gsl_tmp_47 * max(gsl_tmp_41.z + 1.0, 0.0) + max(gsl_tmp_41.z + 1.0, 0.0) * gsl_tmp_47) / (2.0 * (gsl_tmp_48 + 1.0 - select(0.0, 1.0, (gsl_tmp_48 > 0.0))))), vec3<f32>(select(select(gsl_tmp_3.z, select(gsl_tmp_3.y, gsl_tmp_3.x, gsl_tmp_1.x > gsl_tmp_1.y), max(gsl_tmp_1.x, gsl_tmp_1.y) > gsl_tmp_1.z), -0.0, -1.5 < max(max(gsl_tmp_1.x, gsl_tmp_1.y), gsl_tmp_1.z)) + (gsl_tmp_4 * max(gsl_tmp_1.x + 1.5, 0.0) + max(gsl_tmp_1.x + 1.5, 0.0) * gsl_tmp_4 + gsl_tmp_5 * max(gsl_tmp_1.y + 1.5, 0.0) + max(gsl_tmp_1.y + 1.5, 0.0) * gsl_tmp_5 + gsl_tmp_6 * max(gsl_tmp_1.z + 1.5, 0.0) + max(gsl_tmp_1.z + 1.5, 0.0) * gsl_tmp_6) / (2.0 * (gsl_tmp_7 + 1.0 - select(0.0, 1.0, (gsl_tmp_7 > 0.0)))), select(select(gsl_tmp_10.z, select(gsl_tmp_10.y, gsl_tmp_10.x, gsl_tmp_8.x > gsl_tmp_8.y), max(gsl_tmp_8.x, gsl_tmp_8.y) > gsl_tmp_8.z), -0.0, -1.5 < max(max(gsl_tmp_8.x, gsl_tmp_8.y), gsl_tmp_8.z)) + (gsl_tmp_11 * max(gsl_tmp_8.x + 1.5, 0.0) + max(gsl_tmp_8.x + 1.5, 0.0) * gsl_tmp_11 + gsl_tmp_12 * max(gsl_tmp_8.y + 1.5, 0.0) + max(gsl_tmp_8.y + 1.5, 0.0) * gsl_tmp_12 + gsl_tmp_13 * max(gsl_tmp_8.z + 1.5, 0.0) + max(gsl_tmp_8.z + 1.5, 0.0) * gsl_tmp_13) / (2.0 * (gsl_tmp_14 + 1.0 - select(0.0, 1.0, (gsl_tmp_14 > 0.0)))), select(select(gsl_tmp_17.z, select(gsl_tmp_17.y, gsl_tmp_17.x, gsl_tmp_15.x > gsl_tmp_15.y), max(gsl_tmp_15.x, gsl_tmp_15.y) > gsl_tmp_15.z), -0.0, -1.5 < max(max(gsl_tmp_15.x, gsl_tmp_15.y), gsl_tmp_15.z)) + (gsl_tmp_18 * max(gsl_tmp_15.x + 1.5, 0.0) + max(gsl_tmp_15.x + 1.5, 0.0) * gsl_tmp_18 + gsl_tmp_19 * max(gsl_tmp_15.y + 1.5, 0.0) + max(gsl_tmp_15.y + 1.5, 0.0) * gsl_tmp_19 + gsl_tmp_20 * max(gsl_tmp_15.z + 1.5, 0.0) + max(gsl_tmp_15.z + 1.5, 0.0) * gsl_tmp_20) / (2.0 * (gsl_tmp_21 + 1.0 - select(0.0, 1.0, (gsl_tmp_21 > 0.0))))), endGrain > 0.0);
    let material_wood_noStain = mat2x4<f32>(vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let topcoat = mat2x4<f32>(select(vec4<f32>(1.0, 0.99812025, 0.9884352, 4.0), vec4<f32>(1.0, 1.0, 1.0, 3.0), row > 2.5), select(vec4<f32>(1.0, 0.005, 0.91, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.0), row > 2.5));
    return vec4<f32>(wood_speciesColor(p, number, origin, axis, look, -1.0, -1.0, normal, mat3x4<f32>(vec4<f32>(1.0, 180.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), material_wood_noStain, topcoat, noiseTex), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_wood_pineNumber = 14.0;
    let material_wood_walnutNumber = 4.0;
    let column = clamp(floor(p.x / 330.0 + 1.5), 0.0, 2.0);
    let material_wood_whiteOakNumber = 0.0;
    let number = select(select(material_wood_pineNumber, material_wood_walnutNumber, column < 1.5), material_wood_whiteOakNumber, column < 0.5);
    let x = (column - 1.0) * 330.0;
    let row = clamp(floor((p.y + 300.0) / 150.0), 0.0, 3.0);
    let endGrain = select(0.0, 1.0, row > 1.5 && row < 2.5);
    let origin = vec3<f32>(x - 70.0 * endGrain, select(select(225.0, 5.0, row < 2.5), -225.0, row < 1.5), select(select(select(-80.0, 0.0, row < 2.5), 10.0, row < 1.5), -80.0, row < 0.5));
    let axis = select(vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), endGrain > 0.0);
    let look = vec3<f32>(0.0, 0.0, select(0.0, 1.0, row > 2.5));
    let gsl_tmp_22 = p - vec3<f32>(x, -225.0 + 150.0 * row, 10.0);
    let gsl_tmp_24 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_25 = floor((p.y + 300.0) / 150.0);
    let gsl_tmp_26 = sign(gsl_tmp_22) * (vec3<f32>(1.0, 0.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_24 > 2.0), 0.0, gsl_tmp_24 < 0.0) * 330.0, 150.0 * select(select(0.0, 0.0, gsl_tmp_25 > 3.0), 0.0, gsl_tmp_25 < 0.0), 0.0));
    let gsl_tmp_23 = abs(gsl_tmp_22) - vec3<f32>(150.0, 50.0, 10.0);
    let gsl_tmp_27 = select(0.0, gsl_tmp_26.x, gsl_tmp_23.x + 1.0 > 0.0);
    let gsl_tmp_28 = select(0.0, gsl_tmp_26.y, gsl_tmp_23.y + 1.0 > 0.0);
    let gsl_tmp_29 = select(0.0, gsl_tmp_26.z, gsl_tmp_23.z + 1.0 > 0.0);
    let gsl_tmp_30 = sqrt(0.0 + max(gsl_tmp_23.x + 1.0, 0.0) * max(gsl_tmp_23.x + 1.0, 0.0) + max(gsl_tmp_23.y + 1.0, 0.0) * max(gsl_tmp_23.y + 1.0, 0.0) + max(gsl_tmp_23.z + 1.0, 0.0) * max(gsl_tmp_23.z + 1.0, 0.0));
    let gsl_tmp_31 = p - vec3<f32>(x, -225.0 + 150.0 * row, 10.0);
    let gsl_tmp_33 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_34 = floor((p.y + 300.0) / 150.0);
    let gsl_tmp_35 = sign(gsl_tmp_31) * (vec3<f32>(0.0, 1.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_33 > 2.0), 0.0, gsl_tmp_33 < 0.0) * 330.0, 150.0 * select(select(0.0, 0.0, gsl_tmp_34 > 3.0), 0.0, gsl_tmp_34 < 0.0), 0.0));
    let gsl_tmp_32 = abs(gsl_tmp_31) - vec3<f32>(150.0, 50.0, 10.0);
    let gsl_tmp_36 = select(0.0, gsl_tmp_35.x, gsl_tmp_32.x + 1.0 > 0.0);
    let gsl_tmp_37 = select(0.0, gsl_tmp_35.y, gsl_tmp_32.y + 1.0 > 0.0);
    let gsl_tmp_38 = select(0.0, gsl_tmp_35.z, gsl_tmp_32.z + 1.0 > 0.0);
    let gsl_tmp_39 = sqrt(0.0 + max(gsl_tmp_32.x + 1.0, 0.0) * max(gsl_tmp_32.x + 1.0, 0.0) + max(gsl_tmp_32.y + 1.0, 0.0) * max(gsl_tmp_32.y + 1.0, 0.0) + max(gsl_tmp_32.z + 1.0, 0.0) * max(gsl_tmp_32.z + 1.0, 0.0));
    let gsl_tmp_40 = p - vec3<f32>(x, -225.0 + 150.0 * row, 10.0);
    let gsl_tmp_42 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_43 = floor((p.y + 300.0) / 150.0);
    let gsl_tmp_44 = sign(gsl_tmp_40) * (vec3<f32>(0.0, 0.0, 1.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_42 > 2.0), 0.0, gsl_tmp_42 < 0.0) * 330.0, 150.0 * select(select(0.0, 0.0, gsl_tmp_43 > 3.0), 0.0, gsl_tmp_43 < 0.0), 0.0));
    let gsl_tmp_41 = abs(gsl_tmp_40) - vec3<f32>(150.0, 50.0, 10.0);
    let gsl_tmp_45 = select(0.0, gsl_tmp_44.x, gsl_tmp_41.x + 1.0 > 0.0);
    let gsl_tmp_46 = select(0.0, gsl_tmp_44.y, gsl_tmp_41.y + 1.0 > 0.0);
    let gsl_tmp_47 = select(0.0, gsl_tmp_44.z, gsl_tmp_41.z + 1.0 > 0.0);
    let gsl_tmp_48 = sqrt(0.0 + max(gsl_tmp_41.x + 1.0, 0.0) * max(gsl_tmp_41.x + 1.0, 0.0) + max(gsl_tmp_41.y + 1.0, 0.0) * max(gsl_tmp_41.y + 1.0, 0.0) + max(gsl_tmp_41.z + 1.0, 0.0) * max(gsl_tmp_41.z + 1.0, 0.0));
    let gsl_tmp_2 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_3 = sign(p - vec3<f32>(x, 75.0, 35.0)) * (vec3<f32>(1.0, 0.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_2 > 2.0), 0.0, gsl_tmp_2 < 0.0) * 330.0, 0.0, 0.0));
    let gsl_tmp_1 = abs(p - vec3<f32>(x, 75.0, 35.0)) - vec3<f32>(50.0, 50.0, 35.0);
    let gsl_tmp_4 = select(0.0, gsl_tmp_3.x, gsl_tmp_1.x + 1.5 > 0.0);
    let gsl_tmp_5 = select(0.0, gsl_tmp_3.y, gsl_tmp_1.y + 1.5 > 0.0);
    let gsl_tmp_6 = select(0.0, gsl_tmp_3.z, gsl_tmp_1.z + 1.5 > 0.0);
    let gsl_tmp_7 = sqrt(0.0 + max(gsl_tmp_1.x + 1.5, 0.0) * max(gsl_tmp_1.x + 1.5, 0.0) + max(gsl_tmp_1.y + 1.5, 0.0) * max(gsl_tmp_1.y + 1.5, 0.0) + max(gsl_tmp_1.z + 1.5, 0.0) * max(gsl_tmp_1.z + 1.5, 0.0));
    let gsl_tmp_9 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_10 = sign(p - vec3<f32>(x, 75.0, 35.0)) * (vec3<f32>(0.0, 1.0, 0.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_9 > 2.0), 0.0, gsl_tmp_9 < 0.0) * 330.0, 0.0, 0.0));
    let gsl_tmp_8 = abs(p - vec3<f32>(x, 75.0, 35.0)) - vec3<f32>(50.0, 50.0, 35.0);
    let gsl_tmp_11 = select(0.0, gsl_tmp_10.x, gsl_tmp_8.x + 1.5 > 0.0);
    let gsl_tmp_12 = select(0.0, gsl_tmp_10.y, gsl_tmp_8.y + 1.5 > 0.0);
    let gsl_tmp_13 = select(0.0, gsl_tmp_10.z, gsl_tmp_8.z + 1.5 > 0.0);
    let gsl_tmp_14 = sqrt(0.0 + max(gsl_tmp_8.x + 1.5, 0.0) * max(gsl_tmp_8.x + 1.5, 0.0) + max(gsl_tmp_8.y + 1.5, 0.0) * max(gsl_tmp_8.y + 1.5, 0.0) + max(gsl_tmp_8.z + 1.5, 0.0) * max(gsl_tmp_8.z + 1.5, 0.0));
    let gsl_tmp_16 = floor(p.x / 330.0 + 1.5);
    let gsl_tmp_17 = sign(p - vec3<f32>(x, 75.0, 35.0)) * (vec3<f32>(0.0, 0.0, 1.0) - vec3<f32>(select(select(0.0, 0.0, gsl_tmp_16 > 2.0), 0.0, gsl_tmp_16 < 0.0) * 330.0, 0.0, 0.0));
    let gsl_tmp_15 = abs(p - vec3<f32>(x, 75.0, 35.0)) - vec3<f32>(50.0, 50.0, 35.0);
    let gsl_tmp_18 = select(0.0, gsl_tmp_17.x, gsl_tmp_15.x + 1.5 > 0.0);
    let gsl_tmp_19 = select(0.0, gsl_tmp_17.y, gsl_tmp_15.y + 1.5 > 0.0);
    let gsl_tmp_20 = select(0.0, gsl_tmp_17.z, gsl_tmp_15.z + 1.5 > 0.0);
    let gsl_tmp_21 = sqrt(0.0 + max(gsl_tmp_15.x + 1.5, 0.0) * max(gsl_tmp_15.x + 1.5, 0.0) + max(gsl_tmp_15.y + 1.5, 0.0) * max(gsl_tmp_15.y + 1.5, 0.0) + max(gsl_tmp_15.z + 1.5, 0.0) * max(gsl_tmp_15.z + 1.5, 0.0));
    let normal = select(vec3<f32>(select(select(gsl_tmp_26.z, select(gsl_tmp_26.y, gsl_tmp_26.x, gsl_tmp_23.x > gsl_tmp_23.y), max(gsl_tmp_23.x, gsl_tmp_23.y) > gsl_tmp_23.z), -0.0, -1.0 < max(max(gsl_tmp_23.x, gsl_tmp_23.y), gsl_tmp_23.z)) + (gsl_tmp_27 * max(gsl_tmp_23.x + 1.0, 0.0) + max(gsl_tmp_23.x + 1.0, 0.0) * gsl_tmp_27 + gsl_tmp_28 * max(gsl_tmp_23.y + 1.0, 0.0) + max(gsl_tmp_23.y + 1.0, 0.0) * gsl_tmp_28 + gsl_tmp_29 * max(gsl_tmp_23.z + 1.0, 0.0) + max(gsl_tmp_23.z + 1.0, 0.0) * gsl_tmp_29) / (2.0 * (gsl_tmp_30 + 1.0 - select(0.0, 1.0, (gsl_tmp_30 > 0.0)))), select(select(gsl_tmp_35.z, select(gsl_tmp_35.y, gsl_tmp_35.x, gsl_tmp_32.x > gsl_tmp_32.y), max(gsl_tmp_32.x, gsl_tmp_32.y) > gsl_tmp_32.z), -0.0, -1.0 < max(max(gsl_tmp_32.x, gsl_tmp_32.y), gsl_tmp_32.z)) + (gsl_tmp_36 * max(gsl_tmp_32.x + 1.0, 0.0) + max(gsl_tmp_32.x + 1.0, 0.0) * gsl_tmp_36 + gsl_tmp_37 * max(gsl_tmp_32.y + 1.0, 0.0) + max(gsl_tmp_32.y + 1.0, 0.0) * gsl_tmp_37 + gsl_tmp_38 * max(gsl_tmp_32.z + 1.0, 0.0) + max(gsl_tmp_32.z + 1.0, 0.0) * gsl_tmp_38) / (2.0 * (gsl_tmp_39 + 1.0 - select(0.0, 1.0, (gsl_tmp_39 > 0.0)))), select(select(gsl_tmp_44.z, select(gsl_tmp_44.y, gsl_tmp_44.x, gsl_tmp_41.x > gsl_tmp_41.y), max(gsl_tmp_41.x, gsl_tmp_41.y) > gsl_tmp_41.z), -0.0, -1.0 < max(max(gsl_tmp_41.x, gsl_tmp_41.y), gsl_tmp_41.z)) + (gsl_tmp_45 * max(gsl_tmp_41.x + 1.0, 0.0) + max(gsl_tmp_41.x + 1.0, 0.0) * gsl_tmp_45 + gsl_tmp_46 * max(gsl_tmp_41.y + 1.0, 0.0) + max(gsl_tmp_41.y + 1.0, 0.0) * gsl_tmp_46 + gsl_tmp_47 * max(gsl_tmp_41.z + 1.0, 0.0) + max(gsl_tmp_41.z + 1.0, 0.0) * gsl_tmp_47) / (2.0 * (gsl_tmp_48 + 1.0 - select(0.0, 1.0, (gsl_tmp_48 > 0.0))))), vec3<f32>(select(select(gsl_tmp_3.z, select(gsl_tmp_3.y, gsl_tmp_3.x, gsl_tmp_1.x > gsl_tmp_1.y), max(gsl_tmp_1.x, gsl_tmp_1.y) > gsl_tmp_1.z), -0.0, -1.5 < max(max(gsl_tmp_1.x, gsl_tmp_1.y), gsl_tmp_1.z)) + (gsl_tmp_4 * max(gsl_tmp_1.x + 1.5, 0.0) + max(gsl_tmp_1.x + 1.5, 0.0) * gsl_tmp_4 + gsl_tmp_5 * max(gsl_tmp_1.y + 1.5, 0.0) + max(gsl_tmp_1.y + 1.5, 0.0) * gsl_tmp_5 + gsl_tmp_6 * max(gsl_tmp_1.z + 1.5, 0.0) + max(gsl_tmp_1.z + 1.5, 0.0) * gsl_tmp_6) / (2.0 * (gsl_tmp_7 + 1.0 - select(0.0, 1.0, (gsl_tmp_7 > 0.0)))), select(select(gsl_tmp_10.z, select(gsl_tmp_10.y, gsl_tmp_10.x, gsl_tmp_8.x > gsl_tmp_8.y), max(gsl_tmp_8.x, gsl_tmp_8.y) > gsl_tmp_8.z), -0.0, -1.5 < max(max(gsl_tmp_8.x, gsl_tmp_8.y), gsl_tmp_8.z)) + (gsl_tmp_11 * max(gsl_tmp_8.x + 1.5, 0.0) + max(gsl_tmp_8.x + 1.5, 0.0) * gsl_tmp_11 + gsl_tmp_12 * max(gsl_tmp_8.y + 1.5, 0.0) + max(gsl_tmp_8.y + 1.5, 0.0) * gsl_tmp_12 + gsl_tmp_13 * max(gsl_tmp_8.z + 1.5, 0.0) + max(gsl_tmp_8.z + 1.5, 0.0) * gsl_tmp_13) / (2.0 * (gsl_tmp_14 + 1.0 - select(0.0, 1.0, (gsl_tmp_14 > 0.0)))), select(select(gsl_tmp_17.z, select(gsl_tmp_17.y, gsl_tmp_17.x, gsl_tmp_15.x > gsl_tmp_15.y), max(gsl_tmp_15.x, gsl_tmp_15.y) > gsl_tmp_15.z), -0.0, -1.5 < max(max(gsl_tmp_15.x, gsl_tmp_15.y), gsl_tmp_15.z)) + (gsl_tmp_18 * max(gsl_tmp_15.x + 1.5, 0.0) + max(gsl_tmp_15.x + 1.5, 0.0) * gsl_tmp_18 + gsl_tmp_19 * max(gsl_tmp_15.y + 1.5, 0.0) + max(gsl_tmp_15.y + 1.5, 0.0) * gsl_tmp_19 + gsl_tmp_20 * max(gsl_tmp_15.z + 1.5, 0.0) + max(gsl_tmp_15.z + 1.5, 0.0) * gsl_tmp_20) / (2.0 * (gsl_tmp_21 + 1.0 - select(0.0, 1.0, (gsl_tmp_21 > 0.0))))), endGrain > 0.0);
    let topcoat = mat2x4<f32>(select(vec4<f32>(1.0, 0.99812025, 0.9884352, 4.0), vec4<f32>(1.0, 1.0, 1.0, 3.0), row > 2.5), select(vec4<f32>(1.0, 0.005, 0.91, 0.5), vec4<f32>(0.0, 0.0, 0.0, 0.0), row > 2.5));
    let gsl_m = wood_speciesMaterial(p, number, origin, axis, look, -1.0, normal, mat3x4<f32>(vec4<f32>(1.0, 180.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), topcoat, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
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
}
