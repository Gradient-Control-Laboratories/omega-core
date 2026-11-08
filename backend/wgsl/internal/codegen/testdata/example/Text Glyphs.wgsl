fn gsl_remap(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(min(text_arcDist(p, 4.9265, 5.629, 4.7978, 2.7289, 0.2296), text_arcDist(p, 9.3919, 4.5329, 9.3918, 3.0376, 0.1073)), text_arcDist(p, 9.3954, 4.4638, 9.3953, -3.0383, 0.1073)), text_arcDist(p, 4.926, 3.3746, 4.7989, -2.7288, 0.2296)), text_arcDist(p, 2.0001, 1.0884, 1.0892, -1.5721, 0.9983)), text_arcDist(p, -0.9267, 3.3719, 4.7982, -0.4128, 0.2296)), text_arcDist(p, -5.3919, 4.4671, 9.3918, -0.104, 0.1073)), text_arcDist(p, -5.3993, 4.5343, 9.3993, 0.1035, 0.1073)), text_arcDist(p, -0.9305, 5.624, 4.8032, 0.4127, 0.2292)), text_arcDist(p, 1.9997, 7.9114, 1.0895, 1.5691, 0.9986));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(2.0, 9.0)), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(2.0, 0.0)));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.865, 6.647))), text_arcDist(p, 3.0013, 7.1488, 0.9989, -0.0837, 0.4426)), text_arcDist(p, 2.0, 6.9998, 2.0, 1.5705, 1.3179));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 0.0, 1.5708)), text_arcDist(p, 2.0, 3.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 5.0), vec2<f32>(1.0, 5.0)));
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(3.5, 0.0), vec2<f32>(3.5, 4.0)), text_segmentDist(p, vec2<f32>(5.0, 2.0), vec2<f32>(0.0, 2.0))), text_segmentDist(p, vec2<f32>(0.0, 2.0), vec2<f32>(2.0, 9.0)));
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 5.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(2.0, 5.0))), text_arcDist(p, 2.0, 3.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(0.0, 0.0)));
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 6.001, 3.8035, 6.001, 2.618, 0.5235), text_segmentDist(p, vec2<f32>(0.0, 3.804), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 3.0))), text_arcDist(p, 2.0, 3.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 5.0), vec2<f32>(0.1204, 5.0)));
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(4.0, 9.0)), text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(1.5, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 5.0), vec2<f32>(4.0, 5.0)));
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 3.0))), text_arcDist(p, 2.0, 3.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 6.75, 1.75, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(3.75, 6.75), vec2<f32>(3.75, 7.25))), text_arcDist(p, 2.0, 7.25, 1.75, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.25, 7.25), vec2<f32>(0.25, 6.75)));
}

fn gsl_remap_10(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, -2.001, 5.1965, 6.001, -0.5236, 0.5235), text_segmentDist(p, vec2<f32>(4.0, 5.196), vec2<f32>(4.0, 7.0))), text_arcDist(p, 2.0, 7.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 6.0))), text_arcDist(p, 2.0, 6.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 4.0), vec2<f32>(3.88, 4.0)));
}

fn gsl_remap_11(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 3.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 0.5)));
}

fn gsl_remap_12(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.5, 7.0), vec2<f32>(1.0, 9.0)), text_segmentDist(p, vec2<f32>(3.5, 7.0), vec2<f32>(4.0, 9.0)));
}

fn gsl_remap_13(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(2.0, 9.0)), text_segmentDist(p, vec2<f32>(5.0, 9.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 2.5), vec2<f32>(7.0, 2.5))), text_segmentDist(p, vec2<f32>(7.0, 6.5), vec2<f32>(0.0, 6.5)));
}

fn gsl_remap_14(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 1.9994, 2.3723, 6.6277, 1.3358, 0.2349), text_arcDist(p, 2.0001, 7.0, 2.0, 2.906, 1.3351)), text_segmentDist(p, vec2<f32>(1.092, 5.218), vec2<f32>(3.909, 3.782))), text_arcDist(p, 3.0002, 2.0002, 2.0002, -0.2359, 1.335)), text_arcDist(p, 3.0001, 6.6259, 6.6259, -1.8055, 0.2347)), text_segmentDist(p, vec2<f32>(2.5, -1.0), vec2<f32>(2.5, 10.0)));
}

fn gsl_remap_15(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(7.0, 9.0)), text_arcDist(p, 1.6, 7.4, 1.6, 3.1416, 1.5708)), text_arcDist(p, 1.6, 7.4, 1.6, 6.2832, 1.5708)), text_arcDist(p, 5.4, 1.6, 1.6, 6.2832, 1.5708)), text_arcDist(p, 5.4, 1.6, 1.6, 3.1416, 1.5708));
}

fn gsl_remap_16(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(6.0, 3.215), vec2<f32>(3.426, 0.5976)), text_arcDist(p, 1.9998, 1.9997, 2.0, -2.3477, 1.5708)), text_segmentDist(p, vec2<f32>(0.5735, 3.4018), vec2<f32>(3.569, 6.448))), text_arcDist(p, 2.4998, 7.4993, 1.4995, 1.4387, 2.2157)), text_segmentDist(p, vec2<f32>(1.1932, 6.7636), vec2<f32>(5.0, 0.0)));
}

fn gsl_remap_17(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(4.0, 2.0)), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(0.0, 2.0)));
}

fn gsl_remap_18(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(4.0, 4.0)), text_segmentDist(p, vec2<f32>(2.0, 6.0), vec2<f32>(2.0, 2.0)));
}

fn gsl_remap_19(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(1.0, 0.5), vec2<f32>(1.0, 0.0)), text_segmentDist(p, vec2<f32>(1.0, 0.0), vec2<f32>(0.0, -3.0)));
}

fn gsl_remap_20(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 0.5)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 3.5)));
}

fn gsl_remap_21(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(1.0, 0.5), vec2<f32>(1.0, 0.0)), text_segmentDist(p, vec2<f32>(1.0, 0.0), vec2<f32>(0.0, -3.0))), text_segmentDist(p, vec2<f32>(1.0, 4.0), vec2<f32>(1.0, 3.5)));
}

fn gsl_remap_22(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(4.0, 7.0), vec2<f32>(0.0, 3.5)), text_segmentDist(p, vec2<f32>(0.0, 3.5), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_23(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 5.5), vec2<f32>(4.0, 5.5)), text_segmentDist(p, vec2<f32>(0.0, 2.5), vec2<f32>(4.0, 2.5)));
}

fn gsl_remap_24(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(4.0, 3.5)), text_segmentDist(p, vec2<f32>(4.0, 3.5), vec2<f32>(0.0, 0.0)));
}

fn gsl_remap_25(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(2.0, 0.5)), text_segmentDist(p, vec2<f32>(2.0, 3.0), vec2<f32>(2.0, 3.394))), text_arcDist(p, 4.0002, 3.3939, 2.0002, 2.8475, 0.2941)), text_segmentDist(p, vec2<f32>(2.3361, 4.5037), vec2<f32>(3.664, 6.496))), text_arcDist(p, 1.9996, 7.6056, 2.0004, -0.3079, 0.2802)), text_arcDist(p, 2.5007, 7.5007, 1.4993, 0.8021, 0.7691)), text_segmentDist(p, vec2<f32>(2.5, 9.0), vec2<f32>(1.5, 9.0))), text_arcDist(p, 1.5, 7.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.5), vec2<f32>(0.0, 7.0)));
}

fn gsl_remap_26(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(8.0, 1.0), vec2<f32>(6.068, 0.2819)), text_arcDist(p, 4.4993, 4.5005, 4.5008, 2.3029, 2.7655)), text_arcDist(p, 7.4137, 2.9435, 1.2015, -1.8616, 1.4765)), text_arcDist(p, 5.1699, 5.5, 1.5, 0.6868, 0.884)), text_segmentDist(p, vec2<f32>(5.17, 7.0), vec2<f32>(5.05, 7.0))), text_arcDist(p, 5.0495, 4.5, 2.5, 2.2575, 0.6869)), text_segmentDist(p, vec2<f32>(2.598, 4.99), vec2<f32>(2.4788, 4.3919))), text_arcDist(p, 4.4401, 4.0001, 2.0001, -2.4548, 0.884)), text_segmentDist(p, vec2<f32>(4.44, 2.0), vec2<f32>(4.36, 2.0))), text_arcDist(p, 4.3601, 3.9998, 1.9998, -0.884, 0.6869)), text_segmentDist(p, vec2<f32>(6.3212, 3.6081), vec2<f32>(7.0, 7.0)));
}

fn gsl_remap_27(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(1.0, -1.0), vec2<f32>(0.0, -1.0)), text_segmentDist(p, vec2<f32>(0.0, -1.0), vec2<f32>(0.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(1.0, 9.0)));
}

fn gsl_remap_28(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(1.0, 9.0)), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(1.0, -1.0))), text_segmentDist(p, vec2<f32>(1.0, -1.0), vec2<f32>(0.0, -1.0)));
}

fn gsl_remap_29(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 8.0), vec2<f32>(2.0, 10.0)), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(4.0, 8.0)));
}

fn gsl_remap_30(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 2.0, 8.0, 1.0, 2.3562, 0.7854), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(1.0, 5.0))), text_arcDist(p, -0.0, 5.0, 1.0, -0.7854, 0.7854)), text_arcDist(p, -0.0, 3.0, 1.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(1.0, 3.0), vec2<f32>(1.0, 0.0))), text_arcDist(p, 2.0, 0.0, 1.0, -2.3562, 0.7854));
}

