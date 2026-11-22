fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_circleXY(p: vec3<f32>, center: vec2<f32>, radius: f32) -> f32 {
  return length(p.xy - center) - radius;
}

fn builtin_intersectChamferNormalized(d1: f32, d2: f32, depth: f32, normal1: vec2<f32>, normal2: vec2<f32>) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2) / max(length(normal1 + normal2), 0.000001) + depth;
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

fn builtin_intersectChamfer(d1: f32, d2: f32, offset: f32) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2 + offset) / 1.4142135;
  return max(face, edge);
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

fn builtin_unionChamfer(d1: f32, d2: f32, offset: f32) -> f32 {
  return -builtin_intersectChamfer(-d1, -d2, offset);
}

fn builtin_differenceChamfer(gsl_var_target: f32, tool: f32, offset: f32) -> f32 {
  return builtin_intersectChamfer(gsl_var_target, -tool, offset);
}

fn gears_bevelPitchAngle(numTeeth: f32, mateTeeth: f32, shaftAngle: f32) -> f32 {
  return atan2(sin(shaftAngle), mateTeeth / numTeeth + cos(shaftAngle));
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
  let pitchRadius = numTeeth / 2.0 * moduleVal;
  let gsl_cse_1 = vec3<f32>(opx, opy, p.z);
  let pitchDisc = builtin_circleXY(gsl_cse_1, vec2<f32>(0.0, 0.0), pitchRadius);
  let gsl_featureSize = 3.1415927 * moduleVal / 2.0;
  let gsl_featureWeight = gsl_featureWeight(gsl_featureSize);
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
  let addendumRadius = pitchRadius + moduleVal;
  let addClip = builtin_circleXY(gsl_cse_1, vec2<f32>(0.0, 0.0), addendumRadius);
  let baseRadius = pitchRadius * cos(pressureAngle);
  let tipRoll = sqrt(addendumRadius * addendumRadius - baseRadius * baseRadius) / baseRadius;
  let tipCorner = vec2<f32>(cos(tipRoll) + tipRoll * sin(tipRoll), sin(tipRoll) - tipRoll * cos(tipRoll));
  let toothProfile = builtin_intersectChamferNormalized(repeatedTooth, addClip, kSmall, vec2<f32>(sin(tipRoll), -cos(tipRoll)), tipCorner / length(tipCorner));
  let dedendumRadius = pitchRadius - (1.0 + dedendumClearance) * moduleVal;
  let dedBody = builtin_circleXY(gsl_cse_1, vec2<f32>(0.0, 0.0), dedendumRadius);
  let roundK = dedendumClearance * moduleVal;
  let profile = builtin_unionRound(toothProfile, dedBody, roundK);
  var gsl_feature = 100000.0;
  if (gsl_featureWeight > 0.0) {
    gsl_feature = max(profile, dedendumRadius - pRho) + (1.0 - gsl_featureWeight) * 0.5 * gsl_featureSize;
  }
  let tips = gsl_feature;
  var gsl_feature_2 = 100000.0;
  if (gsl_featureWeight > 0.0) {
    gsl_feature_2 = max(-profile, pRho - addendumRadius) + (1.0 - gsl_featureWeight) * 0.5 * gsl_featureSize;
  }
  let gaps = gsl_feature_2;
  return max(min(pitchDisc, tips), -gaps);
}

fn gears_bevelTeeth(p: vec3<f32>, numTeeth: f32, moduleVal: f32, pressureAngle: f32, pitchAngle: f32, spiralAngle: f32, rotOffset: f32, backlash: f32, dedendumClearance: f32, kSmall: f32) -> f32 {
  let coneDist = length(p);
  let sinD = sin(pitchAngle);
  let outerCone = numTeeth * moduleVal / (2.0 * sinD);
  let cosD = cos(pitchAngle);
  let backCone = outerCone * sinD / cosD;
  let theta = atan2(length(p.xy), p.z);
  let u = backCone + outerCone * (theta - pitchAngle);
  let spiral = -tan(spiralAngle) * log(max(coneDist, outerCone * 0.001) / outerCone) / sinD;
  let phi = atan2(p.y, p.x) - rotOffset + spiral;
  let phiWrapped = gsl_floored((gsl_floored((phi + 3.1415927), 6.2831855) + 6.2831855), 6.2831855) - 3.1415927;
  let psi = phiWrapped * cosD;
  let virtualTeeth = numTeeth / cosD;
  let dev = gears_gearProfile(vec3<f32>(u * cos(psi), u * sin(psi), 0.0), virtualTeeth, moduleVal, pressureAngle, 0.0, dedendumClearance, kSmall) + backlash / 2.0;
  return coneDist * sin(clamp(dev / outerCone, -1.5707964, 1.5707964));
}

