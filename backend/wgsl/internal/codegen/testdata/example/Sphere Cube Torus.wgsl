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
    let gsl_tmp_9 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_3 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
    let gsl_tmp_2 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_4 = select(gsl_tmp_3.z, select(gsl_tmp_3.y, gsl_tmp_3.x, gsl_tmp_2.x > gsl_tmp_2.y), max(gsl_tmp_2.x, gsl_tmp_2.y) > gsl_tmp_2.z);
    let gsl_tmp_6 = select(0.0, gsl_tmp_3.x, gsl_tmp_2.x + 0.0 > 0.0);
    let gsl_tmp_7 = select(0.0, gsl_tmp_3.y, gsl_tmp_2.y + 0.0 > 0.0);
    let gsl_tmp_8 = select(0.0, gsl_tmp_3.z, gsl_tmp_2.z + 0.0 > 0.0);
    let gsl_tmp_5 = 0.0 + max(gsl_tmp_2.x + 0.0, 0.0) * max(gsl_tmp_2.x + 0.0, 0.0) + max(gsl_tmp_2.y + 0.0, 0.0) * max(gsl_tmp_2.y + 0.0, 0.0) + max(gsl_tmp_2.z + 0.0, 0.0) * max(gsl_tmp_2.z + 0.0, 0.0);
    let gsl_tmp_17 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_11 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
    let gsl_tmp_10 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_12 = select(gsl_tmp_11.z, select(gsl_tmp_11.y, gsl_tmp_11.x, gsl_tmp_10.x > gsl_tmp_10.y), max(gsl_tmp_10.x, gsl_tmp_10.y) > gsl_tmp_10.z);
    let gsl_tmp_14 = select(0.0, gsl_tmp_11.x, gsl_tmp_10.x + 0.0 > 0.0);
    let gsl_tmp_15 = select(0.0, gsl_tmp_11.y, gsl_tmp_10.y + 0.0 > 0.0);
    let gsl_tmp_16 = select(0.0, gsl_tmp_11.z, gsl_tmp_10.z + 0.0 > 0.0);
    let gsl_tmp_13 = 0.0 + max(gsl_tmp_10.x + 0.0, 0.0) * max(gsl_tmp_10.x + 0.0, 0.0) + max(gsl_tmp_10.y + 0.0, 0.0) * max(gsl_tmp_10.y + 0.0, 0.0) + max(gsl_tmp_10.z + 0.0, 0.0) * max(gsl_tmp_10.z + 0.0, 0.0);
    let gsl_tmp_25 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_19 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
    let gsl_tmp_18 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_20 = select(gsl_tmp_19.z, select(gsl_tmp_19.y, gsl_tmp_19.x, gsl_tmp_18.x > gsl_tmp_18.y), max(gsl_tmp_18.x, gsl_tmp_18.y) > gsl_tmp_18.z);
    let gsl_tmp_22 = select(0.0, gsl_tmp_19.x, gsl_tmp_18.x + 0.0 > 0.0);
    let gsl_tmp_23 = select(0.0, gsl_tmp_19.y, gsl_tmp_18.y + 0.0 > 0.0);
    let gsl_tmp_24 = select(0.0, gsl_tmp_19.z, gsl_tmp_18.z + 0.0 > 0.0);
    let gsl_tmp_21 = 0.0 + max(gsl_tmp_18.x + 0.0, 0.0) * max(gsl_tmp_18.x + 0.0, 0.0) + max(gsl_tmp_18.y + 0.0, 0.0) * max(gsl_tmp_18.y + 0.0, 0.0) + max(gsl_tmp_18.z + 0.0, 0.0) * max(gsl_tmp_18.z + 0.0, 0.0);
    let gsl_tmp_31 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
    let gsl_tmp_30 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_32 = select(gsl_tmp_31.z, select(gsl_tmp_31.y, gsl_tmp_31.x, gsl_tmp_30.x > gsl_tmp_30.y), max(gsl_tmp_30.x, gsl_tmp_30.y) > gsl_tmp_30.z);
    let gsl_tmp_34 = select(0.0, gsl_tmp_31.x, gsl_tmp_30.x + 0.0 > 0.0);
    let gsl_tmp_35 = select(0.0, gsl_tmp_31.y, gsl_tmp_30.y + 0.0 > 0.0);
    let gsl_tmp_36 = select(0.0, gsl_tmp_31.z, gsl_tmp_30.z + 0.0 > 0.0);
    let gsl_tmp_33 = 0.0 + max(gsl_tmp_30.x + 0.0, 0.0) * max(gsl_tmp_30.x + 0.0, 0.0) + max(gsl_tmp_30.y + 0.0, 0.0) * max(gsl_tmp_30.y + 0.0, 0.0) + max(gsl_tmp_30.z + 0.0, 0.0) * max(gsl_tmp_30.z + 0.0, 0.0);
    let gsl_tmp_37 = gsl_tmp_32 - select(0.0, gsl_tmp_32, max(max(gsl_tmp_30.x, gsl_tmp_30.y), gsl_tmp_30.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_34 * max(gsl_tmp_30.x + 0.0, 0.0) + max(gsl_tmp_30.x + 0.0, 0.0) * gsl_tmp_34 + gsl_tmp_35 * max(gsl_tmp_30.y + 0.0, 0.0) + max(gsl_tmp_30.y + 0.0, 0.0) * gsl_tmp_35 + gsl_tmp_36 * max(gsl_tmp_30.z + 0.0, 0.0) + max(gsl_tmp_30.z + 0.0, 0.0) * gsl_tmp_36, gsl_tmp_33 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_33, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_33, 0.000000000000000000000000000001)) > 0.0))));
    let gsl_tmp_38 = dot(p, vec3<f32>(1.0, 0.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0)));
    let gsl_tmp_39 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(1.0, 0.0, 0.0);
    let gsl_tmp_26 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_27 = max(gsl_tmp_26.x, gsl_tmp_26.y) > gsl_tmp_26.z;
    let gsl_tmp_28 = select(max(gsl_tmp_26.x, gsl_tmp_26.y), max(gsl_tmp_26.z, min(gsl_tmp_26.x, gsl_tmp_26.y)), gsl_tmp_27);
    let gsl_tmp_29 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_45 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
    let gsl_tmp_44 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_46 = select(gsl_tmp_45.z, select(gsl_tmp_45.y, gsl_tmp_45.x, gsl_tmp_44.x > gsl_tmp_44.y), max(gsl_tmp_44.x, gsl_tmp_44.y) > gsl_tmp_44.z);
    let gsl_tmp_48 = select(0.0, gsl_tmp_45.x, gsl_tmp_44.x + 0.0 > 0.0);
    let gsl_tmp_49 = select(0.0, gsl_tmp_45.y, gsl_tmp_44.y + 0.0 > 0.0);
    let gsl_tmp_50 = select(0.0, gsl_tmp_45.z, gsl_tmp_44.z + 0.0 > 0.0);
    let gsl_tmp_47 = 0.0 + max(gsl_tmp_44.x + 0.0, 0.0) * max(gsl_tmp_44.x + 0.0, 0.0) + max(gsl_tmp_44.y + 0.0, 0.0) * max(gsl_tmp_44.y + 0.0, 0.0) + max(gsl_tmp_44.z + 0.0, 0.0) * max(gsl_tmp_44.z + 0.0, 0.0);
    let gsl_tmp_51 = gsl_tmp_46 - select(0.0, gsl_tmp_46, max(max(gsl_tmp_44.x, gsl_tmp_44.y), gsl_tmp_44.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_48 * max(gsl_tmp_44.x + 0.0, 0.0) + max(gsl_tmp_44.x + 0.0, 0.0) * gsl_tmp_48 + gsl_tmp_49 * max(gsl_tmp_44.y + 0.0, 0.0) + max(gsl_tmp_44.y + 0.0, 0.0) * gsl_tmp_49 + gsl_tmp_50 * max(gsl_tmp_44.z + 0.0, 0.0) + max(gsl_tmp_44.z + 0.0, 0.0) * gsl_tmp_50, gsl_tmp_47 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_47, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_47, 0.000000000000000000000000000001)) > 0.0))));
    let gsl_tmp_52 = dot(p, vec3<f32>(0.0, 1.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0)));
    let gsl_tmp_53 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 1.0, 0.0);
    let gsl_tmp_40 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_41 = max(gsl_tmp_40.x, gsl_tmp_40.y) > gsl_tmp_40.z;
    let gsl_tmp_42 = select(max(gsl_tmp_40.x, gsl_tmp_40.y), max(gsl_tmp_40.z, min(gsl_tmp_40.x, gsl_tmp_40.y)), gsl_tmp_41);
    let gsl_tmp_43 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let gsl_tmp_59 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
    let gsl_tmp_58 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_60 = select(gsl_tmp_59.z, select(gsl_tmp_59.y, gsl_tmp_59.x, gsl_tmp_58.x > gsl_tmp_58.y), max(gsl_tmp_58.x, gsl_tmp_58.y) > gsl_tmp_58.z);
    let gsl_tmp_62 = select(0.0, gsl_tmp_59.x, gsl_tmp_58.x + 0.0 > 0.0);
    let gsl_tmp_63 = select(0.0, gsl_tmp_59.y, gsl_tmp_58.y + 0.0 > 0.0);
    let gsl_tmp_64 = select(0.0, gsl_tmp_59.z, gsl_tmp_58.z + 0.0 > 0.0);
    let gsl_tmp_61 = 0.0 + max(gsl_tmp_58.x + 0.0, 0.0) * max(gsl_tmp_58.x + 0.0, 0.0) + max(gsl_tmp_58.y + 0.0, 0.0) * max(gsl_tmp_58.y + 0.0, 0.0) + max(gsl_tmp_58.z + 0.0, 0.0) * max(gsl_tmp_58.z + 0.0, 0.0);
    let gsl_tmp_65 = gsl_tmp_60 - select(0.0, gsl_tmp_60, max(max(gsl_tmp_58.x, gsl_tmp_58.y), gsl_tmp_58.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_62 * max(gsl_tmp_58.x + 0.0, 0.0) + max(gsl_tmp_58.x + 0.0, 0.0) * gsl_tmp_62 + gsl_tmp_63 * max(gsl_tmp_58.y + 0.0, 0.0) + max(gsl_tmp_58.y + 0.0, 0.0) * gsl_tmp_63 + gsl_tmp_64 * max(gsl_tmp_58.z + 0.0, 0.0) + max(gsl_tmp_58.z + 0.0, 0.0) * gsl_tmp_64, gsl_tmp_61 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_61, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_61, 0.000000000000000000000000000001)) > 0.0))));
    let gsl_tmp_66 = dot(p, vec3<f32>(0.0, 0.0, 1.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0)));
    let gsl_tmp_67 = sign(p - vec3<f32>(0.3, 0.3, 0.0)) * vec3<f32>(0.0, 0.0, 1.0);
    let gsl_tmp_54 = abs(p - vec3<f32>(0.3, 0.3, 0.0)) - vec3<f32>(0.75, 0.75, 0.6);
    let gsl_tmp_55 = max(gsl_tmp_54.x, gsl_tmp_54.y) > gsl_tmp_54.z;
    let gsl_tmp_56 = select(max(gsl_tmp_54.x, gsl_tmp_54.y), max(gsl_tmp_54.z, min(gsl_tmp_54.x, gsl_tmp_54.y)), gsl_tmp_55);
    let gsl_tmp_57 = vec2<f32>(length(p.xy) - 1.3, p.z);
    let scale = max(abs(dot(vec3<f32>(select(dot(gsl_tmp_9, vec2<f32>(dot(p.xy, vec2<f32>(1.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_9) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_9) > 0.0))), select(gsl_tmp_4 - select(0.0, gsl_tmp_4, max(max(gsl_tmp_2.x, gsl_tmp_2.y), gsl_tmp_2.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_6 * max(gsl_tmp_2.x + 0.0, 0.0) + max(gsl_tmp_2.x + 0.0, 0.0) * gsl_tmp_6 + gsl_tmp_7 * max(gsl_tmp_2.y + 0.0, 0.0) + max(gsl_tmp_2.y + 0.0, 0.0) * gsl_tmp_7 + gsl_tmp_8 * max(gsl_tmp_2.z + 0.0, 0.0) + max(gsl_tmp_2.z + 0.0, 0.0) * gsl_tmp_8, gsl_tmp_5 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_5, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_5, 0.000000000000000000000000000001)) > 0.0)))), dot(p, vec3<f32>(1.0, 0.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), s < r), min(s, r) < t), select(dot(gsl_tmp_17, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 1.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_17) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_17) > 0.0))), select(gsl_tmp_12 - select(0.0, gsl_tmp_12, max(max(gsl_tmp_10.x, gsl_tmp_10.y), gsl_tmp_10.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_14 * max(gsl_tmp_10.x + 0.0, 0.0) + max(gsl_tmp_10.x + 0.0, 0.0) * gsl_tmp_14 + gsl_tmp_15 * max(gsl_tmp_10.y + 0.0, 0.0) + max(gsl_tmp_10.y + 0.0, 0.0) * gsl_tmp_15 + gsl_tmp_16 * max(gsl_tmp_10.z + 0.0, 0.0) + max(gsl_tmp_10.z + 0.0, 0.0) * gsl_tmp_16, gsl_tmp_13 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_13, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_13, 0.000000000000000000000000000001)) > 0.0)))), dot(p, vec3<f32>(0.0, 1.0, 0.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), s < r), min(s, r) < t), select(dot(gsl_tmp_25, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 1.0)) / (length(gsl_tmp_25) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_25) > 0.0))), select(gsl_tmp_20 - select(0.0, gsl_tmp_20, max(max(gsl_tmp_18.x, gsl_tmp_18.y), gsl_tmp_18.z) + 0.0 > 0.0) + select(0.0, gsl_tmp_22 * max(gsl_tmp_18.x + 0.0, 0.0) + max(gsl_tmp_18.x + 0.0, 0.0) * gsl_tmp_22 + gsl_tmp_23 * max(gsl_tmp_18.y + 0.0, 0.0) + max(gsl_tmp_18.y + 0.0, 0.0) * gsl_tmp_23 + gsl_tmp_24 * max(gsl_tmp_18.z + 0.0, 0.0) + max(gsl_tmp_18.z + 0.0, 0.0) * gsl_tmp_24, gsl_tmp_21 > 0.000000000000000000000000000001) / (2.0 * (sqrt(max(gsl_tmp_21, 0.000000000000000000000000000001)) + 1.0 - select(0.0, 1.0, (sqrt(max(gsl_tmp_21, 0.000000000000000000000000000001)) > 0.0)))), dot(p, vec3<f32>(0.0, 0.0, 1.0)) / (length(p) + 1.0 - select(0.0, 1.0, (length(p) > 0.0))), s < r), min(s, r) < t)), vec3<f32>(select(select(gsl_tmp_37, gsl_tmp_38, s < r), select(select(select(select(select(gsl_tmp_39.y, gsl_tmp_39.x, gsl_tmp_26.x > gsl_tmp_26.y), select(select(gsl_tmp_39.y, gsl_tmp_39.x, gsl_tmp_26.x < gsl_tmp_26.y), gsl_tmp_39.z, gsl_tmp_26.z > min(gsl_tmp_26.x, gsl_tmp_26.y)), gsl_tmp_27), gsl_tmp_38, s < gsl_tmp_28), gsl_tmp_37, s < r), dot(gsl_tmp_29, vec2<f32>(dot(p.xy, vec2<f32>(1.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_29) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_29) > 0.0))), t < select(min(s, gsl_tmp_28), r, s < r)), min(s, r) < t), select(select(gsl_tmp_51, gsl_tmp_52, s < r), select(select(select(select(select(gsl_tmp_53.y, gsl_tmp_53.x, gsl_tmp_40.x > gsl_tmp_40.y), select(select(gsl_tmp_53.y, gsl_tmp_53.x, gsl_tmp_40.x < gsl_tmp_40.y), gsl_tmp_53.z, gsl_tmp_40.z > min(gsl_tmp_40.x, gsl_tmp_40.y)), gsl_tmp_41), gsl_tmp_52, s < gsl_tmp_42), gsl_tmp_51, s < r), dot(gsl_tmp_43, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 1.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 0.0)) / (length(gsl_tmp_43) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_43) > 0.0))), t < select(min(s, gsl_tmp_42), r, s < r)), min(s, r) < t), select(select(gsl_tmp_65, gsl_tmp_66, s < r), select(select(select(select(select(gsl_tmp_67.y, gsl_tmp_67.x, gsl_tmp_54.x > gsl_tmp_54.y), select(select(gsl_tmp_67.y, gsl_tmp_67.x, gsl_tmp_54.x < gsl_tmp_54.y), gsl_tmp_67.z, gsl_tmp_54.z > min(gsl_tmp_54.x, gsl_tmp_54.y)), gsl_tmp_55), gsl_tmp_66, s < gsl_tmp_56), gsl_tmp_65, s < r), dot(gsl_tmp_57, vec2<f32>(dot(p.xy, vec2<f32>(0.0, 0.0)) / (length(p.xy) + 1.0 - select(0.0, 1.0, (length(p.xy) > 0.0))), 1.0)) / (length(gsl_tmp_57) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_57) > 0.0))), t < select(min(s, gsl_tmp_56), r, s < r)), min(s, r) < t)))), 0.5);
    let color_darkGray = vec3<f32>(0.25, 0.25, 0.25);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    return vec4<f32>(builtin_ramp(edges, 0.005 * scale, 0.008 * scale, color_darkGray, color_white), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
