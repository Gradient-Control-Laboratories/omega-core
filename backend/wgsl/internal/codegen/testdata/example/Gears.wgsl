fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_circleXY(p: vec3<f32>, center: vec2<f32>, radius: f32) -> f32 {
  return length(p.xy - center) - radius;
}

fn builtin_intersectChamfer(d1: f32, d2: f32, setback: f32) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2 + setback) / 1.4142135;
  return max(face, edge);
}

fn builtin_intersectRound(d1: f32, d2: f32, radius: f32) -> f32 {
  let inset1 = max(d1 + radius, 0.0);
  let inset2 = max(d2 + radius, 0.0);
  return min(-radius, max(d1, d2)) + length(vec2<f32>(inset1, inset2));
}

fn builtin_unionRound(d1: f32, d2: f32, radius: f32) -> f32 {
  return -builtin_intersectRound(-d1, -d2, radius);
}

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn builtin_cylinder(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, radius: f32) -> f32 {
  return builtin_axis(p, center, direction) - radius;
}

fn builtin_cylinderZ(p: vec3<f32>, vector_dirZ: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return builtin_cylinder(p, center, vector_dirZ, radius);
}

fn builtin_unionChamfer(d1: f32, d2: f32, setback: f32) -> f32 {
  return -builtin_intersectChamfer(-d1, -d2, setback);
}

fn builtin_differenceChamfer(gsl_var_target: f32, tool: f32, setback: f32) -> f32 {
  return builtin_intersectChamfer(gsl_var_target, -tool, setback);
}

fn gears_involuteRadial(p: vec3<f32>, radius: f32) -> f32 {
  let radialField = -p.y;
  let rho = length(p.xy);
  let tanLen = sqrt(max(rho * rho - radius * radius, 0.0));
  let ratioVal = radius / max(rho, radius);
  let ang = atan2(p.y, p.x);
  let invField = tanLen - (acos(ratioVal) + ang) * radius;
  let result = select(radialField, invField, rho > radius);
  return result;
}

fn gsl_gears_gearProfile_remap(p: vec3<f32>, dedendumClearance: f32, kSmall: f32, moduleVal: f32, numTeeth: f32, pressureAngle: f32, rotOffset: f32) -> f32 {
  let pitchAngle = 3.1415927 / numTeeth;
  let halfPitchAngle = pitchAngle / 2.0;
  let pitchRadius = numTeeth / 2.0 * moduleVal;
  let baseRadius = pitchRadius * cos(pressureAngle);
  let prRatio = pitchRadius / baseRadius;
  let gammaP = sqrt(prRatio * prRatio - 1.0);
  let contactAngle = gammaP - atan(gammaP);
  let rotAngle = halfPitchAngle - contactAngle;
  let cosR = cos(rotAngle);
  let sinR = sin(rotAngle);
  return gears_involuteRadial(vec3<f32>(p.x * cosR + p.y * sinR, -p.x * sinR + p.y * cosR, p.z), baseRadius);
}

fn gsl_gears_gearProfile_remap_2(p: vec3<f32>, dedendumClearance: f32, kSmall: f32, moduleVal: f32, numTeeth: f32, pressureAngle: f32, rotOffset: f32) -> f32 {
  let pitchAngle = 3.1415927 / numTeeth;
  let halfPitchAngle = pitchAngle / 2.0;
  let pitchRadius = numTeeth / 2.0 * moduleVal;
  let baseRadius = pitchRadius * cos(pressureAngle);
  let prRatio = pitchRadius / baseRadius;
  let gammaP = sqrt(prRatio * prRatio - 1.0);
  let contactAngle = gammaP - atan(gammaP);
  let rotAngle = halfPitchAngle - contactAngle;
  let cosR = cos(rotAngle);
  let sinR = sin(rotAngle);
  return min(gears_involuteRadial(vec3<f32>(p.x * cosR + p.y * sinR, -p.x * sinR + p.y * cosR, p.z), baseRadius), gsl_gears_gearProfile_remap(vec3<f32>(p.x, -p.y, p.z), dedendumClearance, kSmall, moduleVal, numTeeth, pressureAngle, rotOffset));
}

