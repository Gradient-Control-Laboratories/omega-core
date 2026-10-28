fn sphere_dist1(pt: vec3<f32>) -> f32 {
  return sqrt(pt.x * pt.x + pt.y * pt.y + pt.z * pt.z);
}

fn sphere_dist2(localOffset: f32, pt: vec3<f32>) -> f32 {
  return sphere_dist1(pt) + localOffset;
}

fn sphere_dist3(radius: f32, pt: vec3<f32>) -> f32 {
  return sphere_dist1(pt) - radius;
}

fn sphere(p: vec3<f32>, radius: f32) -> f32 {
  let localOffset = 8.0;
  return sphere_dist2(localOffset, p) + sphere_dist3(radius, p);
}

fn sphere2(p: vec3<f32>, radius: f32) -> f32 {
  return sphere(p, radius + 7.0);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let a = sphere2(p, 2.5);
  let b = sphere2(p, 2.5);
  let c = sphere2(p, 2.5);
  let gsl_array = array<f32, 7>(1.0, 2.0, 3.0, 4.0, sphere(p, 2.5), 6.0, 7.0);
  let d = vec4(gsl_array[3], gsl_array[2], gsl_array[1], gsl_array[0]).w;
  let gsl_s1 = gsl_shape(a + b + c + d, 1u);
  return gsl_s1;
}
