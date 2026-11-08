fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_reliefRoughness(roughness: f32, slope: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sqrt(roughness * roughness + (1.0 - texture_fadeDetail(size, noiseTex) * gsl_dispDrawn) * slope * slope);
}

fn material_withRoughness(m: mat4x4<f32>, rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(rough, surface.y, surface.z, surface.w), m[1], m[2], m[3]);
}

fn material_withLay(m: mat4x4<f32>, direction: vec3<f32>, anisotropy: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], m[1], m[2], vec4<f32>(direction.x, direction.y, direction.z, anisotropy));
}

fn material_displacement(m: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let volume = m[2];
  return volume.z;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn material_addDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_withDisplacement(m, material_displacement(m, noiseTex) + height, noiseTex);
}

fn builtin_intersectSharp(fields: array<f32, 6>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(max(max(max(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]);
}

fn builtin_firstOf(list: f32, noiseTex: texture_3d<f32>) -> f32 {
  return list;
}

fn builtin_intersectChamferEdges(fields: array<f32, 6>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  let face = builtin_intersectSharp(fields, noiseTex);
  return builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(select(face, (fields[0] + fields[0] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[0] + fields[1] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[0] + fields[2] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[0] + fields[3] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[0] + fields[4] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[0] + fields[5] + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[1] + fields[0] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[1] + fields[1] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[1] + fields[2] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[1] + fields[3] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[1] + fields[4] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[1] + fields[5] + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[2] + fields[0] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[2] + fields[1] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[2] + fields[2] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[2] + fields[3] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[2] + fields[4] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[2] + fields[5] + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[3] + fields[0] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[3] + fields[1] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[3] + fields[2] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[3] + fields[3] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[3] + fields[4] + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields[3] + fields[5] + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[4] + fields[0] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[4] + fields[1] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[4] + fields[2] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[4] + fields[3] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[4] + fields[4] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[4] + fields[5] + setback) / 1.4142135, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[5] + fields[0] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[5] + fields[1] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[5] + fields[2] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[5] + fields[3] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[5] + fields[4] + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields[5] + fields[5] + setback) / 1.4142135, 0.0 > 0.0)), noiseTex)), noiseTex);
}

fn builtin_lastOf(list: f32, noiseTex: texture_3d<f32>) -> f32 {
  return list;
}

