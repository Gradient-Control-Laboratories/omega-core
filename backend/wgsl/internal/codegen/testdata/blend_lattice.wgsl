fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn gsl_floored_vec3(a: vec3<f32>, n: vec3<f32>) -> vec3<f32> {
  return a - n * floor(a / n);
}

fn builtin_intersectSharp(fields: vec4<f32>) -> f32 {
  return max(max(max(fields.x, fields.y), fields.z), fields.w);
}

fn builtin_intersectRound(fields: vec4<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0) + max(fields.w + radius, 0.0) * max(fields.w + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_unionRound(fields: vec4<f32>, radius: f32) -> f32 {
  let negated = vec4<f32>(-fields.x, -fields.y, -fields.z, -fields.w);
  return -builtin_intersectRound(negated, radius);
}

fn builtin_gcl_closure_unionRound(fields: vec4<f32>) -> f32 {
  return builtin_unionRound(fields, 0.2);
}

fn builtin_unionSharp(fields: vec4<f32>) -> f32 {
  return min(min(min(fields.x, fields.y), fields.z), fields.w);
}

fn builtin_lineSegment(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn builtin_unionSharp_2(fields: vec3<f32>) -> f32 {
  return min(min(fields.x, fields.y), fields.z);
}

fn gsl_remap(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirY: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_unionSharp_2(vec3<f32>(builtin_lineSegment(p, vec3<f32>(-2.0, 0.0, 0.0), vec3<f32>(2.0, 0.0, 0.0)), builtin_lineSegment(p, vec3<f32>(0.0, -2.0, 0.0), vec3<f32>(0.0, 2.0, 0.0)), builtin_lineSegment(p, vec3<f32>(0.0, 0.0, -2.0), vec3<f32>(0.0, 0.0, 2.0))));
}

fn gsl_remap_2(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirY: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_unionSharp(vec4<f32>(builtin_lineSegment(p, vec3<f32>(2.0, 2.0, 2.0), vec3<f32>(0.0, 2.0, 2.0)), builtin_lineSegment(p, vec3<f32>(2.0, 2.0, 2.0), vec3<f32>(2.0, 0.0, 2.0)), builtin_lineSegment(p, vec3<f32>(2.0, 2.0, 2.0), vec3<f32>(2.0, 2.0, 0.0)), builtin_lineSegment(p, vec3<f32>(2.0, 2.0, 2.0), vec3<f32>(0.0, 0.0, 0.0))));
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp_2(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_unionRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let negated = vec2<f32>(-fields.x, -fields.y);
  return -builtin_intersectRound_2(negated, radius);
}

fn builtin_gcl_closure_unionRound_2(fields: vec2<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.15);
}

fn builtin_unionSharp_3(fields: array<f32, 7>) -> f32 {
  return min(min(min(min(min(min(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]), fields[6]);
}

fn gsl_remap_3(p: vec3<f32>, vector_dirX: vec3<f32>, vector_dirY: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return builtin_unionSharp_3(array<f32, 7>(builtin_lineSegment(p, vec3<f32>(0.0, 2.0, 0.8), vec3<f32>(0.0, 0.0, 1.6)), builtin_lineSegment(p, vec3<f32>(0.0, 2.0, 2.0), vec3<f32>(0.0, 2.0, 0.8)), builtin_lineSegment(p, vec3<f32>(2.0, 0.0, 0.8), vec3<f32>(0.0, 0.0, 1.6)), builtin_lineSegment(p, vec3<f32>(2.0, 0.0, 2.0), vec3<f32>(2.0, 0.0, 0.8)), builtin_lineSegment(p, vec3<f32>(2.0, 2.0, 1.6), vec3<f32>(0.0, 2.0, 0.8)), builtin_lineSegment(p, vec3<f32>(2.0, 2.0, 1.6), vec3<f32>(2.0, 0.0, 0.8)), builtin_lineSegment(p, vec3<f32>(2.0, 2.0, 1.6), vec3<f32>(2.0, 2.0, 0.0))));
}

fn builtin_gcl_closure_unionRound_3(fields: vec2<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.1);
}

fn builtin_intersectSharp_3(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound_3(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp_3(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound_3(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn map_repeatCentered(x: vec3<f32>, period: vec3<f32>) -> vec3<f32> {
  let half = period * 0.5;
  return gsl_floored_vec3((gsl_floored_vec3((x + half), period) + period), period) - half;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let block = builtin_gcl_closure_boxCentered(p, vector_zero, vec3<f32>(8.0, 8.0, 8.0));
  let gsl_tmp_1 = vec3<f32>(gsl_floored((gsl_floored((p.x + 2.0), 4.0) + 4.0), 4.0) - 2.0, gsl_floored((gsl_floored((p.y + 2.0), 4.0) + 4.0), 4.0) - 2.0, gsl_floored((gsl_floored((p.z + 2.0), 4.0) + 4.0), 4.0) - 2.0);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let vector_dirY = vec3<f32>(0.0, 1.0, 0.0);
  let filleted = builtin_gcl_closure_unionRound(vec4<f32>(length(gsl_tmp_1 - normalize(vec3<f32>(1.0, 1.0, 0.0) + vector_dirZ) * clamp(dot(gsl_tmp_1, normalize(vec3<f32>(1.0, 1.0, 0.0) + vector_dirZ)), -3.4641016, 3.4641016)) - 0.3, length(gsl_tmp_1 - normalize(vec3<f32>(-1.0, -0.0, -0.0) + vector_dirY + vector_dirZ) * clamp(dot(gsl_tmp_1, normalize(vec3<f32>(-1.0, -0.0, -0.0) + vector_dirY + vector_dirZ)), -3.4641016, 3.4641016)) - 0.3, length(gsl_tmp_1 - normalize(vec3<f32>(1.0, -1.0, 0.0) + vector_dirZ) * clamp(dot(gsl_tmp_1, normalize(vec3<f32>(1.0, -1.0, 0.0) + vector_dirZ)), -3.4641016, 3.4641016)) - 0.3, length(gsl_tmp_1 - normalize(vec3<f32>(1.0, 1.0, 0.0) - vector_dirZ) * clamp(dot(gsl_tmp_1, normalize(vec3<f32>(1.0, 1.0, 0.0) - vector_dirZ)), -3.4641016, 3.4641016)) - 0.3));
  let gsl_tmp_2 = vec3<f32>(gsl_floored((gsl_floored((p.x + 2.0), 4.0) + 4.0), 4.0) - 2.0, gsl_floored((gsl_floored((p.y + 2.0), 4.0) + 4.0), 4.0) - 2.0, gsl_floored((gsl_floored((p.z + 2.0), 4.0) + 4.0), 4.0) - 2.0);
  let sharp = builtin_unionSharp(vec4<f32>(length(gsl_tmp_2 - normalize(vec3<f32>(1.0, 1.0, 0.0) + vector_dirZ) * clamp(dot(gsl_tmp_2, normalize(vec3<f32>(1.0, 1.0, 0.0) + vector_dirZ)), -3.4641016, 3.4641016)) - 0.3, length(gsl_tmp_2 - normalize(vec3<f32>(-1.0, -0.0, -0.0) + vector_dirY + vector_dirZ) * clamp(dot(gsl_tmp_2, normalize(vec3<f32>(-1.0, -0.0, -0.0) + vector_dirY + vector_dirZ)), -3.4641016, 3.4641016)) - 0.3, length(gsl_tmp_2 - normalize(vec3<f32>(1.0, -1.0, 0.0) + vector_dirZ) * clamp(dot(gsl_tmp_2, normalize(vec3<f32>(1.0, -1.0, 0.0) + vector_dirZ)), -3.4641016, 3.4641016)) - 0.3, length(gsl_tmp_2 - normalize(vec3<f32>(1.0, 1.0, 0.0) - vector_dirZ) * clamp(dot(gsl_tmp_2, normalize(vec3<f32>(1.0, 1.0, 0.0) - vector_dirZ)), -3.4641016, 3.4641016)) - 0.3));
  let vector_dirX = vec3<f32>(1.0, 0.0, 0.0);
  let truss = builtin_gcl_closure_unionRound_2(vec2<f32>(gsl_remap(map_repeatCentered(p, vec3<f32>(4.0, 4.0, 4.0)), vector_dirX, vector_dirY, vector_dirZ), gsl_remap_2(abs(map_repeatCentered(p, vec3<f32>(4.0, 4.0, 4.0))), vector_dirX, vector_dirY, vector_dirZ)));
  let folded = builtin_gcl_closure_unionRound_3(vec2<f32>(gsl_remap_3(abs(map_repeatCentered(p, vec3<f32>(4.0, 4.0, 4.0))), vector_dirX, vector_dirY, vector_dirZ), builtin_lineSegment(map_repeatCentered(p, vec3<f32>(4.0, 4.0, 4.0)), vec3<f32>(0.0, 0.0, -1.6), vec3<f32>(0.0, 0.0, 1.6))));
  let gsl_s1 = gsl_shape(builtin_intersectSharp_2(vec2<f32>(block, builtin_unionSharp(vec4<f32>(filleted, sharp, truss, folded)))), 1u);
  return gsl_s1;
}