fn gsl_remap_31(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, -0.0, 0.0, 1.0, -0.7854, 0.7854), text_segmentDist(p, vec2<f32>(1.0, 0.0), vec2<f32>(1.0, 3.0))), text_arcDist(p, 2.0, 3.0, 1.0, 2.3562, 0.7854)), text_arcDist(p, 2.0, 5.0, 1.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.0, 5.0), vec2<f32>(1.0, 8.0))), text_arcDist(p, -0.0, 8.0, 1.0, 0.7854, 0.7854));
}

fn gsl_remap_32(p: vec3<f32>) -> f32 {
  return min(min(text_arcDist(p, 1.6, 4.2001, 1.2, 2.0656, 1.0761), text_segmentDist(p, vec2<f32>(2.259, 5.203), vec2<f32>(3.941, 4.097))), text_arcDist(p, 4.5999, 5.1001, 1.2001, -1.076, 1.076));
}

fn gsl_remap_33(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5)));
}

fn gsl_remap_34(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(2.6, 5.0)), text_arcDist(p, 2.5999, 2.5999, 2.4001, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 2.6), vec2<f32>(5.0, 2.4))), text_arcDist(p, 2.5999, 2.4001, 2.4001, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.6, 0.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(2.5, 9.0))), text_arcDist(p, 2.5, 7.0, 2.0, 0.0, 1.5708));
}

fn gsl_remap_35(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_36(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854), text_segmentDist(p, vec2<f32>(5.0, 2.0), vec2<f32>(5.0, 7.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(0.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 0.0)));
}

fn gsl_remap_37(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0)));
}

fn gsl_remap_38(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(4.0, 5.0)));
}

fn gsl_remap_39(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(5.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 5.0))), text_segmentDist(p, vec2<f32>(5.0, 5.0), vec2<f32>(3.5, 5.0)));
}

fn gsl_remap_40(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(5.0, 5.0)));
}

fn gsl_remap_41(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(3.0, 2.0)), text_arcDist(p, 1.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(1.0, 0.0), vec2<f32>(0.0, 0.0)));
}

fn gsl_remap_42(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 3.5), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(1.671, 5.338), vec2<f32>(5.0, 0.0)));
}

fn gsl_remap_43(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_44(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(3.0, 4.0))), text_segmentDist(p, vec2<f32>(3.0, 4.0), vec2<f32>(6.0, 9.0))), text_segmentDist(p, vec2<f32>(6.0, 9.0), vec2<f32>(6.0, 0.0)));
}

fn gsl_remap_45(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 9.0)));
}

fn gsl_remap_46(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 2.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 0.0), vec2<f32>(2.0, 0.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 2.0), vec2<f32>(0.0, 7.0)));
}

fn gsl_remap_47(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 6.0))), text_arcDist(p, 3.0, 6.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 4.0), vec2<f32>(0.0, 4.0)));
}

fn gsl_remap_48(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 2.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 0.0), vec2<f32>(2.0, 0.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 2.0), vec2<f32>(0.0, 7.0))), text_segmentDist(p, vec2<f32>(6.0, 0.0), vec2<f32>(3.0, 2.0)));
}

fn gsl_remap_49(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 6.0))), text_arcDist(p, 3.0, 6.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 4.0), vec2<f32>(0.0, 4.0))), text_segmentDist(p, vec2<f32>(3.0, 4.0), vec2<f32>(5.0, 0.0)));
}

fn gsl_remap_50(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_arcDist(p, 1.9994, 2.3723, 6.6277, 1.3358, 0.2349), text_arcDist(p, 2.0001, 7.0, 2.0, 2.906, 1.3351)), text_segmentDist(p, vec2<f32>(1.092, 5.218), vec2<f32>(3.909, 3.782))), text_arcDist(p, 3.0002, 2.0002, 2.0002, -0.2359, 1.335)), text_arcDist(p, 3.0001, 6.6259, 6.6259, -1.8055, 0.2347));
}

fn gsl_remap_51(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(6.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(3.0, 0.0)));
}

fn gsl_remap_52(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 2.5)), text_arcDist(p, 2.5, 2.5, 2.5, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(5.0, 2.5), vec2<f32>(5.0, 9.0)));
}

fn gsl_remap_53(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(3.0, 0.0)), text_segmentDist(p, vec2<f32>(3.0, 0.0), vec2<f32>(6.0, 9.0)));
}

fn gsl_remap_54(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(2.0, 0.0)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(6.0, 0.0), vec2<f32>(8.0, 9.0)));
}

fn gsl_remap_55(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(6.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(6.0, 9.0)));
}

fn gsl_remap_56(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(3.0, 5.0)), text_segmentDist(p, vec2<f32>(3.0, 5.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(3.0, 5.0), vec2<f32>(6.0, 9.0)));
}

fn gsl_remap_57(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(5.0, 9.0)), text_segmentDist(p, vec2<f32>(5.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(5.0, 0.0)));
}

fn gsl_remap_58(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0)));
}

fn gsl_remap_59(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(2.5, 0.0))), text_arcDist(p, 2.5, 1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 1.5), vec2<f32>(4.0, 4.5))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.5, 6.0), vec2<f32>(0.0, 6.0)));
}

fn gsl_remap_60(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(3.0, 6.0), vec2<f32>(1.5, 6.0)), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(3.0, 0.0)));
}

fn gsl_remap_61(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(4.0, 0.0)), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 1.5), vec2<f32>(0.0, 4.5))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 6.0), vec2<f32>(4.0, 6.0)));
}

fn gsl_remap_62(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_63(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(1.0, 0.0), vec2<f32>(1.0, 7.5)), text_arcDist(p, 2.5, 7.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.5, 9.0), vec2<f32>(3.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(3.0, 6.0)));
}

fn gsl_remap_64(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(2.5, -3.0)), text_arcDist(p, 2.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, -1.5), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.5, 6.0))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_65(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_66(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(0.0, 8.5), vec2<f32>(0.0, 9.0)));
}

fn gsl_remap_67(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(0.5, -3.0)), text_arcDist(p, 0.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, -1.5), vec2<f32>(2.0, 6.0))), text_segmentDist(p, vec2<f32>(2.0, 8.5), vec2<f32>(2.0, 9.0)));
}

fn gsl_remap_68(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 3.5), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(1.321, 4.326), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_69(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854));
}

fn gsl_remap_70(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(4.5, 6.0))), text_arcDist(p, 4.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(6.0, 4.5), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(3.0, 6.0), vec2<f32>(3.0, 0.0)));
}

fn gsl_remap_71(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_72(p: vec3<f32>) -> f32 {
  return min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0)));
}

fn gsl_remap_73(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(2.5, 0.0)), text_arcDist(p, 2.5, 1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 1.5), vec2<f32>(4.0, 4.5))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.5, 6.0), vec2<f32>(0.0, 6.0))), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(0.0, -3.0)));
}

fn gsl_remap_74(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0)), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 1.5), vec2<f32>(0.0, 4.5))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 6.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(4.0, -3.0)));
}

fn gsl_remap_75(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.0, 6.0))), text_arcDist(p, 2.0, 5.0, 1.0, 0.7854, 0.7854));
}

fn gsl_remap_76(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_arcDist(p, 2.1639, 1.8206, 4.0574, 1.4957, 0.3944), text_arcDist(p, 1.2673, 4.534, 1.1997, 3.108, 1.2174)), text_segmentDist(p, vec2<f32>(0.8146, 3.423), vec2<f32>(3.185, 2.455))), text_arcDist(p, 2.7317, 1.3439, 1.2, -0.0338, 1.2173)), text_arcDist(p, 1.8361, 4.0559, 4.056, -1.646, 0.3944));
}

fn gsl_remap_77(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(3.0, 6.0)), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(1.0, 0.0)));
}

fn gsl_remap_78(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(0.0, 1.5)), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(4.0, 6.0)));
}

fn gsl_remap_79(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.0, 0.0)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 6.0)));
}

fn gsl_remap_80(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(1.5, 0.0)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(3.0, 4.0))), text_segmentDist(p, vec2<f32>(3.0, 4.0), vec2<f32>(4.5, 0.0))), text_segmentDist(p, vec2<f32>(4.5, 0.0), vec2<f32>(6.0, 6.0)));
}

fn gsl_remap_81(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(4.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 6.0)));
}

fn gsl_remap_82(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.0, 0.0)), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.228, -2.316))), text_arcDist(p, 0.2795, -2.0003, 0.9997, -0.9461, 0.6248)), text_segmentDist(p, vec2<f32>(0.2793, -3.0), vec2<f32>(0.0, -3.0)));
}

fn gsl_remap_83(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(4.0, 6.0)), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0)));
}

fn gsl_remap_84(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5))), text_segmentDist(p, vec2<f32>(2.0, 11.0), vec2<f32>(4.0, 10.0)));
}

fn gsl_remap_85(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(4.0, 11.0)));
}

fn gsl_remap_86(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(3.0, 11.0))), text_segmentDist(p, vec2<f32>(3.0, 11.0), vec2<f32>(4.0, 10.0)));
}

