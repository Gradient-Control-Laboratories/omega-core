fn map_axisAt(p: vec3<f32>, point: vec3<f32>, axis: vec3<f32>) -> vec3<f32> {
  let d = p - point;
  let n = normalize(axis);
  let sz = sign(n.z);
  let s = sz + 1.0 - abs(sz);
  let a = -1.0 / (s + n.z);
  let b = n.x * n.y * a;
  let t1 = vec3<f32>(1.0 + s * n.x * n.x * a, s * b, -(s * n.x));
  let t2 = vec3<f32>(b, s + n.y * n.y * a, -n.y);
  return vec3<f32>(dot(d, t1), dot(d, t2), dot(d, n));
}

fn fastener_seatedAt(p: vec3<f32>, seatPoint: vec3<f32>, shankDir: vec3<f32>) -> vec3<f32> {
  return map_axisAt(p, seatPoint, vec3<f32>(-shankDir.x, -shankDir.y, -shankDir.z));
}

fn catalog_onRow(key: f32, rowKey: f32, tol: f32) -> f32 {
  return 1.0 - clamp(sign(abs(key - rowKey) - tol), 0.0, 1.0);
}

fn catalog_pitch(nominal: f32) -> f32 {
  let k = nominal;
  return 0.25 * catalog_onRow(k, 1.0, 0.0001) + 0.25 * catalog_onRow(k, 1.2, 0.0001) + 0.3 * catalog_onRow(k, 1.4, 0.0001) + 0.35 * catalog_onRow(k, 1.6, 0.0001) + 0.35 * catalog_onRow(k, 1.8, 0.0001) + 0.4 * catalog_onRow(k, 2.0, 0.0001) + 0.45 * catalog_onRow(k, 2.2, 0.0001) + 0.45 * catalog_onRow(k, 2.5, 0.0001) + 0.5 * catalog_onRow(k, 3.0, 0.0001) + 0.6 * catalog_onRow(k, 3.5, 0.0001) + 0.7 * catalog_onRow(k, 4.0, 0.0001) + 0.75 * catalog_onRow(k, 4.5, 0.0001) + 0.8 * catalog_onRow(k, 5.0, 0.0001) + 1.0 * catalog_onRow(k, 6.0, 0.0001) + 1.0 * catalog_onRow(k, 7.0, 0.0001) + 1.25 * catalog_onRow(k, 8.0, 0.0001) + 1.5 * catalog_onRow(k, 10.0, 0.0001) + 1.75 * catalog_onRow(k, 12.0, 0.0001) + 2.0 * catalog_onRow(k, 14.0, 0.0001) + 2.0 * catalog_onRow(k, 16.0, 0.0001) + 2.5 * catalog_onRow(k, 18.0, 0.0001) + 2.5 * catalog_onRow(k, 20.0, 0.0001) + 2.5 * catalog_onRow(k, 22.0, 0.0001) + 3.0 * catalog_onRow(k, 24.0, 0.0001) + 3.0 * catalog_onRow(k, 27.0, 0.0001) + 3.5 * catalog_onRow(k, 30.0, 0.0001) + 3.5 * catalog_onRow(k, 33.0, 0.0001) + 4.0 * catalog_onRow(k, 36.0, 0.0001) + 4.0 * catalog_onRow(k, 39.0, 0.0001) + 4.5 * catalog_onRow(k, 42.0, 0.0001) + 4.5 * catalog_onRow(k, 45.0, 0.0001) + 5.0 * catalog_onRow(k, 48.0, 0.0001) + 5.0 * catalog_onRow(k, 52.0, 0.0001) + 5.5 * catalog_onRow(k, 56.0, 0.0001) + 5.5 * catalog_onRow(k, 60.0, 0.0001) + 6.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn thread_metricPitch(nominal: f32) -> f32 {
  return catalog_pitch(nominal);
}

fn catalog_shcsLength(minLength: f32) -> f32 {
  let m = minLength;
  let l = min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(2.5 + 1000000.0 * clamp(sign(m - 2.5001), 0.0, 1.0), 3.0 + 1000000.0 * clamp(sign(m - 3.0001), 0.0, 1.0)), 4.0 + 1000000.0 * clamp(sign(m - 4.0001), 0.0, 1.0)), 5.0 + 1000000.0 * clamp(sign(m - 5.0001), 0.0, 1.0)), 6.0 + 1000000.0 * clamp(sign(m - 6.0001), 0.0, 1.0)), 8.0 + 1000000.0 * clamp(sign(m - 8.0001), 0.0, 1.0)), 10.0 + 1000000.0 * clamp(sign(m - 10.0001), 0.0, 1.0)), 12.0 + 1000000.0 * clamp(sign(m - 12.0001), 0.0, 1.0)), 16.0 + 1000000.0 * clamp(sign(m - 16.0001), 0.0, 1.0)), 20.0 + 1000000.0 * clamp(sign(m - 20.0001), 0.0, 1.0)), 25.0 + 1000000.0 * clamp(sign(m - 25.0001), 0.0, 1.0)), 30.0 + 1000000.0 * clamp(sign(m - 30.0001), 0.0, 1.0)), 35.0 + 1000000.0 * clamp(sign(m - 35.0001), 0.0, 1.0)), 40.0 + 1000000.0 * clamp(sign(m - 40.0001), 0.0, 1.0)), 45.0 + 1000000.0 * clamp(sign(m - 45.0001), 0.0, 1.0)), 50.0 + 1000000.0 * clamp(sign(m - 50.0001), 0.0, 1.0)), 55.0 + 1000000.0 * clamp(sign(m - 55.0001), 0.0, 1.0)), 60.0 + 1000000.0 * clamp(sign(m - 60.0001), 0.0, 1.0)), 65.0 + 1000000.0 * clamp(sign(m - 65.0001), 0.0, 1.0)), 70.0 + 1000000.0 * clamp(sign(m - 70.0001), 0.0, 1.0)), 80.0 + 1000000.0 * clamp(sign(m - 80.0001), 0.0, 1.0)), 90.0 + 1000000.0 * clamp(sign(m - 90.0001), 0.0, 1.0)), 100.0 + 1000000.0 * clamp(sign(m - 100.0001), 0.0, 1.0)), 110.0 + 1000000.0 * clamp(sign(m - 110.0001), 0.0, 1.0)), 120.0 + 1000000.0 * clamp(sign(m - 120.0001), 0.0, 1.0)), 130.0 + 1000000.0 * clamp(sign(m - 130.0001), 0.0, 1.0)), 140.0 + 1000000.0 * clamp(sign(m - 140.0001), 0.0, 1.0)), 150.0 + 1000000.0 * clamp(sign(m - 150.0001), 0.0, 1.0)), 160.0 + 1000000.0 * clamp(sign(m - 160.0001), 0.0, 1.0)), 180.0 + 1000000.0 * clamp(sign(m - 180.0001), 0.0, 1.0)), 200.0 + 1000000.0 * clamp(sign(m - 200.0001), 0.0, 1.0)), 220.0 + 1000000.0 * clamp(sign(m - 220.0001), 0.0, 1.0)), 240.0 + 1000000.0 * clamp(sign(m - 240.0001), 0.0, 1.0)), 260.0 + 1000000.0 * clamp(sign(m - 260.0001), 0.0, 1.0)), 280.0 + 1000000.0 * clamp(sign(m - 280.0001), 0.0, 1.0)), 300.0 + 1000000.0 * clamp(sign(m - 300.0001), 0.0, 1.0));
  return l * clamp(sign(300.0001 - l), 0.0, 1.0);
}

