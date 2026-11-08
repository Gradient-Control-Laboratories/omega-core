fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn builtin_arcExtendedXY(p: vec3<f32>, halfAngleSinCos: vec2<f32>, radius: f32) -> f32 {
  let pxAbs = abs(p.x);
  let py = p.y;
  let l = length(vec2<f32>(pxAbs, py));
  let dOutside = l - radius;
  let dInside = pxAbs * halfAngleSinCos.x + py * halfAngleSinCos.y - radius;
  let inside = halfAngleSinCos.y * pxAbs > halfAngleSinCos.x * py;
  return select(dOutside, dInside, inside);
}

fn builtin_circleXY(p: vec3<f32>, center: vec2<f32>, radius: f32) -> f32 {
  return length(p.xy - center) - radius;
}

fn gsl_remap(p: vec3<f32>, gsl_param_7_7: f32, gsl_param_11_7: f32, gsl_param_15_7: f32, gsl_param_19_7: f32, gsl_param_42_13: f32, vector_zero: vec3<f32>) -> f32 {
  return 1.025 - abs(p.x);
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_7_7: f32, gsl_param_11_7: f32, gsl_param_15_7: f32, gsl_param_19_7: f32, gsl_param_42_13: f32, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_plane(vec3<f32>(abs(p.x), p.y, p.z), vec3<f32>(4.7, 7.3000746, -2.6998363), vec3<f32>(0.70710677, 0.0, 0.70710677));
}

fn builtin_intersectChamfer(d1: f32, d2: f32, setback: f32) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2 + setback) / 1.4142135;
  return max(face, edge);
}

fn gsl_remap_3(p: vec3<f32>, gsl_param_7_7: f32, gsl_param_11_7: f32, gsl_param_15_7: f32, gsl_param_19_7: f32, gsl_param_42_13: f32, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return abs(p.x) - 1.45;
}

fn gsl_remap_4(p: vec3<f32>, gsl_param_7_7: f32, gsl_param_11_7: f32, gsl_param_15_7: f32, gsl_param_19_7: f32, gsl_param_42_13: f32, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return 0.6 - abs(p.x);
}

