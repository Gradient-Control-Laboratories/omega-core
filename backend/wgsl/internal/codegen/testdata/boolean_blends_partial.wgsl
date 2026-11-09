fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_firstOf(list: f32) -> f32 {
  return list;
}

fn builtin_intersectChamferEdges(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp(fields);
  return builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.x + fields.y + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields.x + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.z + setback) / 1.4142135, 0.0 > 0.0)))));
}

fn builtin_lastOf(list: f32) -> f32 {
  return list;
}

fn builtin_intersectChamferVertices(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp(fields);
  return builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.z + setback) / 1.7320508, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0)))))));
}

fn builtin_intersectChamfer(fields: vec3<f32>, setback: f32) -> f32 {
  return max(max(builtin_intersectSharp(fields), builtin_intersectChamferEdges(fields, builtin_firstOf(setback))), builtin_intersectChamferVertices(fields, builtin_lastOf(setback)));
}

fn builtin_gcl_closure_intersectChamfer(fields: vec3<f32>) -> f32 {
  return builtin_intersectChamfer(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_2(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamfer(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_gcl_closure_intersectRound_2(radius: f32, fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, radius);
}

fn builtin_gcl_closure_intersectChamferBlend(radius: f32, fields: vec3<f32>, setback: f32) -> f32 {
  return builtin_gcl_closure_intersectRound_2(radius, vec3<f32>(builtin_intersectSharp(fields), builtin_intersectChamferEdges(fields, builtin_firstOf(setback)), builtin_intersectChamferVertices(fields, builtin_lastOf(setback))));
}

fn builtin_intersectChamferRound(fields: vec3<f32>, setback: f32, radius: f32) -> f32 {
  return builtin_gcl_closure_intersectChamferBlend(radius, fields, setback);
}

fn builtin_gcl_closure_intersectChamferRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectChamferRound(fields, 0.2, 0.1);
}

fn builtin_gcl_closure_boxCentered_3(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamferRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_intersectLp(fields: vec3<f32>, width: f32, exponent: f32) -> f32 {
  let peak = builtin_intersectSharp(fields);
  let peakInset = max(peak + width, 0.0);
  let normalizer = max(peakInset, 0.000000001);
  let sumOfPowers = 0.0 + pow(max(fields.x + width, 0.0) / normalizer, exponent) + pow(max(fields.y + width, 0.0) / normalizer, exponent) + pow(max(fields.z + width, 0.0) / normalizer, exponent);
  return min(-width, peak) + peakInset * pow(sumOfPowers, 1.0 / exponent);
}

fn builtin_gcl_closure_intersectLp(fields: vec3<f32>) -> f32 {
  return builtin_intersectLp(fields, 0.2, 4.0);
}

fn builtin_gcl_closure_boxCentered_4(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectLp(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_intersectExponential(fields: vec3<f32>, width: f32) -> f32 {
  let peak = builtin_intersectSharp(fields);
  let scale = width / 2.7182817;
  let total = 0.0 + exp((fields.x - peak) / scale) + exp((fields.y - peak) / scale) + exp((fields.z - peak) / scale);
  return peak + scale * log(total);
}

fn builtin_gcl_closure_intersectExponential(fields: vec3<f32>) -> f32 {
  return builtin_intersectExponential(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_5(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectExponential(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_intersectRvachev(fields: vec3<f32>, band: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let safeBand = max(band, 0.000001);
  let progress = clamp((sharp + safeBand) / safeBand, 0.0, 1.0);
  let weight = progress * progress * (3.0 - 2.0 * progress);
  let sum = 0.0 + fields.x + fields.y + fields.z;
  let sumOfSquares = 0.0 + fields.x * fields.x + fields.y * fields.y + fields.z * fields.z;
  let rvachev = sum + sqrt(sumOfSquares);
  let blended = (1.0 - weight) * sharp + weight * rvachev;
  return select(blended, sharp, sharp < -safeBand);
}

fn builtin_gcl_closure_intersectRvachev(fields: vec3<f32>) -> f32 {
  return builtin_intersectRvachev(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_6(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRvachev(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_unionRound(fields: vec3<f32>, radius: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectRound(negated, radius);
}

fn builtin_gcl_closure_unionRound(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_7(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_unionRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_unionChamfer(fields: vec3<f32>, setback: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectChamfer(negated, setback);
}

fn builtin_gcl_closure_unionChamfer(fields: vec3<f32>) -> f32 {
  return builtin_unionChamfer(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_8(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_unionChamfer(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_unionSharp(fields: vec3<f32>) -> f32 {
  return min(min(fields.x, fields.y), fields.z);
}

fn builtin_unionChamferEdges(fields: vec3<f32>, setback: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectChamferEdges(negated, setback);
}

fn builtin_unionChamferVertices(fields: vec3<f32>, setback: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectChamferVertices(negated, setback);
}

fn builtin_gcl_closure_unionRound_2(radius: f32, fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, radius);
}

fn builtin_gcl_closure_unionChamferBlend(radius: f32, fields: vec3<f32>, setback: f32) -> f32 {
  return builtin_gcl_closure_unionRound_2(radius, vec3<f32>(builtin_unionSharp(fields), builtin_unionChamferEdges(fields, builtin_firstOf(setback)), builtin_unionChamferVertices(fields, builtin_lastOf(setback))));
}

fn builtin_unionChamferRound(fields: vec3<f32>, setback: f32, radius: f32) -> f32 {
  return builtin_gcl_closure_unionChamferBlend(radius, fields, setback);
}

fn builtin_gcl_closure_unionChamferRound(fields: vec3<f32>) -> f32 {
  return builtin_unionChamferRound(fields, 0.2, 0.1);
}

fn builtin_gcl_closure_boxCentered_9(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_unionChamferRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_unionLp(fields: vec3<f32>, width: f32, exponent: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectLp(negated, width, exponent);
}

fn builtin_gcl_closure_unionLp(fields: vec3<f32>) -> f32 {
  return builtin_unionLp(fields, 0.2, 4.0);
}

fn builtin_gcl_closure_boxCentered_10(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_unionLp(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_unionExponential(fields: vec3<f32>, width: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectExponential(negated, width);
}

fn builtin_gcl_closure_unionExponential(fields: vec3<f32>) -> f32 {
  return builtin_unionExponential(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_11(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_unionExponential(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_unionRvachev(fields: vec3<f32>, band: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectRvachev(negated, band);
}

fn builtin_gcl_closure_unionRvachev(fields: vec3<f32>) -> f32 {
  return builtin_unionRvachev(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_12(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_unionRvachev(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_unionSharp_2(fields: array<f32, 12>) -> f32 {
  return min(min(min(min(min(min(min(min(min(min(min(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]), fields[6]), fields[7]), fields[8]), fields[9]), fields[10]), fields[11]);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let a = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let b = builtin_gcl_closure_boxCentered_2(p, vec3<f32>(3.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let c = builtin_gcl_closure_boxCentered_3(p, vec3<f32>(6.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let d = builtin_gcl_closure_boxCentered_4(p, vec3<f32>(9.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let e = builtin_gcl_closure_boxCentered_5(p, vec3<f32>(12.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let f = builtin_gcl_closure_boxCentered_6(p, vec3<f32>(15.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let g = builtin_gcl_closure_boxCentered_7(p, vec3<f32>(18.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let h = builtin_gcl_closure_boxCentered_8(p, vec3<f32>(21.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let i = builtin_gcl_closure_boxCentered_9(p, vec3<f32>(24.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let j = builtin_gcl_closure_boxCentered_10(p, vec3<f32>(27.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let k = builtin_gcl_closure_boxCentered_11(p, vec3<f32>(30.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let l = builtin_gcl_closure_boxCentered_12(p, vec3<f32>(33.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let gsl_s1 = gsl_shape(builtin_unionSharp_2(array<f32, 12>(a, b, c, d, e, f, g, h, i, j, k, l)), 1u);
  return gsl_s1;
}