fn gears_gearProfile(p: vec3<f32>, numTeeth: f32, moduleVal: f32, pressureAngle: f32, rotOffset: f32, dedendumClearance: f32, kSmall: f32) -> f32 {
  let cosO = cos(rotOffset);
  let sinO = sin(rotOffset);
  let opx = p.x * cosO + p.y * sinO;
  let opy = -p.x * sinO + p.y * cosO;
  let pRho = length(vec2<f32>(opx, opy));
  let pAngle = atan2(opy, opx);
  let pitchAngle = 3.1415927 / numTeeth;
  let pAngleShifted = pAngle + pitchAngle;
  let twicePitchAngle = pitchAngle * 2.0;
  let pAngleMod = gsl_floored((gsl_floored(pAngleShifted, twicePitchAngle) + twicePitchAngle), twicePitchAngle);
  let pAngleCentered = pAngleMod - pitchAngle;
  let rpx = pRho * cos(pAngleCentered);
  let rpy = pRho * sin(pAngleCentered);
  let repeatedTooth = gsl_gears_gearProfile_remap_2(vec3<f32>(rpx, rpy, p.z), dedendumClearance, kSmall, moduleVal, numTeeth, pressureAngle, rotOffset);
  let pitchRadius = numTeeth / 2.0 * moduleVal;
  let addendumRadius = pitchRadius + moduleVal;
  let addClip = builtin_circleXY(vec3<f32>(opx, opy, p.z), vec2<f32>(0.0, 0.0), addendumRadius);
  let toothProfile = builtin_intersectChamfer(repeatedTooth, addClip, kSmall);
  let dedendumRadius = pitchRadius - (1.0 + dedendumClearance) * moduleVal;
  let dedBody = builtin_circleXY(vec3<f32>(opx, opy, p.z), vec2<f32>(0.0, 0.0), dedendumRadius);
  let roundK = dedendumClearance * moduleVal;
  return builtin_unionRound(toothProfile, dedBody, roundK);
}

fn gears_gear(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, numTeeth: f32, moduleVal: f32, pressureAngle: f32, rotOffset: f32, thickness: f32, helixAngle: f32, standoffDiameter: f32, standoffHeight: f32, standoffChamfer: f32, boreDiameter: f32, dedendumClearance: f32, kSmall: f32) -> f32 {
  let pitchRadius = numTeeth / 2.0 * moduleVal;
  let twistRate = tan(helixAngle) / pitchRadius;
  let twistAngle = p.z * twistRate;
  let cosTw = cos(twistAngle);
  let sinTw = sin(twistAngle);
  let tpx = p.x * cosTw + p.y * sinTw;
  let tpy = -p.x * sinTw + p.y * cosTw;
  let profile = gears_gearProfile(vec3<f32>(tpx, tpy, p.z), numTeeth, moduleVal, pressureAngle, rotOffset, dedendumClearance, kSmall);
  let slab = abs(p.z) - thickness / 2.0;
  let gearBody = builtin_intersectChamfer(profile, slab, kSmall);
  let standoffRadius = standoffDiameter / 2.0;
  let standoffCyl = builtin_cylinderZ(p, vector_dirZ, vector_zero, standoffRadius);
  let standoffSlabZ = builtin_intersectChamfer(-p.z, p.z - standoffHeight, kSmall);
  let standoffSolid = builtin_intersectChamfer(standoffCyl, standoffSlabZ, kSmall);
  let withStandoff = builtin_unionChamfer(gearBody, standoffSolid, standoffChamfer);
  let boreRadius = boreDiameter / 2.0;
  let bore = builtin_cylinderZ(p, vector_dirZ, vector_zero, boreRadius);
  let form = builtin_differenceChamfer(withStandoff, bore, kSmall);
  return form;
}

