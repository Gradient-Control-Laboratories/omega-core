fn tpms_sum3(v: vec3<f32>) -> f32 {
  return v.x + v.y + v.z;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = p * vec3<f32>(0.62831855, 0.62831855, 0.62831855);
  let gyr = tpms_sum3(vec3<f32>(1.0, 1.0, 1.0) * sin(gsl_tmp_1) * cos(vec3<f32>(gsl_tmp_1.y, gsl_tmp_1.z, gsl_tmp_1.x))) * 30.0 / 18.0 - 1.0;
  let gsl_tmp_2 = p * vec3<f32>(0.7853982, 0.7853982, 0.7853982);
  let sch = tpms_sum3(vec3<f32>(1.0, 1.0, 2.0) * cos(vec3<f32>(gsl_tmp_2.y, gsl_tmp_2.z, gsl_tmp_2.x)) * (sin(gsl_tmp_2) * 0.5 - 0.5)) * 24.0 / 36.0 - 1.0;
  let gsl_tmp_3 = p * vec3<f32>(0.5235988, 0.5235988, 0.5235988);
  let gsl_tmp_4 = sin(gsl_tmp_3);
  let gsl_tmp_5 = cos(vec3<f32>(gsl_tmp_3.y, gsl_tmp_3.z, gsl_tmp_3.x));
  let dia = (1.0 * gsl_tmp_4.x * gsl_tmp_4.y * gsl_tmp_4.z + tpms_sum3(vec3<f32>(1.0, 1.0, 1.0) * gsl_tmp_4 * gsl_tmp_5 * (vec3<f32>(gsl_tmp_5.y, gsl_tmp_5.z, gsl_tmp_5.x) * 1.0 + 0.0))) * 36.0 / 33.941124 - 1.0;
  let gsl_tmp_6 = p * vec3<f32>(0.62831855, 0.62831855, 0.62831855);
  let gsl_tmp_7 = cos(gsl_tmp_6 * 2.0);
  let lid = (tpms_sum3(vec3<f32>(1.0, 1.0, 1.0) * sin(vec3<f32>(gsl_tmp_6.z, gsl_tmp_6.x, gsl_tmp_6.y)) * cos(vec3<f32>(gsl_tmp_6.y, gsl_tmp_6.z, gsl_tmp_6.x)) * (sin(gsl_tmp_6 * 2.0) * 1.0 + 0.0)) - 1.0 * tpms_sum3(gsl_tmp_7 * vec3<f32>(gsl_tmp_7.y, gsl_tmp_7.z, gsl_tmp_7.x))) * 30.0 / 72.0 - 1.0;
  let gsl_tmp_8 = cos(p * vec3<f32>(0.6981317, 0.6981317, 0.6981317)) * vec3<f32>(1.0, 1.0, 1.0);
  let neo = (-tpms_sum3(gsl_tmp_8) - 1.3333334 * gsl_tmp_8.x * gsl_tmp_8.y * gsl_tmp_8.z) * 27.0 / 52.0 - 1.0;
  let gsl_tmp_9 = p * vec3<f32>(0.57119864, 0.57119864, 0.57119864);
  let iwp = (2.0 * tpms_sum3(vec3<f32>(1.0, 1.0, 1.0) * cos(gsl_tmp_9) * cos(vec3<f32>(gsl_tmp_9.y, gsl_tmp_9.z, gsl_tmp_9.x))) - tpms_sum3(cos(gsl_tmp_9 * 2.0))) * 33.0 / 48.0 - 1.0;
  let gsl_tmp_10 = p * vec3<f32>(0.62831855, 0.62831855, 0.62831855);
  let gsl_tmp_11 = sin(gsl_tmp_10 * 2.0);
  let gsl_tmp_12 = cos(gsl_tmp_10 * 2.0);
  let spl = (-1.0 * tpms_sum3(gsl_tmp_11 * cos(vec3<f32>(gsl_tmp_10.y, gsl_tmp_10.z, gsl_tmp_10.x)) * sin(vec3<f32>(gsl_tmp_10.z, gsl_tmp_10.x, gsl_tmp_10.y))) + 0.5 * tpms_sum3(gsl_tmp_11 * gsl_tmp_12) + 0.25 * tpms_sum3(gsl_tmp_12)) * 30.0 / 36.0 - 1.0;
  let gsl_s1 = gsl_shape(min(min(min(gyr, sch), min(dia, lid)), min(min(neo, iwp), spl)), 1u);
  return gsl_s1;
}