fn builtin_intersectChamferVertices(fields: array<f32, 6>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  let face = builtin_intersectSharp(fields, noiseTex);
  return builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(select(face, (fields[0] + fields[0] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[0] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[0] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[0] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[0] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[0] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[0] + fields[1] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[1] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[1] + fields[2] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[0] + fields[1] + fields[3] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[0] + fields[1] + fields[4] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[0] + fields[1] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[0] + fields[2] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[2] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[2] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[2] + fields[3] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[0] + fields[2] + fields[4] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[0] + fields[2] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[0] + fields[3] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[3] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[3] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[3] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[3] + fields[4] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[0] + fields[3] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[0] + fields[4] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[4] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[4] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[4] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[4] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[4] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[0] + fields[5] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[5] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[5] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[5] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[5] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[0] + fields[5] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(select(face, (fields[1] + fields[0] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[0] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[0] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[0] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[0] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[0] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[1] + fields[1] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[1] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[1] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[1] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[1] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[1] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[1] + fields[2] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[2] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[2] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[2] + fields[3] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[1] + fields[2] + fields[4] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[1] + fields[2] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[1] + fields[3] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[3] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[3] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[3] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[3] + fields[4] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[1] + fields[3] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[1] + fields[4] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[4] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[4] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[4] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[4] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[4] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[1] + fields[5] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[5] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[5] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[5] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[5] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[1] + fields[5] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(select(face, (fields[2] + fields[0] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[0] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[0] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[0] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[0] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[0] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[2] + fields[1] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[1] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[1] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[1] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[1] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[1] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[2] + fields[2] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[2] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[2] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[2] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[2] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[2] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[2] + fields[3] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[3] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[3] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[3] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[3] + fields[4] + setback) / 1.7320508, 1.0 > 0.0), select(face, (fields[2] + fields[3] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[2] + fields[4] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[4] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[4] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[4] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[4] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[4] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[2] + fields[5] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[5] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[5] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[5] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[5] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[2] + fields[5] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(select(face, (fields[3] + fields[0] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[0] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[0] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[0] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[0] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[0] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[3] + fields[1] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[1] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[1] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[1] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[1] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[1] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[3] + fields[2] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[2] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[2] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[2] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[2] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[2] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[3] + fields[3] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[3] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[3] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[3] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[3] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[3] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[3] + fields[4] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[4] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[4] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[4] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[4] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[4] + fields[5] + setback) / 1.7320508, 1.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[3] + fields[5] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[5] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[5] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[5] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[5] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[3] + fields[5] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(select(face, (fields[4] + fields[0] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[0] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[0] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[0] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[0] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[0] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[4] + fields[1] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[1] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[1] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[1] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[1] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[1] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[4] + fields[2] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[2] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[2] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[2] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[2] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[2] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[4] + fields[3] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[3] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[3] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[3] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[3] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[3] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[4] + fields[4] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[4] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[4] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[4] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[4] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[4] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[4] + fields[5] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[5] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[5] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[5] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[5] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[4] + fields[5] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex), builtin_intersectSharp(array<f32, 6>(builtin_intersectSharp(array<f32, 6>(select(face, (fields[5] + fields[0] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[0] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[0] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[0] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[0] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[0] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[5] + fields[1] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[1] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[1] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[1] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[1] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[1] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[5] + fields[2] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[2] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[2] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[2] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[2] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[2] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[5] + fields[3] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[3] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[3] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[3] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[3] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[3] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[5] + fields[4] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[4] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[4] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[4] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[4] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[4] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex), builtin_intersectSharp(array<f32, 6>(select(face, (fields[5] + fields[5] + fields[0] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[5] + fields[1] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[5] + fields[2] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[5] + fields[3] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[5] + fields[4] + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields[5] + fields[5] + fields[5] + setback) / 1.7320508, 0.0 > 0.0)), noiseTex)), noiseTex)), noiseTex);
}

fn builtin_intersectChamfer(fields: array<f32, 6>, setback: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(builtin_intersectSharp(fields, noiseTex), builtin_intersectChamferEdges(fields, builtin_firstOf(setback, noiseTex), noiseTex)), builtin_intersectChamferVertices(fields, builtin_lastOf(setback, noiseTex), noiseTex));
}

fn noise_cellularF1F2CellLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  var gsl_result : gsl_struct;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  gsl_result.v_3 = gsl_cellularScan(q, 1, noiseTex).nearCell;
  return gsl_result;
}

fn noise_cellularCell(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let gsl_result = noise_cellularF1F2CellLattice(p / size, noiseTex);
  let f1 = gsl_result.v_1;
  let f2 = gsl_result.v_2;
  let cell = gsl_result.v_3;
  var gsl_result_2 : gsl_struct;
  gsl_result_2.v_1 = f1 * size;
  gsl_result_2.v_2 = f2 * size;
  gsl_result_2.v_3 = cell;
  return gsl_result_2;
}

fn builtin_sphere(p: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - radius;
}

fn gsl_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(0.77, 0.78, 0.78);
}

fn gsl_remap_2(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(0.9621, 0.9654, 0.9658);
}

fn gsl_remap_3(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let slabZ = max(p.z - 8.0, -8.0 - p.z);
  return slabZ;
}

fn gsl_remap_4(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let slabX = max(p.x - 0.0, -40.0 - p.x);
  return slabX;
}

fn gsl_remap_5(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let slabY = max(p.y - -5.0, -30.0 - p.y);
  return slabY;
}

fn gsl_remap_6(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let slabX = max(p.x - 0.0, -40.0 - p.x);
  let slabY = max(p.y - -5.0, -30.0 - p.y);
  let slabZ = max(p.z - 8.0, -8.0 - p.z);
  let corners = (slabX + slabY + slabZ + 0.5) / 1.7320508;
  return corners;
}

fn gsl_remap_7(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let slabY = max(p.y - -5.0, -30.0 - p.y);
  let slabZ = max(p.z - 8.0, -8.0 - p.z);
  let edgesX = (slabY + slabZ + 0.5) / 1.4142135;
  return edgesX;
}

fn gsl_remap_8(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let slabX = max(p.x - 0.0, -40.0 - p.x);
  let slabZ = max(p.z - 8.0, -8.0 - p.z);
  let edgesY = (slabX + slabZ + 0.5) / 1.4142135;
  return edgesY;
}

fn gsl_remap_9(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let slabX = max(p.x - 0.0, -40.0 - p.x);
  let slabY = max(p.y - -5.0, -30.0 - p.y);
  let edgesZ = (slabX + slabY + 0.5) / 1.4142135;
  return edgesZ;
}

fn gsl_remap_10(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let barChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.05), 0.05, 0.4, noiseTex);
  return barChamferMat[0];
}

fn gsl_remap_11(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let endMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(length(p - vec3<f32>(-40.0, 25.0, 0.0) - normalize(vec3<f32>(1.0, 0.0, 0.0)) * dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0)))) / 0.2), 0.2, 0.8, noiseTex);
  return endMat[0];
}

fn gsl_remap_12(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let sideMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.2), 0.2, 0.8, noiseTex);
  return sideMat[0];
}

fn gsl_remap_13(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let barChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.05), 0.05, 0.4, noiseTex);
  return barChamferMat[1];
}

fn gsl_remap_14(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let endMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(length(p - vec3<f32>(-40.0, 25.0, 0.0) - normalize(vec3<f32>(1.0, 0.0, 0.0)) * dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0)))) / 0.2), 0.2, 0.8, noiseTex);
  return endMat[1];
}

fn gsl_remap_15(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let sideMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.2), 0.2, 0.8, noiseTex);
  return sideMat[1];
}

fn gsl_remap_16(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let barChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.05), 0.05, 0.4, noiseTex);
  return barChamferMat[2];
}

fn gsl_remap_17(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let endMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(length(p - vec3<f32>(-40.0, 25.0, 0.0) - normalize(vec3<f32>(1.0, 0.0, 0.0)) * dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0)))) / 0.2), 0.2, 0.8, noiseTex);
  return endMat[2];
}

fn gsl_remap_18(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let sideMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.2), 0.2, 0.8, noiseTex);
  return sideMat[2];
}

fn gsl_remap_19(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let barChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.05), 0.05, 0.4, noiseTex);
  return barChamferMat[3];
}

