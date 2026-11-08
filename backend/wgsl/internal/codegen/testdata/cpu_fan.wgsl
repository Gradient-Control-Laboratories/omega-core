fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return max(box(p, vec3<f32>(3.5, 0.75, 1.4)), abs(gsl_floored((p.z + 0.075), 0.15) - 0.075) - 0.02);
}

fn box(p: vec3<f32>, b: vec3<f32>) -> f32 {
  let q = abs(p) - b;
  return length(max(q, vec3<f32>(0.0))) + min(max(max(q.x, q.y), q.z), 0.0);
}

fn cylinderY(p: vec3<f32>, r: f32, h: f32) -> f32 {
  let d = p.xz;
  return max(length(d) - r, abs(p.y) - h);
}

fn fan_assembly_rotated_blade(p: vec3<f32>, fan_scale: f32, angle: f32) -> f32 {
  let c = cos(angle);
  let s = sin(angle);
  let p_rot_x = p.x * c + p.z * s;
  let p_rot_z = -p.x * s + p.z * c;
  let blade_offset = vec3<f32>(0.56, 0.0, 0.0);
  let p_remapped = vec3<f32>(p_rot_x, p.y, p_rot_z) - blade_offset;
  return box(p_remapped, vec3<f32>(0.64, 0.08, 0.12));
}

fn fan_assembly(p: vec3<f32>) -> f32 {
  let hub = cylinderY(p, 0.24, 0.16);
  let fan_scale = 0.8;
  let b1 = fan_assembly_rotated_blade(p, fan_scale, 0.0);
  let b2 = fan_assembly_rotated_blade(p, fan_scale, 0.89709926);
  let b3 = fan_assembly_rotated_blade(p, fan_scale, 1.7941985);
  let b4 = fan_assembly_rotated_blade(p, fan_scale, 2.6912978);
  let b5 = fan_assembly_rotated_blade(p, fan_scale, 3.588397);
  let b6 = fan_assembly_rotated_blade(p, fan_scale, 4.485496);
  let b7 = fan_assembly_rotated_blade(p, fan_scale, 5.3825955);
  let blades = min(min(min(min(min(min(b1, b2), b3), b4), b5), b6), b7);
  return min(hub, blades);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let pcb = box(p, vec3<f32>(4.0, 0.1, 1.5));
  let shroud = gsl_remap(p - vec3<f32>(0.0, 0.85, 0.0));
  let fan1_hole = cylinderY(p - vec3<f32>(-1.7, 0.85, 0.0), 1.2, 0.8);
  let fan2_hole = cylinderY(p - vec3<f32>(1.7, 0.85, 0.0), 1.2, 0.8);
  let holes = min(fan1_hole, fan2_hole);
  let shroud_with_holes = max(shroud, -holes);
  let fan1 = fan_assembly(p - vec3<f32>(-1.7, 0.85, 0.0));
  let fan2 = fan_assembly(p - vec3<f32>(1.7, 0.85, 0.0));
  let all_fans = min(fan1, fan2);
  let gpu_chip = box(p - vec3<f32>(0.0, -0.25, 0.0), vec3<f32>(0.8, 0.15, 0.8));
  let io_bracket = box(p - vec3<f32>(-4.1, 0.8, 0.0), vec3<f32>(0.1, 0.8, 1.4));
  let gsl_s1 = gsl_shape(pcb, 1u);
  let gsl_s2 = gsl_shape(shroud_with_holes, 2u);
  let gsl_s3 = gsl_shape(all_fans, 3u);
  let gsl_s4 = gsl_shape(gpu_chip, 4u);
  let gsl_s5 = gsl_shape(io_bracket, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}
