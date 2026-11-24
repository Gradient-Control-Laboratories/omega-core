fn gsl_remap(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_3 = -dot(clamp((p - vec3<f32>(0.0, 0.0, 0.0)) / max(length(p - vec3<f32>(0.0, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(1.0, 0.0, 0.0)) + 0.0;
  let gsl_tmp_4 = -dot(clamp((p - vec3<f32>(0.0, 0.0, 0.0)) / max(length(p - vec3<f32>(0.0, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 1.0, 0.0)) + 0.0;
  let gsl_tmp_5 = -dot(clamp((p - vec3<f32>(0.0, 0.0, 0.0)) / max(length(p - vec3<f32>(0.0, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 0.0, 1.0)) + 0.0;
  return vec3<f32>(gsl_tmp_3, gsl_tmp_4, gsl_tmp_5);
}

<<<<<<< HEAD
fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_cse_7 = sign(p - vec3<f32>(0.0, 0.0, 0.0));
  let gsl_tmp_2 = gsl_cse_7 * vec3<f32>(1.0, 0.0, 0.0);
  let gsl_tmp_1 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8);
  let gsl_tmp_4 = select(0.0, gsl_tmp_2.x, gsl_tmp_1.x + 0.0 > 0.0);
  let gsl_tmp_5 = select(0.0, gsl_tmp_2.y, gsl_tmp_1.y + 0.0 > 0.0);
  let gsl_tmp_6 = select(0.0, gsl_tmp_2.z, gsl_tmp_1.z + 0.0 > 0.0);
  let gsl_cse_9 = max(gsl_tmp_1.x + 0.0, 0.0);
  let gsl_cse_10 = max(gsl_tmp_1.y + 0.0, 0.0);
  let gsl_cse_11 = max(gsl_tmp_1.z + 0.0, 0.0);
  let gsl_tmp_3 = 0.0 + gsl_cse_9 * gsl_cse_9 + gsl_cse_10 * gsl_cse_10 + gsl_cse_11 * gsl_cse_11;
  let gsl_tmp_8 = gsl_cse_7 * vec3<f32>(0.0, 1.0, 0.0);
  let gsl_tmp_10 = select(0.0, gsl_tmp_8.x, gsl_tmp_1.x + 0.0 > 0.0);
  let gsl_tmp_11 = select(0.0, gsl_tmp_8.y, gsl_tmp_1.y + 0.0 > 0.0);
  let gsl_tmp_12 = select(0.0, gsl_tmp_8.z, gsl_tmp_1.z + 0.0 > 0.0);
  let gsl_tmp_14 = gsl_cse_7 * vec3<f32>(0.0, 0.0, 1.0);
  let gsl_tmp_16 = select(0.0, gsl_tmp_14.x, gsl_tmp_1.x + 0.0 > 0.0);
  let gsl_tmp_17 = select(0.0, gsl_tmp_14.y, gsl_tmp_1.y + 0.0 > 0.0);
  let gsl_tmp_18 = select(0.0, gsl_tmp_14.z, gsl_tmp_1.z + 0.0 > 0.0);
  let gsl_cse_4 = sqrt(max(gsl_tmp_3, 0.000000000000000000000000000001));
  let gsl_cse_1 = 2.0 * (gsl_cse_4 + 1.0 - select(0.0, 1.0, (gsl_cse_4 > 0.0)));
  let gsl_cse_8 = p - vec3<f32>(0.8, 0.0, 0.0);
  let gsl_cse_2 = length(gsl_cse_8) + 1.0 - select(0.0, 1.0, (length(gsl_cse_8) > 0.0));
  let gsl_cse_3 = max(max(gsl_tmp_1.x, gsl_tmp_1.y), gsl_tmp_1.z) < 0.000000000000000000000000000001;
  let gsl_cse_5 = gsl_tmp_3 > 0.000000000000000000000000000001;
  let gsl_cse_6 = max(gsl_tmp_1.x, gsl_tmp_1.y) > gsl_tmp_1.z;
  let notch = gcl_closure_intersectRoundExact_2(p, vec3<f32>(select(0.0, select(gsl_tmp_2.z, select(gsl_tmp_2.y, gsl_tmp_2.x, gsl_tmp_1.x > gsl_tmp_1.y), gsl_cse_6), gsl_cse_3) + select(0.0, gsl_tmp_4 * gsl_cse_9 + gsl_cse_9 * gsl_tmp_4 + gsl_tmp_5 * gsl_cse_10 + gsl_cse_10 * gsl_tmp_5 + gsl_tmp_6 * gsl_cse_11 + gsl_cse_11 * gsl_tmp_6, gsl_cse_5) / gsl_cse_1, select(0.0, select(gsl_tmp_8.z, select(gsl_tmp_8.y, gsl_tmp_8.x, gsl_tmp_1.x > gsl_tmp_1.y), gsl_cse_6), gsl_cse_3) + select(0.0, gsl_tmp_10 * gsl_cse_9 + gsl_cse_9 * gsl_tmp_10 + gsl_tmp_11 * gsl_cse_10 + gsl_cse_10 * gsl_tmp_11 + gsl_tmp_12 * gsl_cse_11 + gsl_cse_11 * gsl_tmp_12, gsl_cse_5) / gsl_cse_1, select(0.0, select(gsl_tmp_14.z, select(gsl_tmp_14.y, gsl_tmp_14.x, gsl_tmp_1.x > gsl_tmp_1.y), gsl_cse_6), gsl_cse_3) + select(0.0, gsl_tmp_16 * gsl_cse_9 + gsl_cse_9 * gsl_tmp_16 + gsl_tmp_17 * gsl_cse_10 + gsl_cse_10 * gsl_tmp_17 + gsl_tmp_18 * gsl_cse_11 + gsl_cse_11 * gsl_tmp_18, gsl_cse_5) / gsl_cse_1), vec3<f32>(dot(gsl_cse_8, vec3<f32>(1.0, 0.0, 0.0)) / gsl_cse_2, dot(gsl_cse_8, vec3<f32>(0.0, 1.0, 0.0)) / gsl_cse_2, dot(gsl_cse_8, vec3<f32>(0.0, 0.0, 1.0)) / gsl_cse_2));
=======
fn gsl_remap_2(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_6 = -dot(clamp((p - vec3<f32>(1.2, 0.0, 0.0)) / max(length(p - vec3<f32>(1.2, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(1.0, 0.0, 0.0)) + 0.0;
  let gsl_tmp_7 = -dot(clamp((p - vec3<f32>(1.2, 0.0, 0.0)) / max(length(p - vec3<f32>(1.2, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 1.0, 0.0)) + 0.0;
  let gsl_tmp_8 = -dot(clamp((p - vec3<f32>(1.2, 0.0, 0.0)) / max(length(p - vec3<f32>(1.2, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 0.0, 1.0)) + 0.0;
  return vec3<f32>(gsl_tmp_6, gsl_tmp_7, gsl_tmp_8);
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let gsl_tmp_1 = -(length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0) + 0.0;
  return gsl_tmp_1;
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let gsl_tmp_2 = -(length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0) + 0.0;
  return gsl_tmp_2;
}

fn builtin_edgeProjectFrom(v1: f32, v2: f32, n1: vec3<f32>, n2: vec3<f32>, c: vec3<f32>) -> vec3<f32> {
  let d = dot(n2, n2);
  let b = dot(n1, n2);
  let a = dot(n1, n1);
  let det = max(a * d - b * b, 0.000001);
  let gsl_var_step = (n1 * (d * v1 - b * v2) + n2 * (a * v2 - b * v1)) / det;
  let reach = 4.0 * (abs(v1) + abs(v2)) + 0.001;
  return c - clamp(gsl_var_step, vec3<f32>(-reach), vec3<f32>(reach));
}

fn builtin_edgeSlideFrom(p: vec3<f32>, q: vec3<f32>, n1: vec3<f32>, n2: vec3<f32>) -> vec3<f32> {
  let t = cross(n1, n2);
  let slide = t * dot(p - q, t) / max(dot(t, t), 0.000001);
  let span = length(p - q);
  return q + clamp(slide, vec3<f32>(-span), vec3<f32>(span));
}

fn builtin_edgeVerdict(e: vec3<f32>, n1: vec3<f32>, n2: vec3<f32>, sharp: f32, residual: f32) -> f32 {
  let d = dot(n2, n2);
  let x1 = dot(e, n1);
  let b = dot(n1, n2);
  let x2 = dot(e, n2);
  let a = dot(n1, n1);
  let det = max(a * d - b * b, 0.000001);
  let w1 = (d * x1 - b * x2) / det;
  let w2 = (a * x2 - b * x1) / det;
  let plausible = length(e) * length(n1 + n2) / 2.0 <= 1.5 * sharp + 0.001;
  let solved = residual <= 0.1 * length(e) + 0.001;
  let edged = a * d - b * b >= 0.0001 * a * d;
  return select(sharp, max(length(e), sharp), min(min(w1, w2), sharp) > 0.0 && plausible && solved && edged);
}

fn gsl_remap_5(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_19 = select(0.0, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).z, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).y, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y) > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z), max(max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z) < 0.000000000000000000000000000001) + select(0.0, select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0), 0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) > 0.0)))) + 0.0;
  let gsl_tmp_20 = select(0.0, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).z, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).y, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y) > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z), max(max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z) < 0.000000000000000000000000000001) + select(0.0, select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0), 0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) > 0.0)))) + 0.0;
  let gsl_tmp_21 = select(0.0, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).z, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).y, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y) > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z), max(max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z) < 0.000000000000000000000000000001) + select(0.0, select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0), 0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) > 0.0)))) + 0.0;
  return vec3<f32>(gsl_tmp_19, gsl_tmp_20, gsl_tmp_21);
}

