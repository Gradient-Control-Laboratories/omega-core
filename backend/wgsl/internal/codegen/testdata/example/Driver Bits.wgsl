fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn map_patternPolarXY(p: vec3<f32>, count: f32) -> vec3<f32> {
  let rad = length(p.xy);
  let ang = atan2(p.y, p.x);
  let wedge = 6.2831855 / count;
  let halfWedge = wedge * 0.5;
  let angmod = gsl_floored((ang + halfWedge), wedge) - halfWedge;
  return vec3<f32>(rad * cos(angmod), rad * sin(angmod), p.z);
}

fn drill_hexXY(p: vec3<f32>, acrossFlats: f32) -> f32 {
  let h = acrossFlats / 2.0;
  return max(max(abs(p.x) - h, abs(0.5 * p.x + 0.8660254 * p.y) - h), abs(0.5 * p.x - 0.8660254 * p.y) - h);
}

fn drill_hexSocketAt(p: vec3<f32>, zTop: f32, acrossFlats: f32, depth: f32) -> f32 {
  return drill_prismVoid(p, zTop, drill_hexXY(p, acrossFlats), depth);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  let ph2 = drill_phillipsGauge(p, 0.0, 2.78 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.78 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 4.37 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 6.74 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 8.34 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.64 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.001 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.539 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.497 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 3.574 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.315 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.513 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.102 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.098 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.738 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.813 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.286 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 3.81 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.08 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.25 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), radians(0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 138.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 140.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 146.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 153.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0))), radians(7.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0))), 0.225 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)));
  return min(ph2, driver_insertShank(p, 4.37, 0.7));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return gsl_remap(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let gsl_tmp_17 = length(p.xy);
  let gsl_tmp_18 = abs(p.x);
  let gsl_tmp_20 = (0.851 + (p.z - 0.3245) * 0.10069471) * 0.70710677;
  let gsl_tmp_19 = abs(p.y);
  let pz2 = max(max(max(max(gsl_tmp_17 * 0.89493436 - (p.z - 0.3245) * 0.4461978 - 1.0797383, (0.3245 - p.z) * 0.95105654 + (gsl_tmp_17 - 1.2065) * 0.309017), 0.0 - p.z), p.z - 4.37), min(min(abs(p.y) - 0.495, abs(p.x) - 0.495), max((gsl_tmp_18 - gsl_tmp_20) * 0.99939084 + (gsl_tmp_19 - gsl_tmp_20) * 0.0348995, (gsl_tmp_19 - gsl_tmp_20) * 0.99939084 + (gsl_tmp_18 - gsl_tmp_20) * 0.0348995) / 1.002708));
  return min(pz2, driver_insertShank(p, 4.37, 0.9));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return gsl_remap_3(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let t25 = max(drill_torxAt(p, 5.0, 0.9 * (1.0 - clamp(sign(23.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(22.9999), 0.0, 1.0)) + 1.2 * (1.0 - clamp(sign(21.9999), 0.0, 1.0)) + 1.35 * (1.0 - clamp(sign(20.9999), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(18.9999), 0.0, 1.0)) + 2.1 * (1.0 - clamp(sign(17.9999), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(16.9999), 0.0, 1.0)) + 2.6 * (1.0 - clamp(sign(15.9999), 0.0, 1.0)) + 2.8 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 3.35 * (1.0 - clamp(sign(9.9999), 0.0, 1.0)) + 3.95 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 4.5 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 5.1 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 5.6 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 6.75 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 7.93 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 8.95 * (1.0 - clamp(sign(24.9999), 0.0, 1.0)) + 11.35 * (1.0 - clamp(sign(29.9999), 0.0, 1.0)) + 13.45 * (1.0 - clamp(sign(34.9999), 0.0, 1.0)) + 15.7 * (1.0 - clamp(sign(44.9999), 0.0, 1.0)) + 17.75 * (1.0 - clamp(sign(54.9999), 0.0, 1.0)) + 20.2 * (1.0 - clamp(sign(64.9999), 0.0, 1.0)) + 22.4 * (1.0 - clamp(sign(74.9999), 0.0, 1.0)), 0.6 * (1.0 - clamp(sign(23.9999), 0.0, 1.0)) + 0.7 * (1.0 - clamp(sign(22.9999), 0.0, 1.0)) + 0.85 * (1.0 - clamp(sign(21.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(20.9999), 0.0, 1.0)) + 1.1 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(18.9999), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(17.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(16.9999), 0.0, 1.0)) + 1.9 * (1.0 - clamp(sign(15.9999), 0.0, 1.0)) + 2.05 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(9.9999), 0.0, 1.0)) + 2.85 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 3.25 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 3.68 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 4.05 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 4.85 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 5.64 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 6.45 * (1.0 - clamp(sign(24.9999), 0.0, 1.0)) + 8.05 * (1.0 - clamp(sign(29.9999), 0.0, 1.0)) + 9.6 * (1.0 - clamp(sign(34.9999), 0.0, 1.0)) + 11.2 * (1.0 - clamp(sign(44.9999), 0.0, 1.0)) + 12.8 * (1.0 - clamp(sign(54.9999), 0.0, 1.0)) + 14.4 * (1.0 - clamp(sign(64.9999), 0.0, 1.0)) + 16.0 * (1.0 - clamp(sign(74.9999), 0.0, 1.0)), 5.0), p.z - 5.0);
  return min(t25, driver_insertShank(p, 5.0, 1.6));
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  return gsl_remap_5(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  let r2 = max(max(max(max(max(abs(p.x) - 1.42, abs(p.y) - 1.42), 0.0 - p.z), (abs(p.x) - 1.42 + 0.0 - p.z + 1.42) * 0.70710677), (abs(p.y) - 1.42 + 0.0 - p.z + 1.42) * 0.70710677), p.z - 5.0);
  return min(r2, driver_insertShank(p, 5.0, 1.4));
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  return gsl_remap_7(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  let hex4 = driver_hexTip(p, 4.0, 5.0);
  return min(hex4, driver_insertShank(p, 5.0, 1.8));
}

fn gsl_remap_10(p: vec3<f32>) -> f32 {
  return gsl_remap_9(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_11(p: vec3<f32>) -> f32 {
  let slot55 = driver_slotTip(p, 1.0, 5.5, 5.0);
  return min(slot55, driver_insertShank(p, 5.0, 0.5));
}

fn gsl_remap_12(p: vec3<f32>) -> f32 {
  return gsl_remap_11(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_13(p: vec3<f32>) -> f32 {
  let ph2 = drill_phillipsGauge(p, 0.0, 2.78 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.78 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 4.37 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 6.74 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 8.34 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.64 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.001 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.539 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.497 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 3.574 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.315 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.513 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.102 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.098 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.738 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.813 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.286 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 3.81 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.08 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.25 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), radians(0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 138.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 140.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 146.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 153.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0))), radians(7.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0))), 0.225 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)));
  return min(ph2, driver_powerShank(p, 4.37, 0.7));
}

