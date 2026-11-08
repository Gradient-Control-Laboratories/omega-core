fn torusY(p: vec3<f32>, pos: vec3<f32>, R: f32, r: f32) -> f32 {
  let q = p - pos;
  let t = vec2<f32>(length(q.xz) - R, q.y);
  return length(t) - r;
}

fn capsuleSDF(p: vec3<f32>, a: vec3<f32>, b: vec3<f32>, r: f32) -> f32 {
  let pa = p - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h) - r;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let seatTube = capsuleSDF(p, vec3<f32>(-0.2, 0.0, 0.25), vec3<f32>(-0.25, 0.0, 0.7), 0.025);
  let downTube = capsuleSDF(p, vec3<f32>(-0.2, 0.0, 0.25), vec3<f32>(0.3, 0.0, 0.65), 0.03);
  let topTube = capsuleSDF(p, vec3<f32>(-0.25, 0.0, 0.7), vec3<f32>(0.35, 0.0, 0.75), 0.025);
  let headTube = capsuleSDF(p, vec3<f32>(0.3, 0.0, 0.65), vec3<f32>(0.35, 0.0, 0.75), 0.035);
  let mainFrame = min(seatTube, min(downTube, min(topTube, headTube)));
  let chainStay_l = capsuleSDF(p, vec3<f32>(-0.5, -0.065, 0.35), vec3<f32>(-0.2, -0.08, 0.25), 0.025);
  let chainStay_r = capsuleSDF(p, vec3<f32>(-0.5, 0.065, 0.35), vec3<f32>(-0.2, 0.08, 0.25), 0.025);
  let chainStays = min(chainStay_l, chainStay_r);
  let seatStay_l = capsuleSDF(p, vec3<f32>(-0.5, -0.065, 0.35), vec3<f32>(-0.25, 0.0, 0.7), 0.02);
  let seatStay_r = capsuleSDF(p, vec3<f32>(-0.5, 0.065, 0.35), vec3<f32>(-0.25, 0.0, 0.7), 0.02);
  let seatStays = min(seatStay_l, seatStay_r);
  let rearFrame = min(chainStays, seatStays);
  let seat = capsuleSDF(p, vec3<f32>(-0.33, 0.0, 0.85), vec3<f32>(-0.13, 0.0, 0.85), 0.04);
  let seatPost = capsuleSDF(p, vec3<f32>(-0.25, 0.0, 0.7), vec3<f32>(-0.23, 0.0, 0.85), 0.025);
  let seatAssembly = min(seat, seatPost);
  let fork_l = capsuleSDF(p, vec3<f32>(0.5, -0.065, 0.35), vec3<f32>(0.3, 0.0, 0.65), 0.025);
  let fork_r = capsuleSDF(p, vec3<f32>(0.5, 0.065, 0.35), vec3<f32>(0.3, 0.0, 0.65), 0.025);
  let fork = min(fork_l, fork_r);
  let frame = min(min(min(mainFrame, rearFrame), seatAssembly), fork);
  let stem = capsuleSDF(p, vec3<f32>(0.35, 0.0, 0.75), vec3<f32>(0.45, 0.0, 0.8), 0.025);
  let handlebars = capsuleSDF(p, vec3<f32>(0.45, -0.2, 0.8), vec3<f32>(0.45, 0.2, 0.8), 0.0225);
  let handlebarAssembly = min(stem, handlebars);
  let rearWheel = torusY(p, vec3<f32>(-0.5, 0.0, 0.35), 0.35, 0.03);
  let frontWheel = torusY(p, vec3<f32>(0.5, 0.0, 0.35), 0.35, 0.03);
  let wheels = min(rearWheel, frontWheel);
  let gsl_s1 = gsl_shape(frame, 1u);
  let gsl_s2 = gsl_shape(handlebarAssembly, 2u);
  let gsl_s3 = gsl_shape(wheels, 3u);
  return gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3);
}
