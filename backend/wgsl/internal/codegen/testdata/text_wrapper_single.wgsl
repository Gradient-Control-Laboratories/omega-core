fn gsl_remap(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(5.0, 5.0)));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(0.0, 8.5), vec2<f32>(0.0, 9.0)));
}

fn text_segmentDist(p: vec3<f32>, a: vec2<f32>, b: vec2<f32>) -> f32 {
  let pa = p.xy - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = vec2<f32>(7.9997, min(1000000.0, gsl_remap(vec3<f32>((p.x - 0.0) / 1.1111112 - 0.0, (p.y - 0.0) / 1.1111112, p.z / 1.1111112))));
  let gsl_s1 = gsl_shape(min(gsl_tmp_1.y, gsl_remap_2(vec3<f32>((p.x - 0.0) / 1.1111112 - gsl_tmp_1.x, (p.y - 0.0) / 1.1111112, p.z / 1.1111112))) * 1.1111112 - 0.389, 1u);
  return gsl_s1;
}