fn gsl_remap_87(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5))), text_arcDist(p, 2.3, 10.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(2.6295, 10.9015), vec2<f32>(3.4705, 10.3485))), text_arcDist(p, 3.7999, 10.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_88(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5))), text_segmentDist(p, vec2<f32>(5.0, 10.25), vec2<f32>(5.0, 10.75))), text_segmentDist(p, vec2<f32>(1.0, 10.25), vec2<f32>(1.0, 10.75)));
}

fn gsl_remap_89(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 9.0)), text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(4.0, 9.0))), text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(8.0, 9.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(8.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 5.0), vec2<f32>(7.0, 5.0))), text_segmentDist(p, vec2<f32>(1.1111, 2.5), vec2<f32>(4.0, 2.5)));
}

fn gsl_remap_90(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(2.0, -1.0))), text_segmentDist(p, vec2<f32>(2.0, -1.0), vec2<f32>(2.5925, -1.2725))), text_arcDist(p, 2.3658, -1.8281, 0.6, -0.0338, 1.2173)), text_arcDist(p, 1.9181, -0.472, 2.0281, -1.6461, 0.3944));
}

fn gsl_remap_91(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_segmentDist(p, vec2<f32>(1.0, 11.0), vec2<f32>(3.0, 10.0)));
}

fn gsl_remap_92(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_segmentDist(p, vec2<f32>(1.0, 10.0), vec2<f32>(3.0, 11.0)));
}

fn gsl_remap_93(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_segmentDist(p, vec2<f32>(1.0, 10.0), vec2<f32>(2.0, 11.0))), text_segmentDist(p, vec2<f32>(2.0, 11.0), vec2<f32>(3.0, 10.0)));
}

fn gsl_remap_94(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_segmentDist(p, vec2<f32>(1.0, 10.25), vec2<f32>(1.0, 10.75))), text_segmentDist(p, vec2<f32>(3.0, 10.25), vec2<f32>(3.0, 10.75)));
}

fn gsl_remap_95(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(-1.0, 11.0), vec2<f32>(1.0, 10.0)));
}

fn gsl_remap_96(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(-1.0, 10.0), vec2<f32>(1.0, 11.0)));
}

fn gsl_remap_97(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(-1.0, 10.0), vec2<f32>(0.0, 11.0))), text_segmentDist(p, vec2<f32>(0.0, 11.0), vec2<f32>(1.0, 10.0)));
}

fn gsl_remap_98(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(-1.0, 10.25), vec2<f32>(-1.0, 10.75))), text_segmentDist(p, vec2<f32>(1.0, 10.25), vec2<f32>(1.0, 10.75)));
}

fn gsl_remap_99(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854), text_segmentDist(p, vec2<f32>(5.0, 2.0), vec2<f32>(5.0, 7.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(0.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(-1.0, 4.5), vec2<f32>(1.0, 4.5)));
}

fn gsl_remap_100(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 9.0))), text_arcDist(p, 1.8, 10.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(2.1295, 10.9015), vec2<f32>(2.9705, 10.3485))), text_arcDist(p, 3.2999, 10.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_101(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 2.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 0.0), vec2<f32>(2.0, 0.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 2.0), vec2<f32>(0.0, 7.0))), text_segmentDist(p, vec2<f32>(1.5, 11.0), vec2<f32>(3.5, 10.0)));
}

fn gsl_remap_102(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 2.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 0.0), vec2<f32>(2.0, 0.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 2.0), vec2<f32>(0.0, 7.0))), text_segmentDist(p, vec2<f32>(1.5, 10.0), vec2<f32>(3.5, 11.0)));
}

fn gsl_remap_103(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(min(text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 2.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 0.0), vec2<f32>(2.0, 0.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 2.0), vec2<f32>(0.0, 7.0))), text_segmentDist(p, vec2<f32>(1.5, 10.0), vec2<f32>(2.5, 11.0))), text_segmentDist(p, vec2<f32>(2.5, 11.0), vec2<f32>(3.5, 10.0)));
}

fn gsl_remap_104(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(min(min(text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 9.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 7.0), vec2<f32>(5.0, 2.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 0.0), vec2<f32>(2.0, 0.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 2.0), vec2<f32>(0.0, 7.0))), text_arcDist(p, 1.8, 10.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(2.1295, 10.9015), vec2<f32>(2.9705, 10.3485))), text_arcDist(p, 3.2999, 10.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_105(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0)), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(3.0, 0.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 2.0), vec2<f32>(5.0, 7.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(2.0, 9.0))), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.0, 10.25), vec2<f32>(1.0, 10.75))), text_segmentDist(p, vec2<f32>(4.0, 10.25), vec2<f32>(4.0, 10.75)));
}

fn gsl_remap_106(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0)), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(3.0, 0.0))), text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 2.0), vec2<f32>(5.0, 7.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(2.0, 9.0))), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(5.0, 9.0)));
}

fn gsl_remap_107(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 2.5)), text_arcDist(p, 2.5, 2.5, 2.5, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(5.0, 2.5), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(1.5, 11.0), vec2<f32>(3.5, 10.0)));
}

fn gsl_remap_108(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 2.5)), text_arcDist(p, 2.5, 2.5, 2.5, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(5.0, 2.5), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(1.5, 10.0), vec2<f32>(3.5, 11.0)));
}

fn gsl_remap_109(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 2.5)), text_arcDist(p, 2.5, 2.5, 2.5, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(5.0, 2.5), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(1.5, 10.0), vec2<f32>(2.5, 11.0))), text_segmentDist(p, vec2<f32>(2.5, 11.0), vec2<f32>(3.5, 10.0)));
}

fn gsl_remap_110(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 2.5)), text_arcDist(p, 2.5, 2.5, 2.5, -1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(5.0, 2.5), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(1.0, 10.25), vec2<f32>(1.0, 10.75))), text_segmentDist(p, vec2<f32>(4.0, 10.25), vec2<f32>(4.0, 10.75)));
}

fn gsl_remap_111(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(3.0, 5.0)), text_segmentDist(p, vec2<f32>(3.0, 5.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(3.0, 5.0), vec2<f32>(6.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(4.0, 11.0)));
}

fn gsl_remap_112(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 6.0))), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.0, 6.0))), text_arcDist(p, 2.0, 4.5, 1.5, 0.0, 1.5708)), text_segmentDist(p, vec2<f32>(2.0, 3.0), vec2<f32>(0.0, 3.0)));
}

fn gsl_remap_113(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_arcDist(p, 2.5, 7.0, 2.0, 0.0, 1.5708), text_segmentDist(p, vec2<f32>(2.5, 9.0), vec2<f32>(2.0, 9.0))), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 0.0), vec2<f32>(2.6, 0.0))), text_arcDist(p, 2.5999, 2.4001, 2.4001, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(5.0, 2.4), vec2<f32>(5.0, 2.6))), text_arcDist(p, 2.5999, 2.5999, 2.4001, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.6, 5.0), vec2<f32>(1.0, 5.0)));
}

fn gsl_remap_114(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_115(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_116(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(2.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_117(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_arcDist(p, 1.3, 8.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(1.6295, 8.9015), vec2<f32>(2.4705, 8.3485))), text_arcDist(p, 2.7999, 8.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_118(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(1.0, 8.5))), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(3.0, 8.5)));
}

fn gsl_remap_119(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(8.0, 3.0))), text_segmentDist(p, vec2<f32>(8.0, 3.0), vec2<f32>(8.0, 4.0))), text_arcDist(p, 6.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(4.0, 4.0), vec2<f32>(4.0, 1.5))), text_arcDist(p, 5.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(5.5, 0.0), vec2<f32>(8.0, 0.0)));
}

fn gsl_remap_120(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(3.0, 6.0), vec2<f32>(1.5, 6.0)), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(1.5, -1.0))), text_segmentDist(p, vec2<f32>(1.5, -1.0), vec2<f32>(2.0925, -1.2725))), text_arcDist(p, 1.8658, -1.8281, 0.6, -0.0338, 1.2173)), text_arcDist(p, 1.4181, -0.472, 2.0281, -1.6461, 0.3944));
}

fn gsl_remap_121(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_122(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_123(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(2.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_124(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(1.0, 8.5))), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(3.0, 8.5)));
}

fn gsl_remap_125(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(-1.0, 9.0)));
}

fn gsl_remap_126(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(-1.0, 8.0)));
}

fn gsl_remap_127(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(0.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(-1.0, 8.0)));
}

fn gsl_remap_128(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(-1.0, 8.5), vec2<f32>(-1.0, 9.0))), text_segmentDist(p, vec2<f32>(1.0, 8.5), vec2<f32>(1.0, 9.0)));
}

fn gsl_remap_129(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))), text_segmentDist(p, vec2<f32>(3.6502, 5.13), vec2<f32>(1.0, 9.0))), text_segmentDist(p, vec2<f32>(1.0, 7.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_130(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_arcDist(p, 1.1, 8.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(1.4295, 8.9015), vec2<f32>(2.2705, 8.3485))), text_arcDist(p, 2.5999, 8.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_131(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_132(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_133(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(2.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_134(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 1.3, 8.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(1.6295, 8.9015), vec2<f32>(2.4705, 8.3485))), text_arcDist(p, 2.7999, 8.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_135(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_arcDist(p, 2.0, 2.0, 2.0, -1.5708, 1.5708), text_segmentDist(p, vec2<f32>(4.0, 2.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 2.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(1.0, 8.5))), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(3.0, 8.5)));
}

fn gsl_remap_136(p: vec3<f32>) -> f32 {
  return min(min(text_arcDist(p, 3.0, 4.5, 3.0, 6.0648, 1.5705), text_arcDist(p, 3.0, 4.5, 3.0, 2.9232, 1.5711)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 9.0)));
}

fn gsl_remap_137(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(0.0, 1.5)), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_138(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(0.0, 1.5)), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_139(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(0.0, 1.5)), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(2.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_140(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(0.0, 1.5)), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(1.0, 8.5))), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(3.0, 8.5)));
}

