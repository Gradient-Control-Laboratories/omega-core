fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_1 = -(length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0) + 0.0;
  return gsl_tmp_1;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let gsl_tmp_2 = -(length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0) + 0.0;
  return gsl_tmp_2;
}

fn gsl_remap_3(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_3 = -(dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(0.0, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.0, 0.0, 0.0)) > 0.0)))) + 0.0;
  let gsl_tmp_4 = -(dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(0.0, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.0, 0.0, 0.0)) > 0.0)))) + 0.0;
  let gsl_tmp_5 = -(dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(0.0, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.0, 0.0, 0.0)) > 0.0)))) + 0.0;
  return vec3<f32>(gsl_tmp_3, gsl_tmp_4, gsl_tmp_5);
}

fn gsl_remap_4(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_6 = -(dot(p - vec3<f32>(1.2, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(1.2, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(1.2, 0.0, 0.0)) > 0.0)))) + 0.0;
  let gsl_tmp_7 = -(dot(p - vec3<f32>(1.2, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(1.2, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(1.2, 0.0, 0.0)) > 0.0)))) + 0.0;
  let gsl_tmp_8 = -(dot(p - vec3<f32>(1.2, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(1.2, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(1.2, 0.0, 0.0)) > 0.0)))) + 0.0;
  return vec3<f32>(gsl_tmp_6, gsl_tmp_7, gsl_tmp_8);
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let gsl_tmp_27 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6)) + 0.0;
  return gsl_tmp_27;
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  let gsl_tmp_28 = length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0 + 0.0;
  return gsl_tmp_28;
}

fn gsl_remap_7(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_29 = select(0.0, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).z, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).y, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y) > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z), max(max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z) < 0.000000000000000000000000000001) + select(0.0, select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(1.0, 0.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0), 0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) > 0.0)))) + 0.0;
  let gsl_tmp_30 = select(0.0, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).z, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).y, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y) > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z), max(max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z) < 0.000000000000000000000000000001) + select(0.0, select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 1.0, 0.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0), 0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) > 0.0)))) + 0.0;
  let gsl_tmp_31 = select(0.0, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).z, select((sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).y, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y) > (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z), max(max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y), (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z) < 0.000000000000000000000000000001) + select(0.0, select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).x, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).y, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0 > 0.0) + select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * select(0.0, (sign(p - vec3<f32>(0.0, 0.0, 0.0)) * vec3<f32>(0.0, 0.0, 1.0)).z, (abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0 > 0.0), 0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(0.0 + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).x + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).y + 0.0, 0.0) + max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0) * max((abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(0.8, 0.8, 0.8)).z + 0.0, 0.0), 0.000000000000000000000000000001)) > 0.0)))) + 0.0;
  return vec3<f32>(gsl_tmp_29, gsl_tmp_30, gsl_tmp_31);
}

