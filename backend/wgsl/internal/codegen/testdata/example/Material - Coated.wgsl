fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_reliefRoughness(roughness: f32, slope: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sqrt(roughness * roughness + (1.0 - texture_fadeDetail(size, noiseTex) * gsl_dispDrawn) * slope * slope);
}

fn material_withRoughness(m: mat4x4<f32>, rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(rough, surface.y, surface.z, surface.w), m[1], m[2], m[3]);
}

fn material_withLay(m: mat4x4<f32>, direction: vec3<f32>, anisotropy: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], m[1], m[2], vec4<f32>(direction.x, direction.y, direction.z, anisotropy));
}

fn material_displacement(m: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let volume = m[2];
  return volume.z;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn material_addDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_withDisplacement(m, material_displacement(m, noiseTex) + height, noiseTex);
}

fn material_roughness(m: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let surface = m[0];
  return surface.x;
}

fn material_withMetallic(m: mat4x4<f32>, metal: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(surface.x, metal, surface.z, surface.w), m[1], m[2], m[3]);
}

fn builtin_intersectSharp(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_firstOf(list: f32, noiseTex: texture_3d<f32>) -> f32 {
  return list;
}

fn builtin_intersectChamferEdges(fields: vec2<f32>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  let face = builtin_intersectSharp(fields, noiseTex);
  return builtin_intersectSharp(vec2<f32>(builtin_intersectSharp(vec2<f32>(select(face, (fields.x + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.x + fields.y + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(vec2<f32>(select(face, (fields.y + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.y + setback) / 1.4142135, 0.0 > 0.0)), noiseTex)), noiseTex);
}

fn builtin_lastOf(list: f32, noiseTex: texture_3d<f32>) -> f32 {
  return list;
}

fn builtin_intersectChamferVertices(fields: vec2<f32>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  let face = builtin_intersectSharp(fields, noiseTex);
  return builtin_intersectSharp(vec2<f32>(builtin_intersectSharp(vec2<f32>(builtin_intersectSharp(vec2<f32>(select(face, (fields.x + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(vec2<f32>(select(face, (fields.x + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp(vec2<f32>(builtin_intersectSharp(vec2<f32>(select(face, (fields.y + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(vec2<f32>(select(face, (fields.y + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex)), noiseTex);
}

fn builtin_intersectChamfer(fields: vec2<f32>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(builtin_intersectSharp(fields, noiseTex), builtin_intersectChamferEdges(fields, builtin_firstOf(setback, noiseTex), noiseTex)), builtin_intersectChamferVertices(fields, builtin_lastOf(setback, noiseTex), noiseTex));
}

fn builtin_gcl_closure_intersectChamfer(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectChamfer(vec2<f32>(radialDistance, axialDistance), noiseTex);
  return form;
}

fn builtin_gcl_closure_intersectChamfer_2(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_cylinderCapped_2(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectChamfer_2(vec2<f32>(radialDistance, axialDistance), noiseTex);
  return form;
}

fn builtin_gcl_closure_intersectChamfer_3(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_cylinderCapped_3(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectChamfer_3(vec2<f32>(radialDistance, axialDistance), noiseTex);
  return form;
}

fn builtin_gcl_closure_intersectChamfer_4(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_cylinderCapped_4(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectChamfer_4(vec2<f32>(radialDistance, axialDistance), noiseTex);
  return form;
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

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let la = color_rgb2oklab(a, noiseTex);
  let lb = color_rgb2oklab(b, noiseTex);
  return color_oklab2rgb(la + (lb - la) * t, noiseTex);
}

fn builtin_intersectSharp_2(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectChamferEdges_2(fields: vec3<f32>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  let face = builtin_intersectSharp_2(fields, noiseTex);
  return builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.x + fields.y + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields.x + fields.z + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.z + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.z + setback) / 1.4142135, 0.0 > 0.0)), noiseTex)), noiseTex);
}

fn builtin_intersectChamferVertices_2(fields: vec3<f32>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  let face = builtin_intersectSharp_2(fields, noiseTex);
  return builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.z + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.x + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.y + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp_2(vec3<f32>(builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp_2(vec3<f32>(select(face, (fields.z + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex)), noiseTex);
}

fn builtin_intersectChamfer_2(fields: vec3<f32>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(builtin_intersectSharp_2(fields, noiseTex), builtin_intersectChamferEdges_2(fields, builtin_firstOf(setback, noiseTex), noiseTex)), builtin_intersectChamferVertices_2(fields, builtin_lastOf(setback, noiseTex), noiseTex));
}

fn builtin_gcl_closure_intersectChamfer_5(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer_2(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamfer_5(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered(p, center, size, noiseTex);
}

fn material_withIor(m: mat4x4<f32>, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(surface.x, surface.y, index, surface.w), m[1], m[2], m[3]);
}

fn builtin_gcl_closure_intersectChamfer_6(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer_2(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_2(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamfer_6(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_box_2(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered_2(p, center, size, noiseTex);
}

fn builtin_gcl_closure_intersectChamfer_7(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer_2(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_3(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamfer_7(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_box_3(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered_3(p, center, size, noiseTex);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_4(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_intersectRound_2(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_5(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_2(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn noise_cellularF1Lattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_cellularScan(q, 1, noiseTex).f1;
}

fn builtin_gcl_closure_intersectRound_3(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_6(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_3(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
}

fn builtin_gcl_closure_intersectRound_4(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_7(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_4(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn noise_cellularF1F2CellLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  gsl_result.v_3 = gsl_cellularScan(q, 1, noiseTex).nearCell;
  return gsl_result;
}

fn noise_cellularCell(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  let gsl_result_7 = noise_cellularF1F2CellLattice(p / size, noiseTex);
  let f1 = gsl_result_7.v_1;
  let f2 = gsl_result_7.v_2;
  let cell = gsl_result_7.v_3;
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = f1 * size;
  gsl_result.v_2 = f2 * size;
  gsl_result.v_3 = cell;
  return gsl_result;
}

fn noise_gradient(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / size, noiseTex);
}

fn material_roughnessFromGloss(gloss: f32, noiseTex: texture_3d<f32>) -> f32 {
  let g = clamp(gloss, 0.5, 90.0);
  return clamp(0.647 - 0.1243 * log(g) + 0.008 * log(1.0 - g / 91.5), 0.045, 1.0);
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn builtin_gcl_closure_intersectChamfer_8(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectChamfer_2(fields, 1.0, noiseTex);
}

fn builtin_gcl_closure_boxCentered_8(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamfer_8(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_box_4(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered_8(p, center, size, noiseTex);
}

fn builtin_gcl_closure_intersectRound_5(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_9(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_5(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_intersectRound_6(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_10(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_6(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_intersectRound_7(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_11(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_7(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn material_filmKeeps(wavelength: f32, build: f32, noiseTex: texture_3d<f32>) -> f32 {
  let x = wavelength / (15.0 * max(build, 0.0001));
  return 1.0 - exp(-x * x);
}

fn material_lay(m: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let grooves = m[3];
  return grooves.xyz;
}

fn material_anisotropy(m: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let grooves = m[3];
  return grooves.w;
}

fn builtin_gcl_closure_intersectRound_8(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_12(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_8(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_intersectRound_9(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_13(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_9(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_gcl_closure_intersectRound_10(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 1.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_14(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_10(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn material_make(rough: f32, metal: f32, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(vec4<f32>(rough, metal, index, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
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

fn texture_coverageLine(gsl_var_distance: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let half = max(0.5 * width, 0.5 * gsl_pw);
  return (1.0 - texture_mask(gsl_var_distance, half, 0.25 * half, noiseTex)) * 0.5 * width / half;
}

fn coating_anodized(baseMaterial: mat4x4<f32>, baseColor: vec3<f32>, dyeColor: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  let rough = material_roughness(baseMaterial, noiseTex);
  let coat = material_withRoughness(material_withMetallic(baseMaterial, 1.0, noiseTex), sqrt(rough * rough + 0.01), noiseTex);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = coat;
  gsl_result.v_2 = baseColor * dyeColor * 0.9;
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn coating_hardAnodized(baseMaterial: mat4x4<f32>, baseColor: vec3<f32>, thickness: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let depth = clamp((thickness - 25.0) / 50.0, 0.0, 1.0);
  let rough = material_roughness(baseMaterial, noiseTex);
  let coat = material_withRoughness(material_withMetallic(baseMaterial, 0.6 - 0.4 * depth, noiseTex), sqrt(rough * rough + 0.09), noiseTex);
  let film = color_mixOklab(vec3<f32>(0.46, 0.45, 0.42), vec3<f32>(0.1, 0.1, 0.09), depth, noiseTex);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = coat;
  gsl_result.v_2 = film;
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn coating_blackOxide(baseMaterial: mat4x4<f32>, oiled: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let rough = material_roughness(baseMaterial, noiseTex) * (1.0 - 0.3 * clamp(oiled, 0.0, 1.0));
  let coat = material_withIor(material_withRoughness(material_withMetallic(baseMaterial, 0.0, noiseTex), rough, noiseTex), 2.4, noiseTex);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = coat;
  gsl_result.v_2 = vec3<f32>(0.045, 0.045, 0.05);
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn coating_edgeWear(faceA: f32, faceB: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - clamp(-min(faceA, faceB) / width, 0.0, 1.0);
}

fn coating_chipReach(wear: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 4.2 * wear * wear;
}

fn coating_chipEdge(i: f32, reach: f32, extra: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 2.4 + 0.35 * i + (0.15 + 0.25 * i) * extra - reach;
}

fn coating_normalTail(y: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 / (1.0 + exp(1.5976 * y + 0.070566 * y * y * y));
}

fn coating_chipUnion(reach: f32, extra: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(coating_normalTail((2.06 - reach) / 0.9, noiseTex), coating_normalTail((2.4 - reach) / 0.98, noiseTex), extra > 0.5);
}

fn coating_chipKept(z: f32, i: f32, reach: f32, extra: f32, perShare: f32, shown: f32, noiseTex: texture_3d<f32>) -> f32 {
  let edge = coating_chipEdge(i, reach, extra, noiseTex);
  let gsl_var_target = 1.0 - exp(perShare * coating_normalTail(edge, noiseTex));
  return 1.0 - gsl_var_target - (texture_mask(z, edge, 0.04, noiseTex) - gsl_var_target) * shown;
}

fn coating_chipLayer(z0: f32, z1: f32, z2: f32, z3: f32, z4: f32, w0: f32, w1: f32, w2: f32, w3: f32, w4: f32, reach: f32, extra: f32, noiseTex: texture_3d<f32>) -> f32 {
  let total = 0.0 + coating_normalTail(coating_chipEdge(0.0, reach, extra, noiseTex), noiseTex) + coating_normalTail(coating_chipEdge(1.0, reach, extra, noiseTex), noiseTex) + coating_normalTail(coating_chipEdge(2.0, reach, extra, noiseTex), noiseTex) + coating_normalTail(coating_chipEdge(3.0, reach, extra, noiseTex), noiseTex) + coating_normalTail(coating_chipEdge(4.0, reach, extra, noiseTex), noiseTex);
  let perShare = log(max(1.0 - coating_chipUnion(reach, extra, noiseTex), 0.000001)) / max(total, 0.000001);
  return 1.0 - coating_chipKept(z0, 0.0, reach, extra, perShare, w0, noiseTex) * coating_chipKept(z1, 1.0, reach, extra, perShare, w1, noiseTex) * coating_chipKept(z2, 2.0, reach, extra, perShare, w2, noiseTex) * coating_chipKept(z3, 3.0, reach, extra, perShare, w3, noiseTex) * coating_chipKept(z4, 4.0, reach, extra, perShare, w4, noiseTex);
}

fn coating_chipCover(p: vec3<f32>, chipSize: f32, reach: f32, noiseTex: texture_3d<f32>) -> gsl_struct_3 {
  let s0 = 4.0 * chipSize;
  let w0 = texture_fadeDetail(s0, noiseTex);
  let n0 = w0 * noise_gradient(p, s0, noiseTex) / 0.189;
  let w1 = texture_fadeDetail(s0 / 2.07, noiseTex);
  let n1 = w1 * noise_gradient(p, s0 / 2.07, noiseTex) / 0.189;
  let w2 = texture_fadeDetail(s0 / 4.21, noiseTex);
  let n2 = w2 * noise_gradient(p, s0 / 4.21, noiseTex) / 0.189;
  let w3 = texture_fadeDetail(s0 / 8.6, noiseTex);
  let n3 = w3 * noise_gradient(p, s0 / 8.6, noiseTex) / 0.189;
  let w4 = texture_fadeDetail(s0 / 17.3, noiseTex);
  let n4 = w4 * noise_gradient(p, s0 / 17.3, noiseTex) / 0.189;
  let f4 = n4;
  let f3 = n3 + 0.55 * f4;
  let f2 = n2 + 0.55 * f3;
  let f1 = n1 + 0.55 * f2;
  let f0 = n0 + 0.55 * f1;
  let e4 = w4 * w4;
  let e3 = w3 * w3 + 0.3025 * e4;
  let e2 = w2 * w2 + 0.3025 * e3;
  let e1 = w1 * w1 + 0.3025 * e2;
  let e0 = w0 * w0 + 0.3025 * e1;
  let gsl_result_8 = noise_cellularCell(p, chipSize, noiseTex);
  let cell = gsl_result_8.v_3;
  let facet = 2.4248 * (fract(textureLoad(noiseTex, vec3<i32>(floor(cell + vec3<f32>(29.0, 7.0, 43.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell + vec3<f32>(29.0, 7.0, 43.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5);
  let z0 = (f0 / sqrt(max(e0, 0.000001)) + facet * w2) * 0.8192 - 0.5 * (abs(n1) - 0.83 * w1);
  let z1 = ((f1 / sqrt(max(e1, 0.000001)) + 0.35 * n0) * 0.9439 + facet * w2) * 0.8192 - 0.5 * (abs(n2) - 0.83 * w2);
  let z2 = ((f2 / sqrt(max(e2, 0.000001)) + 0.35 * n0) * 0.9439 + facet * w2) * 0.8192 - 0.5 * (abs(n3) - 0.83 * w3);
  let z3 = (f3 / sqrt(max(e3, 0.000001)) + 0.35 * n0) * 0.9439 - 0.5 * (abs(n4) - 0.83 * w4);
  let z4 = (f4 / sqrt(max(e4, 0.000001)) + 0.35 * n0) * 0.9439;
  var gsl_result : gsl_struct_3;
  gsl_result.v_1 = coating_chipLayer(z0, z1, z2, z3, z4, w0, w1, w2, w3, w4, reach, 0.0, noiseTex);
  gsl_result.v_2 = coating_chipLayer(z0, z1, z2, z3, z4, w0, w1, w2, w3, w4, reach, 1.0, noiseTex);
  return gsl_result;
}

fn coating_glossRoughness(glossUnits: f32, noiseTex: texture_3d<f32>) -> f32 {
  return material_roughnessFromGloss(glossUnits, noiseTex);
}

fn coating_chipUnion__spec1(reach: f32, noiseTex: texture_3d<f32>) -> f32 {
  return coating_normalTail((2.06 - reach) / 0.9, noiseTex);
}

fn coating_chipUnion__spec2(reach: f32, noiseTex: texture_3d<f32>) -> f32 {
  return coating_normalTail((2.4 - reach) / 0.98, noiseTex);
}

fn coating_chippedStandInMaterial(baseMaterial: mat4x4<f32>, glossUnits: f32, wear: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let reach = coating_chipReach(wear, noiseTex);
  let share1 = coating_chipUnion__spec1(reach, noiseTex);
  let share2 = coating_chipUnion__spec2(reach, noiseTex);
  let coats = material_blend(material_make(coating_glossRoughness(glossUnits, noiseTex), 0.0, 1.5, noiseTex), mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), (share1 - share2) / max(1.0 - share2, 0.0001), noiseTex);
  return material_withDisplacement(material_blend(coats, baseMaterial, share2, noiseTex), material_displacement(baseMaterial, noiseTex), noiseTex);
}

fn coating_chippedStandInColor(baseColor: vec3<f32>, paintColor: vec3<f32>, primerColor: vec3<f32>, wear: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let reach = coating_chipReach(wear, noiseTex);
  let share1 = coating_chipUnion__spec1(reach, noiseTex);
  let share2 = coating_chipUnion__spec2(reach, noiseTex);
  return color_oklab2rgb(color_rgb2oklab(paintColor, noiseTex) * (1.0 - share1) + color_rgb2oklab(primerColor, noiseTex) * (share1 - share2) + color_rgb2oklab(baseColor, noiseTex) * share2, noiseTex);
}

fn coating_chippedPaint(p: vec3<f32>, baseMaterial: mat4x4<f32>, baseColor: vec3<f32>, paintColor: vec3<f32>, primerColor: vec3<f32>, glossUnits: f32, wear: f32, chipSize: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let chalk = (noise_gradientLattice(p / 80.0, noiseTex) + 0.8 * noise_gradientLattice(p / 12.0, noiseTex) * (1.0 - texture_mask(gsl_pw, 4.5, 1.5, noiseTex))) / 0.24192;
  let reach = coating_chipReach(wear, noiseTex);
  let gsl_result_9 = coating_chipCover(p, chipSize, reach, noiseTex);
  let paintGone = gsl_result_9.v_1;
  let primerGone = gsl_result_9.v_2;
  let primerShows = clamp((paintGone - primerGone) / max(1.0 - primerGone, 0.0001), 0.0, 1.0);
  var gsl_simpleValue = coating_chippedStandInMaterial(baseMaterial, glossUnits, wear, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let paintMat = material_make(coating_glossRoughness(glossUnits, noiseTex) + 0.02 * chalk, 0.0, 1.5, noiseTex);
    let coated = material_blend(paintMat, mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), primerShows, noiseTex);
    let worn = material_blend(coated, baseMaterial, primerGone, noiseTex);
    let gsl_simpleFull = material_withDisplacement(worn, material_displacement(baseMaterial, noiseTex) + 0.05 * (coating_chipUnion__spec1(reach, noiseTex) - paintGone) + 0.03 * (coating_chipUnion__spec2(reach, noiseTex) - primerGone), noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let chipped = gsl_simpleValue;
  var gsl_simpleValue_2 = coating_chippedStandInColor(baseColor, paintColor, primerColor, wear, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let paintLab = color_rgb2oklab(paintColor, noiseTex);
    let dL = 0.009 * chalk;
    let mottled = color_oklab2rgb(paintLab + vec3<f32>(dL, -8.0 * dL * paintLab.y, -8.0 * dL * paintLab.z), noiseTex);
    let coatColor = color_mixOklab(mottled, primerColor, primerShows, noiseTex);
    gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_mixOklab(coatColor, baseColor, primerGone, noiseTex), gsl_simpleDetail);
  }
  let finish = gsl_simpleValue_2;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = chipped;
  gsl_result.v_2 = finish;
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn coating_levelled(coat: mat4x4<f32>, baseMaterial: mat4x4<f32>, build: f32, underSize: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let coatRough = material_roughness(coat, noiseTex);
  let coatAlpha = coatRough * coatRough;
  let keeps = material_filmKeeps(underSize, build, noiseTex);
  let baseRough = material_roughness(baseMaterial, noiseTex);
  let baseAlpha = baseRough * baseRough;
  let alpha = sqrt(coatAlpha * coatAlpha + keeps * keeps * baseAlpha * baseAlpha);
  let lobe = material_withRoughness(coat, sqrt(alpha), noiseTex);
  let relief = material_displacement(coat, noiseTex) + material_displacement(baseMaterial, noiseTex) * keeps;
  return material_withDisplacement(material_withLay(lobe, material_lay(baseMaterial, noiseTex), material_anisotropy(baseMaterial, noiseTex) * keeps, noiseTex), relief, noiseTex);
}

fn coating_clearCoat(baseMaterial: mat4x4<f32>, baseColor: vec3<f32>, tint: vec3<f32>, glossUnits: f32, build: f32, underSize: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let bare = material_withLay(material_withRoughness(baseMaterial, coating_glossRoughness(glossUnits, noiseTex), noiseTex), vec3<f32>(0.0, 0.0, 0.0), 0.0, noiseTex);
  let film = material_withDisplacement(bare, 0.0, noiseTex);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = coating_levelled(film, baseMaterial, build, underSize, noiseTex);
  gsl_result.v_2 = baseColor * tint * 0.97;
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn coating_powderCoat_2(p: vec3<f32>, coatColor: vec3<f32>, glossUnits: f32, peel: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let rough = coating_glossRoughness(glossUnits, noiseTex);
  let alpha = rough * rough;
  let amount = clamp(peel, 0.0, 1.0);
  var gsl_detailValue_2 = 0.0;
  if (max(gsl_pw, gsl_detailPw) < 0.15) { let gsl_detailOuterPw_2 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_2 = noise_gradientLattice(p / 0.3, noiseTex) * 0.0006 * amount * (1.0 - texture_mask(gsl_pw, 0.1125, 0.0375, noiseTex));
    gsl_pw = gsl_detailOuterPw_2;
  }
  var gsl_detailValue = 0.0;
  if (max(gsl_pw, gsl_detailPw) < 0.75) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let shortWave = gsl_detailValue_2;
    gsl_detailValue = (noise_gradientLattice(p / 1.5, noiseTex) + 0.5 * noise_gradientLattice(p / 0.65, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.24375, 0.08125, noiseTex))) * 0.0065 * amount * (1.0 - texture_mask(gsl_pw, 0.5625, 0.1875, noiseTex)) + shortWave;
    gsl_pw = gsl_detailOuterPw;
  }
  var gsl_simpleValue = material_make(sqrt(sqrt(alpha * alpha + 0.00001124 * amount * amount)), 0.0, 1.5, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let lobe = texture_reliefRoughness(alpha, 0.001 * amount, 0.3, noiseTex);
    let wide = texture_reliefRoughness(lobe, 0.0032 * amount, 1.5, noiseTex);
    let waves = gsl_detailValue;
    let gsl_simpleFull = material_withDisplacement(material_make(sqrt(wide), 0.0, 1.5, noiseTex), waves, noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let coat = gsl_simpleValue;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = coat;
  gsl_result.v_2 = coatColor;
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn coating_powderCoat(p: vec3<f32>, baseMaterial: mat4x4<f32>, coatColor: vec3<f32>, glossUnits: f32, peel: f32, build: f32, underSize: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let rough = coating_glossRoughness(glossUnits, noiseTex);
  let amount = clamp(peel, 0.0, 1.0);
  let flat = material_make(sqrt(sqrt(rough * rough * rough * rough + 0.00001124 * amount * amount)), 0.0, 1.5, noiseTex);
  let gsl_result_15 = coating_powderCoat_2(p, coatColor, glossUnits, peel, noiseTex);
  var gsl_simpleValue = coating_levelled(flat, baseMaterial, build, underSize, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let coat = gsl_result_15.v_1;
    let gsl_simpleFull = coating_levelled(coat, baseMaterial, build, underSize, noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let result = gsl_simpleValue;
  let tone = gsl_result_15.v_2;
  let height = gsl_result_15.v_3;
  let slope = gsl_result_15.v_4;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = result;
  gsl_result.v_2 = tone;
  gsl_result.v_3 = height;
  gsl_result.v_4 = slope;
  return gsl_result;
}

fn machined_layFinish(baseMaterial: mat4x4<f32>, direction: vec3<f32>, along: f32, across: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let alpha = sqrt(along * across);
  let squeeze = clamp(along / across, 0.1, 1.0);
  let stretch = (1.0 - squeeze) / 0.9;
  let unit = select(vec3<f32>(0.0, 0.0, 0.0), direction / max(length(direction), 0.000001), stretch > 0.0);
  return material_withLay(material_withRoughness(baseMaterial, sqrt(alpha), noiseTex), unit, stretch, noiseTex);
}

fn machined_nosePitch(s: f32, spacing: f32, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let x = (s - 0.5) * spacing;
  return sqrt(max(radius * radius - x * x, 0.0)) - sqrt(max(radius * radius - 0.25 * spacing * spacing, 0.0));
}

fn machined_grooveFinish(baseMaterial: mat4x4<f32>, direction: vec3<f32>, s: f32, spacing: f32, radius: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let slope = 0.408 * spacing / radius;
  var gsl_simpleValue = material_addDisplacement(machined_layFinish(baseMaterial, direction, 0.02, sqrt(0.0004 + slope * slope), noiseTex), -spacing * spacing / (12.0 * radius), noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let across = texture_reliefRoughness(0.02, slope, spacing, noiseTex);
    let fade = texture_fadeDetail(spacing, noiseTex);
    let gsl_simpleFull = material_addDisplacement(machined_layFinish(baseMaterial, direction, 0.02, across, noiseTex), -machined_nosePitch(s, spacing, radius, noiseTex) * fade, noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let finish = gsl_simpleValue;
  return finish;
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 18> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-56.01, 8.99, -6.01);
  let gsl_partHi = vec3<f32>(-43.99, 51.01, 6.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-50.0, 15.0, 0.0), vec3<f32>(-50.0, 45.0, 0.0), 6.0, noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let clearBar = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-41.01, 8.99, -6.01);
  let gsl_partHi_2 = vec3<f32>(-28.99, 51.01, 6.01);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_gcl_closure_cylinderCapped_2(p, vec3<f32>(-35.0, 15.0, 0.0), vec3<f32>(-35.0, 45.0, 0.0), 6.0, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let blueBar = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-26.01, 8.99, -6.01);
  let gsl_partHi_3 = vec3<f32>(-13.99, 51.01, 6.01);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_gcl_closure_cylinderCapped_3(p, vec3<f32>(-20.0, 15.0, 0.0), vec3<f32>(-20.0, 45.0, 0.0), 6.0, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let redBar = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-11.01, 8.99, -6.01);
  let gsl_partHi_4 = vec3<f32>(1.01, 51.01, 6.01);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_gcl_closure_cylinderCapped_4(p, vec3<f32>(-5.0, 15.0, 0.0), vec3<f32>(-5.0, 45.0, 0.0), 6.0, noiseTex) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let blackBar = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(9.99, 17.99, -6.01);
  let gsl_partHi_5 = vec3<f32>(30.01, 42.01, 6.01);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = builtin_gcl_closure_box(p, vec3<f32>(10.0, 18.0, -6.0), vec3<f32>(30.0, 42.0, 6.0), noiseTex) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let hardBlock = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(34.99, 17.99, -6.01);
  let gsl_partHi_6 = vec3<f32>(55.01, 42.01, 6.01);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = builtin_gcl_closure_box_2(p, vec3<f32>(35.0, 18.0, -6.0), vec3<f32>(55.0, 42.0, 6.0), noiseTex) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let oxideBlock = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(59.99, 17.99, -6.01);
  let gsl_partHi_7 = vec3<f32>(80.01, 42.01, 6.01);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = builtin_gcl_closure_box_3(p, vec3<f32>(60.0, 18.0, -6.0), vec3<f32>(80.0, 42.0, 6.0), noiseTex) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let phosBlock = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-55.01, -35.01, -3.01);
  let gsl_partHi_8 = vec3<f32>(-34.99, -14.99, 3.01);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = builtin_gcl_closure_boxCentered_4(p, vec3<f32>(-45.0, -25.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let glossPlate = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(-30.01, -35.01, -3.01);
  let gsl_partHi_9 = vec3<f32>(-9.99, -14.99, 3.01);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = builtin_gcl_closure_boxCentered_5(p, vec3<f32>(-20.0, -25.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let wrinklePlate = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(-5.01, -35.01, -3.01);
  let gsl_partHi_10 = vec3<f32>(15.01, -14.99, 3.01);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = builtin_gcl_closure_boxCentered_6(p, vec3<f32>(5.0, -25.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let sandPlate = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(19.99, -35.01, -3.01);
  let gsl_partHi_11 = vec3<f32>(40.01, -14.99, 3.01);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = builtin_gcl_closure_boxCentered_7(p, vec3<f32>(30.0, -25.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let hammerPlate = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(49.99, -38.01, -8.01);
  let gsl_partHi_12 = vec3<f32>(75.01, -11.99, 12.01);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = builtin_gcl_closure_box_4(p, vec3<f32>(50.0, -38.0, -8.0), vec3<f32>(75.0, -12.0, 12.0), noiseTex) + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let paintBox = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(-55.01, -71.01, -1.01);
  let gsl_partHi_13 = vec3<f32>(-4.99, -48.99, 1.01);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = builtin_gcl_closure_boxCentered_9(p, vec3<f32>(-30.0, -60.0, 0.0), vec3<f32>(50.0, 22.0, 2.0), noiseTex) + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let freshSheet = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(4.99, -71.01, -1.01);
  let gsl_partHi_14 = vec3<f32>(55.01, -48.99, 1.01);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = builtin_gcl_closure_boxCentered_10(p, vec3<f32>(30.0, -60.0, 0.0), vec3<f32>(50.0, 22.0, 2.0), noiseTex) + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let agedSheet = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(-55.01, -10.01, -3.01);
  let gsl_partHi_15 = vec3<f32>(-34.99, 10.01, 3.01);
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = builtin_gcl_closure_boxCentered_11(p, vec3<f32>(-45.0, 0.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let finePlate = gsl_part_15;
  let gsl_partLo_16 = vec3<f32>(-30.01, -10.01, -3.01);
  let gsl_partHi_16 = vec3<f32>(-9.99, 10.01, 3.01);
  let gsl_partBox_16 = gsl_boxDistance(p, gsl_partLo_16, gsl_partHi_16);
  var gsl_part_16 = gsl_partBox_16;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_16 <= gsl_vanishView.w) {
    let gsl_partVanish_16 = gsl_vanish(gsl_partLo_16, gsl_partHi_16);
    gsl_part_16 = 1e10;
    if (gsl_partVanish_16.x > 0.0) { gsl_part_16 = builtin_gcl_closure_boxCentered_12(p, vec3<f32>(-20.0, 0.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_16.y; }
    if (abs(gsl_part_16) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_16), gsl_partVanish_16.zw); }
  }
  let lacqPlate = gsl_part_16;
  let gsl_partLo_17 = vec3<f32>(-5.01, -10.01, -3.01);
  let gsl_partHi_17 = vec3<f32>(15.01, 10.01, 3.01);
  let gsl_partBox_17 = gsl_boxDistance(p, gsl_partLo_17, gsl_partHi_17);
  var gsl_part_17 = gsl_partBox_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_17 <= gsl_vanishView.w) {
    let gsl_partVanish_17 = gsl_vanish(gsl_partLo_17, gsl_partHi_17);
    gsl_part_17 = 1e10;
    if (gsl_partVanish_17.x > 0.0) { gsl_part_17 = builtin_gcl_closure_boxCentered_13(p, vec3<f32>(5.0, 0.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_17.y; }
    if (abs(gsl_part_17) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_17), gsl_partVanish_17.zw); }
  }
  let coarseLacqPlate = gsl_part_17;
  let gsl_partLo_18 = vec3<f32>(19.99, -10.01, -3.01);
  let gsl_partHi_18 = vec3<f32>(40.01, 10.01, 3.01);
  let gsl_partBox_18 = gsl_boxDistance(p, gsl_partLo_18, gsl_partHi_18);
  var gsl_part_18 = gsl_partBox_18;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_18 <= gsl_vanishView.w) {
    let gsl_partVanish_18 = gsl_vanish(gsl_partLo_18, gsl_partHi_18);
    gsl_part_18 = 1e10;
    if (gsl_partVanish_18.x > 0.0) { gsl_part_18 = builtin_gcl_closure_boxCentered_14(p, vec3<f32>(30.0, 0.0, 0.0), vec3<f32>(20.0, 20.0, 6.0), noiseTex) + gsl_partVanish_18.y; }
    if (abs(gsl_part_18) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_18), gsl_partVanish_18.zw); }
  }
  let coarsePowderPlate = gsl_part_18;
  return array<f32, 18>(clearBar, blueBar, redBar, blackBar, hardBlock, oxideBlock, phosBlock, glossPlate, wrinklePlate, sandPlate, hammerPlate, paintBox, freshSheet, agedSheet, finePlate, lacqPlate, coarseLacqPlate, coarsePowderPlate);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7], 8u);
  let gsl_s9 = gsl_shape(gsl_d[8], 9u);
  let gsl_s10 = gsl_shape(gsl_d[9], 10u);
  let gsl_s11 = gsl_shape(gsl_d[10], 11u);
  let gsl_s12 = gsl_shape(gsl_d[11], 12u);
  let gsl_s13 = gsl_shape(gsl_d[12], 13u);
  let gsl_s14 = gsl_shape(gsl_d[13], 14u);
  let gsl_s15 = gsl_shape(gsl_d[14], 15u);
  let gsl_s16 = gsl_shape(gsl_d[15], 16u);
  let gsl_s17 = gsl_shape(gsl_d[16], 17u);
  let gsl_s18 = gsl_shape(gsl_d[17], 18u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), noiseTex);
    let clearColor = gsl_result.v_2;
    return vec4<f32>(clearColor, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_2 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.15, 0.35, 0.85), noiseTex);
    let blueColor = gsl_result_2.v_2;
    return vec4<f32>(blueColor, 1.0);
  }
  if (gsl_colorIdx == 3u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_3 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.85, 0.12, 0.12), noiseTex);
    let redColor = gsl_result_3.v_2;
    return vec4<f32>(redColor, 1.0);
  }
  if (gsl_colorIdx == 4u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_4 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.06, 0.06, 0.07), noiseTex);
    let blackColor = gsl_result_4.v_2;
    return vec4<f32>(blackColor, 1.0);
  }
  if (gsl_colorIdx == 5u) {
    let millMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) * sqrt(625.0 - clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975)) - normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975), fract((dot(p - vec3<f32>(20.0, 30.0, 6.0), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) - sqrt(625.0 - clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975))) / 0.3), 0.3, 0.8, noiseTex);
    let gsl_result_5 = coating_hardAnodized(millMat, vec3<f32>(0.9621, 0.9654, 0.9658), 50.0, noiseTex);
    let hardColor = gsl_result_5.v_2;
    return vec4<f32>(hardColor, 1.0);
  }
  if (gsl_colorIdx == 6u) {
    let steelMillMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) * sqrt(625.0 - clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975)) - normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975), fract((dot(p - vec3<f32>(45.0, 30.0, 6.0), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) - sqrt(625.0 - clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975))) / 0.3), 0.3, 0.8, noiseTex);
    let gsl_result_6 = coating_blackOxide(steelMillMat, 1.0, noiseTex);
    let oxideColor = gsl_result_6.v_2;
    return vec4<f32>(oxideColor, 1.0);
  }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(0.07, 0.07, 0.07499999), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(vec3<f32>(0.62, 0.06, 0.05), 1.0); }
  if (gsl_colorIdx == 9u) { return vec4<f32>(vec3<f32>(0.04, 0.04, 0.045), 1.0); }
  if (gsl_colorIdx == 10u) { return vec4<f32>(vec3<f32>(0.12, 0.22, 0.12), 1.0); }
  if (gsl_colorIdx == 11u) {
    var gsl_detailValue = 1.0;
    if (max(gsl_pw, gsl_detailPw) < 2.0) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_3 = p / 4.0;
      gsl_detailValue = texture_fadeTo(0.85 + 0.3 * fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_cellularScan(gsl_tmp_3, 1, noiseTex).nearCell + vec3<f32>(41.0, 7.0, 13.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_cellularScan(gsl_tmp_3, 1, noiseTex).nearCell + vec3<f32>(41.0, 7.0, 13.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 4.0, noiseTex);
      gsl_pw = gsl_detailOuterPw;
    }
    var gsl_simpleValue = vec3<f32>(0.18, 0.32, 0.45);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue = mix(gsl_simpleValue, vec3<f32>(0.18, 0.32, 0.45) * gsl_detailValue, gsl_simpleDetail);
    }
    let hammerColor = gsl_simpleValue;
    return vec4<f32>(hammerColor, 1.0);
  }
  if (gsl_colorIdx == 12u) {
    let slabX = max(p.x - 75.0, 50.0 - p.x);
    let slabY = max(p.y - -12.0, -38.0 - p.y);
    let slabZ = max(p.z - 12.0, -8.0 - p.z);
    let wear = max(coating_edgeWear(slabX, slabY, 3.0, noiseTex), max(coating_edgeWear(slabX, slabZ, 3.0, noiseTex), coating_edgeWear(slabY, slabZ, 3.0, noiseTex)));
    let gsl_result_10 = coating_chippedPaint(p, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.14, 0.32, 0.18), vec3<f32>(0.55, 0.22, 0.14), 60.0, wear, 1.5, noiseTex);
    let paintColor = gsl_result_10.v_2;
    return vec4<f32>(paintColor, 1.0);
  }
  if (gsl_colorIdx == 13u) {
    var gsl_simpleValue_2 = color_mixOklab(vec3<f32>(0.8, 0.82, 0.84), vec3<f32>(0.58, 0.6, 0.61), clamp(0.0, 0.0, 1.0), noiseTex);
    if (gsl_simpleDetail > 0.0) {
      let gsl_result_11 = noise_cellularCell(p, 1.0, noiseTex);
      let spangleCell = gsl_result_11.v_3;
      let spangleF2 = gsl_result_11.v_2;
      let spangleF1 = gsl_result_11.v_1;
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_mixOklab(vec3<f32>(0.8, 0.82, 0.84) * texture_fadeTo(0.92 + 0.16 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 1.0, noiseTex) * (1.0 - 0.5 * texture_coverageLine(spangleF2 - spangleF1, 0.02, noiseTex)), vec3<f32>(0.58, 0.6, 0.61), clamp(0.0, 0.0, 1.0), noiseTex), gsl_simpleDetail);
    }
    let freshColor = gsl_simpleValue_2;
    return vec4<f32>(freshColor, 1.0);
  }
  if (gsl_colorIdx == 14u) {
    var gsl_simpleValue_3 = color_mixOklab(vec3<f32>(0.8, 0.82, 0.84), vec3<f32>(0.58, 0.6, 0.61), clamp(0.7, 0.0, 1.0), noiseTex);
    if (gsl_simpleDetail > 0.0) {
      let gsl_result_12 = noise_cellularCell(p, 1.0, noiseTex);
      let spangleCell_2 = gsl_result_12.v_3;
      let spangleF2_2 = gsl_result_12.v_2;
      let spangleF1_2 = gsl_result_12.v_1;
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, color_mixOklab(vec3<f32>(0.8, 0.82, 0.84) * texture_fadeTo(0.92 + 0.16 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 1.0, noiseTex) * (1.0 - 0.5 * texture_coverageLine(spangleF2_2 - spangleF1_2, 0.02, noiseTex)), vec3<f32>(0.58, 0.6, 0.61), clamp(0.7, 0.0, 1.0), noiseTex), gsl_simpleDetail);
    }
    let agedColor = gsl_simpleValue_3;
    return vec4<f32>(agedColor, 1.0);
  }
  if (gsl_colorIdx == 15u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 16u) {
    let fineCutMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(-20.0, 0.0, 3.0)), fract(length(p - vec3<f32>(-20.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(-20.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_13 = coating_clearCoat(fineCutMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), 45.0, 0.08, 0.15, noiseTex);
    let lacqCol = gsl_result_13.v_2;
    return vec4<f32>(lacqCol, 1.0);
  }
  if (gsl_colorIdx == 17u) {
    let coarseCutMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(5.0, 0.0, 3.0)), fract(length(p - vec3<f32>(5.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(5.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 1.2), 1.2, 0.8, noiseTex);
    let gsl_result_14 = coating_clearCoat(coarseCutMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), 45.0, 0.08, 1.2, noiseTex);
    let coarseLacqCol = gsl_result_14.v_2;
    return vec4<f32>(coarseLacqCol, 1.0);
  }
  if (gsl_colorIdx == 18u) {
    let coarsePowderBase = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(30.0, 0.0, 3.0)), fract(length(p - vec3<f32>(30.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(30.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 1.2), 1.2, 0.8, noiseTex);
    let gsl_result_16 = coating_powderCoat(p, coarsePowderBase, vec3<f32>(0.16, 0.3, 0.42), 85.0, 0.5, 0.08, 1.2, noiseTex);
    let coarsePowderCol = gsl_result_16.v_2;
    return vec4<f32>(coarsePowderCol, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), noiseTex);
    let clearMat = gsl_result.v_1;
    let gsl_m = clearMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_2 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.15, 0.35, 0.85), noiseTex);
    let blueMat = gsl_result_2.v_1;
    let gsl_m = blueMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_3 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.85, 0.12, 0.12), noiseTex);
    let redMat = gsl_result_3.v_1;
    let gsl_m = redMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
    let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_4 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.06, 0.06, 0.07), noiseTex);
    let blackMat = gsl_result_4.v_1;
    let gsl_m = blackMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) {
    let millMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) * sqrt(625.0 - clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975)) - normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975), fract((dot(p - vec3<f32>(20.0, 30.0, 6.0), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) - sqrt(625.0 - clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975))) / 0.3), 0.3, 0.8, noiseTex);
    let gsl_result_5 = coating_hardAnodized(millMat, vec3<f32>(0.9621, 0.9654, 0.9658), 50.0, noiseTex);
    let hardMat = gsl_result_5.v_1;
    let gsl_m = hardMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    let steelMillMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) * sqrt(625.0 - clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975)) - normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975), fract((dot(p - vec3<f32>(45.0, 30.0, 6.0), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) - sqrt(625.0 - clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975))) / 0.3), 0.3, 0.8, noiseTex);
    let gsl_result_6 = coating_blackOxide(steelMillMat, 1.0, noiseTex);
    let oxideMat = gsl_result_6.v_1;
    let gsl_m = oxideMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.8, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) {
    var gsl_detailValue_3 = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 0.15) { let gsl_detailOuterPw_3 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      gsl_detailValue_3 = noise_gradientLattice(p / 0.3, noiseTex) * 0.0006 * clamp(0.7, 0.0, 1.0) * (1.0 - texture_mask(gsl_pw, 0.1125, 0.0375, noiseTex));
      gsl_pw = gsl_detailOuterPw_3;
    }
    var gsl_detailValue_2 = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 0.75) { let gsl_detailOuterPw_2 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      gsl_detailValue_2 = (noise_gradientLattice(p / 1.5, noiseTex) + 0.5 * noise_gradientLattice(p / 0.65, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.24375, 0.08125, noiseTex))) * 0.0065 * clamp(0.7, 0.0, 1.0) * (1.0 - texture_mask(gsl_pw, 0.5625, 0.1875, noiseTex)) + gsl_detailValue_3;
      gsl_pw = gsl_detailOuterPw_2;
    }
    var gsl_simpleValue_4 = material_make(sqrt(sqrt(clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) + 0.00001124 * clamp(0.7, 0.0, 1.0) * clamp(0.7, 0.0, 1.0))), 0.0, 1.5, noiseTex);
    if (gsl_simpleDetail > 0.0) {
      let gsl_simpleFull = material_withDisplacement(material_make(sqrt(texture_reliefRoughness(texture_reliefRoughness(clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0), 0.001 * clamp(0.7, 0.0, 1.0), 0.3, noiseTex), 0.0032 * clamp(0.7, 0.0, 1.0), 1.5, noiseTex)), 0.0, 1.5, noiseTex), gsl_detailValue_2, noiseTex);
      gsl_simpleValue_4 = mat4x4<f32>(mix(gsl_simpleValue_4[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_4[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_4[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_4[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let glossMat = gsl_simpleValue_4;
    let gsl_m = glossMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 9u) {
    var gsl_detailValue_4 = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 0.3) { let gsl_detailOuterPw_4 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      gsl_detailValue_4 = (0.234 - (abs(noise_gradientLattice(p / 0.6, noiseTex)) + 0.5 * abs(noise_gradientLattice(p / 0.2857143, noiseTex)))) * 0.09 * (1.0 - texture_mask(gsl_pw, 0.225, 0.075, noiseTex));
      gsl_pw = gsl_detailOuterPw_4;
    }
    var gsl_simpleValue_5 = mat4x4<f32>(vec4<f32>(0.7338368, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_simpleFull_2 = material_withDisplacement(material_make(sqrt(sqrt(0.04 + (1.0 - (1.0 - texture_mask(gsl_pw, 0.225, 0.075, noiseTex)) * gsl_dispDrawn) * 0.25)), 0.0, 1.5, noiseTex), gsl_detailValue_4, noiseTex);
      gsl_simpleValue_5 = mat4x4<f32>(mix(gsl_simpleValue_5[0], gsl_simpleFull_2[0], gsl_simpleDetail), mix(gsl_simpleValue_5[1], gsl_simpleFull_2[1], gsl_simpleDetail), mix(gsl_simpleValue_5[2], gsl_simpleFull_2[2], gsl_simpleDetail), mix(gsl_simpleValue_5[3], gsl_simpleFull_2[3], gsl_simpleDetail));
    }
    let wrinkleMat = gsl_simpleValue_5;
    let gsl_m = wrinkleMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 10u) {
    var gsl_detailValue_5 = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 0.1) { let gsl_detailOuterPw_5 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_2 = clamp(1.0 - noise_cellularF1Lattice(p / 0.2, noiseTex) * 0.2 / 0.09, 0.0, 1.0);
      gsl_detailValue_5 = gsl_tmp_2 * gsl_tmp_2 * 0.08 * (1.0 - texture_mask(gsl_pw, 0.075, 0.025, noiseTex));
      gsl_pw = gsl_detailOuterPw_5;
    }
    var gsl_simpleValue_6 = mat4x4<f32>(vec4<f32>(0.81903625, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_simpleFull_3 = material_withDisplacement(material_make(sqrt(sqrt(0.09 + (1.0 - (1.0 - texture_mask(gsl_pw, 0.075, 0.025, noiseTex)) * gsl_dispDrawn) * 0.36)), 0.0, 1.5, noiseTex), gsl_detailValue_5, noiseTex);
      gsl_simpleValue_6 = mat4x4<f32>(mix(gsl_simpleValue_6[0], gsl_simpleFull_3[0], gsl_simpleDetail), mix(gsl_simpleValue_6[1], gsl_simpleFull_3[1], gsl_simpleDetail), mix(gsl_simpleValue_6[2], gsl_simpleFull_3[2], gsl_simpleDetail), mix(gsl_simpleValue_6[3], gsl_simpleFull_3[3], gsl_simpleDetail));
    }
    let sandMat = gsl_simpleValue_6;
    let gsl_m = sandMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 11u) {
    var gsl_detailValue_6 = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 2.0) { let gsl_detailOuterPw_6 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_3 = p / 4.0;
      let gsl_tmp_4 = clamp(gsl_cellularScan(gsl_tmp_3, 1, noiseTex).f1 * 4.0 / 2.8, 0.0, 1.0);
      gsl_detailValue_6 = (gsl_tmp_4 * gsl_tmp_4 - 0.578) * 0.24 * (1.0 - texture_mask(gsl_pw, 1.5, 0.5, noiseTex));
      gsl_pw = gsl_detailOuterPw_6;
    }
    var gsl_simpleValue_7 = mat4x4<f32>(vec4<f32>(0.38729835, 0.5, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_simpleFull_4 = material_withDisplacement(material_make(sqrt(sqrt(0.0144 + (1.0 - (1.0 - texture_mask(gsl_pw, 1.5, 0.5, noiseTex)) * gsl_dispDrawn) * 0.0081)), 0.5, 1.5, noiseTex), gsl_detailValue_6, noiseTex);
      gsl_simpleValue_7 = mat4x4<f32>(mix(gsl_simpleValue_7[0], gsl_simpleFull_4[0], gsl_simpleDetail), mix(gsl_simpleValue_7[1], gsl_simpleFull_4[1], gsl_simpleDetail), mix(gsl_simpleValue_7[2], gsl_simpleFull_4[2], gsl_simpleDetail), mix(gsl_simpleValue_7[3], gsl_simpleFull_4[3], gsl_simpleDetail));
    }
    let hammerMat = gsl_simpleValue_7;
    let gsl_m = hammerMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 12u) {
    let slabX = max(p.x - 75.0, 50.0 - p.x);
    let slabY = max(p.y - -12.0, -38.0 - p.y);
    let slabZ = max(p.z - 12.0, -8.0 - p.z);
    let wear = max(coating_edgeWear(slabX, slabY, 3.0, noiseTex), max(coating_edgeWear(slabX, slabZ, 3.0, noiseTex), coating_edgeWear(slabY, slabZ, 3.0, noiseTex)));
    let gsl_result_10 = coating_chippedPaint(p, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.14, 0.32, 0.18), vec3<f32>(0.55, 0.22, 0.14), 60.0, wear, 1.5, noiseTex);
    let paintMat = gsl_result_10.v_1;
    let gsl_m = paintMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 13u) {
    var gsl_simpleValue_8 = material_blend(mat4x4<f32>(vec4<f32>(0.325, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.55)), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.0, 0.0, 1.0), noiseTex);
    if (gsl_simpleDetail > 0.0) {
      let gsl_result_11 = noise_cellularCell(p, 1.0, noiseTex);
      let spangleCell = gsl_result_11.v_3;
      let gsl_simpleFull_5 = material_blend(material_withLay(material_make(texture_fadeTo(0.25 + 0.15 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).z, 0.325, 1.0, noiseTex), 1.0, 1.5, noiseTex), normalize(fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb) - vec3<f32>(0.5, 0.5, 0.5) + vec3<f32>(0.0001, 0.0, 0.0)) * (1.0 - texture_mask(gsl_pw, 0.375, 0.125, noiseTex)), texture_fadeTo(0.3 + 0.5 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y, 0.55, 1.0, noiseTex), noiseTex), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.0, 0.0, 1.0), noiseTex);
      gsl_simpleValue_8 = mat4x4<f32>(mix(gsl_simpleValue_8[0], gsl_simpleFull_5[0], gsl_simpleDetail), mix(gsl_simpleValue_8[1], gsl_simpleFull_5[1], gsl_simpleDetail), mix(gsl_simpleValue_8[2], gsl_simpleFull_5[2], gsl_simpleDetail), mix(gsl_simpleValue_8[3], gsl_simpleFull_5[3], gsl_simpleDetail));
    }
    let freshMat = gsl_simpleValue_8;
    let gsl_m = freshMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 14u) {
    var gsl_simpleValue_9 = material_blend(mat4x4<f32>(vec4<f32>(0.325, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.55)), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.7, 0.0, 1.0), noiseTex);
    if (gsl_simpleDetail > 0.0) {
      let gsl_result_12 = noise_cellularCell(p, 1.0, noiseTex);
      let spangleCell_2 = gsl_result_12.v_3;
      let gsl_simpleFull_6 = material_blend(material_withLay(material_make(texture_fadeTo(0.25 + 0.15 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).z, 0.325, 1.0, noiseTex), 1.0, 1.5, noiseTex), normalize(fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb) - vec3<f32>(0.5, 0.5, 0.5) + vec3<f32>(0.0001, 0.0, 0.0)) * (1.0 - texture_mask(gsl_pw, 0.375, 0.125, noiseTex)), texture_fadeTo(0.3 + 0.5 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y, 0.55, 1.0, noiseTex), noiseTex), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.7, 0.0, 1.0), noiseTex);
      gsl_simpleValue_9 = mat4x4<f32>(mix(gsl_simpleValue_9[0], gsl_simpleFull_6[0], gsl_simpleDetail), mix(gsl_simpleValue_9[1], gsl_simpleFull_6[1], gsl_simpleDetail), mix(gsl_simpleValue_9[2], gsl_simpleFull_6[2], gsl_simpleDetail), mix(gsl_simpleValue_9[3], gsl_simpleFull_6[3], gsl_simpleDetail));
    }
    let agedMat = gsl_simpleValue_9;
    let gsl_m = agedMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 15u) {
    let fineMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(-45.0, 0.0, 3.0)), fract(length(p - vec3<f32>(-45.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(-45.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_m = fineMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 16u) {
    let fineCutMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(-20.0, 0.0, 3.0)), fract(length(p - vec3<f32>(-20.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(-20.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
    let gsl_result_13 = coating_clearCoat(fineCutMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), 45.0, 0.08, 0.15, noiseTex);
    let lacqMat = gsl_result_13.v_1;
    let gsl_m = lacqMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 17u) {
    let coarseCutMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(5.0, 0.0, 3.0)), fract(length(p - vec3<f32>(5.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(5.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 1.2), 1.2, 0.8, noiseTex);
    let gsl_result_14 = coating_clearCoat(coarseCutMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), 45.0, 0.08, 1.2, noiseTex);
    let coarseLacqMat = gsl_result_14.v_1;
    let gsl_m = coarseLacqMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 18u) {
    let coarsePowderBase = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(30.0, 0.0, 3.0)), fract(length(p - vec3<f32>(30.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(30.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 1.2), 1.2, 0.8, noiseTex);
    let gsl_result_16 = coating_powderCoat(p, coarsePowderBase, vec3<f32>(0.16, 0.3, 0.42), 85.0, 0.5, 0.08, 1.2, noiseTex);
    let coarsePowderMat = gsl_result_16.v_1;
    let gsl_m = coarsePowderMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

var<private> gsl_dispDrawn: f32 = 0.0;

var<private> gsl_detailPw: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

fn gsl_heights(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 18> {
  let gsl_tmp_1 = p - vec3<f32>(0.0, 0.0, 0.0);
  let turnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(vec3<f32>(0.0, 1.0, 0.0), gsl_tmp_1), fract(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)) / 0.15), 0.15, 0.8, noiseTex);
  let gsl_result = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), noiseTex);
  let clearMat = gsl_result.v_1;
  let gsl_result_2 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.15, 0.35, 0.85), noiseTex);
  let blueMat = gsl_result_2.v_1;
  let gsl_result_3 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.85, 0.12, 0.12), noiseTex);
  let redMat = gsl_result_3.v_1;
  let gsl_result_4 = coating_anodized(turnedMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(0.06, 0.06, 0.07), noiseTex);
  let blackMat = gsl_result_4.v_1;
  let millMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) * sqrt(625.0 - clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975)) - normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975), fract((dot(p - vec3<f32>(20.0, 30.0, 6.0), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) - sqrt(625.0 - clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(20.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975))) / 0.3), 0.3, 0.8, noiseTex);
  let gsl_result_5 = coating_hardAnodized(millMat, vec3<f32>(0.9621, 0.9654, 0.9658), 50.0, noiseTex);
  let hardMat = gsl_result_5.v_1;
  let steelMillMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) * sqrt(625.0 - clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975)) - normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975), fract((dot(p - vec3<f32>(45.0, 30.0, 6.0), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))))) - sqrt(625.0 - clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975) * clamp(dot(p - vec3<f32>(45.0, 30.0, 6.0), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), normalize(vec3<f32>(1.0, 0.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(vec3<f32>(1.0, 0.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))))), -24.975, 24.975))) / 0.3), 0.3, 0.8, noiseTex);
  let gsl_result_6 = coating_blackOxide(steelMillMat, 1.0, noiseTex);
  let oxideMat = gsl_result_6.v_1;
  var gsl_detailValue_7 = 0.0;
  var gsl_detailValue_8 = 0.0;
  if (max(gsl_pw, gsl_detailPw) < 0.15) { let gsl_detailOuterPw_8 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_8 = noise_gradientLattice(p / 0.3, noiseTex) * 0.0006 * clamp(0.7, 0.0, 1.0) * (1.0 - texture_mask(gsl_pw, 0.1125, 0.0375, noiseTex));
    gsl_pw = gsl_detailOuterPw_8;
  }
  if (max(gsl_pw, gsl_detailPw) < 0.75) { let gsl_detailOuterPw_7 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_7 = (noise_gradientLattice(p / 1.5, noiseTex) + 0.5 * noise_gradientLattice(p / 0.65, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.24375, 0.08125, noiseTex))) * 0.0065 * clamp(0.7, 0.0, 1.0) * (1.0 - texture_mask(gsl_pw, 0.5625, 0.1875, noiseTex)) + gsl_detailValue_8;
    gsl_pw = gsl_detailOuterPw_7;
  }
  var gsl_simpleValue_10 = material_make(sqrt(sqrt(clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) + 0.00001124 * clamp(0.7, 0.0, 1.0) * clamp(0.7, 0.0, 1.0))), 0.0, 1.5, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_7 = material_withDisplacement(material_make(sqrt(texture_reliefRoughness(texture_reliefRoughness(clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0) * clamp(0.647 - 0.1243 * log(clamp(85.0, 0.5, 90.0)) + 0.008 * log(1.0 - clamp(85.0, 0.5, 90.0) / 91.5), 0.045, 1.0), 0.001 * clamp(0.7, 0.0, 1.0), 0.3, noiseTex), 0.0032 * clamp(0.7, 0.0, 1.0), 1.5, noiseTex)), 0.0, 1.5, noiseTex), gsl_detailValue_7, noiseTex);
    gsl_simpleValue_10 = mat4x4<f32>(mix(gsl_simpleValue_10[0], gsl_simpleFull_7[0], gsl_simpleDetail), mix(gsl_simpleValue_10[1], gsl_simpleFull_7[1], gsl_simpleDetail), mix(gsl_simpleValue_10[2], gsl_simpleFull_7[2], gsl_simpleDetail), mix(gsl_simpleValue_10[3], gsl_simpleFull_7[3], gsl_simpleDetail));
  }
  let glossMat = gsl_simpleValue_10;
  var gsl_detailValue_9 = 0.0;
  if (max(gsl_pw, gsl_detailPw) < 0.3) { let gsl_detailOuterPw_9 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_9 = (0.234 - (abs(noise_gradientLattice(p / 0.6, noiseTex)) + 0.5 * abs(noise_gradientLattice(p / 0.2857143, noiseTex)))) * 0.09 * (1.0 - texture_mask(gsl_pw, 0.225, 0.075, noiseTex));
    gsl_pw = gsl_detailOuterPw_9;
  }
  var gsl_simpleValue_11 = mat4x4<f32>(vec4<f32>(0.7338368, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_8 = material_withDisplacement(material_make(sqrt(sqrt(0.04 + (1.0 - (1.0 - texture_mask(gsl_pw, 0.225, 0.075, noiseTex)) * gsl_dispDrawn) * 0.25)), 0.0, 1.5, noiseTex), gsl_detailValue_9, noiseTex);
    gsl_simpleValue_11 = mat4x4<f32>(mix(gsl_simpleValue_11[0], gsl_simpleFull_8[0], gsl_simpleDetail), mix(gsl_simpleValue_11[1], gsl_simpleFull_8[1], gsl_simpleDetail), mix(gsl_simpleValue_11[2], gsl_simpleFull_8[2], gsl_simpleDetail), mix(gsl_simpleValue_11[3], gsl_simpleFull_8[3], gsl_simpleDetail));
  }
  let wrinkleMat = gsl_simpleValue_11;
  var gsl_detailValue_10 = 0.0;
  let gsl_tmp_2 = clamp(1.0 - noise_cellularF1Lattice(p / 0.2, noiseTex) * 0.2 / 0.09, 0.0, 1.0);
  if (max(gsl_pw, gsl_detailPw) < 0.1) { let gsl_detailOuterPw_10 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_10 = gsl_tmp_2 * gsl_tmp_2 * 0.08 * (1.0 - texture_mask(gsl_pw, 0.075, 0.025, noiseTex));
    gsl_pw = gsl_detailOuterPw_10;
  }
  var gsl_simpleValue_12 = mat4x4<f32>(vec4<f32>(0.81903625, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_9 = material_withDisplacement(material_make(sqrt(sqrt(0.09 + (1.0 - (1.0 - texture_mask(gsl_pw, 0.075, 0.025, noiseTex)) * gsl_dispDrawn) * 0.36)), 0.0, 1.5, noiseTex), gsl_detailValue_10, noiseTex);
    gsl_simpleValue_12 = mat4x4<f32>(mix(gsl_simpleValue_12[0], gsl_simpleFull_9[0], gsl_simpleDetail), mix(gsl_simpleValue_12[1], gsl_simpleFull_9[1], gsl_simpleDetail), mix(gsl_simpleValue_12[2], gsl_simpleFull_9[2], gsl_simpleDetail), mix(gsl_simpleValue_12[3], gsl_simpleFull_9[3], gsl_simpleDetail));
  }
  let sandMat = gsl_simpleValue_12;
  var gsl_detailValue_11 = 0.0;
  let gsl_tmp_3 = p / 4.0;
  let gsl_tmp_4 = clamp(gsl_cellularScan(gsl_tmp_3, 1, noiseTex).f1 * 4.0 / 2.8, 0.0, 1.0);
  if (max(gsl_pw, gsl_detailPw) < 2.0) { let gsl_detailOuterPw_11 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_11 = (gsl_tmp_4 * gsl_tmp_4 - 0.578) * 0.24 * (1.0 - texture_mask(gsl_pw, 1.5, 0.5, noiseTex));
    gsl_pw = gsl_detailOuterPw_11;
  }
  var gsl_simpleValue_13 = mat4x4<f32>(vec4<f32>(0.38729835, 0.5, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_10 = material_withDisplacement(material_make(sqrt(sqrt(0.0144 + (1.0 - (1.0 - texture_mask(gsl_pw, 1.5, 0.5, noiseTex)) * gsl_dispDrawn) * 0.0081)), 0.5, 1.5, noiseTex), gsl_detailValue_11, noiseTex);
    gsl_simpleValue_13 = mat4x4<f32>(mix(gsl_simpleValue_13[0], gsl_simpleFull_10[0], gsl_simpleDetail), mix(gsl_simpleValue_13[1], gsl_simpleFull_10[1], gsl_simpleDetail), mix(gsl_simpleValue_13[2], gsl_simpleFull_10[2], gsl_simpleDetail), mix(gsl_simpleValue_13[3], gsl_simpleFull_10[3], gsl_simpleDetail));
  }
  let hammerMat = gsl_simpleValue_13;
  let slabX = max(p.x - 75.0, 50.0 - p.x);
  let slabY = max(p.y - -12.0, -38.0 - p.y);
  let slabZ = max(p.z - 12.0, -8.0 - p.z);
  let wear = max(coating_edgeWear(slabX, slabY, 3.0, noiseTex), max(coating_edgeWear(slabX, slabZ, 3.0, noiseTex), coating_edgeWear(slabY, slabZ, 3.0, noiseTex)));
  let gsl_result_10 = coating_chippedPaint(p, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.14, 0.32, 0.18), vec3<f32>(0.55, 0.22, 0.14), 60.0, wear, 1.5, noiseTex);
  let paintMat = gsl_result_10.v_1;
  let gsl_result_11 = noise_cellularCell(p, 1.0, noiseTex);
  let spangleCell = gsl_result_11.v_3;
  var gsl_simpleValue_14 = material_blend(mat4x4<f32>(vec4<f32>(0.325, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.55)), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.0, 0.0, 1.0), noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_11 = material_blend(material_withLay(material_make(texture_fadeTo(0.25 + 0.15 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).z, 0.325, 1.0, noiseTex), 1.0, 1.5, noiseTex), normalize(fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb) - vec3<f32>(0.5, 0.5, 0.5) + vec3<f32>(0.0001, 0.0, 0.0)) * (1.0 - texture_mask(gsl_pw, 0.375, 0.125, noiseTex)), texture_fadeTo(0.3 + 0.5 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y, 0.55, 1.0, noiseTex), noiseTex), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.0, 0.0, 1.0), noiseTex);
    gsl_simpleValue_14 = mat4x4<f32>(mix(gsl_simpleValue_14[0], gsl_simpleFull_11[0], gsl_simpleDetail), mix(gsl_simpleValue_14[1], gsl_simpleFull_11[1], gsl_simpleDetail), mix(gsl_simpleValue_14[2], gsl_simpleFull_11[2], gsl_simpleDetail), mix(gsl_simpleValue_14[3], gsl_simpleFull_11[3], gsl_simpleDetail));
  }
  let freshMat = gsl_simpleValue_14;
  let gsl_result_12 = noise_cellularCell(p, 1.0, noiseTex);
  let spangleCell_2 = gsl_result_12.v_3;
  var gsl_simpleValue_15 = material_blend(mat4x4<f32>(vec4<f32>(0.325, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.55)), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.7, 0.0, 1.0), noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_12 = material_blend(material_withLay(material_make(texture_fadeTo(0.25 + 0.15 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).z, 0.325, 1.0, noiseTex), 1.0, 1.5, noiseTex), normalize(fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb) - vec3<f32>(0.5, 0.5, 0.5) + vec3<f32>(0.0001, 0.0, 0.0)) * (1.0 - texture_mask(gsl_pw, 0.375, 0.125, noiseTex)), texture_fadeTo(0.3 + 0.5 * fract(textureLoad(noiseTex, vec3<i32>(floor(spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((spangleCell_2 + vec3<f32>(17.0, 53.0, 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y, 0.55, 1.0, noiseTex), noiseTex), mat4x4<f32>(vec4<f32>(0.65, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.7, 0.0, 1.0), noiseTex);
    gsl_simpleValue_15 = mat4x4<f32>(mix(gsl_simpleValue_15[0], gsl_simpleFull_12[0], gsl_simpleDetail), mix(gsl_simpleValue_15[1], gsl_simpleFull_12[1], gsl_simpleDetail), mix(gsl_simpleValue_15[2], gsl_simpleFull_12[2], gsl_simpleDetail), mix(gsl_simpleValue_15[3], gsl_simpleFull_12[3], gsl_simpleDetail));
  }
  let agedMat = gsl_simpleValue_15;
  let fineMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(-45.0, 0.0, 3.0)), fract(length(p - vec3<f32>(-45.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(-45.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
  let fineCutMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(-20.0, 0.0, 3.0)), fract(length(p - vec3<f32>(-20.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(-20.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
  let gsl_result_13 = coating_clearCoat(fineCutMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), 45.0, 0.08, 0.15, noiseTex);
  let lacqMat = gsl_result_13.v_1;
  let coarseCutMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(5.0, 0.0, 3.0)), fract(length(p - vec3<f32>(5.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(5.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 1.2), 1.2, 0.8, noiseTex);
  let gsl_result_14 = coating_clearCoat(coarseCutMat, vec3<f32>(0.9621, 0.9654, 0.9658), vec3<f32>(1.0, 1.0, 1.0), 45.0, 0.08, 1.2, noiseTex);
  let coarseLacqMat = gsl_result_14.v_1;
  let coarsePowderBase = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(30.0, 0.0, 3.0)), fract(length(p - vec3<f32>(30.0, 0.0, 3.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(30.0, 0.0, 3.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 1.2), 1.2, 0.8, noiseTex);
  let gsl_result_16 = coating_powderCoat(p, coarsePowderBase, vec3<f32>(0.16, 0.3, 0.42), 85.0, 0.5, 0.08, 1.2, noiseTex);
  let coarsePowderMat = gsl_result_16.v_1;
  return array<f32, 18>(clearMat[2].z, blueMat[2].z, redMat[2].z, blackMat[2].z, hardMat[2].z, oxideMat[2].z, 0.0, glossMat[2].z, wrinkleMat[2].z, sandMat[2].z, hammerMat[2].z, paintMat[2].z, freshMat[2].z, agedMat[2].z, fineMat[2].z, lacqMat[2].z, coarseLacqMat[2].z, coarsePowderMat[2].z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_h = gsl_heights(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  let gsl_s3 = gsl_shape(gsl_d[2] - gsl_h[2] * gsl_dispK, 3u);
  let gsl_s4 = gsl_shape(gsl_d[3] - gsl_h[3] * gsl_dispK, 4u);
  let gsl_s5 = gsl_shape(gsl_d[4] - gsl_h[4] * gsl_dispK, 5u);
  let gsl_s6 = gsl_shape(gsl_d[5] - gsl_h[5] * gsl_dispK, 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7] - gsl_h[7] * gsl_dispK, 8u);
  let gsl_s9 = gsl_shape(gsl_d[8] - gsl_h[8] * gsl_dispK, 9u);
  let gsl_s10 = gsl_shape(gsl_d[9] - gsl_h[9] * gsl_dispK, 10u);
  let gsl_s11 = gsl_shape(gsl_d[10] - gsl_h[10] * gsl_dispK, 11u);
  let gsl_s12 = gsl_shape(gsl_d[11] - gsl_h[11] * gsl_dispK, 12u);
  let gsl_s13 = gsl_shape(gsl_d[12] - gsl_h[12] * gsl_dispK, 13u);
  let gsl_s14 = gsl_shape(gsl_d[13] - gsl_h[13] * gsl_dispK, 14u);
  let gsl_s15 = gsl_shape(gsl_d[14] - gsl_h[14] * gsl_dispK, 15u);
  let gsl_s16 = gsl_shape(gsl_d[15] - gsl_h[15] * gsl_dispK, 16u);
  let gsl_s17 = gsl_shape(gsl_d[16] - gsl_h[16] * gsl_dispK, 17u);
  let gsl_s18 = gsl_shape(gsl_d[17] - gsl_h[17] * gsl_dispK, 18u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18);
}

struct gsl_struct {
  v_1 : mat4x4<f32>,
  v_2 : vec3<f32>,
  v_3 : f32,
  v_4 : f32,
}
struct gsl_struct_2 {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
}
struct gsl_struct_3 {
  v_1 : f32,
  v_2 : f32,
}
