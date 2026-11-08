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
  let glyphs = array<f32, 5>(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854)), min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854)), min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))));
  let glyphs2 = array<f32, 5>(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854)), min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854)), min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))));
  let gsl_s1 = gsl_shape(glyphs[0], 1u);
  let gsl_s2 = gsl_shape(glyphs2[0], 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