fn gsl_remap_6(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_22 = dot(clamp((p - vec3<f32>(0.8, 0.0, 0.0)) / max(length(p - vec3<f32>(0.8, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(1.0, 0.0, 0.0)) + 0.0;
  let gsl_tmp_23 = dot(clamp((p - vec3<f32>(0.8, 0.0, 0.0)) / max(length(p - vec3<f32>(0.8, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 1.0, 0.0)) + 0.0;
  let gsl_tmp_24 = dot(clamp((p - vec3<f32>(0.8, 0.0, 0.0)) / max(length(p - vec3<f32>(0.8, 0.0, 0.0)), 0.00000000000000000001), vec3<f32>(-1.0), vec3<f32>(1.0)), vec3<f32>(0.0, 0.0, 1.0)) + 0.0;
  return vec3<f32>(gsl_tmp_22, gsl_tmp_23, gsl_tmp_24);
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  let gsl_tmp_17 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6)) + 0.0;
  return gsl_tmp_17;
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  let gsl_tmp_18 = length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0 + 0.0;
  return gsl_tmp_18;
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  let gsl_tmp_25 = p + vec3<f32>(0.0001, 0.00007, 0.00003);
  let gsl_tmp_39 = gsl_remap_5(gsl_tmp_25);
  let gsl_tmp_40 = gsl_remap_6(gsl_tmp_25);
  let gsl_tmp_26 = builtin_edgeSlideFrom(p, builtin_edgeProjectFrom(gsl_remap_7(gsl_tmp_25), gsl_remap_8(gsl_tmp_25), gsl_tmp_39, gsl_tmp_40, gsl_tmp_25), gsl_tmp_39, gsl_tmp_40);
  let gsl_tmp_41 = gsl_remap_5(gsl_tmp_26);
  let gsl_tmp_42 = gsl_remap_6(gsl_tmp_26);
  let gsl_tmp_27 = builtin_edgeSlideFrom(p, builtin_edgeProjectFrom(gsl_remap_7(gsl_tmp_26), gsl_remap_8(gsl_tmp_26), gsl_tmp_41, gsl_tmp_42, gsl_tmp_26), gsl_tmp_41, gsl_tmp_42);
  let gsl_tmp_43 = gsl_remap_5(gsl_tmp_27);
  let gsl_tmp_44 = gsl_remap_6(gsl_tmp_27);
  let gsl_tmp_28 = builtin_edgeSlideFrom(p, builtin_edgeProjectFrom(gsl_remap_7(gsl_tmp_27), gsl_remap_8(gsl_tmp_27), gsl_tmp_43, gsl_tmp_44, gsl_tmp_27), gsl_tmp_43, gsl_tmp_44);
  let gsl_tmp_32 = builtin_edgeProjectFrom(gsl_remap_7(gsl_tmp_28), gsl_remap_8(gsl_tmp_28), gsl_remap_5(gsl_tmp_28), gsl_remap_6(gsl_tmp_28), gsl_tmp_28);
  let gsl_tmp_17 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6)) + 0.0;
  let gsl_tmp_18 = length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0 + 0.0;
  let notch = builtin_edgeVerdict(p - gsl_tmp_32, gsl_remap_5(gsl_tmp_32), gsl_remap_6(gsl_tmp_32), max(gsl_tmp_17, gsl_tmp_18), max(abs(gsl_remap_7(gsl_tmp_32)), abs(gsl_remap_8(gsl_tmp_32)))) - 0.0;
>>>>>>> origin/main
  return notch;
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(sharp, 0.000000000000000000000000000001 - radius) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
<<<<<<< HEAD
  let gsl_cse_3 = p - vec3<f32>(0.0, 0.0, 0.0);
  let gsl_cse_1 = length(gsl_cse_3) + 1.0 - select(0.0, 1.0, (length(gsl_cse_3) > 0.0));
  let gsl_cse_4 = p - vec3<f32>(1.2, 0.0, 0.0);
  let gsl_cse_2 = length(gsl_cse_4) + 1.0 - select(0.0, 1.0, (length(gsl_cse_4) > 0.0));
  let lens = gcl_closure_unionRoundExact_2(p, vec3<f32>(dot(gsl_cse_3, vec3<f32>(1.0, 0.0, 0.0)) / gsl_cse_1, dot(gsl_cse_3, vec3<f32>(0.0, 1.0, 0.0)) / gsl_cse_1, dot(gsl_cse_3, vec3<f32>(0.0, 0.0, 1.0)) / gsl_cse_1), vec3<f32>(dot(gsl_cse_4, vec3<f32>(1.0, 0.0, 0.0)) / gsl_cse_2, dot(gsl_cse_4, vec3<f32>(0.0, 1.0, 0.0)) / gsl_cse_2, dot(gsl_cse_4, vec3<f32>(0.0, 0.0, 1.0)) / gsl_cse_2));
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(lens, gsl_remap(p - vec3<f32>(4.0, 0.0, 0.0)))), 1u);
=======
  let gsl_tmp_9 = p + vec3<f32>(0.0001, 0.00007, 0.00003);
  let gsl_tmp_33 = gsl_remap(gsl_tmp_9);
  let gsl_tmp_34 = gsl_remap_2(gsl_tmp_9);
  let gsl_tmp_10 = builtin_edgeSlideFrom(p, builtin_edgeProjectFrom(gsl_remap_3(gsl_tmp_9), gsl_remap_4(gsl_tmp_9), gsl_tmp_33, gsl_tmp_34, gsl_tmp_9), gsl_tmp_33, gsl_tmp_34);
  let gsl_tmp_35 = gsl_remap(gsl_tmp_10);
  let gsl_tmp_36 = gsl_remap_2(gsl_tmp_10);
  let gsl_tmp_11 = builtin_edgeSlideFrom(p, builtin_edgeProjectFrom(gsl_remap_3(gsl_tmp_10), gsl_remap_4(gsl_tmp_10), gsl_tmp_35, gsl_tmp_36, gsl_tmp_10), gsl_tmp_35, gsl_tmp_36);
  let gsl_tmp_37 = gsl_remap(gsl_tmp_11);
  let gsl_tmp_38 = gsl_remap_2(gsl_tmp_11);
  let gsl_tmp_12 = builtin_edgeSlideFrom(p, builtin_edgeProjectFrom(gsl_remap_3(gsl_tmp_11), gsl_remap_4(gsl_tmp_11), gsl_tmp_37, gsl_tmp_38, gsl_tmp_11), gsl_tmp_37, gsl_tmp_38);
  let gsl_tmp_16 = builtin_edgeProjectFrom(gsl_remap_3(gsl_tmp_12), gsl_remap_4(gsl_tmp_12), gsl_remap(gsl_tmp_12), gsl_remap_2(gsl_tmp_12), gsl_tmp_12);
  let gsl_tmp_1 = -(length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0) + 0.0;
  let gsl_tmp_2 = -(length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0) + 0.0;
  let lens = -(builtin_edgeVerdict(p - gsl_tmp_16, gsl_remap(gsl_tmp_16), gsl_remap_2(gsl_tmp_16), max(gsl_tmp_1, gsl_tmp_2), max(abs(gsl_remap_3(gsl_tmp_16)), abs(gsl_remap_4(gsl_tmp_16)))) - 0.0);
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(lens, gsl_remap_9(p - vec3<f32>(4.0, 0.0, 0.0)))), 1u);
>>>>>>> origin/main
  return gsl_s1;
}
