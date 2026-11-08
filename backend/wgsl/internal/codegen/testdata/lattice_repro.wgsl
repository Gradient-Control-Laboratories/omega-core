fn gsl_remap(p: vec3<f32>, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let gsl_msUvw = (p - sdfBMin0) / sdfBSize0;
  let gsl_msSample = textureSampleLevel(sdfTex0, sdfSmp0, clamp(gsl_msUvw, vec3<f32>(0.0), vec3<f32>(1.0)), 0.0).r;
  let gsl_msOutside = length(max(vec3<f32>(0.0), max(sdfBMin0 - p, p - (sdfBMin0 + sdfBSize0))));
  let gsl_msVal = select(gsl_msOutside + gsl_msSample, gsl_msSample, gsl_msOutside == 0.0);
  let meshDist = gsl_msVal;
  return meshDist;
}

fn cylinder(p: vec3<f32>, radius: f32, halfHeight: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let radialDist = length(p.xy) - radius;
  let heightDist = abs(p.z) - halfHeight;
  return max(radialDist, heightDist);
}

fn smin(a: f32, b: f32, k: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> f32 {
  let h = clamp(0.5 + 0.5 * (b - a) / k, 0.0, 1.0);
  return b + (a - b) * h - k * h * (1.0 - h);
}

fn sdShape(p: vec3<f32>, time: f32, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>) -> gsl_shape {
  let tree = min(gsl_remap(vec3<f32>(p.x - 1.0, p.y, p.z), sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0), cylinder(p, 1.1, 8.0, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0));
  let innerShell = tree + 0.01;
  let inner = smin(innerShell, innerShell, 0.02, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0);
  let gsl_s1 = gsl_shape(max(min(1e30, cylinder(p, 1.1, 8.0, sdfTex0, sdfSmp0, sdfBMin0, sdfBSize0)), -inner), 1u);
  return gsl_s1;
}