fn builtin_quadXY(p: vec3<f32>, corner0: vec2<f32>, corner1: vec2<f32>, corner2: vec2<f32>, corner3: vec2<f32>) -> f32 {
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

fn map_rotateZX(v: vec3<f32>, ang: f32) -> vec3<f32> {
  let ca = cos(ang);
  let sa = sin(ang);
  return vec3<f32>(v.x * ca - v.z * sa, v.y, v.z * ca + v.x * sa);
}

fn planeXY(p: vec3<f32>, p1: vec2<f32>, p2: vec2<f32>) -> f32 {
  let dDir = p2 - p1;
  let n = normalize(vec2<f32>(-dDir.y, dDir.x));
  return builtin_plane(p, vec3<f32>(p1.x, p1.y, 0.0), vec3<f32>(n.x, n.y, 0.0));
}

fn arcShape(p: vec3<f32>, vector_zero: vec3<f32>) -> gsl_struct {
  let d = p - vec3<f32>(0.0, 0.0, 0.0);
  let lx = dot(d, vec3<f32>(0.0, -0.629699, -0.77683926));
  let ly = dot(d, vec3<f32>(0.0, -0.77683926, 0.629699));
  let rx = lx * 0.9998388 - ly * 0.017953383;
  let ry = lx * 0.017953383 + ly * 0.9998388;
  let lz = dot(d, vec3<f32>(-1.0, 0.0, 0.0));
  let arc = builtin_arcExtendedXY(vec3<f32>(rx, ry, lz), vec2<f32>(0.9271787, 0.3746194), 1.4);
  let hole = builtin_circleXY(p.zyx, vec2<f32>(0.0, 0.0), 0.75);
  let horizPlane = -rx * 0.9271787 + ry * 0.3746194 - 1.4;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = max(arc, -hole);
  gsl_result.v_2 = horizPlane;
  return gsl_result;
}

fn gcl_closure_gsl_anon(p: vec3<f32>) -> f32 {
  return p.z - -2.7;
}

fn boltHole(p: vec3<f32>, vector_dirZ: vec3<f32>, corner2D: vec2<f32>) -> f32 {
  return builtin_cylinderZ(p, vector_dirZ, vec3<f32>(corner2D.x, corner2D.y, 0.0), 0.42);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_7_7: f32, gsl_param_11_7: f32, gsl_param_15_7: f32, gsl_param_19_7: f32, gsl_param_42_13: f32) -> gsl_shape {
  let p0x = gsl_param_7_7;
  let quadV0X = -5.9 + p0x;
  let p0y = gsl_param_11_7;
  let quadV0Y = 1.5 + p0y;
  let v0 = vec2<f32>(quadV0X, quadV0Y);
  let p2x = gsl_param_15_7;
  let quadV2X = 5.9 + p2x;
  let p2y = gsl_param_19_7;
  let quadV2Y = 6.1 + p2y;
  let v2 = vec2<f32>(quadV2X, quadV2Y);
  let bottomPlane = -(p.z - -3.5);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let gsl_tmp_1 = p.z - -2.7;
  let gsl_tmp_2 = builtin_plane(vec3<f32>(abs(p.x), p.y, p.z), vec3<f32>(4.7, 7.3000746, -2.6998363), vec3<f32>(0.70710677, 0.0, 0.70710677));
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let unionTop = select(-length(vec2<f32>((gsl_tmp_1 + gsl_tmp_2) / 1.847759, (gsl_tmp_1 - gsl_tmp_2) / 0.76536685)), min(gsl_tmp_1, gsl_tmp_2), max(gcl_closure_gsl_anon(p - vec3<f32>(0.70710677, 0.0, 0.70710677) * gsl_tmp_2), gsl_remap_2(p - vector_dirZ * gsl_tmp_1, gsl_param_7_7, gsl_param_11_7, gsl_param_15_7, gsl_param_19_7, gsl_param_42_13, vector_zero, vector_dirZ)) > 0.0);
  let clevisAng = gsl_param_42_13;
  let clevisRot = map_rotateZX(p, radians(clevisAng));
  let gsl_result = arcShape(clevisRot, vector_zero);
  let horizPlane = gsl_result.v_2;
  let frontPlane = planeXY(p, vec2<f32>(5.9, 1.5), v0) - 1.2;
  let eyeSpan = gsl_remap(clevisRot, gsl_param_7_7, gsl_param_11_7, gsl_param_15_7, gsl_param_19_7, gsl_param_42_13, vector_zero);
  let eyeDraft = builtin_plane(clevisRot, vec3<f32>(1.45, 0.30003822, -1.3846282), vec3<f32>(0.0, -1.0, 1.7320508));
  let eyeStep = max(min(eyeSpan, eyeDraft + 0.8), eyeDraft);
  let topSurf = builtin_intersectChamfer(builtin_intersectChamfer(unionTop, horizPlane, 0.05), builtin_intersectChamfer(frontPlane, eyeStep, 0.05), 0.05);
  let base = max(builtin_intersectChamfer(max(max(builtin_quadXY(p, v0, vec2<f32>(-5.9, 6.1), v2, vec2<f32>(5.9, 1.5)) - 1.2, bottomPlane), -min(min(min(boltHole(p, vector_dirZ, v0), builtin_axis(p, vec3<f32>(-5.9, 6.1, 0.0), vector_dirZ) - 0.42), boltHole(p, vector_dirZ, v2)), builtin_axis(p, vec3<f32>(5.9, 1.5, 0.0), vector_dirZ) - 0.42)), topSurf, 0.05), -topSurf - 0.8);
  let clevisShape = gsl_result.v_1;
  let clevisSpan = gsl_remap_3(clevisRot, gsl_param_7_7, gsl_param_11_7, gsl_param_15_7, gsl_param_19_7, gsl_param_42_13, vector_zero, vector_dirZ);
  let rearPlane = planeXY(p, vec2<f32>(-5.9, 6.1), v2) - 1.2;
  let clevis = max(max(max(builtin_intersectChamfer(clevisShape, clevisSpan, 0.05), bottomPlane), -topSurf), rearPlane);
  let clevisCut = max(min(-eyeStep - 0.8, rearPlane), gsl_remap_4(clevisRot, gsl_param_7_7, gsl_param_11_7, gsl_param_15_7, gsl_param_19_7, gsl_param_42_13, vector_zero, vector_dirZ));
  let clevisFinal = builtin_intersectChamfer(clevis, clevisCut, 0.05);
  let body = min(base, clevisFinal);
  let gsl_s1 = gsl_shape(body, 1u);
  return gsl_s1;
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
}