fn gears_bevelGear(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, numTeeth: f32, moduleVal: f32, pressureAngle: f32, pitchAngle: f32, spiralAngle: f32, rotOffset: f32, faceWidth: f32, crownToBack: f32, recess: f32, backlash: f32, standoffDiameter: f32, standoffHeight: f32, standoffChamfer: f32, boreDiameter: f32, dedendumClearance: f32, kSmall: f32) -> f32 {
  let teeth = gears_bevelTeeth(p, numTeeth, moduleVal, pressureAngle, pitchAngle, spiralAngle, rotOffset, backlash, dedendumClearance, kSmall);
  let sinD = sin(pitchAngle);
  let cosD = cos(pitchAngle);
  let elementDist = length(p.xy) * sinD + p.z * cosD;
  let outerCone = numTeeth * moduleVal / (2.0 * sinD);
  let addendumAngle = moduleVal / outerCone;
  let tipCornerZ = cos(pitchAngle + addendumAngle) / cos(addendumAngle);
  let heelFace = outerCone * tipCornerZ + crownToBack;
  let heelCap = max(elementDist - outerCone, p.z - heelFace);
  let innerCone = outerCone - faceWidth;
  let toeFace = innerCone * tipCornerZ + recess;
  let toeCap = min(innerCone - elementDist, toeFace - p.z);
  let body = builtin_intersectChamfer(teeth, max(heelCap, toeCap), kSmall);
  let standoffRadius = standoffDiameter / 2.0;
  let standoffCyl = builtin_cylinderZ(p, vector_dirZ, vector_zero, standoffRadius);
  let standoffSlabZ = builtin_intersectChamfer(toeFace - p.z, p.z - (heelFace + standoffHeight), kSmall);
  let standoffSolid = builtin_intersectChamfer(standoffCyl, standoffSlabZ, kSmall);
  let withStandoff = builtin_unionChamfer(body, standoffSolid, standoffChamfer);
  let boreRadius = boreDiameter / 2.0;
  let bore = builtin_cylinderZ(p, vector_dirZ, vector_zero, boreRadius);
  let form = builtin_differenceChamfer(withStandoff, bore, kSmall);
  return form;
}

fn gsl_gears_bevelPair_remap(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, backlash: f32, boreDiam1: f32, boreDiam2: f32, crownToBack: f32, dedendumClearance: f32, faceWidth: f32, kSmall: f32, moduleVal: f32, numTeeth1: f32, numTeeth2: f32, phase: f32, pressureAngle: f32, recess: f32, shaftAngle: f32, spiralAngle: f32, standoffChamfer: f32, standoffDiam1: f32, standoffDiam2: f32, standoffHeight: f32) -> f32 {
  let pitchAngle1 = gears_bevelPitchAngle(numTeeth1, numTeeth2, shaftAngle);
  let pitchAngle2 = shaftAngle - pitchAngle1;
  let rot2 = 3.1415927 * (numTeeth2 - 1.0) / numTeeth2 - 6.2831855 * phase / numTeeth2;
  let g2 = gears_bevelGear(p, vector_zero, vector_dirZ, numTeeth2, moduleVal, pressureAngle, pitchAngle2, -spiralAngle, rot2, faceWidth, crownToBack, recess, backlash, standoffDiam2, standoffHeight, standoffChamfer, boreDiam2, dedendumClearance, kSmall);
  return g2;
}