fn gsl_remap_141(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.0, 0.0)), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.228, -2.316))), text_arcDist(p, 0.2795, -2.0003, 0.9997, -0.9461, 0.6248)), text_segmentDist(p, vec2<f32>(0.2793, -3.0), vec2<f32>(0.0, -3.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_142(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(2.5, 0.0)), text_arcDist(p, 2.5, 1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 1.5), vec2<f32>(4.0, 4.5))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.5, 6.0), vec2<f32>(0.0, 6.0))), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, -3.0)));
}

fn gsl_remap_143(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.0, 0.0)), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.228, -2.316))), text_arcDist(p, 0.2795, -2.0003, 0.9997, -0.9461, 0.6248)), text_segmentDist(p, vec2<f32>(0.2793, -3.0), vec2<f32>(0.0, -3.0))), text_segmentDist(p, vec2<f32>(1.0, 8.5), vec2<f32>(1.0, 9.0))), text_segmentDist(p, vec2<f32>(3.0, 8.5), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_144(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5))), text_segmentDist(p, vec2<f32>(1.5, 11.0), vec2<f32>(4.5, 11.0)));
}

fn gsl_remap_145(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5))), text_arcDist(p, 3.0, 11.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_146(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 9.0)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(6.0, 0.0))), text_arcDist(p, 6.0, -1.0, 1.0, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(0.8333, 2.5), vec2<f32>(5.167, 2.5)));
}

fn gsl_remap_147(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 10.0), vec2<f32>(3.0, 11.0)));
}

fn gsl_remap_148(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 10.0), vec2<f32>(2.0, 11.0))), text_segmentDist(p, vec2<f32>(2.0, 11.0), vec2<f32>(3.0, 10.0)));
}

fn gsl_remap_149(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 11.0), vec2<f32>(2.0, 10.5)));
}

fn gsl_remap_150(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 11.0), vec2<f32>(2.0, 10.0))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(3.0, 11.0)));
}

fn gsl_remap_151(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_arcDist(p, 3.0, 2.0, 2.0, -0.7854, 0.7854), text_segmentDist(p, vec2<f32>(5.0, 2.0), vec2<f32>(5.0, 7.0))), text_arcDist(p, 3.0, 7.0, 2.0, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(0.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 11.0), vec2<f32>(2.0, 10.0))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(3.0, 11.0)));
}

fn gsl_remap_152(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_segmentDist(p, vec2<f32>(0.5, 11.0), vec2<f32>(3.5, 11.0)));
}

fn gsl_remap_153(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_arcDist(p, 2.0, 11.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_154(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_segmentDist(p, vec2<f32>(2.0, 11.0), vec2<f32>(2.0, 10.5)));
}

fn gsl_remap_155(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_arcDist(p, 4.0, -1.0, 1.0, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0)));
}

fn gsl_remap_156(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 9.0)), text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(3.0, 5.0))), text_segmentDist(p, vec2<f32>(1.0, 11.0), vec2<f32>(2.0, 10.0))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(3.0, 11.0)));
}

fn gsl_remap_157(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(5.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 5.0))), text_segmentDist(p, vec2<f32>(5.0, 5.0), vec2<f32>(3.5, 5.0))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(3.0, 11.0))), text_segmentDist(p, vec2<f32>(3.0, 11.0), vec2<f32>(4.0, 10.0)));
}

fn gsl_remap_158(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(5.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 5.0))), text_segmentDist(p, vec2<f32>(5.0, 5.0), vec2<f32>(3.5, 5.0))), text_arcDist(p, 3.0, 11.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_159(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(5.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 5.0))), text_segmentDist(p, vec2<f32>(5.0, 5.0), vec2<f32>(3.5, 5.0))), text_segmentDist(p, vec2<f32>(2.5, 11.0), vec2<f32>(2.5, 10.5)));
}

fn gsl_remap_160(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(5.0, 9.0), vec2<f32>(2.0, 9.0)), text_arcDist(p, 2.0, 7.0, 2.0, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 7.0), vec2<f32>(0.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 5.0))), text_segmentDist(p, vec2<f32>(5.0, 5.0), vec2<f32>(3.5, 5.0))), text_segmentDist(p, vec2<f32>(2.0, -3.0), vec2<f32>(3.0, -1.0))), text_segmentDist(p, vec2<f32>(3.0, -1.0), vec2<f32>(3.0, -0.5)));
}

fn gsl_remap_161(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(5.0, 5.0))), text_segmentDist(p, vec2<f32>(1.5, 10.0), vec2<f32>(2.5, 11.0))), text_segmentDist(p, vec2<f32>(2.5, 11.0), vec2<f32>(3.5, 10.0)));
}

fn gsl_remap_162(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(5.0, 0.0), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, 5.0), vec2<f32>(5.0, 5.0))), text_segmentDist(p, vec2<f32>(-1.0, 7.0), vec2<f32>(6.0, 7.0)));
}

fn gsl_remap_163(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_arcDist(p, -0.7, 10.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(-0.3705, 10.9015), vec2<f32>(0.4705, 10.3485))), text_arcDist(p, 0.7999, 10.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_164(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(-1.5, 11.0), vec2<f32>(1.5, 11.0)));
}

fn gsl_remap_165(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_arcDist(p, 0.0, 11.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_166(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_arcDist(p, 0.0, -1.0, 1.0, 3.1416, 1.5708));
}

fn gsl_remap_167(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 11.0), vec2<f32>(0.0, 10.5)));
}

fn gsl_remap_168(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(4.0, 2.0))), text_arcDist(p, 2.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, 0.0), vec2<f32>(1.0, 0.0)));
}

fn gsl_remap_169(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(3.0, 9.0), vec2<f32>(3.0, 2.0)), text_arcDist(p, 1.0, 2.0, 2.0, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(1.0, 0.0), vec2<f32>(0.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 10.0), vec2<f32>(3.0, 11.0))), text_segmentDist(p, vec2<f32>(3.0, 11.0), vec2<f32>(4.0, 10.0)));
}

fn gsl_remap_170(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 3.5), vec2<f32>(5.0, 9.0))), text_segmentDist(p, vec2<f32>(1.671, 5.338), vec2<f32>(5.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, -3.0), vec2<f32>(3.0, -1.0))), text_segmentDist(p, vec2<f32>(3.0, -1.0), vec2<f32>(3.0, -0.5)));
}

fn gsl_remap_171(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.0, 10.0), vec2<f32>(2.0, 11.0)));
}

fn gsl_remap_172(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, -3.0), vec2<f32>(2.0, -1.0))), text_segmentDist(p, vec2<f32>(2.0, -1.0), vec2<f32>(2.0, -0.5)));
}

fn gsl_remap_173(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_segmentDist(p, vec2<f32>(0.5, 8.0), vec2<f32>(3.5, 8.0)));
}

fn gsl_remap_174(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_arcDist(p, 2.0, 9.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_175(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.5, 6.0), vec2<f32>(2.5, 6.0)), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, 3.1416, 1.5708)), text_segmentDist(p, vec2<f32>(1.5, 3.0), vec2<f32>(4.0, 3.0))), text_arcDist(p, 4.0, -1.0, 1.0, 3.1416, 1.5708));
}

fn gsl_remap_176(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(3.0, 6.0), vec2<f32>(1.5, 6.0)), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_177(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(3.0, 6.0), vec2<f32>(1.5, 6.0)), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(2.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_178(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(3.0, 6.0), vec2<f32>(1.5, 6.0)), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(2.0, 8.5)));
}

fn gsl_remap_179(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(3.0, 6.0), vec2<f32>(1.5, 6.0)), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(3.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(2.0, 8.0))), text_segmentDist(p, vec2<f32>(2.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_180(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(4.0, 0.0)), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 1.5), vec2<f32>(0.0, 4.5))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 6.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.5, 7.0), vec2<f32>(5.0, 9.0)));
}

fn gsl_remap_181(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(4.0, 9.0), vec2<f32>(4.0, 0.0)), text_segmentDist(p, vec2<f32>(4.0, 0.0), vec2<f32>(1.5, 0.0))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 1.5), vec2<f32>(0.0, 4.5))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 6.0), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(5.0, 8.0), vec2<f32>(1.0, 8.0)));
}

fn gsl_remap_182(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(0.5, 8.0), vec2<f32>(3.5, 8.0)));
}

fn gsl_remap_183(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_arcDist(p, 2.0, 9.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_184(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(2.0, 8.5)));
}

fn gsl_remap_185(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_arcDist(p, 4.0, -1.0, 1.0, 3.1416, 1.5708));
}

fn gsl_remap_186(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 3.0), vec2<f32>(4.0, 3.0)), text_segmentDist(p, vec2<f32>(4.0, 3.0), vec2<f32>(4.0, 4.0))), text_arcDist(p, 2.0, 4.0, 2.0, 1.5708, 1.5708)), text_segmentDist(p, vec2<f32>(0.0, 4.0), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 9.0), vec2<f32>(2.0, 8.0))), text_segmentDist(p, vec2<f32>(2.0, 8.0), vec2<f32>(3.0, 9.0)));
}

