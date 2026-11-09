fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn builtin_arcExtendedXY(p: vec3<f32>, halfAngleSinCos: vec2<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pxAbs = abs(p.x);
  let py = p.y;
  let l = length(vec2<f32>(pxAbs, py));
  let dOutside = l - radius;
  let dInside = pxAbs * halfAngleSinCos.x + py * halfAngleSinCos.y - radius;
  let inside = halfAngleSinCos.y * pxAbs > halfAngleSinCos.x * py;
  return select(dOutside, dInside, inside);
}

fn builtin_circleXY(p: vec3<f32>, center: vec2<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p.xy - center) - radius;
}

fn gsl_remap(p: vec3<f32>, gsl_param_14_7: f32, gsl_param_18_7: f32, gsl_param_22_7: f32, gsl_param_26_7: f32, gsl_param_49_13: f32, vector_zero: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return 1.025 - abs(p.x);
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_14_7: f32, gsl_param_18_7: f32, gsl_param_22_7: f32, gsl_param_26_7: f32, gsl_param_49_13: f32, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_plane(vec3<f32>(abs(p.x), p.y, p.z), vec3<f32>(4.7, 7.3000746, -2.6998363), vec3<f32>(0.70710677, 0.0, 0.70710677), noiseTex);
}

fn builtin_intersectChamfer(d1: f32, d2: f32, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2 + setback) / 1.4142135;
  return max(face, edge);
}

fn gsl_remap_3(p: vec3<f32>, gsl_param_14_7: f32, gsl_param_18_7: f32, gsl_param_22_7: f32, gsl_param_26_7: f32, gsl_param_49_13: f32, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return abs(p.x) - 1.45;
}

fn gsl_remap_4(p: vec3<f32>, gsl_param_14_7: f32, gsl_param_18_7: f32, gsl_param_22_7: f32, gsl_param_26_7: f32, gsl_param_49_13: f32, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return 0.6 - abs(p.x);
}

fn builtin_quadXY(p: vec3<f32>, corner0: vec2<f32>, corner1: vec2<f32>, corner2: vec2<f32>, corner3: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let q = p.xy;
  let w0 = q - corner0;
  let e0 = corner1 - corner0;
  let t0 = clamp(dot(w0, e0) / dot(e0, e0), 0.0, 1.0);
  let qq0 = w0 - e0 * t0;
  let d0 = dot(qq0, qq0);
  let w1 = q - corner1;
  let e1 = corner2 - corner1;
  let t1 = clamp(dot(w1, e1) / dot(e1, e1), 0.0, 1.0);
  let qq1 = w1 - e1 * t1;
  let d1 = dot(qq1, qq1);
  let w2 = q - corner2;
  let e2 = corner3 - corner2;
  let t2 = clamp(dot(w2, e2) / dot(e2, e2), 0.0, 1.0);
  let qq2 = w2 - e2 * t2;
  let d2 = dot(qq2, qq2);
  let w3 = q - corner3;
  let e3 = corner0 - corner3;
  let t3 = clamp(dot(w3, e3) / dot(e3, e3), 0.0, 1.0);
  let qq3 = w3 - e3 * t3;
  let d3 = dot(qq3, qq3);
  let dmin = min(min(min(d0, d1), d2), d3);
  let gs = (corner0.x - corner3.x) * (corner1.y - corner0.y) - (corner0.y - corner3.y) * (corner1.x - corner0.x);
  let s0 = gs * (w0.x * e0.y - w0.y * e0.x);
  let s1 = gs * (w1.x * e1.y - w1.y * e1.x);
  let s2 = gs * (w2.x * e2.y - w2.y * e2.x);
  let s3 = gs * (w3.x * e3.y - w3.y * e3.x);
  let smax = max(max(max(s0, s1), s2), s3);
  return sqrt(dmin) * sign(smax);
}

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn builtin_cylinder(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_axis(p, center, direction, noiseTex) - radius;
}

fn builtin_cylinderZ(p: vec3<f32>, vector_dirZ: vec3<f32>, center: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_cylinder(p, center, vector_dirZ, radius, noiseTex);
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

fn noise_cellularF1Lattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_cellularScan(q, 1, noiseTex).f1;
}

