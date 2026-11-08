fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let a = length(p) - 1.0;
  let b = length(p) - 1.0 + 0.5;
  let c = length(p) - 1.0 - 0.5;
  let shapes = vec3<f32>(a, b, c);
  let unioned = min(min(shapes.x, shapes.y), shapes.z);
  let intersected = max(max(a, b), c);
  let diffOne = max(a, -(b));
  let diffSubList = max(max(a, -(b)), -(c));
  let diffBaseList = max(min(a, b), -(c));
  let diffBoth = max(max(min(a, b), -(c)), -(b));
  let gsl_s1 = gsl_shape(min(min(min(min(min(unioned, intersected), diffOne), diffSubList), diffBaseList), diffBoth), 1u);
  return gsl_s1;
}
