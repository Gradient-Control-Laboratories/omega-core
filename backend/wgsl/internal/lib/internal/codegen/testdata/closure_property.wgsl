fn color_sampleMap(palette: array<vec3<f32>, 7>, t: f32) -> vec3<f32> {
  let tt = clamp(t, 0.0, 1.0);
  let weighted = array<vec3<f32>, 7>(palette[0] * max(0.0, 1.0 - abs(tt * 6.0 - 0.0)), palette[1] * max(0.0, 1.0 - abs(tt * 6.0 - 1.0)), palette[2] * max(0.0, 1.0 - abs(tt * 6.0 - 2.0)), palette[3] * max(0.0, 1.0 - abs(tt * 6.0 - 3.0)), palette[4] * max(0.0, 1.0 - abs(tt * 6.0 - 4.0)), palette[5] * max(0.0, 1.0 - abs(tt * 6.0 - 5.0)), palette[6] * max(0.0, 1.0 - abs(tt * 6.0 - 6.0)));
  return weighted[0] + weighted[1] + weighted[2] + weighted[3] + weighted[4] + weighted[5] + weighted[6];
}

fn color_cyclicRainbow(t: f32) -> vec3<f32> {
  return color_sampleMap(array<vec3<f32>, 7>(vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 0.0)), t);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-10.0, -10.0, -10.0);
  let gsl_partHi = vec3<f32>(10.0, 10.0, 10.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = length(p) - 10.0 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let ball = gsl_part;
  let gsl_s1 = gsl_shape(ball, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    var gsl_simpleValue = vec3<f32>(0.6604224, 0.6490635, 0.6588749);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue = mix(gsl_simpleValue, color_cyclicRainbow(fract(p.z / 10.0 - time / 5.0)), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
var<private> gsl_vanishEye: vec4<f32>;
var<private> gsl_vanishView: vec4<f32>;
var<private> gsl_partNear: vec3<f32>;

fn gsl_vanish(lo: vec3<f32>, hi: vec3<f32>) -> vec4<f32> {
  if (gsl_vanishEye.w <= 0.0) { return vec4<f32>(1.0, 0.0, 1e10, 1e10); }
  let size = hi - lo;
  let dist = length(gsl_vanishEye.xyz - clamp(gsl_vanishEye.xyz, lo, hi));
  let pixel = gsl_vanishEye.w * mix(max(dist, 1.0), 1.0, gsl_vanishView.x);
  let edge = min(size.x, min(size.y, size.z));
  let px = edge / pixel;
  let w = smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), px);
  return vec4<f32>(w, (1.0 - w) * 0.5 * edge, px, edge);
}

fn gsl_boxDistance(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  return length(max(max(lo - q, q - hi), vec3<f32>(0.0)));
}

var<private> gsl_simpleDetail: f32 = 1.0;
