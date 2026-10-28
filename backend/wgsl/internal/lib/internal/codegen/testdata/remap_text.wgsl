fn gsl_remap(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5)));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 0.0, 1.5708)), text_arcDist(p, 2.0, 3.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 5.0), vec2<f32>(1.0, 5.0)));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 6.001, 3.8035, 6.001, 2.618, 0.5235), text_segmentDist(p, vec2<f32>(0.0, 3.804), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 3.0))), text_arcDist(p, 2.0, 3.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 5.0), vec2<f32>(0.1204, 5.0)));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let gsl_tmp_1 = vec2<f32>(8.9997, min(1000000.0, gsl_remap(vec3<f32>((p.x - 0.0) / 1.1111112 - 0.0, (p.y - -5.0) / 1.1111112, p.z / 1.1111112))));
  let gsl_tmp_2 = vec2<f32>(gsl_tmp_1.x + 6.9997, min(gsl_tmp_1.y, gsl_remap_2(vec3<f32>((p.x - 0.0) / 1.1111112 - gsl_tmp_1.x, (p.y - -5.0) / 1.1111112, p.z / 1.1111112))));
  return min(gsl_tmp_2.y, gsl_remap_3(vec3<f32>((p.x - 0.0) / 1.1111112 - gsl_tmp_2.x, (p.y - -5.0) / 1.1111112, p.z / 1.1111112))) * 1.1111112 - 0.389;
}

fn text_segmentDist(p: vec3<f32>, a: vec2<f32>, b: vec2<f32>) -> f32 {
  let pa = p.xy - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn text_arcDist(p: vec3<f32>, cx: f32, cy: f32, ra: f32, mid: f32, half: f32) -> f32 {
  let dx = p.x - cx;
  let rot = 1.5707964 - mid;
  let ca = cos(rot);
  let dy = p.y - cy;
  let sa = sin(rot);
  let rx = dx * ca - dy * sa;
  let pxAbs = abs(rx);
  let ry = dx * sa + dy * ca;
  let l = length(vec2<f32>(pxAbs, ry));
  let dOutside = abs(l - ra);
  let scx = sin(half);
  let wx = pxAbs - ra * scx;
  let scy = cos(half);
  let wy = ry - ra * scy;
  let dInside = length(vec2<f32>(wx, wy));
  let inside = scy * pxAbs > scx * ry;
  return select(dOutside, dInside, inside);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(gsl_remap_4(p.yzx), 1u);
  return gsl_s1;
}
