fn gsl_remap(p: vec3<f32>, gsl_param_4_15: f32, gsl_param_7_14: f32, gsl_param_13_12: f32, vector_zero: vec3<f32>, vector_dirY: vec3<f32>) -> f32 {
  return max(builtin_axis(p, vector_zero, vector_dirY) - 0.5, abs(p.y) - 6.5);
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_4_15: f32, gsl_param_7_14: f32, gsl_param_13_12: f32, vector_zero: vec3<f32>, vector_dirY: vec3<f32>) -> vec2<f32> {
  let slidePos = gsl_param_13_12;
  return vec2<f32>(-1.7, slidePos);
}

fn builtin_sphere(p: vec3<f32>, radius: f32) -> f32 {
  return length(p) - radius;
}

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn normalizeField__d(v: f32, v__dp0: f32, v__dp1: f32, v__dp2: f32) -> f32 {
  return v / length(vec3<f32>(v__dp0, v__dp1, v__dp2));
}

fn normalizeCurve__d(a: f32, b: f32, a__dp0: f32, a__dp1: f32, a__dp2: f32, b__dp0: f32, b__dp1: f32, b__dp2: f32) -> gsl_struct {
  let sum = normalizeField__d(a + b, a__dp0 + b__dp0, a__dp1 + b__dp1, a__dp2 + b__dp2);
  let difference = normalizeField__d(a - b, a__dp0 - b__dp0, a__dp1 - b__dp1, a__dp2 - b__dp2);
  var gsl_result : gsl_struct;
  gsl_result.v_1 = length(vec2<f32>(sum, difference));
  gsl_result.v_2 = atan2(difference, sum);
  return gsl_result;
}

fn arm(p: vec3<f32>, ax: f32, ay: f32, bx: f32, by: f32, halfWidth: f32, halfThick: f32) -> f32 {
  let va = vec2<f32>(ax, ay);
  let vb = vec2<f32>(bx, by);
  let ab = vb - va;
  let ap = p.xy - va;
  let t = clamp(dot(ap, ab) / dot(ab, ab), 0.0, 1.0);
  let np = va + ab * t;
  let d2d = length(p.xy - np);
  let dPerp = d2d - halfWidth;
  let dZ = abs(p.z) - halfThick;
  return max(dPerp, dZ);
}

