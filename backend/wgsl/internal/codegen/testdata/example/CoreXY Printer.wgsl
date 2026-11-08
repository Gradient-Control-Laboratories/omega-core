fn builtin_gcl_closure_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered(p, center, size, noiseTex);
}

fn builtin_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_gcl_closure_box(p, pointA, pointB, noiseTex);
}

fn gsl_remap(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  let motorCase = max(min(min(min(min(max(nemaSection(p.x, p.y, noiseTex), abs(p.z + 5.0) - 5.0), max(nemaSection(p.x, p.y, noiseTex), abs(p.z + 43.0) - 5.0)), max(length(p.xy) - 11.0, abs(p.z - 1.0) - 1.0)), builtin_gcl_closure_boxCentered(p, vec3<f32>(47.15, 0.0, -86.0) * 0.5, abs(vec3<f32>(4.85, 16.0, 8.0)), noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(18.0, 0.0, -58.0), vec3<f32>(25.0, 0.0, -58.0), 4.0, noiseTex)), -max(length(vec2<f32>(abs(p.x) - 15.5, abs(p.y) - 15.5)) - 1.5, -4.5 - p.z));
  return motorCase;
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn gsl_remap_3(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_1 = nemaSection(p.x, p.y, noiseTex);
  let motorCore = min(max(max(gsl_tmp_1 + 0.05, gsl_tmp_1 + 0.05 + 0.3 * tri(p.z / 1.2, noiseTex)), abs(p.z + 24.0) - 14.0), max(max(length(p.xy) - 2.5, abs(p.z - 12.0) - 12.0), min(p.x - 2.0, p.z - 9.0)));
  return motorCore;
}

fn gsl_remap_4(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap_3(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn gsl_remap_5(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_2 = length(p.xy) - 19.5;
  let motorEncoder = max(max(gsl_tmp_2, abs(p.z + 58.0) - 10.0), (gsl_tmp_2 - (p.z + 68.0) + 1.5) * 0.7071);
  return motorEncoder;
}

fn gsl_remap_6(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap_5(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn builtin_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_gcl_closure_cylinderCapped(p, pointA, pointB, radius, noiseTex);
}

fn builtin_coneCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radiusA: f32, radiusB: f32, noiseTex: texture_3d<f32>) -> f32 {
  let papa = dot(p - pointA, p - pointA);
  let baba = dot(pointB - pointA, pointB - pointA);
  let paba = dot(p - pointA, pointB - pointA) / baba;
  let x = sqrt(papa - paba * paba * baba);
  let rba = radiusB - radiusA;
  let k = rba * rba + baba;
  let f = clamp((rba * (x - radiusA) + paba * baba) / k, 0.0, 1.0);
  let cbx = x - radiusA - f * rba;
  let cay = abs(paba - 0.5) - 0.5;
  let s = select(1.0, -1.0, cbx < 0.0 && cay < 0.0);
  let cax = max(0.0, x - select(radiusB, radiusA, paba < 0.5));
  let cby = paba - f;
  return s * sqrt(min(cax * cax + cay * cay * baba, cbx * cbx + cby * cby * baba));
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn gsl_remap_7(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  let m5 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -12.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 12.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5, noiseTex), noiseTex);
  return m5;
}

fn gsl_remap_8(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap_7(vec3<f32>(p.x * 0.00000000000000006123234 - p.z * -1.0, p.y, p.z * 0.00000000000000006123234 + p.x * -1.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn gsl_remap_9(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap_7(vec3<f32>(p.x * 0.00000000000000006123234 - p.z * 1.0, p.y, p.z * 0.00000000000000006123234 + p.x * 1.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn gsl_remap_10(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap_7(vec3<f32>(p.x, p.y * 0.00000000000000006123234 + p.z * 1.0, -p.y * 1.0 + p.z * 0.00000000000000006123234), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn gsl_remap_11(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap_7(vec3<f32>(p.x, p.y * 0.00000000000000006123234 + p.z * -1.0, -p.y * -1.0 + p.z * 0.00000000000000006123234), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn gsl_remap_12(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  let m3 = fastener_screw(max(length(vec3<f32>(p.x, p.y, p.z - -1.6363636)) - 3.2863636, -p.z), max(max(max(length(p.xy) - 1.5, -8.0 - p.z), p.z - 0.825), (length(p.xy) - 1.5 + -p.z - 8.0 + 0.45) * 0.7071), drill_prismVoid(p, 1.65, max(max(abs(p.x) - 1.0, abs(0.5 * p.x + 0.8660254 * p.y) - 1.0), abs(0.5 * p.x - 0.8660254 * p.y) - 1.0), 1.04, noiseTex), noiseTex);
  return m3;
}

fn gsl_remap_13(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  let m5long = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -30.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 30.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5, noiseTex), noiseTex);
  return m5long;
}

fn gsl_remap_14(p: vec3<f32>, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_remap_12(vec3<f32>(p.x, p.y * 0.00000000000000006123234 + p.z * 1.0, -p.y * 1.0 + p.z * 0.00000000000000006123234), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0, noiseTex);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_intersectSharp_2(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound_2(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0, noiseTex);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound_2(vec2<f32>(radialDistance, axialDistance), noiseTex);
  return form;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn fastener_screw(head: f32, shank: f32, drive: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(min(head, shank), -drive);
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

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn nemaSection(x: f32, y: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(abs(x) - 21.15, abs(y) - 21.15), (abs(x) + abs(y) - 36.3) * 0.7071);
}

fn tri(u: f32, noiseTex: texture_3d<f32>) -> f32 {
  return abs(fract(u) - 0.5) * 2.0;
}

fn slot(u: f32, v: f32, noiseTex: texture_3d<f32>) -> f32 {
  let opening = max(abs(u - 9.35) - 1.15, abs(v) - 3.1);
  let cavity = max(abs(u - 6.05) - 2.15, abs(v) - 5.5);
  return min(opening, cavity);
}

fn section2020(a: f32, b: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(max(max(abs(a) - 10.0, abs(b) - 10.0), -slot(abs(a), b, noiseTex)), -slot(abs(b), a, noiseTex)), 2.1 - length(vec2<f32>(a, b)));
}

fn beam2020(a: f32, b: f32, t: f32, span: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(section2020(a, b, noiseTex), abs(t - span / 2.0) - span / 2.0);
}

fn rail(a: f32, b: f32, t: f32, span: f32, w: f32, h: f32, pitch: f32, noiseTex: texture_3d<f32>) -> f32 {
  let body = max(max(abs(a) - w / 2.0, abs(b - h / 2.0) - h / 2.0), abs(t - span / 2.0) - span / 2.0);
  let groove = length(vec2<f32>(abs(a) - w / 2.0, b - 0.7 * h)) - 0.1 * w;
  let k = (fract((t - pitch / 2.0) / pitch + 0.5) - 0.5) * pitch;
  let counterbore = max(length(vec2<f32>(a, k)) - 0.25 * w, 0.6 * h - b);
  let hole = length(vec2<f32>(a, k)) - 0.14 * w;
  return max(max(body, -groove), -min(counterbore, hole));
}

fn carriage(a: f32, b: f32, t: f32, w: f32, h: f32, len: f32, railW: f32, railH: f32, noiseTex: texture_3d<f32>) -> f32 {
  let outer = max(max(abs(a) - w / 2.0, abs(b - (h + 3.0) / 2.0) - (h - 3.0) / 2.0), abs(t) - len / 2.0);
  let channel = max(abs(a) - (railW / 2.0 + 0.5), b - (railH + 0.5));
  let tapped = max(length(vec2<f32>(abs(a) - 0.37 * w, abs(t) - 0.22 * len)) - 1.3, h - 4.0 - b);
  return max(max(outer, -channel), -tapped);
}

fn driveBase(p: vec3<f32>, cx: f32, cy: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pillars = max(max(abs(abs(p.x - cx) - 20.0) - 4.0, abs(abs(p.y - cy) - 20.0) - 4.0), abs(p.z - 246.0) - 10.0);
  return min(builtin_box(p, vec3<f32>(cx - 26.0, cy - 26.0, 200.0), vec3<f32>(cx + 26.0, cy + 26.0, 236.0), noiseTex), pillars);
}

fn jointBase(p: vec3<f32>, gy: f32, side: f32, noiseTex: texture_3d<f32>) -> f32 {
  let x0 = side * 190.0;
  let x1 = side * 236.0;
  let base = builtin_box(p, vec3<f32>(min(x0, x1), gy - 26.0, 213.0), vec3<f32>(max(x0, x1), gy + 26.0, 238.0), noiseTex);
  let pillars = max(max(abs(p.x - side * 213.0) - 4.0, abs(abs(p.y - gy) - 20.0) - 4.0), abs(p.z - 247.0) - 9.0);
  return min(base, pillars);
}

fn idler(p: vec3<f32>, cx: f32, cy: f32, cz: f32, noiseTex: texture_3d<f32>) -> f32 {
  return min(min(builtin_cylinderCapped(p, vec3<f32>(cx, cy, cz - 3.5), vec3<f32>(cx, cy, cz + 3.5), 6.0, noiseTex), builtin_cylinderCapped(p, vec3<f32>(cx, cy, cz - 4.5), vec3<f32>(cx, cy, cz - 3.5), 9.0, noiseTex)), builtin_cylinderCapped(p, vec3<f32>(cx, cy, cz + 3.5), vec3<f32>(cx, cy, cz + 4.5), 9.0, noiseTex));
}

fn beltY(p: vec3<f32>, x: f32, cz: f32, y0: f32, y1: f32, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_box(p, vec3<f32>(x - 0.75, y0, cz - 3.0), vec3<f32>(x + 0.75, y1, cz + 3.0), noiseTex);
}

fn beltX(p: vec3<f32>, y: f32, cz: f32, x0: f32, x1: f32, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_box(p, vec3<f32>(x0, y - 0.75, cz - 3.0), vec3<f32>(x1, y + 0.75, cz + 3.0), noiseTex);
}

fn leadScrew(p: vec3<f32>, sx: f32, sy: f32, noiseTex: texture_3d<f32>) -> f32 {
  let r = length(vec2<f32>(p.x - sx, p.y - sy)) - 4.0;
  let u = (p.z - -300.0 - 8.0 * atan2(p.y - sy, p.x - sx) / 6.2831855) / 2.0;
  return max(max(r, r + 0.4 * tri(u, noiseTex)), abs(p.z - -1.0) - 171.0);
}

fn leadNut(p: vec3<f32>, bz: f32, sx: f32, sy: f32, noiseTex: texture_3d<f32>) -> f32 {
  return min(builtin_cylinderCapped(p, vec3<f32>(sx, sy, bz - 28.0), vec3<f32>(sx, sy, bz - 24.5), 11.0, noiseTex), builtin_cylinderCapped(p, vec3<f32>(sx, sy, bz - 43.0), vec3<f32>(sx, sy, bz - 28.0), 5.1, noiseTex));
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let carriageX = gsl_param_54_13;
  let gx = carriageX;
  let gantryY = gsl_param_51_11;
  let gy = gantryY;
  let headY = gy - 46.0;
  let bedHeight = gsl_param_48_13;
  let bz = -300.0 + bedHeight;
  let gsl_partLo = vec3<f32>(min(-230.0, min(min(min(min(-219.5, min(gx - 10.0, gx + 10.0)), -200.0), 10.0), 29.0)), min(min(min(min(min(-230.0, -(-0.0 - gy) - 10.0), -190.0), -190.0), 150.0), min(min(min(min(min(min(min(min(min(min(-209.0, min(min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0), min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0))), min(min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0), min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0))), -194.5), -194.5), 175.5), -155.0), min(headY - 8.0, headY + 8.0)), 45.0), -146.0), -265.0)), min(min(min(min(-300.0, -(-0.0 - (bz - 21.0)) - 10.0), -(-0.0 - (bz - 21.0)) - 10.0), -(-0.0 - (bz - 21.0)) - 10.0), min(min(min(min(min(-268.0, min(bz - 9.0, bz - 1.0)), 126.0), -277.0), -255.4), -257.0)));
  let gsl_partHi = vec3<f32>(max(230.0, max(max(max(max(219.5, max(gx - 10.0, gx + 10.0)), 15.0), 100.0), 43.0)), max(max(max(max(max(230.0, -(-0.0 - gy) + 10.0), 190.0), 190.0), 170.0), max(max(max(max(max(max(max(max(max(max(219.5, max(max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0), max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0))), max(max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0), max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0))), -175.5), -175.5), 194.5), 155.0), max(headY - 8.0, headY + 8.0)), 160.0), -134.0), -243.0)), max(max(max(max(233.0, -(-0.0 - (bz - 21.0)) + 10.0), -(-0.0 - (bz - 21.0)) + 10.0), -(-0.0 - (bz - 21.0)) + 10.0), max(max(max(max(max(328.0, max(bz - 9.0, bz - 1.0)), 136.0), -247.0), -247.4), -243.0)));
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let gsl_partLo_2 = vec3<f32>(-230.0, min(min(min(min(-230.0, -(-0.0 - gy) - 10.0), -190.0), -190.0), 150.0), min(min(min(-300.0, -(-0.0 - (bz - 21.0)) - 10.0), -(-0.0 - (bz - 21.0)) - 10.0), -(-0.0 - (bz - 21.0)) - 10.0));
  let gsl_partHi_2 = vec3<f32>(230.0, max(max(max(max(230.0, -(-0.0 - gy) + 10.0), 190.0), 190.0), 170.0), max(max(max(233.0, -(-0.0 - (bz - 21.0)) + 10.0), -(-0.0 - (bz - 21.0)) + 10.0), -(-0.0 - (bz - 21.0)) + 10.0));
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  let gsl_partLo_3 = vec3<f32>(-230.0, -230.0, -300.0);
  let gsl_partHi_3 = vec3<f32>(-210.0, -210.0, 200.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = beam2020(p.x + 220.0, p.y + 220.0, p.z - -300.0, 500.0, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let postFL = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(210.0, -230.0, -300.0);
  let gsl_partHi_4 = vec3<f32>(230.0, -210.0, 200.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = beam2020(p.x - 220.0, p.y + 220.0, p.z - -300.0, 500.0, noiseTex) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let postFR = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-230.0, 210.0, -300.0);
  let gsl_partHi_5 = vec3<f32>(-210.0, 230.0, 200.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = beam2020(p.x + 220.0, p.y - 220.0, p.z - -300.0, 500.0, noiseTex) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let postBL = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(210.0, 210.0, -300.0);
  let gsl_partHi_6 = vec3<f32>(230.0, 230.0, 200.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = beam2020(p.x - 220.0, p.y - 220.0, p.z - -300.0, 500.0, noiseTex) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let postBR = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(-210.0, -230.0, -300.0);
  let gsl_partHi_7 = vec3<f32>(210.0, -210.0, -280.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = beam2020(p.y + 220.0, p.z - -290.0, p.x + 210.0, 420.0, noiseTex) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let floorFront = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-210.0, 210.0, -300.0);
  let gsl_partHi_8 = vec3<f32>(210.0, 230.0, -280.0);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = beam2020(p.y - 220.0, p.z - -290.0, p.x + 210.0, 420.0, noiseTex) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let floorBack = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(-230.0, -210.0, -300.0);
  let gsl_partHi_9 = vec3<f32>(-210.0, 210.0, -280.0);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = beam2020(p.x + 220.0, p.z - -290.0, p.y + 210.0, 420.0, noiseTex) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let floorLeft = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(210.0, -210.0, -300.0);
  let gsl_partHi_10 = vec3<f32>(230.0, 210.0, -280.0);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = beam2020(p.x - 220.0, p.z - -290.0, p.y + 210.0, 420.0, noiseTex) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let floorRight = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(-210.0, -230.0, -220.0);
  let gsl_partHi_11 = vec3<f32>(210.0, -210.0, -200.0);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = beam2020(p.y + 220.0, p.z - -210.0, p.x + 210.0, 420.0, noiseTex) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let deckFront = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(-210.0, 210.0, -220.0);
  let gsl_partHi_12 = vec3<f32>(210.0, 230.0, -200.0);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = beam2020(p.y - 220.0, p.z - -210.0, p.x + 210.0, 420.0, noiseTex) + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let deckBack = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(-230.0, -210.0, -220.0);
  let gsl_partHi_13 = vec3<f32>(-210.0, 210.0, -200.0);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = beam2020(p.x + 220.0, p.z - -210.0, p.y + 210.0, 420.0, noiseTex) + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let deckLeft = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(210.0, -210.0, -220.0);
  let gsl_partHi_14 = vec3<f32>(230.0, 210.0, -200.0);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = beam2020(p.x - 220.0, p.z - -210.0, p.y + 210.0, 420.0, noiseTex) + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let deckRight = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(-210.0, -230.0, 180.0);
  let gsl_partHi_15 = vec3<f32>(210.0, -210.0, 200.0);
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = beam2020(p.y + 220.0, p.z - 190.0, p.x + 210.0, 420.0, noiseTex) + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let topFront = gsl_part_15;
  let gsl_partLo_16 = vec3<f32>(-210.0, 210.0, 180.0);
  let gsl_partHi_16 = vec3<f32>(210.0, 230.0, 200.0);
  let gsl_partBox_16 = gsl_boxDistance(p, gsl_partLo_16, gsl_partHi_16);
  var gsl_part_16 = gsl_partBox_16;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_16 <= gsl_vanishView.w) {
    let gsl_partVanish_16 = gsl_vanish(gsl_partLo_16, gsl_partHi_16);
    gsl_part_16 = 1e10;
    if (gsl_partVanish_16.x > 0.0) { gsl_part_16 = beam2020(p.y - 220.0, p.z - 190.0, p.x + 210.0, 420.0, noiseTex) + gsl_partVanish_16.y; }
    if (abs(gsl_part_16) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_16), gsl_partVanish_16.zw); }
  }
  let topBack = gsl_part_16;
  let gsl_partLo_17 = vec3<f32>(-230.0, -210.0, 180.0);
  let gsl_partHi_17 = vec3<f32>(-210.0, 210.0, 200.0);
  let gsl_partBox_17 = gsl_boxDistance(p, gsl_partLo_17, gsl_partHi_17);
  var gsl_part_17 = gsl_partBox_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_17 <= gsl_vanishView.w) {
    let gsl_partVanish_17 = gsl_vanish(gsl_partLo_17, gsl_partHi_17);
    gsl_part_17 = 1e10;
    if (gsl_partVanish_17.x > 0.0) { gsl_part_17 = beam2020(p.x + 220.0, p.z - 190.0, p.y + 210.0, 420.0, noiseTex) + gsl_partVanish_17.y; }
    if (abs(gsl_part_17) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_17), gsl_partVanish_17.zw); }
  }
  let topLeft = gsl_part_17;
  let gsl_partLo_18 = vec3<f32>(210.0, -210.0, 180.0);
  let gsl_partHi_18 = vec3<f32>(230.0, 210.0, 200.0);
  let gsl_partBox_18 = gsl_boxDistance(p, gsl_partLo_18, gsl_partHi_18);
  var gsl_part_18 = gsl_partBox_18;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_18 <= gsl_vanishView.w) {
    let gsl_partVanish_18 = gsl_vanish(gsl_partLo_18, gsl_partHi_18);
    gsl_part_18 = 1e10;
    if (gsl_partVanish_18.x > 0.0) { gsl_part_18 = beam2020(p.x - 220.0, p.z - 190.0, p.y + 210.0, 420.0, noiseTex) + gsl_partVanish_18.y; }
    if (abs(gsl_part_18) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_18), gsl_partVanish_18.zw); }
  }
  let topRight = gsl_part_18;
  let gsl_partLo_19 = vec3<f32>(-190.0, -(-0.0 - gy) - 10.0, 213.0);
  let gsl_partHi_19 = vec3<f32>(190.0, -(-0.0 - gy) + 10.0, 233.0);
  let gsl_partBox_19 = gsl_boxDistance(p, gsl_partLo_19, gsl_partHi_19);
  var gsl_part_19 = gsl_partBox_19;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_19 <= gsl_vanishView.w) {
    let gsl_partVanish_19 = gsl_vanish(gsl_partLo_19, gsl_partHi_19);
    gsl_part_19 = 1e10;
    if (gsl_partVanish_19.x > 0.0) { gsl_part_19 = beam2020(p.y - gy, p.z - 223.0, p.x + 190.0, 380.0, noiseTex) + gsl_partVanish_19.y; }
    if (abs(gsl_part_19) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_19), gsl_partVanish_19.zw); }
  }
  let gantryBeam = gsl_part_19;
  let gsl_partLo_20 = vec3<f32>(-110.0, -190.0, -(-0.0 - (bz - 21.0)) - 10.0);
  let gsl_partHi_20 = vec3<f32>(-90.0, 190.0, -(-0.0 - (bz - 21.0)) + 10.0);
  let gsl_partBox_20 = gsl_boxDistance(p, gsl_partLo_20, gsl_partHi_20);
  var gsl_part_20 = gsl_partBox_20;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_20 <= gsl_vanishView.w) {
    let gsl_partVanish_20 = gsl_vanish(gsl_partLo_20, gsl_partHi_20);
    gsl_part_20 = 1e10;
    if (gsl_partVanish_20.x > 0.0) { gsl_part_20 = beam2020(p.x + 100.0, p.z - (bz - 21.0), p.y + 190.0, 380.0, noiseTex) + gsl_partVanish_20.y; }
    if (abs(gsl_part_20) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_20), gsl_partVanish_20.zw); }
  }
  let bedRailL = gsl_part_20;
  let gsl_partLo_21 = vec3<f32>(90.0, -190.0, -(-0.0 - (bz - 21.0)) - 10.0);
  let gsl_partHi_21 = vec3<f32>(110.0, 190.0, -(-0.0 - (bz - 21.0)) + 10.0);
  let gsl_partBox_21 = gsl_boxDistance(p, gsl_partLo_21, gsl_partHi_21);
  var gsl_part_21 = gsl_partBox_21;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_21 <= gsl_vanishView.w) {
    let gsl_partVanish_21 = gsl_vanish(gsl_partLo_21, gsl_partHi_21);
    gsl_part_21 = 1e10;
    if (gsl_partVanish_21.x > 0.0) { gsl_part_21 = beam2020(p.x - 100.0, p.z - (bz - 21.0), p.y + 190.0, 380.0, noiseTex) + gsl_partVanish_21.y; }
    if (abs(gsl_part_21) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_21), gsl_partVanish_21.zw); }
  }
  let bedRailR = gsl_part_21;
  let gsl_partLo_22 = vec3<f32>(-90.0, 150.0, -(-0.0 - (bz - 21.0)) - 10.0);
  let gsl_partHi_22 = vec3<f32>(90.0, 170.0, -(-0.0 - (bz - 21.0)) + 10.0);
  let gsl_partBox_22 = gsl_boxDistance(p, gsl_partLo_22, gsl_partHi_22);
  var gsl_part_22 = gsl_partBox_22;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_22 <= gsl_vanishView.w) {
    let gsl_partVanish_22 = gsl_vanish(gsl_partLo_22, gsl_partHi_22);
    gsl_part_22 = 1e10;
    if (gsl_partVanish_22.x > 0.0) { gsl_part_22 = beam2020(p.y - 160.0, p.z - (bz - 21.0), p.x + 90.0, 180.0, noiseTex) + gsl_partVanish_22.y; }
    if (abs(gsl_part_22) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_22), gsl_partVanish_22.zw); }
  }
  let bedCross = gsl_part_22;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(postFL, postFR), postBL), postBR), floorFront), floorBack), floorLeft), floorRight), deckFront), deckBack), deckLeft), deckRight), topFront), topBack), topLeft), topRight), gantryBeam), bedRailL), bedRailR), bedCross) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let frame = gsl_part_2;
  let gsl_partLo_23 = vec3<f32>(min(min(min(min(-219.5, min(gx - 10.0, gx + 10.0)), -200.0), 10.0), 29.0), min(min(min(min(min(min(min(min(min(min(-209.0, min(min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0), min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0))), min(min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0), min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0))), -194.5), -194.5), 175.5), -155.0), min(headY - 8.0, headY + 8.0)), 45.0), -146.0), -265.0), min(min(min(min(min(-268.0, min(bz - 9.0, bz - 1.0)), 126.0), -277.0), -255.4), -257.0));
  let gsl_partHi_23 = vec3<f32>(max(max(max(max(219.5, max(gx - 10.0, gx + 10.0)), 15.0), 100.0), 43.0), max(max(max(max(max(max(max(max(max(max(219.5, max(max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0), max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0))), max(max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0), max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0))), -175.5), -175.5), 194.5), 155.0), max(headY - 8.0, headY + 8.0)), 160.0), -134.0), -243.0), max(max(max(max(max(328.0, max(bz - 9.0, bz - 1.0)), 136.0), -247.0), -247.4), -243.0));
  let gsl_partBox_23 = gsl_boxDistance(p, gsl_partLo_23, gsl_partHi_23);
  var gsl_part_23 = gsl_partBox_23;
  let gsl_partLo_24 = vec3<f32>(-219.5, 180.5, 308.0);
  let gsl_partHi_24 = vec3<f32>(-180.5, 219.5, 328.0);
  let gsl_partBox_24 = gsl_boxDistance(p, gsl_partLo_24, gsl_partHi_24);
  var gsl_part_24 = gsl_partBox_24;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_24 <= gsl_vanishView.w) {
    let gsl_partVanish_24 = gsl_vanish(gsl_partLo_24, gsl_partHi_24);
    gsl_part_24 = 1e10;
    if (gsl_partVanish_24.x > 0.0) { gsl_part_24 = gsl_remap_6(p - vec3<f32>(-200.0, 200.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_24.y; }
    if (abs(gsl_part_24) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_24), gsl_partVanish_24.zw); }
  }
  let encoderA = gsl_part_24;
  let gsl_partLo_25 = vec3<f32>(180.5, 180.5, 308.0);
  let gsl_partHi_25 = vec3<f32>(219.5, 219.5, 328.0);
  let gsl_partBox_25 = gsl_boxDistance(p, gsl_partLo_25, gsl_partHi_25);
  var gsl_part_25 = gsl_partBox_25;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_25 <= gsl_vanishView.w) {
    let gsl_partVanish_25 = gsl_vanish(gsl_partLo_25, gsl_partHi_25);
    gsl_part_25 = 1e10;
    if (gsl_partVanish_25.x > 0.0) { gsl_part_25 = gsl_remap_6(p - vec3<f32>(200.0, 200.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_25.y; }
    if (abs(gsl_part_25) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_25), gsl_partVanish_25.zw); }
  }
  let encoderB = gsl_part_25;
  let gsl_partLo_26 = vec3<f32>(-204.5, -204.5, -268.0);
  let gsl_partHi_26 = vec3<f32>(-165.5, -165.5, -248.0);
  let gsl_partBox_26 = gsl_boxDistance(p, gsl_partLo_26, gsl_partHi_26);
  var gsl_part_26 = gsl_partBox_26;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_26 <= gsl_vanishView.w) {
    let gsl_partVanish_26 = gsl_vanish(gsl_partLo_26, gsl_partHi_26);
    gsl_part_26 = 1e10;
    if (gsl_partVanish_26.x > 0.0) { gsl_part_26 = gsl_remap_5(p - vec3<f32>(-185.0, -185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_26.y; }
    if (abs(gsl_part_26) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_26), gsl_partVanish_26.zw); }
  }
  let zEncoderFL = gsl_part_26;
  let gsl_partLo_27 = vec3<f32>(165.5, -204.5, -268.0);
  let gsl_partHi_27 = vec3<f32>(204.5, -165.5, -248.0);
  let gsl_partBox_27 = gsl_boxDistance(p, gsl_partLo_27, gsl_partHi_27);
  var gsl_part_27 = gsl_partBox_27;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_27 <= gsl_vanishView.w) {
    let gsl_partVanish_27 = gsl_vanish(gsl_partLo_27, gsl_partHi_27);
    gsl_part_27 = 1e10;
    if (gsl_partVanish_27.x > 0.0) { gsl_part_27 = gsl_remap_5(p - vec3<f32>(185.0, -185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_27.y; }
    if (abs(gsl_part_27) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_27), gsl_partVanish_27.zw); }
  }
  let zEncoderFR = gsl_part_27;
  let gsl_partLo_28 = vec3<f32>(-19.5, 165.5, -268.0);
  let gsl_partHi_28 = vec3<f32>(19.5, 204.5, -248.0);
  let gsl_partBox_28 = gsl_boxDistance(p, gsl_partLo_28, gsl_partHi_28);
  var gsl_part_28 = gsl_partBox_28;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_28 <= gsl_vanishView.w) {
    let gsl_partVanish_28 = gsl_vanish(gsl_partLo_28, gsl_partHi_28);
    gsl_part_28 = 1e10;
    if (gsl_partVanish_28.x > 0.0) { gsl_part_28 = gsl_remap_5(p - vec3<f32>(0.0, 185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_28.y; }
    if (abs(gsl_part_28) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_28), gsl_partVanish_28.zw); }
  }
  let zEncoderB = gsl_part_28;
  let gsl_partLo_29 = vec3<f32>(-208.0, 192.0, 230.5);
  let gsl_partHi_29 = vec3<f32>(-192.0, 208.0, 262.0);
  let gsl_partBox_29 = gsl_boxDistance(p, gsl_partLo_29, gsl_partHi_29);
  var gsl_part_29 = gsl_partBox_29;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_29 <= gsl_vanishView.w) {
    let gsl_partVanish_29 = gsl_vanish(gsl_partLo_29, gsl_partHi_29);
    gsl_part_29 = 1e10;
    if (gsl_partVanish_29.x > 0.0) { gsl_part_29 = min(min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, 200.0, 239.5), vec3<f32>(-200.0, 200.0, 246.5), 6.1, noiseTex), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, 200.0, 238.5), vec3<f32>(-200.0, 200.0, 239.5), 8.0, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, 200.0, 246.5), vec3<f32>(-200.0, 200.0, 254.0), 8.0, noiseTex)) + gsl_partVanish_29.y; }
    if (abs(gsl_part_29) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_29), gsl_partVanish_29.zw); }
  }
  let pulleyA = gsl_part_29;
  let gsl_partLo_30 = vec3<f32>(192.0, 192.0, 238.5);
  let gsl_partHi_30 = vec3<f32>(208.0, 208.0, 270.0);
  let gsl_partBox_30 = gsl_boxDistance(p, gsl_partLo_30, gsl_partHi_30);
  var gsl_part_30 = gsl_partBox_30;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_30 <= gsl_vanishView.w) {
    let gsl_partVanish_30 = gsl_vanish(gsl_partLo_30, gsl_partHi_30);
    gsl_part_30 = 1e10;
    if (gsl_partVanish_30.x > 0.0) { gsl_part_30 = min(min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, 200.0, 247.5), vec3<f32>(200.0, 200.0, 254.5), 6.1, noiseTex), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, 200.0, 246.5), vec3<f32>(200.0, 200.0, 247.5), 8.0, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, 200.0, 254.5), vec3<f32>(200.0, 200.0, 262.0), 8.0, noiseTex)) + gsl_partVanish_30.y; }
    if (abs(gsl_part_30) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_30), gsl_partVanish_30.zw); }
  }
  let pulleyB = gsl_part_30;
  let gsl_partLo_31 = vec3<f32>(-209.0, -209.0, 229.5);
  let gsl_partHi_31 = vec3<f32>(-191.0, -191.0, 264.5);
  let gsl_partBox_31 = gsl_boxDistance(p, gsl_partLo_31, gsl_partHi_31);
  var gsl_part_31 = gsl_partBox_31;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_31 <= gsl_vanishView.w) {
    let gsl_partVanish_31 = gsl_vanish(gsl_partLo_31, gsl_partHi_31);
    gsl_part_31 = 1e10;
    if (gsl_partVanish_31.x > 0.0) { gsl_part_31 = min(min(min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, -200.0, 239.5), vec3<f32>(-200.0, -200.0, 246.5), 6.0, noiseTex), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, -200.0, 238.5), vec3<f32>(-200.0, -200.0, 239.5), 9.0, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, -200.0, 246.5), vec3<f32>(-200.0, -200.0, 247.5), 9.0, noiseTex)), min(min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, -200.0, 247.5), vec3<f32>(-200.0, -200.0, 254.5), 6.0, noiseTex), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, -200.0, 246.5), vec3<f32>(-200.0, -200.0, 247.5), 9.0, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-200.0, -200.0, 254.5), vec3<f32>(-200.0, -200.0, 255.5), 9.0, noiseTex))) + gsl_partVanish_31.y; }
    if (abs(gsl_part_31) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_31), gsl_partVanish_31.zw); }
  }
  let idlersL = gsl_part_31;
  let gsl_partLo_32 = vec3<f32>(191.0, -209.0, 229.5);
  let gsl_partHi_32 = vec3<f32>(209.0, -191.0, 264.5);
  let gsl_partBox_32 = gsl_boxDistance(p, gsl_partLo_32, gsl_partHi_32);
  var gsl_part_32 = gsl_partBox_32;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_32 <= gsl_vanishView.w) {
    let gsl_partVanish_32 = gsl_vanish(gsl_partLo_32, gsl_partHi_32);
    gsl_part_32 = 1e10;
    if (gsl_partVanish_32.x > 0.0) { gsl_part_32 = min(min(min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, -200.0, 239.5), vec3<f32>(200.0, -200.0, 246.5), 6.0, noiseTex), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, -200.0, 238.5), vec3<f32>(200.0, -200.0, 239.5), 9.0, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, -200.0, 246.5), vec3<f32>(200.0, -200.0, 247.5), 9.0, noiseTex)), min(min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, -200.0, 247.5), vec3<f32>(200.0, -200.0, 254.5), 6.0, noiseTex), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, -200.0, 246.5), vec3<f32>(200.0, -200.0, 247.5), 9.0, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(200.0, -200.0, 254.5), vec3<f32>(200.0, -200.0, 255.5), 9.0, noiseTex))) + gsl_partVanish_32.y; }
    if (abs(gsl_part_32) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_32), gsl_partVanish_32.zw); }
  }
  let idlersR = gsl_part_32;
  let gsl_partLo_33 = vec3<f32>(-209.0, min(min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0), min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0)), 229.5);
  let gsl_partHi_33 = vec3<f32>(-191.0, max(max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0), max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0)), 264.5);
  let gsl_partBox_33 = gsl_boxDistance(p, gsl_partLo_33, gsl_partHi_33);
  var gsl_part_33 = gsl_partBox_33;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_33 <= gsl_vanishView.w) {
    let gsl_partVanish_33 = gsl_vanish(gsl_partLo_33, gsl_partHi_33);
    gsl_part_33 = 1e10;
    if (gsl_partVanish_33.x > 0.0) { gsl_part_33 = min(idler(p, -200.0, gy, 243.0, noiseTex), idler(p, -200.0, gy, 251.0, noiseTex)) + gsl_partVanish_33.y; }
    if (abs(gsl_part_33) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_33), gsl_partVanish_33.zw); }
  }
  let jointIdlersL = gsl_part_33;
  let gsl_partLo_34 = vec3<f32>(191.0, min(min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0), min(min(min(gy, gy) - 6.0, min(gy, gy) - 9.0), min(gy, gy) - 9.0)), 229.5);
  let gsl_partHi_34 = vec3<f32>(209.0, max(max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0), max(max(max(gy, gy) + 6.0, max(gy, gy) + 9.0), max(gy, gy) + 9.0)), 264.5);
  let gsl_partBox_34 = gsl_boxDistance(p, gsl_partLo_34, gsl_partHi_34);
  var gsl_part_34 = gsl_partBox_34;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_34 <= gsl_vanishView.w) {
    let gsl_partVanish_34 = gsl_vanish(gsl_partLo_34, gsl_partHi_34);
    gsl_part_34 = 1e10;
    if (gsl_partVanish_34.x > 0.0) { gsl_part_34 = min(idler(p, 200.0, gy, 243.0, noiseTex), idler(p, 200.0, gy, 251.0, noiseTex)) + gsl_partVanish_34.y; }
    if (abs(gsl_part_34) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_34), gsl_partVanish_34.zw); }
  }
  let jointIdlersR = gsl_part_34;
  let gsl_partLo_35 = vec3<f32>(-194.5, -194.5, -206.5);
  let gsl_partHi_35 = vec3<f32>(-175.5, -175.5, -162.5);
  let gsl_partBox_35 = gsl_boxDistance(p, gsl_partLo_35, gsl_partHi_35);
  var gsl_part_35 = gsl_partBox_35;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_35 <= gsl_vanishView.w) {
    let gsl_partVanish_35 = gsl_vanish(gsl_partLo_35, gsl_partHi_35);
    gsl_part_35 = 1e10;
    if (gsl_partVanish_35.x > 0.0) { gsl_part_35 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-185.0, -185.0, -197.0), vec3<f32>(-185.0, -185.0, -172.0), 9.5, noiseTex) + gsl_partVanish_35.y; }
    if (abs(gsl_part_35) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_35), gsl_partVanish_35.zw); }
  }
  let couplerFL = gsl_part_35;
  let gsl_partLo_36 = vec3<f32>(175.5, -194.5, -206.5);
  let gsl_partHi_36 = vec3<f32>(194.5, -175.5, -162.5);
  let gsl_partBox_36 = gsl_boxDistance(p, gsl_partLo_36, gsl_partHi_36);
  var gsl_part_36 = gsl_partBox_36;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_36 <= gsl_vanishView.w) {
    let gsl_partVanish_36 = gsl_vanish(gsl_partLo_36, gsl_partHi_36);
    gsl_part_36 = 1e10;
    if (gsl_partVanish_36.x > 0.0) { gsl_part_36 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(185.0, -185.0, -197.0), vec3<f32>(185.0, -185.0, -172.0), 9.5, noiseTex) + gsl_partVanish_36.y; }
    if (abs(gsl_part_36) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_36), gsl_partVanish_36.zw); }
  }
  let couplerFR = gsl_part_36;
  let gsl_partLo_37 = vec3<f32>(-9.5, 175.5, -206.5);
  let gsl_partHi_37 = vec3<f32>(9.5, 194.5, -162.5);
  let gsl_partBox_37 = gsl_boxDistance(p, gsl_partLo_37, gsl_partHi_37);
  var gsl_part_37 = gsl_partBox_37;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_37 <= gsl_vanishView.w) {
    let gsl_partVanish_37 = gsl_vanish(gsl_partLo_37, gsl_partHi_37);
    gsl_part_37 = 1e10;
    if (gsl_partVanish_37.x > 0.0) { gsl_part_37 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(0.0, 185.0, -197.0), vec3<f32>(0.0, 185.0, -172.0), 9.5, noiseTex) + gsl_partVanish_37.y; }
    if (abs(gsl_part_37) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_37), gsl_partVanish_37.zw); }
  }
  let couplerB = gsl_part_37;
  let gsl_partLo_38 = vec3<f32>(-155.0, -155.0, min(bz - 9.0, bz - 1.0));
  let gsl_partHi_38 = vec3<f32>(155.0, 155.0, max(bz - 9.0, bz - 1.0));
  let gsl_partBox_38 = gsl_boxDistance(p, gsl_partLo_38, gsl_partHi_38);
  var gsl_part_38 = gsl_partBox_38;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_38 <= gsl_vanishView.w) {
    let gsl_partVanish_38 = gsl_vanish(gsl_partLo_38, gsl_partHi_38);
    gsl_part_38 = 1e10;
    if (gsl_partVanish_38.x > 0.0) { gsl_part_38 = builtin_box(p, vec3<f32>(-155.0, -155.0, bz - 9.0), vec3<f32>(155.0, 155.0, bz - 1.0), noiseTex) + gsl_partVanish_38.y; }
    if (abs(gsl_part_38) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_38), gsl_partVanish_38.zw); }
  }
  let bedPlate = gsl_part_38;
  let gsl_partLo_39 = vec3<f32>(min(gx - 10.0, gx + 10.0), min(headY - 8.0, headY + 8.0), 126.0);
  let gsl_partHi_39 = vec3<f32>(max(gx - 10.0, gx + 10.0), max(headY - 8.0, headY + 8.0), 136.0);
  let gsl_partBox_39 = gsl_boxDistance(p, gsl_partLo_39, gsl_partHi_39);
  var gsl_part_39 = gsl_partBox_39;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_39 <= gsl_vanishView.w) {
    let gsl_partVanish_39 = gsl_vanish(gsl_partLo_39, gsl_partHi_39);
    gsl_part_39 = 1e10;
    if (gsl_partVanish_39.x > 0.0) { gsl_part_39 = builtin_box(p, vec3<f32>(gx - 10.0, headY - 8.0, 126.0), vec3<f32>(gx + 10.0, headY + 8.0, 136.0), noiseTex) + gsl_partVanish_39.y; }
    if (abs(gsl_part_39) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_39), gsl_partVanish_39.zw); }
  }
  let heaterBlock = gsl_part_39;
  let gsl_partLo_40 = vec3<f32>(-200.0, 45.0, -277.0);
  let gsl_partHi_40 = vec3<f32>(15.0, 160.0, -247.0);
  let gsl_partBox_40 = gsl_boxDistance(p, gsl_partLo_40, gsl_partHi_40);
  var gsl_part_40 = gsl_partBox_40;
  let psuVents = max(max(abs(fract(p.x / 6.0) - 0.5) * 6.0 - 1.6, abs(fract(p.y / 6.0) - 0.5) * 6.0 - 1.6), -247.8 - p.z);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_40 <= gsl_vanishView.w) {
    let gsl_partVanish_40 = gsl_vanish(gsl_partLo_40, gsl_partHi_40);
    gsl_part_40 = 1e10;
    if (gsl_partVanish_40.x > 0.0) { gsl_part_40 = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-185.0, 205.0, -524.0) * 0.5, abs(vec3<f32>(215.0, 115.0, 30.0)), noiseTex), -psuVents) + gsl_partVanish_40.y; }
    if (abs(gsl_part_40) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_40), gsl_partVanish_40.zw); }
  }
  let psu = gsl_part_40;
  let gsl_partLo_41 = vec3<f32>(10.0, -146.0, -255.4);
  let gsl_partHi_41 = vec3<f32>(100.0, -134.0, -247.4);
  let gsl_partBox_41 = gsl_boxDistance(p, gsl_partLo_41, gsl_partHi_41);
  var gsl_part_41 = gsl_partBox_41;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_41 <= gsl_vanishView.w) {
    let gsl_partVanish_41 = gsl_vanish(gsl_partLo_41, gsl_partHi_41);
    gsl_part_41 = 1e10;
    if (gsl_partVanish_41.x > 0.0) { gsl_part_41 = max(max(max(max(abs(fract((p.x - 10.0) / 25.0) * 25.0 - 7.5) - 6.0, abs(p.y + 140.0) - 6.0), abs(p.z - -251.4) - 4.0), abs(p.x - 55.0) - 45.0), -max(abs(fract(p.y / 2.4) - 0.5) * 2.4 - 0.5, -253.4 - p.z)) + gsl_partVanish_41.y; }
    if (abs(gsl_part_41) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_41), gsl_partVanish_41.zw); }
  }
  let heatSinks = gsl_part_41;
  let gsl_partLo_42 = vec3<f32>(29.0, -265.0, -257.0);
  let gsl_partHi_42 = vec3<f32>(43.0, -243.0, -243.0);
  let gsl_partBox_42 = gsl_boxDistance(p, gsl_partLo_42, gsl_partHi_42);
  var gsl_part_42 = gsl_partBox_42;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_42 <= gsl_vanishView.w) {
    let gsl_partVanish_42 = gsl_vanish(gsl_partLo_42, gsl_partHi_42);
    gsl_part_42 = 1e10;
    if (gsl_partVanish_42.x > 0.0) { gsl_part_42 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(36.0, -258.0, -250.0), vec3<f32>(36.0, -250.0, -250.0), 7.0, noiseTex) + gsl_partVanish_42.y; }
    if (abs(gsl_part_42) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_42), gsl_partVanish_42.zw); }
  }
  let displayKnob = gsl_part_42;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_23 <= gsl_vanishView.w) {
    let gsl_partVanish_23 = gsl_vanish(gsl_partLo_23, gsl_partHi_23);
    gsl_part_23 = 1e10;
    if (gsl_partVanish_23.x > 0.0) { gsl_part_23 = min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(encoderA, encoderB), zEncoderFL), zEncoderFR), zEncoderB), pulleyA), pulleyB), idlersL), idlersR), jointIdlersL), jointIdlersR), couplerFL), couplerFR), couplerB), bedPlate), heaterBlock), psu), heatSinks), displayKnob) + gsl_partVanish_23.y; }
    if (abs(gsl_part_23) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_23), gsl_partVanish_23.zw); }
  }
  let aluminumParts = gsl_part_23;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = min(frame, aluminumParts) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let aluminum = gsl_part;
  let gsl_partLo_43 = vec3<f32>(min(-226.0, min(min(-221.15, min(gx, gx) - 18.0), min(gx, gx) - 7.0)), min(-210.0, min(min(-206.15, min(headY, headY) - 18.0), min(gy - 74.0, gy - 71.0) - 7.0)), min(min(min(min(-280.0, min(bz - 33.0, bz - 28.0)), min(bz - 33.0, bz - 28.0)), min(bz - 33.0, bz - 28.0)), -262.0));
  let gsl_partHi_43 = vec3<f32>(max(226.0, max(max(229.0, max(gx, gx) + 18.0), max(gx, gx) + 7.0)), max(226.0, max(max(221.15, max(headY, headY) + 18.0), max(gy - 74.0, gy - 71.0) + 7.0)), max(max(max(max(260.0, max(bz - 33.0, bz - 28.0)), max(bz - 33.0, bz - 28.0)), max(bz - 33.0, bz - 28.0)), 322.0));
  let gsl_partBox_43 = gsl_boxDistance(p, gsl_partLo_43, gsl_partHi_43);
  var gsl_part_43 = gsl_partBox_43;
  let gsl_partLo_44 = vec3<f32>(-226.0, -210.0, min(min(min(-280.0, min(bz - 33.0, bz - 28.0)), min(bz - 33.0, bz - 28.0)), min(bz - 33.0, bz - 28.0)));
  let gsl_partHi_44 = vec3<f32>(226.0, 226.0, max(max(max(260.0, max(bz - 33.0, bz - 28.0)), max(bz - 33.0, bz - 28.0)), max(bz - 33.0, bz - 28.0)));
  let gsl_partBox_44 = gsl_boxDistance(p, gsl_partLo_44, gsl_partHi_44);
  var gsl_part_44 = gsl_partBox_44;
  let gsl_partLo_45 = vec3<f32>(-226.0, 174.0, 256.0);
  let gsl_partHi_45 = vec3<f32>(-174.0, 226.0, 260.0);
  let gsl_partBox_45 = gsl_boxDistance(p, gsl_partLo_45, gsl_partHi_45);
  var gsl_part_45 = gsl_partBox_45;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_45 <= gsl_vanishView.w) {
    let gsl_partVanish_45 = gsl_vanish(gsl_partLo_45, gsl_partHi_45);
    gsl_part_45 = 1e10;
    if (gsl_partVanish_45.x > 0.0) { gsl_part_45 = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-400.0, 400.0, 516.0) * 0.5, abs(vec3<f32>(52.0, 52.0, 4.0)), noiseTex), 11.5 - length(vec2<f32>(p.x - -200.0, p.y - 200.0))) + gsl_partVanish_45.y; }
    if (abs(gsl_part_45) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_45), gsl_partVanish_45.zw); }
  }
  let drivePlateA = gsl_part_45;
  let gsl_partLo_46 = vec3<f32>(174.0, 174.0, 256.0);
  let gsl_partHi_46 = vec3<f32>(226.0, 226.0, 260.0);
  let gsl_partBox_46 = gsl_boxDistance(p, gsl_partLo_46, gsl_partHi_46);
  var gsl_part_46 = gsl_partBox_46;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_46 <= gsl_vanishView.w) {
    let gsl_partVanish_46 = gsl_vanish(gsl_partLo_46, gsl_partHi_46);
    gsl_part_46 = 1e10;
    if (gsl_partVanish_46.x > 0.0) { gsl_part_46 = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(400.0, 400.0, 516.0) * 0.5, abs(vec3<f32>(52.0, 52.0, 4.0)), noiseTex), 11.5 - length(vec2<f32>(p.x - 200.0, p.y - 200.0))) + gsl_partVanish_46.y; }
    if (abs(gsl_part_46) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_46), gsl_partVanish_46.zw); }
  }
  let drivePlateB = gsl_part_46;
  let gsl_partLo_47 = vec3<f32>(-210.0, -210.0, -200.0);
  let gsl_partHi_47 = vec3<f32>(210.0, 210.0, -197.0);
  let gsl_partBox_47 = gsl_boxDistance(p, gsl_partLo_47, gsl_partHi_47);
  var gsl_part_47 = gsl_partBox_47;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_47 <= gsl_vanishView.w) {
    let gsl_partVanish_47 = gsl_vanish(gsl_partLo_47, gsl_partHi_47);
    gsl_part_47 = 1e10;
    if (gsl_partVanish_47.x > 0.0) { gsl_part_47 = max(max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -397.0) * 0.5, abs(vec3<f32>(420.0, 420.0, 3.0)), noiseTex), 11.0 - length(vec2<f32>(p.x - -185.0, p.y - -185.0))), 11.0 - length(vec2<f32>(p.x - 185.0, p.y - -185.0))), 11.0 - length(vec2<f32>(p.x - 0.0, p.y - 185.0))) + gsl_partVanish_47.y; }
    if (abs(gsl_part_47) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_47), gsl_partVanish_47.zw); }
  }
  let deckPlate = gsl_part_47;
  let gsl_partLo_48 = vec3<f32>(-210.0, -210.0, -280.0);
  let gsl_partHi_48 = vec3<f32>(210.0, 210.0, -277.0);
  let gsl_partBox_48 = gsl_boxDistance(p, gsl_partLo_48, gsl_partHi_48);
  var gsl_part_48 = gsl_partBox_48;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_48 <= gsl_vanishView.w) {
    let gsl_partVanish_48 = gsl_vanish(gsl_partLo_48, gsl_partHi_48);
    gsl_part_48 = 1e10;
    if (gsl_partVanish_48.x > 0.0) { gsl_part_48 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -557.0) * 0.5, abs(vec3<f32>(420.0, 420.0, 3.0)), noiseTex) + gsl_partVanish_48.y; }
    if (abs(gsl_part_48) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_48), gsl_partVanish_48.zw); }
  }
  let bayFloor = gsl_part_48;
  let gsl_partLo_49 = vec3<f32>(-200.0, -200.0, min(bz - 33.0, bz - 28.0));
  let gsl_partHi_49 = vec3<f32>(-90.0, -170.0, max(bz - 33.0, bz - 28.0));
  let gsl_partBox_49 = gsl_boxDistance(p, gsl_partLo_49, gsl_partHi_49);
  var gsl_part_49 = gsl_partBox_49;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_49 <= gsl_vanishView.w) {
    let gsl_partVanish_49 = gsl_vanish(gsl_partLo_49, gsl_partHi_49);
    gsl_part_49 = 1e10;
    if (gsl_partVanish_49.x > 0.0) { gsl_part_49 = builtin_box(p, vec3<f32>(-200.0, -200.0, bz - 33.0), vec3<f32>(-90.0, -170.0, bz - 28.0), noiseTex) + gsl_partVanish_49.y; }
    if (abs(gsl_part_49) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_49), gsl_partVanish_49.zw); }
  }
  let armFL = gsl_part_49;
  let gsl_partLo_50 = vec3<f32>(90.0, -200.0, min(bz - 33.0, bz - 28.0));
  let gsl_partHi_50 = vec3<f32>(200.0, -170.0, max(bz - 33.0, bz - 28.0));
  let gsl_partBox_50 = gsl_boxDistance(p, gsl_partLo_50, gsl_partHi_50);
  var gsl_part_50 = gsl_partBox_50;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_50 <= gsl_vanishView.w) {
    let gsl_partVanish_50 = gsl_vanish(gsl_partLo_50, gsl_partHi_50);
    gsl_part_50 = 1e10;
    if (gsl_partVanish_50.x > 0.0) { gsl_part_50 = builtin_box(p, vec3<f32>(90.0, -200.0, bz - 33.0), vec3<f32>(200.0, -170.0, bz - 28.0), noiseTex) + gsl_partVanish_50.y; }
    if (abs(gsl_part_50) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_50), gsl_partVanish_50.zw); }
  }
  let armFR = gsl_part_50;
  let gsl_partLo_51 = vec3<f32>(-15.0, 150.0, min(bz - 33.0, bz - 28.0));
  let gsl_partHi_51 = vec3<f32>(15.0, 200.0, max(bz - 33.0, bz - 28.0));
  let gsl_partBox_51 = gsl_boxDistance(p, gsl_partLo_51, gsl_partHi_51);
  var gsl_part_51 = gsl_partBox_51;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_51 <= gsl_vanishView.w) {
    let gsl_partVanish_51 = gsl_vanish(gsl_partLo_51, gsl_partHi_51);
    gsl_part_51 = 1e10;
    if (gsl_partVanish_51.x > 0.0) { gsl_part_51 = builtin_box(p, vec3<f32>(-15.0, 150.0, bz - 33.0), vec3<f32>(15.0, 200.0, bz - 28.0), noiseTex) + gsl_partVanish_51.y; }
    if (abs(gsl_part_51) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_51), gsl_partVanish_51.zw); }
  }
  let armB = gsl_part_51;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_44 <= gsl_vanishView.w) {
    let gsl_partVanish_44 = gsl_vanish(gsl_partLo_44, gsl_partHi_44);
    gsl_part_44 = 1e10;
    if (gsl_partVanish_44.x > 0.0) { gsl_part_44 = min(min(min(min(min(min(drivePlateA, drivePlateB), deckPlate), bayFloor), armFL), armFR), armB) + gsl_partVanish_44.y; }
    if (abs(gsl_part_44) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_44), gsl_partVanish_44.zw); }
  }
  let plates = gsl_part_44;
  let gsl_partLo_52 = vec3<f32>(min(min(-221.15, min(gx, gx) - 18.0), min(gx, gx) - 7.0), min(min(-206.15, min(headY, headY) - 18.0), min(gy - 74.0, gy - 71.0) - 7.0), -262.0);
  let gsl_partHi_52 = vec3<f32>(max(max(229.0, max(gx, gx) + 18.0), max(gx, gx) + 7.0), max(max(221.15, max(headY, headY) + 18.0), max(gy - 74.0, gy - 71.0) + 7.0), 322.0);
  let gsl_partBox_52 = gsl_boxDistance(p, gsl_partLo_52, gsl_partHi_52);
  var gsl_part_52 = gsl_partBox_52;
  let gsl_partLo_53 = vec3<f32>(-221.15, 178.85, 258.0);
  let gsl_partHi_53 = vec3<f32>(-171.0, 221.15, 322.0);
  let gsl_partBox_53 = gsl_boxDistance(p, gsl_partLo_53, gsl_partHi_53);
  var gsl_part_53 = gsl_partBox_53;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_53 <= gsl_vanishView.w) {
    let gsl_partVanish_53 = gsl_vanish(gsl_partLo_53, gsl_partHi_53);
    gsl_part_53 = 1e10;
    if (gsl_partVanish_53.x > 0.0) { gsl_part_53 = gsl_remap_2(p - vec3<f32>(-200.0, 200.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_53.y; }
    if (abs(gsl_part_53) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_53), gsl_partVanish_53.zw); }
  }
  let motorCaseA = gsl_part_53;
  let gsl_partLo_54 = vec3<f32>(178.85, 178.85, 258.0);
  let gsl_partHi_54 = vec3<f32>(229.0, 221.15, 322.0);
  let gsl_partBox_54 = gsl_boxDistance(p, gsl_partLo_54, gsl_partHi_54);
  var gsl_part_54 = gsl_partBox_54;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_54 <= gsl_vanishView.w) {
    let gsl_partVanish_54 = gsl_vanish(gsl_partLo_54, gsl_partHi_54);
    gsl_part_54 = 1e10;
    if (gsl_partVanish_54.x > 0.0) { gsl_part_54 = gsl_remap_2(p - vec3<f32>(200.0, 200.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_54.y; }
    if (abs(gsl_part_54) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_54), gsl_partVanish_54.zw); }
  }
  let motorCaseB = gsl_part_54;
  let gsl_partLo_55 = vec3<f32>(-206.15, -206.15, -262.0);
  let gsl_partHi_55 = vec3<f32>(-156.0, -163.85, -198.0);
  let gsl_partBox_55 = gsl_boxDistance(p, gsl_partLo_55, gsl_partHi_55);
  var gsl_part_55 = gsl_partBox_55;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_55 <= gsl_vanishView.w) {
    let gsl_partVanish_55 = gsl_vanish(gsl_partLo_55, gsl_partHi_55);
    gsl_part_55 = 1e10;
    if (gsl_partVanish_55.x > 0.0) { gsl_part_55 = gsl_remap(p - vec3<f32>(-185.0, -185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_55.y; }
    if (abs(gsl_part_55) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_55), gsl_partVanish_55.zw); }
  }
  let zCaseFL = gsl_part_55;
  let gsl_partLo_56 = vec3<f32>(163.85, -206.15, -262.0);
  let gsl_partHi_56 = vec3<f32>(214.0, -163.85, -198.0);
  let gsl_partBox_56 = gsl_boxDistance(p, gsl_partLo_56, gsl_partHi_56);
  var gsl_part_56 = gsl_partBox_56;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_56 <= gsl_vanishView.w) {
    let gsl_partVanish_56 = gsl_vanish(gsl_partLo_56, gsl_partHi_56);
    gsl_part_56 = 1e10;
    if (gsl_partVanish_56.x > 0.0) { gsl_part_56 = gsl_remap(p - vec3<f32>(185.0, -185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_56.y; }
    if (abs(gsl_part_56) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_56), gsl_partVanish_56.zw); }
  }
  let zCaseFR = gsl_part_56;
  let gsl_partLo_57 = vec3<f32>(-21.15, 163.85, -262.0);
  let gsl_partHi_57 = vec3<f32>(29.0, 206.15, -198.0);
  let gsl_partBox_57 = gsl_boxDistance(p, gsl_partLo_57, gsl_partHi_57);
  var gsl_part_57 = gsl_partBox_57;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_57 <= gsl_vanishView.w) {
    let gsl_partVanish_57 = gsl_vanish(gsl_partLo_57, gsl_partHi_57);
    gsl_part_57 = 1e10;
    if (gsl_partVanish_57.x > 0.0) { gsl_part_57 = gsl_remap(p - vec3<f32>(0.0, 185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_57.y; }
    if (abs(gsl_part_57) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_57), gsl_partVanish_57.zw); }
  }
  let zCaseB = gsl_part_57;
  let gsl_partLo_58 = vec3<f32>(min(gx, gx) - 18.0, min(headY, headY) - 18.0, 238.0);
  let gsl_partHi_58 = vec3<f32>(max(gx, gx) + 18.0, max(headY, headY) + 18.0, 294.0);
  let gsl_partBox_58 = gsl_boxDistance(p, gsl_partLo_58, gsl_partHi_58);
  var gsl_part_58 = gsl_partBox_58;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_58 <= gsl_vanishView.w) {
    let gsl_partVanish_58 = gsl_vanish(gsl_partLo_58, gsl_partHi_58);
    gsl_part_58 = 1e10;
    if (gsl_partVanish_58.x > 0.0) { gsl_part_58 = builtin_cylinderCapped(p, vec3<f32>(gx, headY, 256.0), vec3<f32>(gx, headY, 276.0), 18.0, noiseTex) + gsl_partVanish_58.y; }
    if (abs(gsl_part_58) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_58), gsl_partVanish_58.zw); }
  }
  let extruderMotor = gsl_part_58;
  let gsl_partLo_59 = vec3<f32>(min(gx, gx) - 7.0, min(gy - 74.0, gy - 71.0) - 7.0, 211.0);
  let gsl_partHi_59 = vec3<f32>(max(gx, gx) + 7.0, max(gy - 74.0, gy - 71.0) + 7.0, 225.0);
  let gsl_partBox_59 = gsl_boxDistance(p, gsl_partLo_59, gsl_partHi_59);
  var gsl_part_59 = gsl_partBox_59;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_59 <= gsl_vanishView.w) {
    let gsl_partVanish_59 = gsl_vanish(gsl_partLo_59, gsl_partHi_59);
    gsl_part_59 = 1e10;
    if (gsl_partVanish_59.x > 0.0) { gsl_part_59 = builtin_cylinderCapped(p, vec3<f32>(gx, gy - 74.0, 218.0), vec3<f32>(gx, gy - 71.0, 218.0), 7.0, noiseTex) + gsl_partVanish_59.y; }
    if (abs(gsl_part_59) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_59), gsl_partVanish_59.zw); }
  }
  let fanHub = gsl_part_59;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_52 <= gsl_vanishView.w) {
    let gsl_partVanish_52 = gsl_vanish(gsl_partLo_52, gsl_partHi_52);
    gsl_part_52 = 1e10;
    if (gsl_partVanish_52.x > 0.0) { gsl_part_52 = min(min(min(min(min(min(motorCaseA, motorCaseB), zCaseFL), zCaseFR), zCaseB), extruderMotor), fanHub) + gsl_partVanish_52.y; }
    if (abs(gsl_part_52) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_52), gsl_partVanish_52.zw); }
  }
  let motorCases = gsl_part_52;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_43 <= gsl_vanishView.w) {
    let gsl_partVanish_43 = gsl_vanish(gsl_partLo_43, gsl_partHi_43);
    gsl_part_43 = 1e10;
    if (gsl_partVanish_43.x > 0.0) { gsl_part_43 = min(plates, motorCases) + gsl_partVanish_43.y; }
    if (abs(gsl_part_43) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_43), gsl_partVanish_43.zw); }
  }
  let blackMetal = gsl_part_43;
  let gsl_partLo_60 = vec3<f32>(min(min(min(min(min(min(-236.0, -(-0.0 - gx) - 30.0), min(gx - 30.0, gx + 30.0)), min(gx - 38.0, gx - 24.0)), min(gx + 24.0, gx + 38.0)), 10.0), min(min(min(min(min(min(min(min(-208.75, min(-192.0, gx - 30.0)), min(gx + 30.0, 192.0)), min(-192.0, gx - 30.0)), min(gx + 30.0, 192.0)), -235.0), 205.0), -235.0), 205.0)), min(min(min(min(min(min(min(min(-226.0, min(min(gy - 26.0, gy + 26.0), -(-0.0 - gy) - 20.0 - 4.0)), min(min(gy - 26.0, gy + 26.0), -(-0.0 - gy) - 20.0 - 4.0)), -(-0.0 - headY) - 26.0), min(gy - 20.0, gy - 4.0)), min(gy - 64.0, gy - 28.0)), min(gy - 64.0, gy - 28.0)), -150.0), min(min(min(min(min(min(min(min(min(min(min(min(min(min(-200.0, min(gy, 200.0)), min(-200.0, gy)), -200.0), -200.0), min(gy, 200.0)), min(-200.0, gy)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), -235.0), -235.0), 205.0), 205.0)), -330.0);
  let gsl_partHi_60 = vec3<f32>(max(max(max(max(max(max(236.0, -(-0.0 - gx) + 30.0), max(gx - 30.0, gx + 30.0)), max(gx - 38.0, gx - 24.0)), max(gx + 24.0, gx + 38.0)), 124.0), max(max(max(max(max(max(max(max(208.75, max(-192.0, gx - 30.0)), max(gx + 30.0, 192.0)), max(-192.0, gx - 30.0)), max(gx + 30.0, 192.0)), -205.0), 235.0), -205.0), 235.0)), max(max(max(max(max(max(max(max(226.0, max(max(gy - 26.0, gy + 26.0), -(-0.0 - gy) + 20.0 + 4.0)), max(max(gy - 26.0, gy + 26.0), -(-0.0 - gy) + 20.0 + 4.0)), -(-0.0 - headY) + 26.0), max(gy - 20.0, gy - 4.0)), max(gy - 64.0, gy - 28.0)), max(gy - 64.0, gy - 28.0)), 150.0), max(max(max(max(max(max(max(max(max(max(max(max(max(max(200.0, max(gy, 200.0)), max(-200.0, gy)), 200.0), 200.0), max(gy, 200.0)), max(-200.0, gy)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), -205.0), -205.0), 235.0), 235.0)), 258.0);
  let gsl_partBox_60 = gsl_boxDistance(p, gsl_partLo_60, gsl_partHi_60);
  var gsl_part_60 = gsl_partBox_60;
  let gsl_partLo_61 = vec3<f32>(min(min(min(min(-236.0, -(-0.0 - gx) - 30.0), min(gx - 30.0, gx + 30.0)), min(gx - 38.0, gx - 24.0)), min(gx + 24.0, gx + 38.0)), min(min(min(min(min(min(-226.0, min(min(gy - 26.0, gy + 26.0), -(-0.0 - gy) - 20.0 - 4.0)), min(min(gy - 26.0, gy + 26.0), -(-0.0 - gy) - 20.0 - 4.0)), -(-0.0 - headY) - 26.0), min(gy - 20.0, gy - 4.0)), min(gy - 64.0, gy - 28.0)), min(gy - 64.0, gy - 28.0)), 128.0);
  let gsl_partHi_61 = vec3<f32>(max(max(max(max(236.0, -(-0.0 - gx) + 30.0), max(gx - 30.0, gx + 30.0)), max(gx - 38.0, gx - 24.0)), max(gx + 24.0, gx + 38.0)), max(max(max(max(max(max(226.0, max(max(gy - 26.0, gy + 26.0), -(-0.0 - gy) + 20.0 + 4.0)), max(max(gy - 26.0, gy + 26.0), -(-0.0 - gy) + 20.0 + 4.0)), -(-0.0 - headY) + 26.0), max(gy - 20.0, gy - 4.0)), max(gy - 64.0, gy - 28.0)), max(gy - 64.0, gy - 28.0)), 258.0);
  let gsl_partBox_61 = gsl_boxDistance(p, gsl_partLo_61, gsl_partHi_61);
  var gsl_part_61 = gsl_partBox_61;
  let gsl_partLo_62 = vec3<f32>(-226.0, 174.0, 200.0);
  let gsl_partHi_62 = vec3<f32>(-174.0, 226.0, 256.0);
  let gsl_partBox_62 = gsl_boxDistance(p, gsl_partLo_62, gsl_partHi_62);
  var gsl_part_62 = gsl_partBox_62;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_62 <= gsl_vanishView.w) {
    let gsl_partVanish_62 = gsl_vanish(gsl_partLo_62, gsl_partHi_62);
    gsl_part_62 = 1e10;
    if (gsl_partVanish_62.x > 0.0) { gsl_part_62 = driveBase(p, -200.0, 200.0, noiseTex) + gsl_partVanish_62.y; }
    if (abs(gsl_part_62) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_62), gsl_partVanish_62.zw); }
  }
  let driveBaseA = gsl_part_62;
  let gsl_partLo_63 = vec3<f32>(174.0, 174.0, 200.0);
  let gsl_partHi_63 = vec3<f32>(226.0, 226.0, 256.0);
  let gsl_partBox_63 = gsl_boxDistance(p, gsl_partLo_63, gsl_partHi_63);
  var gsl_part_63 = gsl_partBox_63;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_63 <= gsl_vanishView.w) {
    let gsl_partVanish_63 = gsl_vanish(gsl_partLo_63, gsl_partHi_63);
    gsl_part_63 = 1e10;
    if (gsl_partVanish_63.x > 0.0) { gsl_part_63 = driveBase(p, 200.0, 200.0, noiseTex) + gsl_partVanish_63.y; }
    if (abs(gsl_part_63) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_63), gsl_partVanish_63.zw); }
  }
  let driveBaseB = gsl_part_63;
  let gsl_partLo_64 = vec3<f32>(-226.0, -226.0, 200.0);
  let gsl_partHi_64 = vec3<f32>(-174.0, -174.0, 256.0);
  let gsl_partBox_64 = gsl_boxDistance(p, gsl_partLo_64, gsl_partHi_64);
  var gsl_part_64 = gsl_partBox_64;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_64 <= gsl_vanishView.w) {
    let gsl_partVanish_64 = gsl_vanish(gsl_partLo_64, gsl_partHi_64);
    gsl_part_64 = 1e10;
    if (gsl_partVanish_64.x > 0.0) { gsl_part_64 = driveBase(p, -200.0, -200.0, noiseTex) + gsl_partVanish_64.y; }
    if (abs(gsl_part_64) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_64), gsl_partVanish_64.zw); }
  }
  let idlerBaseL = gsl_part_64;
  let gsl_partLo_65 = vec3<f32>(174.0, -226.0, 200.0);
  let gsl_partHi_65 = vec3<f32>(226.0, -174.0, 256.0);
  let gsl_partBox_65 = gsl_boxDistance(p, gsl_partLo_65, gsl_partHi_65);
  var gsl_part_65 = gsl_partBox_65;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_65 <= gsl_vanishView.w) {
    let gsl_partVanish_65 = gsl_vanish(gsl_partLo_65, gsl_partHi_65);
    gsl_part_65 = 1e10;
    if (gsl_partVanish_65.x > 0.0) { gsl_part_65 = driveBase(p, 200.0, -200.0, noiseTex) + gsl_partVanish_65.y; }
    if (abs(gsl_part_65) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_65), gsl_partVanish_65.zw); }
  }
  let idlerBaseR = gsl_part_65;
  let gsl_partLo_66 = vec3<f32>(-236.0, min(min(gy - 26.0, gy + 26.0), -(-0.0 - gy) - 20.0 - 4.0), 213.0);
  let gsl_partHi_66 = vec3<f32>(-190.0, max(max(gy - 26.0, gy + 26.0), -(-0.0 - gy) + 20.0 + 4.0), 256.0);
  let gsl_partBox_66 = gsl_boxDistance(p, gsl_partLo_66, gsl_partHi_66);
  var gsl_part_66 = gsl_partBox_66;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_66 <= gsl_vanishView.w) {
    let gsl_partVanish_66 = gsl_vanish(gsl_partLo_66, gsl_partHi_66);
    gsl_part_66 = 1e10;
    if (gsl_partVanish_66.x > 0.0) { gsl_part_66 = jointBase(p, gy, -1.0, noiseTex) + gsl_partVanish_66.y; }
    if (abs(gsl_part_66) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_66), gsl_partVanish_66.zw); }
  }
  let jointBaseL = gsl_part_66;
  let gsl_partLo_67 = vec3<f32>(190.0, min(min(gy - 26.0, gy + 26.0), -(-0.0 - gy) - 20.0 - 4.0), 213.0);
  let gsl_partHi_67 = vec3<f32>(236.0, max(max(gy - 26.0, gy + 26.0), -(-0.0 - gy) + 20.0 + 4.0), 256.0);
  let gsl_partBox_67 = gsl_boxDistance(p, gsl_partLo_67, gsl_partHi_67);
  var gsl_part_67 = gsl_partBox_67;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_67 <= gsl_vanishView.w) {
    let gsl_partVanish_67 = gsl_vanish(gsl_partLo_67, gsl_partHi_67);
    gsl_part_67 = 1e10;
    if (gsl_partVanish_67.x > 0.0) { gsl_part_67 = jointBase(p, gy, 1.0, noiseTex) + gsl_partVanish_67.y; }
    if (abs(gsl_part_67) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_67), gsl_partVanish_67.zw); }
  }
  let jointBaseR = gsl_part_67;
  let gsl_partLo_68 = vec3<f32>(-(-0.0 - gx) - 30.0, -(-0.0 - headY) - 26.0, 136.0);
  let gsl_partHi_68 = vec3<f32>(-(-0.0 - gx) + 30.0, -(-0.0 - headY) + 26.0, 256.0);
  let gsl_partBox_68 = gsl_boxDistance(p, gsl_partLo_68, gsl_partHi_68);
  var gsl_part_68 = gsl_partBox_68;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_68 <= gsl_vanishView.w) {
    let gsl_partVanish_68 = gsl_vanish(gsl_partLo_68, gsl_partHi_68);
    gsl_part_68 = 1e10;
    if (gsl_partVanish_68.x > 0.0) { gsl_part_68 = max(max(max(abs(p.x - gx) - 30.0, abs(p.y - headY) - 26.0), abs(p.z - 196.0) - 60.0), abs(p.x - gx) + abs(p.y - headY) - 50.0) + gsl_partVanish_68.y; }
    if (abs(gsl_part_68) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_68), gsl_partVanish_68.zw); }
  }
  let headBody = gsl_part_68;
  let gsl_partLo_69 = vec3<f32>(min(gx - 30.0, gx + 30.0), min(gy - 20.0, gy - 4.0), 236.0);
  let gsl_partHi_69 = vec3<f32>(max(gx - 30.0, gx + 30.0), max(gy - 20.0, gy - 4.0), 258.0);
  let gsl_partBox_69 = gsl_boxDistance(p, gsl_partLo_69, gsl_partHi_69);
  var gsl_part_69 = gsl_partBox_69;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_69 <= gsl_vanishView.w) {
    let gsl_partVanish_69 = gsl_vanish(gsl_partLo_69, gsl_partHi_69);
    gsl_part_69 = 1e10;
    if (gsl_partVanish_69.x > 0.0) { gsl_part_69 = builtin_box(p, vec3<f32>(gx - 30.0, gy - 20.0, 236.0), vec3<f32>(gx + 30.0, gy - 4.0, 258.0), noiseTex) + gsl_partVanish_69.y; }
    if (abs(gsl_part_69) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_69), gsl_partVanish_69.zw); }
  }
  let headClamp = gsl_part_69;
  let gsl_partLo_70 = vec3<f32>(min(gx - 38.0, gx - 24.0), min(gy - 64.0, gy - 28.0), 128.0);
  let gsl_partHi_70 = vec3<f32>(max(gx - 38.0, gx - 24.0), max(gy - 64.0, gy - 28.0), 160.0);
  let gsl_partBox_70 = gsl_boxDistance(p, gsl_partLo_70, gsl_partHi_70);
  var gsl_part_70 = gsl_partBox_70;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_70 <= gsl_vanishView.w) {
    let gsl_partVanish_70 = gsl_vanish(gsl_partLo_70, gsl_partHi_70);
    gsl_part_70 = 1e10;
    if (gsl_partVanish_70.x > 0.0) { gsl_part_70 = builtin_box(p, vec3<f32>(gx - 38.0, gy - 64.0, 128.0), vec3<f32>(gx - 24.0, gy - 28.0, 160.0), noiseTex) + gsl_partVanish_70.y; }
    if (abs(gsl_part_70) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_70), gsl_partVanish_70.zw); }
  }
  let ductL = gsl_part_70;
  let gsl_partLo_71 = vec3<f32>(min(gx + 24.0, gx + 38.0), min(gy - 64.0, gy - 28.0), 128.0);
  let gsl_partHi_71 = vec3<f32>(max(gx + 24.0, gx + 38.0), max(gy - 64.0, gy - 28.0), 160.0);
  let gsl_partBox_71 = gsl_boxDistance(p, gsl_partLo_71, gsl_partHi_71);
  var gsl_part_71 = gsl_partBox_71;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_71 <= gsl_vanishView.w) {
    let gsl_partVanish_71 = gsl_vanish(gsl_partLo_71, gsl_partHi_71);
    gsl_part_71 = 1e10;
    if (gsl_partVanish_71.x > 0.0) { gsl_part_71 = builtin_box(p, vec3<f32>(gx + 24.0, gy - 64.0, 128.0), vec3<f32>(gx + 38.0, gy - 28.0, 160.0), noiseTex) + gsl_partVanish_71.y; }
    if (abs(gsl_part_71) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_71), gsl_partVanish_71.zw); }
  }
  let ductR = gsl_part_71;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_61 <= gsl_vanishView.w) {
    let gsl_partVanish_61 = gsl_vanish(gsl_partLo_61, gsl_partHi_61);
    gsl_part_61 = 1e10;
    if (gsl_partVanish_61.x > 0.0) { gsl_part_61 = min(min(min(min(min(min(min(min(min(driveBaseA, driveBaseB), idlerBaseL), idlerBaseR), jointBaseL), jointBaseR), headBody), headClamp), ductL), ductR) + gsl_partVanish_61.y; }
    if (abs(gsl_part_61) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_61), gsl_partVanish_61.zw); }
  }
  let printedBlack = gsl_part_61;
  let gsl_partLo_72 = vec3<f32>(10.0, -150.0, -277.0);
  let gsl_partHi_72 = vec3<f32>(124.0, 150.0, -255.4);
  let gsl_partBox_72 = gsl_boxDistance(p, gsl_partLo_72, gsl_partHi_72);
  var gsl_part_72 = gsl_partBox_72;
  let gsl_partLo_73 = vec3<f32>(110.0, -120.0, -265.4);
  let gsl_partHi_73 = vec3<f32>(124.0, -106.0, -263.9);
  let gsl_partBox_73 = gsl_boxDistance(p, gsl_partLo_73, gsl_partHi_73);
  var gsl_part_73 = gsl_partBox_73;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_73 <= gsl_vanishView.w) {
    let gsl_partVanish_73 = gsl_vanish(gsl_partLo_73, gsl_partHi_73);
    gsl_part_73 = 1e10;
    if (gsl_partVanish_73.x > 0.0) { gsl_part_73 = builtin_gcl_closure_boxCentered(p, vec3<f32>(234.0, -226.0, -529.3) * 0.5, abs(vec3<f32>(14.0, 14.0, 1.5)), noiseTex) + gsl_partVanish_73.y; }
    if (abs(gsl_part_73) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_73), gsl_partVanish_73.zw); }
  }
  let processor = gsl_part_73;
  let gsl_partLo_74 = vec3<f32>(10.0, -150.0, -265.4);
  let gsl_partHi_74 = vec3<f32>(100.0, -130.0, -255.4);
  let gsl_partBox_74 = gsl_boxDistance(p, gsl_partLo_74, gsl_partHi_74);
  var gsl_part_74 = gsl_partBox_74;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_74 <= gsl_vanishView.w) {
    let gsl_partVanish_74 = gsl_vanish(gsl_partLo_74, gsl_partHi_74);
    gsl_part_74 = 1e10;
    if (gsl_partVanish_74.x > 0.0) { gsl_part_74 = max(max(max(abs(fract((p.x - 10.0) / 25.0) * 25.0 - 7.5) - 7.5, abs(p.y + 140.0) - 10.0), abs(p.z - -260.4) - 5.0), abs(p.x - 55.0) - 45.0) + gsl_partVanish_74.y; }
    if (abs(gsl_part_74) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_74), gsl_partVanish_74.zw); }
  }
  let drivers = gsl_part_74;
  let gsl_partLo_75 = vec3<f32>(15.0, 55.0, -277.0);
  let gsl_partHi_75 = vec3<f32>(27.0, 150.0, -260.0);
  let gsl_partBox_75 = gsl_boxDistance(p, gsl_partLo_75, gsl_partHi_75);
  var gsl_part_75 = gsl_partBox_75;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_75 <= gsl_vanishView.w) {
    let gsl_partVanish_75 = gsl_vanish(gsl_partLo_75, gsl_partHi_75);
    gsl_part_75 = 1e10;
    if (gsl_partVanish_75.x > 0.0) { gsl_part_75 = builtin_gcl_closure_boxCentered(p, vec3<f32>(42.0, 205.0, -537.0) * 0.5, abs(vec3<f32>(12.0, 95.0, 17.0)), noiseTex) + gsl_partVanish_75.y; }
    if (abs(gsl_part_75) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_75), gsl_partVanish_75.zw); }
  }
  let psuTerminals = gsl_part_75;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_72 <= gsl_vanishView.w) {
    let gsl_partVanish_72 = gsl_vanish(gsl_partLo_72, gsl_partHi_72);
    gsl_part_72 = 1e10;
    if (gsl_partVanish_72.x > 0.0) { gsl_part_72 = min(min(processor, drivers), psuTerminals) + gsl_partVanish_72.y; }
    if (abs(gsl_part_72) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_72), gsl_partVanish_72.zw); }
  }
  let chips = gsl_part_72;
  let gsl_partLo_76 = vec3<f32>(min(min(min(min(min(min(min(min(-208.75, min(-192.0, gx - 30.0)), min(gx + 30.0, 192.0)), min(-192.0, gx - 30.0)), min(gx + 30.0, 192.0)), -235.0), 205.0), -235.0), 205.0), min(min(min(min(min(min(min(min(min(min(min(min(min(min(-200.0, min(gy, 200.0)), min(-200.0, gy)), -200.0), -200.0), min(gy, 200.0)), min(-200.0, gy)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), -235.0), -235.0), 205.0), 205.0), -330.0);
  let gsl_partHi_76 = vec3<f32>(max(max(max(max(max(max(max(max(208.75, max(-192.0, gx - 30.0)), max(gx + 30.0, 192.0)), max(-192.0, gx - 30.0)), max(gx + 30.0, 192.0)), -205.0), 235.0), -205.0), 235.0), max(max(max(max(max(max(max(max(max(max(max(max(max(max(200.0, max(gy, 200.0)), max(-200.0, gy)), 200.0), 200.0), max(gy, 200.0)), max(-200.0, gy)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75)), -205.0), -205.0), 235.0), 235.0), 254.0);
  let gsl_partBox_76 = gsl_boxDistance(p, gsl_partLo_76, gsl_partHi_76);
  var gsl_part_76 = gsl_partBox_76;
  let gsl_partLo_77 = vec3<f32>(-208.75, -200.0, 240.0);
  let gsl_partHi_77 = vec3<f32>(-207.25, 200.0, 246.0);
  let gsl_partBox_77 = gsl_boxDistance(p, gsl_partLo_77, gsl_partHi_77);
  var gsl_part_77 = gsl_partBox_77;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_77 <= gsl_vanishView.w) {
    let gsl_partVanish_77 = gsl_vanish(gsl_partLo_77, gsl_partHi_77);
    gsl_part_77 = 1e10;
    if (gsl_partVanish_77.x > 0.0) { gsl_part_77 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-416.0, 0.0, 486.0) * 0.5, abs(vec3<f32>(1.5, 400.0, 6.0)), noiseTex) + gsl_partVanish_77.y; }
    if (abs(gsl_part_77) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_77), gsl_partVanish_77.zw); }
  }
  let beltLAo = gsl_part_77;
  let gsl_partLo_78 = vec3<f32>(-208.75, -200.0, 248.0);
  let gsl_partHi_78 = vec3<f32>(-207.25, 200.0, 254.0);
  let gsl_partBox_78 = gsl_boxDistance(p, gsl_partLo_78, gsl_partHi_78);
  var gsl_part_78 = gsl_partBox_78;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_78 <= gsl_vanishView.w) {
    let gsl_partVanish_78 = gsl_vanish(gsl_partLo_78, gsl_partHi_78);
    gsl_part_78 = 1e10;
    if (gsl_partVanish_78.x > 0.0) { gsl_part_78 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-416.0, 0.0, 502.0) * 0.5, abs(vec3<f32>(1.5, 400.0, 6.0)), noiseTex) + gsl_partVanish_78.y; }
    if (abs(gsl_part_78) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_78), gsl_partVanish_78.zw); }
  }
  let beltLBo = gsl_part_78;
  let gsl_partLo_79 = vec3<f32>(-192.75, min(gy, 200.0), 240.0);
  let gsl_partHi_79 = vec3<f32>(-191.25, max(gy, 200.0), 246.0);
  let gsl_partBox_79 = gsl_boxDistance(p, gsl_partLo_79, gsl_partHi_79);
  var gsl_part_79 = gsl_partBox_79;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_79 <= gsl_vanishView.w) {
    let gsl_partVanish_79 = gsl_vanish(gsl_partLo_79, gsl_partHi_79);
    gsl_part_79 = 1e10;
    if (gsl_partVanish_79.x > 0.0) { gsl_part_79 = beltY(p, -192.0, 243.0, gy, 200.0, noiseTex) + gsl_partVanish_79.y; }
    if (abs(gsl_part_79) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_79), gsl_partVanish_79.zw); }
  }
  let beltLAi = gsl_part_79;
  let gsl_partLo_80 = vec3<f32>(-192.75, min(-200.0, gy), 248.0);
  let gsl_partHi_80 = vec3<f32>(-191.25, max(-200.0, gy), 254.0);
  let gsl_partBox_80 = gsl_boxDistance(p, gsl_partLo_80, gsl_partHi_80);
  var gsl_part_80 = gsl_partBox_80;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_80 <= gsl_vanishView.w) {
    let gsl_partVanish_80 = gsl_vanish(gsl_partLo_80, gsl_partHi_80);
    gsl_part_80 = 1e10;
    if (gsl_partVanish_80.x > 0.0) { gsl_part_80 = beltY(p, -192.0, 251.0, -200.0, gy, noiseTex) + gsl_partVanish_80.y; }
    if (abs(gsl_part_80) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_80), gsl_partVanish_80.zw); }
  }
  let beltLBi = gsl_part_80;
  let gsl_partLo_81 = vec3<f32>(207.25, -200.0, 240.0);
  let gsl_partHi_81 = vec3<f32>(208.75, 200.0, 246.0);
  let gsl_partBox_81 = gsl_boxDistance(p, gsl_partLo_81, gsl_partHi_81);
  var gsl_part_81 = gsl_partBox_81;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_81 <= gsl_vanishView.w) {
    let gsl_partVanish_81 = gsl_vanish(gsl_partLo_81, gsl_partHi_81);
    gsl_part_81 = 1e10;
    if (gsl_partVanish_81.x > 0.0) { gsl_part_81 = builtin_gcl_closure_boxCentered(p, vec3<f32>(416.0, 0.0, 486.0) * 0.5, abs(vec3<f32>(1.5, 400.0, 6.0)), noiseTex) + gsl_partVanish_81.y; }
    if (abs(gsl_part_81) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_81), gsl_partVanish_81.zw); }
  }
  let beltRAo = gsl_part_81;
  let gsl_partLo_82 = vec3<f32>(207.25, -200.0, 248.0);
  let gsl_partHi_82 = vec3<f32>(208.75, 200.0, 254.0);
  let gsl_partBox_82 = gsl_boxDistance(p, gsl_partLo_82, gsl_partHi_82);
  var gsl_part_82 = gsl_partBox_82;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_82 <= gsl_vanishView.w) {
    let gsl_partVanish_82 = gsl_vanish(gsl_partLo_82, gsl_partHi_82);
    gsl_part_82 = 1e10;
    if (gsl_partVanish_82.x > 0.0) { gsl_part_82 = builtin_gcl_closure_boxCentered(p, vec3<f32>(416.0, 0.0, 502.0) * 0.5, abs(vec3<f32>(1.5, 400.0, 6.0)), noiseTex) + gsl_partVanish_82.y; }
    if (abs(gsl_part_82) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_82), gsl_partVanish_82.zw); }
  }
  let beltRBo = gsl_part_82;
  let gsl_partLo_83 = vec3<f32>(191.25, min(gy, 200.0), 248.0);
  let gsl_partHi_83 = vec3<f32>(192.75, max(gy, 200.0), 254.0);
  let gsl_partBox_83 = gsl_boxDistance(p, gsl_partLo_83, gsl_partHi_83);
  var gsl_part_83 = gsl_partBox_83;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_83 <= gsl_vanishView.w) {
    let gsl_partVanish_83 = gsl_vanish(gsl_partLo_83, gsl_partHi_83);
    gsl_part_83 = 1e10;
    if (gsl_partVanish_83.x > 0.0) { gsl_part_83 = beltY(p, 192.0, 251.0, gy, 200.0, noiseTex) + gsl_partVanish_83.y; }
    if (abs(gsl_part_83) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_83), gsl_partVanish_83.zw); }
  }
  let beltRBi = gsl_part_83;
  let gsl_partLo_84 = vec3<f32>(191.25, min(-200.0, gy), 240.0);
  let gsl_partHi_84 = vec3<f32>(192.75, max(-200.0, gy), 246.0);
  let gsl_partBox_84 = gsl_boxDistance(p, gsl_partLo_84, gsl_partHi_84);
  var gsl_part_84 = gsl_partBox_84;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_84 <= gsl_vanishView.w) {
    let gsl_partVanish_84 = gsl_vanish(gsl_partLo_84, gsl_partHi_84);
    gsl_part_84 = 1e10;
    if (gsl_partVanish_84.x > 0.0) { gsl_part_84 = beltY(p, 192.0, 243.0, -200.0, gy, noiseTex) + gsl_partVanish_84.y; }
    if (abs(gsl_part_84) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_84), gsl_partVanish_84.zw); }
  }
  let beltRAi = gsl_part_84;
  let gsl_partLo_85 = vec3<f32>(min(-192.0, gx - 30.0), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 240.0);
  let gsl_partHi_85 = vec3<f32>(max(-192.0, gx - 30.0), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 246.0);
  let gsl_partBox_85 = gsl_boxDistance(p, gsl_partLo_85, gsl_partHi_85);
  var gsl_part_85 = gsl_partBox_85;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_85 <= gsl_vanishView.w) {
    let gsl_partVanish_85 = gsl_vanish(gsl_partLo_85, gsl_partHi_85);
    gsl_part_85 = 1e10;
    if (gsl_partVanish_85.x > 0.0) { gsl_part_85 = beltX(p, gy - 8.0, 243.0, -192.0, gx - 30.0, noiseTex) + gsl_partVanish_85.y; }
    if (abs(gsl_part_85) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_85), gsl_partVanish_85.zw); }
  }
  let beltXAL = gsl_part_85;
  let gsl_partLo_86 = vec3<f32>(min(gx + 30.0, 192.0), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 240.0);
  let gsl_partHi_86 = vec3<f32>(max(gx + 30.0, 192.0), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 246.0);
  let gsl_partBox_86 = gsl_boxDistance(p, gsl_partLo_86, gsl_partHi_86);
  var gsl_part_86 = gsl_partBox_86;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_86 <= gsl_vanishView.w) {
    let gsl_partVanish_86 = gsl_vanish(gsl_partLo_86, gsl_partHi_86);
    gsl_part_86 = 1e10;
    if (gsl_partVanish_86.x > 0.0) { gsl_part_86 = beltX(p, gy - 8.0, 243.0, gx + 30.0, 192.0, noiseTex) + gsl_partVanish_86.y; }
    if (abs(gsl_part_86) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_86), gsl_partVanish_86.zw); }
  }
  let beltXAR = gsl_part_86;
  let gsl_partLo_87 = vec3<f32>(min(-192.0, gx - 30.0), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 248.0);
  let gsl_partHi_87 = vec3<f32>(max(-192.0, gx - 30.0), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 254.0);
  let gsl_partBox_87 = gsl_boxDistance(p, gsl_partLo_87, gsl_partHi_87);
  var gsl_part_87 = gsl_partBox_87;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_87 <= gsl_vanishView.w) {
    let gsl_partVanish_87 = gsl_vanish(gsl_partLo_87, gsl_partHi_87);
    gsl_part_87 = 1e10;
    if (gsl_partVanish_87.x > 0.0) { gsl_part_87 = beltX(p, gy - 8.0, 251.0, -192.0, gx - 30.0, noiseTex) + gsl_partVanish_87.y; }
    if (abs(gsl_part_87) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_87), gsl_partVanish_87.zw); }
  }
  let beltXBL = gsl_part_87;
  let gsl_partLo_88 = vec3<f32>(min(gx + 30.0, 192.0), min(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 248.0);
  let gsl_partHi_88 = vec3<f32>(max(gx + 30.0, 192.0), max(gy - 8.0 - 0.75, gy - 8.0 + 0.75), 254.0);
  let gsl_partBox_88 = gsl_boxDistance(p, gsl_partLo_88, gsl_partHi_88);
  var gsl_part_88 = gsl_partBox_88;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_88 <= gsl_vanishView.w) {
    let gsl_partVanish_88 = gsl_vanish(gsl_partLo_88, gsl_partHi_88);
    gsl_part_88 = 1e10;
    if (gsl_partVanish_88.x > 0.0) { gsl_part_88 = beltX(p, gy - 8.0, 251.0, gx + 30.0, 192.0, noiseTex) + gsl_partVanish_88.y; }
    if (abs(gsl_part_88) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_88), gsl_partVanish_88.zw); }
  }
  let beltXBR = gsl_part_88;
  let gsl_partLo_89 = vec3<f32>(-235.0, -235.0, -330.0);
  let gsl_partHi_89 = vec3<f32>(-205.0, -205.0, -285.0);
  let gsl_partBox_89 = gsl_boxDistance(p, gsl_partLo_89, gsl_partHi_89);
  var gsl_part_89 = gsl_partBox_89;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_89 <= gsl_vanishView.w) {
    let gsl_partVanish_89 = gsl_vanish(gsl_partLo_89, gsl_partHi_89);
    gsl_part_89 = 1e10;
    if (gsl_partVanish_89.x > 0.0) { gsl_part_89 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-220.0, -220.0, -315.0), vec3<f32>(-220.0, -220.0, -300.0), 15.0, noiseTex) + gsl_partVanish_89.y; }
    if (abs(gsl_part_89) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_89), gsl_partVanish_89.zw); }
  }
  let footFL = gsl_part_89;
  let gsl_partLo_90 = vec3<f32>(205.0, -235.0, -330.0);
  let gsl_partHi_90 = vec3<f32>(235.0, -205.0, -285.0);
  let gsl_partBox_90 = gsl_boxDistance(p, gsl_partLo_90, gsl_partHi_90);
  var gsl_part_90 = gsl_partBox_90;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_90 <= gsl_vanishView.w) {
    let gsl_partVanish_90 = gsl_vanish(gsl_partLo_90, gsl_partHi_90);
    gsl_part_90 = 1e10;
    if (gsl_partVanish_90.x > 0.0) { gsl_part_90 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(220.0, -220.0, -315.0), vec3<f32>(220.0, -220.0, -300.0), 15.0, noiseTex) + gsl_partVanish_90.y; }
    if (abs(gsl_part_90) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_90), gsl_partVanish_90.zw); }
  }
  let footFR = gsl_part_90;
  let gsl_partLo_91 = vec3<f32>(-235.0, 205.0, -330.0);
  let gsl_partHi_91 = vec3<f32>(-205.0, 235.0, -285.0);
  let gsl_partBox_91 = gsl_boxDistance(p, gsl_partLo_91, gsl_partHi_91);
  var gsl_part_91 = gsl_partBox_91;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_91 <= gsl_vanishView.w) {
    let gsl_partVanish_91 = gsl_vanish(gsl_partLo_91, gsl_partHi_91);
    gsl_part_91 = 1e10;
    if (gsl_partVanish_91.x > 0.0) { gsl_part_91 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(-220.0, 220.0, -315.0), vec3<f32>(-220.0, 220.0, -300.0), 15.0, noiseTex) + gsl_partVanish_91.y; }
    if (abs(gsl_part_91) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_91), gsl_partVanish_91.zw); }
  }
  let footBL = gsl_part_91;
  let gsl_partLo_92 = vec3<f32>(205.0, 205.0, -330.0);
  let gsl_partHi_92 = vec3<f32>(235.0, 235.0, -285.0);
  let gsl_partBox_92 = gsl_boxDistance(p, gsl_partLo_92, gsl_partHi_92);
  var gsl_part_92 = gsl_partBox_92;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_92 <= gsl_vanishView.w) {
    let gsl_partVanish_92 = gsl_vanish(gsl_partLo_92, gsl_partHi_92);
    gsl_part_92 = 1e10;
    if (gsl_partVanish_92.x > 0.0) { gsl_part_92 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(220.0, 220.0, -315.0), vec3<f32>(220.0, 220.0, -300.0), 15.0, noiseTex) + gsl_partVanish_92.y; }
    if (abs(gsl_part_92) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_92), gsl_partVanish_92.zw); }
  }
  let footBR = gsl_part_92;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_76 <= gsl_vanishView.w) {
    let gsl_partVanish_76 = gsl_vanish(gsl_partLo_76, gsl_partHi_76);
    gsl_part_76 = 1e10;
    if (gsl_partVanish_76.x > 0.0) { gsl_part_76 = min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(beltLAo, beltLBo), beltLAi), beltLBi), beltRAo), beltRBo), beltRBi), beltRAi), beltXAL), beltXAR), beltXBL), beltXBR), footFL), footFR), footBL), footBR) + gsl_partVanish_76.y; }
    if (abs(gsl_part_76) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_76), gsl_partVanish_76.zw); }
  }
  let rubber = gsl_part_76;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_60 <= gsl_vanishView.w) {
    let gsl_partVanish_60 = gsl_vanish(gsl_partLo_60, gsl_partHi_60);
    gsl_part_60 = 1e10;
    if (gsl_partVanish_60.x > 0.0) { gsl_part_60 = min(min(printedBlack, chips), rubber) + gsl_partVanish_60.y; }
    if (abs(gsl_part_60) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_60), gsl_partVanish_60.zw); }
  }
  let darkParts = gsl_part_60;
  let gsl_partLo_93 = vec3<f32>(min(min(min(min(min(min(-236.0, min(min(gx - 26.0, gx + 26.0), -(-0.0 - gx) - 16.0)), -202.0), 168.0), -17.0), -52.0), -135.0), min(min(min(min(min(min(min(min(-226.0, min(gy - 26.0, gy + 26.0)), min(gy - 26.0, gy + 26.0)), min(min(gy - 75.0, gy - 72.0), -(-0.0 - (gy - 74.0)) - 1.0)), -210.0), -210.0), 168.0), -250.0), -135.0), min(-278.0, min(bz - 10.5, bz - 9.0)));
  let gsl_partHi_93 = vec3<f32>(max(max(max(max(max(max(236.0, max(max(gx - 26.0, gx + 26.0), -(-0.0 - gx) + 16.0)), -168.0), 202.0), 17.0), 52.0), 135.0), max(max(max(max(max(max(max(max(230.0, max(gy - 26.0, gy + 26.0)), max(gy - 26.0, gy + 26.0)), max(max(gy - 75.0, gy - 72.0), -(-0.0 - (gy - 74.0)) + 1.0)), -168.0), -168.0), 210.0), -230.0), 135.0), max(262.0, max(bz - 10.5, bz - 9.0)));
  let gsl_partBox_93 = gsl_boxDistance(p, gsl_partLo_93, gsl_partHi_93);
  var gsl_part_93 = gsl_partBox_93;
  let gsl_partLo_94 = vec3<f32>(min(min(min(min(min(-236.0, min(min(gx - 26.0, gx + 26.0), -(-0.0 - gx) - 16.0)), -202.0), 168.0), -17.0), -52.0), min(min(min(min(min(min(min(-226.0, min(gy - 26.0, gy + 26.0)), min(gy - 26.0, gy + 26.0)), min(min(gy - 75.0, gy - 72.0), -(-0.0 - (gy - 74.0)) - 1.0)), -210.0), -210.0), 168.0), -250.0), -278.0);
  let gsl_partHi_94 = vec3<f32>(max(max(max(max(max(236.0, max(max(gx - 26.0, gx + 26.0), -(-0.0 - gx) + 16.0)), -168.0), 202.0), 17.0), 52.0), max(max(max(max(max(max(max(230.0, max(gy - 26.0, gy + 26.0)), max(gy - 26.0, gy + 26.0)), max(max(gy - 75.0, gy - 72.0), -(-0.0 - (gy - 74.0)) + 1.0)), -168.0), -168.0), 210.0), -230.0), 262.0);
  let gsl_partBox_94 = gsl_boxDistance(p, gsl_partLo_94, gsl_partHi_94);
  var gsl_part_94 = gsl_partBox_94;
  let gsl_partLo_95 = vec3<f32>(-226.0, 226.0, 200.0);
  let gsl_partHi_95 = vec3<f32>(-174.0, 230.0, 256.0);
  let gsl_partBox_95 = gsl_boxDistance(p, gsl_partLo_95, gsl_partHi_95);
  var gsl_part_95 = gsl_partBox_95;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_95 <= gsl_vanishView.w) {
    let gsl_partVanish_95 = gsl_vanish(gsl_partLo_95, gsl_partHi_95);
    gsl_part_95 = 1e10;
    if (gsl_partVanish_95.x > 0.0) { gsl_part_95 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-400.0, 456.0, 456.0) * 0.5, abs(vec3<f32>(52.0, 4.0, 56.0)), noiseTex) + gsl_partVanish_95.y; }
    if (abs(gsl_part_95) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_95), gsl_partVanish_95.zw); }
  }
  let driveCoverA = gsl_part_95;
  let gsl_partLo_96 = vec3<f32>(174.0, 226.0, 200.0);
  let gsl_partHi_96 = vec3<f32>(226.0, 230.0, 256.0);
  let gsl_partBox_96 = gsl_boxDistance(p, gsl_partLo_96, gsl_partHi_96);
  var gsl_part_96 = gsl_partBox_96;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_96 <= gsl_vanishView.w) {
    let gsl_partVanish_96 = gsl_vanish(gsl_partLo_96, gsl_partHi_96);
    gsl_part_96 = 1e10;
    if (gsl_partVanish_96.x > 0.0) { gsl_part_96 = builtin_gcl_closure_boxCentered(p, vec3<f32>(400.0, 456.0, 456.0) * 0.5, abs(vec3<f32>(52.0, 4.0, 56.0)), noiseTex) + gsl_partVanish_96.y; }
    if (abs(gsl_part_96) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_96), gsl_partVanish_96.zw); }
  }
  let driveCoverB = gsl_part_96;
  let gsl_partLo_97 = vec3<f32>(-226.0, -226.0, 256.0);
  let gsl_partHi_97 = vec3<f32>(-174.0, -174.0, 262.0);
  let gsl_partBox_97 = gsl_boxDistance(p, gsl_partLo_97, gsl_partHi_97);
  var gsl_part_97 = gsl_partBox_97;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_97 <= gsl_vanishView.w) {
    let gsl_partVanish_97 = gsl_vanish(gsl_partLo_97, gsl_partHi_97);
    gsl_part_97 = 1e10;
    if (gsl_partVanish_97.x > 0.0) { gsl_part_97 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-400.0, -400.0, 518.0) * 0.5, abs(vec3<f32>(52.0, 52.0, 6.0)), noiseTex) + gsl_partVanish_97.y; }
    if (abs(gsl_part_97) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_97), gsl_partVanish_97.zw); }
  }
  let idlerCapL = gsl_part_97;
  let gsl_partLo_98 = vec3<f32>(174.0, -226.0, 256.0);
  let gsl_partHi_98 = vec3<f32>(226.0, -174.0, 262.0);
  let gsl_partBox_98 = gsl_boxDistance(p, gsl_partLo_98, gsl_partHi_98);
  var gsl_part_98 = gsl_partBox_98;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_98 <= gsl_vanishView.w) {
    let gsl_partVanish_98 = gsl_vanish(gsl_partLo_98, gsl_partHi_98);
    gsl_part_98 = 1e10;
    if (gsl_partVanish_98.x > 0.0) { gsl_part_98 = builtin_gcl_closure_boxCentered(p, vec3<f32>(400.0, -400.0, 518.0) * 0.5, abs(vec3<f32>(52.0, 52.0, 6.0)), noiseTex) + gsl_partVanish_98.y; }
    if (abs(gsl_part_98) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_98), gsl_partVanish_98.zw); }
  }
  let idlerCapR = gsl_part_98;
  let gsl_partLo_99 = vec3<f32>(-236.0, min(gy - 26.0, gy + 26.0), 256.0);
  let gsl_partHi_99 = vec3<f32>(-190.0, max(gy - 26.0, gy + 26.0), 260.0);
  let gsl_partBox_99 = gsl_boxDistance(p, gsl_partLo_99, gsl_partHi_99);
  var gsl_part_99 = gsl_partBox_99;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_99 <= gsl_vanishView.w) {
    let gsl_partVanish_99 = gsl_vanish(gsl_partLo_99, gsl_partHi_99);
    gsl_part_99 = 1e10;
    if (gsl_partVanish_99.x > 0.0) { gsl_part_99 = builtin_box(p, vec3<f32>(-236.0, gy - 26.0, 256.0), vec3<f32>(-190.0, gy + 26.0, 260.0), noiseTex) + gsl_partVanish_99.y; }
    if (abs(gsl_part_99) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_99), gsl_partVanish_99.zw); }
  }
  let jointCapL = gsl_part_99;
  let gsl_partLo_100 = vec3<f32>(190.0, min(gy - 26.0, gy + 26.0), 256.0);
  let gsl_partHi_100 = vec3<f32>(236.0, max(gy - 26.0, gy + 26.0), 260.0);
  let gsl_partBox_100 = gsl_boxDistance(p, gsl_partLo_100, gsl_partHi_100);
  var gsl_part_100 = gsl_partBox_100;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_100 <= gsl_vanishView.w) {
    let gsl_partVanish_100 = gsl_vanish(gsl_partLo_100, gsl_partHi_100);
    gsl_part_100 = 1e10;
    if (gsl_partVanish_100.x > 0.0) { gsl_part_100 = builtin_box(p, vec3<f32>(190.0, gy - 26.0, 256.0), vec3<f32>(236.0, gy + 26.0, 260.0), noiseTex) + gsl_partVanish_100.y; }
    if (abs(gsl_part_100) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_100), gsl_partVanish_100.zw); }
  }
  let jointCapR = gsl_part_100;
  let gsl_partLo_101 = vec3<f32>(min(min(gx - 26.0, gx + 26.0), -(-0.0 - gx) - 16.0), min(min(gy - 75.0, gy - 72.0), -(-0.0 - (gy - 74.0)) - 1.0), 160.0);
  let gsl_partHi_101 = vec3<f32>(max(max(gx - 26.0, gx + 26.0), -(-0.0 - gx) + 16.0), max(max(gy - 75.0, gy - 72.0), -(-0.0 - (gy - 74.0)) + 1.0), 250.0);
  let gsl_partBox_101 = gsl_boxDistance(p, gsl_partLo_101, gsl_partHi_101);
  var gsl_part_101 = gsl_partBox_101;
  let fanOpening = builtin_cylinderCapped(p, vec3<f32>(gx, gy - 76.0, 218.0), vec3<f32>(gx, gy - 70.0, 218.0), 17.0, noiseTex);
  let gsl_partLo_102 = vec3<f32>(-(-0.0 - gx) - 16.0, -(-0.0 - (gy - 74.0)) - 1.0, 211.0);
  let gsl_partHi_102 = vec3<f32>(-(-0.0 - gx) + 16.0, -(-0.0 - (gy - 74.0)) + 1.0, 225.0);
  let gsl_partBox_102 = gsl_boxDistance(p, gsl_partLo_102, gsl_partHi_102);
  var gsl_part_102 = gsl_partBox_102;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_102 <= gsl_vanishView.w) {
    let gsl_partVanish_102 = gsl_vanish(gsl_partLo_102, gsl_partHi_102);
    gsl_part_102 = 1e10;
    if (gsl_partVanish_102.x > 0.0) { gsl_part_102 = max(max(abs(abs(p.z - 218.0) - 6.0) - 1.0, abs(p.x - gx) - 16.0), abs(p.y - (gy - 74.0)) - 1.0) + gsl_partVanish_102.y; }
    if (abs(gsl_part_102) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_102), gsl_partVanish_102.zw); }
  }
  let grille = gsl_part_102;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_101 <= gsl_vanishView.w) {
    let gsl_partVanish_101 = gsl_vanish(gsl_partLo_101, gsl_partHi_101);
    gsl_part_101 = 1e10;
    if (gsl_partVanish_101.x > 0.0) { gsl_part_101 = min(max(builtin_box(p, vec3<f32>(gx - 26.0, gy - 75.0, 160.0), vec3<f32>(gx + 26.0, gy - 72.0, 250.0), noiseTex), -fanOpening), grille) + gsl_partVanish_101.y; }
    if (abs(gsl_part_101) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_101), gsl_partVanish_101.zw); }
  }
  let headCover = gsl_part_101;
  let gsl_partLo_103 = vec3<f32>(-202.0, -210.0, 155.0);
  let gsl_partHi_103 = vec3<f32>(-168.0, -168.0, 180.0);
  let gsl_partBox_103 = gsl_boxDistance(p, gsl_partLo_103, gsl_partHi_103);
  var gsl_part_103 = gsl_partBox_103;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_103 <= gsl_vanishView.w) {
    let gsl_partVanish_103 = gsl_vanish(gsl_partLo_103, gsl_partHi_103);
    gsl_part_103 = 1e10;
    if (gsl_partVanish_103.x > 0.0) { gsl_part_103 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-370.0, -378.0, 335.0) * 0.5, abs(vec3<f32>(34.0, 42.0, 25.0)), noiseTex) + gsl_partVanish_103.y; }
    if (abs(gsl_part_103) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_103), gsl_partVanish_103.zw); }
  }
  let bearingFL = gsl_part_103;
  let gsl_partLo_104 = vec3<f32>(168.0, -210.0, 155.0);
  let gsl_partHi_104 = vec3<f32>(202.0, -168.0, 180.0);
  let gsl_partBox_104 = gsl_boxDistance(p, gsl_partLo_104, gsl_partHi_104);
  var gsl_part_104 = gsl_partBox_104;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_104 <= gsl_vanishView.w) {
    let gsl_partVanish_104 = gsl_vanish(gsl_partLo_104, gsl_partHi_104);
    gsl_part_104 = 1e10;
    if (gsl_partVanish_104.x > 0.0) { gsl_part_104 = builtin_gcl_closure_boxCentered(p, vec3<f32>(370.0, -378.0, 335.0) * 0.5, abs(vec3<f32>(34.0, 42.0, 25.0)), noiseTex) + gsl_partVanish_104.y; }
    if (abs(gsl_part_104) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_104), gsl_partVanish_104.zw); }
  }
  let bearingFR = gsl_part_104;
  let gsl_partLo_105 = vec3<f32>(-17.0, 168.0, 155.0);
  let gsl_partHi_105 = vec3<f32>(17.0, 210.0, 180.0);
  let gsl_partBox_105 = gsl_boxDistance(p, gsl_partLo_105, gsl_partHi_105);
  var gsl_part_105 = gsl_partBox_105;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_105 <= gsl_vanishView.w) {
    let gsl_partVanish_105 = gsl_vanish(gsl_partLo_105, gsl_partHi_105);
    gsl_part_105 = 1e10;
    if (gsl_partVanish_105.x > 0.0) { gsl_part_105 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 378.0, 335.0) * 0.5, abs(vec3<f32>(34.0, 42.0, 25.0)), noiseTex) + gsl_partVanish_105.y; }
    if (abs(gsl_part_105) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_105), gsl_partVanish_105.zw); }
  }
  let bearingB = gsl_part_105;
  let gsl_partLo_106 = vec3<f32>(-52.0, -250.0, -278.0);
  let gsl_partHi_106 = vec3<f32>(52.0, -230.0, -218.0);
  let gsl_partBox_106 = gsl_boxDistance(p, gsl_partLo_106, gsl_partHi_106);
  var gsl_part_106 = gsl_partBox_106;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_106 <= gsl_vanishView.w) {
    let gsl_partVanish_106 = gsl_vanish(gsl_partLo_106, gsl_partHi_106);
    gsl_part_106 = 1e10;
    if (gsl_partVanish_106.x > 0.0) { gsl_part_106 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, -480.0, -496.0) * 0.5, abs(vec3<f32>(104.0, 20.0, 60.0)), noiseTex) + gsl_partVanish_106.y; }
    if (abs(gsl_part_106) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_106), gsl_partVanish_106.zw); }
  }
  let displayCase = gsl_part_106;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_94 <= gsl_vanishView.w) {
    let gsl_partVanish_94 = gsl_vanish(gsl_partLo_94, gsl_partHi_94);
    gsl_part_94 = 1e10;
    if (gsl_partVanish_94.x > 0.0) { gsl_part_94 = min(min(min(min(min(min(min(min(min(min(driveCoverA, driveCoverB), idlerCapL), idlerCapR), jointCapL), jointCapR), headCover), bearingFL), bearingFR), bearingB), displayCase) + gsl_partVanish_94.y; }
    if (abs(gsl_part_94) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_94), gsl_partVanish_94.zw); }
  }
  let printedRed = gsl_part_94;
  let gsl_partLo_107 = vec3<f32>(-135.0, -135.0, min(bz - 10.5, bz - 9.0));
  let gsl_partHi_107 = vec3<f32>(135.0, 135.0, max(bz - 10.5, bz - 9.0));
  let gsl_partBox_107 = gsl_boxDistance(p, gsl_partLo_107, gsl_partHi_107);
  var gsl_part_107 = gsl_partBox_107;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_107 <= gsl_vanishView.w) {
    let gsl_partVanish_107 = gsl_vanish(gsl_partLo_107, gsl_partHi_107);
    gsl_part_107 = 1e10;
    if (gsl_partVanish_107.x > 0.0) { gsl_part_107 = builtin_box(p, vec3<f32>(-135.0, -135.0, bz - 10.5), vec3<f32>(135.0, 135.0, bz - 9.0), noiseTex) + gsl_partVanish_107.y; }
    if (abs(gsl_part_107) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_107), gsl_partVanish_107.zw); }
  }
  let bedHeater = gsl_part_107;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_93 <= gsl_vanishView.w) {
    let gsl_partVanish_93 = gsl_vanish(gsl_partLo_93, gsl_partHi_93);
    gsl_part_93 = 1e10;
    if (gsl_partVanish_93.x > 0.0) { gsl_part_93 = min(printedRed, bedHeater) + gsl_partVanish_93.y; }
    if (abs(gsl_part_93) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_93), gsl_partVanish_93.zw); }
  }
  let red = gsl_part_93;
  let gsl_partLo_108 = vec3<f32>(min(min(-235.01, -2.86 + gx - 20.0), -2.86 + gx + 20.0), min(min(min(min(-235.01, -4.26 + gy), -4.26 + gy), -1.66 + gy - 75.0), -1.66 + gy - 75.0), -294.26);
  let gsl_partHi_108 = vec3<f32>(max(max(235.01, 2.86 + gx - 20.0), 2.86 + gx + 20.0), max(max(max(max(235.01, 4.26 + gy), 4.26 + gy), 8.01 + gy - 75.0), 8.01 + gy - 75.0), 267.01);
  let gsl_partBox_108 = gsl_boxDistance(p, gsl_partLo_108, gsl_partHi_108);
  var gsl_part_108 = gsl_partBox_108;
  let gsl_partLo_109 = vec3<f32>(-235.01, -235.01, -294.26);
  let gsl_partHi_109 = vec3<f32>(235.01, 235.01, 194.26);
  let gsl_partBox_109 = gsl_boxDistance(p, gsl_partLo_109, gsl_partHi_109);
  var gsl_part_109 = gsl_partBox_109;
  let gsl_partLo_110 = vec3<f32>(-235.01, -224.26, -294.26);
  let gsl_partHi_110 = vec3<f32>(-217.99, -215.74, -285.74);
  let gsl_partBox_110 = gsl_boxDistance(p, gsl_partLo_110, gsl_partHi_110);
  var gsl_part_110 = gsl_partBox_110;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_110 <= gsl_vanishView.w) {
    let gsl_partVanish_110 = gsl_vanish(gsl_partLo_110, gsl_partHi_110);
    gsl_part_110 = 1e10;
    if (gsl_partVanish_110.x > 0.0) { gsl_part_110 = gsl_remap_8(p - vec3<f32>(-230.0, -220.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_110.y; }
    if (abs(gsl_part_110) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_110), gsl_partVanish_110.zw); }
  }
  let floorJ1 = gsl_part_110;
  let gsl_partLo_111 = vec3<f32>(-235.01, 215.74, -294.26);
  let gsl_partHi_111 = vec3<f32>(-217.99, 224.26, -285.74);
  let gsl_partBox_111 = gsl_boxDistance(p, gsl_partLo_111, gsl_partHi_111);
  var gsl_part_111 = gsl_partBox_111;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_111 <= gsl_vanishView.w) {
    let gsl_partVanish_111 = gsl_vanish(gsl_partLo_111, gsl_partHi_111);
    gsl_part_111 = 1e10;
    if (gsl_partVanish_111.x > 0.0) { gsl_part_111 = gsl_remap_8(p - vec3<f32>(-230.0, 220.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_111.y; }
    if (abs(gsl_part_111) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_111), gsl_partVanish_111.zw); }
  }
  let floorJ2 = gsl_part_111;
  let gsl_partLo_112 = vec3<f32>(217.99, -224.26, -294.26);
  let gsl_partHi_112 = vec3<f32>(235.01, -215.74, -285.74);
  let gsl_partBox_112 = gsl_boxDistance(p, gsl_partLo_112, gsl_partHi_112);
  var gsl_part_112 = gsl_partBox_112;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_112 <= gsl_vanishView.w) {
    let gsl_partVanish_112 = gsl_vanish(gsl_partLo_112, gsl_partHi_112);
    gsl_part_112 = 1e10;
    if (gsl_partVanish_112.x > 0.0) { gsl_part_112 = gsl_remap_9(p - vec3<f32>(230.0, -220.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_112.y; }
    if (abs(gsl_part_112) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_112), gsl_partVanish_112.zw); }
  }
  let floorJ3 = gsl_part_112;
  let gsl_partLo_113 = vec3<f32>(217.99, 215.74, -294.26);
  let gsl_partHi_113 = vec3<f32>(235.01, 224.26, -285.74);
  let gsl_partBox_113 = gsl_boxDistance(p, gsl_partLo_113, gsl_partHi_113);
  var gsl_part_113 = gsl_partBox_113;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_113 <= gsl_vanishView.w) {
    let gsl_partVanish_113 = gsl_vanish(gsl_partLo_113, gsl_partHi_113);
    gsl_part_113 = 1e10;
    if (gsl_partVanish_113.x > 0.0) { gsl_part_113 = gsl_remap_9(p - vec3<f32>(230.0, 220.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_113.y; }
    if (abs(gsl_part_113) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_113), gsl_partVanish_113.zw); }
  }
  let floorJ4 = gsl_part_113;
  let gsl_partLo_114 = vec3<f32>(-224.26, -235.01, -294.26);
  let gsl_partHi_114 = vec3<f32>(-215.74, -217.99, -285.74);
  let gsl_partBox_114 = gsl_boxDistance(p, gsl_partLo_114, gsl_partHi_114);
  var gsl_part_114 = gsl_partBox_114;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_114 <= gsl_vanishView.w) {
    let gsl_partVanish_114 = gsl_vanish(gsl_partLo_114, gsl_partHi_114);
    gsl_part_114 = 1e10;
    if (gsl_partVanish_114.x > 0.0) { gsl_part_114 = gsl_remap_10(p - vec3<f32>(-220.0, -230.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_114.y; }
    if (abs(gsl_part_114) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_114), gsl_partVanish_114.zw); }
  }
  let floorJ5 = gsl_part_114;
  let gsl_partLo_115 = vec3<f32>(215.74, -235.01, -294.26);
  let gsl_partHi_115 = vec3<f32>(224.26, -217.99, -285.74);
  let gsl_partBox_115 = gsl_boxDistance(p, gsl_partLo_115, gsl_partHi_115);
  var gsl_part_115 = gsl_partBox_115;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_115 <= gsl_vanishView.w) {
    let gsl_partVanish_115 = gsl_vanish(gsl_partLo_115, gsl_partHi_115);
    gsl_part_115 = 1e10;
    if (gsl_partVanish_115.x > 0.0) { gsl_part_115 = gsl_remap_10(p - vec3<f32>(220.0, -230.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_115.y; }
    if (abs(gsl_part_115) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_115), gsl_partVanish_115.zw); }
  }
  let floorJ6 = gsl_part_115;
  let gsl_partLo_116 = vec3<f32>(-224.26, 217.99, -294.26);
  let gsl_partHi_116 = vec3<f32>(-215.74, 235.01, -285.74);
  let gsl_partBox_116 = gsl_boxDistance(p, gsl_partLo_116, gsl_partHi_116);
  var gsl_part_116 = gsl_partBox_116;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_116 <= gsl_vanishView.w) {
    let gsl_partVanish_116 = gsl_vanish(gsl_partLo_116, gsl_partHi_116);
    gsl_part_116 = 1e10;
    if (gsl_partVanish_116.x > 0.0) { gsl_part_116 = gsl_remap_11(p - vec3<f32>(-220.0, 230.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_116.y; }
    if (abs(gsl_part_116) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_116), gsl_partVanish_116.zw); }
  }
  let floorJ7 = gsl_part_116;
  let gsl_partLo_117 = vec3<f32>(215.74, 217.99, -294.26);
  let gsl_partHi_117 = vec3<f32>(224.26, 235.01, -285.74);
  let gsl_partBox_117 = gsl_boxDistance(p, gsl_partLo_117, gsl_partHi_117);
  var gsl_part_117 = gsl_partBox_117;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_117 <= gsl_vanishView.w) {
    let gsl_partVanish_117 = gsl_vanish(gsl_partLo_117, gsl_partHi_117);
    gsl_part_117 = 1e10;
    if (gsl_partVanish_117.x > 0.0) { gsl_part_117 = gsl_remap_11(p - vec3<f32>(220.0, 230.0, -290.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_117.y; }
    if (abs(gsl_part_117) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_117), gsl_partVanish_117.zw); }
  }
  let floorJ8 = gsl_part_117;
  let gsl_partLo_118 = vec3<f32>(-235.01, -224.26, -214.26);
  let gsl_partHi_118 = vec3<f32>(-217.99, -215.74, -205.74);
  let gsl_partBox_118 = gsl_boxDistance(p, gsl_partLo_118, gsl_partHi_118);
  var gsl_part_118 = gsl_partBox_118;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_118 <= gsl_vanishView.w) {
    let gsl_partVanish_118 = gsl_vanish(gsl_partLo_118, gsl_partHi_118);
    gsl_part_118 = 1e10;
    if (gsl_partVanish_118.x > 0.0) { gsl_part_118 = gsl_remap_8(p - vec3<f32>(-230.0, -220.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_118.y; }
    if (abs(gsl_part_118) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_118), gsl_partVanish_118.zw); }
  }
  let deckJ1 = gsl_part_118;
  let gsl_partLo_119 = vec3<f32>(-235.01, 215.74, -214.26);
  let gsl_partHi_119 = vec3<f32>(-217.99, 224.26, -205.74);
  let gsl_partBox_119 = gsl_boxDistance(p, gsl_partLo_119, gsl_partHi_119);
  var gsl_part_119 = gsl_partBox_119;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_119 <= gsl_vanishView.w) {
    let gsl_partVanish_119 = gsl_vanish(gsl_partLo_119, gsl_partHi_119);
    gsl_part_119 = 1e10;
    if (gsl_partVanish_119.x > 0.0) { gsl_part_119 = gsl_remap_8(p - vec3<f32>(-230.0, 220.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_119.y; }
    if (abs(gsl_part_119) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_119), gsl_partVanish_119.zw); }
  }
  let deckJ2 = gsl_part_119;
  let gsl_partLo_120 = vec3<f32>(217.99, -224.26, -214.26);
  let gsl_partHi_120 = vec3<f32>(235.01, -215.74, -205.74);
  let gsl_partBox_120 = gsl_boxDistance(p, gsl_partLo_120, gsl_partHi_120);
  var gsl_part_120 = gsl_partBox_120;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_120 <= gsl_vanishView.w) {
    let gsl_partVanish_120 = gsl_vanish(gsl_partLo_120, gsl_partHi_120);
    gsl_part_120 = 1e10;
    if (gsl_partVanish_120.x > 0.0) { gsl_part_120 = gsl_remap_9(p - vec3<f32>(230.0, -220.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_120.y; }
    if (abs(gsl_part_120) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_120), gsl_partVanish_120.zw); }
  }
  let deckJ3 = gsl_part_120;
  let gsl_partLo_121 = vec3<f32>(217.99, 215.74, -214.26);
  let gsl_partHi_121 = vec3<f32>(235.01, 224.26, -205.74);
  let gsl_partBox_121 = gsl_boxDistance(p, gsl_partLo_121, gsl_partHi_121);
  var gsl_part_121 = gsl_partBox_121;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_121 <= gsl_vanishView.w) {
    let gsl_partVanish_121 = gsl_vanish(gsl_partLo_121, gsl_partHi_121);
    gsl_part_121 = 1e10;
    if (gsl_partVanish_121.x > 0.0) { gsl_part_121 = gsl_remap_9(p - vec3<f32>(230.0, 220.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_121.y; }
    if (abs(gsl_part_121) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_121), gsl_partVanish_121.zw); }
  }
  let deckJ4 = gsl_part_121;
  let gsl_partLo_122 = vec3<f32>(-224.26, -235.01, -214.26);
  let gsl_partHi_122 = vec3<f32>(-215.74, -217.99, -205.74);
  let gsl_partBox_122 = gsl_boxDistance(p, gsl_partLo_122, gsl_partHi_122);
  var gsl_part_122 = gsl_partBox_122;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_122 <= gsl_vanishView.w) {
    let gsl_partVanish_122 = gsl_vanish(gsl_partLo_122, gsl_partHi_122);
    gsl_part_122 = 1e10;
    if (gsl_partVanish_122.x > 0.0) { gsl_part_122 = gsl_remap_10(p - vec3<f32>(-220.0, -230.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_122.y; }
    if (abs(gsl_part_122) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_122), gsl_partVanish_122.zw); }
  }
  let deckJ5 = gsl_part_122;
  let gsl_partLo_123 = vec3<f32>(215.74, -235.01, -214.26);
  let gsl_partHi_123 = vec3<f32>(224.26, -217.99, -205.74);
  let gsl_partBox_123 = gsl_boxDistance(p, gsl_partLo_123, gsl_partHi_123);
  var gsl_part_123 = gsl_partBox_123;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_123 <= gsl_vanishView.w) {
    let gsl_partVanish_123 = gsl_vanish(gsl_partLo_123, gsl_partHi_123);
    gsl_part_123 = 1e10;
    if (gsl_partVanish_123.x > 0.0) { gsl_part_123 = gsl_remap_10(p - vec3<f32>(220.0, -230.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_123.y; }
    if (abs(gsl_part_123) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_123), gsl_partVanish_123.zw); }
  }
  let deckJ6 = gsl_part_123;
  let gsl_partLo_124 = vec3<f32>(-224.26, 217.99, -214.26);
  let gsl_partHi_124 = vec3<f32>(-215.74, 235.01, -205.74);
  let gsl_partBox_124 = gsl_boxDistance(p, gsl_partLo_124, gsl_partHi_124);
  var gsl_part_124 = gsl_partBox_124;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_124 <= gsl_vanishView.w) {
    let gsl_partVanish_124 = gsl_vanish(gsl_partLo_124, gsl_partHi_124);
    gsl_part_124 = 1e10;
    if (gsl_partVanish_124.x > 0.0) { gsl_part_124 = gsl_remap_11(p - vec3<f32>(-220.0, 230.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_124.y; }
    if (abs(gsl_part_124) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_124), gsl_partVanish_124.zw); }
  }
  let deckJ7 = gsl_part_124;
  let gsl_partLo_125 = vec3<f32>(215.74, 217.99, -214.26);
  let gsl_partHi_125 = vec3<f32>(224.26, 235.01, -205.74);
  let gsl_partBox_125 = gsl_boxDistance(p, gsl_partLo_125, gsl_partHi_125);
  var gsl_part_125 = gsl_partBox_125;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_125 <= gsl_vanishView.w) {
    let gsl_partVanish_125 = gsl_vanish(gsl_partLo_125, gsl_partHi_125);
    gsl_part_125 = 1e10;
    if (gsl_partVanish_125.x > 0.0) { gsl_part_125 = gsl_remap_11(p - vec3<f32>(220.0, 230.0, -210.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_125.y; }
    if (abs(gsl_part_125) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_125), gsl_partVanish_125.zw); }
  }
  let deckJ8 = gsl_part_125;
  let gsl_partLo_126 = vec3<f32>(-235.01, -224.26, 185.74);
  let gsl_partHi_126 = vec3<f32>(-217.99, -215.74, 194.26);
  let gsl_partBox_126 = gsl_boxDistance(p, gsl_partLo_126, gsl_partHi_126);
  var gsl_part_126 = gsl_partBox_126;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_126 <= gsl_vanishView.w) {
    let gsl_partVanish_126 = gsl_vanish(gsl_partLo_126, gsl_partHi_126);
    gsl_part_126 = 1e10;
    if (gsl_partVanish_126.x > 0.0) { gsl_part_126 = gsl_remap_8(p - vec3<f32>(-230.0, -220.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_126.y; }
    if (abs(gsl_part_126) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_126), gsl_partVanish_126.zw); }
  }
  let topJ1 = gsl_part_126;
  let gsl_partLo_127 = vec3<f32>(-235.01, 215.74, 185.74);
  let gsl_partHi_127 = vec3<f32>(-217.99, 224.26, 194.26);
  let gsl_partBox_127 = gsl_boxDistance(p, gsl_partLo_127, gsl_partHi_127);
  var gsl_part_127 = gsl_partBox_127;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_127 <= gsl_vanishView.w) {
    let gsl_partVanish_127 = gsl_vanish(gsl_partLo_127, gsl_partHi_127);
    gsl_part_127 = 1e10;
    if (gsl_partVanish_127.x > 0.0) { gsl_part_127 = gsl_remap_8(p - vec3<f32>(-230.0, 220.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_127.y; }
    if (abs(gsl_part_127) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_127), gsl_partVanish_127.zw); }
  }
  let topJ2 = gsl_part_127;
  let gsl_partLo_128 = vec3<f32>(217.99, -224.26, 185.74);
  let gsl_partHi_128 = vec3<f32>(235.01, -215.74, 194.26);
  let gsl_partBox_128 = gsl_boxDistance(p, gsl_partLo_128, gsl_partHi_128);
  var gsl_part_128 = gsl_partBox_128;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_128 <= gsl_vanishView.w) {
    let gsl_partVanish_128 = gsl_vanish(gsl_partLo_128, gsl_partHi_128);
    gsl_part_128 = 1e10;
    if (gsl_partVanish_128.x > 0.0) { gsl_part_128 = gsl_remap_9(p - vec3<f32>(230.0, -220.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_128.y; }
    if (abs(gsl_part_128) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_128), gsl_partVanish_128.zw); }
  }
  let topJ3 = gsl_part_128;
  let gsl_partLo_129 = vec3<f32>(217.99, 215.74, 185.74);
  let gsl_partHi_129 = vec3<f32>(235.01, 224.26, 194.26);
  let gsl_partBox_129 = gsl_boxDistance(p, gsl_partLo_129, gsl_partHi_129);
  var gsl_part_129 = gsl_partBox_129;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_129 <= gsl_vanishView.w) {
    let gsl_partVanish_129 = gsl_vanish(gsl_partLo_129, gsl_partHi_129);
    gsl_part_129 = 1e10;
    if (gsl_partVanish_129.x > 0.0) { gsl_part_129 = gsl_remap_9(p - vec3<f32>(230.0, 220.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_129.y; }
    if (abs(gsl_part_129) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_129), gsl_partVanish_129.zw); }
  }
  let topJ4 = gsl_part_129;
  let gsl_partLo_130 = vec3<f32>(-224.26, -235.01, 185.74);
  let gsl_partHi_130 = vec3<f32>(-215.74, -217.99, 194.26);
  let gsl_partBox_130 = gsl_boxDistance(p, gsl_partLo_130, gsl_partHi_130);
  var gsl_part_130 = gsl_partBox_130;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_130 <= gsl_vanishView.w) {
    let gsl_partVanish_130 = gsl_vanish(gsl_partLo_130, gsl_partHi_130);
    gsl_part_130 = 1e10;
    if (gsl_partVanish_130.x > 0.0) { gsl_part_130 = gsl_remap_10(p - vec3<f32>(-220.0, -230.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_130.y; }
    if (abs(gsl_part_130) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_130), gsl_partVanish_130.zw); }
  }
  let topJ5 = gsl_part_130;
  let gsl_partLo_131 = vec3<f32>(215.74, -235.01, 185.74);
  let gsl_partHi_131 = vec3<f32>(224.26, -217.99, 194.26);
  let gsl_partBox_131 = gsl_boxDistance(p, gsl_partLo_131, gsl_partHi_131);
  var gsl_part_131 = gsl_partBox_131;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_131 <= gsl_vanishView.w) {
    let gsl_partVanish_131 = gsl_vanish(gsl_partLo_131, gsl_partHi_131);
    gsl_part_131 = 1e10;
    if (gsl_partVanish_131.x > 0.0) { gsl_part_131 = gsl_remap_10(p - vec3<f32>(220.0, -230.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_131.y; }
    if (abs(gsl_part_131) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_131), gsl_partVanish_131.zw); }
  }
  let topJ6 = gsl_part_131;
  let gsl_partLo_132 = vec3<f32>(-224.26, 217.99, 185.74);
  let gsl_partHi_132 = vec3<f32>(-215.74, 235.01, 194.26);
  let gsl_partBox_132 = gsl_boxDistance(p, gsl_partLo_132, gsl_partHi_132);
  var gsl_part_132 = gsl_partBox_132;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_132 <= gsl_vanishView.w) {
    let gsl_partVanish_132 = gsl_vanish(gsl_partLo_132, gsl_partHi_132);
    gsl_part_132 = 1e10;
    if (gsl_partVanish_132.x > 0.0) { gsl_part_132 = gsl_remap_11(p - vec3<f32>(-220.0, 230.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_132.y; }
    if (abs(gsl_part_132) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_132), gsl_partVanish_132.zw); }
  }
  let topJ7 = gsl_part_132;
  let gsl_partLo_133 = vec3<f32>(215.74, 217.99, 185.74);
  let gsl_partHi_133 = vec3<f32>(224.26, 235.01, 194.26);
  let gsl_partBox_133 = gsl_boxDistance(p, gsl_partLo_133, gsl_partHi_133);
  var gsl_part_133 = gsl_partBox_133;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_133 <= gsl_vanishView.w) {
    let gsl_partVanish_133 = gsl_vanish(gsl_partLo_133, gsl_partHi_133);
    gsl_part_133 = 1e10;
    if (gsl_partVanish_133.x > 0.0) { gsl_part_133 = gsl_remap_11(p - vec3<f32>(220.0, 230.0, 190.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_133.y; }
    if (abs(gsl_part_133) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_133), gsl_partVanish_133.zw); }
  }
  let topJ8 = gsl_part_133;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_109 <= gsl_vanishView.w) {
    let gsl_partVanish_109 = gsl_vanish(gsl_partLo_109, gsl_partHi_109);
    gsl_part_109 = 1e10;
    if (gsl_partVanish_109.x > 0.0) { gsl_part_109 = min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(floorJ1, floorJ2), floorJ3), floorJ4), floorJ5), floorJ6), floorJ7), floorJ8), deckJ1), deckJ2), deckJ3), deckJ4), deckJ5), deckJ6), deckJ7), deckJ8), topJ1), topJ2), topJ3), topJ4), topJ5), topJ6), topJ7), topJ8) + gsl_partVanish_109.y; }
    if (abs(gsl_part_109) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_109), gsl_partVanish_109.zw); }
  }
  let jointScrews = gsl_part_109;
  let gsl_partLo_134 = vec3<f32>(-222.86, 177.14, 251.99);
  let gsl_partHi_134 = vec3<f32>(222.86, 222.86, 261.66);
  let gsl_partBox_134 = gsl_boxDistance(p, gsl_partLo_134, gsl_partHi_134);
  var gsl_part_134 = gsl_partBox_134;
  let gsl_partLo_135 = vec3<f32>(-222.86, 177.14, 251.99);
  let gsl_partHi_135 = vec3<f32>(-217.14, 182.86, 261.66);
  let gsl_partBox_135 = gsl_boxDistance(p, gsl_partLo_135, gsl_partHi_135);
  var gsl_part_135 = gsl_partBox_135;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_135 <= gsl_vanishView.w) {
    let gsl_partVanish_135 = gsl_vanish(gsl_partLo_135, gsl_partHi_135);
    gsl_part_135 = 1e10;
    if (gsl_partVanish_135.x > 0.0) { gsl_part_135 = gsl_remap_12(p - vec3<f32>(-220.0, 180.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_135.y; }
    if (abs(gsl_part_135) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_135), gsl_partVanish_135.zw); }
  }
  let plateA1 = gsl_part_135;
  let gsl_partLo_136 = vec3<f32>(-182.86, 177.14, 251.99);
  let gsl_partHi_136 = vec3<f32>(-177.14, 182.86, 261.66);
  let gsl_partBox_136 = gsl_boxDistance(p, gsl_partLo_136, gsl_partHi_136);
  var gsl_part_136 = gsl_partBox_136;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_136 <= gsl_vanishView.w) {
    let gsl_partVanish_136 = gsl_vanish(gsl_partLo_136, gsl_partHi_136);
    gsl_part_136 = 1e10;
    if (gsl_partVanish_136.x > 0.0) { gsl_part_136 = gsl_remap_12(p - vec3<f32>(-180.0, 180.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_136.y; }
    if (abs(gsl_part_136) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_136), gsl_partVanish_136.zw); }
  }
  let plateA2 = gsl_part_136;
  let gsl_partLo_137 = vec3<f32>(-222.86, 217.14, 251.99);
  let gsl_partHi_137 = vec3<f32>(-217.14, 222.86, 261.66);
  let gsl_partBox_137 = gsl_boxDistance(p, gsl_partLo_137, gsl_partHi_137);
  var gsl_part_137 = gsl_partBox_137;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_137 <= gsl_vanishView.w) {
    let gsl_partVanish_137 = gsl_vanish(gsl_partLo_137, gsl_partHi_137);
    gsl_part_137 = 1e10;
    if (gsl_partVanish_137.x > 0.0) { gsl_part_137 = gsl_remap_12(p - vec3<f32>(-220.0, 220.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_137.y; }
    if (abs(gsl_part_137) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_137), gsl_partVanish_137.zw); }
  }
  let plateA3 = gsl_part_137;
  let gsl_partLo_138 = vec3<f32>(-182.86, 217.14, 251.99);
  let gsl_partHi_138 = vec3<f32>(-177.14, 222.86, 261.66);
  let gsl_partBox_138 = gsl_boxDistance(p, gsl_partLo_138, gsl_partHi_138);
  var gsl_part_138 = gsl_partBox_138;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_138 <= gsl_vanishView.w) {
    let gsl_partVanish_138 = gsl_vanish(gsl_partLo_138, gsl_partHi_138);
    gsl_part_138 = 1e10;
    if (gsl_partVanish_138.x > 0.0) { gsl_part_138 = gsl_remap_12(p - vec3<f32>(-180.0, 220.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_138.y; }
    if (abs(gsl_part_138) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_138), gsl_partVanish_138.zw); }
  }
  let plateA4 = gsl_part_138;
  let gsl_partLo_139 = vec3<f32>(177.14, 177.14, 251.99);
  let gsl_partHi_139 = vec3<f32>(182.86, 182.86, 261.66);
  let gsl_partBox_139 = gsl_boxDistance(p, gsl_partLo_139, gsl_partHi_139);
  var gsl_part_139 = gsl_partBox_139;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_139 <= gsl_vanishView.w) {
    let gsl_partVanish_139 = gsl_vanish(gsl_partLo_139, gsl_partHi_139);
    gsl_part_139 = 1e10;
    if (gsl_partVanish_139.x > 0.0) { gsl_part_139 = gsl_remap_12(p - vec3<f32>(180.0, 180.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_139.y; }
    if (abs(gsl_part_139) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_139), gsl_partVanish_139.zw); }
  }
  let plateB1 = gsl_part_139;
  let gsl_partLo_140 = vec3<f32>(217.14, 177.14, 251.99);
  let gsl_partHi_140 = vec3<f32>(222.86, 182.86, 261.66);
  let gsl_partBox_140 = gsl_boxDistance(p, gsl_partLo_140, gsl_partHi_140);
  var gsl_part_140 = gsl_partBox_140;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_140 <= gsl_vanishView.w) {
    let gsl_partVanish_140 = gsl_vanish(gsl_partLo_140, gsl_partHi_140);
    gsl_part_140 = 1e10;
    if (gsl_partVanish_140.x > 0.0) { gsl_part_140 = gsl_remap_12(p - vec3<f32>(220.0, 180.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_140.y; }
    if (abs(gsl_part_140) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_140), gsl_partVanish_140.zw); }
  }
  let plateB2 = gsl_part_140;
  let gsl_partLo_141 = vec3<f32>(177.14, 217.14, 251.99);
  let gsl_partHi_141 = vec3<f32>(182.86, 222.86, 261.66);
  let gsl_partBox_141 = gsl_boxDistance(p, gsl_partLo_141, gsl_partHi_141);
  var gsl_part_141 = gsl_partBox_141;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_141 <= gsl_vanishView.w) {
    let gsl_partVanish_141 = gsl_vanish(gsl_partLo_141, gsl_partHi_141);
    gsl_part_141 = 1e10;
    if (gsl_partVanish_141.x > 0.0) { gsl_part_141 = gsl_remap_12(p - vec3<f32>(180.0, 220.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_141.y; }
    if (abs(gsl_part_141) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_141), gsl_partVanish_141.zw); }
  }
  let plateB3 = gsl_part_141;
  let gsl_partLo_142 = vec3<f32>(217.14, 217.14, 251.99);
  let gsl_partHi_142 = vec3<f32>(222.86, 222.86, 261.66);
  let gsl_partBox_142 = gsl_boxDistance(p, gsl_partLo_142, gsl_partHi_142);
  var gsl_part_142 = gsl_partBox_142;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_142 <= gsl_vanishView.w) {
    let gsl_partVanish_142 = gsl_vanish(gsl_partLo_142, gsl_partHi_142);
    gsl_part_142 = 1e10;
    if (gsl_partVanish_142.x > 0.0) { gsl_part_142 = gsl_remap_12(p - vec3<f32>(220.0, 220.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_142.y; }
    if (abs(gsl_part_142) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_142), gsl_partVanish_142.zw); }
  }
  let plateB4 = gsl_part_142;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_134 <= gsl_vanishView.w) {
    let gsl_partVanish_134 = gsl_vanish(gsl_partLo_134, gsl_partHi_134);
    gsl_part_134 = 1e10;
    if (gsl_partVanish_134.x > 0.0) { gsl_part_134 = min(min(min(min(min(min(min(plateA1, plateA2), plateA3), plateA4), plateB1), plateB2), plateB3), plateB4) + gsl_partVanish_134.y; }
    if (abs(gsl_part_134) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_134), gsl_partVanish_134.zw); }
  }
  let plateScrews = gsl_part_134;
  let gsl_partLo_143 = vec3<f32>(-204.26, -204.26, 231.99);
  let gsl_partHi_143 = vec3<f32>(-195.74, -195.74, 267.01);
  let gsl_partBox_143 = gsl_boxDistance(p, gsl_partLo_143, gsl_partHi_143);
  var gsl_part_143 = gsl_partBox_143;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_143 <= gsl_vanishView.w) {
    let gsl_partVanish_143 = gsl_vanish(gsl_partLo_143, gsl_partHi_143);
    gsl_part_143 = 1e10;
    if (gsl_partVanish_143.x > 0.0) { gsl_part_143 = gsl_remap_13(p - vec3<f32>(-200.0, -200.0, 262.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_143.y; }
    if (abs(gsl_part_143) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_143), gsl_partVanish_143.zw); }
  }
  let idlerBoltL = gsl_part_143;
  let gsl_partLo_144 = vec3<f32>(195.74, -204.26, 231.99);
  let gsl_partHi_144 = vec3<f32>(204.26, -195.74, 267.01);
  let gsl_partBox_144 = gsl_boxDistance(p, gsl_partLo_144, gsl_partHi_144);
  var gsl_part_144 = gsl_partBox_144;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_144 <= gsl_vanishView.w) {
    let gsl_partVanish_144 = gsl_vanish(gsl_partLo_144, gsl_partHi_144);
    gsl_part_144 = 1e10;
    if (gsl_partVanish_144.x > 0.0) { gsl_part_144 = gsl_remap_13(p - vec3<f32>(200.0, -200.0, 262.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_144.y; }
    if (abs(gsl_part_144) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_144), gsl_partVanish_144.zw); }
  }
  let idlerBoltR = gsl_part_144;
  let gsl_partLo_145 = vec3<f32>(-204.26, -4.26 + gy, 229.99);
  let gsl_partHi_145 = vec3<f32>(-195.74, 4.26 + gy, 265.01);
  let gsl_partBox_145 = gsl_boxDistance(p, gsl_partLo_145, gsl_partHi_145);
  var gsl_part_145 = gsl_partBox_145;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_145 <= gsl_vanishView.w) {
    let gsl_partVanish_145 = gsl_vanish(gsl_partLo_145, gsl_partHi_145);
    gsl_part_145 = 1e10;
    if (gsl_partVanish_145.x > 0.0) { gsl_part_145 = gsl_remap_13(p - vec3<f32>(-200.0, gy, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_145.y; }
    if (abs(gsl_part_145) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_145), gsl_partVanish_145.zw); }
  }
  let jointBoltL = gsl_part_145;
  let gsl_partLo_146 = vec3<f32>(195.74, -4.26 + gy, 229.99);
  let gsl_partHi_146 = vec3<f32>(204.26, 4.26 + gy, 265.01);
  let gsl_partBox_146 = gsl_boxDistance(p, gsl_partLo_146, gsl_partHi_146);
  var gsl_part_146 = gsl_partBox_146;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_146 <= gsl_vanishView.w) {
    let gsl_partVanish_146 = gsl_vanish(gsl_partLo_146, gsl_partHi_146);
    gsl_part_146 = 1e10;
    if (gsl_partVanish_146.x > 0.0) { gsl_part_146 = gsl_remap_13(p - vec3<f32>(200.0, gy, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_146.y; }
    if (abs(gsl_part_146) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_146), gsl_partVanish_146.zw); }
  }
  let jointBoltR = gsl_part_146;
  let gsl_partLo_147 = vec3<f32>(-2.86 + gx - 20.0, -1.66 + gy - 75.0, 241.14);
  let gsl_partHi_147 = vec3<f32>(2.86 + gx - 20.0, 8.01 + gy - 75.0, 246.86);
  let gsl_partBox_147 = gsl_boxDistance(p, gsl_partLo_147, gsl_partHi_147);
  var gsl_part_147 = gsl_partBox_147;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_147 <= gsl_vanishView.w) {
    let gsl_partVanish_147 = gsl_vanish(gsl_partLo_147, gsl_partHi_147);
    gsl_part_147 = 1e10;
    if (gsl_partVanish_147.x > 0.0) { gsl_part_147 = gsl_remap_14(p - vec3<f32>(gx - 20.0, gy - 75.0, 244.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_147.y; }
    if (abs(gsl_part_147) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_147), gsl_partVanish_147.zw); }
  }
  let coverScrewL = gsl_part_147;
  let gsl_partLo_148 = vec3<f32>(-2.86 + gx + 20.0, -1.66 + gy - 75.0, 241.14);
  let gsl_partHi_148 = vec3<f32>(2.86 + gx + 20.0, 8.01 + gy - 75.0, 246.86);
  let gsl_partBox_148 = gsl_boxDistance(p, gsl_partLo_148, gsl_partHi_148);
  var gsl_part_148 = gsl_partBox_148;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_148 <= gsl_vanishView.w) {
    let gsl_partVanish_148 = gsl_vanish(gsl_partLo_148, gsl_partHi_148);
    gsl_part_148 = 1e10;
    if (gsl_partVanish_148.x > 0.0) { gsl_part_148 = gsl_remap_14(p - vec3<f32>(gx + 20.0, gy - 75.0, 244.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_148.y; }
    if (abs(gsl_part_148) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_148), gsl_partVanish_148.zw); }
  }
  let coverScrewR = gsl_part_148;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_108 <= gsl_vanishView.w) {
    let gsl_partVanish_108 = gsl_vanish(gsl_partLo_108, gsl_partHi_108);
    gsl_part_108 = 1e10;
    if (gsl_partVanish_108.x > 0.0) { gsl_part_108 = min(min(min(min(min(min(min(jointScrews, plateScrews), idlerBoltL), idlerBoltR), jointBoltL), jointBoltR), coverScrewL), coverScrewR) + gsl_partVanish_108.y; }
    if (abs(gsl_part_108) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_108), gsl_partVanish_108.zw); }
  }
  let fasteners = gsl_part_108;
  let gsl_partLo_149 = vec3<f32>(min(min(-233.5, -(-0.0 - gx) - 19.95), -221.15), min(min(min(min(min(-170.0, -(gy - 10.0 - 3.25) / -1.0 - 3.25), -(-0.0 - gy) - 22.7), -(-0.0 - gy) - 22.7), -(gy - 10.0 - 6.5) / -1.0 - 3.5), -206.15), -238.0);
  let gsl_partHi_149 = vec3<f32>(max(max(233.5, -(-0.0 - gx) + 19.95), 221.15), max(max(max(max(max(170.0, -(gy - 10.0 - 3.25) / -1.0 + 3.25), -(-0.0 - gy) + 22.7), -(-0.0 - gy) + 22.7), -(gy - 10.0 - 6.5) / -1.0 + 3.5), 221.15), 298.0);
  let gsl_partBox_149 = gsl_boxDistance(p, gsl_partLo_149, gsl_partHi_149);
  var gsl_part_149 = gsl_partBox_149;
  let gsl_partLo_150 = vec3<f32>(min(-233.5, -(-0.0 - gx) - 19.95), min(min(min(min(-170.0, -(gy - 10.0 - 3.25) / -1.0 - 3.25), -(-0.0 - gy) - 22.7), -(-0.0 - gy) - 22.7), -(gy - 10.0 - 6.5) / -1.0 - 3.5), 200.0);
  let gsl_partHi_150 = vec3<f32>(max(233.5, -(-0.0 - gx) + 19.95), max(max(max(max(170.0, -(gy - 10.0 - 3.25) / -1.0 + 3.25), -(-0.0 - gy) + 22.7), -(-0.0 - gy) + 22.7), -(gy - 10.0 - 6.5) / -1.0 + 3.5), 233.0);
  let gsl_partBox_150 = gsl_boxDistance(p, gsl_partLo_150, gsl_partHi_150);
  var gsl_part_150 = gsl_partBox_150;
  let gsl_partLo_151 = vec3<f32>(-226.0, -170.0, 200.0);
  let gsl_partHi_151 = vec3<f32>(-214.0, 170.0, 208.0);
  let gsl_partBox_151 = gsl_boxDistance(p, gsl_partLo_151, gsl_partHi_151);
  var gsl_part_151 = gsl_partBox_151;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_151 <= gsl_vanishView.w) {
    let gsl_partVanish_151 = gsl_vanish(gsl_partLo_151, gsl_partHi_151);
    gsl_part_151 = 1e10;
    if (gsl_partVanish_151.x > 0.0) { gsl_part_151 = rail(p.x + 220.0, p.z - 200.0, p.y + 170.0, 340.0, 12.0, 8.0, 25.0, noiseTex) + gsl_partVanish_151.y; }
    if (abs(gsl_part_151) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_151), gsl_partVanish_151.zw); }
  }
  let railL = gsl_part_151;
  let gsl_partLo_152 = vec3<f32>(214.0, -170.0, 200.0);
  let gsl_partHi_152 = vec3<f32>(226.0, 170.0, 208.0);
  let gsl_partBox_152 = gsl_boxDistance(p, gsl_partLo_152, gsl_partHi_152);
  var gsl_part_152 = gsl_partBox_152;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_152 <= gsl_vanishView.w) {
    let gsl_partVanish_152 = gsl_vanish(gsl_partLo_152, gsl_partHi_152);
    gsl_part_152 = 1e10;
    if (gsl_partVanish_152.x > 0.0) { gsl_part_152 = rail(p.x - 220.0, p.z - 200.0, p.y + 170.0, 340.0, 12.0, 8.0, 25.0, noiseTex) + gsl_partVanish_152.y; }
    if (abs(gsl_part_152) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_152), gsl_partVanish_152.zw); }
  }
  let railR = gsl_part_152;
  let gsl_partLo_153 = vec3<f32>(-175.0, -(gy - 10.0 - 3.25) / -1.0 - 3.25, 218.5);
  let gsl_partHi_153 = vec3<f32>(175.0, -(gy - 10.0 - 3.25) / -1.0 + 3.25, 227.5);
  let gsl_partBox_153 = gsl_boxDistance(p, gsl_partLo_153, gsl_partHi_153);
  var gsl_part_153 = gsl_partBox_153;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_153 <= gsl_vanishView.w) {
    let gsl_partVanish_153 = gsl_vanish(gsl_partLo_153, gsl_partHi_153);
    gsl_part_153 = 1e10;
    if (gsl_partVanish_153.x > 0.0) { gsl_part_153 = rail(p.z - 223.0, gy - 10.0 - p.y, p.x + 175.0, 350.0, 9.0, 6.5, 20.0, noiseTex) + gsl_partVanish_153.y; }
    if (abs(gsl_part_153) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_153), gsl_partVanish_153.zw); }
  }
  let railX = gsl_part_153;
  let gsl_partLo_154 = vec3<f32>(-233.5, -(-0.0 - gy) - 22.7, 203.0);
  let gsl_partHi_154 = vec3<f32>(-206.5, -(-0.0 - gy) + 22.7, 213.0);
  let gsl_partBox_154 = gsl_boxDistance(p, gsl_partLo_154, gsl_partHi_154);
  var gsl_part_154 = gsl_partBox_154;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_154 <= gsl_vanishView.w) {
    let gsl_partVanish_154 = gsl_vanish(gsl_partLo_154, gsl_partHi_154);
    gsl_part_154 = 1e10;
    if (gsl_partVanish_154.x > 0.0) { gsl_part_154 = carriage(p.x + 220.0, p.z - 200.0, p.y - gy, 27.0, 13.0, 45.4, 12.0, 8.0, noiseTex) + gsl_partVanish_154.y; }
    if (abs(gsl_part_154) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_154), gsl_partVanish_154.zw); }
  }
  let carriageL = gsl_part_154;
  let gsl_partLo_155 = vec3<f32>(206.5, -(-0.0 - gy) - 22.7, 203.0);
  let gsl_partHi_155 = vec3<f32>(233.5, -(-0.0 - gy) + 22.7, 213.0);
  let gsl_partBox_155 = gsl_boxDistance(p, gsl_partLo_155, gsl_partHi_155);
  var gsl_part_155 = gsl_partBox_155;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_155 <= gsl_vanishView.w) {
    let gsl_partVanish_155 = gsl_vanish(gsl_partLo_155, gsl_partHi_155);
    gsl_part_155 = 1e10;
    if (gsl_partVanish_155.x > 0.0) { gsl_part_155 = carriage(p.x - 220.0, p.z - 200.0, p.y - gy, 27.0, 13.0, 45.4, 12.0, 8.0, noiseTex) + gsl_partVanish_155.y; }
    if (abs(gsl_part_155) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_155), gsl_partVanish_155.zw); }
  }
  let carriageR = gsl_part_155;
  let gsl_partLo_156 = vec3<f32>(-(-0.0 - gx) - 19.95, -(gy - 10.0 - 6.5) / -1.0 - 3.5, 213.0);
  let gsl_partHi_156 = vec3<f32>(-(-0.0 - gx) + 19.95, -(gy - 10.0 - 6.5) / -1.0 + 3.5, 233.0);
  let gsl_partBox_156 = gsl_boxDistance(p, gsl_partLo_156, gsl_partHi_156);
  var gsl_part_156 = gsl_partBox_156;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_156 <= gsl_vanishView.w) {
    let gsl_partVanish_156 = gsl_vanish(gsl_partLo_156, gsl_partHi_156);
    gsl_part_156 = 1e10;
    if (gsl_partVanish_156.x > 0.0) { gsl_part_156 = carriage(p.z - 223.0, gy - 10.0 - p.y, p.x - gx, 20.0, 10.0, 39.9, 9.0, 6.5, noiseTex) + gsl_partVanish_156.y; }
    if (abs(gsl_part_156) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_156), gsl_partVanish_156.zw); }
  }
  let carriageT = gsl_part_156;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_150 <= gsl_vanishView.w) {
    let gsl_partVanish_150 = gsl_vanish(gsl_partLo_150, gsl_partHi_150);
    gsl_part_150 = 1e10;
    if (gsl_partVanish_150.x > 0.0) { gsl_part_150 = min(min(min(min(min(railL, railR), railX), carriageL), carriageR), carriageT) + gsl_partVanish_150.y; }
    if (abs(gsl_part_150) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_150), gsl_partVanish_150.zw); }
  }
  let rails = gsl_part_150;
  let gsl_partLo_157 = vec3<f32>(-221.15, -206.15, -238.0);
  let gsl_partHi_157 = vec3<f32>(221.15, 221.15, 298.0);
  let gsl_partBox_157 = gsl_boxDistance(p, gsl_partLo_157, gsl_partHi_157);
  var gsl_part_157 = gsl_partBox_157;
  let gsl_partLo_158 = vec3<f32>(-221.15, 178.85, 236.0);
  let gsl_partHi_158 = vec3<f32>(-178.85, 221.15, 298.0);
  let gsl_partBox_158 = gsl_boxDistance(p, gsl_partLo_158, gsl_partHi_158);
  var gsl_part_158 = gsl_partBox_158;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_158 <= gsl_vanishView.w) {
    let gsl_partVanish_158 = gsl_vanish(gsl_partLo_158, gsl_partHi_158);
    gsl_part_158 = 1e10;
    if (gsl_partVanish_158.x > 0.0) { gsl_part_158 = gsl_remap_4(p - vec3<f32>(-200.0, 200.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_158.y; }
    if (abs(gsl_part_158) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_158), gsl_partVanish_158.zw); }
  }
  let motorCoreA = gsl_part_158;
  let gsl_partLo_159 = vec3<f32>(178.85, 178.85, 236.0);
  let gsl_partHi_159 = vec3<f32>(221.15, 221.15, 298.0);
  let gsl_partBox_159 = gsl_boxDistance(p, gsl_partLo_159, gsl_partHi_159);
  var gsl_part_159 = gsl_partBox_159;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_159 <= gsl_vanishView.w) {
    let gsl_partVanish_159 = gsl_vanish(gsl_partLo_159, gsl_partHi_159);
    gsl_part_159 = 1e10;
    if (gsl_partVanish_159.x > 0.0) { gsl_part_159 = gsl_remap_4(p - vec3<f32>(200.0, 200.0, 260.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_159.y; }
    if (abs(gsl_part_159) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_159), gsl_partVanish_159.zw); }
  }
  let motorCoreB = gsl_part_159;
  let gsl_partLo_160 = vec3<f32>(-206.15, -206.15, -238.0);
  let gsl_partHi_160 = vec3<f32>(-163.85, -163.85, -176.0);
  let gsl_partBox_160 = gsl_boxDistance(p, gsl_partLo_160, gsl_partHi_160);
  var gsl_part_160 = gsl_partBox_160;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_160 <= gsl_vanishView.w) {
    let gsl_partVanish_160 = gsl_vanish(gsl_partLo_160, gsl_partHi_160);
    gsl_part_160 = 1e10;
    if (gsl_partVanish_160.x > 0.0) { gsl_part_160 = gsl_remap_3(p - vec3<f32>(-185.0, -185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_160.y; }
    if (abs(gsl_part_160) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_160), gsl_partVanish_160.zw); }
  }
  let zCoreFL = gsl_part_160;
  let gsl_partLo_161 = vec3<f32>(163.85, -206.15, -238.0);
  let gsl_partHi_161 = vec3<f32>(206.15, -163.85, -176.0);
  let gsl_partBox_161 = gsl_boxDistance(p, gsl_partLo_161, gsl_partHi_161);
  var gsl_part_161 = gsl_partBox_161;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_161 <= gsl_vanishView.w) {
    let gsl_partVanish_161 = gsl_vanish(gsl_partLo_161, gsl_partHi_161);
    gsl_part_161 = 1e10;
    if (gsl_partVanish_161.x > 0.0) { gsl_part_161 = gsl_remap_3(p - vec3<f32>(185.0, -185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_161.y; }
    if (abs(gsl_part_161) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_161), gsl_partVanish_161.zw); }
  }
  let zCoreFR = gsl_part_161;
  let gsl_partLo_162 = vec3<f32>(-21.15, 163.85, -238.0);
  let gsl_partHi_162 = vec3<f32>(21.15, 206.15, -176.0);
  let gsl_partBox_162 = gsl_boxDistance(p, gsl_partLo_162, gsl_partHi_162);
  var gsl_part_162 = gsl_partBox_162;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_162 <= gsl_vanishView.w) {
    let gsl_partVanish_162 = gsl_vanish(gsl_partLo_162, gsl_partHi_162);
    gsl_part_162 = 1e10;
    if (gsl_partVanish_162.x > 0.0) { gsl_part_162 = gsl_remap_3(p - vec3<f32>(0.0, 185.0, -200.0), gsl_param_48_13, gsl_param_51_11, gsl_param_54_13, noiseTex) + gsl_partVanish_162.y; }
    if (abs(gsl_part_162) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_162), gsl_partVanish_162.zw); }
  }
  let zCoreB = gsl_part_162;
  let gsl_partLo_163 = vec3<f32>(-189.0, -189.0, -172.0);
  let gsl_partHi_163 = vec3<f32>(-181.0, -181.0, 170.0);
  let gsl_partBox_163 = gsl_boxDistance(p, gsl_partLo_163, gsl_partHi_163);
  var gsl_part_163 = gsl_partBox_163;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_163 <= gsl_vanishView.w) {
    let gsl_partVanish_163 = gsl_vanish(gsl_partLo_163, gsl_partHi_163);
    gsl_part_163 = 1e10;
    if (gsl_partVanish_163.x > 0.0) { gsl_part_163 = leadScrew(p, -185.0, -185.0, noiseTex) + gsl_partVanish_163.y; }
    if (abs(gsl_part_163) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_163), gsl_partVanish_163.zw); }
  }
  let screwFL = gsl_part_163;
  let gsl_partLo_164 = vec3<f32>(181.0, -189.0, -172.0);
  let gsl_partHi_164 = vec3<f32>(189.0, -181.0, 170.0);
  let gsl_partBox_164 = gsl_boxDistance(p, gsl_partLo_164, gsl_partHi_164);
  var gsl_part_164 = gsl_partBox_164;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_164 <= gsl_vanishView.w) {
    let gsl_partVanish_164 = gsl_vanish(gsl_partLo_164, gsl_partHi_164);
    gsl_part_164 = 1e10;
    if (gsl_partVanish_164.x > 0.0) { gsl_part_164 = leadScrew(p, 185.0, -185.0, noiseTex) + gsl_partVanish_164.y; }
    if (abs(gsl_part_164) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_164), gsl_partVanish_164.zw); }
  }
  let screwFR = gsl_part_164;
  let gsl_partLo_165 = vec3<f32>(-4.0, 181.0, -172.0);
  let gsl_partHi_165 = vec3<f32>(4.0, 189.0, 170.0);
  let gsl_partBox_165 = gsl_boxDistance(p, gsl_partLo_165, gsl_partHi_165);
  var gsl_part_165 = gsl_partBox_165;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_165 <= gsl_vanishView.w) {
    let gsl_partVanish_165 = gsl_vanish(gsl_partLo_165, gsl_partHi_165);
    gsl_part_165 = 1e10;
    if (gsl_partVanish_165.x > 0.0) { gsl_part_165 = leadScrew(p, 0.0, 185.0, noiseTex) + gsl_partVanish_165.y; }
    if (abs(gsl_part_165) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_165), gsl_partVanish_165.zw); }
  }
  let screwB = gsl_part_165;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_157 <= gsl_vanishView.w) {
    let gsl_partVanish_157 = gsl_vanish(gsl_partLo_157, gsl_partHi_157);
    gsl_part_157 = 1e10;
    if (gsl_partVanish_157.x > 0.0) { gsl_part_157 = min(min(min(min(min(min(min(motorCoreA, motorCoreB), zCoreFL), zCoreFR), zCoreB), screwFL), screwFR), screwB) + gsl_partVanish_157.y; }
    if (abs(gsl_part_157) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_157), gsl_partVanish_157.zw); }
  }
  let motorSteel = gsl_part_157;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_149 <= gsl_vanishView.w) {
    let gsl_partVanish_149 = gsl_vanish(gsl_partLo_149, gsl_partHi_149);
    gsl_part_149 = 1e10;
    if (gsl_partVanish_149.x > 0.0) { gsl_part_149 = min(rails, motorSteel) + gsl_partVanish_149.y; }
    if (abs(gsl_part_149) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_149), gsl_partVanish_149.zw); }
  }
  let steel = gsl_part_149;
  let gsl_partLo_166 = vec3<f32>(min(min(min(min(min(min(gx, gx) - 4.0, min(gx, gx) - 3.0), -196.0), 174.0), -11.0), 3.5), min(min(min(min(min(min(headY, headY) - 4.0, min(headY, headY) - 3.0), -196.0), -196.0), 174.0), -161.5), min(min(min(min(115.0, min(min(bz - 28.0, bz - 24.5) - 11.0, min(bz - 43.0, bz - 28.0) - 5.1)), min(min(bz - 28.0, bz - 24.5) - 11.0, min(bz - 43.0, bz - 28.0) - 5.1)), min(min(bz - 28.0, bz - 24.5) - 11.0, min(bz - 43.0, bz - 28.0) - 5.1)), -279.5));
  let gsl_partHi_166 = vec3<f32>(max(max(max(max(max(max(gx, gx) + 4.0, max(gx, gx) + 3.0), -174.0), 196.0), 11.0), 156.5), max(max(max(max(max(max(headY, headY) + 4.0, max(headY, headY) + 3.0), -174.0), -174.0), 196.0), -68.5), max(max(max(max(130.0, max(max(bz - 28.0, bz - 24.5) + 11.0, max(bz - 43.0, bz - 28.0) + 5.1)), max(max(bz - 28.0, bz - 24.5) + 11.0, max(bz - 43.0, bz - 28.0) + 5.1)), max(max(bz - 28.0, bz - 24.5) + 11.0, max(bz - 43.0, bz - 28.0) + 5.1)), -264.5));
  let gsl_partBox_166 = gsl_boxDistance(p, gsl_partLo_166, gsl_partHi_166);
  var gsl_part_166 = gsl_partBox_166;
  let gsl_partLo_167 = vec3<f32>(min(min(gx, gx) - 4.0, min(gx, gx) - 3.0), min(min(headY, headY) - 4.0, min(headY, headY) - 3.0), 115.0);
  let gsl_partHi_167 = vec3<f32>(max(max(gx, gx) + 4.0, max(gx, gx) + 3.0), max(max(headY, headY) + 4.0, max(headY, headY) + 3.0), 130.0);
  let gsl_partBox_167 = gsl_boxDistance(p, gsl_partLo_167, gsl_partHi_167);
  var gsl_part_167 = gsl_partBox_167;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_167 <= gsl_vanishView.w) {
    let gsl_partVanish_167 = gsl_vanish(gsl_partLo_167, gsl_partHi_167);
    gsl_part_167 = 1e10;
    if (gsl_partVanish_167.x > 0.0) { gsl_part_167 = min(builtin_cylinderCapped(p, vec3<f32>(gx, headY, 122.0), vec3<f32>(gx, headY, 126.0), 4.0, noiseTex), builtin_coneCapped(p, vec3<f32>(gx, headY, 122.0), vec3<f32>(gx, headY, 118.0), 3.0, 0.5, noiseTex)) + gsl_partVanish_167.y; }
    if (abs(gsl_part_167) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_167), gsl_partVanish_167.zw); }
  }
  let nozzle = gsl_part_167;
  let gsl_partLo_168 = vec3<f32>(-196.0, -196.0, min(min(bz - 28.0, bz - 24.5) - 11.0, min(bz - 43.0, bz - 28.0) - 5.1));
  let gsl_partHi_168 = vec3<f32>(-174.0, -174.0, max(max(bz - 28.0, bz - 24.5) + 11.0, max(bz - 43.0, bz - 28.0) + 5.1));
  let gsl_partBox_168 = gsl_boxDistance(p, gsl_partLo_168, gsl_partHi_168);
  var gsl_part_168 = gsl_partBox_168;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_168 <= gsl_vanishView.w) {
    let gsl_partVanish_168 = gsl_vanish(gsl_partLo_168, gsl_partHi_168);
    gsl_part_168 = 1e10;
    if (gsl_partVanish_168.x > 0.0) { gsl_part_168 = leadNut(p, bz, -185.0, -185.0, noiseTex) + gsl_partVanish_168.y; }
    if (abs(gsl_part_168) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_168), gsl_partVanish_168.zw); }
  }
  let nutFL = gsl_part_168;
  let gsl_partLo_169 = vec3<f32>(174.0, -196.0, min(min(bz - 28.0, bz - 24.5) - 11.0, min(bz - 43.0, bz - 28.0) - 5.1));
  let gsl_partHi_169 = vec3<f32>(196.0, -174.0, max(max(bz - 28.0, bz - 24.5) + 11.0, max(bz - 43.0, bz - 28.0) + 5.1));
  let gsl_partBox_169 = gsl_boxDistance(p, gsl_partLo_169, gsl_partHi_169);
  var gsl_part_169 = gsl_partBox_169;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_169 <= gsl_vanishView.w) {
    let gsl_partVanish_169 = gsl_vanish(gsl_partLo_169, gsl_partHi_169);
    gsl_part_169 = 1e10;
    if (gsl_partVanish_169.x > 0.0) { gsl_part_169 = leadNut(p, bz, 185.0, -185.0, noiseTex) + gsl_partVanish_169.y; }
    if (abs(gsl_part_169) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_169), gsl_partVanish_169.zw); }
  }
  let nutFR = gsl_part_169;
  let gsl_partLo_170 = vec3<f32>(-11.0, 174.0, min(min(bz - 28.0, bz - 24.5) - 11.0, min(bz - 43.0, bz - 28.0) - 5.1));
  let gsl_partHi_170 = vec3<f32>(11.0, 196.0, max(max(bz - 28.0, bz - 24.5) + 11.0, max(bz - 43.0, bz - 28.0) + 5.1));
  let gsl_partBox_170 = gsl_boxDistance(p, gsl_partLo_170, gsl_partHi_170);
  var gsl_part_170 = gsl_partBox_170;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_170 <= gsl_vanishView.w) {
    let gsl_partVanish_170 = gsl_vanish(gsl_partLo_170, gsl_partHi_170);
    gsl_part_170 = 1e10;
    if (gsl_partVanish_170.x > 0.0) { gsl_part_170 = leadNut(p, bz, 0.0, 185.0, noiseTex) + gsl_partVanish_170.y; }
    if (abs(gsl_part_170) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_170), gsl_partVanish_170.zw); }
  }
  let nutB = gsl_part_170;
  let gsl_partLo_171 = vec3<f32>(3.5, -161.5, -279.5);
  let gsl_partHi_171 = vec3<f32>(156.5, -68.5, -264.5);
  let gsl_partBox_171 = gsl_boxDistance(p, gsl_partLo_171, gsl_partHi_171);
  var gsl_part_171 = gsl_partBox_171;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_171 <= gsl_vanishView.w) {
    let gsl_partVanish_171 = gsl_vanish(gsl_partLo_171, gsl_partHi_171);
    gsl_part_171 = 1e10;
    if (gsl_partVanish_171.x > 0.0) { gsl_part_171 = min(min(min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(6.0, -159.0, -277.0), vec3<f32>(6.0, -159.0, -267.0), 2.5, noiseTex), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(154.0, -159.0, -277.0), vec3<f32>(154.0, -159.0, -267.0), 2.5, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(6.0, -71.0, -277.0), vec3<f32>(6.0, -71.0, -267.0), 2.5, noiseTex)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(154.0, -71.0, -277.0), vec3<f32>(154.0, -71.0, -267.0), 2.5, noiseTex)) + gsl_partVanish_171.y; }
    if (abs(gsl_part_171) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_171), gsl_partVanish_171.zw); }
  }
  let standoffs = gsl_part_171;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_166 <= gsl_vanishView.w) {
    let gsl_partVanish_166 = gsl_vanish(gsl_partLo_166, gsl_partHi_166);
    gsl_part_166 = 1e10;
    if (gsl_partVanish_166.x > 0.0) { gsl_part_166 = min(min(min(min(nozzle, nutFL), nutFR), nutB), standoffs) + gsl_partVanish_166.y; }
    if (abs(gsl_part_166) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_166), gsl_partVanish_166.zw); }
  }
  let brassParts = gsl_part_166;
  let gsl_partLo_172 = vec3<f32>(0.0, -165.0, -267.0);
  let gsl_partHi_172 = vec3<f32>(160.0, -65.0, -265.4);
  let gsl_partBox_172 = gsl_boxDistance(p, gsl_partLo_172, gsl_partHi_172);
  var gsl_part_172 = gsl_partBox_172;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_172 <= gsl_vanishView.w) {
    let gsl_partVanish_172 = gsl_vanish(gsl_partLo_172, gsl_partHi_172);
    gsl_part_172 = 1e10;
    if (gsl_partVanish_172.x > 0.0) { gsl_part_172 = builtin_gcl_closure_boxCentered(p, vec3<f32>(160.0, -230.0, -532.4) * 0.5, abs(vec3<f32>(160.0, 100.0, 1.6)), noiseTex) + gsl_partVanish_172.y; }
    if (abs(gsl_part_172) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_172), gsl_partVanish_172.zw); }
  }
  let pcb = gsl_part_172;
  let gsl_partLo_173 = vec3<f32>(-155.0, -155.0, min(bz - 1.0, bz));
  let gsl_partHi_173 = vec3<f32>(155.0, 155.0, max(bz - 1.0, bz));
  let gsl_partBox_173 = gsl_boxDistance(p, gsl_partLo_173, gsl_partHi_173);
  var gsl_part_173 = gsl_partBox_173;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_173 <= gsl_vanishView.w) {
    let gsl_partVanish_173 = gsl_vanish(gsl_partLo_173, gsl_partHi_173);
    gsl_part_173 = 1e10;
    if (gsl_partVanish_173.x > 0.0) { gsl_part_173 = builtin_box(p, vec3<f32>(-155.0, -155.0, bz - 1.0), vec3<f32>(155.0, 155.0, bz), noiseTex) + gsl_partVanish_173.y; }
    if (abs(gsl_part_173) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_173), gsl_partVanish_173.zw); }
  }
  let bedSheet = gsl_part_173;
  let gsl_partLo_174 = vec3<f32>(min(min(gx - 18.0, gx + 18.0), -38.0), min(min(gy - 75.6, gy - 74.8), -250.6), -256.0);
  let gsl_partHi_174 = vec3<f32>(max(max(gx - 18.0, gx + 18.0), 22.0), max(max(gy - 75.6, gy - 74.8), -249.9), 170.0);
  let gsl_partBox_174 = gsl_boxDistance(p, gsl_partLo_174, gsl_partHi_174);
  var gsl_part_174 = gsl_partBox_174;
  let gsl_partLo_175 = vec3<f32>(min(gx - 18.0, gx + 18.0), min(gy - 75.6, gy - 74.8), 166.0);
  let gsl_partHi_175 = vec3<f32>(max(gx - 18.0, gx + 18.0), max(gy - 75.6, gy - 74.8), 170.0);
  let gsl_partBox_175 = gsl_boxDistance(p, gsl_partLo_175, gsl_partHi_175);
  var gsl_part_175 = gsl_partBox_175;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_175 <= gsl_vanishView.w) {
    let gsl_partVanish_175 = gsl_vanish(gsl_partLo_175, gsl_partHi_175);
    gsl_part_175 = 1e10;
    if (gsl_partVanish_175.x > 0.0) { gsl_part_175 = builtin_box(p, vec3<f32>(gx - 18.0, gy - 75.6, 166.0), vec3<f32>(gx + 18.0, gy - 74.8, 170.0), noiseTex) + gsl_partVanish_175.y; }
    if (abs(gsl_part_175) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_175), gsl_partVanish_175.zw); }
  }
  let ledBar = gsl_part_175;
  let gsl_partLo_176 = vec3<f32>(-38.0, -250.6, -256.0);
  let gsl_partHi_176 = vec3<f32>(22.0, -249.9, -224.0);
  let gsl_partBox_176 = gsl_boxDistance(p, gsl_partLo_176, gsl_partHi_176);
  var gsl_part_176 = gsl_partBox_176;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_176 <= gsl_vanishView.w) {
    let gsl_partVanish_176 = gsl_vanish(gsl_partLo_176, gsl_partHi_176);
    gsl_part_176 = 1e10;
    if (gsl_partVanish_176.x > 0.0) { gsl_part_176 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-16.0, -500.5, -480.0) * 0.5, abs(vec3<f32>(60.0, 0.7, 32.0)), noiseTex) + gsl_partVanish_176.y; }
    if (abs(gsl_part_176) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_176), gsl_partVanish_176.zw); }
  }
  let displayScreen = gsl_part_176;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_174 <= gsl_vanishView.w) {
    let gsl_partVanish_174 = gsl_vanish(gsl_partLo_174, gsl_partHi_174);
    gsl_part_174 = 1e10;
    if (gsl_partVanish_174.x > 0.0) { gsl_part_174 = min(ledBar, displayScreen) + gsl_partVanish_174.y; }
    if (abs(gsl_part_174) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_174), gsl_partVanish_174.zw); }
  }
  let lights = gsl_part_174;
  let gsl_s1 = gsl_shape(aluminum, 1u);
  let gsl_s2 = gsl_shape(blackMetal, 2u);
  let gsl_s3 = gsl_shape(darkParts, 3u);
  let gsl_s4 = gsl_shape(red, 4u);
  let gsl_s5 = gsl_shape(fasteners, 5u);
  let gsl_s6 = gsl_shape(steel, 6u);
  let gsl_s7 = gsl_shape(brassParts, 7u);
  let gsl_s8 = gsl_shape(pcb, 8u);
  let gsl_s9 = gsl_shape(bedSheet, 9u);
  let gsl_s10 = gsl_shape(lights, 10u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.05, 0.05, 0.06), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.06, 0.06, 0.07), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.8, 0.07, 0.05), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.12, 0.12, 0.13), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(0.9593, 0.8952, 0.6822), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(vec3<f32>(0.05, 0.32, 0.14), 1.0); }
  if (gsl_colorIdx == 9u) {
    let speck = noise_gradientLattice(p / 0.5, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.1875, 0.0625, noiseTex));
    return vec4<f32>(vec3<f32>(0.8, 0.58, 0.24) * (1.0 + 0.2 * speck), 1.0);
  }
  if (gsl_colorIdx == 10u) { return vec4<f32>(vec3<f32>(0.45, 0.72, 1.0), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_48_13: f32, gsl_param_51_11: f32, gsl_param_54_13: f32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = mat4x4<f32>(vec4<f32>(0.6, 0.0, 1.46, 0.0), material_plastic[1], material_plastic[2], material_plastic[3]); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = mat4x4<f32>(vec4<f32>(0.55, 0.0, 1.46, 0.0), material_plastic[1], material_plastic[2], material_plastic[3]); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.3, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.2, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.25, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = mat4x4<f32>(vec4<f32>(0.4, 0.0, 1.46, 0.0), material_plastic[1], material_plastic[2], material_plastic[3]); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 9u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let speck = noise_gradientLattice(p / 0.5, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.1875, 0.0625, noiseTex));
    let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.75, 0.0, 1.46, 0.0), material_plastic[1], material_plastic[2], material_plastic[3]), 0.05 * speck, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 10u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.45, 0.72, 1.0, 3.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
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

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));