fn thread_formDepth(pitchVal: f32) -> f32 {
  return 0.5412659 * pitchVal;
}

fn thread_helixPhase(p: vec3<f32>, pitchVal: f32) -> f32 {
  let ang = 6.2831855 * p.z / max(pitchVal, 0.01);
  let radial = max(length(p.xy), 0.2);
  let along = cos(ang) * p.x / radial + sin(ang) * p.y / radial;
  return acos(clamp(along, -1.0, 1.0)) / 3.1415927;
}

fn thread_profileDepth(p: vec3<f32>, pitchVal: f32) -> f32 {
  let offset = thread_helixPhase(p, pitchVal) * pitchVal * 0.5;
  return clamp((offset - pitchVal / 16.0) * 1.7320508, 0.0, thread_formDepth(pitchVal));
}

fn thread_runout(p: vec3<f32>, zTop: f32, zBottom: f32, outerR: f32) -> f32 {
  return max(0.0, 0.5 * min(min(zTop - p.z, p.z - zBottom), outerR - length(p.xy)));
}

fn thread_maleRelief(p: vec3<f32>, major: f32, pitchVal: f32, zTop: f32, zBottom: f32) -> f32 {
  let outerR = major / 2.0 + thread_formDepth(pitchVal);
  let innerR = major / 2.0 - 2.0 * thread_formDepth(pitchVal);
  return -min(min(thread_profileDepth(p, pitchVal), thread_runout(p, zTop, zBottom, outerR)), max(0.0, length(p.xy) - innerR));
}

fn color_srgbToLinearChannel(s: f32) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x), color_srgbToLinearChannel(c.y), color_srgbToLinearChannel(c.z));
}

fn color_cbrt(x: f32) -> f32 {
  return sign(x) * pow(abs(x), 0.33333334);
}

fn color_rgb2oklab(c: vec3<f32>) -> vec3<f32> {
  let lin = color_linearize(c);
  let l = dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), lin);
  let l_ = color_cbrt(l);
  let m = dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), lin);
  let m_ = color_cbrt(m);
  let s = dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), lin);
  let s_ = color_cbrt(s);
  let lms = vec3<f32>(l_, m_, s_);
  return vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), lms), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), lms), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), lms));
}

fn color_linearToSrgbChannel(l: f32) -> f32 {
  return select(1.055 * pow(l, 0.41666666) - 0.055, l * 12.92, l <= 0.0031308);
}

fn color_gammaEncode(c: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(color_linearToSrgbChannel(c.x), color_linearToSrgbChannel(c.y), color_linearToSrgbChannel(c.z));
}

fn color_oklab2rgb(c: vec3<f32>) -> vec3<f32> {
  let l_ = c.x + 0.39633778 * c.y + 0.21580376 * c.z;
  let l = l_ * l_ * l_;
  let m_ = c.x - 0.105561346 * c.y - 0.06385417 * c.z;
  let m = m_ * m_ * m_;
  let s_ = c.x - 0.08948418 * c.y - 1.2914855 * c.z;
  let s = s_ * s_ * s_;
  let lms = vec3<f32>(l, m, s);
  let lin = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), lms), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), lms), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), lms));
  return color_gammaEncode(lin);
}

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32) -> vec3<f32> {
  let la = color_rgb2oklab(a);
  let lb = color_rgb2oklab(b);
  return color_oklab2rgb(la + (lb - la) * t);
}

fn thread_reliefColor(base: vec3<f32>, height: f32, pitchVal: f32) -> vec3<f32> {
  var gsl_simpleValue = base;
  if (gsl_simpleDetail > 0.0) {
    let depth = clamp(-height / thread_formDepth(pitchVal), 0.0, 1.0);
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(base, base * 0.35, depth * (1.0 - gsl_dispDrawn)), gsl_simpleDetail);
  }
  let painted = gsl_simpleValue;
  return painted;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn thread_minorDia(major: f32, pitchVal: f32) -> f32 {
  return major - 2.0 * thread_formDepth(pitchVal);
}

fn thread_femaleRelief_2(p: vec3<f32>, major: f32, pitchVal: f32, boreDia: f32, zTop: f32, zBottom: f32) -> f32 {
  let majorR = major / 2.0;
  let reach = majorR - boreDia / 2.0;
  let cut = max(reach - thread_profileDepth(p, pitchVal), 0.0);
  return -min(cut, thread_runout(p, zTop, zBottom, majorR + 2.0 * reach));
}

fn thread_femaleRelief(p: vec3<f32>, major: f32, pitchVal: f32, zTop: f32, zBottom: f32) -> f32 {
  return thread_femaleRelief_2(p, major, pitchVal, thread_minorDia(major, pitchVal), zTop, zBottom);
}