fn gears_bevelPair(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, numTeeth1: f32, numTeeth2: f32, moduleVal: f32, pressureAngle: f32, shaftAngle: f32, spiralAngle: f32, faceWidth: f32, crownToBack: f32, recess: f32, backlash: f32, standoffDiam1: f32, standoffDiam2: f32, standoffHeight: f32, standoffChamfer: f32, boreDiam1: f32, boreDiam2: f32, dedendumClearance: f32, kSmall: f32, phase: f32) -> gsl_struct {
  let pitchAngle1 = gears_bevelPitchAngle(numTeeth1, numTeeth2, shaftAngle);
  let rot1 = 6.2831855 * phase / numTeeth1;
  let g1 = gears_bevelGear(p, vector_zero, vector_dirZ, numTeeth1, moduleVal, pressureAngle, pitchAngle1, spiralAngle, rot1, faceWidth, crownToBack, recess, backlash, standoffDiam1, standoffHeight, standoffChamfer, boreDiam1, dedendumClearance, kSmall);
  let cosS = cos(shaftAngle);
  let sinS = sin(shaftAngle);
  let turned = gsl_gears_bevelPair_remap(vec3<f32>(cosS * p.x - sinS * p.z, p.y, sinS * p.x + cosS * p.z), vector_zero, vector_dirZ, backlash, boreDiam1, boreDiam2, crownToBack, dedendumClearance, faceWidth, kSmall, moduleVal, numTeeth1, numTeeth2, phase, pressureAngle, recess, shaftAngle, spiralAngle, standoffChamfer, standoffDiam1, standoffDiam2, standoffHeight);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = g1;
  gsl_result.v_2 = turned;
  return gsl_result;
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_14_28: f32, gsl_param_15_28: f32, gsl_param_16_30: f32, gsl_param_17_34: f32, gsl_param_23_32: f32, gsl_param_28_33: f32, gsl_param_32_33: f32, gsl_param_40_32: f32, gsl_param_41_27: f32, gsl_param_46_35: f32) -> gsl_shape {
  gsl_featurePixel = select(0.0, gsl_vanishEye.w * mix(max(length(gsl_vanishEye.xyz - p), 1.0), 1.0, gsl_vanishView.x), gsl_vanishEye.w > 0.0);
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let numTeeth1 = gsl_param_14_28;
  let moduleVal = gsl_param_16_30;
  let shaftAngle = gsl_param_23_32;
  let numTeeth2 = gsl_param_15_28;
  let gsl_cse_2 = sin(shaftAngle);
  let gsl_cse_3 = cos(shaftAngle);
  let gslBox_6 = atan2(gsl_cse_2, numTeeth2 / numTeeth1 + gsl_cse_3);
  let gslBox = numTeeth1 * moduleVal / (2.0 * sin(gslBox_6));
  let gslBox_3 = gslBox_6 + moduleVal / gslBox;
  let gslBox_2 = max(gslBox * sin(gslBox_3) / cos(moduleVal / gslBox), 1.6);
  let faceFraction = gsl_param_32_33;
  let faceWidth = gslBox * faceFraction;
  let gslBox_4 = (gslBox - faceWidth) * cos(gslBox_3) / cos(moduleVal / gslBox);
  let recess = gsl_param_41_27;
  let gslBox_7 = min(gslBox_4 + recess - 0.1, gslBox_4);
  let crownToBack = gsl_param_40_32;
  let standoffHeight = gsl_param_46_35;
  let gslBox_5 = gslBox * cos(gslBox_3) / cos(moduleVal / gslBox) + crownToBack + max(standoffHeight, 0.1);
  let gsl_partLo = vec3<f32>(min(-gslBox_2, gslBox_2) - 0.01, min(-gslBox_2, gslBox_2) - 0.01, min(gslBox_7, gslBox_5) - 0.01);
  let gsl_partHi = vec3<f32>(max(-gslBox_2, gslBox_2) + 0.01, max(-gslBox_2, gslBox_2) + 0.01, max(gslBox_7, gslBox_5) + 0.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let pressureAngle = gsl_param_17_34;
  let spiralAngle = gsl_param_28_33;
  let animate = gsl_floored(time, 1.0);
  let gsl_result = gears_bevelPair(p, vector_zero, vector_dirZ, numTeeth1, numTeeth2, moduleVal, pressureAngle, shaftAngle, spiralAngle, faceWidth, crownToBack, recess, 0.025, 3.0, 3.6, standoffHeight, 0.1, 1.5, 2.0, 0.25, 0.025, animate);
  let gsl_part_g1 = gsl_result.v_1;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_part_g1 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let g1 = gsl_part;
  let gslBox_17 = shaftAngle - gslBox_6;
  let gslBox_8 = numTeeth2 * moduleVal / (2.0 * sin(gslBox_17));
  let gslBox_9 = gslBox_17 + moduleVal / gslBox_8;
  let gsl_cse_1 = cos(moduleVal / gslBox_8);
  let gslBox_18 = (gslBox_8 - faceWidth) * cos(gslBox_9) / gsl_cse_1;
  let gslBox_10 = min(gslBox_18 + recess - 0.1, gslBox_18);
  let gslBox_12 = gslBox_8 * cos(gslBox_9) / gsl_cse_1 + crownToBack + max(standoffHeight, 0.1);
  let gslBox_11 = max(gslBox_8 * sin(gslBox_9) / gsl_cse_1, 1.9);
  let gslBox_15 = min(gsl_cse_2 * gslBox_10, gsl_cse_2 * gslBox_12) - abs(gsl_cse_3) * gslBox_11;
  let gslBox_13 = max(gsl_cse_2 * gslBox_10, gsl_cse_2 * gslBox_12) + abs(gsl_cse_3) * gslBox_11;
  let gslBox_16 = min(gsl_cse_3 * gslBox_10, gsl_cse_3 * gslBox_12) - abs(gsl_cse_2) * gslBox_11;
  let gslBox_14 = max(gsl_cse_3 * gslBox_10, gsl_cse_3 * gslBox_12) + abs(gsl_cse_2) * gslBox_11;
  let gsl_partLo_2 = vec3<f32>(min(gslBox_15, gslBox_13) - 0.01, min(-gslBox_11, gslBox_11) - 0.01, min(gslBox_16, gslBox_14) - 0.01);
  let gsl_partHi_2 = vec3<f32>(max(gslBox_15, gslBox_13) + 0.01, max(-gslBox_11, gslBox_11) + 0.01, max(gslBox_16, gslBox_14) + 0.01);
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

var<private> gsl_featurePixel: f32 = 0.0;

fn gsl_featureWeight(size: f32) -> f32 {
  if (gsl_featurePixel <= 0.0) { return 1.0; }
  return smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), size / gsl_featurePixel);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
}
