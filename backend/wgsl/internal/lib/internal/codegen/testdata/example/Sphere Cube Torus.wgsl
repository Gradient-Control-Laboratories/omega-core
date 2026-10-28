fn builtin_torus(p: vec3<f32>, majorRadius: f32, minorRadius: f32) -> f32 {
  let radial = length(p.xy) - majorRadius;
  return length(vec2<f32>(radial, p.z)) - minorRadius;
}

fn builtin_unionSharp(d1: f32, d2: f32, d3: f32) -> f32 {
  return min(min(d1, d2), d3);
}

fn builtin_linearMap(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: vec3<f32>, outMax: vec3<f32>) -> vec3<f32> {
  let slope = (outMax - outMin) / (inMax - inMin);
  return slope * (gsl_var_value - inMin) + outMin;
}

fn builtin_ramp(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: vec3<f32>, outMax: vec3<f32>) -> vec3<f32> {
  let mapped = builtin_linearMap(gsl_var_value, inMin, inMax, outMin, outMax);
  let low = min(outMin, outMax);
  let high = max(outMin, outMax);
  return clamp(mapped, low, high);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
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

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let s = length(p) - 0.8;
  let r = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.3, 0.3, 0.0), vec3<f32>(1.5, 1.5, 1.2));
  let t = builtin_torus(p, 1.3, 0.2);
  let combined = builtin_unionSharp(s, r, t);
  let gsl_s1 = gsl_shape(combined, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let s = length(p) - 0.8;
    let r = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.3, 0.3, 0.0), vec3<f32>(1.5, 1.5, 1.2));
    let t = builtin_torus(p, 1.3, 0.2);
    let combined = builtin_unionSharp(s, r, t);
    let gsl_tmp_1 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let edges = length(vec2<f32>(combined, select(min(s, r), min(t, select(min(s, select(max(gsl_tmp_1.x, gsl_tmp_1.y), max(gsl_tmp_1.z, min(gsl_tmp_1.x, gsl_tmp_1.y)), max(gsl_tmp_1.x, gsl_tmp_1.y) > gsl_tmp_1.z)), r, s < r)), min(s, r) < t)));
    let gsl_tmp_8 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_3 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
    let gsl_tmp_2 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_4 = select(0.0, gsl_tmp_3.x, gsl_tmp_2.x + 0.0 > 0.0);
    let gsl_tmp_5 = select(0.0, gsl_tmp_3.y, gsl_tmp_2.y + 0.0 > 0.0);
    let gsl_tmp_6 = select(0.0, gsl_tmp_3.z, gsl_tmp_2.z + 0.0 > 0.0);
    let gsl_tmp_7 = sqrt(0.0 + max(gsl_tmp_2.x + 0.0, 0.0) * max(gsl_tmp_2.x + 0.0, 0.0) + max(gsl_tmp_2.y + 0.0, 0.0) * max(gsl_tmp_2.y + 0.0, 0.0) + max(gsl_tmp_2.z + 0.0, 0.0) * max(gsl_tmp_2.z + 0.0, 0.0));
    let gsl_tmp_15 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_10 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
    let gsl_tmp_9 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_11 = select(0.0, gsl_tmp_10.x, gsl_tmp_9.x + 0.0 > 0.0);
    let gsl_tmp_12 = select(0.0, gsl_tmp_10.y, gsl_tmp_9.y + 0.0 > 0.0);
    let gsl_tmp_13 = select(0.0, gsl_tmp_10.z, gsl_tmp_9.z + 0.0 > 0.0);
    let gsl_tmp_14 = sqrt(0.0 + max(gsl_tmp_9.x + 0.0, 0.0) * max(gsl_tmp_9.x + 0.0, 0.0) + max(gsl_tmp_9.y + 0.0, 0.0) * max(gsl_tmp_9.y + 0.0, 0.0) + max(gsl_tmp_9.z + 0.0, 0.0) * max(gsl_tmp_9.z + 0.0, 0.0));
    let gsl_tmp_22 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_17 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
    let gsl_tmp_16 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_18 = select(0.0, gsl_tmp_17.x, gsl_tmp_16.x + 0.0 > 0.0);
    let gsl_tmp_19 = select(0.0, gsl_tmp_17.y, gsl_tmp_16.y + 0.0 > 0.0);
    let gsl_tmp_20 = select(0.0, gsl_tmp_17.z, gsl_tmp_16.z + 0.0 > 0.0);
    let gsl_tmp_21 = sqrt(0.0 + max(gsl_tmp_16.x + 0.0, 0.0) * max(gsl_tmp_16.x + 0.0, 0.0) + max(gsl_tmp_16.y + 0.0, 0.0) * max(gsl_tmp_16.y + 0.0, 0.0) + max(gsl_tmp_16.z + 0.0, 0.0) * max(gsl_tmp_16.z + 0.0, 0.0));
    let gsl_tmp_28 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
    let gsl_tmp_27 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_29 = select(0.0, gsl_tmp_28.x, gsl_tmp_27.x + 0.0 > 0.0);
    let gsl_tmp_30 = select(0.0, gsl_tmp_28.y, gsl_tmp_27.y + 0.0 > 0.0);
    let gsl_tmp_31 = select(0.0, gsl_tmp_28.z, gsl_tmp_27.z + 0.0 > 0.0);
    let gsl_tmp_32 = sqrt(0.0 + max(gsl_tmp_27.x + 0.0, 0.0) * max(gsl_tmp_27.x + 0.0, 0.0) + max(gsl_tmp_27.y + 0.0, 0.0) * max(gsl_tmp_27.y + 0.0, 0.0) + max(gsl_tmp_27.z + 0.0, 0.0) * max(gsl_tmp_27.z + 0.0, 0.0));
    let gsl_tmp_33 = select(select(gsl_tmp_28.z, select(gsl_tmp_28.y, gsl_tmp_28.x, gsl_tmp_27.x > gsl_tmp_27.y), max(gsl_tmp_27.x, gsl_tmp_27.y) > gsl_tmp_27.z), -0.0, -0.0 < max(max(gsl_tmp_27.x, gsl_tmp_27.y), gsl_tmp_27.z)) + (gsl_tmp_29 * max(gsl_tmp_27.x + 0.0, 0.0) + max(gsl_tmp_27.x + 0.0, 0.0) * gsl_tmp_29 + gsl_tmp_30 * max(gsl_tmp_27.y + 0.0, 0.0) + max(gsl_tmp_27.y + 0.0, 0.0) * gsl_tmp_30 + gsl_tmp_31 * max(gsl_tmp_27.z + 0.0, 0.0) + max(gsl_tmp_27.z + 0.0, 0.0) * gsl_tmp_31) / (2.0 * (gsl_tmp_32 + 1.0 - select(0.0, 1.0, (gsl_tmp_32 > 0.0))));
    let gsl_tmp_34 = dot(p, vec3<f32>(1.0, 0.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0)));
    let gsl_tmp_35 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
    let gsl_tmp_23 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_24 = max(gsl_tmp_23.x, gsl_tmp_23.y) > gsl_tmp_23.z;
    let gsl_tmp_25 = select(max(gsl_tmp_23.x, gsl_tmp_23.y), max(gsl_tmp_23.z, min(gsl_tmp_23.x, gsl_tmp_23.y)), gsl_tmp_24);
    let gsl_tmp_26 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_41 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
    let gsl_tmp_40 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_42 = select(0.0, gsl_tmp_41.x, gsl_tmp_40.x + 0.0 > 0.0);
    let gsl_tmp_43 = select(0.0, gsl_tmp_41.y, gsl_tmp_40.y + 0.0 > 0.0);
    let gsl_tmp_44 = select(0.0, gsl_tmp_41.z, gsl_tmp_40.z + 0.0 > 0.0);
    let gsl_tmp_45 = sqrt(0.0 + max(gsl_tmp_40.x + 0.0, 0.0) * max(gsl_tmp_40.x + 0.0, 0.0) + max(gsl_tmp_40.y + 0.0, 0.0) * max(gsl_tmp_40.y + 0.0, 0.0) + max(gsl_tmp_40.z + 0.0, 0.0) * max(gsl_tmp_40.z + 0.0, 0.0));
    let gsl_tmp_46 = select(select(gsl_tmp_41.z, select(gsl_tmp_41.y, gsl_tmp_41.x, gsl_tmp_40.x > gsl_tmp_40.y), max(gsl_tmp_40.x, gsl_tmp_40.y) > gsl_tmp_40.z), -0.0, -0.0 < max(max(gsl_tmp_40.x, gsl_tmp_40.y), gsl_tmp_40.z)) + (gsl_tmp_42 * max(gsl_tmp_40.x + 0.0, 0.0) + max(gsl_tmp_40.x + 0.0, 0.0) * gsl_tmp_42 + gsl_tmp_43 * max(gsl_tmp_40.y + 0.0, 0.0) + max(gsl_tmp_40.y + 0.0, 0.0) * gsl_tmp_43 + gsl_tmp_44 * max(gsl_tmp_40.z + 0.0, 0.0) + max(gsl_tmp_40.z + 0.0, 0.0) * gsl_tmp_44) / (2.0 * (gsl_tmp_45 + 1.0 - select(0.0, 1.0, (gsl_tmp_45 > 0.0))));
    let gsl_tmp_47 = dot(p, vec3<f32>(0.0, 1.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0)));
    let gsl_tmp_48 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
    let gsl_tmp_36 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_37 = max(gsl_tmp_36.x, gsl_tmp_36.y) > gsl_tmp_36.z;
    let gsl_tmp_38 = select(max(gsl_tmp_36.x, gsl_tmp_36.y), max(gsl_tmp_36.z, min(gsl_tmp_36.x, gsl_tmp_36.y)), gsl_tmp_37);
    let gsl_tmp_39 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_54 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
    let gsl_tmp_53 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_55 = select(0.0, gsl_tmp_54.x, gsl_tmp_53.x + 0.0 > 0.0);
    let gsl_tmp_56 = select(0.0, gsl_tmp_54.y, gsl_tmp_53.y + 0.0 > 0.0);
    let gsl_tmp_57 = select(0.0, gsl_tmp_54.z, gsl_tmp_53.z + 0.0 > 0.0);
    let gsl_tmp_58 = sqrt(0.0 + max(gsl_tmp_53.x + 0.0, 0.0) * max(gsl_tmp_53.x + 0.0, 0.0) + max(gsl_tmp_53.y + 0.0, 0.0) * max(gsl_tmp_53.y + 0.0, 0.0) + max(gsl_tmp_53.z + 0.0, 0.0) * max(gsl_tmp_53.z + 0.0, 0.0));
    let gsl_tmp_59 = select(select(gsl_tmp_54.z, select(gsl_tmp_54.y, gsl_tmp_54.x, gsl_tmp_53.x > gsl_tmp_53.y), max(gsl_tmp_53.x, gsl_tmp_53.y) > gsl_tmp_53.z), -0.0, -0.0 < max(max(gsl_tmp_53.x, gsl_tmp_53.y), gsl_tmp_53.z)) + (gsl_tmp_55 * max(gsl_tmp_53.x + 0.0, 0.0) + max(gsl_tmp_53.x + 0.0, 0.0) * gsl_tmp_55 + gsl_tmp_56 * max(gsl_tmp_53.y + 0.0, 0.0) + max(gsl_tmp_53.y + 0.0, 0.0) * gsl_tmp_56 + gsl_tmp_57 * max(gsl_tmp_53.z + 0.0, 0.0) + max(gsl_tmp_53.z + 0.0, 0.0) * gsl_tmp_57) / (2.0 * (gsl_tmp_58 + 1.0 - select(0.0, 1.0, (gsl_tmp_58 > 0.0))));
    let gsl_tmp_60 = dot(p, vec3<f32>(0.0, 0.0, 1.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0)));
    let gsl_tmp_61 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
    let gsl_tmp_49 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_50 = max(gsl_tmp_49.x, gsl_tmp_49.y) > gsl_tmp_49.z;
    let gsl_tmp_51 = select(max(gsl_tmp_49.x, gsl_tmp_49.y), max(gsl_tmp_49.z, min(gsl_tmp_49.x, gsl_tmp_49.y)), gsl_tmp_50);
    let gsl_tmp_52 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let scale = max(abs(dot(vec3<f32>(select(dot(gsl_tmp_8, vec2<f32>(dot(p.xy, vec2<f32>(1.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_8) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_8) > 0.0))), select(select(select(gsl_tmp_3.z, select(gsl_tmp_3.y, gsl_tmp_3.x, gsl_tmp_2.x > gsl_tmp_2.y), max(gsl_tmp_2.x, gsl_tmp_2.y) > gsl_tmp_2.z), -0.0, -0.0 < max(max(gsl_tmp_2.x, gsl_tmp_2.y), gsl_tmp_2.z)) + (gsl_tmp_4 * max(gsl_tmp_2.x + 0.0, 0.0) + max(gsl_tmp_2.x + 0.0, 0.0) * gsl_tmp_4 + gsl_tmp_5 * max(gsl_tmp_2.y + 0.0, 0.0) + max(gsl_tmp_2.y + 0.0, 0.0) * gsl_tmp_5 + gsl_tmp_6 * max(gsl_tmp_2.z + 0.0, 0.0) + max(gsl_tmp_2.z + 0.0, 0.0) * gsl_tmp_6) / (2.0 * (gsl_tmp_7 + 1.0 - select(0.0, 1.0, (gsl_tmp_7 > 0.0)))), dot(p, vec3<f32>(1.0, 0.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), s < r), min(s, r) < t), select(dot(gsl_tmp_15, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 1.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_15) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_15) > 0.0))), select(select(select(gsl_tmp_10.z, select(gsl_tmp_10.y, gsl_tmp_10.x, gsl_tmp_9.x > gsl_tmp_9.y), max(gsl_tmp_9.x, gsl_tmp_9.y) > gsl_tmp_9.z), -0.0, -0.0 < max(max(gsl_tmp_9.x, gsl_tmp_9.y), gsl_tmp_9.z)) + (gsl_tmp_11 * max(gsl_tmp_9.x + 0.0, 0.0) + max(gsl_tmp_9.x + 0.0, 0.0) * gsl_tmp_11 + gsl_tmp_12 * max(gsl_tmp_9.y + 0.0, 0.0) + max(gsl_tmp_9.y + 0.0, 0.0) * gsl_tmp_12 + gsl_tmp_13 * max(gsl_tmp_9.z + 0.0, 0.0) + max(gsl_tmp_9.z + 0.0, 0.0) * gsl_tmp_13) / (2.0 * (gsl_tmp_14 + 1.0 - select(0.0, 1.0, (gsl_tmp_14 > 0.0)))), dot(p, vec3<f32>(0.0, 1.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), s < r), min(s, r) < t), select(dot(gsl_tmp_22, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 1.0)) / (length(gsl_tmp_22) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_22) > 0.0))), select(select(select(gsl_tmp_17.z, select(gsl_tmp_17.y, gsl_tmp_17.x, gsl_tmp_16.x > gsl_tmp_16.y), max(gsl_tmp_16.x, gsl_tmp_16.y) > gsl_tmp_16.z), -0.0, -0.0 < max(max(gsl_tmp_16.x, gsl_tmp_16.y), gsl_tmp_16.z)) + (gsl_tmp_18 * max(gsl_tmp_16.x + 0.0, 0.0) + max(gsl_tmp_16.x + 0.0, 0.0) * gsl_tmp_18 + gsl_tmp_19 * max(gsl_tmp_16.y + 0.0, 0.0) + max(gsl_tmp_16.y + 0.0, 0.0) * gsl_tmp_19 + gsl_tmp_20 * max(gsl_tmp_16.z + 0.0, 0.0) + max(gsl_tmp_16.z + 0.0, 0.0) * gsl_tmp_20) / (2.0 * (gsl_tmp_21 + 1.0 - select(0.0, 1.0, (gsl_tmp_21 > 0.0)))), dot(p, vec3<f32>(0.0, 0.0, 1.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), s < r), min(s, r) < t)), vec3<f32>(select(select(gsl_tmp_33, gsl_tmp_34, s < r), select(select(select(select(select(gsl_tmp_35.y, gsl_tmp_35.x, gsl_tmp_23.x > gsl_tmp_23.y), select(select(gsl_tmp_35.y, gsl_tmp_35.x, gsl_tmp_23.x < gsl_tmp_23.y), gsl_tmp_35.z, gsl_tmp_23.z > min(gsl_tmp_23.x, gsl_tmp_23.y)), gsl_tmp_24), gsl_tmp_34, s < gsl_tmp_25), gsl_tmp_33, s < r), dot(gsl_tmp_26, vec2<f32>(dot(p.xy, vec2<f32>(1.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_26) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_26) > 0.0))), t < select(min(s, gsl_tmp_25), r, s < r)), min(s, r) < t), select(select(gsl_tmp_46, gsl_tmp_47, s < r), select(select(select(select(select(gsl_tmp_48.y, gsl_tmp_48.x, gsl_tmp_36.x > gsl_tmp_36.y), select(select(gsl_tmp_48.y, gsl_tmp_48.x, gsl_tmp_36.x < gsl_tmp_36.y), gsl_tmp_48.z, gsl_tmp_36.z > min(gsl_tmp_36.x, gsl_tmp_36.y)), gsl_tmp_37), gsl_tmp_47, s < gsl_tmp_38), gsl_tmp_46, s < r), dot(gsl_tmp_39, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 1.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_39) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_39) > 0.0))), t < select(min(s, gsl_tmp_38), r, s < r)), min(s, r) < t), select(select(gsl_tmp_59, gsl_tmp_60, s < r), select(select(select(select(select(gsl_tmp_61.y, gsl_tmp_61.x, gsl_tmp_49.x > gsl_tmp_49.y), select(select(gsl_tmp_61.y, gsl_tmp_61.x, gsl_tmp_49.x < gsl_tmp_49.y), gsl_tmp_61.z, gsl_tmp_49.z > min(gsl_tmp_49.x, gsl_tmp_49.y)), gsl_tmp_50), gsl_tmp_60, s < gsl_tmp_51), gsl_tmp_59, s < r), dot(gsl_tmp_52, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 1.0)) / (length(gsl_tmp_52) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_52) > 0.0))), t < select(min(s, gsl_tmp_51), r, s < r)), min(s, r) < t)))), 0.5);
    let color_darkGray = vec3<f32>(0.25, 0.25, 0.25);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    return vec4<f32>(builtin_ramp(edges, 0.005 * scale, 0.008 * scale, color_darkGray, color_white), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