fn noise_cellular(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_cellularF1Lattice(p / size, noiseTex) * size;
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn noise_cellularF1F2CellLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  gsl_result.v_3 = gsl_cellularScan(q, 1, noiseTex).nearCell;
  return gsl_result;
}

fn noise_cellularCell(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  let gsl_result_2 = noise_cellularF1F2CellLattice(p / size, noiseTex);
  let f1 = gsl_result_2.v_1;
  let f2 = gsl_result_2.v_2;
  let cell = gsl_result_2.v_3;
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = f1 * size;
  gsl_result.v_2 = f2 * size;
  gsl_result.v_3 = cell;
  return gsl_result;
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

fn noise_octaveFade(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - ease_quadratic(clamp((gsl_pw - 0.375 * size) / (0.125 * size) * 0.5 + 0.5, 0.0, 1.0), noiseTex);
}

fn noise_warp(p: vec3<f32>, size: f32, amount: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let coarse = noise_gradientLatticeVector(p / size, noiseTex);
  let fine = noise_gradientLatticeVector(p * 2.03 / size + vec3<f32>(17.3, 5.1, 9.7), noiseTex);
  return p + (coarse + fine * 0.5 * noise_octaveFade(0.5 * size, noiseTex)) * amount * 1.2;
}

fn stone_grainAt(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_3 {
  let f1 = gsl_cellularScan(q, 1, noiseTex).f1;
  let f2 = gsl_cellularScan(q, 1, noiseTex).f2;
  let far = gsl_cellularScan(q, 1, noiseTex).farCell;
  let near = gsl_cellularScan(q, 1, noiseTex).nearCell;
  let gap = length(far + fract(textureLoad(noiseTex, vec3<i32>(floor(far)) & vec3<i32>(63), 0).rgb) - near - fract(textureLoad(noiseTex, vec3<i32>(floor(near)) & vec3<i32>(63), 0).rgb));
  var gsl_result : gsl_struct_3;
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

fn stone_graniteGrains(p: vec3<f32>, s: f32, noiseTex: texture_3d<f32>) -> gsl_struct_4 {
  let w = noise_warp(p, 1.3 * s, 0.22 * s, noiseTex);
  let g = 0.45 * s;
  let sq = (w + noise_warp(p, 0.5 * s, 0.08 * s, noiseTex) - p + vec3<f32>(37.1, 11.3, 5.9) * s) / g;
  let gsl_result_5 = stone_grainAt(sq, noiseTex);
  let small = gsl_result_5.v_3;
  let sf1 = gsl_result_5.v_1;
  let swall = gsl_result_5.v_2;
  let sh = fract(textureLoad(noiseTex, vec3<i32>(floor(small + vec3<f32>(7.0, 29.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((small + vec3<f32>(7.0, 29.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let quartzIn = stone_quartzOf(small, g, s, noiseTex);
  let sAcross = 2.0 * texture_mask(swall * g, 0.0, 0.05 * s, noiseTex) - 1.0;
  let smallFar = gsl_result_5.v_4;
  let inQuartz = quartzIn * (0.5 + 0.5 * sAcross) + stone_quartzOf(smallFar, g, s, noiseTex) * (0.5 - 0.5 * sAcross);
  let t = 0.15 * s;
  let gsl_result_6 = stone_grainAt((w + vec3<f32>(5.3, 61.7, 17.1) * s) / t, noiseTex);
  let tiny = gsl_result_6.v_3;
  let twall = gsl_result_6.v_2;
  let tAcross = 2.0 * texture_mask(twall * t, 0.0, 0.05 * s, noiseTex) - 1.0;
  let tinyFar = gsl_result_6.v_4;
  let mafic = stone_maficOf(tiny, t, s, quartzIn, noiseTex) * (0.5 + 0.5 * tAcross) + stone_maficOf(tinyFar, t, s, quartzIn, noiseTex) * (0.5 - 0.5 * tAcross);
  let plag = select(0.0, 1.0 - quartzIn, sh.x < 0.15);
  let gsl_result_4 = stone_grainAt(w / s, noiseTex);
  let bwall = gsl_result_4.v_2;
  let big = gsl_result_4.v_3;
  let bh = fract(textureLoad(noiseTex, vec3<i32>(floor(big + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((big + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let th = fract(textureLoad(noiseTex, vec3<i32>(floor(tiny + vec3<f32>(3.0, 0.0, 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((tiny + vec3<f32>(3.0, 0.0, 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  var gsl_result : gsl_struct_4;
  gsl_result.v_1 = sq;
  gsl_result.v_2 = small;
  gsl_result.v_3 = sf1;
  gsl_result.v_4 = swall * g;
  gsl_result.v_5 = sh;
  gsl_result.v_6 = inQuartz;
  gsl_result.v_7 = mafic;
  gsl_result.v_8 = plag;
  gsl_result.v_9 = bwall * s;
  gsl_result.v_10 = bh;
  gsl_result.v_11 = th;
  return gsl_result;
}

fn stone_darkFloor(L: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 0.5 * (L + 0.3 + sqrt((L - 0.3) * (L - 0.3) + 0.0064));
}

fn stone_graniteMosaic(p: vec3<f32>, s: f32, baseColor: vec3<f32>, fleckColorA: vec3<f32>, fleckColorB: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_5 {
  let feldLab = color_rgb2oklab(baseColor, noiseTex);
  let gsl_result_7 = stone_graniteGrains(p, s, noiseTex);
  let bh = gsl_result_7.v_10;
  let cloud = noise_gradientLattice(p / (0.3 * s) + vec3<f32>(1.7, 9.1, 4.4), noiseTex) + 0.15 * texture_fadeDetail(0.12 * s, noiseTex) * noise_gradientLattice(p / (0.12 * s) + vec3<f32>(8.2, 2.6, 7.3), noiseTex) + 0.4 * noise_gradientLattice(p / (0.9 * s) + vec3<f32>(4.2, 6.6, 1.3), noiseTex);
  let dir = normalize(bh - 0.5 + vec3<f32>(0.0001, 0.0, 0.0));
  let along = dot(p, dir);
  let streak = noise_gradientLattice((p - dir * along) / (0.08 * s) + dir * along / (0.8 * s), noiseTex) * texture_fadeDetail(0.08 * s, noiseTex);
  let bwall = gsl_result_7.v_9;
  let rim = 1.0 - texture_mask(bwall, 0.06 * s, 0.05 * s, noiseTex);
  let frost = texture_fadeTo(abs(noise_gradientLattice(p / (0.15 * s) + vec3<f32>(6.1, 4.9, 2.2), noiseTex)), 0.157, 0.15 * s, noiseTex) + 0.5 * texture_fadeTo(abs(noise_gradientLattice(p / (0.06 * s) + vec3<f32>(1.1, 3.5, 8.8), noiseTex)), 0.157, 0.06 * s, noiseTex);
  let feldspar = feldLab + vec3<f32>((bh.y - 0.5) * 0.1 + 0.1 * cloud + 0.015 * streak + 0.04 * rim + 0.12 * (frost - 0.235) * (0.5 + rim), 0.0, 0.0) - vec3<f32>(0.0, feldLab.y, feldLab.z) * 0.3 * rim;
  let sh = gsl_result_7.v_5;
  let plagLab = feldLab + vec3<f32>(0.06 + (sh.y - 0.5) * 0.06 + 0.04 * cloud + 0.12 * (frost - 0.235), -0.5 * feldLab.y, -0.5 * feldLab.z);
  let plag = gsl_result_7.v_8;
  let ground = feldspar + (plagLab - feldspar) * plag;
  let quartzLab = color_rgb2oklab(fleckColorB, noiseTex);
  let glass = vec3<f32>(quartzLab.x + (0.5 - quartzLab.x) * 0.45, quartzLab.y * 0.8, quartzLab.z * 0.8);
  let sf1 = gsl_result_7.v_3;
  let core = 1.0 - texture_mask(sf1, 0.3, 0.25, noiseTex);
  let sq = gsl_result_7.v_1;
  let small = gsl_result_7.v_2;
  let lean = dot(sq - small - fract(textureLoad(noiseTex, vec3<i32>(floor(small)) & vec3<i32>(63), 0).rgb), normalize(sh - 0.5 + vec3<f32>(0.0, 0.0001, 0.0)));
  let swall = gsl_result_7.v_4;
  let crack = (1.0 - texture_mask(swall, 0.02 * s, 0.015 * s, noiseTex)) * texture_fadeDetail(0.05 * s, noiseTex);
  let quartz = glass + vec3<f32>((sh.z - 0.5) * 0.12 - 0.1 * core + 0.05 * lean + 0.01 * crack + 0.02 * cloud + 0.06 * (frost - 0.235), -0.003 * core, -0.01 * core + (sh.y - 0.5) * 0.02);
  let inQuartz = gsl_result_7.v_6;
  let broad = 0.15 * noise_gradientLattice(p / (3.0 * s) + vec3<f32>(2.9, 8.3, 6.1), noiseTex) + 0.08 * noise_gradientLattice(p / (1.4 * s) + vec3<f32>(9.4, 1.2, 3.8), noiseTex);
  let felsic = ground + (quartz - ground) * inQuartz + vec3<f32>(broad, 0.0, 0.0);
  let darkLab = color_rgb2oklab(fleckColorA, noiseTex);
  let th = gsl_result_7.v_11;
  let dark = vec3<f32>(stone_darkFloor(darkLab.x, noiseTex), darkLab.y, darkLab.z) + vec3<f32>((th.y - 0.5) * 0.06 + select(0.0, 0.12, th.z > 0.96), 0.0, 0.0) + select(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.03, 0.01, 0.02), th.z < 0.08);
  let mafic = gsl_result_7.v_7;
  let crystal = felsic + (dark - felsic) * mafic;
  let wPlag = (1.0 - mafic) * (1.0 - inQuartz) * plag;
  let wQuartz = (1.0 - mafic) * inQuartz;
  var gsl_result : gsl_struct_5;
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
    let gsl_result_8 = stone_graniteMosaic(p, cellSize, baseColor, fleckColorA, fleckColorB, noiseTex);
    let crystal = gsl_result_8.v_1;
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

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn texture_paintMaterial(field: f32, edge: f32, width: f32, matA: mat4x4<f32>, matB: mat4x4<f32>, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_blend(matA, matB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn map_rotateZX(v: vec3<f32>, ang: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let ca = cos(ang);
  let sa = sin(ang);
  return vec3<f32>(v.x * ca - v.z * sa, v.y, v.z * ca + v.x * sa);
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

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_grainRelief(p: vec3<f32>, size: f32, height: f32, noiseTex: texture_3d<f32>) -> f32 {
  var gsl_detailValue = 0.0;
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (size)) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let bump = clamp(1.0 - noise_cellular(p, size, noiseTex) / (0.45 * size), 0.0, 1.0);
    gsl_detailValue = bump * bump * height * texture_fadeDetail(size, noiseTex);
    gsl_pw = gsl_detailOuterPw;
  }
  let relief = gsl_detailValue;
  return relief;
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
}

fn texture_grainSpecks(p: vec3<f32>, size: f32, share: f32, noiseTex: texture_3d<f32>) -> f32 {
  var gsl_detailValue = share;
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (size)) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let gsl_result_3 = noise_cellularCell(p, size, noiseTex);
    let cell = gsl_result_3.v_3;
    let pick = fract(textureLoad(noiseTex, vec3<i32>(floor(cell + vec3<f32>(23.0, 61.0, 7.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell + vec3<f32>(23.0, 61.0, 7.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    gsl_detailValue = texture_fadeTo(select(0.0, 1.0, pick.x < share), share, size, noiseTex);
    gsl_pw = gsl_detailOuterPw;
  }
  let specks = gsl_detailValue;
  return specks;
}

fn natural_terrainSnowShare(snowBias: f32, noiseTex: texture_3d<f32>) -> f32 {
  let snow = clamp(0.01 + 0.09 * snowBias * snowBias, 0.0, 1.0);
  return clamp((snow - 0.01) / 0.99, 0.0, 1.0);
}

fn natural_terrainStandColor(snowBias: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return color_oklab2rgb(vec3<f32>(0.3796, -0.0065, -0.0341) + (vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex)))) - vec3<f32>(0.3796, -0.0065, -0.0341)) * natural_terrainSnowShare(snowBias, noiseTex), noiseTex);
}

fn natural_terrainStandMaterial(snowBias: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let snowMat = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.3098, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  return material_blend(mat4x4<f32>(vec4<f32>(0.274, 0.0, 1.39, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), snowMat, natural_terrainSnowShare(snowBias, noiseTex), noiseTex);
}

fn natural_terrain(p: vec3<f32>, scale: f32, snowBias: f32, noiseTex: texture_3d<f32>) -> gsl_struct_6 {
  let unit = scale / 12.0;
  let continents = noise_fbmBump4(p, scale, 0.5, 2.0, noiseTex) - 0.25 * unit;
  let ridges = noise_turbulenceBump2(p, scale / 4.0, 0.5, 2.0, noiseTex);
  let height = 1.2 * continents + 2.2 * (0.45 * unit - ridges) * clamp(continents / unit, 0.0, 1.0);
  let band = (height + 0.25 * unit * noise_gradient(p, scale / 4.0, noiseTex)) / unit;
  let seaLevel = height / unit;
  var gsl_simpleValue = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let measured = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_simpleFull = material_withDisplacement(measured, texture_grainRelief(p, 0.3, 0.12, noiseTex), noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  var gsl_simpleValue_2 = natural_terrainStandMaterial(snowBias, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let sandMat = gsl_simpleValue;
    let sandy = texture_paintMaterial(seaLevel, 0.0, 0.02, mat4x4<f32>(vec4<f32>(0.08, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), sandMat, noiseTex);
    let grassMat = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let grassy = texture_paintMaterial(band, 0.2, 0.12, sandy, grassMat, noiseTex);
    let stony = texture_paintMaterial(band, 1.2, 0.25, grassy, mat4x4<f32>(vec4<f32>(0.8, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
    let snowMat = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.3098, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_simpleFull_2 = texture_paintMaterial(band + snowBias, 2.0, 0.2, stony, snowMat, noiseTex);
    gsl_simpleValue_2 = mat4x4<f32>(mix(gsl_simpleValue_2[0], gsl_simpleFull_2[0], gsl_simpleDetail), mix(gsl_simpleValue_2[1], gsl_simpleFull_2[1], gsl_simpleDetail), mix(gsl_simpleValue_2[2], gsl_simpleFull_2[2], gsl_simpleDetail), mix(gsl_simpleValue_2[3], gsl_simpleFull_2[3], gsl_simpleDetail));
  }
  let landMat = gsl_simpleValue_2;
  var gsl_simpleValue_3 = vec3<f32>(0.6596159, 0.61961985, 0.49138752);
  if (gsl_simpleDetail > 0.0) {
    let measuredColor = vec3<f32>(0.6944, 0.6546, 0.5176);
    gsl_simpleValue_3 = mix(gsl_simpleValue_3, color_mixOklab(measuredColor, vec3<f32>(0.3, 0.26, 0.22), texture_grainSpecks(p, 0.3, 0.083, noiseTex), noiseTex), gsl_simpleDetail);
  }
  var gsl_simpleValue_4 = natural_terrainStandColor(snowBias, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let sea = vec3<f32>(0.49309835, -0.060483743, -0.05850419) + vec3<f32>(-0.2985687, 0.053410567, -0.009546458) * clamp(-seaLevel / 1.2, 0.0, 1.0);
    let sandColor = gsl_simpleValue_3;
    let coast = sea + (color_rgb2oklab(sandColor, noiseTex) - sea) * texture_mask(seaLevel, 0.0, 0.02, noiseTex);
    let grassLab = vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.10500539, 0.02767028, 0.0 > 0.0), select(0.13300854, 0.030975232, 0.0 > 0.0), select(0.04101066, 0.017321981, 0.0 > 0.0))), noiseTex))));
    let meadow = grassLab + (vec3<f32>(0.34344932, -0.06014222, 0.05241708) - grassLab) * texture_mask(noise_gradient(p, scale / 3.0, noiseTex), 0.0, 0.3, noiseTex);
    let green = coast + (meadow - coast) * texture_mask(band, 0.2, 0.12, noiseTex);
    let rock = color_rgb2oklab(stone_granite(p, 0.6 * unit, vec3<f32>(0.42, 0.4, 0.38), vec3<f32>(0.22, 0.21, 0.2), vec3<f32>(0.62, 0.6, 0.58), noiseTex), noiseTex);
    let rocky = green + (rock - green) * texture_mask(band, 1.2, 0.25, noiseTex);
    let land = rocky + (vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex))), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), vec3<f32>(color_cbrt(dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex), color_cbrt(dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), vec3<f32>(select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0), select(0.84994924, 0.072051086, 0.0 > 0.0))), noiseTex)))) - rocky) * texture_mask(band + snowBias, 2.0, 0.2, noiseTex);
    gsl_simpleValue_4 = mix(gsl_simpleValue_4, color_oklab2rgb(land, noiseTex), gsl_simpleDetail);
  }
  let landColor = gsl_simpleValue_4;
  var gsl_result : gsl_struct_6;
  gsl_result.v_1 = landMat;
  gsl_result.v_2 = landColor;
  gsl_result.v_3 = height;
  return gsl_result;
}

fn planeXY(p: vec3<f32>, p1: vec2<f32>, p2: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let dDir = p2 - p1;
  let n = normalize(vec2<f32>(-dDir.y, dDir.x));
  return builtin_plane(p, vec3<f32>(p1.x, p1.y, 0.0), vec3<f32>(n.x, n.y, 0.0), noiseTex);
}

fn arcShape(p: vec3<f32>, vector_zero: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  let d = p - vec3<f32>(0.0, 0.0, 0.0);
  let lx = dot(d, vec3<f32>(0.0, -0.629699, -0.77683926));
  let ly = dot(d, vec3<f32>(0.0, -0.77683926, 0.629699));
  let rx = lx * 0.9998388 - ly * 0.017953383;
  let ry = lx * 0.017953383 + ly * 0.9998388;
  let lz = dot(d, vec3<f32>(-1.0, 0.0, 0.0));
  let arc = builtin_arcExtendedXY(vec3<f32>(rx, ry, lz), vec2<f32>(0.9271787, 0.3746194), 1.4, noiseTex);
  let hole = builtin_circleXY(p.zyx, vec2<f32>(0.0, 0.0), 0.75, noiseTex);
  let horizPlane = -rx * 0.9271787 + ry * 0.3746194 - 1.4;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = max(arc, -hole);
  gsl_result.v_2 = horizPlane;
  return gsl_result;
}

fn gcl_closure_gsl_anon(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return p.z - -2.7;
}

fn boltHole(p: vec3<f32>, vector_dirZ: vec3<f32>, corner2D: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_cylinderZ(p, vector_dirZ, vec3<f32>(corner2D.x, corner2D.y, 0.0), 0.42, noiseTex);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_14_7: f32, gsl_param_18_7: f32, gsl_param_22_7: f32, gsl_param_26_7: f32, gsl_param_49_13: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let p0x = gsl_param_14_7;
  let quadV0X = -5.9 + p0x;
  let p0y = gsl_param_18_7;
  let quadV0Y = 1.5 + p0y;
  let v0 = vec2<f32>(quadV0X, quadV0Y);
  let p2x = gsl_param_22_7;
  let quadV2X = 5.9 + p2x;
  let p2y = gsl_param_26_7;
  let quadV2Y = 6.1 + p2y;
  let v2 = vec2<f32>(quadV2X, quadV2Y);
  let bottomPlane = -(p.z - -3.5);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let gsl_tmp_1 = p.z - -2.7;
  let gsl_tmp_2 = builtin_plane(vec3<f32>(abs(p.x), p.y, p.z), vec3<f32>(4.7, 7.3000746, -2.6998363), vec3<f32>(0.70710677, 0.0, 0.70710677), noiseTex);
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let unionTop = select(-length(vec2<f32>((gsl_tmp_1 + gsl_tmp_2) / 1.847759, (gsl_tmp_1 - gsl_tmp_2) / 0.76536685)), min(gsl_tmp_1, gsl_tmp_2), max(gcl_closure_gsl_anon(p - vec3<f32>(0.70710677, 0.0, 0.70710677) * gsl_tmp_2, noiseTex), gsl_remap_2(p - vector_dirZ * gsl_tmp_1, gsl_param_14_7, gsl_param_18_7, gsl_param_22_7, gsl_param_26_7, gsl_param_49_13, vector_zero, vector_dirZ, noiseTex)) > 0.0);
  let clevisAng = gsl_param_49_13;
  let clevisRot = map_rotateZX(p, radians(clevisAng), noiseTex);
  let gsl_result = arcShape(clevisRot, vector_zero, noiseTex);
  let horizPlane = gsl_result.v_2;
  let frontPlane = planeXY(p, vec2<f32>(5.9, 1.5), v0, noiseTex) - 1.2;
  let eyeSpan = gsl_remap(clevisRot, gsl_param_14_7, gsl_param_18_7, gsl_param_22_7, gsl_param_26_7, gsl_param_49_13, vector_zero, noiseTex);
  let eyeDraft = builtin_plane(clevisRot, vec3<f32>(1.45, 0.30003822, -1.3846282), vec3<f32>(0.0, -1.0, 1.7320508), noiseTex);
  let eyeStep = max(min(eyeSpan, eyeDraft + 0.8), eyeDraft);
  let topSurf = builtin_intersectChamfer(builtin_intersectChamfer(unionTop, horizPlane, 0.05, noiseTex), builtin_intersectChamfer(frontPlane, eyeStep, 0.05, noiseTex), 0.05, noiseTex);
  let base = max(builtin_intersectChamfer(max(max(builtin_quadXY(p, v0, vec2<f32>(-5.9, 6.1), v2, vec2<f32>(5.9, 1.5), noiseTex) - 1.2, bottomPlane), -min(min(min(boltHole(p, vector_dirZ, v0, noiseTex), builtin_axis(p, vec3<f32>(-5.9, 6.1, 0.0), vector_dirZ, noiseTex) - 0.42), boltHole(p, vector_dirZ, v2, noiseTex)), builtin_axis(p, vec3<f32>(5.9, 1.5, 0.0), vector_dirZ, noiseTex) - 0.42)), topSurf, 0.05, noiseTex), -topSurf - 0.8);
  let clevisShape = gsl_result.v_1;
  let clevisSpan = gsl_remap_3(clevisRot, gsl_param_14_7, gsl_param_18_7, gsl_param_22_7, gsl_param_26_7, gsl_param_49_13, vector_zero, vector_dirZ, noiseTex);
  let rearPlane = planeXY(p, vec2<f32>(-5.9, 6.1), v2, noiseTex) - 1.2;
  let clevis = max(max(max(builtin_intersectChamfer(clevisShape, clevisSpan, 0.05, noiseTex), bottomPlane), -topSurf), rearPlane);
  let clevisCut = max(min(-eyeStep - 0.8, rearPlane), gsl_remap_4(clevisRot, gsl_param_14_7, gsl_param_18_7, gsl_param_22_7, gsl_param_26_7, gsl_param_49_13, vector_zero, vector_dirZ, noiseTex));
  let clevisFinal = builtin_intersectChamfer(clevis, clevisCut, 0.05, noiseTex);
  let body = min(base, clevisFinal);
  let gsl_s1 = gsl_shape(body, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_14_7: f32, gsl_param_18_7: f32, gsl_param_22_7: f32, gsl_param_26_7: f32, gsl_param_49_13: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let altitude = 3.0 * clamp((p.z - 0.2) / 1.2, 0.0, 1.0);
    let gsl_result_9 = natural_terrain(p, 3.0, altitude, noiseTex);
    let terrainColor = gsl_result_9.v_2;
    return vec4<f32>(terrainColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_14_7: f32, gsl_param_18_7: f32, gsl_param_22_7: f32, gsl_param_26_7: f32, gsl_param_49_13: f32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let altitude = 3.0 * clamp((p.z - 0.2) / 1.2, 0.0, 1.0);
    let gsl_result_9 = natural_terrain(p, 3.0, altitude, noiseTex);
    let terrainMaterial = gsl_result_9.v_1;
    let height = gsl_result_9.v_3;
    let gsl_m = material_withDisplacement(terrainMaterial, 0.3 * max(height, 0.0), noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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
}
struct gsl_struct_2 {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
}
struct gsl_struct_3 {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
  v_4 : vec3<f32>,
}
struct gsl_struct_4 {
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
struct gsl_struct_5 {
  v_1 : vec3<f32>,
  v_2 : f32,
  v_3 : f32,
  v_4 : f32,
  v_5 : f32,
}
struct gsl_struct_6 {
  v_1 : mat4x4<f32>,
  v_2 : vec3<f32>,
  v_3 : f32,
}