fn catalog_shcsHeadDia(nominal: f32) -> f32 {
  let k = nominal;
  return 0.0 * catalog_onRow(k, 1.0, 0.0001) + 0.0 * catalog_onRow(k, 1.2, 0.0001) + 0.0 * catalog_onRow(k, 1.4, 0.0001) + 3.0 * catalog_onRow(k, 1.6, 0.0001) + 0.0 * catalog_onRow(k, 1.8, 0.0001) + 3.8 * catalog_onRow(k, 2.0, 0.0001) + 0.0 * catalog_onRow(k, 2.2, 0.0001) + 4.5 * catalog_onRow(k, 2.5, 0.0001) + 5.5 * catalog_onRow(k, 3.0, 0.0001) + 0.0 * catalog_onRow(k, 3.5, 0.0001) + 7.0 * catalog_onRow(k, 4.0, 0.0001) + 0.0 * catalog_onRow(k, 4.5, 0.0001) + 8.5 * catalog_onRow(k, 5.0, 0.0001) + 10.0 * catalog_onRow(k, 6.0, 0.0001) + 0.0 * catalog_onRow(k, 7.0, 0.0001) + 13.0 * catalog_onRow(k, 8.0, 0.0001) + 16.0 * catalog_onRow(k, 10.0, 0.0001) + 18.0 * catalog_onRow(k, 12.0, 0.0001) + 21.0 * catalog_onRow(k, 14.0, 0.0001) + 24.0 * catalog_onRow(k, 16.0, 0.0001) + 0.0 * catalog_onRow(k, 18.0, 0.0001) + 30.0 * catalog_onRow(k, 20.0, 0.0001) + 0.0 * catalog_onRow(k, 22.0, 0.0001) + 36.0 * catalog_onRow(k, 24.0, 0.0001) + 0.0 * catalog_onRow(k, 27.0, 0.0001) + 45.0 * catalog_onRow(k, 30.0, 0.0001) + 0.0 * catalog_onRow(k, 33.0, 0.0001) + 54.0 * catalog_onRow(k, 36.0, 0.0001) + 0.0 * catalog_onRow(k, 39.0, 0.0001) + 63.0 * catalog_onRow(k, 42.0, 0.0001) + 0.0 * catalog_onRow(k, 45.0, 0.0001) + 72.0 * catalog_onRow(k, 48.0, 0.0001) + 0.0 * catalog_onRow(k, 52.0, 0.0001) + 84.0 * catalog_onRow(k, 56.0, 0.0001) + 0.0 * catalog_onRow(k, 60.0, 0.0001) + 96.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn fastener_cylindricalHead_2(p: vec3<f32>, dk: f32, k: f32, edge: f32) -> f32 {
  let head = max(max(max(length(p.xy) - dk / 2.0, -p.z), p.z - k), (length(p.xy) - dk / 2.0 + p.z - k + edge) * 0.7071);
  return head;
}

fn fastener_cylindricalHead(p: vec3<f32>, dk: f32, k: f32) -> f32 {
  return fastener_cylindricalHead_2(p, dk, k, 0.2);
}

fn fastener_plainShank(p: vec3<f32>, dia: f32, shankLength: f32, zTop: f32) -> f32 {
  let form = max(max(max(length(p.xy) - dia / 2.0, -shankLength - p.z), p.z - zTop), (length(p.xy) - dia / 2.0 + -p.z - shankLength + 0.15 * dia) * 0.7071);
  return form;
}

fn catalog_shcsSocket(nominal: f32) -> f32 {
  let k = nominal;
  return 0.0 * catalog_onRow(k, 1.0, 0.0001) + 0.0 * catalog_onRow(k, 1.2, 0.0001) + 0.0 * catalog_onRow(k, 1.4, 0.0001) + 1.5 * catalog_onRow(k, 1.6, 0.0001) + 0.0 * catalog_onRow(k, 1.8, 0.0001) + 1.5 * catalog_onRow(k, 2.0, 0.0001) + 0.0 * catalog_onRow(k, 2.2, 0.0001) + 2.0 * catalog_onRow(k, 2.5, 0.0001) + 2.5 * catalog_onRow(k, 3.0, 0.0001) + 0.0 * catalog_onRow(k, 3.5, 0.0001) + 3.0 * catalog_onRow(k, 4.0, 0.0001) + 0.0 * catalog_onRow(k, 4.5, 0.0001) + 4.0 * catalog_onRow(k, 5.0, 0.0001) + 5.0 * catalog_onRow(k, 6.0, 0.0001) + 0.0 * catalog_onRow(k, 7.0, 0.0001) + 6.0 * catalog_onRow(k, 8.0, 0.0001) + 8.0 * catalog_onRow(k, 10.0, 0.0001) + 10.0 * catalog_onRow(k, 12.0, 0.0001) + 12.0 * catalog_onRow(k, 14.0, 0.0001) + 14.0 * catalog_onRow(k, 16.0, 0.0001) + 0.0 * catalog_onRow(k, 18.0, 0.0001) + 17.0 * catalog_onRow(k, 20.0, 0.0001) + 0.0 * catalog_onRow(k, 22.0, 0.0001) + 19.0 * catalog_onRow(k, 24.0, 0.0001) + 0.0 * catalog_onRow(k, 27.0, 0.0001) + 22.0 * catalog_onRow(k, 30.0, 0.0001) + 0.0 * catalog_onRow(k, 33.0, 0.0001) + 27.0 * catalog_onRow(k, 36.0, 0.0001) + 0.0 * catalog_onRow(k, 39.0, 0.0001) + 32.0 * catalog_onRow(k, 42.0, 0.0001) + 0.0 * catalog_onRow(k, 45.0, 0.0001) + 36.0 * catalog_onRow(k, 48.0, 0.0001) + 0.0 * catalog_onRow(k, 52.0, 0.0001) + 41.0 * catalog_onRow(k, 56.0, 0.0001) + 0.0 * catalog_onRow(k, 60.0, 0.0001) + 46.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn catalog_shcsSocketDepth(nominal: f32) -> f32 {
  let k = nominal;
  return 0.7 * catalog_onRow(k, 1.6, 0.0001) + 1.0 * catalog_onRow(k, 2.0, 0.0001) + 1.1 * catalog_onRow(k, 2.5, 0.0001) + 1.3 * catalog_onRow(k, 3.0, 0.0001) + 2.0 * catalog_onRow(k, 4.0, 0.0001) + 2.5 * catalog_onRow(k, 5.0, 0.0001) + 3.0 * catalog_onRow(k, 6.0, 0.0001) + 4.0 * catalog_onRow(k, 8.0, 0.0001) + 5.0 * catalog_onRow(k, 10.0, 0.0001) + 6.0 * catalog_onRow(k, 12.0, 0.0001) + 7.0 * catalog_onRow(k, 14.0, 0.0001) + 8.0 * catalog_onRow(k, 16.0, 0.0001) + 10.0 * catalog_onRow(k, 20.0, 0.0001) + 12.0 * catalog_onRow(k, 24.0, 0.0001) + 15.5 * catalog_onRow(k, 30.0, 0.0001) + 19.0 * catalog_onRow(k, 36.0, 0.0001) + 24.0 * catalog_onRow(k, 42.0, 0.0001) + 28.0 * catalog_onRow(k, 48.0, 0.0001) + 34.0 * catalog_onRow(k, 56.0, 0.0001) + 38.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn drill_hexXY(p: vec3<f32>, acrossFlats: f32) -> f32 {
  let h = acrossFlats / 2.0;
  return max(max(abs(p.x) - h, abs(0.5 * p.x + 0.8660254 * p.y) - h), abs(0.5 * p.x - 0.8660254 * p.y) - h);
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn drill_hexSocketAt(p: vec3<f32>, zTop: f32, acrossFlats: f32, depth: f32) -> f32 {
  return drill_prismVoid(p, zTop, drill_hexXY(p, acrossFlats), depth);
}

fn fastener_screw(head: f32, shank: f32, drive: f32) -> f32 {
  return max(min(head, shank), -drive);
}

fn fastener_shcs(p: vec3<f32>, nominalDia: f32, shankLength: f32) -> f32 {
  let headH = nominalDia;
  let form = fastener_screw(fastener_cylindricalHead(p, catalog_shcsHeadDia(nominalDia), headH), fastener_plainShank(p, nominalDia, shankLength, headH / 2.0), drill_hexSocketAt(p, headH, catalog_shcsSocket(nominalDia), catalog_shcsSocketDepth(nominalDia)));
  return form;
}

fn catalog_counterbore(nominal: f32) -> f32 {
  let k = nominal;
  return 0.0 * catalog_onRow(k, 1.0, 0.0001) + 0.0 * catalog_onRow(k, 1.2, 0.0001) + 0.0 * catalog_onRow(k, 1.4, 0.0001) + 3.5 * catalog_onRow(k, 1.6, 0.0001) + 0.0 * catalog_onRow(k, 1.8, 0.0001) + 4.4 * catalog_onRow(k, 2.0, 0.0001) + 0.0 * catalog_onRow(k, 2.2, 0.0001) + 5.5 * catalog_onRow(k, 2.5, 0.0001) + 6.5 * catalog_onRow(k, 3.0, 0.0001) + 0.0 * catalog_onRow(k, 3.5, 0.0001) + 8.0 * catalog_onRow(k, 4.0, 0.0001) + 0.0 * catalog_onRow(k, 4.5, 0.0001) + 10.0 * catalog_onRow(k, 5.0, 0.0001) + 11.0 * catalog_onRow(k, 6.0, 0.0001) + 0.0 * catalog_onRow(k, 7.0, 0.0001) + 15.0 * catalog_onRow(k, 8.0, 0.0001) + 18.0 * catalog_onRow(k, 10.0, 0.0001) + 20.0 * catalog_onRow(k, 12.0, 0.0001) + 24.0 * catalog_onRow(k, 14.0, 0.0001) + 26.0 * catalog_onRow(k, 16.0, 0.0001) + 0.0 * catalog_onRow(k, 18.0, 0.0001) + 33.0 * catalog_onRow(k, 20.0, 0.0001) + 0.0 * catalog_onRow(k, 22.0, 0.0001) + 40.0 * catalog_onRow(k, 24.0, 0.0001) + 0.0 * catalog_onRow(k, 27.0, 0.0001) + 50.0 * catalog_onRow(k, 30.0, 0.0001) + 0.0 * catalog_onRow(k, 33.0, 0.0001) + 58.0 * catalog_onRow(k, 36.0, 0.0001) + 0.0 * catalog_onRow(k, 39.0, 0.0001) + 0.0 * catalog_onRow(k, 42.0, 0.0001) + 0.0 * catalog_onRow(k, 45.0, 0.0001) + 0.0 * catalog_onRow(k, 48.0, 0.0001) + 0.0 * catalog_onRow(k, 52.0, 0.0001) + 0.0 * catalog_onRow(k, 56.0, 0.0001) + 0.0 * catalog_onRow(k, 60.0, 0.0001) + 0.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn catalog_clearance_2(nominal: f32) -> f32 {
  let k = nominal;
  return 1.2 * catalog_onRow(k, 1.0, 0.0001) + 1.4 * catalog_onRow(k, 1.2, 0.0001) + 1.6 * catalog_onRow(k, 1.4, 0.0001) + 1.8 * catalog_onRow(k, 1.6, 0.0001) + 2.1 * catalog_onRow(k, 1.8, 0.0001) + 2.4 * catalog_onRow(k, 2.0, 0.0001) + 0.0 * catalog_onRow(k, 2.2, 0.0001) + 2.9 * catalog_onRow(k, 2.5, 0.0001) + 3.4 * catalog_onRow(k, 3.0, 0.0001) + 3.9 * catalog_onRow(k, 3.5, 0.0001) + 4.5 * catalog_onRow(k, 4.0, 0.0001) + 5.0 * catalog_onRow(k, 4.5, 0.0001) + 5.5 * catalog_onRow(k, 5.0, 0.0001) + 6.6 * catalog_onRow(k, 6.0, 0.0001) + 7.6 * catalog_onRow(k, 7.0, 0.0001) + 9.0 * catalog_onRow(k, 8.0, 0.0001) + 11.0 * catalog_onRow(k, 10.0, 0.0001) + 13.5 * catalog_onRow(k, 12.0, 0.0001) + 15.5 * catalog_onRow(k, 14.0, 0.0001) + 17.5 * catalog_onRow(k, 16.0, 0.0001) + 20.0 * catalog_onRow(k, 18.0, 0.0001) + 22.0 * catalog_onRow(k, 20.0, 0.0001) + 24.0 * catalog_onRow(k, 22.0, 0.0001) + 26.0 * catalog_onRow(k, 24.0, 0.0001) + 30.0 * catalog_onRow(k, 27.0, 0.0001) + 33.0 * catalog_onRow(k, 30.0, 0.0001) + 36.0 * catalog_onRow(k, 33.0, 0.0001) + 39.0 * catalog_onRow(k, 36.0, 0.0001) + 42.0 * catalog_onRow(k, 39.0, 0.0001) + 45.0 * catalog_onRow(k, 42.0, 0.0001) + 48.0 * catalog_onRow(k, 45.0, 0.0001) + 52.0 * catalog_onRow(k, 48.0, 0.0001) + 56.0 * catalog_onRow(k, 52.0, 0.0001) + 62.0 * catalog_onRow(k, 56.0, 0.0001) + 66.0 * catalog_onRow(k, 60.0, 0.0001) + 70.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn catalog_clearance(nominal: f32, fit: f32) -> f32 {
  let k = nominal;
  let fine = 1.1 * catalog_onRow(k, 1.0, 0.0001) + 1.3 * catalog_onRow(k, 1.2, 0.0001) + 1.5 * catalog_onRow(k, 1.4, 0.0001) + 1.7 * catalog_onRow(k, 1.6, 0.0001) + 2.0 * catalog_onRow(k, 1.8, 0.0001) + 2.2 * catalog_onRow(k, 2.0, 0.0001) + 2.7 * catalog_onRow(k, 2.5, 0.0001) + 3.2 * catalog_onRow(k, 3.0, 0.0001) + 3.7 * catalog_onRow(k, 3.5, 0.0001) + 4.3 * catalog_onRow(k, 4.0, 0.0001) + 4.8 * catalog_onRow(k, 4.5, 0.0001) + 5.3 * catalog_onRow(k, 5.0, 0.0001) + 6.4 * catalog_onRow(k, 6.0, 0.0001) + 7.4 * catalog_onRow(k, 7.0, 0.0001) + 8.4 * catalog_onRow(k, 8.0, 0.0001) + 10.5 * catalog_onRow(k, 10.0, 0.0001) + 13.0 * catalog_onRow(k, 12.0, 0.0001) + 15.0 * catalog_onRow(k, 14.0, 0.0001) + 17.0 * catalog_onRow(k, 16.0, 0.0001) + 19.0 * catalog_onRow(k, 18.0, 0.0001) + 21.0 * catalog_onRow(k, 20.0, 0.0001) + 23.0 * catalog_onRow(k, 22.0, 0.0001) + 25.0 * catalog_onRow(k, 24.0, 0.0001) + 28.0 * catalog_onRow(k, 27.0, 0.0001) + 31.0 * catalog_onRow(k, 30.0, 0.0001) + 34.0 * catalog_onRow(k, 33.0, 0.0001) + 37.0 * catalog_onRow(k, 36.0, 0.0001) + 40.0 * catalog_onRow(k, 39.0, 0.0001) + 43.0 * catalog_onRow(k, 42.0, 0.0001) + 46.0 * catalog_onRow(k, 45.0, 0.0001) + 50.0 * catalog_onRow(k, 48.0, 0.0001) + 54.0 * catalog_onRow(k, 52.0, 0.0001) + 58.0 * catalog_onRow(k, 56.0, 0.0001) + 62.0 * catalog_onRow(k, 60.0, 0.0001) + 66.0 * catalog_onRow(k, 64.0, 0.0001);
  let f = fit;
  let coarse = 1.3 * catalog_onRow(k, 1.0, 0.0001) + 1.5 * catalog_onRow(k, 1.2, 0.0001) + 1.8 * catalog_onRow(k, 1.4, 0.0001) + 2.0 * catalog_onRow(k, 1.6, 0.0001) + 2.2 * catalog_onRow(k, 1.8, 0.0001) + 2.6 * catalog_onRow(k, 2.0, 0.0001) + 3.1 * catalog_onRow(k, 2.5, 0.0001) + 3.6 * catalog_onRow(k, 3.0, 0.0001) + 4.2 * catalog_onRow(k, 3.5, 0.0001) + 4.8 * catalog_onRow(k, 4.0, 0.0001) + 5.3 * catalog_onRow(k, 4.5, 0.0001) + 5.8 * catalog_onRow(k, 5.0, 0.0001) + 7.0 * catalog_onRow(k, 6.0, 0.0001) + 8.0 * catalog_onRow(k, 7.0, 0.0001) + 10.0 * catalog_onRow(k, 8.0, 0.0001) + 12.0 * catalog_onRow(k, 10.0, 0.0001) + 14.5 * catalog_onRow(k, 12.0, 0.0001) + 16.5 * catalog_onRow(k, 14.0, 0.0001) + 18.5 * catalog_onRow(k, 16.0, 0.0001) + 21.0 * catalog_onRow(k, 18.0, 0.0001) + 24.0 * catalog_onRow(k, 20.0, 0.0001) + 26.0 * catalog_onRow(k, 22.0, 0.0001) + 28.0 * catalog_onRow(k, 24.0, 0.0001) + 32.0 * catalog_onRow(k, 27.0, 0.0001) + 35.0 * catalog_onRow(k, 30.0, 0.0001) + 38.0 * catalog_onRow(k, 33.0, 0.0001) + 42.0 * catalog_onRow(k, 36.0, 0.0001) + 45.0 * catalog_onRow(k, 39.0, 0.0001) + 48.0 * catalog_onRow(k, 42.0, 0.0001) + 52.0 * catalog_onRow(k, 45.0, 0.0001) + 56.0 * catalog_onRow(k, 48.0, 0.0001) + 62.0 * catalog_onRow(k, 52.0, 0.0001) + 66.0 * catalog_onRow(k, 56.0, 0.0001) + 70.0 * catalog_onRow(k, 60.0, 0.0001) + 74.0 * catalog_onRow(k, 64.0, 0.0001);
  return fine * catalog_onRow(f, 1.0, 0.0001) + catalog_clearance_2(nominal) * catalog_onRow(f, 2.0, 0.0001) + coarse * catalog_onRow(f, 3.0, 0.0001);
}

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn builtin_cylinder(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, radius: f32) -> f32 {
  return builtin_axis(p, center, direction) - radius;
}

fn builtin_cylinderZ(p: vec3<f32>, vector_dirZ: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return builtin_cylinder(p, center, vector_dirZ, radius);
}

fn drill_gcl_closure_bore(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, dia: f32, depth: f32, zTop: f32) -> f32 {
  return max(builtin_cylinderZ(p, vector_dirZ, vector_zero, dia / 2.0), zTop - depth - p.z);
}

fn drill_gcl_closure_thru(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, dia: f32, zTop: f32) -> f32 {
  return builtin_cylinderZ(p, vector_dirZ, vector_zero, dia / 2.0);
}

fn fastener_gcl_closure_hole_2(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, bored: f32, clear: f32, zTop: f32) -> f32 {
  let mouth = zTop;
  return min(min(100000.0, drill_gcl_closure_bore(p, vector_zero, vector_dirZ, bored, 0.0, mouth)), drill_gcl_closure_thru(p, vector_zero, vector_dirZ, clear, mouth));
}

fn fastener_shcsHole(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, nominalDia: f32, fit: f32) -> f32 {
  let bored = catalog_counterbore(nominalDia);
  let clear = catalog_clearance(nominalDia, fit);
  return fastener_gcl_closure_hole_2(p, vector_zero, vector_dirZ, bored, clear, 0.0);
}

fn builtin_cone(p: vec3<f32>, origin: vec3<f32>, direction: vec3<f32>, angle: f32) -> f32 {
  let offset = p - origin;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = length(offset - unitDirection * along);
  return perp * cos(angle) - along * sin(angle);
}

fn drill_drillVoid(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, zTop: f32, dia: f32, depth: f32, included: f32) -> f32 {
  let half = included / 2.0;
  let apex = zTop - depth - dia / 2.0 / tan(half);
  return max(builtin_cylinderZ(p, vector_dirZ, vector_zero, dia / 2.0), builtin_cone(p, vec3<f32>(0.0, 0.0, apex), vector_dirZ, half));
}

fn drill_gcl_closure_blind(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, dia: f32, depth: f32, zTop: f32) -> f32 {
  return drill_drillVoid(p, vector_zero, vector_dirZ, zTop, dia, depth, 2.0594885);
}

fn fastener_gcl_closure_hole_3(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, minor: f32, depth: f32, zTop: f32) -> f32 {
  let mouth = zTop;
  return min(100000.0, drill_gcl_closure_blind(p, vector_zero, vector_dirZ, minor, depth, mouth));
}

fn fastener_tappedHole(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, nominalDia: f32, shankLength: f32) -> f32 {
  let pitchVal = thread_metricPitch(nominalDia);
  let minor = thread_minorDia(nominalDia, pitchVal);
  let depth = shankLength + 3.0 * pitchVal;
  return fastener_gcl_closure_hole_3(p, vector_zero, vector_dirZ, minor, depth, 0.0);
}

fn catalog_cskLength(minLength: f32) -> f32 {
  let m = minLength;
  let l = min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(min(8.0 + 1000000.0 * clamp(sign(m - 8.0001), 0.0, 1.0), 10.0 + 1000000.0 * clamp(sign(m - 10.0001), 0.0, 1.0)), 12.0 + 1000000.0 * clamp(sign(m - 12.0001), 0.0, 1.0)), 16.0 + 1000000.0 * clamp(sign(m - 16.0001), 0.0, 1.0)), 20.0 + 1000000.0 * clamp(sign(m - 20.0001), 0.0, 1.0)), 25.0 + 1000000.0 * clamp(sign(m - 25.0001), 0.0, 1.0)), 30.0 + 1000000.0 * clamp(sign(m - 30.0001), 0.0, 1.0)), 35.0 + 1000000.0 * clamp(sign(m - 35.0001), 0.0, 1.0)), 40.0 + 1000000.0 * clamp(sign(m - 40.0001), 0.0, 1.0)), 45.0 + 1000000.0 * clamp(sign(m - 45.0001), 0.0, 1.0)), 50.0 + 1000000.0 * clamp(sign(m - 50.0001), 0.0, 1.0)), 55.0 + 1000000.0 * clamp(sign(m - 55.0001), 0.0, 1.0)), 60.0 + 1000000.0 * clamp(sign(m - 60.0001), 0.0, 1.0)), 65.0 + 1000000.0 * clamp(sign(m - 65.0001), 0.0, 1.0)), 70.0 + 1000000.0 * clamp(sign(m - 70.0001), 0.0, 1.0)), 80.0 + 1000000.0 * clamp(sign(m - 80.0001), 0.0, 1.0)), 90.0 + 1000000.0 * clamp(sign(m - 90.0001), 0.0, 1.0)), 100.0 + 1000000.0 * clamp(sign(m - 100.0001), 0.0, 1.0));
  return l * clamp(sign(100.0001 - l), 0.0, 1.0);
}