fn gsl_remap_14(p: vec3<f32>) -> f32 {
  return gsl_remap_13(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_15(p: vec3<f32>) -> f32 {
  let gsl_tmp_17 = length(p.xy);
  let gsl_tmp_18 = abs(p.x);
  let gsl_tmp_20 = (0.851 + (p.z - 0.3245) * 0.10069471) * 0.70710677;
  let gsl_tmp_19 = abs(p.y);
  let pz2 = max(max(max(max(gsl_tmp_17 * 0.89493436 - (p.z - 0.3245) * 0.4461978 - 1.0797383, (0.3245 - p.z) * 0.95105654 + (gsl_tmp_17 - 1.2065) * 0.309017), 0.0 - p.z), p.z - 4.37), min(min(abs(p.y) - 0.495, abs(p.x) - 0.495), max((gsl_tmp_18 - gsl_tmp_20) * 0.99939084 + (gsl_tmp_19 - gsl_tmp_20) * 0.0348995, (gsl_tmp_19 - gsl_tmp_20) * 0.99939084 + (gsl_tmp_18 - gsl_tmp_20) * 0.0348995) / 1.002708));
  return min(pz2, driver_powerShank(p, 4.37, 0.9));
}

fn gsl_remap_16(p: vec3<f32>) -> f32 {
  return gsl_remap_15(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_17(p: vec3<f32>) -> f32 {
  let t25 = max(drill_torxAt(p, 5.0, 0.9 * (1.0 - clamp(sign(23.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(22.9999), 0.0, 1.0)) + 1.2 * (1.0 - clamp(sign(21.9999), 0.0, 1.0)) + 1.35 * (1.0 - clamp(sign(20.9999), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(18.9999), 0.0, 1.0)) + 2.1 * (1.0 - clamp(sign(17.9999), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(16.9999), 0.0, 1.0)) + 2.6 * (1.0 - clamp(sign(15.9999), 0.0, 1.0)) + 2.8 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 3.35 * (1.0 - clamp(sign(9.9999), 0.0, 1.0)) + 3.95 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 4.5 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 5.1 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 5.6 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 6.75 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 7.93 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 8.95 * (1.0 - clamp(sign(24.9999), 0.0, 1.0)) + 11.35 * (1.0 - clamp(sign(29.9999), 0.0, 1.0)) + 13.45 * (1.0 - clamp(sign(34.9999), 0.0, 1.0)) + 15.7 * (1.0 - clamp(sign(44.9999), 0.0, 1.0)) + 17.75 * (1.0 - clamp(sign(54.9999), 0.0, 1.0)) + 20.2 * (1.0 - clamp(sign(64.9999), 0.0, 1.0)) + 22.4 * (1.0 - clamp(sign(74.9999), 0.0, 1.0)), 0.6 * (1.0 - clamp(sign(23.9999), 0.0, 1.0)) + 0.7 * (1.0 - clamp(sign(22.9999), 0.0, 1.0)) + 0.85 * (1.0 - clamp(sign(21.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(20.9999), 0.0, 1.0)) + 1.1 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(18.9999), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(17.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(16.9999), 0.0, 1.0)) + 1.9 * (1.0 - clamp(sign(15.9999), 0.0, 1.0)) + 2.05 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(9.9999), 0.0, 1.0)) + 2.85 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 3.25 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 3.68 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 4.05 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 4.85 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 5.64 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 6.45 * (1.0 - clamp(sign(24.9999), 0.0, 1.0)) + 8.05 * (1.0 - clamp(sign(29.9999), 0.0, 1.0)) + 9.6 * (1.0 - clamp(sign(34.9999), 0.0, 1.0)) + 11.2 * (1.0 - clamp(sign(44.9999), 0.0, 1.0)) + 12.8 * (1.0 - clamp(sign(54.9999), 0.0, 1.0)) + 14.4 * (1.0 - clamp(sign(64.9999), 0.0, 1.0)) + 16.0 * (1.0 - clamp(sign(74.9999), 0.0, 1.0)), 5.0), p.z - 5.0);
  return min(t25, driver_powerShank(p, 5.0, 1.6));
}

fn gsl_remap_18(p: vec3<f32>) -> f32 {
  return gsl_remap_17(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_19(p: vec3<f32>) -> f32 {
  let r2 = max(max(max(max(max(abs(p.x) - 1.42, abs(p.y) - 1.42), 0.0 - p.z), (abs(p.x) - 1.42 + 0.0 - p.z + 1.42) * 0.70710677), (abs(p.y) - 1.42 + 0.0 - p.z + 1.42) * 0.70710677), p.z - 5.0);
  return min(r2, driver_powerShank(p, 5.0, 1.4));
}

fn gsl_remap_20(p: vec3<f32>) -> f32 {
  return gsl_remap_19(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_21(p: vec3<f32>) -> f32 {
  let hex4 = driver_hexTip(p, 4.0, 5.0);
  return min(hex4, driver_powerShank(p, 5.0, 1.8));
}

fn gsl_remap_22(p: vec3<f32>) -> f32 {
  return gsl_remap_21(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_23(p: vec3<f32>) -> f32 {
  let slot55 = driver_slotTip(p, 1.0, 5.5, 5.0);
  return min(slot55, driver_powerShank(p, 5.0, 0.5));
}

fn gsl_remap_24(p: vec3<f32>) -> f32 {
  return gsl_remap_23(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_25(p: vec3<f32>) -> f32 {
  return min(drill_phillipsGauge(p, 0.0, 2.78 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.78 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 4.37 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 6.74 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 8.34 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)), 0.64 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 1.001 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.539 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.497 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 3.574 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)), 0.315 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.513 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.102 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.098 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 2.738 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)), 0.813 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.286 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 3.81 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 5.08 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)), 0.25 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)), radians(0.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 138.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 140.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 146.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 153.0 * (1.0 - clamp(sign(3.9999), 0.0, 1.0))), radians(7.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(3.9999), 0.0, 1.0))), 0.225 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(3.9999), 0.0, 1.0))), driver_microShank(p, 2.78, 0.3));
}

fn gsl_remap_26(p: vec3<f32>) -> f32 {
  return gsl_remap_25(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_27(p: vec3<f32>) -> f32 {
  return min(drill_phillipsGauge(p, 0.0, 2.78 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.78 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 4.37 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 6.74 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 8.34 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.64 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.001 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 1.539 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.497 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 3.574 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.315 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.513 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 1.102 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.098 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.738 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.813 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.286 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 3.81 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 5.08 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.25 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), radians(0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 138.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 140.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 146.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 153.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0))), radians(7.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0))), 0.225 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0))), driver_microShank(p, 2.78, 0.45));
}

fn gsl_remap_28(p: vec3<f32>) -> f32 {
  return gsl_remap_27(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_29(p: vec3<f32>) -> f32 {
  return min(max(drill_torxAt(p, 3.0, 0.9 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)) + 1.2 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 1.35 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.1 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.6 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 2.8 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)) + 3.35 * (1.0 - clamp(sign(8.9999), 0.0, 1.0)) + 3.95 * (1.0 - clamp(sign(13.9999), 0.0, 1.0)) + 4.5 * (1.0 - clamp(sign(18.9999), 0.0, 1.0)) + 5.1 * (1.0 - clamp(sign(20.9999), 0.0, 1.0)) + 5.6 * (1.0 - clamp(sign(23.9999), 0.0, 1.0)) + 6.75 * (1.0 - clamp(sign(33.9999), 0.0, 1.0)) + 7.93 * (1.0 - clamp(sign(38.9999), 0.0, 1.0)) + 8.95 * (1.0 - clamp(sign(43.9999), 0.0, 1.0)) + 11.35 * (1.0 - clamp(sign(48.9999), 0.0, 1.0)) + 13.45 * (1.0 - clamp(sign(53.9999), 0.0, 1.0)) + 15.7 * (1.0 - clamp(sign(63.9999), 0.0, 1.0)) + 17.75 * (1.0 - clamp(sign(73.9999), 0.0, 1.0)) + 20.2 * (1.0 - clamp(sign(83.9999), 0.0, 1.0)) + 22.4 * (1.0 - clamp(sign(93.9999), 0.0, 1.0)), 0.6 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 0.7 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)) + 0.85 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.1 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.9 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 2.05 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(8.9999), 0.0, 1.0)) + 2.85 * (1.0 - clamp(sign(13.9999), 0.0, 1.0)) + 3.25 * (1.0 - clamp(sign(18.9999), 0.0, 1.0)) + 3.68 * (1.0 - clamp(sign(20.9999), 0.0, 1.0)) + 4.05 * (1.0 - clamp(sign(23.9999), 0.0, 1.0)) + 4.85 * (1.0 - clamp(sign(33.9999), 0.0, 1.0)) + 5.64 * (1.0 - clamp(sign(38.9999), 0.0, 1.0)) + 6.45 * (1.0 - clamp(sign(43.9999), 0.0, 1.0)) + 8.05 * (1.0 - clamp(sign(48.9999), 0.0, 1.0)) + 9.6 * (1.0 - clamp(sign(53.9999), 0.0, 1.0)) + 11.2 * (1.0 - clamp(sign(63.9999), 0.0, 1.0)) + 12.8 * (1.0 - clamp(sign(73.9999), 0.0, 1.0)) + 14.4 * (1.0 - clamp(sign(83.9999), 0.0, 1.0)) + 16.0 * (1.0 - clamp(sign(93.9999), 0.0, 1.0)), 3.0), p.z - 3.0), driver_microShank(p, 3.0, 0.6));
}

