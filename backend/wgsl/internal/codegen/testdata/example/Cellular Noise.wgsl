fn gsl_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let f1 = noise_cellularF1Lattice(p / 5.0, noiseTex) * 5.0 - 3.0 + length(p) - 7.0;
  return f1;
}

fn gsl_remap_2(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_1 = p / 5.0;
  let f2 = -(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 * 5.0 - 4.0) + length(p) - 7.0;
  return f2;
}

fn gsl_remap_3(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_2 = p / 5.0;
  let edges = max((gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f1) * 5.0 - 0.6, length(p) - 7.0);
  return edges;
}

fn gsl_remap_4(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let walls = max(gsl_cellularWallScan(p / 5.0, 1, noiseTex).wall * 5.0 - 0.3, length(p) - 7.0);
  return walls;
}

fn noise_cellularF1Lattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_cellularScan(q, 1, noiseTex).f1;
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_s1 = gsl_shape(gsl_remap(vec3<f32>(p.x + 27.0, p.y, p.z), noiseTex), 1u);
  let gsl_s2 = gsl_shape(gsl_remap_2(vec3<f32>(p.x + 9.0, p.y, p.z), noiseTex), 2u);
  let gsl_s3 = gsl_shape(gsl_remap_3(vec3<f32>(p.x - 9.0, p.y, p.z), noiseTex), 3u);
  let gsl_s4 = gsl_shape(gsl_remap_4(vec3<f32>(p.x - 27.0, p.y, p.z), noiseTex), 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
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

struct gsl_CellularWalls {
  wall: f32,
  wallFeature: vec3<f32>,
}

fn gsl_cellularWallScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularWalls {
  let pi = floor(q);
  let pf = q - pi;
  var d: array<vec3<f32>, 27>;
  var h: array<vec3<f32>, 27>;
  var best = 64.000000;
  var a = 0;
  for (var i = 0; i < 27; i = i + 1) {
    let o = vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
    h[i] = gsl_cellularHash(pi + o, oct, noiseTex);
    d[i] = pf - (o + h[i]);
    let d2 = dot(d[i], d[i]);
    if (d2 < best) {
      best = d2;
      a = i;
    }
  }
  let da = d[a];
  var wall = 64.000000;
  var wallCell = pi;
  var wallH = vec3<f32>(0.0);
  for (var i = 0; i < 27; i = i + 1) {
    if (i == a) {
      continue;
    }
    let ac = d[i] - da;
    let w = (dot(d[i], d[i]) - best) / (2.0 * sqrt(dot(ac, ac)));
    if (w < wall) {
      wall = w;
      wallCell = pi + vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
      wallH = h[i];
    }
  }
  // Only a point within F1 + 2*wall of q can own a nearer plane.
  let reach = sqrt(best) + 2.0 * wall;
  let lo = vec3<i32>(max(floor(pf - vec3<f32>(reach)), vec3<f32>(-2.0)));
  let hi = vec3<i32>(min(floor(pf + vec3<f32>(reach)), vec3<f32>(2.0)));
  for (var oz = lo.z; oz <= hi.z; oz = oz + 1) {
    for (var oy = lo.y; oy <= hi.y; oy = oy + 1) {
      for (var ox = lo.x; ox <= hi.x; ox = ox + 1) {
        if (abs(ox) <= 1 && abs(oy) <= 1 && abs(oz) <= 1) {
          continue;
        }
        let o = vec3<f32>(f32(ox), f32(oy), f32(oz));
        let hc = gsl_cellularHash(pi + o, oct, noiseTex);
        let dc = pf - (o + hc);
        let ac = dc - da;
        let w = (dot(dc, dc) - best) / (2.0 * sqrt(dot(ac, ac)));
        if (w < wall) {
          wall = w;
          wallCell = pi + o;
          wallH = hc;
        }
      }
    }
  }
  var out: gsl_CellularWalls;
  out.wall = wall;
  out.wallFeature = wallCell + wallH;
  return out;
}