fn gsl_remap_20(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let endMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(length(p - vec3<f32>(-40.0, 25.0, 0.0) - normalize(vec3<f32>(1.0, 0.0, 0.0)) * dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0)))) / 0.2), 0.2, 0.8, noiseTex);
  return endMat[3];
}

fn gsl_remap_21(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let sideMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(1.0, 0.0, 0.0)), p - vec3<f32>(-40.0, 25.0, 0.0)), fract(dot(p - vec3<f32>(-40.0, 25.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0))) / 0.2), 0.2, 0.8, noiseTex);
  return sideMat[3];
}

fn gsl_remap_22(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.254573, 1.0, 1.5, 0.0);
}

fn gsl_remap_23(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_1 = p - vec3<f32>(-20.0, -17.5, 8.0);
  let gsl_tmp_2 = clamp(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)), -24.975, 24.975);
  let gsl_tmp_3 = sqrt(625.0 - gsl_tmp_2 * gsl_tmp_2);
  let blockMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(0.0, 1.0, 0.0) * gsl_tmp_3 - vec3<f32>(1.0, 0.0, 0.0) * gsl_tmp_2, fract((dot(gsl_tmp_1, vec3<f32>(1.0, 0.0, 0.0)) - gsl_tmp_3) / 0.3), 0.3, 0.8, noiseTex);
  return blockMat[0];
}

fn gsl_remap_24(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}

fn gsl_remap_25(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_1 = p - vec3<f32>(-20.0, -17.5, 8.0);
  let gsl_tmp_2 = clamp(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)), -24.975, 24.975);
  let gsl_tmp_3 = sqrt(625.0 - gsl_tmp_2 * gsl_tmp_2);
  let blockMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(0.0, 1.0, 0.0) * gsl_tmp_3 - vec3<f32>(1.0, 0.0, 0.0) * gsl_tmp_2, fract((dot(gsl_tmp_1, vec3<f32>(1.0, 0.0, 0.0)) - gsl_tmp_3) / 0.3), 0.3, 0.8, noiseTex);
  return blockMat[1];
}

fn gsl_remap_26(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_1 = p - vec3<f32>(-20.0, -17.5, 8.0);
  let gsl_tmp_2 = clamp(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)), -24.975, 24.975);
  let gsl_tmp_3 = sqrt(625.0 - gsl_tmp_2 * gsl_tmp_2);
  let blockMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(0.0, 1.0, 0.0) * gsl_tmp_3 - vec3<f32>(1.0, 0.0, 0.0) * gsl_tmp_2, fract((dot(gsl_tmp_1, vec3<f32>(1.0, 0.0, 0.0)) - gsl_tmp_3) / 0.3), 0.3, 0.8, noiseTex);
  return blockMat[2];
}

fn gsl_remap_27(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.0, 1.0, 0.0, 1.0);
}

fn gsl_remap_28(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(1.0, 0.0, 0.0, 1.0);
}

fn gsl_remap_29(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_1 = p - vec3<f32>(-20.0, -17.5, 8.0);
  let gsl_tmp_2 = clamp(dot(gsl_tmp_1, vec3<f32>(0.0, 1.0, 0.0)), -24.975, 24.975);
  let gsl_tmp_3 = sqrt(625.0 - gsl_tmp_2 * gsl_tmp_2);
  let blockMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(0.0, 1.0, 0.0) * gsl_tmp_3 - vec3<f32>(1.0, 0.0, 0.0) * gsl_tmp_2, fract((dot(gsl_tmp_1, vec3<f32>(1.0, 0.0, 0.0)) - gsl_tmp_3) / 0.3), 0.3, 0.8, noiseTex);
  return blockMat[3];
}

fn gsl_remap_30(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.0, 0.0, 1.0, 1.0);
}

fn gsl_remap_31(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05), 0.05, 0.4, noiseTex);
  return knobChamferMat[0];
}

fn gsl_remap_32(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobEndMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(length(p - vec3<f32>(0.0, 50.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
  return knobEndMat[0];
}

fn gsl_remap_33(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobTurnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.15), 0.15, 0.8, noiseTex);
  let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
  return material_blend(knobTurnedMat, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), runOut, noiseTex)[0];
}

fn gsl_remap_34(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05), 0.05, 0.4, noiseTex);
  return knobChamferMat[1];
}

fn gsl_remap_35(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobEndMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(length(p - vec3<f32>(0.0, 50.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
  return knobEndMat[1];
}

fn gsl_remap_36(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobTurnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.15), 0.15, 0.8, noiseTex);
  let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
  return material_blend(knobTurnedMat, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), runOut, noiseTex)[1];
}

fn gsl_remap_37(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05), 0.05, 0.4, noiseTex);
  return knobChamferMat[2];
}

fn gsl_remap_38(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobEndMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(length(p - vec3<f32>(0.0, 50.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
  return knobEndMat[2];
}

fn gsl_remap_39(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobTurnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.15), 0.15, 0.8, noiseTex);
  let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
  return material_blend(knobTurnedMat, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), runOut, noiseTex)[2];
}

fn gsl_remap_40(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobChamferMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05), 0.05, 0.4, noiseTex);
  return knobChamferMat[3];
}

fn gsl_remap_41(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobEndMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(length(p - vec3<f32>(0.0, 50.0, 0.0) - normalize(vec3<f32>(0.0, 0.0, 1.0)) * dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0)))) / 0.15), 0.15, 0.8, noiseTex);
  return knobEndMat[3];
}