fn gsl_remap_30(p: vec3<f32>) -> f32 {
  return gsl_remap_29(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_31(p: vec3<f32>) -> f32 {
  return min(max(drill_torxAt(p, 3.0, 0.9 * (1.0 - clamp(sign(8.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(7.9999), 0.0, 1.0)) + 1.2 * (1.0 - clamp(sign(6.9999), 0.0, 1.0)) + 1.35 * (1.0 - clamp(sign(5.9999), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)) + 2.1 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.6 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.8 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 3.35 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 3.95 * (1.0 - clamp(sign(9.9999), 0.0, 1.0)) + 4.5 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 5.1 * (1.0 - clamp(sign(16.9999), 0.0, 1.0)) + 5.6 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 6.75 * (1.0 - clamp(sign(29.9999), 0.0, 1.0)) + 7.93 * (1.0 - clamp(sign(34.9999), 0.0, 1.0)) + 8.95 * (1.0 - clamp(sign(39.9999), 0.0, 1.0)) + 11.35 * (1.0 - clamp(sign(44.9999), 0.0, 1.0)) + 13.45 * (1.0 - clamp(sign(49.9999), 0.0, 1.0)) + 15.7 * (1.0 - clamp(sign(59.9999), 0.0, 1.0)) + 17.75 * (1.0 - clamp(sign(69.9999), 0.0, 1.0)) + 20.2 * (1.0 - clamp(sign(79.9999), 0.0, 1.0)) + 22.4 * (1.0 - clamp(sign(89.9999), 0.0, 1.0)), 0.6 * (1.0 - clamp(sign(8.9999), 0.0, 1.0)) + 0.7 * (1.0 - clamp(sign(7.9999), 0.0, 1.0)) + 0.85 * (1.0 - clamp(sign(6.9999), 0.0, 1.0)) + 1.0 * (1.0 - clamp(sign(5.9999), 0.0, 1.0)) + 1.1 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(3.9999), 0.0, 1.0)) + 1.5 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)) + 1.75 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.9 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.05 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.4 * (1.0 - clamp(sign(4.9999), 0.0, 1.0)) + 2.85 * (1.0 - clamp(sign(9.9999), 0.0, 1.0)) + 3.25 * (1.0 - clamp(sign(14.9999), 0.0, 1.0)) + 3.68 * (1.0 - clamp(sign(16.9999), 0.0, 1.0)) + 4.05 * (1.0 - clamp(sign(19.9999), 0.0, 1.0)) + 4.85 * (1.0 - clamp(sign(29.9999), 0.0, 1.0)) + 5.64 * (1.0 - clamp(sign(34.9999), 0.0, 1.0)) + 6.45 * (1.0 - clamp(sign(39.9999), 0.0, 1.0)) + 8.05 * (1.0 - clamp(sign(44.9999), 0.0, 1.0)) + 9.6 * (1.0 - clamp(sign(49.9999), 0.0, 1.0)) + 11.2 * (1.0 - clamp(sign(59.9999), 0.0, 1.0)) + 12.8 * (1.0 - clamp(sign(69.9999), 0.0, 1.0)) + 14.4 * (1.0 - clamp(sign(79.9999), 0.0, 1.0)) + 16.0 * (1.0 - clamp(sign(89.9999), 0.0, 1.0)), 3.0), p.z - 3.0), driver_microShank(p, 3.0, 0.9));
}

fn gsl_remap_32(p: vec3<f32>) -> f32 {
  return gsl_remap_31(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_33(p: vec3<f32>) -> f32 {
  return min(driver_hexTip(p, 1.5, 3.0), driver_microShank(p, 3.0, 0.6));
}

fn gsl_remap_34(p: vec3<f32>) -> f32 {
  return gsl_remap_33(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_35(p: vec3<f32>) -> f32 {
  return min(driver_slotTip(p, 0.4, 2.5, 3.0), driver_microShank(p, 3.0, 0.2));
}

fn gsl_remap_36(p: vec3<f32>) -> f32 {
  return gsl_remap_35(vec3<f32>(p.x, p.y * -1.0 + p.z * 0.00000000000000012246469, -p.y * 0.00000000000000012246469 + p.z * -1.0));
}

fn gsl_remap_37(p: vec3<f32>) -> f32 {
  let gsl_tmp_63 = length(p.xy);
  let gsl_tmp_64 = abs(p.x);
  let gsl_tmp_65 = abs(p.y);
  let gsl_tmp_66 = abs(gsl_tmp_64 - gsl_tmp_65) * 0.70710677;
  let gsl_tmp_67 = ((gsl_tmp_64 + gsl_tmp_65) * 0.70710677 - 0.735) * 0.99496853 - (p.z - 0.38236797) * 0.10018806;
  return max(min(max(max(length(p.xy) - 4.0, -p.z), length(vec3<f32>(p.x, p.y, p.z - -3.4)) - 6.5), max(max(max(length(p.xy) - 2.0, -10.0 - p.z), p.z - 1.55), (length(p.xy) - 2.0 + -p.z - 10.0 + 0.6) * 0.7071)), -max(max(gsl_tmp_63 * 0.89493436 + (2.7163715 - p.z) * 0.4461978 - 1.9688556, (-0.008438476 - p.z) * 0.88294756 + gsl_tmp_63 * 0.46947157), -max(gsl_tmp_66 * 0.34202015 - gsl_tmp_67 * 0.9396926, (gsl_tmp_66 - 0.4075) * 0.6946584 - (gsl_tmp_67 - 0.14831787) * 0.7193398)));
}

fn gsl_remap_38(p: vec3<f32>) -> f32 {
  return min(max(max(length(p.xy) - 4.0, -p.z), length(vec3<f32>(p.x, p.y, p.z - -3.4)) - 6.5), max(max(length(p.xy) - 2.0, -10.0 - p.z), p.z - 1.55));
}

fn gsl_remap_39(p: vec3<f32>) -> f32 {
  let gsl_tmp_118 = length(p.xy);
  let gsl_tmp_119 = abs(p.x);
  let gsl_tmp_121 = (0.915 + (p.z - 0.7685499) * 0.10069471) * 0.70710677;
  let gsl_tmp_120 = abs(p.y);
  return max(min(max(max(length(p.xy) - 4.0, -p.z), length(vec3<f32>(p.x, p.y, p.z - -3.4)) - 6.5), max(max(max(length(p.xy) - 2.0, -10.0 - p.z), p.z - 1.55), (length(p.xy) - 2.0 + -p.z - 10.0 + 0.6) * 0.7071)), -min(max(max(gsl_tmp_118 * 0.89493436 + (2.7341259 - p.z) * 0.4461978 - 1.9241089, (0.1464499 - p.z) * 0.88294756 + gsl_tmp_118 * 0.46947157), min(min(abs(p.y) - 0.515, abs(p.x) - 0.515), max((gsl_tmp_119 - gsl_tmp_121) * 0.99939084 + (gsl_tmp_120 - gsl_tmp_121) * 0.0348995, (gsl_tmp_120 - gsl_tmp_121) * 0.99939084 + (gsl_tmp_119 - gsl_tmp_121) * 0.0348995) / 1.002708)), (abs(abs(p.x) - abs(p.y)) * 0.70710677 * 0.8660254 - (p.z - (2.5841258 + ((abs(p.x) + abs(p.y)) * 0.70710677 - 1.1129231) * 0.1446373)) * 0.5) / 1.0026116));
}

fn gsl_remap_40(p: vec3<f32>) -> f32 {
  return length(vec2<f32>(p.x - 2.125, p.y)) - 0.5;
}

fn gsl_remap_41(p: vec3<f32>) -> f32 {
  return max(min(max(max(length(p.xy) - 4.75, -p.z), length(vec3<f32>(p.x, p.y, p.z - -4.3)) - 8.0), max(max(max(length(p.xy) - 2.5, -10.0 - p.z), p.z - 1.85), (length(p.xy) - 2.5 + -p.z - 10.0 + 0.75) * 0.7071)), -drill_prismVoid(p, 3.3770761, max(length(p.xy) - 2.25, -gsl_remap_40(map_patternPolarXY(p, 6.0))), 1.91));
}

fn gsl_remap_42(p: vec3<f32>) -> f32 {
  return min(max(max(length(p.xy) - 4.75, -p.z), length(vec3<f32>(p.x, p.y, p.z - -4.3)) - 8.0), max(max(length(p.xy) - 2.5, -10.0 - p.z), p.z - 1.85));
}

fn gsl_remap_43(p: vec3<f32>) -> f32 {
  return min(drill_phillipsGauge(p, 0.0, 2.78 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.78 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 4.37 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 6.74 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 8.34 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.64 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.001 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 1.539 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.497 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 3.574 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.315 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.513 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 1.102 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.098 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.738 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.813 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.286 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 3.81 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 5.08 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), 0.25 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(2.9999), 0.0, 1.0)), radians(0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 138.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 140.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 146.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 153.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0))), radians(7.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0))), 0.225 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(2.9999), 0.0, 1.0))), driver_screwdriverBlade(p, 2.78, 0.45, 4.5, 75.0));
}

