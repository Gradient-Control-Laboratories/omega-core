fn torus(p: vec3<f32>, majorRadius: f32, minorRadius: f32) -> f32 {
  let q = vec2<f32>(length(p.xy) - majorRadius, p.z);
  return length(q) - minorRadius;
}

fn capsule(p: vec3<f32>, a: vec3<f32>, b: vec3<f32>, r: f32) -> f32 {
  let pa = p - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h) - r;
}

fn box(p: vec3<f32>, c1: vec3<f32>, c2: vec3<f32>) -> f32 {
  return max(max(max(max(max(c1.x - p.x, p.x - c2.x), c1.y - p.y), p.y - c2.y), c1.z - p.z), p.z - c2.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let frontWheel = torus(vec3<f32>(p.z, p.y - 1.5, p.x), 0.9, 0.1);
  let rearWheelL = torus(vec3<f32>(p.z, p.y - -1.2, p.x - 1.0), 0.9, 0.1);
  let rearWheelR = torus(vec3<f32>(p.z, p.y - -1.2, p.x + 1.0), 0.9, 0.1);
  let allWheels = min(min(frontWheel, rearWheelL), rearWheelR);
  let backbone = capsule(p, vec3<f32>(0.0, -1.2, 0.9), vec3<f32>(0.0, 1.5, 0.9), 0.08);
  let topTube = capsule(p, vec3<f32>(0.0, 1.5, 0.9), vec3<f32>(0.0, -0.5, 2.1), 0.08);
  let seatTube = capsule(p, vec3<f32>(0.0, -1.2, 0.9), vec3<f32>(0.0, -0.5, 2.1), 0.08);
  let rearAxle = capsule(p, vec3<f32>(1.0, -1.2, 0.9), vec3<f32>(-1.0, -1.2, 0.9), 0.08);
  let frontFork = capsule(p, vec3<f32>(0.0, 1.5, 0.9), vec3<f32>(0.0, 1.5, 2.6), 0.08);
  let allFrame = min(min(min(min(backbone, topTube), seatTube), rearAxle), frontFork);
  let handlebarTube = capsule(p, vec3<f32>(-0.45, 1.5, 2.6), vec3<f32>(0.45, 1.5, 2.6), 0.08);
  let seat = box(p - vec3<f32>(0.0, -0.65, 2.1), -vec3<f32>(0.3, 0.35, 0.075), vec3<f32>(0.3, 0.35, 0.075));
  let crank = capsule(p, vec3<f32>(-0.4, 1.5, 0.9), vec3<f32>(0.4, 1.5, 0.9), 0.056);
  let pedalL = box(p - vec3<f32>(-0.4, 1.5, 0.9), -vec3<f32>(0.15, 0.075, 0.025), vec3<f32>(0.15, 0.075, 0.025));
  let pedalR = box(p - vec3<f32>(0.4, 1.5, 0.9), -vec3<f32>(0.15, 0.075, 0.025), vec3<f32>(0.15, 0.075, 0.025));
  let allPedals = min(min(crank, pedalL), pedalR);
  let tricycle = min(min(min(min(allWheels, allFrame), handlebarTube), seat), allPedals);
  let gsl_s1 = gsl_shape(tricycle, 1u);
  return gsl_s1;
}
