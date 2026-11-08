fn gsl_remap(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0)));
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
  let gsl_tmp_1 = vec2<f32>(6.9997, min(1000000.0, gsl_remap(vec3<f32>((p.x - 0.0) / 0.055555556 - 0.0, (p.y - 0.0) / 0.055555556, p.z / 0.055555556))));
  let gsl_tmp_2 = vec2<f32>(gsl_tmp_1.x + 6.9997, min(gsl_tmp_1.y, gsl_remap_2(vec3<f32>((p.x - 0.0) / 0.055555556 - gsl_tmp_1.x, (p.y - 0.0) / 0.055555556, p.z / 0.055555556))));
  let gsl_tmp_3 = vec2<f32>(gsl_tmp_2.x + 3.9997, min(gsl_tmp_2.y, gsl_remap_3(vec3<f32>((p.x - 0.0) / 0.055555556 - gsl_tmp_2.x, (p.y - 0.0) / 0.055555556, p.z / 0.055555556))));
  let gsl_tmp_4 = vec2<f32>(gsl_tmp_3.x + 3.9997, min(gsl_tmp_3.y, gsl_remap_3(vec3<f32>((p.x - 0.0) / 0.055555556 - gsl_tmp_3.x, (p.y - 0.0) / 0.055555556, p.z / 0.055555556))));
  let gsl_s1 = gsl_shape(min(gsl_tmp_4.y, gsl_remap_4(vec3<f32>((p.x - 0.0) / 0.055555556 - gsl_tmp_4.x, (p.y - 0.0) / 0.055555556, p.z / 0.055555556))) * 0.055555556 - 0.01945, 1u);
  return gsl_s1;
}