fn coupler(p: vec3<f32>, x0: f32, x1: f32, y: f32, height: f32, width: f32) -> f32 {
  let va = vec3<f32>(x0, y, 0.0);
  let vb = vec3<f32>(x1, y, 0.0);
  let ab = vb - va;
  let ap = p - va;
  let t = clamp(dot(ap, ab) / dot(ab, ab), 0.0, 1.0);
  let np = va + ab * t;
  let cap1 = length(p.zx - np.zx) - height / 2.0;
  let slab1 = abs(p.y - y) - width / 2.0;
  let cap0 = length(p.xy - np.xy) - width / 2.0;
  let slab0 = abs(p.z) - height / 2.0;
  return select(max(cap1, slab1), max(cap0, slab0), t > 0.5);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_4_15: f32, gsl_param_7_14: f32, gsl_param_13_12: f32) -> gsl_shape {
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let vector_dirY = vec3<f32>(0.0, 1.0, 0.0);
  let shaft = gsl_remap(vec3<f32>(p.x + 3.3, p.y - 1.7, p.z), gsl_param_4_15, gsl_param_7_14, gsl_param_13_12, vector_zero, vector_dirY);
  let baseRadius = gsl_param_7_14;
  let r1 = baseRadius;
  let slideRadius = gsl_param_4_15;
  let r2 = slideRadius;
  let slidePos = gsl_param_13_12;
  let slideCenter = gsl_remap_2(p - vec3<f32>(0.0, 0.0, slidePos), gsl_param_4_15, gsl_param_7_14, gsl_param_13_12, vector_zero, vector_dirY);
  let dx_c = slideCenter.x - -1.7;
  let dy_c = slideCenter.y - -4.0;
  let d = sqrt(dx_c * dx_c + dy_c * dy_c);
  let a = (r1 * r1 - r2 * r2 + d * d) / (2.0 * d);
  let mx = -1.7 + a * dx_c / d;
  let h = sqrt(r1 * r1 - a * a);
  let perpX = -(dy_c / d);
  let my = -4.0 + a * dy_c / d;
  let perpY = dx_c / d;
  let intPt = vec2<f32>(mx - h * perpX, my - h * perpY);
  let intersectionMarker = builtin_sphere(vec3<f32>(p.x - intPt.x, p.y - intPt.y, p.z), 0.3);
  let lowerArm = arm(p, -1.7, -4.0, intPt.x, intPt.y, 0.7, 0.8);
  let upperArm = arm(p, slideCenter.x, slideCenter.y, intPt.x, intPt.y, 0.7, 0.8);
  let couplerArmSlider = coupler(p, -3.3, -1.7, slideCenter.y, 1.6, 1.4);
  let couplerArmBase = coupler(p, -3.3, -1.7, -4.0, 1.6, 1.4);
  let cylBase = length(p.xy - vec2<f32>(-1.7, -4.0)) - r1;
  let cylSlide = length(p.xy - slideCenter - vec2<f32>(0.0, p.z)) - r2;
  let gsl_result = normalizeCurve__d(cylBase, cylSlide, dot(p.xy - vec2<f32>(-1.7, -4.0), vec2<f32>(1.0, 0.0)) / (length(p.xy - vec2<f32>(-1.7, -4.0)) + 1.0 - select(0.0, 1.0, (length(p.xy - vec2<f32>(-1.7, -4.0)) > 0.0))), dot(p.xy - vec2<f32>(-1.7, -4.0), vec2<f32>(0.0, 1.0)) / (length(p.xy - vec2<f32>(-1.7, -4.0)) + 1.0 - select(0.0, 1.0, (length(p.xy - vec2<f32>(-1.7, -4.0)) > 0.0))), dot(p.xy - vec2<f32>(-1.7, -4.0), vec2<f32>(0.0, 0.0)) / (length(p.xy - vec2<f32>(-1.7, -4.0)) + 1.0 - select(0.0, 1.0, (length(p.xy - vec2<f32>(-1.7, -4.0)) > 0.0))), dot(p.xy - slideCenter - vec2<f32>(0.0, p.z), vec2<f32>(1.0, 0.0)) / (length(p.xy - slideCenter - vec2<f32>(0.0, p.z)) + 1.0 - select(0.0, 1.0, (length(p.xy - slideCenter - vec2<f32>(0.0, p.z)) > 0.0))), dot(p.xy - slideCenter - vec2<f32>(0.0, p.z), vec2<f32>(0.0, 1.0)) / (length(p.xy - slideCenter - vec2<f32>(0.0, p.z)) + 1.0 - select(0.0, 1.0, (length(p.xy - slideCenter - vec2<f32>(0.0, p.z)) > 0.0))), dot(p.xy - slideCenter - vec2<f32>(0.0, p.z), vec2<f32>(0.0, -1.0)) / (length(p.xy - slideCenter - vec2<f32>(0.0, p.z)) + 1.0 - select(0.0, 1.0, (length(p.xy - slideCenter - vec2<f32>(0.0, p.z)) > 0.0))));
  let dist = gsl_result.v_1;
  let cylpath = dist - 0.02;
  let gsl_s1 = gsl_shape(shaft, 1u);
  let gsl_s2 = gsl_shape(intersectionMarker, 2u);
  let gsl_s3 = gsl_shape(lowerArm, 3u);
  let gsl_s4 = gsl_shape(upperArm, 4u);
  let gsl_s5 = gsl_shape(couplerArmSlider, 5u);
  let gsl_s6 = gsl_shape(couplerArmBase, 6u);
  let gsl_s7 = gsl_shape(cylpath, 7u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
}