fn gsl_remap_44(p: vec3<f32>) -> f32 {
  return gsl_remap_43(vec3<f32>(0.0, 0.0, 0.0) + (p - vec3<f32>(0.0, 0.0, 0.0)) * -0.5 - cross(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 0.8660254 + vec3<f32>(0.57735026, 0.57735026, 0.57735026) * dot(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 1.5);
}

fn gsl_remap_45(p: vec3<f32>) -> f32 {
  let ph2 = drill_phillipsGauge(p, 0.0, 2.78 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 2.78 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 4.37 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 6.74 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 8.34 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.64 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.001 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.539 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.497 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 3.574 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.315 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.513 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 1.102 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 2.098 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.738 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.813 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 1.27 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 2.286 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 3.81 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.08 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), 0.25 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.38 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)), radians(0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 138.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 140.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 146.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 153.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0))), radians(7.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 5.75 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 7.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0))), 0.225 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(-0.0001), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(0.9999), 0.0, 1.0)) + 0.0 * (1.0 - clamp(sign(1.9999), 0.0, 1.0)));
  return min(ph2, driver_screwdriverBlade(p, 4.37, 0.7, 6.0, 100.0));
}

fn gsl_remap_46(p: vec3<f32>) -> f32 {
  return gsl_remap_45(vec3<f32>(0.0, 0.0, 0.0) + (p - vec3<f32>(0.0, 0.0, 0.0)) * -0.5 - cross(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 0.8660254 + vec3<f32>(0.57735026, 0.57735026, 0.57735026) * dot(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 1.5);
}

fn gsl_remap_47(p: vec3<f32>) -> f32 {
  let gsl_tmp_17 = length(p.xy);
  let gsl_tmp_18 = abs(p.x);
  let gsl_tmp_20 = (0.851 + (p.z - 0.3245) * 0.10069471) * 0.70710677;
  let gsl_tmp_19 = abs(p.y);
  let pz2 = max(max(max(max(gsl_tmp_17 * 0.89493436 - (p.z - 0.3245) * 0.4461978 - 1.0797383, (0.3245 - p.z) * 0.95105654 + (gsl_tmp_17 - 1.2065) * 0.309017), 0.0 - p.z), p.z - 4.37), min(min(abs(p.y) - 0.495, abs(p.x) - 0.495), max((gsl_tmp_18 - gsl_tmp_20) * 0.99939084 + (gsl_tmp_19 - gsl_tmp_20) * 0.0348995, (gsl_tmp_19 - gsl_tmp_20) * 0.99939084 + (gsl_tmp_18 - gsl_tmp_20) * 0.0348995) / 1.002708));
  return min(pz2, driver_screwdriverBlade(p, 4.37, 0.9, 6.0, 100.0));
}

fn gsl_remap_48(p: vec3<f32>) -> f32 {
  return gsl_remap_47(vec3<f32>(0.0, 0.0, 0.0) + (p - vec3<f32>(0.0, 0.0, 0.0)) * -0.5 - cross(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 0.8660254 + vec3<f32>(0.57735026, 0.57735026, 0.57735026) * dot(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 1.5);
}

fn gsl_remap_49(p: vec3<f32>) -> f32 {
  return min(driver_slotTip(p, 1.2, 6.5, 5.0), driver_screwdriverBlade(p, 5.0, 0.6, 6.0, 125.0));
}

fn gsl_remap_50(p: vec3<f32>) -> f32 {
  return gsl_remap_49(vec3<f32>(0.0, 0.0, 0.0) + (p - vec3<f32>(0.0, 0.0, 0.0)) * -0.5 - cross(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 0.8660254 + vec3<f32>(0.57735026, 0.57735026, 0.57735026) * dot(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 1.5);
}

fn gsl_remap_51(p: vec3<f32>) -> f32 {
  return driver_screwdriverHandle(vec3<f32>(0.0, 0.0, 0.0) + (p - vec3<f32>(0.0, 0.0, 0.0)) * -0.5 - cross(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 0.8660254 + vec3<f32>(0.57735026, 0.57735026, 0.57735026) * dot(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 1.5, 77.78, 30.0, 111.0);
}

fn gsl_remap_52(p: vec3<f32>) -> f32 {
  return driver_screwdriverHandle(vec3<f32>(0.0, 0.0, 0.0) + (p - vec3<f32>(0.0, 0.0, 0.0)) * -0.5 - cross(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 0.8660254 + vec3<f32>(0.57735026, 0.57735026, 0.57735026) * dot(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 1.5, 104.37, 36.0, 118.0);
}

fn gsl_remap_53(p: vec3<f32>) -> f32 {
  return driver_screwdriverHandle(vec3<f32>(0.0, 0.0, 0.0) + (p - vec3<f32>(0.0, 0.0, 0.0)) * -0.5 - cross(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 0.8660254 + vec3<f32>(0.57735026, 0.57735026, 0.57735026) * dot(vec3<f32>(0.57735026, 0.57735026, 0.57735026), p - vec3<f32>(0.0, 0.0, 0.0)) * 1.5, 130.0, 36.0, 118.0);
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

fn drill_phillipsRibs(p: vec3<f32>, core: f32, facet: f32, ridge: f32, lean: f32, root: f32, zb: f32) -> f32 {
  let x = abs(p.x);
  let y = abs(p.y);
  let w = abs(x - y) * 0.70710677;
  let half = ridge / 2.0;
  let u = (x + y) * 0.70710677;
  let s = (u - core / 2.0) * cos(lean) - (p.z - zb) * sin(lean);
  let sFacet = facet / 2.0 / tan(max(half, 0.1));
  let faceted = max(w * cos(half) - s * sin(half), (w - facet / 2.0) * 0.6946584 - (s - sFacet) * 0.7193398);
  let qs = s - root;
  let t = max(qs * 0.6946584 + w * 0.7193398, 0.0);
  let rounded = sign(w * 0.6946584 - qs * 0.7193398) * length(vec2<f32>(qs - t * 0.6946584, w - t * 0.7193398)) - root;
  return select(faceted, rounded, root > 0.0);
}

fn drill_phillipsGauge(p: vec3<f32>, zTip: f32, zTop: f32, core: f32, facet: f32, g: f32, flat: f32, ridge: f32, lean: f32, root: f32) -> f32 {
  let r = length(p.xy);
  let zg = zTip + (g - flat) / 2.0 * 0.3249197;
  let zb = zTip + (core - flat) / 2.0 * 0.3249197;
  let form = max(max(max(max(r * 0.89493436 - (p.z - zg) * 0.4461978 - g / 2.0 * 0.89493436, zTip - p.z), (zTip - p.z) * 0.95105654 + (r - flat / 2.0) * 0.309017), p.z - zTop), -drill_phillipsRibs(p, core, facet, ridge, lean, root, zb));
  return form;
}

fn gsl_drill_torxXY_remap(p: vec3<f32>, filletR: f32, major: f32, minor: f32) -> f32 {
  let bite = length(vec2<f32>(p.x - (minor / 2.0 + filletR), p.y)) - filletR;
  return bite;
}

fn drill_torxXY_2(p: vec3<f32>, major: f32, minor: f32, filletR: f32) -> f32 {
  let bites = gsl_drill_torxXY_remap(map_patternPolarXY(p, 6.0), filletR, major, minor);
  return max(length(p.xy) - major / 2.0, -bites);
}

fn drill_torxXY(p: vec3<f32>, major: f32, minor: f32) -> f32 {
  return drill_torxXY_2(p, major, minor, 0.4 * (major - minor));
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn drill_torxAt(p: vec3<f32>, zTop: f32, major: f32, minor: f32, depth: f32) -> f32 {
  return drill_prismVoid(p, zTop, drill_torxXY(p, major, minor), depth);
}

fn driver_hexTip(p: vec3<f32>, acrossFlats: f32, gsl_var_length: f32) -> f32 {
  let form = max(drill_hexSocketAt(p, gsl_var_length, acrossFlats, gsl_var_length), p.z - gsl_var_length);
  return form;
}

fn driver_slotTip(p: vec3<f32>, thickness: f32, width: f32, gsl_var_length: f32) -> f32 {
  let edge = abs(p.y) * 0.9961947 - p.z * 0.08715574 - thickness / 2.0 * 0.9961947;
  let ramp = abs(p.y) * 0.9659258 - (p.z - 1.5) * 0.25881904 - thickness / 2.0 * 0.9659258;
  let form = max(max(max(min(edge, ramp), abs(p.x) - width / 2.0), -p.z), p.z - gsl_var_length);
  return form;
}

fn driver_neck(p: vec3<f32>, zLow: f32, neckR: f32) -> f32 {
  return (length(p.xy) - neckR - (p.z - zLow)) * 0.70710677;
}

fn driver_driveEnd(p: vec3<f32>, acrossFlats: f32, zTop: f32) -> f32 {
  return (length(p.xy) - 0.52 * acrossFlats + p.z - zTop) * 0.70710677;
}

fn driver_insertShank(p: vec3<f32>, zLow: f32, neckR: f32) -> f32 {
  let groove = (3.35 + abs(p.z - 16.8) - length(p.xy)) * 0.70710677;
  let form = max(max(max(max(max(drill_hexXY(p, 6.32), zLow - p.z), p.z - 25.0), driver_neck(p, zLow, neckR)), driver_driveEnd(p, 6.32, 25.0)), -groove);
  return form;
}

fn driver_powerShank(p: vec3<f32>, zLow: f32, neckR: f32) -> f32 {
  let torsion = max(max(max(length(p.xy) - 2.2, zLow - p.z), p.z - 25.5), driver_neck(p, zLow, neckR));
  let hexPart = max(max(max(drill_hexXY(p, 6.32), p.z - 50.0), driver_driveEnd(p, 6.32, 50.0)), (length(p.xy) - 2.2 - (p.z - 25.0)) * 0.70710677);
  let ball = length(vec2<f32>(length(p.xy) - 4.75, p.z - 40.5)) - 2.4;
  let form = max(min(torsion, hexPart), -ball);
  return form;
}

fn driver_microShank(p: vec3<f32>, zLow: f32, neckR: f32) -> f32 {
  let form = max(max(max(max(drill_hexXY(p, 3.94), zLow - p.z), p.z - 28.0), driver_neck(p, zLow, neckR)), driver_driveEnd(p, 3.94, 28.0));
  return form;
}

fn driver_screwdriverBlade(p: vec3<f32>, zLow: f32, neckR: f32, bladeDia: f32, bladeLen: f32) -> f32 {
  let form = max(max(max(length(p.xy) - bladeDia / 2.0, zLow - p.z), p.z - (zLow + bladeLen + 25.0)), driver_neck(p, zLow, neckR));
  return form;
}

fn gsl_driver_screwdriverHandle_remap(p: vec3<f32>, dia: f32, len: f32, zBase: f32) -> f32 {
  let r = dia / 2.0;
  let top = zBase + len;
  let flute = max(max(length(vec2<f32>(p.x - 1.1 * r, p.y)) - 0.28 * r, zBase + 0.3 * len - p.z), p.z - (top - 0.25 * len));
  return flute;
}

fn driver_screwdriverHandle(p: vec3<f32>, zBase: f32, dia: f32, len: f32) -> f32 {
  let r = dia / 2.0;
  let top = zBase + len;
  let body = max(max(max(length(p.xy) - r, zBase - p.z), (length(p.xy) - 0.55 * r - (p.z - zBase) * 0.8) * 0.7808688), min(p.z - (top - r), length(vec3<f32>(p.x, p.y, p.z - (top - r))) - r));
  let form = max(body, -gsl_driver_screwdriverHandle_remap(map_patternPolarXY(p, 6.0), dia, len, zBase));
  return form;
}

fn gsl_shapeParts(p: vec3<f32>, time: f32) -> array<f32, 32> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-43.177418, -73.65885, -0.01);
  let gsl_partHi = vec3<f32>(-36.822582, -66.34115, 25.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_remap_2(p - vec3<f32>(-40.0, -70.0, 25.0)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let insertPH2 = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-27.233511, -73.65885, -0.01);
  let gsl_partHi_2 = vec3<f32>(-20.766489, -66.34115, 25.01);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_remap_4(p - vec3<f32>(-24.0, -70.0, 25.0)) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let insertPZ2 = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-11.17, -73.65885, -0.01);
  let gsl_partHi_3 = vec3<f32>(-4.83, -66.34115, 25.01);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = gsl_remap_6(p - vec3<f32>(-8.0, -70.0, 25.0)) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let insertT25 = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(4.83, -73.65885, -0.01);
  let gsl_partHi_4 = vec3<f32>(11.17, -66.34115, 25.01);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = gsl_remap_8(p - vec3<f32>(8.0, -70.0, 25.0)) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let insertR2 = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(20.83, -73.65885, -0.01);
  let gsl_partHi_5 = vec3<f32>(27.17, -66.34115, 25.01);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = gsl_remap_10(p - vec3<f32>(24.0, -70.0, 25.0)) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let insertHex4 = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(36.83, -73.65885, -0.01);
  let gsl_partHi_6 = vec3<f32>(43.17, -66.34115, 25.01);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = gsl_remap_12(p - vec3<f32>(40.0, -70.0, 25.0)) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let insertSlot = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(-43.177418, -48.658855, -0.01);
  let gsl_partHi_7 = vec3<f32>(-36.822582, -41.341145, 50.01);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = gsl_remap_14(p - vec3<f32>(-40.0, -45.0, 50.0)) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let powerPH2 = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-27.233511, -48.658855, -0.01);
  let gsl_partHi_8 = vec3<f32>(-20.766489, -41.341145, 50.01);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = gsl_remap_16(p - vec3<f32>(-24.0, -45.0, 50.0)) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let powerPZ2 = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(-11.17, -48.658855, -0.01);
  let gsl_partHi_9 = vec3<f32>(-4.83, -41.341145, 50.01);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = gsl_remap_18(p - vec3<f32>(-8.0, -45.0, 50.0)) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let powerT25 = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(4.83, -48.658855, -0.01);
  let gsl_partHi_10 = vec3<f32>(11.17, -41.341145, 50.01);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = gsl_remap_20(p - vec3<f32>(8.0, -45.0, 50.0)) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let powerR2 = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(20.83, -48.658855, -0.01);
  let gsl_partHi_11 = vec3<f32>(27.17, -41.341145, 50.01);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = gsl_remap_22(p - vec3<f32>(24.0, -45.0, 50.0)) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let powerHex4 = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(36.83, -48.658855, -0.01);
  let gsl_partHi_12 = vec3<f32>(43.17, -41.341145, 50.01);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = gsl_remap_24(p - vec3<f32>(40.0, -45.0, 50.0)) + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let powerSlot = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(-41.98, -22.28476, -0.01);
  let gsl_partHi_13 = vec3<f32>(-38.02, -17.71524, 28.01);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = gsl_remap_26(p - vec3<f32>(-40.0, -20.0, 28.0)) + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let microPH0 = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(-25.98, -22.28476, -0.01);
  let gsl_partHi_14 = vec3<f32>(-22.02, -17.71524, 28.01);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = gsl_remap_28(p - vec3<f32>(-24.0, -20.0, 28.0)) + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let microPH1 = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(-9.98, -22.28476, -0.01);
  let gsl_partHi_15 = vec3<f32>(-6.02, -17.71524, 28.01);
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = gsl_remap_30(p - vec3<f32>(-8.0, -20.0, 28.0)) + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let microT6 = gsl_part_15;
  let gsl_partLo_16 = vec3<f32>(6.02, -22.28476, -0.01);
  let gsl_partHi_16 = vec3<f32>(9.98, -17.71524, 28.01);
  let gsl_partBox_16 = gsl_boxDistance(p, gsl_partLo_16, gsl_partHi_16);
  var gsl_part_16 = gsl_partBox_16;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_16 <= gsl_vanishView.w) {
    let gsl_partVanish_16 = gsl_vanish(gsl_partLo_16, gsl_partHi_16);
    gsl_part_16 = 1e10;
    if (gsl_partVanish_16.x > 0.0) { gsl_part_16 = gsl_remap_32(p - vec3<f32>(8.0, -20.0, 28.0)) + gsl_partVanish_16.y; }
    if (abs(gsl_part_16) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_16), gsl_partVanish_16.zw); }
  }
  let microT10 = gsl_part_16;
  let gsl_partLo_17 = vec3<f32>(22.02, -22.28476, -0.01);
  let gsl_partHi_17 = vec3<f32>(25.98, -17.71524, 28.01);
  let gsl_partBox_17 = gsl_boxDistance(p, gsl_partLo_17, gsl_partHi_17);
  var gsl_part_17 = gsl_partBox_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_17 <= gsl_vanishView.w) {
    let gsl_partVanish_17 = gsl_vanish(gsl_partLo_17, gsl_partHi_17);
    gsl_part_17 = 1e10;
    if (gsl_partVanish_17.x > 0.0) { gsl_part_17 = gsl_remap_34(p - vec3<f32>(24.0, -20.0, 28.0)) + gsl_partVanish_17.y; }
    if (abs(gsl_part_17) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_17), gsl_partVanish_17.zw); }
  }
  let microHex15 = gsl_part_17;
  let gsl_partLo_18 = vec3<f32>(38.02, -22.28476, -0.01);
  let gsl_partHi_18 = vec3<f32>(41.98, -17.71524, 28.01);
  let gsl_partBox_18 = gsl_boxDistance(p, gsl_partLo_18, gsl_partHi_18);
  var gsl_part_18 = gsl_partBox_18;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_18 <= gsl_vanishView.w) {
    let gsl_partVanish_18 = gsl_vanish(gsl_partLo_18, gsl_partHi_18);
    gsl_part_18 = 1e10;
    if (gsl_partVanish_18.x > 0.0) { gsl_part_18 = gsl_remap_36(p - vec3<f32>(40.0, -20.0, 28.0)) + gsl_partVanish_18.y; }
    if (abs(gsl_part_18) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_18), gsl_partVanish_18.zw); }
  }
  let microSlot = gsl_part_18;
  let gsl_partLo_19 = vec3<f32>(-34.01, 0.99, -0.01);
  let gsl_partHi_19 = vec3<f32>(-25.99, 9.01, 13.11);
  let gsl_partBox_19 = gsl_boxDistance(p, gsl_partLo_19, gsl_partHi_19);
  var gsl_part_19 = gsl_partBox_19;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_19 <= gsl_vanishView.w) {
    let gsl_partVanish_19 = gsl_vanish(gsl_partLo_19, gsl_partHi_19);
    gsl_part_19 = 1e10;
    let gsl_simpleWeight = gsl_simple(vec3<f32>(-34.01, 0.99, -0.01), vec3<f32>(-25.99, 9.01, 13.11));
    var gsl_simplePart = 1e10;
    if (gsl_simpleWeight > 0.0) {
      gsl_simplePart = gsl_remap_37(p - vec3<f32>(-30.0, 5.0, 10.0));
    }
    if (gsl_simpleWeight < 1.0) {
      gsl_simplePart = mix(max(gsl_remap_38(p - vec3<f32>(-30.0, 5.0, 10.0)), gsl_boxSdf(p, vec3<f32>(-34.01, 0.99, -0.01), vec3<f32>(-25.99, 9.01, 13.11))), gsl_simplePart, gsl_simpleWeight);
    }
    if (gsl_partVanish_19.x > 0.0) { gsl_part_19 = gsl_simplePart + gsl_partVanish_19.y; }
    if (abs(gsl_part_19) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_19), gsl_partVanish_19.zw); }
  }
  let screwPH = gsl_part_19;
  let gsl_partLo_20 = vec3<f32>(-4.01, 0.99, -0.01);
  let gsl_partHi_20 = vec3<f32>(4.01, 9.01, 13.11);
  let gsl_partBox_20 = gsl_boxDistance(p, gsl_partLo_20, gsl_partHi_20);
  var gsl_part_20 = gsl_partBox_20;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_20 <= gsl_vanishView.w) {
    let gsl_partVanish_20 = gsl_vanish(gsl_partLo_20, gsl_partHi_20);
    gsl_part_20 = 1e10;
    let gsl_simpleWeight_2 = gsl_simple(vec3<f32>(-4.01, 0.99, -0.01), vec3<f32>(4.01, 9.01, 13.11));
    var gsl_simplePart_2 = 1e10;
    if (gsl_simpleWeight_2 > 0.0) {
      gsl_simplePart_2 = gsl_remap_39(p - vec3<f32>(0.0, 5.0, 10.0));
    }
    if (gsl_simpleWeight_2 < 1.0) {
      gsl_simplePart_2 = mix(max(gsl_remap_38(p - vec3<f32>(0.0, 5.0, 10.0)), gsl_boxSdf(p, vec3<f32>(-4.01, 0.99, -0.01), vec3<f32>(4.01, 9.01, 13.11))), gsl_simplePart_2, gsl_simpleWeight_2);
    }
    if (gsl_partVanish_20.x > 0.0) { gsl_part_20 = gsl_simplePart_2 + gsl_partVanish_20.y; }
    if (abs(gsl_part_20) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_20), gsl_partVanish_20.zw); }
  }
  let screwPZ = gsl_part_20;
  let gsl_partLo_21 = vec3<f32>(25.24, 0.24, -0.01);
  let gsl_partHi_21 = vec3<f32>(34.76, 9.76, 13.71);
  let gsl_partBox_21 = gsl_boxDistance(p, gsl_partLo_21, gsl_partHi_21);
  var gsl_part_21 = gsl_partBox_21;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_21 <= gsl_vanishView.w) {
    let gsl_partVanish_21 = gsl_vanish(gsl_partLo_21, gsl_partHi_21);
    gsl_part_21 = 1e10;
    let gsl_simpleWeight_3 = gsl_simple(vec3<f32>(25.24, 0.24, -0.01), vec3<f32>(34.76, 9.76, 13.71));
    var gsl_simplePart_3 = 1e10;
    if (gsl_simpleWeight_3 > 0.0) {
      gsl_simplePart_3 = gsl_remap_41(p - vec3<f32>(30.0, 5.0, 10.0));
    }
    if (gsl_simpleWeight_3 < 1.0) {
      gsl_simplePart_3 = mix(max(gsl_remap_42(p - vec3<f32>(30.0, 5.0, 10.0)), gsl_boxSdf(p, vec3<f32>(25.24, 0.24, -0.01), vec3<f32>(34.76, 9.76, 13.71))), gsl_simplePart_3, gsl_simpleWeight_3);
    }
    if (gsl_partVanish_21.x > 0.0) { gsl_part_21 = gsl_simplePart_3 + gsl_partVanish_21.y; }
    if (abs(gsl_part_21) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_21), gsl_partVanish_21.zw); }
  }
  let screwTX = gsl_part_21;
  let gsl_partLo_22 = vec3<f32>(-33.177418, 1.3411463, 10.806372);
  let gsl_partHi_22 = vec3<f32>(-26.822584, 8.658854, 35.82637);
  let gsl_partBox_22 = gsl_boxDistance(p, gsl_partLo_22, gsl_partHi_22);
  var gsl_part_22 = gsl_partBox_22;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_22 <= gsl_vanishView.w) {
    let gsl_partVanish_22 = gsl_vanish(gsl_partLo_22, gsl_partHi_22);
    gsl_part_22 = 1e10;
    if (gsl_partVanish_22.x > 0.0) { gsl_part_22 = gsl_remap(p - vec3<f32>(-30.0, 5.0, 10.816372)) + gsl_partVanish_22.y; }
    if (abs(gsl_part_22) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_22), gsl_partVanish_22.zw); }
  }
  let seatedPH = gsl_part_22;
  let gsl_partLo_23 = vec3<f32>(-3.233512, 1.3411463, 10.8341255);
  let gsl_partHi_23 = vec3<f32>(3.233512, 8.658854, 35.854126);
  let gsl_partBox_23 = gsl_boxDistance(p, gsl_partLo_23, gsl_partHi_23);
  var gsl_part_23 = gsl_partBox_23;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_23 <= gsl_vanishView.w) {
    let gsl_partVanish_23 = gsl_vanish(gsl_partLo_23, gsl_partHi_23);
    gsl_part_23 = 1e10;
    if (gsl_partVanish_23.x > 0.0) { gsl_part_23 = gsl_remap_3(p - vec3<f32>(0.0, 5.0, 10.844126)) + gsl_partVanish_23.y; }
    if (abs(gsl_part_23) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_23), gsl_partVanish_23.zw); }
  }
  let seatedPZ = gsl_part_23;
  let gsl_partLo_24 = vec3<f32>(26.83, 1.3411463, 11.457076);
  let gsl_partHi_24 = vec3<f32>(33.17, 8.658854, 36.477077);
  let gsl_partBox_24 = gsl_boxDistance(p, gsl_partLo_24, gsl_partHi_24);
  var gsl_part_24 = gsl_partBox_24;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_24 <= gsl_vanishView.w) {
    let gsl_partVanish_24 = gsl_vanish(gsl_partLo_24, gsl_partHi_24);
    gsl_part_24 = 1e10;
    if (gsl_partVanish_24.x > 0.0) { gsl_part_24 = gsl_remap_5(p - vec3<f32>(30.0, 5.0, 11.467076)) + gsl_partVanish_24.y; }
    if (abs(gsl_part_24) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_24), gsl_partVanish_24.zw); }
  }
  let seatedTX = gsl_part_24;
  let gsl_partLo_25 = vec3<f32>(-110.01, 32.74, 12.74);
  let gsl_partHi_25 = vec3<f32>(-7.21, 37.26, 17.26);
  let gsl_partBox_25 = gsl_boxDistance(p, gsl_partLo_25, gsl_partHi_25);
  var gsl_part_25 = gsl_partBox_25;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_25 <= gsl_vanishView.w) {
    let gsl_partVanish_25 = gsl_vanish(gsl_partLo_25, gsl_partHi_25);
    gsl_part_25 = 1e10;
    if (gsl_partVanish_25.x > 0.0) { gsl_part_25 = gsl_remap_44(p - vec3<f32>(-110.0, 35.0, 15.0)) + gsl_partVanish_25.y; }
    if (abs(gsl_part_25) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_25), gsl_partVanish_25.zw); }
  }
  let bladePH1 = gsl_part_25;
  let gsl_partLo_26 = vec3<f32>(-110.01, 61.822582, 14.822583);
  let gsl_partHi_26 = vec3<f32>(19.38, 68.177414, 21.177416);
  let gsl_partBox_26 = gsl_boxDistance(p, gsl_partLo_26, gsl_partHi_26);
  var gsl_part_26 = gsl_partBox_26;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_26 <= gsl_vanishView.w) {
    let gsl_partVanish_26 = gsl_vanish(gsl_partLo_26, gsl_partHi_26);
    gsl_part_26 = 1e10;
    if (gsl_partVanish_26.x > 0.0) { gsl_part_26 = gsl_remap_46(p - vec3<f32>(-110.0, 65.0, 18.0)) + gsl_partVanish_26.y; }
    if (abs(gsl_part_26) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_26), gsl_partVanish_26.zw); }
  }
  let bladePH2 = gsl_part_26;
  let gsl_partLo_27 = vec3<f32>(-110.01, 96.76649, 14.766488);
  let gsl_partHi_27 = vec3<f32>(19.38, 103.23351, 21.233511);
  let gsl_partBox_27 = gsl_boxDistance(p, gsl_partLo_27, gsl_partHi_27);
  var gsl_part_27 = gsl_partBox_27;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_27 <= gsl_vanishView.w) {
    let gsl_partVanish_27 = gsl_vanish(gsl_partLo_27, gsl_partHi_27);
    gsl_part_27 = 1e10;
    if (gsl_partVanish_27.x > 0.0) { gsl_part_27 = gsl_remap_48(p - vec3<f32>(-110.0, 100.0, 18.0)) + gsl_partVanish_27.y; }
    if (abs(gsl_part_27) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_27), gsl_partVanish_27.zw); }
  }
  let bladePZ2 = gsl_part_27;
  let gsl_partLo_28 = vec3<f32>(-110.01, 131.74, 14.99);
  let gsl_partHi_28 = vec3<f32>(45.01, 138.26, 21.01);
  let gsl_partBox_28 = gsl_boxDistance(p, gsl_partLo_28, gsl_partHi_28);
  var gsl_part_28 = gsl_partBox_28;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_28 <= gsl_vanishView.w) {
    let gsl_partVanish_28 = gsl_vanish(gsl_partLo_28, gsl_partHi_28);
    gsl_part_28 = 1e10;
    if (gsl_partVanish_28.x > 0.0) { gsl_part_28 = gsl_remap_50(p - vec3<f32>(-110.0, 135.0, 18.0)) + gsl_partVanish_28.y; }
    if (abs(gsl_part_28) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_28), gsl_partVanish_28.zw); }
  }
  let bladeSlot = gsl_part_28;
  let gsl_partLo_29 = vec3<f32>(-32.23, 19.99, -0.01);
  let gsl_partHi_29 = vec3<f32>(78.79, 50.01, 30.01);
  let gsl_partBox_29 = gsl_boxDistance(p, gsl_partLo_29, gsl_partHi_29);
  var gsl_part_29 = gsl_partBox_29;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_29 <= gsl_vanishView.w) {
    let gsl_partVanish_29 = gsl_vanish(gsl_partLo_29, gsl_partHi_29);
    gsl_part_29 = 1e10;
    if (gsl_partVanish_29.x > 0.0) { gsl_part_29 = gsl_remap_51(p - vec3<f32>(-110.0, 35.0, 15.0)) + gsl_partVanish_29.y; }
    if (abs(gsl_part_29) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_29), gsl_partVanish_29.zw); }
  }
  let handlePH1 = gsl_part_29;
  let gsl_partLo_30 = vec3<f32>(-5.64, 46.99, -0.01);
  let gsl_partHi_30 = vec3<f32>(112.38, 83.01, 36.01);
  let gsl_partBox_30 = gsl_boxDistance(p, gsl_partLo_30, gsl_partHi_30);
  var gsl_part_30 = gsl_partBox_30;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_30 <= gsl_vanishView.w) {
    let gsl_partVanish_30 = gsl_vanish(gsl_partLo_30, gsl_partHi_30);
    gsl_part_30 = 1e10;
    if (gsl_partVanish_30.x > 0.0) { gsl_part_30 = gsl_remap_52(p - vec3<f32>(-110.0, 65.0, 18.0)) + gsl_partVanish_30.y; }
    if (abs(gsl_part_30) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_30), gsl_partVanish_30.zw); }
  }
  let handlePH2 = gsl_part_30;
  let gsl_partLo_31 = vec3<f32>(-5.64, 81.99, -0.01);
  let gsl_partHi_31 = vec3<f32>(112.38, 118.01, 36.01);
  let gsl_partBox_31 = gsl_boxDistance(p, gsl_partLo_31, gsl_partHi_31);
  var gsl_part_31 = gsl_partBox_31;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_31 <= gsl_vanishView.w) {
    let gsl_partVanish_31 = gsl_vanish(gsl_partLo_31, gsl_partHi_31);
    gsl_part_31 = 1e10;
    if (gsl_partVanish_31.x > 0.0) { gsl_part_31 = gsl_remap_52(p - vec3<f32>(-110.0, 100.0, 18.0)) + gsl_partVanish_31.y; }
    if (abs(gsl_part_31) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_31), gsl_partVanish_31.zw); }
  }
  let handlePZ2 = gsl_part_31;
  let gsl_partLo_32 = vec3<f32>(19.99, 116.99, -0.01);
  let gsl_partHi_32 = vec3<f32>(138.01, 153.01, 36.01);
  let gsl_partBox_32 = gsl_boxDistance(p, gsl_partLo_32, gsl_partHi_32);
  var gsl_part_32 = gsl_partBox_32;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_32 <= gsl_vanishView.w) {
    let gsl_partVanish_32 = gsl_vanish(gsl_partLo_32, gsl_partHi_32);
    gsl_part_32 = 1e10;
    if (gsl_partVanish_32.x > 0.0) { gsl_part_32 = gsl_remap_53(p - vec3<f32>(-110.0, 135.0, 18.0)) + gsl_partVanish_32.y; }
    if (abs(gsl_part_32) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_32), gsl_partVanish_32.zw); }
  }
  let handleSlot = gsl_part_32;
  return array<f32, 32>(insertPH2, insertPZ2, insertT25, insertR2, insertHex4, insertSlot, powerPH2, powerPZ2, powerT25, powerR2, powerHex4, powerSlot, microPH0, microPH1, microT6, microT10, microHex15, microSlot, screwPH, screwPZ, screwTX, seatedPH, seatedPZ, seatedTX, bladePH1, bladePH2, bladePZ2, bladeSlot, handlePH1, handlePH2, handlePZ2, handleSlot);
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
  let gsl_s9 = gsl_shape(gsl_d[8], 9u);
  let gsl_s10 = gsl_shape(gsl_d[9], 10u);
  let gsl_s11 = gsl_shape(gsl_d[10], 11u);
  let gsl_s12 = gsl_shape(gsl_d[11], 12u);
  let gsl_s13 = gsl_shape(gsl_d[12], 13u);
  let gsl_s14 = gsl_shape(gsl_d[13], 14u);
  let gsl_s15 = gsl_shape(gsl_d[14], 15u);
  let gsl_s16 = gsl_shape(gsl_d[15], 16u);
  let gsl_s17 = gsl_shape(gsl_d[16], 17u);
  let gsl_s18 = gsl_shape(gsl_d[17], 18u);
  let gsl_s19 = gsl_shape(gsl_d[18], 19u);
  let gsl_s20 = gsl_shape(gsl_d[19], 20u);
  let gsl_s21 = gsl_shape(gsl_d[20], 21u);
  let gsl_s22 = gsl_shape(gsl_d[21], 22u);
  let gsl_s23 = gsl_shape(gsl_d[22], 23u);
  let gsl_s24 = gsl_shape(gsl_d[23], 24u);
  let gsl_s25 = gsl_shape(gsl_d[24], 25u);
  let gsl_s26 = gsl_shape(gsl_d[25], 26u);
  let gsl_s27 = gsl_shape(gsl_d[26], 27u);
  let gsl_s28 = gsl_shape(gsl_d[27], 28u);
  let gsl_s29 = gsl_shape(gsl_d[28], 29u);
  let gsl_s30 = gsl_shape(gsl_d[29], 30u);
  let gsl_s31 = gsl_shape(gsl_d[30], 31u);
  let gsl_s32 = gsl_shape(gsl_d[31], 32u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21), gsl_s22), gsl_s23), gsl_s24), gsl_s25), gsl_s26), gsl_s27), gsl_s28), gsl_s29), gsl_s30), gsl_s31), gsl_s32);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 9u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 10u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 11u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 12u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 13u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 14u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 15u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 16u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 17u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 18u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 19u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.9103, 0.928, 0.9381), thread_maleRelief(vec3<f32>(p.x + 30.0, p.y - 5.0, p.z - 10.0), 4.0, 0.7, 0.0, -10.0), 0.7), 1.0); }
  if (gsl_colorIdx == 20u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.9103, 0.928, 0.9381), thread_maleRelief(vec3<f32>(p.x, p.y - 5.0, p.z - 10.0), 4.0, 0.7, 0.0, -10.0), 0.7), 1.0); }
  if (gsl_colorIdx == 21u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.9103, 0.928, 0.9381), thread_maleRelief(vec3<f32>(p.x - 30.0, p.y - 5.0, p.z - 10.0), 5.0, 0.8, 0.0, -10.0), 0.8), 1.0); }
  if (gsl_colorIdx == 22u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 23u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 24u) { return vec4<f32>(vec3<f32>(0.46, 0.47, 0.5), 1.0); }
  if (gsl_colorIdx == 25u) { return vec4<f32>(vec3<f32>(0.8289, 0.8461, 0.8548), 1.0); }
  if (gsl_colorIdx == 26u) { return vec4<f32>(vec3<f32>(0.8289, 0.8461, 0.8548), 1.0); }
  if (gsl_colorIdx == 27u) { return vec4<f32>(vec3<f32>(0.8289, 0.8461, 0.8548), 1.0); }
  if (gsl_colorIdx == 28u) { return vec4<f32>(vec3<f32>(0.8289, 0.8461, 0.8548), 1.0); }
  if (gsl_colorIdx == 29u) {
    let color_red = vec3<f32>(1.0, 0.0, 0.0);
    return vec4<f32>(color_red, 1.0);
  }
  if (gsl_colorIdx == 30u) {
    let color_red = vec3<f32>(1.0, 0.0, 0.0);
    return vec4<f32>(color_red, 1.0);
  }
  if (gsl_colorIdx == 31u) {
    let color_yellow = vec3<f32>(1.0, 1.0, 0.0);
    return vec4<f32>(color_yellow, 1.0);
  }
  if (gsl_colorIdx == 32u) {
    let color_blue = vec3<f32>(0.0, 0.0, 1.0);
    return vec4<f32>(color_blue, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 9u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 10u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 11u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 12u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 13u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 14u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 15u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 16u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 17u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 18u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 19u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(vec3<f32>(p.x + 30.0, p.y - 5.0, p.z - 10.0), 4.0, 0.7, 0.0, -10.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 20u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(vec3<f32>(p.x, p.y - 5.0, p.z - 10.0), 4.0, 0.7, 0.0, -10.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 21u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(vec3<f32>(p.x - 30.0, p.y - 5.0, p.z - 10.0), 5.0, 0.8, 0.0, -10.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 22u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 23u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 24u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 25u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.04, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 26u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.04, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 27u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.04, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 28u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.04, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 29u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_plastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 30u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_plastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 31u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_plastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 32u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_plastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

var<private> gsl_simpleRung: f32;

fn gsl_simple(lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  if (gsl_vanishEye.w <= 0.0 || gsl_simpleRung <= 0.0) { return 1.0; }
  let px = gsl_vanish(lo, hi).z;
  return smoothstep(0.5 * gsl_simpleRung, gsl_simpleRung, px);
}

fn gsl_boxSdf(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  let d = abs(q - 0.5 * (lo + hi)) - 0.5 * (hi - lo);
  return length(max(d, vec3<f32>(0.0))) + min(max(d.x, max(d.y, d.z)), 0.0);
}

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;

fn gsl_heights(p: vec3<f32>, time: f32) -> array<f32, 32> {
  return array<f32, 32>(0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(vec3<f32>(p.x + 30.0, p.y - 5.0, p.z - 10.0), 4.0, 0.7, 0.0, -10.0))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(vec3<f32>(p.x, p.y - 5.0, p.z - 10.0), 4.0, 0.7, 0.0, -10.0))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(vec3<f32>(p.x - 30.0, p.y - 5.0, p.z - 10.0), 5.0, 0.8, 0.0, -10.0))[2].z, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_h = gsl_heights(p, time);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7], 8u);
  let gsl_s9 = gsl_shape(gsl_d[8], 9u);
  let gsl_s10 = gsl_shape(gsl_d[9], 10u);
  let gsl_s11 = gsl_shape(gsl_d[10], 11u);
  let gsl_s12 = gsl_shape(gsl_d[11], 12u);
  let gsl_s13 = gsl_shape(gsl_d[12], 13u);
  let gsl_s14 = gsl_shape(gsl_d[13], 14u);
  let gsl_s15 = gsl_shape(gsl_d[14], 15u);
  let gsl_s16 = gsl_shape(gsl_d[15], 16u);
  let gsl_s17 = gsl_shape(gsl_d[16], 17u);
  let gsl_s18 = gsl_shape(gsl_d[17], 18u);
  let gsl_s19 = gsl_shape(gsl_d[18] - gsl_h[18] * gsl_dispK, 19u);
  let gsl_s20 = gsl_shape(gsl_d[19] - gsl_h[19] * gsl_dispK, 20u);
  let gsl_s21 = gsl_shape(gsl_d[20] - gsl_h[20] * gsl_dispK, 21u);
  let gsl_s22 = gsl_shape(gsl_d[21], 22u);
  let gsl_s23 = gsl_shape(gsl_d[22], 23u);
  let gsl_s24 = gsl_shape(gsl_d[23], 24u);
  let gsl_s25 = gsl_shape(gsl_d[24], 25u);
  let gsl_s26 = gsl_shape(gsl_d[25], 26u);
  let gsl_s27 = gsl_shape(gsl_d[26], 27u);
  let gsl_s28 = gsl_shape(gsl_d[27], 28u);
  let gsl_s29 = gsl_shape(gsl_d[28], 29u);
  let gsl_s30 = gsl_shape(gsl_d[29], 30u);
  let gsl_s31 = gsl_shape(gsl_d[30], 31u);
  let gsl_s32 = gsl_shape(gsl_d[31], 32u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21), gsl_s22), gsl_s23), gsl_s24), gsl_s25), gsl_s26), gsl_s27), gsl_s28), gsl_s29), gsl_s30), gsl_s31), gsl_s32);
}
