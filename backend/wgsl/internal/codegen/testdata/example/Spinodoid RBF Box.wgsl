fn builtin_intersectSharp(fields: vec3<f32>, spinodoidTex: texture_2d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, spinodoidTex: texture_2d<f32>) -> f32 {
  let sharp = builtin_intersectSharp(fields, spinodoidTex);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, spinodoidTex: texture_2d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0, spinodoidTex);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, spinodoidTex: texture_2d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), spinodoidTex);
  return solid;
}

fn lattice_solidify(field: f32, bias: f32, thickness: f32, spinodoidTex: texture_2d<f32>) -> f32 {
  return abs(field - bias) - thickness * 0.5;
}

fn sdShape(p: vec3<f32>, time: f32, spinodoidTex: texture_2d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-15.0, -15.0, -15.0);
  let gsl_partHi = vec3<f32>(15.0, 15.0, 15.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let block = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(30.0, 30.0, 30.0)), spinodoidTex);
  let isotropic = gsl_spinodoid(p, spinodoidTex, 0, 10.896585, 21.79317, 0.055555556, 0.66489935, 1.0471976, 0.0, true);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = max(block, lattice_solidify(isotropic, 0.0, 0.6, spinodoidTex)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let specimen = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-16.5, -16.5, -16.5);
  let gsl_partHi_2 = vec3<f32>(16.5, 1.5, 16.5);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = min(min(min(min(1000000000.0, length(p - vec3<f32>(0.0, 0.0, 15.0))), length(p - vec3<f32>(15.0, 0.0, 0.0))), length(p - vec3<f32>(-15.0, -15.0, -15.0))), length(p - vec3<f32>(0.0, -15.0, 5.0))) - 1.5 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let markers = gsl_part_2;
  let gsl_s1 = gsl_shape(specimen, 1u);
  let gsl_s2 = gsl_shape(markers, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

fn gsl_spinodoid(q: vec3<f32>, tex: texture_2d<f32>, slot: i32, h: f32, reach: f32, a: f32, sigma: f32, baseBeta: f32, level: f32, localNorm: bool) -> f32 {
  let originRow = textureLoad(tex, vec2<i32>(slot * 4, 0), 0);
  let sizePeriodic = textureLoad(tex, vec2<i32>(slot * 4 + 1, 0), 0);
  let origin = vec3<i32>(originRow.xyz);
  let size = vec3<i32>(sizePeriodic.xyz);
  let row = i32(originRow.w);
  let periodic = sizePeriodic.w > 0.5;
  if (!periodic) {
    let domainMin = textureLoad(tex, vec2<i32>(slot * 4 + 2, 0), 0).xyz;
    let domainMax = textureLoad(tex, vec2<i32>(slot * 4 + 3, 0), 0).xyz;
    if (any(q < domainMin) || any(q > domainMax)) {
      let outsideDelta = q - clamp(q, domainMin, domainMax);
      return length(outsideDelta);
    }
  }
  var acc = 0.0;
  var wAcc = 0.0;
  var bwAcc = 0.0;
  let pi = 3.141592653589793;
  let a2 = a * a;
  let r2max = reach * reach;
  for (var z = i32(ceil((q.z - reach) / h)); z <= i32(floor((q.z + reach) / h)); z++) {
    let dz = q.z - f32(z) * h;
    let ry2 = r2max - dz * dz;
    if (ry2 < 0.0) { continue; }
    let ry = sqrt(ry2);
    var wz = z - origin.z;
    if (periodic) { wz = ((wz % size.z) + size.z) % size.z; }
    if (wz < 0 || wz >= size.z) { continue; }
    for (var y = i32(ceil((q.y - ry) / h)); y <= i32(floor((q.y + ry) / h)); y++) {
      let dy = q.y - f32(y) * h;
      let rx2 = ry2 - dy * dy;
      if (rx2 < 0.0) { continue; }
      let rx = sqrt(rx2);
      var wy = y - origin.y;
      if (periodic) { wy = ((wy % size.y) + size.y) % size.y; }
      if (wy < 0 || wy >= size.y) { continue; }
      for (var x = i32(ceil((q.x - rx) / h)); x <= i32(floor((q.x + rx) / h)); x++) {
        let r = q - vec3<f32>(f32(x), f32(y), f32(z)) * h;
        let d2 = dot(r, r);
        if (d2 >= r2max) { continue; }
        var wx = x - origin.x;
        if (periodic) { wx = ((wx % size.x) + size.x) % size.x; }
        if (wx < 0 || wx >= size.x) { continue; }
        let index = (wy * size.z + wz) * size.x + wx;
        let t = textureLoad(tex, vec2<i32>(index % 4096, row + index / 4096), 0);
        let kv = t.xyz;
        let phase = dot(r, kv) + t.w;
        let g = exp(-pi * a2 * d2);
        let taper = 1.0 - d2 / r2max;
        let k = g * taper * taper;
        acc += k * cos(phase);
        if (localNorm) {
          let w = k * k;
          wAcc += w;
          bwAcc += w * length(kv);
        }
      }
    }
  }
  var b = baseBeta;
  if (localNorm && wAcc > 0.0 && bwAcc > 0.0) {
    b = bwAcc / wAcc;
  }
  return (acc / sigma - level) / b;
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