fn catalog_cskHeadDia(nominal: f32) -> f32 {
  let k = nominal;
  return 0.0 * catalog_onRow(k, 1.0, 0.0001) + 0.0 * catalog_onRow(k, 1.2, 0.0001) + 0.0 * catalog_onRow(k, 1.4, 0.0001) + 0.0 * catalog_onRow(k, 1.6, 0.0001) + 0.0 * catalog_onRow(k, 1.8, 0.0001) + 0.0 * catalog_onRow(k, 2.0, 0.0001) + 0.0 * catalog_onRow(k, 2.2, 0.0001) + 0.0 * catalog_onRow(k, 2.5, 0.0001) + 6.72 * catalog_onRow(k, 3.0, 0.0001) + 0.0 * catalog_onRow(k, 3.5, 0.0001) + 8.96 * catalog_onRow(k, 4.0, 0.0001) + 0.0 * catalog_onRow(k, 4.5, 0.0001) + 11.2 * catalog_onRow(k, 5.0, 0.0001) + 13.44 * catalog_onRow(k, 6.0, 0.0001) + 0.0 * catalog_onRow(k, 7.0, 0.0001) + 17.92 * catalog_onRow(k, 8.0, 0.0001) + 22.4 * catalog_onRow(k, 10.0, 0.0001) + 26.88 * catalog_onRow(k, 12.0, 0.0001) + 30.8 * catalog_onRow(k, 14.0, 0.0001) + 33.6 * catalog_onRow(k, 16.0, 0.0001) + 0.0 * catalog_onRow(k, 18.0, 0.0001) + 40.32 * catalog_onRow(k, 20.0, 0.0001) + 0.0 * catalog_onRow(k, 22.0, 0.0001) + 0.0 * catalog_onRow(k, 24.0, 0.0001) + 0.0 * catalog_onRow(k, 27.0, 0.0001) + 0.0 * catalog_onRow(k, 30.0, 0.0001) + 0.0 * catalog_onRow(k, 33.0, 0.0001) + 0.0 * catalog_onRow(k, 36.0, 0.0001) + 0.0 * catalog_onRow(k, 39.0, 0.0001) + 0.0 * catalog_onRow(k, 42.0, 0.0001) + 0.0 * catalog_onRow(k, 45.0, 0.0001) + 0.0 * catalog_onRow(k, 48.0, 0.0001) + 0.0 * catalog_onRow(k, 52.0, 0.0001) + 0.0 * catalog_onRow(k, 56.0, 0.0001) + 0.0 * catalog_onRow(k, 60.0, 0.0001) + 0.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn fastener_countersunkHead_2(p: vec3<f32>, dk: f32, dSmall: f32, included: f32) -> f32 {
  let half = included / 2.0;
  let slope = tan(half);
  let depth = (dk - dSmall) / (2.0 * slope);
  let head = max(max((length(p.xy) - (dk / 2.0 + p.z * slope)) * cos(half), -depth - p.z), p.z);
  return head;
}

fn fastener_countersunkHead(p: vec3<f32>, dk: f32, dSmall: f32) -> f32 {
  return fastener_countersunkHead_2(p, dk, dSmall, 1.5707963);
}

fn fastener_plainShank_2(p: vec3<f32>, dia: f32, shankLength: f32) -> f32 {
  return fastener_plainShank(p, dia, shankLength, 0.0);
}

fn catalog_cskSocket(nominal: f32) -> f32 {
  let k = nominal;
  return 2.0 * catalog_onRow(k, 3.0, 0.0001) + 2.5 * catalog_onRow(k, 4.0, 0.0001) + 3.0 * catalog_onRow(k, 5.0, 0.0001) + 4.0 * catalog_onRow(k, 6.0, 0.0001) + 5.0 * catalog_onRow(k, 8.0, 0.0001) + 6.0 * catalog_onRow(k, 10.0, 0.0001) + 8.0 * catalog_onRow(k, 12.0, 0.0001) + 10.0 * catalog_onRow(k, 14.0, 0.0001) + 10.0 * catalog_onRow(k, 16.0, 0.0001) + 12.0 * catalog_onRow(k, 20.0, 0.0001);
}

fn catalog_cskSocketDepth(nominal: f32) -> f32 {
  let k = nominal;
  return 1.1 * catalog_onRow(k, 3.0, 0.0001) + 1.5 * catalog_onRow(k, 4.0, 0.0001) + 1.9 * catalog_onRow(k, 5.0, 0.0001) + 2.2 * catalog_onRow(k, 6.0, 0.0001) + 3.0 * catalog_onRow(k, 8.0, 0.0001) + 3.6 * catalog_onRow(k, 10.0, 0.0001) + 4.3 * catalog_onRow(k, 12.0, 0.0001) + 4.5 * catalog_onRow(k, 14.0, 0.0001) + 4.8 * catalog_onRow(k, 16.0, 0.0001) + 5.6 * catalog_onRow(k, 20.0, 0.0001);
}

fn fastener_countersunk(p: vec3<f32>, nominalDia: f32, overallLength: f32) -> f32 {
  let form = fastener_screw(fastener_countersunkHead(p, catalog_cskHeadDia(nominalDia), nominalDia), fastener_plainShank_2(p, nominalDia, overallLength), drill_hexSocketAt(p, 0.0, catalog_cskSocket(nominalDia), catalog_cskSocketDepth(nominalDia)));
  return form;
}

fn drill_coneVoid(p: vec3<f32>, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, plane: f32, dia: f32, included: f32) -> f32 {
  let half = included / 2.0;
  let apex = plane - dia / 2.0 / tan(half);
  return max(builtin_cone(p, vec3<f32>(0.0, 0.0, apex), vector_dirZ, half), builtin_cylinderZ(p, vector_dirZ, vector_zero, dia / 2.0));
}

fn drill_gcl_closure_countersink(p: vec3<f32>, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, dia: f32, included: f32, zTop: f32) -> f32 {
  return drill_coneVoid(p, vector_dirZ, vector_zero, zTop, dia, included);
}

fn fastener_gcl_closure_hole(p: vec3<f32>, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, rim: f32, clear: f32, zTop: f32) -> f32 {
  let mouth = zTop;
  return min(min(100000.0, drill_gcl_closure_countersink(p, vector_dirZ, vector_zero, rim, 1.5707963, mouth)), drill_gcl_closure_thru(p, vector_zero, vector_dirZ, clear, mouth));
}

fn fastener_countersunkHole(p: vec3<f32>, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, nominalDia: f32, fit: f32) -> f32 {
  let rim = catalog_cskHeadDia(nominalDia);
  let clear = catalog_clearance(nominalDia, fit);
  return fastener_gcl_closure_hole(p, vector_dirZ, vector_zero, rim, clear, 0.0);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn joint_seat(iface: mat2x4<f32>) -> vec3<f32> {
  let r = iface[0];
  return r.xyz;
}

fn joint_direction(iface: mat2x4<f32>) -> vec3<f32> {
  let r = iface[1];
  return r.xyz;
}

fn joint_diameter(iface: mat2x4<f32>) -> f32 {
  let r = iface[0];
  return r.w;
}

fn joint_boltedLength(nominalDia: f32, grip: f32) -> f32 {
  return catalog_shcsLength(grip + 1.5 * nominalDia);
}

fn joint_insert__color__0__spec1(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> vec3<f32> {
  let c = vec3<f32>(0.77, 0.78, 0.78);
  let steelColor = c;
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let len = joint_boltedLength(nominalDia, grip);
  return thread_reliefColor(steelColor, thread_maleRelief(at, nominalDia, pitchVal, 0.0, 0.0 - len), pitchVal);
}

fn joint_insert__material__0__spec1(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> mat4x4<f32> {
  let steelMat = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let len = joint_boltedLength(nominalDia, grip);
  return material_withDisplacement(steelMat, thread_maleRelief(at, nominalDia, pitchVal, 0.0, 0.0 - len));
}

fn joint_insert__color__2__spec1(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> vec3<f32> {
  let c_3 = vec3<f32>(0.77, 0.78, 0.78);
  let steelColor = c_3;
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let len = joint_boltedLength(nominalDia, grip);
  let tapDepth = len + 3.0 * pitchVal;
  return thread_reliefColor(steelColor, thread_femaleRelief(at, nominalDia, pitchVal, 0.0 - grip, 0.0 - tapDepth), pitchVal);
}

fn joint_insert__material__2__spec1(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> mat4x4<f32> {
  let steelMat = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let len = joint_boltedLength(nominalDia, grip);
  let tapDepth = len + 3.0 * pitchVal;
  return material_withDisplacement(steelMat, thread_femaleRelief(at, nominalDia, pitchVal, 0.0 - grip, 0.0 - tapDepth));
}

fn joint_fit(iface: mat2x4<f32>) -> f32 {
  let r = iface[1];
  return r.w;
}

fn joint_insert__spec1(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> gsl_struct {
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let len = joint_boltedLength(nominalDia, grip);
  let screw = fastener_shcs(at, nominalDia, len);
  let fitSeries = joint_fit(iface);
  let held = max(clamped, -fastener_shcsHole(at, vector_zero, vector_dirZ, nominalDia, fitSeries));
  let tapped = max(threaded, -fastener_tappedHole(at, vector_zero, vector_dirZ, nominalDia, len));
  var gsl_result : gsl_struct;
  gsl_result.v_1 = screw;
  gsl_result.v_2 = held;
  gsl_result.v_3 = tapped;
  return gsl_result;
}

fn joint_countersunkLength(nominalDia: f32, grip: f32) -> f32 {
  return catalog_cskLength(grip + 1.5 * nominalDia);
}

fn joint_insert__color__0__spec2(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> vec3<f32> {
  let c_6 = vec3<f32>(0.77, 0.78, 0.78);
  let steelColor = c_6;
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let threadTop = 0.0 - (catalog_cskHeadDia(nominalDia) - nominalDia) / 2.0;
  let len = joint_countersunkLength(nominalDia, grip);
  return thread_reliefColor(steelColor, thread_maleRelief(at, nominalDia, pitchVal, threadTop, 0.0 - len), pitchVal);
}

fn joint_insert__material__0__spec2(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> mat4x4<f32> {
  let steelMat = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let threadTop = 0.0 - (catalog_cskHeadDia(nominalDia) - nominalDia) / 2.0;
  let len = joint_countersunkLength(nominalDia, grip);
  return material_withDisplacement(steelMat, thread_maleRelief(at, nominalDia, pitchVal, threadTop, 0.0 - len));
}

fn joint_insert__color__2__spec2(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> vec3<f32> {
  let c_8 = vec3<f32>(0.77, 0.78, 0.78);
  let steelColor = c_8;
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let len = joint_countersunkLength(nominalDia, grip);
  let tapDepth = len + 3.0 * pitchVal;
  return thread_reliefColor(steelColor, thread_femaleRelief(at, nominalDia, pitchVal, 0.0 - grip, 0.0 - tapDepth), pitchVal);
}

fn joint_insert__material__2__spec2(p: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> mat4x4<f32> {
  let steelMat = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let pitchVal = thread_metricPitch(nominalDia);
  let len = joint_countersunkLength(nominalDia, grip);
  let tapDepth = len + 3.0 * pitchVal;
  return material_withDisplacement(steelMat, thread_femaleRelief(at, nominalDia, pitchVal, 0.0 - grip, 0.0 - tapDepth));
}

fn joint_insert__spec2(p: vec3<f32>, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, iface: mat2x4<f32>, clamped: f32, threaded: f32, grip: f32) -> gsl_struct {
  let seatPoint = joint_seat(iface);
  let shankDir = joint_direction(iface);
  let at = fastener_seatedAt(p, seatPoint, shankDir);
  let nominalDia = joint_diameter(iface);
  let len = joint_countersunkLength(nominalDia, grip);
  let screw = fastener_countersunk(at, nominalDia, len);
  let fitSeries = joint_fit(iface);
  let held = max(clamped, -fastener_countersunkHole(at, vector_dirZ, vector_zero, nominalDia, fitSeries));
  let tapped = max(threaded, -fastener_tappedHole(at, vector_zero, vector_dirZ, nominalDia, len));
  var gsl_result : gsl_struct;
  gsl_result.v_1 = screw;
  gsl_result.v_2 = held;
  gsl_result.v_3 = tapped;
  return gsl_result;
}

fn gsl_shapeParts(p: vec3<f32>, time: f32) -> array<f32, 8> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-40.0, -20.0, -25.0);
  let gsl_partHi = vec3<f32>(40.0, 20.0, 0.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
  let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
  let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
  let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
  let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
  let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
  let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
  let lid1 = gsl_result.v_2;
  let base1 = gsl_result.v_3;
  let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
  let base2 = gsl_result_2.v_3;
  let gsl_result_3 = joint_insert__spec2(p, vector_dirZ, vector_zero, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
  let cover1 = gsl_result_3.v_2;
  let base3 = gsl_result_3.v_3;
  let gsl_result_4 = joint_insert__spec2(p, vector_dirZ, vector_zero, mat2x4<f32>(vec4<f32>(30.0, -13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), cover1, base3, 3.0);
  let gsl_part_base = gsl_result_4.v_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_part_base + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let base = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-38.0, -18.0, 0.0);
  let gsl_partHi_2 = vec3<f32>(-2.0, 18.0, 14.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  let gsl_part_lid = gsl_result_2.v_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_part_lid + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let lid = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(2.0, -18.0, 0.0);
  let gsl_partHi_3 = vec3<f32>(38.0, 18.0, 3.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  let gsl_part_cover = gsl_result_4.v_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = gsl_part_cover + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let cover = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(25.0, -8.0, -10.0);
  let gsl_partHi_4 = vec3<f32>(55.0, 8.0, 2.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  let gsl_tmp_12 = p - vec3<f32>(0.0, 0.0, -4.0);
  let gsl_tmp_13 = abs(dot(gsl_tmp_12, vec3<f32>(0.0, 1.0, 0.0)));
  let gsl_tmp_14 = abs(dot(gsl_tmp_12, vec3<f32>(0.0, 0.0, 1.0)));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(80.0, 0.0, -8.0) * 0.5, abs(vec3<f32>(30.0, 16.0, 12.0))), max(gsl_tmp_13 - 5.0, gsl_tmp_14 - 3.0)) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let slider = gsl_part_4;
  let lidScrew1 = gsl_result.v_1;
  let lidScrew2 = gsl_result_2.v_1;
  let coverScrew1 = gsl_result_3.v_1;
  let coverScrew2 = gsl_result_4.v_1;
  return array<f32, 8>(base, lid, cover, slider, lidScrew1, lidScrew2, coverScrew1, coverScrew2);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7], 8u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    let base2 = gsl_result_2.v_3;
    let gsl_result_3 = joint_insert__spec2(p, vector_dirZ, vector_zero, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
    let cover1 = gsl_result_3.v_2;
    let base3 = gsl_result_3.v_3;
    let gslCarry15 = joint_insert__color__2__spec2(p, mat2x4<f32>(vec4<f32>(30.0, -13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), cover1, base3, 3.0);
    return vec4<f32>(gslCarry15, 1.0);
  }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.15, 0.35, 0.65), 1.0); }
  if (gsl_colorIdx == 5u) {
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gslCarry1 = joint_insert__color__0__spec1(p, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    return vec4<f32>(gslCarry1, 1.0);
  }
  if (gsl_colorIdx == 6u) {
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gslCarry5 = joint_insert__color__0__spec1(p, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    return vec4<f32>(gslCarry5, 1.0);
  }
  if (gsl_colorIdx == 7u) {
    let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    let base2 = gsl_result_2.v_3;
    let gslCarry9 = joint_insert__color__0__spec2(p, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
    return vec4<f32>(gslCarry9, 1.0);
  }
  if (gsl_colorIdx == 8u) {
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    let base2 = gsl_result_2.v_3;
    let gsl_result_3 = joint_insert__spec2(p, vector_dirZ, vector_zero, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
    let cover1 = gsl_result_3.v_2;
    let base3 = gsl_result_3.v_3;
    let gslCarry13 = joint_insert__color__0__spec2(p, mat2x4<f32>(vec4<f32>(30.0, -13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), cover1, base3, 3.0);
    return vec4<f32>(gslCarry13, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    let base2 = gsl_result_2.v_3;
    let gsl_result_3 = joint_insert__spec2(p, vector_dirZ, vector_zero, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
    let cover1 = gsl_result_3.v_2;
    let base3 = gsl_result_3.v_3;
    let gslCarry16 = joint_insert__material__2__spec2(p, mat2x4<f32>(vec4<f32>(30.0, -13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), cover1, base3, 3.0);
    let gsl_m = gslCarry16; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) {
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gslCarry2 = joint_insert__material__0__spec1(p, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let gsl_m = gslCarry2; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gslCarry6 = joint_insert__material__0__spec1(p, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    let gsl_m = gslCarry6; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 7u) {
    let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    let base2 = gsl_result_2.v_3;
    let gslCarry10 = joint_insert__material__0__spec2(p, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
    let gsl_m = gslCarry10; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 8u) {
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
    let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
    let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
    let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
    let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
    let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
    let lid1 = gsl_result.v_2;
    let base1 = gsl_result.v_3;
    let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
    let base2 = gsl_result_2.v_3;
    let gsl_result_3 = joint_insert__spec2(p, vector_dirZ, vector_zero, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
    let cover1 = gsl_result_3.v_2;
    let base3 = gsl_result_3.v_3;
    let gslCarry14 = joint_insert__material__0__spec2(p, mat2x4<f32>(vec4<f32>(30.0, -13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), cover1, base3, 3.0);
    let gsl_m = gslCarry14; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
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

var<private> gsl_dispDrawn: f32 = 0.0;

fn gsl_heights(p: vec3<f32>, time: f32) -> array<f32, 8> {
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let coverBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(40.0, 0.0, 3.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 3.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
  let lidBlank = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(-40.0, 0.0, 14.0) * 0.5, abs(vec3<f32>(36.0, 36.0, 14.0))), -dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0)));
  let gsl_tmp_5 = p - vec3<f32>(0.0, 0.0, -4.0);
  let gsl_tmp_6 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 1.0, 0.0)));
  let gsl_tmp_7 = abs(dot(gsl_tmp_5, vec3<f32>(0.0, 0.0, 1.0)));
  let baseBlank = max(max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -25.0) * 0.5, abs(vec3<f32>(80.0, 40.0, 25.0))), dot(p - vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 1.0))), -max(gsl_tmp_6 - 5.3, gsl_tmp_7 - 3.3));
  let gsl_result = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
  let lid1 = gsl_result.v_2;
  let base1 = gsl_result.v_3;
  let gsl_result_2 = joint_insert__spec1(p, vector_zero, vector_dirZ, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
  let base2 = gsl_result_2.v_3;
  let gsl_result_3 = joint_insert__spec2(p, vector_dirZ, vector_zero, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
  let cover1 = gsl_result_3.v_2;
  let base3 = gsl_result_3.v_3;
  let gslCarry16 = joint_insert__material__2__spec2(p, mat2x4<f32>(vec4<f32>(30.0, -13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), cover1, base3, 3.0);
  let gslCarry2 = joint_insert__material__0__spec1(p, mat2x4<f32>(vec4<f32>(-20.0, 12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lidBlank, baseBlank, 7.5);
  let gslCarry6 = joint_insert__material__0__spec1(p, mat2x4<f32>(vec4<f32>(-20.0, -12.0, 7.5, 6.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), lid1, base1, 7.5);
  let gslCarry10 = joint_insert__material__0__spec2(p, mat2x4<f32>(vec4<f32>(30.0, 13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), coverBlank, base2, 3.0);
  let gslCarry14 = joint_insert__material__0__spec2(p, mat2x4<f32>(vec4<f32>(30.0, -13.0, 3.0, 3.0), vec4<f32>(-0.0, -0.0, -1.0, 2.0)), cover1, base3, 3.0);
  return array<f32, 8>(gslCarry16[2].z, 0.0, 0.0, 0.0, gslCarry2[2].z, gslCarry6[2].z, gslCarry10[2].z, gslCarry14[2].z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_h = gsl_heights(p, time);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4] - gsl_h[4] * gsl_dispK, 5u);
  let gsl_s6 = gsl_shape(gsl_d[5] - gsl_h[5] * gsl_dispK, 6u);
  let gsl_s7 = gsl_shape(gsl_d[6] - gsl_h[6] * gsl_dispK, 7u);
  let gsl_s8 = gsl_shape(gsl_d[7] - gsl_h[7] * gsl_dispK, 8u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : f32,
}