fn gears_gearPair(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, numTeeth1: f32, numTeeth2: f32, moduleVal: f32, pressureAngle: f32, thickness: f32, helixAngle: f32, standoffDiam1: f32, standoffDiam2: f32, standoffHeight: f32, standoffChamfer: f32, boreDiam1: f32, boreDiam2: f32, dedendumClearance: f32, kSmall: f32, phase: f32) -> gsl_struct {
  let pitchR1 = numTeeth1 / 2.0 * moduleVal;
  let pitchAng1 = 3.1415927 / numTeeth1;
  let rot1 = pitchAng1 * 0.5 + pitchAng1 * phase * 2.0;
  let g1 = gears_gear(vec3<f32>(p.x + pitchR1, p.y, p.z), vector_zero, vector_dirZ, numTeeth1, moduleVal, pressureAngle, rot1, thickness, helixAngle, standoffDiam1, standoffHeight, standoffChamfer, boreDiam1, dedendumClearance, kSmall);
  let pitchR2 = numTeeth2 / 2.0 * moduleVal;
  let pitchAng2 = 3.1415927 / numTeeth2;
  let rot2 = -pitchAng2 * 0.5 + pitchAng2 * phase * 2.0;
  let g2 = gears_gear(vec3<f32>(-p.x + pitchR2, p.y, p.z), vector_zero, vector_dirZ, numTeeth2, moduleVal, pressureAngle, rot2, thickness, helixAngle, standoffDiam2, standoffHeight, standoffChamfer, boreDiam2, dedendumClearance, kSmall);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = g1;
  gsl_result.v_2 = g2;
  return gsl_result;
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_6_13: f32, gsl_param_9_13: f32, gsl_param_12_13: f32, gsl_param_15_17: f32, gsl_param_21_14: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let numTeeth1 = gsl_param_6_13;
  let moduleVal = gsl_param_12_13;
  let gsl_partLo = vec3<f32>(min(-(numTeeth1 / 2.0 * moduleVal) - max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6), max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6) - numTeeth1 / 2.0 * moduleVal) - 0.01, min(-max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6), max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6)) - 0.01, -1.76);
  let gsl_partHi = vec3<f32>(max(-(numTeeth1 / 2.0 * moduleVal) - max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6), max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6) - numTeeth1 / 2.0 * moduleVal) + 0.01, max(-max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6), max((numTeeth1 / 2.0 + 1.0) * moduleVal, 1.6)) + 0.01, 2.51);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let numTeeth2 = gsl_param_9_13;
  let pressureAngle = gsl_param_15_17;
  let helixAngle = gsl_param_21_14;
  let animate = gsl_floored(time, 1.0);
  let gsl_result = gears_gearPair(p, vector_zero, vector_dirZ, numTeeth1, numTeeth2, moduleVal, pressureAngle, 3.5, helixAngle, 3.0, 4.0, 2.5, 0.1, 1.2, 1.6, 0.25, 0.025, animate);
  let gsl_part_g1 = gsl_result.v_1;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_part_g1 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let g1 = gsl_part;
  let gsl_partLo_2 = vec3<f32>(min(numTeeth2 / 2.0 * moduleVal - max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1), numTeeth2 / 2.0 * moduleVal + max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1)) - 0.01, min(-max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1), max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1)) - 0.01, -1.76);
  let gsl_partHi_2 = vec3<f32>(max(numTeeth2 / 2.0 * moduleVal - max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1), numTeeth2 / 2.0 * moduleVal + max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1)) + 0.01, max(-max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1), max((numTeeth2 / 2.0 + 1.0) * moduleVal, 2.1)) + 0.01, 2.51);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  let gsl_part_g2 = gsl_result.v_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_part_g2 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let g2 = gsl_part_2;
  let gsl_s1 = gsl_shape(g1, 1u);
  let gsl_s2 = gsl_shape(g2, 2u);
  return gsl_min(gsl_s1, gsl_s2);
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

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
}
