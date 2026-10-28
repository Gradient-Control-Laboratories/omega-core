fn gsl_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let f1 = noise_cellularF1Lattice(p / 5.0, noiseTex) * 5.0 - 3.0 + length(p) - 7.0;
  return f1;
}

fn gsl_remap_2(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_2 = p / 5.0;
  let edges = max((gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f1) * 5.0 - 0.6, length(p) - 7.0);
  return edges;
}

fn noise_cellularF1Lattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_cellularScan(q, 1, noiseTex).f1;
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_tmp_1 = p / 5.0;
  let f2 = -(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 * 5.0 - 4.0) + length(p) - 7.0;
  let gsl_s1 = gsl_shape(gsl_remap(vec3<f32>(p.x + 18.0, p.y, p.z), noiseTex), 1u);
  let gsl_s2 = gsl_shape(f2, 2u);
  let gsl_s3 = gsl_shape(gsl_remap_2(vec3<f32>(p.x - 18.0, p.y, p.z), noiseTex), 3u);
  return gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3);
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
