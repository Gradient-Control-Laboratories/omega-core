fn basis(p: vec3<f32>) -> gsl_struct {
  let px = p.x;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = px;
  gsl_result.v_2 = p.y;
  gsl_result.v_3 = p.z;
  return gsl_result;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_result = basis(p);
  let bx = gsl_result.v_1;
  let gsl_result_2 = basis(p + 7.0);
  let cy = gsl_result_2.v_2;
  let gsl_s1 = gsl_shape(bx, 1u);
  let gsl_s2 = gsl_shape(cy, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : f32,
}
