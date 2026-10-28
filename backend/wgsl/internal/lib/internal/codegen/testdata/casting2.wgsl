fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn gsl_remap(p: vec3<f32>, vector_zero: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let gsl_msUvw = (p - sdfBMin0) / sdfBSize0;
  let gsl_msSample = textureSampleLevel(sdfTex0, sdfSmp0, clamp(gsl_msUvw, vec3<f32>(0.0), vec3<f32>(1.0)), 0.0).r;
  let gsl_msOutside = length(max(vec3<f32>(0.0), max(sdfBMin0 - p, p - (sdfBMin0 + sdfBSize0))));
  let gsl_msVal = select(gsl_msOutside + gsl_msSample, gsl_msSample, gsl_msOutside == 0.0);
  return gsl_msVal;
}

fn gsl_remap_2(p: vec3<f32>, vector_zero: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return gsl_remap(map_eulerAngles(p, vec3<f32>(-0.0, -1.5707964, -0.0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn builtin_intersectRound(d1: f32, d2: f32, radius: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let inset1 = max(d1 + radius, 0.0);
  let inset2 = max(d2 + radius, 0.0);
  return min(-radius, max(d1, d2)) + length(vec2<f32>(inset1, inset2));
}

fn builtin_unionSharp(fields: vec2<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn gsl_remap_3(p: vec3<f32>, vector_zero: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec2<f32>(gatePortArcA(p, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), gatePortArcB(p, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0)), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn builtin_intersectSharp(d1: f32, d2: f32, d3: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return max(max(d1, d2), d3);
}

fn builtin_circleXY(p: vec3<f32>, center: vec2<f32>, radius: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return length(p.xy - center) - radius;
}

fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn gsl_remap_4(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirX: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return max(builtin_circleXY(vec3<f32>(max(p.x, 0.0), p.y, p.z), vec2<f32>(0.0, 0.0), 1.4403487, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), builtin_plane(p, vec3<f32>(-4.0, 0.0, 0.0), vec3<f32>(-1.0, -0.0, -0.0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0));
}

fn builtin_bezierTubeDir(p: vec3<f32>, start: vec3<f32>, control: vec3<f32>, end: vec3<f32>, radius: f32, orientation: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_struct_2 {
  let d = start - p;
  let e = control - start;
  let g = e * 2.0;
  let f = start - control * 2.0 + end;
  let ff = dot(f, f);
  let kk = 1.0 / (ff + 0.0000000001);
  let ef = dot(e, f);
  let kx = kk * ef;
  let ee = dot(e, e);
  let df = dot(d, f);
  let ky = kk * (2.0 * ee + df) / 3.0;
  let de = dot(d, e);
  let kz = kk * de;
  let qq = kx * (2.0 * kx * kx - 3.0 * ky) + kz;
  let pp = ky - kx * kx;
  let zz = sqrt(max(-pp, 0.0));
  let v = acos(clamp(qq / (2.0 * pp * zz - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let m = cos(v);
  let ta = clamp((m + m) * zz - kx, 0.0, 1.0);
  let eva = d + (g + f * ta) * ta;
  let da = dot(eva, eva);
  let n = sin(v) * 1.7320508;
  let tb = clamp((-n - m) * zz - kx, 0.0, 1.0);
  let evb = d + (g + f * tb) * tb;
  let dbb = dot(evb, evb);
  let tc = clamp((n - m) * zz - kx, 0.0, 1.0);
  let evc = d + (g + f * tc) * tc;
  let dc = dot(evc, evc);
  let d3 = min(min(da, dbb), dc);
  let h = qq * qq + 4.0 * pp * pp * pp;
  let sh = sqrt(max(h, 0.0));
  let w1 = (sh - qq) / 2.0;
  let u1 = sign(w1) * exp(log(max(abs(w1), 0.000000000001)) / 3.0);
  let w2 = (-sh - qq) / 2.0;
  let u2 = sign(w2) * exp(log(max(abs(w2), 0.000000000001)) / 3.0);
  let t1 = clamp(u1 + u2 - kx, 0.0, 1.0);
  let ev1 = d + (g + f * t1) * t1;
  let d1 = dot(ev1, ev1);
  let dist2 = select(d3, d1, h >= 0.0);
  let unsignedDistance = sqrt(dist2) - radius;
  let tBest3 = select(select(tc, tb, dbb <= dc), select(tc, ta, da <= dc), da <= dbb);
  let tBest = select(tBest3, t1, h >= 0.0);
  let closestPoint = start + g * tBest + f * tBest * tBest;
  let offset = p - closestPoint;
  let orientationSign = sign(dot(offset, normalize(orientation)));
  let signedDistance = select(unsignedDistance, orientationSign * unsignedDistance, orientationSign != 0.0);
  let chordLength = length(end - start);
  let polygonLength = length(control - start) + length(end - control);
  let arcLength = (chordLength + polygonLength) * 0.5;
  let tangentRaw = g + f * 2.0 * tBest;
  let tangentLength = length(tangentRaw);
  let tangent = tangentRaw / (tangentLength + 0.000000000001);
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = signedDistance;
  gsl_result.v_2 = tBest;
  gsl_result.v_3 = arcLength;
  gsl_result.v_4 = closestPoint;
  gsl_result.v_5 = tangent;
  return gsl_result;
}

fn builtin_mix(a: f32, b: f32, t: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return a + (b - a) * t;
}

fn builtin_bezierQuadPointRound(p: vec3<f32>, start: vec3<f32>, control: vec3<f32>, end: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_struct_4 {
  let d = start - p;
  let e = control - start;
  let g = e * 2.0;
  let f = start - control * 2.0 + end;
  let ff = dot(f, f);
  let ee = dot(e, e);
  let kk = 1.0 / max(ff, 0.000001 * ee + 0.00000000000000000001);
  let ef = dot(e, f);
  let kx = kk * ef;
  let df = dot(d, f);
  let ky = kk * (2.0 * ee + df) / 3.0;
  let de = dot(d, e);
  let kz = kk * de;
  let qq = kx * (2.0 * kx * kx - 3.0 * ky) + kz;
  let pp = ky - kx * kx;
  let zz = sqrt(max(-pp, 0.0));
  let v = acos(clamp(qq / (2.0 * pp * zz - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let m = cos(v);
  let ta = clamp((m + m) * zz - kx, 0.0, 1.0);
  let eva = d + (g + f * ta) * ta;
  let da = dot(eva, eva);
  let n = sin(v) * 1.7320508;
  let tb = clamp((-n - m) * zz - kx, 0.0, 1.0);
  let evb = d + (g + f * tb) * tb;
  let dbb = dot(evb, evb);
  let tc = clamp((n - m) * zz - kx, 0.0, 1.0);
  let evc = d + (g + f * tc) * tc;
  let dc = dot(evc, evc);
  let d3 = min(min(da, dbb), dc);
  let h = qq * qq + 4.0 * pp * pp * pp;
  let sh = sqrt(max(h, 0.0));
  let w1 = (sh - qq) / 2.0;
  let u1 = sign(w1) * exp(log(max(abs(w1), 0.000000000001)) / 3.0);
  let w2 = (-sh - qq) / 2.0;
  let u2 = sign(w2) * exp(log(max(abs(w2), 0.000000000001)) / 3.0);
  let t1 = clamp(u1 + u2 - kx, 0.0, 1.0);
  let ev1 = d + (g + f * t1) * t1;
  let d1 = dot(ev1, ev1);
  let tLine = clamp(-dot(d, g) / (dot(g, g) + 0.00000000000000000001), 0.0, 1.0);
  let evLine = d + (g + f * tLine) * tLine;
  let dLine = dot(evLine, evLine);
  let straight = ff < 0.000001 * ee;
  let dist2 = select(select(d3, d1, h >= 0.0), dLine, straight);
  let distanceToCurve = sqrt(dist2);
  let tBest3 = select(select(tc, tb, dbb <= dc), select(tc, ta, da <= dc), da <= dbb);
  let tBest = select(select(tBest3, t1, h >= 0.0), tLine, straight);
  let closestPoint = start + g * tBest + f * tBest * tBest;
  var gsl_result : gsl_struct_4;
  gsl_result.v_1 = distanceToCurve;
  gsl_result.v_2 = tBest;
  gsl_result.v_3 = closestPoint;
  return gsl_result;
}

fn builtin_bezierQuadRound(p: vec3<f32>, start: vec3<f32>, control: vec3<f32>, end: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let gsl_result_6 = builtin_bezierQuadPointRound(p, start, control, end, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let distanceToCurve = gsl_result_6.v_1;
  return distanceToCurve;
}

fn gsl_remap_5(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let gsl_tmp_1 = vec3<f32>(p.x - 4.0, p.y, p.z);
  let topPlane = dot(p - vector_zero, normalize(vec3<f32>(0.0, 0.0, 1.0)));
  let gates = min(builtin_intersectSharp(gsl_remap_3(gsl_tmp_1, vector_zero, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), dot(p - vec3<f32>(0.0, 0.0, 2.0), normalize(vec3<f32>(0.0, 0.0, 1.0))), -(topPlane + 0.1), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), builtin_intersectSharp(gsl_remap_4(gsl_tmp_1, vector_zero, vector_dirX, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), builtin_plane(p, vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(-0.0, -0.0, -1.0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), dot(p - vec3<f32>(0.0, 0.0, 2.0), normalize(vector_dirZ)), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0));
  let gsl_tmp_2 = p.z - -1.15;
  let drainsAt = min(min(gcl_closure_buildDrainBlended(p, vector_zero, vec3<f32>(0.0, 0.0, -1.35), vec3<f32>(0.0, 0.0, -3.75), vec3<f32>(-2.2215927, 0.0, -4.55), 0.2157614, gsl_tmp_1, gsl_tmp_2, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), gcl_closure_buildDrainBlended_2(p, vector_zero, vec3<f32>(0.0, 0.0, -1.35), vec3<f32>(0.0, 0.0, -3.75), vec3<f32>(-2.2215927, 0.0, -4.55), 0.89762163, gsl_tmp_1, gsl_tmp_2, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0)), length(gsl_tmp_1 - vec3<f32>(-2.2215927, 0.0, -4.55)) - 0.27840745);
  let vent = builtin_bezierQuadRound(p, vec3<f32>(4.0, 0.0, 2.0), vec3<f32>(4.0, 0.0, 6.125), vec3<f32>(3.0, 0.0, 7.5), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0) - 0.3;
  let sideNoPartAt = min(min(gates, drainsAt), vent);
  let partAt = gsl_remap_2(gsl_tmp_1, vector_zero, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let side = min(sideNoPartAt, partAt);
  return side;
}

fn gsl_remap_6(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let gsl_tmp_1 = vec3<f32>(p.x - 4.0, p.y, p.z);
  let topPlane = dot(p - vector_zero, normalize(vec3<f32>(0.0, 0.0, 1.0)));
  let gates = min(builtin_intersectSharp(gsl_remap_3(gsl_tmp_1, vector_zero, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), dot(p - vec3<f32>(0.0, 0.0, 2.0), normalize(vec3<f32>(0.0, 0.0, 1.0))), -(topPlane + 0.1), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), builtin_intersectSharp(gsl_remap_4(gsl_tmp_1, vector_zero, vector_dirX, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), builtin_plane(p, vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(-0.0, -0.0, -1.0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), dot(p - vec3<f32>(0.0, 0.0, 2.0), normalize(vector_dirZ)), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0));
  let gsl_tmp_2 = p.z - -1.15;
  let drainsAt = min(min(gcl_closure_buildDrainBlended(p, vector_zero, vec3<f32>(0.0, 0.0, -1.35), vec3<f32>(0.0, 0.0, -3.75), vec3<f32>(-2.2215927, 0.0, -4.55), 0.2157614, gsl_tmp_1, gsl_tmp_2, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), gcl_closure_buildDrainBlended_2(p, vector_zero, vec3<f32>(0.0, 0.0, -1.35), vec3<f32>(0.0, 0.0, -3.75), vec3<f32>(-2.2215927, 0.0, -4.55), 0.89762163, gsl_tmp_1, gsl_tmp_2, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0)), length(gsl_tmp_1 - vec3<f32>(-2.2215927, 0.0, -4.55)) - 0.27840745);
  let vent = builtin_bezierQuadRound(p, vec3<f32>(4.0, 0.0, 2.0), vec3<f32>(4.0, 0.0, 6.125), vec3<f32>(3.0, 0.0, 7.5), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0) - 0.3;
  let sideNoPartAt = min(min(gates, drainsAt), vent);
  return sideNoPartAt;
}

fn builtin_linearMap(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: f32, outMax: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let slope = (outMax - outMin) / (inMax - inMin);
  return (gsl_var_value - inMin) * slope + outMin;
}

fn builtin_ramp(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: f32, outMax: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let mapped = builtin_linearMap(gsl_var_value, inMin, inMax, outMin, outMax, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let low = min(outMin, outMax);
  let high = max(outMin, outMax);
  return clamp(mapped, low, high);
}

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn builtin_cylinder(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, radius: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return builtin_axis(p, center, direction, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0) - radius;
}

fn builtin_cylinderZ(p: vec3<f32>, vector_dirZ: vec3<f32>, center: vec3<f32>, radius: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return builtin_cylinder(p, center, vector_dirZ, radius, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn builtin_intersectChamfer(d1: f32, d2: f32, setback: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2 + setback) / 1.4142135;
  return max(face, edge);
}

fn gsl_remap_7(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirX: vec3<f32>, vector_dirZ: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let sprueCone = builtin_cylinderZ(p, vector_dirZ, vector_zero, builtin_ramp(p.z, 2.5, 8.0, 1.5, 4.0, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let sprueTop = p.z - 8.0;
  let sprueHole = (abs(sprueCone - sprueTop) / 1.4142135 + abs(p.y) * 1.5 - (sprueCone + sprueTop + 0.42426407) / 1.4142135) / 1.7320508 - 0.15;
  return sprueHole;
}

fn builtin_unionSharp_2(fields: vec4<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return min(min(min(fields.x, fields.y), fields.z), fields.w);
}

fn builtin_intersectSharp_2(fields: vec2<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec2<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound(vec2<f32>(radialDistance, axialDistance), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  return form;
}

fn map_rotateYZ(v: vec3<f32>, ang: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> vec3<f32> {
  let ca = cos(ang);
  let sa = sin(ang);
  return vec3<f32>(v.x, v.y * ca + v.z * sa, -v.y * sa + v.z * ca);
}

fn map_rotateZX(v: vec3<f32>, ang: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> vec3<f32> {
  let ca = cos(ang);
  let sa = sin(ang);
  return vec3<f32>(v.x * ca - v.z * sa, v.y, v.z * ca + v.x * sa);
}

fn map_rotateXY(v: vec3<f32>, ang: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> vec3<f32> {
  let ca = cos(ang);
  let sa = sin(ang);
  return vec3<f32>(v.x * ca + v.y * sa, -v.x * sa + v.y * ca, v.z);
}

fn map_eulerAngles(v: vec3<f32>, angles: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> vec3<f32> {
  return map_rotateXY(map_rotateZX(map_rotateYZ(v, angles.x, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), angles.y, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), angles.z, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn gatePortFoldedXY(p: vec3<f32>, count: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_struct {
  let rad = length(p.xy);
  let ang = atan2(p.x, p.y);
  let wedge = 6.2831855 / count;
  let halfWedge = wedge * 0.5;
  let angmod = gsl_floored((ang + halfWedge), wedge) - halfWedge;
  let foldedX = rad * sin(angmod);
  let foldedY = rad * cos(angmod);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = foldedX;
  gsl_result.v_2 = foldedY;
  return gsl_result;
}

fn gatePortArc(p: vec3<f32>, gatePortPatternCount: f32, gatePortAnnulusRadius: f32, gatePortAnnulusThickness: f32, gatePortXThickness: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let annulusField = abs(length(p.xy) - gatePortAnnulusRadius) - gatePortAnnulusThickness * 0.5;
  let gsl_result = gatePortFoldedXY(p, gatePortPatternCount, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let foldedX = gsl_result.v_1;
  let xSlabField = abs(foldedX) - gatePortXThickness * 0.5;
  return builtin_intersectRound(annulusField, xSlabField, 0.0, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn gatePortArcA(p: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return gatePortArc(p, 6.0, 1.2903488, 0.15, 0.5, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn gatePortArcB(p: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return gatePortArc(p, 2.0, 0.31016126, 0.15, 0.35, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn drainEndProfile(p: vec3<f32>, vector_zero: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  return builtin_circleXY(p, vec2<f32>(0.0, 0.0), 0.27840745, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
}

fn gcl_closure_sweepLoftScaleSmoothSeparateCaps(p: vec3<f32>, vector_zero: vec3<f32>, start: vec3<f32>, control: vec3<f32>, end: vec3<f32>, orientation: vec3<f32>, scale: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_struct_3 {
  let gsl_result_2 = builtin_bezierTubeDir(p, start, control, end, 0.0, orientation, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let closestPoint = gsl_result_2.v_4;
  let offset = p - closestPoint;
  let tangent = gsl_result_2.v_5;
  let unitOrientation = normalize(orientation);
  let orientationInPlane = unitOrientation - tangent * dot(unitOrientation, tangent);
  let orientationInPlaneLength = length(orientationInPlane);
  let binormal = orientationInPlane / (orientationInPlaneLength + 0.000000000001);
  let normal = cross(tangent, binormal);
  let localX = dot(offset, normal);
  let curveParameter = gsl_result_2.v_2;
  let clampedParameter = clamp(curveParameter, 0.0, 1.0);
  let easedParameter = clampedParameter * clampedParameter * (3.0 - 2.0 * clampedParameter);
  let startScale = builtin_mix(1.0, scale, easedParameter, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let localY = dot(offset, binormal);
  let startLocalPoint = vec3<f32>(localX / startScale, localY / startScale, 0.0);
  let startDistance = gatePortArcA(startLocalPoint, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0) * startScale;
  let inverseScale = 1.0 / scale;
  let endScale = builtin_mix(inverseScale, 1.0, easedParameter, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let endLocalPoint = vec3<f32>(localX / endScale, localY / endScale, 0.0);
  let endDistance = drainEndProfile(endLocalPoint, vector_zero, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0) * endScale;
  let profileDistance = builtin_mix(startDistance, endDistance, easedParameter, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let startTangent = normalize((control - start) * 2.0);
  let startCap = -dot(p - start, startTangent);
  let endTangent = normalize((end - control) * 2.0);
  let endCap = dot(p - end, endTangent);
  var gsl_result : gsl_struct_3;
  gsl_result.v_1 = profileDistance;
  gsl_result.v_2 = startCap;
  gsl_result.v_3 = endCap;
  return gsl_result;
}

fn gcl_closure_buildDrainBlended(p: vec3<f32>, vector_zero: vec3<f32>, a: vec3<f32>, b: vec3<f32>, c: vec3<f32>, sweepScale: f32, shiftedP: vec3<f32>, drainCap: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let gsl_result_3 = gcl_closure_sweepLoftScaleSmoothSeparateCaps(shiftedP, vector_zero, a, b, c, vec3<f32>(0.0, 1.0, 0.0), sweepScale, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let profile = gsl_result_3.v_1;
  let arcField = gatePortArcA(shiftedP, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let transZ = shiftedP.z - -1.25 + 0.1;
  let blended = builtin_mix(profile, arcField, clamp(transZ / 0.1, 0.0, 1.0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let capC = gsl_result_3.v_3;
  return max(max(blended, capC), drainCap);
}

fn gcl_closure_sweepLoftScaleSmoothSeparateCaps_2(p: vec3<f32>, vector_zero: vec3<f32>, start: vec3<f32>, control: vec3<f32>, end: vec3<f32>, orientation: vec3<f32>, scale: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_struct_3 {
  let gsl_result_4 = builtin_bezierTubeDir(p, start, control, end, 0.0, orientation, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let closestPoint = gsl_result_4.v_4;
  let offset = p - closestPoint;
  let tangent = gsl_result_4.v_5;
  let unitOrientation = normalize(orientation);
  let orientationInPlane = unitOrientation - tangent * dot(unitOrientation, tangent);
  let orientationInPlaneLength = length(orientationInPlane);
  let binormal = orientationInPlane / (orientationInPlaneLength + 0.000000000001);
  let normal = cross(tangent, binormal);
  let localX = dot(offset, normal);
  let curveParameter = gsl_result_4.v_2;
  let clampedParameter = clamp(curveParameter, 0.0, 1.0);
  let easedParameter = clampedParameter * clampedParameter * (3.0 - 2.0 * clampedParameter);
  let startScale = builtin_mix(1.0, scale, easedParameter, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let localY = dot(offset, binormal);
  let startLocalPoint = vec3<f32>(localX / startScale, localY / startScale, 0.0);
  let startDistance = gatePortArcB(startLocalPoint, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0) * startScale;
  let inverseScale = 1.0 / scale;
  let endScale = builtin_mix(inverseScale, 1.0, easedParameter, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let endLocalPoint = vec3<f32>(localX / endScale, localY / endScale, 0.0);
  let endDistance = drainEndProfile(endLocalPoint, vector_zero, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0) * endScale;
  let profileDistance = builtin_mix(startDistance, endDistance, easedParameter, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let startTangent = normalize((control - start) * 2.0);
  let startCap = -dot(p - start, startTangent);
  let endTangent = normalize((end - control) * 2.0);
  let endCap = dot(p - end, endTangent);
  var gsl_result : gsl_struct_3;
  gsl_result.v_1 = profileDistance;
  gsl_result.v_2 = startCap;
  gsl_result.v_3 = endCap;
  return gsl_result;
}

fn gcl_closure_buildDrainBlended_2(p: vec3<f32>, vector_zero: vec3<f32>, a: vec3<f32>, b: vec3<f32>, c: vec3<f32>, sweepScale: f32, shiftedP: vec3<f32>, drainCap: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let gsl_result_5 = gcl_closure_sweepLoftScaleSmoothSeparateCaps_2(shiftedP, vector_zero, a, b, c, vec3<f32>(0.0, 1.0, 0.0), sweepScale, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let profile = gsl_result_5.v_1;
  let arcField = gatePortArcB(shiftedP, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let transZ = shiftedP.z - -1.25 + 0.1;
  let blended = builtin_mix(profile, arcField, clamp(transZ / 0.1, 0.0, 1.0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let capC = gsl_result_5.v_3;
  return max(max(blended, capC), drainCap);
}

fn sdShape(p: vec3<f32>, time: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_shape {
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let vector_dirX = vec3<f32>(1.0, 0.0, 0.0);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let sideNoPartAt1 = gsl_remap_6(map_rotateXY(p, -1.0471976, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let sideNoPartAt2 = gsl_remap_6(map_rotateXY(p, -2.0943952, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let sideNoPartAt3 = gsl_remap_6(map_rotateXY(p, 1.0471976, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let sideNoPartAt4 = gsl_remap_6(map_rotateXY(p, 2.0943952, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let patternNoPart = min(min(min(sideNoPartAt1, sideNoPartAt2), sideNoPartAt3), sideNoPartAt4);
  let sprueCone = builtin_cylinderZ(p, vector_dirZ, vector_zero, builtin_ramp(p.z, 2.5, 8.0, 1.5, 4.0, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let sprueBottom = -p.z + -4.55 - 1.0;
  let sprueTop = p.z - 8.0;
  let sprueMountHole = max(length(p.xy - vec2<f32>(0.0, 0.0)) - 0.5, -p.z + 8.0 - 3.0);
  let sprueShaft = max(builtin_intersectChamfer(max(sprueCone, sprueBottom), sprueTop, 0.5, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), -sprueMountHole);
  let sprue = min(builtin_gcl_closure_cylinderCapped(p, vec3<f32>(0.0, 0.0, -5.55), vec3<f32>(0.0, 0.0, -4.55), 2.056815, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), sprueShaft);
  let gsl_tmp_3 = length(p.xy);
  let gsl_tmp_4 = gsl_floored((atan2(p.y, p.x) + 0.2617994), 0.5235988) - 0.2617994;
  let sprueHoles = gsl_remap_7(vec3<f32>(gsl_tmp_3 * cos(gsl_tmp_4), gsl_tmp_3 * sin(gsl_tmp_4), p.z), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let solidOutside = max(min(patternNoPart, sprue), -sprueHoles);
  let gsl_tmp_5 = vec3<f32>(gsl_floored((gsl_floored((p.x + 0.25), 0.5) + 0.5), 0.5) - 0.25, gsl_floored((gsl_floored((p.y + 0.25), 0.5) + 0.5), 0.5) - 0.25, gsl_floored((gsl_floored((p.z + 0.25), 0.5) + 0.5), 0.5) - 0.25);
  let vector_dirY = vec3<f32>(0.0, 1.0, 0.0);
  let latticeField = builtin_unionSharp_2(vec4<f32>(length(gsl_tmp_5 - normalize(vec3<f32>(1.0, 1.0, 0.0) + vector_dirZ) * clamp(dot(gsl_tmp_5, normalize(vec3<f32>(1.0, 1.0, 0.0) + vector_dirZ)), -0.4330127, 0.4330127)) - 0.05, length(gsl_tmp_5 - normalize(vec3<f32>(-1.0, -0.0, -0.0) + vector_dirY + vector_dirZ) * clamp(dot(gsl_tmp_5, normalize(vec3<f32>(-1.0, -0.0, -0.0) + vector_dirY + vector_dirZ)), -0.4330127, 0.4330127)) - 0.05, length(gsl_tmp_5 - normalize(vec3<f32>(1.0, -1.0, 0.0) + vector_dirZ) * clamp(dot(gsl_tmp_5, normalize(vec3<f32>(1.0, -1.0, 0.0) + vector_dirZ)), -0.4330127, 0.4330127)) - 0.05, length(gsl_tmp_5 - normalize(vec3<f32>(1.0, 1.0, 0.0) - vector_dirZ) * clamp(dot(gsl_tmp_5, normalize(vec3<f32>(1.0, 1.0, 0.0) - vector_dirZ)), -0.4330127, 0.4330127)) - 0.05), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let side1 = gsl_remap_5(map_rotateXY(p, -1.0471976, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let side2 = gsl_remap_5(map_rotateXY(p, -2.0943952, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let side3 = gsl_remap_5(map_rotateXY(p, 1.0471976, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let side4 = gsl_remap_5(map_rotateXY(p, 2.0943952, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), vector_zero, vector_dirX, vector_dirZ, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let pattern = min(min(min(side1, side2), side3), side4);
  let solidInside = min(pattern, sprue) + 0.08;
  let latticedInside = builtin_intersectRound(-latticeField, solidInside, 0.05, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let gsl_s1 = gsl_shape(max(min(min(min(min(1e30, 1e30), 1e30), 1e30), solidOutside), -(latticedInside)), 1u);
  return gsl_s1;
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
}
struct gsl_struct_2 {
  v_1 : f32,
  v_2 : f32,
  v_3 : f32,
  v_4 : vec3<f32>,
  v_5 : vec3<f32>,
}
struct gsl_struct_3 {
  v_1 : f32,
  v_2 : f32,
  v_3 : f32,
}
struct gsl_struct_4 {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
}
