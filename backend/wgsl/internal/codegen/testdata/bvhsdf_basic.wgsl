fn sdShape(p: vec3<f32>, time: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_shape {
  let gsl_msUvw = (p - sdfBMin0) / sdfBSize0;
  let gsl_msSample = textureSampleLevel(sdfTex0, sdfSmp0, clamp(gsl_msUvw, vec3<f32>(0.0), vec3<f32>(1.0)), 0.0).r;
  let gsl_msOutside = length(max(vec3<f32>(0.0), max(sdfBMin0 - p, p - (sdfBMin0 + sdfBSize0))));
  let gsl_msVal = select(gsl_msOutside + gsl_msSample, gsl_msSample, gsl_msOutside == 0.0);
  let gsl_s1 = gsl_shape(gsl_msVal - 0.1, 1u);
  return gsl_s1;
}
