fn sdShape(p: vec3<f32>, time: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>, sdfTex1: texture_3d<f32>, sdfSmp1: sampler, sdfBMin1: vec3<f32>, sdfBSize1: vec3<f32>) -> gsl_shape {
  let gsl_msUvw = (p - sdfBMin0) / sdfBSize0;
  let gsl_msSample = textureSampleLevel(sdfTex0, sdfSmp0, clamp(gsl_msUvw, vec3<f32>(0.0), vec3<f32>(1.0)), 0.0).r;
  let gsl_msOutside = length(max(vec3<f32>(0.0), max(sdfBMin0 - p, p - (sdfBMin0 + sdfBSize0))));
  let gsl_msVal = select(gsl_msOutside + gsl_msSample, gsl_msSample, gsl_msOutside == 0.0);
  let gsl_msUvw_2 = (p - sdfBMin1) / sdfBSize1;
  let gsl_msSample_2 = textureSampleLevel(sdfTex1, sdfSmp1, clamp(gsl_msUvw_2, vec3<f32>(0.0), vec3<f32>(1.0)), 0.0).r;
  let gsl_msOutside_2 = length(max(vec3<f32>(0.0), max(sdfBMin1 - p, p - (sdfBMin1 + sdfBSize1))));
  let gsl_msVal_2 = select(gsl_msOutside_2 + gsl_msSample_2, gsl_msSample_2, gsl_msOutside_2 == 0.0);
  let gsl_s1 = gsl_shape(min(gsl_msVal, gsl_msVal_2), 1u);
  return gsl_s1;
}