fn gsl_remap_187(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(2.5, -3.0)), text_arcDist(p, 2.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, -1.5), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.5, 6.0))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(2.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_188(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(2.5, -3.0)), text_arcDist(p, 2.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, -1.5), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.5, 6.0))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_arcDist(p, 2.0, 9.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_189(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(2.5, -3.0)), text_arcDist(p, 2.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, -1.5), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.5, 6.0))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(2.0, 8.5)));
}

fn gsl_remap_190(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(2.5, -3.0)), text_arcDist(p, 2.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, -1.5), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(4.0, 6.0), vec2<f32>(1.5, 6.0))), text_arcDist(p, 1.5, 4.5, 1.5, 2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 4.5), vec2<f32>(0.0, 1.5))), text_arcDist(p, 1.5, 1.5, 1.5, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(1.5, 0.0), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(2.0, 7.0), vec2<f32>(2.5, 9.0)));
}

fn gsl_remap_191(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, 10.0), vec2<f32>(2.0, 11.0))), text_segmentDist(p, vec2<f32>(2.0, 11.0), vec2<f32>(3.0, 10.0)));
}

fn gsl_remap_192(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 6.0), vec2<f32>(2.5, 6.0))), text_arcDist(p, 2.5, 4.5, 1.5, 0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(4.0, 4.5), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(-1.0, 8.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_193(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_arcDist(p, -0.7, 8.4001, 0.6, 2.0656, 1.0761)), text_segmentDist(p, vec2<f32>(-0.3705, 8.9015), vec2<f32>(0.4705, 8.3485))), text_arcDist(p, 0.7999, 8.85, 0.6001, -1.076, 1.076));
}

fn gsl_remap_194(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(-1.5, 9.0), vec2<f32>(1.5, 9.0)));
}

fn gsl_remap_195(p: vec3<f32>) -> f32 {
  return min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_arcDist(p, 0.0, 9.0, 1.0, 4.7124, 1.5708));
}

fn gsl_remap_196(p: vec3<f32>) -> f32 {
  return min(min(text_arcDist(p, 0.0, -1.0, 1.0, 3.1416, 1.5708), text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0))), text_segmentDist(p, vec2<f32>(0.0, 8.5), vec2<f32>(0.0, 9.0)));
}

fn gsl_remap_197(p: vec3<f32>) -> f32 {
  return min(min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0)), text_segmentDist(p, vec2<f32>(0.0, 8.5), vec2<f32>(0.0, 9.0))), text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(0.5, -3.0))), text_arcDist(p, 0.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, -1.5), vec2<f32>(2.0, 6.0))), text_segmentDist(p, vec2<f32>(2.0, 8.5), vec2<f32>(2.0, 9.0)));
}

fn gsl_remap_198(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(0.5, -3.0)), text_arcDist(p, 0.5, -1.5, 1.5, -0.7854, 0.7854)), text_segmentDist(p, vec2<f32>(2.0, -1.5), vec2<f32>(2.0, 6.0))), text_segmentDist(p, vec2<f32>(1.0, 8.0), vec2<f32>(2.0, 9.0))), text_segmentDist(p, vec2<f32>(2.0, 9.0), vec2<f32>(3.0, 8.0)));
}

fn gsl_remap_199(p: vec3<f32>) -> f32 {
  return min(min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0)), text_segmentDist(p, vec2<f32>(0.0, 3.5), vec2<f32>(4.0, 6.0))), text_segmentDist(p, vec2<f32>(1.321, 4.326), vec2<f32>(4.0, 0.0))), text_segmentDist(p, vec2<f32>(1.0, -3.0), vec2<f32>(2.0, -1.0))), text_segmentDist(p, vec2<f32>(2.0, -1.0), vec2<f32>(2.0, -0.5)));
}

fn gsl_remap_200(p: vec3<f32>) -> f32 {
  return min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, 10.0), vec2<f32>(2.0, 11.0)));
}

fn gsl_remap_201(p: vec3<f32>) -> f32 {
  return min(min(min(text_segmentDist(p, vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 1.0)), text_arcDist(p, 1.0, 1.0, 1.0, -2.3562, 0.7854)), text_segmentDist(p, vec2<f32>(0.0, -3.0), vec2<f32>(1.0, -1.0))), text_segmentDist(p, vec2<f32>(1.0, -1.0), vec2<f32>(1.0, -0.5)));
}

fn text_arcDist(p: vec3<f32>, cx: f32, cy: f32, ra: f32, mid: f32, half: f32) -> f32 {
  let dx = p.x - cx;
  let rot = 1.5707964 - mid;
  let ca = cos(rot);
  let dy = p.y - cy;
  let sa = sin(rot);
  let rx = dx * ca - dy * sa;
  let pxAbs = abs(rx);
  let ry = dx * sa + dy * ca;
  let l = length(vec2<f32>(pxAbs, ry));
  let dOutside = abs(l - ra);
  let scx = sin(half);
  let wx = pxAbs - ra * scx;
  let scy = cos(half);
  let wy = ry - ra * scy;
  let dInside = length(vec2<f32>(wx, wy));
  let inside = scy * pxAbs > scx * ry;
  return select(dOutside, dInside, inside);
}