fn gsl_remap_42(p: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let knobTurnedMat = machined_grooveFinish(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), cross(normalize(vec3<f32>(0.0, 0.0, 1.0)), p - vec3<f32>(0.0, 50.0, 0.0)), fract(dot(p - vec3<f32>(0.0, 50.0, 0.0), normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.15), 0.15, 0.8, noiseTex);
  let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
  return material_blend(knobTurnedMat, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), runOut, noiseTex)[3];
}

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
}

fn texture_grainShade(p: vec3<f32>, size: f32, spread: f32, noiseTex: texture_3d<f32>) -> f32 {
  var gsl_detailValue = 1.0;
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (size)) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let gsl_result_2 = noise_cellularCell(p, size, noiseTex);
    let cell = gsl_result_2.v_3;
    let pick = fract(textureLoad(noiseTex, vec3<i32>(floor(cell + vec3<f32>(47.0, 13.0, 29.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell + vec3<f32>(47.0, 13.0, 29.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    gsl_detailValue = texture_fadeTo(1.0 + spread * (2.0 * pick.x - 1.0), 1.0, size, noiseTex);
    gsl_pw = gsl_detailOuterPw;
  }
  let shade = gsl_detailValue;
  return shade;
}

fn machined_layFinish(baseMaterial: mat4x4<f32>, direction: vec3<f32>, along: f32, across: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let alpha = sqrt(along * across);
  let squeeze = clamp(along / across, 0.1, 1.0);
  let stretch = (1.0 - squeeze) / 0.9;
  let unit = select(vec3<f32>(0.0, 0.0, 0.0), direction / max(length(direction), 0.000001), stretch > 0.0);
  return material_withLay(material_withRoughness(baseMaterial, sqrt(alpha), noiseTex), unit, stretch, noiseTex);
}

fn machined_nosePitch(s: f32, spacing: f32, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let x = (s - 0.5) * spacing;
  return sqrt(max(radius * radius - x * x, 0.0)) - sqrt(max(radius * radius - 0.25 * spacing * spacing, 0.0));
}

fn machined_grooveFinish(baseMaterial: mat4x4<f32>, direction: vec3<f32>, s: f32, spacing: f32, radius: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let slope = 0.408 * spacing / radius;
  var gsl_simpleValue = material_addDisplacement(machined_layFinish(baseMaterial, direction, 0.02, sqrt(0.0004 + slope * slope), noiseTex), -spacing * spacing / (12.0 * radius), noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let across = texture_reliefRoughness(0.02, slope, spacing, noiseTex);
    let fade = texture_fadeDetail(spacing, noiseTex);
    let gsl_simpleFull = material_addDisplacement(machined_layFinish(baseMaterial, direction, 0.02, across, noiseTex), -machined_nosePitch(s, spacing, radius, noiseTex) * fade, noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let finish = gsl_simpleValue;
  return finish;
}

fn machined_knurlTooth(u: f32, noiseTex: texture_3d<f32>) -> f32 {
  return -abs(fract(u) - 0.5);
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 7> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let barSide = builtin_axis(p, vec3<f32>(-40.0, 25.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), noiseTex) - 10.0;
  let barEnds = max(p.x - 5.0, -p.x - 85.0);
  let barChamfer = (barSide + barEnds + 0.5) / 1.4142135;
  let bar = max(max(barSide, barEnds), barChamfer);
  let gsl_partLo = vec3<f32>(-40.0, -30.0, -8.0);
  let gsl_partHi = vec3<f32>(0.0, -5.0, 8.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let slabX = max(p.x - 0.0, -40.0 - p.x);
  let slabY = max(p.y - -5.0, -30.0 - p.y);
  let slabZ = max(p.z - 8.0, -8.0 - p.z);
  let edgesX = (slabY + slabZ + 0.5) / 1.4142135;
  let edgesY = (slabX + slabZ + 0.5) / 1.4142135;
  let edgesZ = (slabX + slabY + 0.5) / 1.4142135;
  let corners = (slabX + slabY + slabZ + 0.5) / 1.7320508;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = max(max(max(max(max(max(slabX, slabY), slabZ), edgesX), edgesY), edgesZ), corners) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let block = gsl_part;
  let gsl_partLo_2 = vec3<f32>(10.0, -30.0, -3.0);
  let gsl_partHi_2 = vec3<f32>(30.0, -5.0, 3.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_intersectChamfer(array<f32, 6>(p.x - 30.0, 10.0 - p.x, p.y - -5.0, -30.0 - p.y, p.z - 3.0, -3.0 - p.z), 0.3, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let groundPlate = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(35.0, -30.0, -3.0);
  let gsl_partHi_3 = vec3<f32>(55.0, -5.0, 3.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_intersectChamfer(array<f32, 6>(p.x - 55.0, 35.0 - p.x, p.y - -5.0, -30.0 - p.y, p.z - 3.0, -3.0 - p.z), 0.3, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let brushedPlate = gsl_part_3;
  let beads = builtin_sphere(p - vec3<f32>(20.0, 25.0, 0.0), 8.0, noiseTex);
  let sand = builtin_sphere(p - vec3<f32>(45.0, 25.0, 0.0), 8.0, noiseTex);
  let knobBlank = builtin_axis(p, vec3<f32>(0.0, 50.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex) - 8.0;
  let gsl_tmp_4 = p - vec3<f32>(0.0, 50.0, 0.0);
  let gsl_tmp_5 = gsl_tmp_4 - vec3<f32>(0.0, 0.0, 1.0) * dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0));
  let gsl_tmp_6 = (atan2(dot(gsl_tmp_5, vec3<f32>(-0.70710737, -0.7071062, 0.0)), dot(gsl_tmp_5, vec3<f32>(-0.7071062, 0.70710737, 0.0))) / 6.2831855 + 0.5) * 50.0;
  let gsl_tmp_7 = dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0)) / 1.0;
  let gsl_tmp_8 = machined_knurlTooth(gsl_tmp_6 + 0.57735 * gsl_tmp_7, noiseTex);
  let gsl_tmp_9 = machined_knurlTooth(gsl_tmp_6 - 0.57735 * gsl_tmp_7, noiseTex);
  let knurlHeight = min(gsl_tmp_8, gsl_tmp_9) * 1.0 + 0.25 + 0.0;
  let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
  let knobSide = (knobBlank - knurlHeight * runOut) / 3.2;
  let knobEnds = abs(p.z) - 12.0;
  let knobChamfer = (knobBlank + knobEnds + 0.5) / 1.4142135;
  let knob = max(max(knobSide, knobEnds), knobChamfer);
  return array<f32, 7>(bar, block, groundPlate, brushedPlate, beads, sand, knob);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let barSide = builtin_axis(p, vec3<f32>(-40.0, 25.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), noiseTex) - 10.0;
    let barEnds = max(p.x - 5.0, -p.x - 85.0);
    let barChamfer = (barSide + barEnds + 0.5) / 1.4142135;
    return vec4<f32>(clamp(gsl_remap(p, noiseTex) * select(1.0, select(0.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap(p, noiseTex) * select(0.0, select(1.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap(p, noiseTex) * select(0.0, select(0.0, 1.0, barSide > barEnds), max(barSide, barEnds) > barChamfer), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  if (gsl_colorIdx == 2u) {
    let slabX = max(p.x - 0.0, -40.0 - p.x);
    let slabY = max(p.y - -5.0, -30.0 - p.y);
    let slabZ = max(p.z - 8.0, -8.0 - p.z);
    let edgesY = (slabX + slabZ + 0.5) / 1.4142135;
    let edgesZ = (slabX + slabY + 0.5) / 1.4142135;
    let corners = (slabX + slabY + slabZ + 0.5) / 1.7320508;
    let edgesX = (slabY + slabZ + 0.5) / 1.4142135;
    return vec4<f32>(clamp((clamp((gsl_remap_2(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_2(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_2(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_2(p, noiseTex), gsl_remap_2(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_2(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)) * select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_2(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_2(p, noiseTex) * select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_2(p, noiseTex) * select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_2(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_2(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_2(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(select(select(select(select(clamp((gsl_remap_2(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_2(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_2(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_2(p, noiseTex), gsl_remap_2(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_2(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_remap_2(p, noiseTex), gsl_remap_7(p, noiseTex) < gsl_remap_6(p, noiseTex)), gsl_remap_2(p, noiseTex), gsl_remap_8(p, noiseTex) < min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex))), gsl_remap_2(p, noiseTex), gsl_remap_9(p, noiseTex) < min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex))), gsl_remap_2(p, noiseTex), gsl_remap_4(p, noiseTex) < min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex))), gsl_remap_2(p, noiseTex), gsl_remap_5(p, noiseTex) < min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex))), gsl_remap_2(p, noiseTex), gsl_remap_3(p, noiseTex) < min(min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex)), gsl_remap_5(p, noiseTex))) * max(1.0 - (select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners))), 0.0)) / max(select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)), 1.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.77, 0.78, 0.78), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 7u) {
    let knobBlank = builtin_axis(p, vec3<f32>(0.0, 50.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex) - 8.0;
    let gsl_tmp_4 = p - vec3<f32>(0.0, 50.0, 0.0);
    let gsl_tmp_5 = gsl_tmp_4 - vec3<f32>(0.0, 0.0, 1.0) * dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0));
    let gsl_tmp_6 = (atan2(dot(gsl_tmp_5, vec3<f32>(-0.70710737, -0.7071062, 0.0)), dot(gsl_tmp_5, vec3<f32>(-0.7071062, 0.70710737, 0.0))) / 6.2831855 + 0.5) * 50.0;
    let gsl_tmp_7 = dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0)) / 1.0;
    let gsl_tmp_8 = machined_knurlTooth(gsl_tmp_6 + 0.57735 * gsl_tmp_7, noiseTex);
    let gsl_tmp_9 = machined_knurlTooth(gsl_tmp_6 - 0.57735 * gsl_tmp_7, noiseTex);
    let knurlHeight = min(gsl_tmp_8, gsl_tmp_9) * 1.0 + 0.25 + 0.0;
    let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
    let knobSide = (knobBlank - knurlHeight * runOut) / 3.2;
    let knobEnds = abs(p.z) - 12.0;
    let knobChamfer = (knobBlank + knobEnds + 0.5) / 1.4142135;
    return vec4<f32>(clamp(gsl_remap(p, noiseTex) * select(1.0, select(0.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap(p, noiseTex) * select(0.0, select(1.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap(p, noiseTex) * select(0.0, select(0.0, 1.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let barSide = builtin_axis(p, vec3<f32>(-40.0, 25.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), noiseTex) - 10.0;
    let barEnds = max(p.x - 5.0, -p.x - 85.0);
    let barChamfer = (barSide + barEnds + 0.5) / 1.4142135;
    let gsl_m = mat4x4<f32>(clamp(gsl_remap_10(p, noiseTex) * select(1.0, select(0.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_11(p, noiseTex) * select(0.0, select(1.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_12(p, noiseTex) * select(0.0, select(0.0, 1.0, barSide > barEnds), max(barSide, barEnds) > barChamfer), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_13(p, noiseTex) * select(1.0, select(0.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_14(p, noiseTex) * select(0.0, select(1.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_15(p, noiseTex) * select(0.0, select(0.0, 1.0, barSide > barEnds), max(barSide, barEnds) > barChamfer), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_16(p, noiseTex) * select(1.0, select(0.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_17(p, noiseTex) * select(0.0, select(1.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_18(p, noiseTex) * select(0.0, select(0.0, 1.0, barSide > barEnds), max(barSide, barEnds) > barChamfer), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), clamp(gsl_remap_19(p, noiseTex) * select(1.0, select(0.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_20(p, noiseTex) * select(0.0, select(1.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_21(p, noiseTex) * select(0.0, select(0.0, 1.0, barSide > barEnds), max(barSide, barEnds) > barChamfer), vec4<f32>(-10000.0, -10000.0, -10000.0, 0.0), vec4<f32>(10000.0, 10000.0, 10000.0, 1.0))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let slabX = max(p.x - 0.0, -40.0 - p.x);
    let slabY = max(p.y - -5.0, -30.0 - p.y);
    let slabZ = max(p.z - 8.0, -8.0 - p.z);
    let edgesY = (slabX + slabZ + 0.5) / 1.4142135;
    let edgesZ = (slabX + slabY + 0.5) / 1.4142135;
    let corners = (slabX + slabY + slabZ + 0.5) / 1.7320508;
    let edgesX = (slabY + slabZ + 0.5) / 1.4142135;
    let gsl_m = mat4x4<f32>(clamp((clamp((gsl_remap_22(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_22(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_23(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_22(p, noiseTex), gsl_remap_22(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_23(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)) * select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_22(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_22(p, noiseTex) * select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_22(p, noiseTex) * select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_22(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_22(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_23(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(select(select(select(select(clamp((gsl_remap_22(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_22(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_23(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_22(p, noiseTex), gsl_remap_22(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_23(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), gsl_remap_22(p, noiseTex), gsl_remap_7(p, noiseTex) < gsl_remap_6(p, noiseTex)), gsl_remap_22(p, noiseTex), gsl_remap_8(p, noiseTex) < min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex))), gsl_remap_22(p, noiseTex), gsl_remap_9(p, noiseTex) < min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex))), gsl_remap_22(p, noiseTex), gsl_remap_4(p, noiseTex) < min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex))), gsl_remap_22(p, noiseTex), gsl_remap_5(p, noiseTex) < min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex))), gsl_remap_23(p, noiseTex), gsl_remap_3(p, noiseTex) < min(min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex)), gsl_remap_5(p, noiseTex))) * max(1.0 - (select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners))), 0.0)) / max(select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)), 1.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp((clamp((gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_25(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_24(p, noiseTex), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_25(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)) * select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_25(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(select(select(select(select(clamp((gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_25(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_24(p, noiseTex), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_25(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), gsl_remap_24(p, noiseTex), gsl_remap_7(p, noiseTex) < gsl_remap_6(p, noiseTex)), gsl_remap_24(p, noiseTex), gsl_remap_8(p, noiseTex) < min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_9(p, noiseTex) < min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_4(p, noiseTex) < min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex))), gsl_remap_25(p, noiseTex), gsl_remap_3(p, noiseTex) < min(min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex)), gsl_remap_5(p, noiseTex))) * max(1.0 - (select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners))), 0.0)) / max(select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)), 1.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp((clamp((gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_26(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_24(p, noiseTex), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_26(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)) * select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_26(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(select(select(select(select(clamp((gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_26(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_24(p, noiseTex), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_26(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), gsl_remap_24(p, noiseTex), gsl_remap_7(p, noiseTex) < gsl_remap_6(p, noiseTex)), gsl_remap_24(p, noiseTex), gsl_remap_8(p, noiseTex) < min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_9(p, noiseTex) < min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_4(p, noiseTex) < min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex))), gsl_remap_26(p, noiseTex), gsl_remap_3(p, noiseTex) < min(min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex)), gsl_remap_5(p, noiseTex))) * max(1.0 - (select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners))), 0.0)) / max(select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)), 1.0), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), clamp((clamp((gsl_remap_27(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_28(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_29(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_27(p, noiseTex), gsl_remap_28(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_29(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(-10000.0, -10000.0, -10000.0, 0.0), vec4<f32>(10000.0, 10000.0, 10000.0, 1.0)) * select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_28(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_27(p, noiseTex) * select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_30(p, noiseTex) * select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_27(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_28(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_29(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(select(select(select(select(clamp((gsl_remap_27(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_28(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_29(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_27(p, noiseTex), gsl_remap_28(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_29(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(-10000.0, -10000.0, -10000.0, 0.0), vec4<f32>(10000.0, 10000.0, 10000.0, 1.0)), gsl_remap_28(p, noiseTex), gsl_remap_7(p, noiseTex) < gsl_remap_6(p, noiseTex)), gsl_remap_27(p, noiseTex), gsl_remap_8(p, noiseTex) < min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex))), gsl_remap_30(p, noiseTex), gsl_remap_9(p, noiseTex) < min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex))), gsl_remap_27(p, noiseTex), gsl_remap_4(p, noiseTex) < min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex))), gsl_remap_28(p, noiseTex), gsl_remap_5(p, noiseTex) < min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex))), gsl_remap_29(p, noiseTex), gsl_remap_3(p, noiseTex) < min(min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex)), gsl_remap_5(p, noiseTex))) * max(1.0 - (select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners))), 0.0)) / max(select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)), 1.0), vec4<f32>(-10000.0, -10000.0, -10000.0, 0.0), vec4<f32>(10000.0, 10000.0, 10000.0, 1.0))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.21657367, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 1.0, 0.0, 0.90909094)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.30800703, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.0, 0.0, 1.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5732115, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.80178374, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) {
    let knobBlank = builtin_axis(p, vec3<f32>(0.0, 50.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex) - 8.0;
    let gsl_tmp_4 = p - vec3<f32>(0.0, 50.0, 0.0);
    let gsl_tmp_5 = gsl_tmp_4 - vec3<f32>(0.0, 0.0, 1.0) * dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0));
    let gsl_tmp_6 = (atan2(dot(gsl_tmp_5, vec3<f32>(-0.70710737, -0.7071062, 0.0)), dot(gsl_tmp_5, vec3<f32>(-0.7071062, 0.70710737, 0.0))) / 6.2831855 + 0.5) * 50.0;
    let gsl_tmp_7 = dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0)) / 1.0;
    let gsl_tmp_8 = machined_knurlTooth(gsl_tmp_6 + 0.57735 * gsl_tmp_7, noiseTex);
    let gsl_tmp_9 = machined_knurlTooth(gsl_tmp_6 - 0.57735 * gsl_tmp_7, noiseTex);
    let knurlHeight = min(gsl_tmp_8, gsl_tmp_9) * 1.0 + 0.25 + 0.0;
    let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
    let knobSide = (knobBlank - knurlHeight * runOut) / 3.2;
    let knobEnds = abs(p.z) - 12.0;
    let knobChamfer = (knobBlank + knobEnds + 0.5) / 1.4142135;
    let gsl_m = mat4x4<f32>(clamp(gsl_remap_31(p, noiseTex) * select(1.0, select(0.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_32(p, noiseTex) * select(0.0, select(1.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_33(p, noiseTex) * select(0.0, select(0.0, 1.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_34(p, noiseTex) * select(1.0, select(0.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_35(p, noiseTex) * select(0.0, select(1.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_36(p, noiseTex) * select(0.0, select(0.0, 1.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_37(p, noiseTex) * select(1.0, select(0.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_38(p, noiseTex) * select(0.0, select(1.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_39(p, noiseTex) * select(0.0, select(0.0, 1.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), clamp(gsl_remap_40(p, noiseTex) * select(1.0, select(0.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_41(p, noiseTex) * select(0.0, select(1.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_42(p, noiseTex) * select(0.0, select(0.0, 1.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer), vec4<f32>(-10000.0, -10000.0, -10000.0, 0.0), vec4<f32>(10000.0, 10000.0, 10000.0, 1.0))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
struct gsl_CellularScan {
  f1: f32,
  f2: f32,
  feature: vec3<f32>,
  farFeature: vec3<f32>,
  nearCell: vec3<f32>,
  farCell: vec3<f32>,
}

fn gsl_cellularHash(cell: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let fine = fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
  if (oct <= 1) { return fine; }
  // fract(fract(a) + b) == fract(a + b): the inner fract only removes an
  // integer, so folding the supercell on top of the already-fracted fine
  // lookup is the same value noiseHash3's two-lookup form produces.
  return fract(fine + textureLoad(noiseTex, vec3<i32>(floor(cell / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn gsl_cellularScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularScan {
  let pi = floor(q);
  let pf = q - pi;
  // Seeded with an unreachable distance, squared, because the scan compares
  // squared distances and roots the two winners at the end.
  var best = 64.000000;
  var second = 64.000000;
  var bestCell = pi;
  var bestH = vec3<f32>(0.0);
  var secondCell = pi;
  var secondH = vec3<f32>(0.0);
  for (var i = 0; i < 27; i = i + 1) {
    // The same offset order cellularOffsets() lists: x fastest, then y, then z.
    let o = vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
    let h = gsl_cellularHash(pi + o, oct, noiseTex);
    let v = pf - (o + h);
    let d2 = dot(v, v);
    if (d2 < best) {
      second = best;
      secondCell = bestCell;
      secondH = bestH;
      best = d2;
      bestCell = pi + o;
      bestH = h;
    } else if (d2 < second) {
      second = d2;
      secondCell = pi + o;
      secondH = h;
    }
  }
  var out: gsl_CellularScan;
  out.f1 = sqrt(best);
  out.f2 = sqrt(second);
  // Both closest points, so both gradients are available without a second hash.
  out.feature = bestCell + bestH;
  out.farFeature = secondCell + secondH;
  out.nearCell = bestCell;
  out.farCell = secondCell;
  return out;
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

var<private> gsl_detailPw: f32 = 0.0;

fn gsl_heights(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 7> {
  let barSide = builtin_axis(p, vec3<f32>(-40.0, 25.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), noiseTex) - 10.0;
  let barEnds = max(p.x - 5.0, -p.x - 85.0);
  let barChamfer = (barSide + barEnds + 0.5) / 1.4142135;
  let slabX = max(p.x - 0.0, -40.0 - p.x);
  let slabY = max(p.y - -5.0, -30.0 - p.y);
  let slabZ = max(p.z - 8.0, -8.0 - p.z);
  let edgesY = (slabX + slabZ + 0.5) / 1.4142135;
  let edgesZ = (slabX + slabY + 0.5) / 1.4142135;
  let corners = (slabX + slabY + slabZ + 0.5) / 1.7320508;
  let edgesX = (slabY + slabZ + 0.5) / 1.4142135;
  let knobBlank = builtin_axis(p, vec3<f32>(0.0, 50.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex) - 8.0;
  let gsl_tmp_4 = p - vec3<f32>(0.0, 50.0, 0.0);
  let gsl_tmp_5 = gsl_tmp_4 - vec3<f32>(0.0, 0.0, 1.0) * dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0));
  let gsl_tmp_6 = (atan2(dot(gsl_tmp_5, vec3<f32>(-0.70710737, -0.7071062, 0.0)), dot(gsl_tmp_5, vec3<f32>(-0.7071062, 0.70710737, 0.0))) / 6.2831855 + 0.5) * 50.0;
  let gsl_tmp_7 = dot(gsl_tmp_4, vec3<f32>(0.0, 0.0, 1.0)) / 1.0;
  let gsl_tmp_8 = machined_knurlTooth(gsl_tmp_6 + 0.57735 * gsl_tmp_7, noiseTex);
  let gsl_tmp_9 = machined_knurlTooth(gsl_tmp_6 - 0.57735 * gsl_tmp_7, noiseTex);
  let knurlHeight = min(gsl_tmp_8, gsl_tmp_9) * 1.0 + 0.25 + 0.0;
  let runOut = clamp((10.5 - abs(p.z)) / 0.5, 0.0, 1.0);
  let knobSide = (knobBlank - knurlHeight * runOut) / 3.2;
  let knobEnds = abs(p.z) - 12.0;
  let knobChamfer = (knobBlank + knobEnds + 0.5) / 1.4142135;
  return array<f32, 7>(clamp(gsl_remap_16(p, noiseTex) * select(1.0, select(0.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_17(p, noiseTex) * select(0.0, select(1.0, 0.0, barSide > barEnds), max(barSide, barEnds) > barChamfer) + gsl_remap_18(p, noiseTex) * select(0.0, select(0.0, 1.0, barSide > barEnds), max(barSide, barEnds) > barChamfer), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)).z, clamp((clamp((gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_26(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_24(p, noiseTex), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_26(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)) * select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_24(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + gsl_remap_26(p, noiseTex) * select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(select(select(select(select(clamp((gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_24(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + gsl_remap_26(p, noiseTex) * (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + select(select(gsl_remap_24(p, noiseTex), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < gsl_remap_4(p, noiseTex)), gsl_remap_26(p, noiseTex), gsl_remap_3(p, noiseTex) < min(gsl_remap_4(p, noiseTex), gsl_remap_5(p, noiseTex))) * max(1.0 - ((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0), 0.0)) / max((1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0 + (1.7320508 - (slabX + slabY + slabZ + 0.5) * 0.0) / 3.0, 1.0), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), gsl_remap_24(p, noiseTex), gsl_remap_7(p, noiseTex) < gsl_remap_6(p, noiseTex)), gsl_remap_24(p, noiseTex), gsl_remap_8(p, noiseTex) < min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_9(p, noiseTex) < min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_4(p, noiseTex) < min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex))), gsl_remap_24(p, noiseTex), gsl_remap_5(p, noiseTex) < min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex))), gsl_remap_26(p, noiseTex), gsl_remap_3(p, noiseTex) < min(min(min(min(min(gsl_remap_6(p, noiseTex), gsl_remap_7(p, noiseTex)), gsl_remap_8(p, noiseTex)), gsl_remap_9(p, noiseTex)), gsl_remap_4(p, noiseTex)), gsl_remap_5(p, noiseTex))) * max(1.0 - (select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners))), 0.0)) / max(select(select(1.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(1.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 1.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(1.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(0.0, 1.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 0.0, slabZ > edgesX), select(1.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)) + select(select(0.0, select(0.0, 0.0, edgesY > edgesZ), max(edgesY, edgesZ) > corners), select(select(0.0, 1.0, slabZ > edgesX), select(0.0, 0.0, slabX > slabY), max(slabX, slabY) > max(slabZ, edgesX)), max(max(slabX, slabY), max(slabZ, edgesX)) > max(max(edgesY, edgesZ), corners)), 1.0), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)).z, 0.0, 0.0, 0.0, 0.0, clamp(gsl_remap_37(p, noiseTex) * select(1.0, select(0.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_38(p, noiseTex) * select(0.0, select(1.0, 0.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer) + gsl_remap_39(p, noiseTex) * select(0.0, select(0.0, 1.0, knobSide > knobEnds), max(knobSide, knobEnds) > knobChamfer), vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)).z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_h = gsl_heights(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6] - gsl_h[6] * gsl_dispK, 7u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
}
