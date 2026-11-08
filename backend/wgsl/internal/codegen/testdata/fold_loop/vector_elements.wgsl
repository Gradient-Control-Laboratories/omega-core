fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn corners_gsl_anon(q: vec3<f32>, acc: vec3<f32>, o: vec3<f32>) -> vec3<f32> {
  return acc + (q - o) * length(q - o);
}

fn corners(q: vec3<f32>) -> vec3<f32> {
  var gsl_fold = vec3<f32>(0.0, 0.0, 0.0);
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = corners_gsl_anon(q, gsl_fold, gsl_foldK[gsl_i]);
  }
  return gsl_fold;
}

fn nearest_gsl_anon(q: vec3<f32>, best: f32, o: vec3<f32>) -> f32 {
  return min(best, length(q - o));
}

fn nearest(q: vec3<f32>) -> f32 {
  var gsl_fold = 1000.0;
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = nearest_gsl_anon(q, gsl_fold, gsl_foldK[gsl_i]);
  }
  return gsl_fold;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(builtin_sphere(p, p, 10.0 + corners(p / 10.0).x + nearest(p / 10.0)), 1u);
  return gsl_s1;
}

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));