fn text_segmentDist(p: vec3<f32>, a: vec2<f32>, b: vec2<f32>) -> f32 {
  let pa = p.xy - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = vec2<f32>(6.6999, min(1000000.0, gsl_remap(vec3<f32>((p.x - -1.0694444) / 0.027777778 - 0.0, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_2 = vec2<f32>(gsl_tmp_1.x + 4.7, min(gsl_tmp_1.y, gsl_remap_2(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_1.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_3 = vec2<f32>(gsl_tmp_2.x + 6.7, min(gsl_tmp_2.y, gsl_remap_3(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_2.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_4 = vec2<f32>(gsl_tmp_3.x + 6.7, min(gsl_tmp_3.y, gsl_remap_4(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_3.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_5 = vec2<f32>(gsl_tmp_4.x + 7.7, min(gsl_tmp_4.y, gsl_remap_5(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_4.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_6 = vec2<f32>(gsl_tmp_5.x + 6.7, min(gsl_tmp_5.y, gsl_remap_6(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_5.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_7 = vec2<f32>(gsl_tmp_6.x + 6.7, min(gsl_tmp_6.y, gsl_remap_7(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_6.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_8 = vec2<f32>(gsl_tmp_7.x + 6.7, min(gsl_tmp_7.y, gsl_remap_8(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_7.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_9 = vec2<f32>(gsl_tmp_8.x + 6.7, min(gsl_tmp_8.y, gsl_remap_9(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_8.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))));
  let l01 = min(gsl_tmp_9.y, gsl_remap_10(vec3<f32>((p.x - -1.0694444) / 0.027777778 - gsl_tmp_9.x, (p.y - 3.125) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let gsl_tmp_10 = vec2<f32>(2.7, min(1000000.0, gsl_remap_11(vec3<f32>((p.x - -3.4222221) / 0.027777778 - 0.0, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_11 = vec2<f32>(gsl_tmp_10.x + 6.7, min(gsl_tmp_10.y, gsl_remap_12(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_10.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_12 = vec2<f32>(gsl_tmp_11.x + 9.7, min(gsl_tmp_11.y, gsl_remap_13(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_11.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_13 = vec2<f32>(gsl_tmp_12.x + 7.7005, min(gsl_tmp_12.y, gsl_remap_14(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_12.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_14 = vec2<f32>(gsl_tmp_13.x + 9.7, min(gsl_tmp_13.y, gsl_remap_15(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_13.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_15 = vec2<f32>(gsl_tmp_14.x + 8.7, min(gsl_tmp_14.y, gsl_remap_16(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_14.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_16 = vec2<f32>(gsl_tmp_15.x + 3.7, min(gsl_tmp_15.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_15.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.5, 7.0), vec2<f32>(1.0, 9.0))));
  let gsl_tmp_17 = vec2<f32>(gsl_tmp_16.x + 3.6983, min(gsl_tmp_16.y, text_arcDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_16.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), 13.0, 4.0, 13.0, 3.1416, 0.3944)));
  let gsl_tmp_18 = vec2<f32>(gsl_tmp_17.x + 2.7017, min(gsl_tmp_17.y, text_arcDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_17.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), -12.0, 4.0, 13.0, 6.2834, 0.3946)));
  let gsl_tmp_19 = vec2<f32>(gsl_tmp_18.x + 6.7, min(gsl_tmp_18.y, gsl_remap_17(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_18.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_20 = vec2<f32>(gsl_tmp_19.x + 6.7, min(gsl_tmp_19.y, gsl_remap_18(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_19.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_21 = vec2<f32>(gsl_tmp_20.x + 3.7, min(gsl_tmp_20.y, gsl_remap_19(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_20.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_22 = vec2<f32>(gsl_tmp_21.x + 6.7, min(gsl_tmp_21.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_21.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.0, 4.0), vec2<f32>(4.0, 4.0))));
  let gsl_tmp_23 = vec2<f32>(gsl_tmp_22.x + 2.7, min(gsl_tmp_22.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_22.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 0.5))));
  let gsl_tmp_24 = vec2<f32>(gsl_tmp_23.x + 6.7, min(gsl_tmp_23.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_23.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(4.0, 9.0), vec2<f32>(0.0, 0.0))));
  let gsl_tmp_25 = vec2<f32>(gsl_tmp_24.x + 2.7, min(gsl_tmp_24.y, gsl_remap_20(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_24.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_26 = vec2<f32>(gsl_tmp_25.x + 3.7, min(gsl_tmp_25.y, gsl_remap_21(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_25.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_27 = vec2<f32>(gsl_tmp_26.x + 6.7, min(gsl_tmp_26.y, gsl_remap_22(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_26.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_28 = vec2<f32>(gsl_tmp_27.x + 6.7, min(gsl_tmp_27.y, gsl_remap_23(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_27.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_29 = vec2<f32>(gsl_tmp_28.x + 6.7, min(gsl_tmp_28.y, gsl_remap_24(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_28.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_30 = vec2<f32>(gsl_tmp_29.x + 6.6992, min(gsl_tmp_29.y, gsl_remap_25(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_29.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_31 = vec2<f32>(gsl_tmp_30.x + 11.2272, min(gsl_tmp_30.y, gsl_remap_26(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_30.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_32 = vec2<f32>(gsl_tmp_31.x + 3.7, min(gsl_tmp_31.y, gsl_remap_27(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_31.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_33 = vec2<f32>(gsl_tmp_32.x + 6.7, min(gsl_tmp_32.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_32.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.0, 9.0), vec2<f32>(4.0, 0.0))));
  let gsl_tmp_34 = vec2<f32>(gsl_tmp_33.x + 3.7, min(gsl_tmp_33.y, gsl_remap_28(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_33.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_35 = vec2<f32>(gsl_tmp_34.x + 6.7, min(gsl_tmp_34.y, gsl_remap_29(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_34.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_36 = vec2<f32>(gsl_tmp_35.x + 6.7, min(gsl_tmp_35.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_35.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.0, -1.0), vec2<f32>(4.0, -1.0))));
  let gsl_tmp_37 = vec2<f32>(gsl_tmp_36.x + 5.7, min(gsl_tmp_36.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_36.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(1.0, 9.0), vec2<f32>(3.0, 8.0))));
  let gsl_tmp_38 = vec2<f32>(gsl_tmp_37.x + 4.7, min(gsl_tmp_37.y, gsl_remap_30(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_37.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_39 = vec2<f32>(gsl_tmp_38.x + 2.7, min(gsl_tmp_38.y, text_segmentDist(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_38.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 9.0))));
  let gsl_tmp_40 = vec2<f32>(gsl_tmp_39.x + 4.7, min(gsl_tmp_39.y, gsl_remap_31(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_39.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))));
  let l02 = min(gsl_tmp_40.y, gsl_remap_32(vec3<f32>((p.x - -3.4222221) / 0.027777778 - gsl_tmp_40.x, (p.y - 2.625) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let gsl_tmp_41 = vec2<f32>(8.7, min(1000000.0, gsl_remap_33(vec3<f32>((p.x - -2.7805555) / 0.027777778 - 0.0, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_42 = vec2<f32>(gsl_tmp_41.x + 7.7, min(gsl_tmp_41.y, gsl_remap_34(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_41.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_43 = vec2<f32>(gsl_tmp_42.x + 6.7, min(gsl_tmp_42.y, gsl_remap_35(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_42.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_44 = vec2<f32>(gsl_tmp_43.x + 7.7, min(gsl_tmp_43.y, gsl_remap_36(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_43.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_45 = vec2<f32>(gsl_tmp_44.x + 6.7, min(gsl_tmp_44.y, gsl_remap_37(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_44.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_46 = vec2<f32>(gsl_tmp_45.x + 6.7, min(gsl_tmp_45.y, gsl_remap_38(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_45.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_47 = vec2<f32>(gsl_tmp_46.x + 7.7, min(gsl_tmp_46.y, gsl_remap_39(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_46.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_48 = vec2<f32>(gsl_tmp_47.x + 7.7, min(gsl_tmp_47.y, gsl_remap_40(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_47.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_49 = vec2<f32>(gsl_tmp_48.x + 2.7, min(gsl_tmp_48.y, text_segmentDist(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_48.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.0, 9.0), vec2<f32>(0.0, 0.0))));
  let gsl_tmp_50 = vec2<f32>(gsl_tmp_49.x + 5.7, min(gsl_tmp_49.y, gsl_remap_41(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_49.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_51 = vec2<f32>(gsl_tmp_50.x + 7.7, min(gsl_tmp_50.y, gsl_remap_42(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_50.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_52 = vec2<f32>(gsl_tmp_51.x + 6.7, min(gsl_tmp_51.y, gsl_remap_43(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_51.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_53 = vec2<f32>(gsl_tmp_52.x + 8.7, min(gsl_tmp_52.y, gsl_remap_44(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_52.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_54 = vec2<f32>(gsl_tmp_53.x + 7.7, min(gsl_tmp_53.y, gsl_remap_45(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_53.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_55 = vec2<f32>(gsl_tmp_54.x + 7.7, min(gsl_tmp_54.y, gsl_remap_46(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_54.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_56 = vec2<f32>(gsl_tmp_55.x + 7.7, min(gsl_tmp_55.y, gsl_remap_47(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_55.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_57 = vec2<f32>(gsl_tmp_56.x + 8.7, min(gsl_tmp_56.y, gsl_remap_48(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_56.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_58 = vec2<f32>(gsl_tmp_57.x + 7.7, min(gsl_tmp_57.y, gsl_remap_49(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_57.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_59 = vec2<f32>(gsl_tmp_58.x + 7.7005, min(gsl_tmp_58.y, gsl_remap_50(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_58.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_60 = vec2<f32>(gsl_tmp_59.x + 8.7, min(gsl_tmp_59.y, gsl_remap_51(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_59.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_61 = vec2<f32>(gsl_tmp_60.x + 7.7, min(gsl_tmp_60.y, gsl_remap_52(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_60.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_62 = vec2<f32>(gsl_tmp_61.x + 8.7, min(gsl_tmp_61.y, gsl_remap_53(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_61.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_63 = vec2<f32>(gsl_tmp_62.x + 10.7, min(gsl_tmp_62.y, gsl_remap_54(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_62.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_64 = vec2<f32>(gsl_tmp_63.x + 8.7, min(gsl_tmp_63.y, gsl_remap_55(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_63.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_65 = vec2<f32>(gsl_tmp_64.x + 8.7, min(gsl_tmp_64.y, gsl_remap_56(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_64.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))));
  let l03 = min(gsl_tmp_65.y, gsl_remap_57(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_65.x, (p.y - 2.125) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let gsl_tmp_66 = vec2<f32>(6.7, min(1000000.0, gsl_remap_58(vec3<f32>((p.x - -2.7805555) / 0.027777778 - 0.0, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_67 = vec2<f32>(gsl_tmp_66.x + 6.7, min(gsl_tmp_66.y, gsl_remap_59(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_66.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_68 = vec2<f32>(gsl_tmp_67.x + 5.7, min(gsl_tmp_67.y, gsl_remap_60(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_67.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_69 = vec2<f32>(gsl_tmp_68.x + 6.7, min(gsl_tmp_68.y, gsl_remap_61(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_68.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_70 = vec2<f32>(gsl_tmp_69.x + 6.7, min(gsl_tmp_69.y, gsl_remap_62(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_69.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_71 = vec2<f32>(gsl_tmp_70.x + 5.7, min(gsl_tmp_70.y, gsl_remap_63(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_70.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_72 = vec2<f32>(gsl_tmp_71.x + 6.7, min(gsl_tmp_71.y, gsl_remap_64(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_71.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_73 = vec2<f32>(gsl_tmp_72.x + 6.7, min(gsl_tmp_72.y, gsl_remap_65(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_72.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_74 = vec2<f32>(gsl_tmp_73.x + 2.7, min(gsl_tmp_73.y, gsl_remap_66(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_73.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_75 = vec2<f32>(gsl_tmp_74.x + 4.7, min(gsl_tmp_74.y, gsl_remap_67(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_74.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_76 = vec2<f32>(gsl_tmp_75.x + 6.7, min(gsl_tmp_75.y, gsl_remap_68(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_75.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_77 = vec2<f32>(gsl_tmp_76.x + 3.7, min(gsl_tmp_76.y, gsl_remap_69(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_76.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_78 = vec2<f32>(gsl_tmp_77.x + 8.7, min(gsl_tmp_77.y, gsl_remap_70(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_77.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_79 = vec2<f32>(gsl_tmp_78.x + 6.7, min(gsl_tmp_78.y, gsl_remap_71(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_78.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_80 = vec2<f32>(gsl_tmp_79.x + 6.7, min(gsl_tmp_79.y, gsl_remap_72(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_79.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_81 = vec2<f32>(gsl_tmp_80.x + 6.7, min(gsl_tmp_80.y, gsl_remap_73(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_80.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_82 = vec2<f32>(gsl_tmp_81.x + 6.7, min(gsl_tmp_81.y, gsl_remap_74(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_81.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_83 = vec2<f32>(gsl_tmp_82.x + 5.7, min(gsl_tmp_82.y, gsl_remap_75(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_82.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_84 = vec2<f32>(gsl_tmp_83.x + 6.6995, min(gsl_tmp_83.y, gsl_remap_76(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_83.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_85 = vec2<f32>(gsl_tmp_84.x + 5.7, min(gsl_tmp_84.y, gsl_remap_77(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_84.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_86 = vec2<f32>(gsl_tmp_85.x + 6.7, min(gsl_tmp_85.y, gsl_remap_78(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_85.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_87 = vec2<f32>(gsl_tmp_86.x + 6.7, min(gsl_tmp_86.y, gsl_remap_79(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_86.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_88 = vec2<f32>(gsl_tmp_87.x + 8.7, min(gsl_tmp_87.y, gsl_remap_80(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_87.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_89 = vec2<f32>(gsl_tmp_88.x + 6.7, min(gsl_tmp_88.y, gsl_remap_81(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_88.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_90 = vec2<f32>(gsl_tmp_89.x + 6.7, min(gsl_tmp_89.y, gsl_remap_82(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_89.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))));
  let l04 = min(gsl_tmp_90.y, gsl_remap_83(vec3<f32>((p.x - -2.7805555) / 0.027777778 - gsl_tmp_90.x, (p.y - 1.625) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let gsl_tmp_91 = vec2<f32>(8.7, min(1000000.0, gsl_remap_84(vec3<f32>((p.x - -3.2083333) / 0.027777778 - 0.0, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_92 = vec2<f32>(gsl_tmp_91.x + 8.7, min(gsl_tmp_91.y, gsl_remap_85(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_91.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_93 = vec2<f32>(gsl_tmp_92.x + 8.7, min(gsl_tmp_92.y, gsl_remap_86(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_92.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_94 = vec2<f32>(gsl_tmp_93.x + 8.7, min(gsl_tmp_93.y, gsl_remap_87(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_93.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_95 = vec2<f32>(gsl_tmp_94.x + 8.7, min(gsl_tmp_94.y, gsl_remap_88(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_94.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_96 = vec2<f32>(gsl_tmp_95.x + 8.7, min(gsl_tmp_95.y, gsl_remap_33(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_95.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_97 = vec2<f32>(gsl_tmp_96.x + 10.7, min(gsl_tmp_96.y, gsl_remap_89(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_96.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_98 = vec2<f32>(gsl_tmp_97.x + 6.7, min(gsl_tmp_97.y, gsl_remap_90(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_97.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_99 = vec2<f32>(gsl_tmp_98.x + 6.7, min(gsl_tmp_98.y, gsl_remap_91(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_98.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_100 = vec2<f32>(gsl_tmp_99.x + 6.7, min(gsl_tmp_99.y, gsl_remap_92(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_99.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_101 = vec2<f32>(gsl_tmp_100.x + 6.7, min(gsl_tmp_100.y, gsl_remap_93(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_100.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_102 = vec2<f32>(gsl_tmp_101.x + 6.7, min(gsl_tmp_101.y, gsl_remap_94(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_101.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_103 = vec2<f32>(gsl_tmp_102.x + 3.7, min(gsl_tmp_102.y, gsl_remap_95(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_102.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_104 = vec2<f32>(gsl_tmp_103.x + 3.7, min(gsl_tmp_103.y, gsl_remap_96(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_103.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_105 = vec2<f32>(gsl_tmp_104.x + 3.7, min(gsl_tmp_104.y, gsl_remap_97(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_104.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_106 = vec2<f32>(gsl_tmp_105.x + 3.7, min(gsl_tmp_105.y, gsl_remap_98(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_105.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_107 = vec2<f32>(gsl_tmp_106.x + 7.7, min(gsl_tmp_106.y, gsl_remap_99(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_106.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_108 = vec2<f32>(gsl_tmp_107.x + 7.7, min(gsl_tmp_107.y, gsl_remap_100(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_107.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_109 = vec2<f32>(gsl_tmp_108.x + 7.7, min(gsl_tmp_108.y, gsl_remap_101(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_108.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_110 = vec2<f32>(gsl_tmp_109.x + 7.7, min(gsl_tmp_109.y, gsl_remap_102(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_109.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_111 = vec2<f32>(gsl_tmp_110.x + 7.7, min(gsl_tmp_110.y, gsl_remap_103(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_110.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_112 = vec2<f32>(gsl_tmp_111.x + 7.7, min(gsl_tmp_111.y, gsl_remap_104(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_111.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_113 = vec2<f32>(gsl_tmp_112.x + 7.7, min(gsl_tmp_112.y, gsl_remap_105(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_112.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_114 = vec2<f32>(gsl_tmp_113.x + 7.7, min(gsl_tmp_113.y, gsl_remap_106(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_113.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_115 = vec2<f32>(gsl_tmp_114.x + 7.7, min(gsl_tmp_114.y, gsl_remap_107(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_114.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_116 = vec2<f32>(gsl_tmp_115.x + 7.7, min(gsl_tmp_115.y, gsl_remap_108(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_115.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_117 = vec2<f32>(gsl_tmp_116.x + 7.7, min(gsl_tmp_116.y, gsl_remap_109(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_116.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_118 = vec2<f32>(gsl_tmp_117.x + 7.7, min(gsl_tmp_117.y, gsl_remap_110(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_117.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_119 = vec2<f32>(gsl_tmp_118.x + 8.7, min(gsl_tmp_118.y, gsl_remap_111(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_118.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_120 = vec2<f32>(gsl_tmp_119.x + 4.7, min(gsl_tmp_119.y, gsl_remap_112(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_119.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))));
  let l05 = min(gsl_tmp_120.y, gsl_remap_113(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_120.x, (p.y - 1.125) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let gsl_tmp_121 = vec2<f32>(6.7, min(1000000.0, gsl_remap_114(vec3<f32>((p.x - -3.2083333) / 0.027777778 - 0.0, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_122 = vec2<f32>(gsl_tmp_121.x + 6.7, min(gsl_tmp_121.y, gsl_remap_115(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_121.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_123 = vec2<f32>(gsl_tmp_122.x + 6.7, min(gsl_tmp_122.y, gsl_remap_116(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_122.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_124 = vec2<f32>(gsl_tmp_123.x + 6.7, min(gsl_tmp_123.y, gsl_remap_117(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_123.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_125 = vec2<f32>(gsl_tmp_124.x + 6.7, min(gsl_tmp_124.y, gsl_remap_118(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_124.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_126 = vec2<f32>(gsl_tmp_125.x + 6.7, min(gsl_tmp_125.y, gsl_remap_58(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_125.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_127 = vec2<f32>(gsl_tmp_126.x + 10.7, min(gsl_tmp_126.y, gsl_remap_119(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_126.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_128 = vec2<f32>(gsl_tmp_127.x + 5.7, min(gsl_tmp_127.y, gsl_remap_120(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_127.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_129 = vec2<f32>(gsl_tmp_128.x + 6.7, min(gsl_tmp_128.y, gsl_remap_121(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_128.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_130 = vec2<f32>(gsl_tmp_129.x + 6.7, min(gsl_tmp_129.y, gsl_remap_122(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_129.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_131 = vec2<f32>(gsl_tmp_130.x + 6.7, min(gsl_tmp_130.y, gsl_remap_123(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_130.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_132 = vec2<f32>(gsl_tmp_131.x + 6.7, min(gsl_tmp_131.y, gsl_remap_124(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_131.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_133 = vec2<f32>(gsl_tmp_132.x + 3.7, min(gsl_tmp_132.y, gsl_remap_125(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_132.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_134 = vec2<f32>(gsl_tmp_133.x + 3.7, min(gsl_tmp_133.y, gsl_remap_126(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_133.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_135 = vec2<f32>(gsl_tmp_134.x + 3.7, min(gsl_tmp_134.y, gsl_remap_127(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_134.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_136 = vec2<f32>(gsl_tmp_135.x + 3.7, min(gsl_tmp_135.y, gsl_remap_128(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_135.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_137 = vec2<f32>(gsl_tmp_136.x + 6.7, min(gsl_tmp_136.y, gsl_remap_129(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_136.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_138 = vec2<f32>(gsl_tmp_137.x + 6.7, min(gsl_tmp_137.y, gsl_remap_130(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_137.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_139 = vec2<f32>(gsl_tmp_138.x + 6.7, min(gsl_tmp_138.y, gsl_remap_131(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_138.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_140 = vec2<f32>(gsl_tmp_139.x + 6.7, min(gsl_tmp_139.y, gsl_remap_132(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_139.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_141 = vec2<f32>(gsl_tmp_140.x + 6.7, min(gsl_tmp_140.y, gsl_remap_133(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_140.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_142 = vec2<f32>(gsl_tmp_141.x + 6.7, min(gsl_tmp_141.y, gsl_remap_134(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_141.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_143 = vec2<f32>(gsl_tmp_142.x + 6.7, min(gsl_tmp_142.y, gsl_remap_135(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_142.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_144 = vec2<f32>(gsl_tmp_143.x + 6.7, min(gsl_tmp_143.y, gsl_remap_136(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_143.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_145 = vec2<f32>(gsl_tmp_144.x + 6.7, min(gsl_tmp_144.y, gsl_remap_137(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_144.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_146 = vec2<f32>(gsl_tmp_145.x + 6.7, min(gsl_tmp_145.y, gsl_remap_138(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_145.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_147 = vec2<f32>(gsl_tmp_146.x + 6.7, min(gsl_tmp_146.y, gsl_remap_139(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_146.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_148 = vec2<f32>(gsl_tmp_147.x + 6.7, min(gsl_tmp_147.y, gsl_remap_140(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_147.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_149 = vec2<f32>(gsl_tmp_148.x + 6.7, min(gsl_tmp_148.y, gsl_remap_141(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_148.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_150 = vec2<f32>(gsl_tmp_149.x + 6.7, min(gsl_tmp_149.y, gsl_remap_142(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_149.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))));
  let l06 = min(gsl_tmp_150.y, gsl_remap_143(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_150.x, (p.y - 0.625) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let gsl_tmp_151 = vec2<f32>(8.7, min(1000000.0, gsl_remap_144(vec3<f32>((p.x - -3.2083333) / 0.027777778 - 0.0, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_152 = vec2<f32>(gsl_tmp_151.x + 8.7, min(gsl_tmp_151.y, gsl_remap_145(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_151.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_153 = vec2<f32>(gsl_tmp_152.x + 8.7, min(gsl_tmp_152.y, gsl_remap_146(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_152.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_154 = vec2<f32>(gsl_tmp_153.x + 6.7, min(gsl_tmp_153.y, gsl_remap_147(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_153.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_155 = vec2<f32>(gsl_tmp_154.x + 6.7, min(gsl_tmp_154.y, gsl_remap_148(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_154.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_156 = vec2<f32>(gsl_tmp_155.x + 6.7, min(gsl_tmp_155.y, gsl_remap_149(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_155.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_157 = vec2<f32>(gsl_tmp_156.x + 6.7, min(gsl_tmp_156.y, gsl_remap_150(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_156.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_158 = vec2<f32>(gsl_tmp_157.x + 7.7, min(gsl_tmp_157.y, gsl_remap_151(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_157.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_159 = vec2<f32>(gsl_tmp_158.x + 7.7, min(gsl_tmp_158.y, gsl_remap_99(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_158.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_160 = vec2<f32>(gsl_tmp_159.x + 6.7, min(gsl_tmp_159.y, gsl_remap_152(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_159.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_161 = vec2<f32>(gsl_tmp_160.x + 6.7, min(gsl_tmp_160.y, gsl_remap_153(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_160.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_162 = vec2<f32>(gsl_tmp_161.x + 6.7, min(gsl_tmp_161.y, gsl_remap_154(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_161.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_163 = vec2<f32>(gsl_tmp_162.x + 6.7, min(gsl_tmp_162.y, gsl_remap_155(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_162.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_164 = vec2<f32>(gsl_tmp_163.x + 6.7, min(gsl_tmp_163.y, gsl_remap_156(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_163.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_165 = vec2<f32>(gsl_tmp_164.x + 7.7, min(gsl_tmp_164.y, gsl_remap_157(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_164.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_166 = vec2<f32>(gsl_tmp_165.x + 7.7, min(gsl_tmp_165.y, gsl_remap_158(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_165.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_167 = vec2<f32>(gsl_tmp_166.x + 7.7, min(gsl_tmp_166.y, gsl_remap_159(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_166.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_168 = vec2<f32>(gsl_tmp_167.x + 7.7, min(gsl_tmp_167.y, gsl_remap_160(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_167.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_169 = vec2<f32>(gsl_tmp_168.x + 7.7, min(gsl_tmp_168.y, gsl_remap_161(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_168.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_170 = vec2<f32>(gsl_tmp_169.x + 8.7, min(gsl_tmp_169.y, gsl_remap_162(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_169.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_171 = vec2<f32>(gsl_tmp_170.x + 4.1, min(gsl_tmp_170.y, gsl_remap_163(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_170.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_172 = vec2<f32>(gsl_tmp_171.x + 4.2, min(gsl_tmp_171.y, gsl_remap_164(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_171.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_173 = vec2<f32>(gsl_tmp_172.x + 3.7, min(gsl_tmp_172.y, gsl_remap_165(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_172.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_174 = vec2<f32>(gsl_tmp_173.x + 2.7, min(gsl_tmp_173.y, gsl_remap_166(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_173.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_175 = vec2<f32>(gsl_tmp_174.x + 2.7, min(gsl_tmp_174.y, gsl_remap_167(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_174.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_176 = vec2<f32>(gsl_tmp_175.x + 6.7, min(gsl_tmp_175.y, gsl_remap_168(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_175.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_177 = vec2<f32>(gsl_tmp_176.x + 6.7, min(gsl_tmp_176.y, gsl_remap_169(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_176.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_178 = vec2<f32>(gsl_tmp_177.x + 7.7, min(gsl_tmp_177.y, gsl_remap_170(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_177.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_179 = vec2<f32>(gsl_tmp_178.x + 6.7, min(gsl_tmp_178.y, gsl_remap_171(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_178.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))));
  let l07 = min(gsl_tmp_179.y, gsl_remap_172(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_179.x, (p.y - 0.125) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let gsl_tmp_180 = vec2<f32>(6.7, min(1000000.0, gsl_remap_173(vec3<f32>((p.x - -3.2083333) / 0.027777778 - 0.0, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_181 = vec2<f32>(gsl_tmp_180.x + 6.7, min(gsl_tmp_180.y, gsl_remap_174(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_180.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_182 = vec2<f32>(gsl_tmp_181.x + 6.7, min(gsl_tmp_181.y, gsl_remap_175(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_181.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_183 = vec2<f32>(gsl_tmp_182.x + 5.7, min(gsl_tmp_182.y, gsl_remap_176(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_182.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_184 = vec2<f32>(gsl_tmp_183.x + 5.7, min(gsl_tmp_183.y, gsl_remap_177(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_183.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_185 = vec2<f32>(gsl_tmp_184.x + 5.7, min(gsl_tmp_184.y, gsl_remap_178(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_184.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_186 = vec2<f32>(gsl_tmp_185.x + 5.7, min(gsl_tmp_185.y, gsl_remap_179(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_185.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_187 = vec2<f32>(gsl_tmp_186.x + 7.7, min(gsl_tmp_186.y, gsl_remap_180(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_186.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_188 = vec2<f32>(gsl_tmp_187.x + 7.7, min(gsl_tmp_187.y, gsl_remap_181(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_187.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_189 = vec2<f32>(gsl_tmp_188.x + 6.7, min(gsl_tmp_188.y, gsl_remap_182(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_188.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_190 = vec2<f32>(gsl_tmp_189.x + 6.7, min(gsl_tmp_189.y, gsl_remap_183(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_189.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_191 = vec2<f32>(gsl_tmp_190.x + 6.7, min(gsl_tmp_190.y, gsl_remap_184(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_190.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_192 = vec2<f32>(gsl_tmp_191.x + 6.7, min(gsl_tmp_191.y, gsl_remap_185(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_191.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_193 = vec2<f32>(gsl_tmp_192.x + 6.7, min(gsl_tmp_192.y, gsl_remap_186(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_192.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_194 = vec2<f32>(gsl_tmp_193.x + 6.7, min(gsl_tmp_193.y, gsl_remap_187(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_193.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_195 = vec2<f32>(gsl_tmp_194.x + 6.7, min(gsl_tmp_194.y, gsl_remap_188(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_194.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_196 = vec2<f32>(gsl_tmp_195.x + 6.7, min(gsl_tmp_195.y, gsl_remap_189(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_195.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_197 = vec2<f32>(gsl_tmp_196.x + 6.7, min(gsl_tmp_196.y, gsl_remap_190(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_196.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_198 = vec2<f32>(gsl_tmp_197.x + 6.7, min(gsl_tmp_197.y, gsl_remap_191(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_197.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_199 = vec2<f32>(gsl_tmp_198.x + 6.7, min(gsl_tmp_198.y, gsl_remap_192(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_198.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_200 = vec2<f32>(gsl_tmp_199.x + 4.1, min(gsl_tmp_199.y, gsl_remap_193(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_199.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_201 = vec2<f32>(gsl_tmp_200.x + 4.2, min(gsl_tmp_200.y, gsl_remap_194(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_200.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_202 = vec2<f32>(gsl_tmp_201.x + 3.7, min(gsl_tmp_201.y, gsl_remap_195(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_201.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_203 = vec2<f32>(gsl_tmp_202.x + 2.7, min(gsl_tmp_202.y, gsl_remap_196(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_202.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_204 = vec2<f32>(gsl_tmp_203.x + 2.7, min(gsl_tmp_203.y, text_segmentDist(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_203.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778), vec2<f32>(0.0, 0.0), vec2<f32>(0.0, 6.0))));
  let gsl_tmp_205 = vec2<f32>(gsl_tmp_204.x + 4.7, min(gsl_tmp_204.y, gsl_remap_197(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_204.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_206 = vec2<f32>(gsl_tmp_205.x + 5.7, min(gsl_tmp_205.y, gsl_remap_198(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_205.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_207 = vec2<f32>(gsl_tmp_206.x + 6.7, min(gsl_tmp_206.y, gsl_remap_199(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_206.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let gsl_tmp_208 = vec2<f32>(gsl_tmp_207.x + 4.7, min(gsl_tmp_207.y, gsl_remap_200(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_207.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))));
  let l08 = min(gsl_tmp_208.y, gsl_remap_201(vec3<f32>((p.x - -3.2083333) / 0.027777778 - gsl_tmp_208.x, (p.y - -0.375) / 0.027777778, p.z / 0.027777778))) * 0.027777778 - 0.01;
  let roman = min(min(min(min(min(min(min(l01, l02), l03), l04), l05), l06), l07), l08);
  let gsl_s1 = gsl_shape(max(roman, abs(p.z) - 0.05), 1u);
  return gsl_s1;
}