fn gsl_remap_8(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_32 = dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0))) + 0.0;
  let gsl_tmp_33 = dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0))) + 0.0;
  let gsl_tmp_34 = dot(p - vec3<f32>(0.8, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)) / (length(p - vec3<f32>(0.8, 0.0, 0.0)) + 1.0 - select(0.0, 1.0, (length(p - vec3<f32>(0.8, 0.0, 0.0)) > 0.0))) + 0.0;
  return vec3<f32>(gsl_tmp_32, gsl_tmp_33, gsl_tmp_34);
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  let gsl_tmp_27 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6)) + 0.0;
  let gsl_tmp_28 = length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0 + 0.0;
  let gsl_tmp_52 = max(gsl_tmp_27, gsl_tmp_28);
  let gsl_tmp_35 = p + vec3<f32>(0.0001, 0.00007, 0.00003);
  let gsl_tmp_96 = gsl_remap_7(gsl_tmp_35);
  let gsl_tmp_97 = gsl_remap_8(gsl_tmp_35);
  let gsl_tmp_100 = dot(gsl_tmp_97, gsl_tmp_97);
  let gsl_tmp_94 = gsl_remap_5(gsl_tmp_35);
  let gsl_tmp_99 = dot(gsl_tmp_96, gsl_tmp_97);
  let gsl_tmp_95 = gsl_remap_6(gsl_tmp_35);
  let gsl_tmp_98 = dot(gsl_tmp_96, gsl_tmp_96);
  let gsl_tmp_101 = 4.0 * (abs(gsl_tmp_94) + abs(gsl_tmp_95)) + 0.001;
  let gsl_tmp_102 = gsl_tmp_35 - clamp((gsl_tmp_96 * (gsl_tmp_100 * gsl_tmp_94 - gsl_tmp_99 * gsl_tmp_95) + gsl_tmp_97 * (gsl_tmp_98 * gsl_tmp_95 - gsl_tmp_99 * gsl_tmp_94)) / max(gsl_tmp_98 * gsl_tmp_100 - gsl_tmp_99 * gsl_tmp_99, 0.000001), vec3<f32>(-gsl_tmp_101), vec3<f32>(gsl_tmp_101));
  let gsl_tmp_103 = cross(gsl_tmp_96, gsl_tmp_97);
  let gsl_tmp_104 = length(p - gsl_tmp_102);
  let gsl_tmp_36 = gsl_tmp_102 + clamp(gsl_tmp_103 * dot(p - gsl_tmp_102, gsl_tmp_103) / max(dot(gsl_tmp_103, gsl_tmp_103), 0.000001), vec3<f32>(-gsl_tmp_104), vec3<f32>(gsl_tmp_104));
  let gsl_tmp_107 = gsl_remap_7(gsl_tmp_36);
  let gsl_tmp_108 = gsl_remap_8(gsl_tmp_36);
  let gsl_tmp_111 = dot(gsl_tmp_108, gsl_tmp_108);
  let gsl_tmp_105 = gsl_remap_5(gsl_tmp_36);
  let gsl_tmp_110 = dot(gsl_tmp_107, gsl_tmp_108);
  let gsl_tmp_106 = gsl_remap_6(gsl_tmp_36);
  let gsl_tmp_109 = dot(gsl_tmp_107, gsl_tmp_107);
  let gsl_tmp_112 = 4.0 * (abs(gsl_tmp_105) + abs(gsl_tmp_106)) + 0.001;
  let gsl_tmp_113 = gsl_tmp_36 - clamp((gsl_tmp_107 * (gsl_tmp_111 * gsl_tmp_105 - gsl_tmp_110 * gsl_tmp_106) + gsl_tmp_108 * (gsl_tmp_109 * gsl_tmp_106 - gsl_tmp_110 * gsl_tmp_105)) / max(gsl_tmp_109 * gsl_tmp_111 - gsl_tmp_110 * gsl_tmp_110, 0.000001), vec3<f32>(-gsl_tmp_112), vec3<f32>(gsl_tmp_112));
  let gsl_tmp_114 = cross(gsl_tmp_107, gsl_tmp_108);
  let gsl_tmp_115 = length(p - gsl_tmp_113);
  let gsl_tmp_37 = gsl_tmp_113 + clamp(gsl_tmp_114 * dot(p - gsl_tmp_113, gsl_tmp_114) / max(dot(gsl_tmp_114, gsl_tmp_114), 0.000001), vec3<f32>(-gsl_tmp_115), vec3<f32>(gsl_tmp_115));
  let gsl_tmp_118 = gsl_remap_7(gsl_tmp_37);
  let gsl_tmp_119 = gsl_remap_8(gsl_tmp_37);
  let gsl_tmp_122 = dot(gsl_tmp_119, gsl_tmp_119);
  let gsl_tmp_116 = gsl_remap_5(gsl_tmp_37);
  let gsl_tmp_121 = dot(gsl_tmp_118, gsl_tmp_119);
  let gsl_tmp_117 = gsl_remap_6(gsl_tmp_37);
  let gsl_tmp_120 = dot(gsl_tmp_118, gsl_tmp_118);
  let gsl_tmp_123 = 4.0 * (abs(gsl_tmp_116) + abs(gsl_tmp_117)) + 0.001;
  let gsl_tmp_124 = gsl_tmp_37 - clamp((gsl_tmp_118 * (gsl_tmp_122 * gsl_tmp_116 - gsl_tmp_121 * gsl_tmp_117) + gsl_tmp_119 * (gsl_tmp_120 * gsl_tmp_117 - gsl_tmp_121 * gsl_tmp_116)) / max(gsl_tmp_120 * gsl_tmp_122 - gsl_tmp_121 * gsl_tmp_121, 0.000001), vec3<f32>(-gsl_tmp_123), vec3<f32>(gsl_tmp_123));
  let gsl_tmp_125 = cross(gsl_tmp_118, gsl_tmp_119);
  let gsl_tmp_126 = length(p - gsl_tmp_124);
  let gsl_tmp_38 = gsl_tmp_124 + clamp(gsl_tmp_125 * dot(p - gsl_tmp_124, gsl_tmp_125) / max(dot(gsl_tmp_125, gsl_tmp_125), 0.000001), vec3<f32>(-gsl_tmp_126), vec3<f32>(gsl_tmp_126));
  let gsl_tmp_129 = gsl_remap_7(gsl_tmp_38);
  let gsl_tmp_130 = gsl_remap_8(gsl_tmp_38);
  let gsl_tmp_133 = dot(gsl_tmp_130, gsl_tmp_130);
  let gsl_tmp_127 = gsl_remap_5(gsl_tmp_38);
  let gsl_tmp_132 = dot(gsl_tmp_129, gsl_tmp_130);
  let gsl_tmp_128 = gsl_remap_6(gsl_tmp_38);
  let gsl_tmp_131 = dot(gsl_tmp_129, gsl_tmp_129);
  let gsl_tmp_134 = 4.0 * (abs(gsl_tmp_127) + abs(gsl_tmp_128)) + 0.001;
  let gsl_tmp_42 = gsl_tmp_38 - clamp((gsl_tmp_129 * (gsl_tmp_133 * gsl_tmp_127 - gsl_tmp_132 * gsl_tmp_128) + gsl_tmp_130 * (gsl_tmp_131 * gsl_tmp_128 - gsl_tmp_132 * gsl_tmp_127)) / max(gsl_tmp_131 * gsl_tmp_133 - gsl_tmp_132 * gsl_tmp_132, 0.000001), vec3<f32>(-gsl_tmp_134), vec3<f32>(gsl_tmp_134));
  let gsl_tmp_45 = p - gsl_tmp_42;
  let gsl_tmp_44 = gsl_remap_8(gsl_tmp_42);
  let gsl_tmp_48 = dot(gsl_tmp_44, gsl_tmp_44);
  let gsl_tmp_43 = gsl_remap_7(gsl_tmp_42);
  let gsl_tmp_50 = dot(gsl_tmp_45, gsl_tmp_43);
  let gsl_tmp_47 = dot(gsl_tmp_43, gsl_tmp_44);
  let gsl_tmp_51 = dot(gsl_tmp_45, gsl_tmp_44);
  let gsl_tmp_46 = dot(gsl_tmp_43, gsl_tmp_43);
  let gsl_tmp_49 = max(gsl_tmp_46 * gsl_tmp_48 - gsl_tmp_47 * gsl_tmp_47, 0.000001);
  let notch = select(gsl_tmp_52, max(length(gsl_tmp_45), gsl_tmp_52), min(min((gsl_tmp_48 * gsl_tmp_50 - gsl_tmp_47 * gsl_tmp_51) / gsl_tmp_49, (gsl_tmp_46 * gsl_tmp_51 - gsl_tmp_47 * gsl_tmp_50) / gsl_tmp_49), gsl_tmp_52) > 0.0 && length(gsl_tmp_45) * length(gsl_tmp_43 + gsl_tmp_44) / 2.0 <= 1.5 * gsl_tmp_52 + 0.001) - 0.0;
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
  let gsl_tmp_1 = -(length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0) + 0.0;
  let gsl_tmp_2 = -(length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0) + 0.0;
  let gsl_tmp_26 = max(gsl_tmp_1, gsl_tmp_2);
  let gsl_tmp_9 = p + vec3<f32>(0.0001, 0.00007, 0.00003);
  let gsl_tmp_55 = gsl_remap_3(gsl_tmp_9);
  let gsl_tmp_56 = gsl_remap_4(gsl_tmp_9);
  let gsl_tmp_59 = dot(gsl_tmp_56, gsl_tmp_56);
  let gsl_tmp_53 = gsl_remap(gsl_tmp_9);
  let gsl_tmp_58 = dot(gsl_tmp_55, gsl_tmp_56);
  let gsl_tmp_54 = gsl_remap_2(gsl_tmp_9);
  let gsl_tmp_57 = dot(gsl_tmp_55, gsl_tmp_55);
  let gsl_tmp_60 = 4.0 * (abs(gsl_tmp_53) + abs(gsl_tmp_54)) + 0.001;
  let gsl_tmp_61 = gsl_tmp_9 - clamp((gsl_tmp_55 * (gsl_tmp_59 * gsl_tmp_53 - gsl_tmp_58 * gsl_tmp_54) + gsl_tmp_56 * (gsl_tmp_57 * gsl_tmp_54 - gsl_tmp_58 * gsl_tmp_53)) / max(gsl_tmp_57 * gsl_tmp_59 - gsl_tmp_58 * gsl_tmp_58, 0.000001), vec3<f32>(-gsl_tmp_60), vec3<f32>(gsl_tmp_60));
  let gsl_tmp_62 = cross(gsl_tmp_55, gsl_tmp_56);
  let gsl_tmp_63 = length(p - gsl_tmp_61);
  let gsl_tmp_10 = gsl_tmp_61 + clamp(gsl_tmp_62 * dot(p - gsl_tmp_61, gsl_tmp_62) / max(dot(gsl_tmp_62, gsl_tmp_62), 0.000001), vec3<f32>(-gsl_tmp_63), vec3<f32>(gsl_tmp_63));
  let gsl_tmp_66 = gsl_remap_3(gsl_tmp_10);
  let gsl_tmp_67 = gsl_remap_4(gsl_tmp_10);
  let gsl_tmp_70 = dot(gsl_tmp_67, gsl_tmp_67);
  let gsl_tmp_64 = gsl_remap(gsl_tmp_10);
  let gsl_tmp_69 = dot(gsl_tmp_66, gsl_tmp_67);
  let gsl_tmp_65 = gsl_remap_2(gsl_tmp_10);
  let gsl_tmp_68 = dot(gsl_tmp_66, gsl_tmp_66);
  let gsl_tmp_71 = 4.0 * (abs(gsl_tmp_64) + abs(gsl_tmp_65)) + 0.001;
  let gsl_tmp_72 = gsl_tmp_10 - clamp((gsl_tmp_66 * (gsl_tmp_70 * gsl_tmp_64 - gsl_tmp_69 * gsl_tmp_65) + gsl_tmp_67 * (gsl_tmp_68 * gsl_tmp_65 - gsl_tmp_69 * gsl_tmp_64)) / max(gsl_tmp_68 * gsl_tmp_70 - gsl_tmp_69 * gsl_tmp_69, 0.000001), vec3<f32>(-gsl_tmp_71), vec3<f32>(gsl_tmp_71));
  let gsl_tmp_73 = cross(gsl_tmp_66, gsl_tmp_67);
  let gsl_tmp_74 = length(p - gsl_tmp_72);
  let gsl_tmp_11 = gsl_tmp_72 + clamp(gsl_tmp_73 * dot(p - gsl_tmp_72, gsl_tmp_73) / max(dot(gsl_tmp_73, gsl_tmp_73), 0.000001), vec3<f32>(-gsl_tmp_74), vec3<f32>(gsl_tmp_74));
  let gsl_tmp_77 = gsl_remap_3(gsl_tmp_11);
  let gsl_tmp_78 = gsl_remap_4(gsl_tmp_11);
  let gsl_tmp_81 = dot(gsl_tmp_78, gsl_tmp_78);
  let gsl_tmp_75 = gsl_remap(gsl_tmp_11);
  let gsl_tmp_80 = dot(gsl_tmp_77, gsl_tmp_78);
  let gsl_tmp_76 = gsl_remap_2(gsl_tmp_11);
  let gsl_tmp_79 = dot(gsl_tmp_77, gsl_tmp_77);
  let gsl_tmp_82 = 4.0 * (abs(gsl_tmp_75) + abs(gsl_tmp_76)) + 0.001;
  let gsl_tmp_83 = gsl_tmp_11 - clamp((gsl_tmp_77 * (gsl_tmp_81 * gsl_tmp_75 - gsl_tmp_80 * gsl_tmp_76) + gsl_tmp_78 * (gsl_tmp_79 * gsl_tmp_76 - gsl_tmp_80 * gsl_tmp_75)) / max(gsl_tmp_79 * gsl_tmp_81 - gsl_tmp_80 * gsl_tmp_80, 0.000001), vec3<f32>(-gsl_tmp_82), vec3<f32>(gsl_tmp_82));
  let gsl_tmp_84 = cross(gsl_tmp_77, gsl_tmp_78);
  let gsl_tmp_85 = length(p - gsl_tmp_83);
  let gsl_tmp_12 = gsl_tmp_83 + clamp(gsl_tmp_84 * dot(p - gsl_tmp_83, gsl_tmp_84) / max(dot(gsl_tmp_84, gsl_tmp_84), 0.000001), vec3<f32>(-gsl_tmp_85), vec3<f32>(gsl_tmp_85));
  let gsl_tmp_88 = gsl_remap_3(gsl_tmp_12);
  let gsl_tmp_89 = gsl_remap_4(gsl_tmp_12);
  let gsl_tmp_92 = dot(gsl_tmp_89, gsl_tmp_89);
  let gsl_tmp_86 = gsl_remap(gsl_tmp_12);
  let gsl_tmp_91 = dot(gsl_tmp_88, gsl_tmp_89);
  let gsl_tmp_87 = gsl_remap_2(gsl_tmp_12);
  let gsl_tmp_90 = dot(gsl_tmp_88, gsl_tmp_88);
  let gsl_tmp_93 = 4.0 * (abs(gsl_tmp_86) + abs(gsl_tmp_87)) + 0.001;
  let gsl_tmp_16 = gsl_tmp_12 - clamp((gsl_tmp_88 * (gsl_tmp_92 * gsl_tmp_86 - gsl_tmp_91 * gsl_tmp_87) + gsl_tmp_89 * (gsl_tmp_90 * gsl_tmp_87 - gsl_tmp_91 * gsl_tmp_86)) / max(gsl_tmp_90 * gsl_tmp_92 - gsl_tmp_91 * gsl_tmp_91, 0.000001), vec3<f32>(-gsl_tmp_93), vec3<f32>(gsl_tmp_93));
  let gsl_tmp_19 = p - gsl_tmp_16;
  let gsl_tmp_18 = gsl_remap_4(gsl_tmp_16);
  let gsl_tmp_22 = dot(gsl_tmp_18, gsl_tmp_18);
  let gsl_tmp_17 = gsl_remap_3(gsl_tmp_16);
  let gsl_tmp_24 = dot(gsl_tmp_19, gsl_tmp_17);
  let gsl_tmp_21 = dot(gsl_tmp_17, gsl_tmp_18);
  let gsl_tmp_25 = dot(gsl_tmp_19, gsl_tmp_18);
  let gsl_tmp_20 = dot(gsl_tmp_17, gsl_tmp_17);
  let gsl_tmp_23 = max(gsl_tmp_20 * gsl_tmp_22 - gsl_tmp_21 * gsl_tmp_21, 0.000001);
  let lens = -(select(gsl_tmp_26, max(length(gsl_tmp_19), gsl_tmp_26), min(min((gsl_tmp_22 * gsl_tmp_24 - gsl_tmp_21 * gsl_tmp_25) / gsl_tmp_23, (gsl_tmp_20 * gsl_tmp_25 - gsl_tmp_21 * gsl_tmp_24) / gsl_tmp_23), gsl_tmp_26) > 0.0 && length(gsl_tmp_19) * length(gsl_tmp_17 + gsl_tmp_18) / 2.0 <= 1.5 * gsl_tmp_26 + 0.001) - 0.0);
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(lens, gsl_remap_9(p - vec3<f32>(4.0, 0.0, 0.0)))), 1u);
  return gsl_s1;
}
