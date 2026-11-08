fn gsl_ball_remap(p: vec3<f32>) -> f32 {
  return length(p) - 25.0;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-250.0, 87.5, -25.0);
  let gsl_partHi = vec3<f32>(-200.0, 137.5, 25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_ball_remap(p - vec3<f32>(-225.0, 112.5, 0.0)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let concrete = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-175.0, 87.5, -25.0);
  let gsl_partHi_2 = vec3<f32>(-125.0, 137.5, 25.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_ball_remap(p - vec3<f32>(-150.0, 112.5, 0.0)) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let brick = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-100.0, 87.5, -25.0);
  let gsl_partHi_3 = vec3<f32>(-50.0, 137.5, 25.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = gsl_ball_remap(p - vec3<f32>(-75.0, 112.5, 0.0)) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let asphaltFresh = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-25.0, 87.5, -25.0);
  let gsl_partHi_4 = vec3<f32>(25.0, 137.5, 25.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = gsl_ball_remap(p - vec3<f32>(0.0, 112.5, 0.0)) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let clay = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(50.0, 87.5, -25.0);
  let gsl_partHi_5 = vec3<f32>(100.0, 137.5, 25.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = gsl_ball_remap(p - vec3<f32>(75.0, 112.5, 0.0)) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let terracotta = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(125.0, 87.5, -25.0);
  let gsl_partHi_6 = vec3<f32>(175.0, 137.5, 25.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = gsl_ball_remap(p - vec3<f32>(150.0, 112.5, 0.0)) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let porcelain = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(200.0, 87.5, -25.0);
  let gsl_partHi_7 = vec3<f32>(250.0, 137.5, 25.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = gsl_ball_remap(p - vec3<f32>(225.0, 112.5, 0.0)) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let marble = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-250.0, 12.5, -25.0);
  let gsl_partHi_8 = vec3<f32>(-200.0, 62.5, 25.0);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = gsl_ball_remap(p - vec3<f32>(-225.0, 37.5, 0.0)) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let cardboard = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(-175.0, 12.5, -25.0);
  let gsl_partHi_9 = vec3<f32>(-125.0, 62.5, 25.0);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = gsl_ball_remap(p - vec3<f32>(-150.0, 37.5, 0.0)) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let officePaper = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(-100.0, 12.5, -25.0);
  let gsl_partHi_10 = vec3<f32>(-50.0, 62.5, 25.0);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = gsl_ball_remap(p - vec3<f32>(-75.0, 37.5, 0.0)) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let toiletPaper = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(-25.0, 12.5, -25.0);
  let gsl_partHi_11 = vec3<f32>(25.0, 62.5, 25.0);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = gsl_ball_remap(p - vec3<f32>(0.0, 37.5, 0.0)) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let whiteboard = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(50.0, 12.5, -25.0);
  let gsl_partHi_12 = vec3<f32>(100.0, 62.5, 25.0);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = gsl_ball_remap(p - vec3<f32>(75.0, 37.5, 0.0)) + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let blackboard = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(125.0, 12.5, -25.0);
  let gsl_partHi_13 = vec3<f32>(175.0, 62.5, 25.0);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = gsl_ball_remap(p - vec3<f32>(150.0, 37.5, 0.0)) + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let grayCard = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(200.0, 12.5, -25.0);
  let gsl_partHi_14 = vec3<f32>(250.0, 62.5, 25.0);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = gsl_ball_remap(p - vec3<f32>(225.0, 37.5, 0.0)) + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let tonerBlack = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(-250.0, -62.5, -25.0);
  let gsl_partHi_15 = vec3<f32>(-200.0, -12.5, 25.0);
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = gsl_ball_remap(p - vec3<f32>(-225.0, -37.5, 0.0)) + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let sand = gsl_part_15;
  let gsl_partLo_16 = vec3<f32>(-175.0, -62.5, -25.0);
  let gsl_partHi_16 = vec3<f32>(-125.0, -12.5, 25.0);
  let gsl_partBox_16 = gsl_boxDistance(p, gsl_partLo_16, gsl_partHi_16);
  var gsl_part_16 = gsl_partBox_16;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_16 <= gsl_vanishView.w) {
    let gsl_partVanish_16 = gsl_vanish(gsl_partLo_16, gsl_partHi_16);
    gsl_part_16 = 1e10;
    if (gsl_partVanish_16.x > 0.0) { gsl_part_16 = gsl_ball_remap(p - vec3<f32>(-150.0, -37.5, 0.0)) + gsl_partVanish_16.y; }
    if (abs(gsl_part_16) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_16), gsl_partVanish_16.zw); }
  }
  let snow = gsl_part_16;
  let gsl_partLo_17 = vec3<f32>(-100.0, -62.5, -25.0);
  let gsl_partHi_17 = vec3<f32>(-50.0, -12.5, 25.0);
  let gsl_partBox_17 = gsl_boxDistance(p, gsl_partLo_17, gsl_partHi_17);
  var gsl_part_17 = gsl_partBox_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_17 <= gsl_vanishView.w) {
    let gsl_partVanish_17 = gsl_vanish(gsl_partLo_17, gsl_partHi_17);
    gsl_part_17 = 1e10;
    if (gsl_partVanish_17.x > 0.0) { gsl_part_17 = gsl_ball_remap(p - vec3<f32>(-75.0, -37.5, 0.0)) + gsl_partVanish_17.y; }
    if (abs(gsl_part_17) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_17), gsl_partVanish_17.zw); }
  }
  let grass = gsl_part_17;
  let gsl_partLo_18 = vec3<f32>(-25.0, -62.5, -25.0);
  let gsl_partHi_18 = vec3<f32>(25.0, -12.5, 25.0);
  let gsl_partBox_18 = gsl_boxDistance(p, gsl_partLo_18, gsl_partHi_18);
  var gsl_part_18 = gsl_partBox_18;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_18 <= gsl_vanishView.w) {
    let gsl_partVanish_18 = gsl_vanish(gsl_partLo_18, gsl_partHi_18);
    gsl_part_18 = 1e10;
    if (gsl_partVanish_18.x > 0.0) { gsl_part_18 = gsl_ball_remap(p - vec3<f32>(0.0, -37.5, 0.0)) + gsl_partVanish_18.y; }
    if (abs(gsl_part_18) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_18), gsl_partVanish_18.zw); }
  }
  let charcoal = gsl_part_18;
  let gsl_partLo_19 = vec3<f32>(50.0, -62.5, -25.0);
  let gsl_partHi_19 = vec3<f32>(100.0, -12.5, 25.0);
  let gsl_partBox_19 = gsl_boxDistance(p, gsl_partLo_19, gsl_partHi_19);
  var gsl_part_19 = gsl_partBox_19;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_19 <= gsl_vanishView.w) {
    let gsl_partVanish_19 = gsl_vanish(gsl_partLo_19, gsl_partHi_19);
    gsl_part_19 = 1e10;
    if (gsl_partVanish_19.x > 0.0) { gsl_part_19 = gsl_ball_remap(p - vec3<f32>(75.0, -37.5, 0.0)) + gsl_partVanish_19.y; }
    if (abs(gsl_part_19) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_19), gsl_partVanish_19.zw); }
  }
  let bone = gsl_part_19;
  let gsl_partLo_20 = vec3<f32>(125.0, -62.5, -25.0);
  let gsl_partHi_20 = vec3<f32>(175.0, -12.5, 25.0);
  let gsl_partBox_20 = gsl_boxDistance(p, gsl_partLo_20, gsl_partHi_20);
  var gsl_part_20 = gsl_partBox_20;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_20 <= gsl_vanishView.w) {
    let gsl_partVanish_20 = gsl_vanish(gsl_partLo_20, gsl_partHi_20);
    gsl_part_20 = 1e10;
    if (gsl_partVanish_20.x > 0.0) { gsl_part_20 = gsl_ball_remap(p - vec3<f32>(150.0, -37.5, 0.0)) + gsl_partVanish_20.y; }
    if (abs(gsl_part_20) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_20), gsl_partVanish_20.zw); }
  }
  let eggShellWhite = gsl_part_20;
  let gsl_partLo_21 = vec3<f32>(200.0, -62.5, -25.0);
  let gsl_partHi_21 = vec3<f32>(250.0, -12.5, 25.0);
  let gsl_partBox_21 = gsl_boxDistance(p, gsl_partLo_21, gsl_partHi_21);
  var gsl_part_21 = gsl_partBox_21;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_21 <= gsl_vanishView.w) {
    let gsl_partVanish_21 = gsl_vanish(gsl_partLo_21, gsl_partHi_21);
    gsl_part_21 = 1e10;
    if (gsl_partVanish_21.x > 0.0) { gsl_part_21 = gsl_ball_remap(p - vec3<f32>(225.0, -37.5, 0.0)) + gsl_partVanish_21.y; }
    if (abs(gsl_part_21) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_21), gsl_partVanish_21.zw); }
  }
  let eggShellBrown = gsl_part_21;
  let gsl_partLo_22 = vec3<f32>(-250.0, -137.5, -25.0);
  let gsl_partHi_22 = vec3<f32>(-200.0, -87.5, 25.0);
  let gsl_partBox_22 = gsl_boxDistance(p, gsl_partLo_22, gsl_partHi_22);
  var gsl_part_22 = gsl_partBox_22;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_22 <= gsl_vanishView.w) {
    let gsl_partVanish_22 = gsl_vanish(gsl_partLo_22, gsl_partHi_22);
    gsl_part_22 = 1e10;
    if (gsl_partVanish_22.x > 0.0) { gsl_part_22 = gsl_ball_remap(p - vec3<f32>(-225.0, -112.5, 0.0)) + gsl_partVanish_22.y; }
    if (abs(gsl_part_22) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_22), gsl_partVanish_22.zw); }
  }
  let rust = gsl_part_22;
  let gsl_partLo_23 = vec3<f32>(-175.0, -137.5, -25.0);
  let gsl_partHi_23 = vec3<f32>(-125.0, -87.5, 25.0);
  let gsl_partBox_23 = gsl_boxDistance(p, gsl_partLo_23, gsl_partHi_23);
  var gsl_part_23 = gsl_partBox_23;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_23 <= gsl_vanishView.w) {
    let gsl_partVanish_23 = gsl_vanish(gsl_partLo_23, gsl_partHi_23);
    gsl_part_23 = 1e10;
    if (gsl_partVanish_23.x > 0.0) { gsl_part_23 = gsl_ball_remap(p - vec3<f32>(-150.0, -112.5, 0.0)) + gsl_partVanish_23.y; }
    if (abs(gsl_part_23) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_23), gsl_partVanish_23.zw); }
  }
  let tire = gsl_part_23;
  let gsl_partLo_24 = vec3<f32>(-100.0, -137.5, -25.0);
  let gsl_partHi_24 = vec3<f32>(-50.0, -87.5, 25.0);
  let gsl_partBox_24 = gsl_boxDistance(p, gsl_partLo_24, gsl_partHi_24);
  var gsl_part_24 = gsl_partBox_24;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_24 <= gsl_vanishView.w) {
    let gsl_partVanish_24 = gsl_vanish(gsl_partLo_24, gsl_partHi_24);
    gsl_part_24 = 1e10;
    if (gsl_partVanish_24.x > 0.0) { gsl_part_24 = gsl_ball_remap(p - vec3<f32>(-75.0, -112.5, 0.0)) + gsl_partVanish_24.y; }
    if (abs(gsl_part_24) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_24), gsl_partVanish_24.zw); }
  }
  let carPaint = gsl_part_24;
  let gsl_partLo_25 = vec3<f32>(-25.0, -137.5, -25.0);
  let gsl_partHi_25 = vec3<f32>(25.0, -87.5, 25.0);
  let gsl_partBox_25 = gsl_boxDistance(p, gsl_partLo_25, gsl_partHi_25);
  var gsl_part_25 = gsl_partBox_25;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_25 <= gsl_vanishView.w) {
    let gsl_partVanish_25 = gsl_vanish(gsl_partLo_25, gsl_partHi_25);
    gsl_part_25 = 1e10;
    if (gsl_partVanish_25.x > 0.0) { gsl_part_25 = gsl_ball_remap(p - vec3<f32>(0.0, -112.5, 0.0)) + gsl_partVanish_25.y; }
    if (abs(gsl_part_25) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_25), gsl_partVanish_25.zw); }
  }
  let toothpaste = gsl_part_25;
  let gsl_partLo_26 = vec3<f32>(50.0, -137.5, -25.0);
  let gsl_partHi_26 = vec3<f32>(100.0, -87.5, 25.0);
  let gsl_partBox_26 = gsl_boxDistance(p, gsl_partLo_26, gsl_partHi_26);
  var gsl_part_26 = gsl_partBox_26;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_26 <= gsl_vanishView.w) {
    let gsl_partVanish_26 = gsl_vanish(gsl_partLo_26, gsl_partHi_26);
    gsl_part_26 = 1e10;
    if (gsl_partVanish_26.x > 0.0) { gsl_part_26 = gsl_ball_remap(p - vec3<f32>(75.0, -112.5, 0.0)) + gsl_partVanish_26.y; }
    if (abs(gsl_part_26) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_26), gsl_partVanish_26.zw); }
  }
  let spectralon = gsl_part_26;
  let gsl_partLo_27 = vec3<f32>(125.0, -137.5, -25.0);
  let gsl_partHi_27 = vec3<f32>(175.0, -87.5, 25.0);
  let gsl_partBox_27 = gsl_boxDistance(p, gsl_partLo_27, gsl_partHi_27);
  var gsl_part_27 = gsl_partBox_27;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_27 <= gsl_vanishView.w) {
    let gsl_partVanish_27 = gsl_vanish(gsl_partLo_27, gsl_partHi_27);
    gsl_part_27 = 1e10;
    if (gsl_partVanish_27.x > 0.0) { gsl_part_27 = gsl_ball_remap(p - vec3<f32>(150.0, -112.5, 0.0)) + gsl_partVanish_27.y; }
    if (abs(gsl_part_27) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_27), gsl_partVanish_27.zw); }
  }
  let musouBlack = gsl_part_27;
  let gsl_partLo_28 = vec3<f32>(200.0, -137.5, -25.0);
  let gsl_partHi_28 = vec3<f32>(250.0, -87.5, 25.0);
  let gsl_partBox_28 = gsl_boxDistance(p, gsl_partLo_28, gsl_partHi_28);
  var gsl_part_28 = gsl_partBox_28;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_28 <= gsl_vanishView.w) {
    let gsl_partVanish_28 = gsl_vanish(gsl_partLo_28, gsl_partHi_28);
    gsl_part_28 = 1e10;
    if (gsl_partVanish_28.x > 0.0) { gsl_part_28 = gsl_ball_remap(p - vec3<f32>(225.0, -112.5, 0.0)) + gsl_partVanish_28.y; }
    if (abs(gsl_part_28) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_28), gsl_partVanish_28.zw); }
  }
  let mitBlack = gsl_part_28;
  let gsl_s1 = gsl_shape(concrete, 1u);
  let gsl_s2 = gsl_shape(brick, 2u);
  let gsl_s3 = gsl_shape(asphaltFresh, 3u);
  let gsl_s4 = gsl_shape(clay, 4u);
  let gsl_s5 = gsl_shape(terracotta, 5u);
  let gsl_s6 = gsl_shape(porcelain, 6u);
  let gsl_s7 = gsl_shape(marble, 7u);
  let gsl_s8 = gsl_shape(cardboard, 8u);
  let gsl_s9 = gsl_shape(officePaper, 9u);
  let gsl_s10 = gsl_shape(toiletPaper, 10u);
  let gsl_s11 = gsl_shape(whiteboard, 11u);
  let gsl_s12 = gsl_shape(blackboard, 12u);
  let gsl_s13 = gsl_shape(grayCard, 13u);
  let gsl_s14 = gsl_shape(tonerBlack, 14u);
  let gsl_s15 = gsl_shape(sand, 15u);
  let gsl_s16 = gsl_shape(snow, 16u);
  let gsl_s17 = gsl_shape(grass, 17u);
  let gsl_s18 = gsl_shape(charcoal, 18u);
  let gsl_s19 = gsl_shape(bone, 19u);
  let gsl_s20 = gsl_shape(eggShellWhite, 20u);
  let gsl_s21 = gsl_shape(eggShellBrown, 21u);
  let gsl_s22 = gsl_shape(rust, 22u);
  let gsl_s23 = gsl_shape(tire, 23u);
  let gsl_s24 = gsl_shape(carPaint, 24u);
  let gsl_s25 = gsl_shape(toothpaste, 25u);
  let gsl_s26 = gsl_shape(spectralon, 26u);
  let gsl_s27 = gsl_shape(musouBlack, 27u);
  let gsl_s28 = gsl_shape(mitBlack, 28u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21), gsl_s22), gsl_s23), gsl_s24), gsl_s25), gsl_s26), gsl_s27), gsl_s28);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.7419, 0.7419, 0.7419), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.5488, 0.3406, 0.274), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.2291, 0.2249, 0.2212), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.543, 0.378, 0.3336), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.7705, 0.4978, 0.3656), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(0.8782, 0.8782, 0.8666), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(0.9212, 0.9018, 0.8824), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(vec3<f32>(0.627, 0.4934, 0.3656), 1.0); }
  if (gsl_colorIdx == 9u) { return vec4<f32>(vec3<f32>(0.9033, 0.9231, 0.9472), 1.0); }
  if (gsl_colorIdx == 10u) { return vec4<f32>(vec3<f32>(0.9212, 0.9236, 0.8983), 1.0); }
  if (gsl_colorIdx == 11u) { return vec4<f32>(vec3<f32>(0.94, 0.9391, 0.8917), 1.0); }
  if (gsl_colorIdx == 12u) { return vec4<f32>(vec3<f32>(0.218, 0.218, 0.218), 1.0); }
  if (gsl_colorIdx == 13u) { return vec4<f32>(vec3<f32>(0.4614, 0.4614, 0.4614), 1.0); }
  if (gsl_colorIdx == 14u) { return vec4<f32>(vec3<f32>(0.2478, 0.2478, 0.2478), 1.0); }
  if (gsl_colorIdx == 15u) { return vec4<f32>(vec3<f32>(0.6944, 0.6546, 0.5176), 1.0); }
  if (gsl_colorIdx == 16u) { return vec4<f32>(vec3<f32>(0.9309, 0.9309, 0.9309), 1.0); }
  if (gsl_colorIdx == 17u) { return vec4<f32>(vec3<f32>(0.3575, 0.4002, 0.2238), 1.0); }
  if (gsl_colorIdx == 18u) { return vec4<f32>(vec3<f32>(0.1517, 0.1517, 0.1517), 1.0); }
  if (gsl_colorIdx == 19u) { return vec4<f32>(vec3<f32>(0.9028, 0.9028, 0.8345), 1.0); }
  if (gsl_colorIdx == 20u) { return vec4<f32>(vec3<f32>(0.8036, 0.8118, 0.8158), 1.0); }
  if (gsl_colorIdx == 21u) { return vec4<f32>(vec3<f32>(0.7307, 0.5351, 0.3856), 1.0); }
  if (gsl_colorIdx == 22u) { return vec4<f32>(vec3<f32>(0.4625, 0.2648, 0.1382), 1.0); }
  if (gsl_colorIdx == 23u) { return vec4<f32>(vec3<f32>(0.1641, 0.1641, 0.1641), 1.0); }
  if (gsl_colorIdx == 24u) { return vec4<f32>(vec3<f32>(0.3492, 0.3492, 0.3492), 1.0); }
  if (gsl_colorIdx == 25u) { return vec4<f32>(vec3<f32>(0.9697, 0.9717, 0.9679), 1.0); }
  if (gsl_colorIdx == 26u) { return vec4<f32>(vec3<f32>(0.9956, 0.9956, 0.9956), 1.0); }
  if (gsl_colorIdx == 27u) { return vec4<f32>(vec3<f32>(0.0702, 0.0702, 0.0702), 1.0); }
  if (gsl_colorIdx == 28u) { return vec4<f32>(vec3<f32>(0.0006, 0.0006, 0.0006), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.6, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 9u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.8, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 10u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.9, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 11u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 12u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.9, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 13u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.9, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 14u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 15u) { let gsl_m = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 16u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.3098, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 17u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 18u) { let gsl_m = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 19u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.9, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 20u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.9, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 21u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.9, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 22u) { let gsl_m = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 23u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.7, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 24u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 25u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 26u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.99, 0.0, 1.35, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 27u) { let gsl_m = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 28u) { let gsl_m = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
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
