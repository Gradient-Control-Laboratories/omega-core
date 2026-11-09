// GSL runtime helpers for JavaScript audio
function gsl_clamp(x, min, max) {
    return Math.max(min, Math.min(max, x));
}

function gsl_fract(x) {
    return x - Math.floor(x);
}

function gsl_sign(x) {
    return x > 0 ? 1 : (x < 0 ? -1 : 0);
}

function gsl_exp2(x) {
    return Math.pow(2, x);
}

function gsl_radians(degrees) {
    return degrees * Math.PI / 180;
}

function gsl_degrees(radians) {
    return radians * 180 / Math.PI;
}

function gsl_normalize(v) {
    if (!Array.isArray(v)) return v;
    const len = Math.sqrt(v.reduce((sum, x) => sum + x * x, 0));
    return len === 0 ? v : v.map(x => x / len);
}

function gsl_length(v) {
    if (!Array.isArray(v)) return Math.abs(v);
    return Math.sqrt(v.reduce((sum, x) => sum + x * x, 0));
}

function gsl_distance(a, b) {
    if (!Array.isArray(a)) return Math.abs(a - b);
    return gsl_length(a.map((x, i) => x - b[i]));
}

function gsl_dot(a, b) {
    if (!Array.isArray(a)) return a * b;
    return a.reduce((sum, x, i) => sum + x * b[i], 0);
}

function gsl_cross(a, b) {
    if (a.length !== 3 || b.length !== 3) throw new Error("cross product requires 3D vectors");
    return [
        a[1] * b[2] - a[2] * b[1],
        a[2] * b[0] - a[0] * b[2],
        a[0] * b[1] - a[1] * b[0]
    ];
}

function computeAudio(gsl_time) {
const gsl_p = [0, 0, 0];
const gsl_pi = Math.PI;
const gsl_e = Math.E;

const plateW = 3;
    const plateD = 5;
    const fixLeft = 1;
    const fixRight = 0;
    const fixFront = 1;
    const fixBack = 0;
    const bounceRate = 0.9;
    const tLocal = gsl_fract(0.9 * gsl_time);
    const horizRateX = 0.23;
    const horizRateY = 0.17;
    const bounceStartTime = gsl_time - tLocal / 0.9;
    const horizPhaseXAtHit = gsl_fract(0.23 * bounceStartTime);
    const triWaveXAtHit = Math.abs(2 * horizPhaseXAtHit - 1);
    const hitX = 3 * (2 * triWaveXAtHit - 1);
    const horizPhaseYAtHit = gsl_fract(0.17 * bounceStartTime);
    const triWaveYAtHit = Math.abs(2 * horizPhaseYAtHit - 1);
    const hitY = 5 * (2 * triWaveYAtHit - 1);
    const a = 6;
    const b = 10;
    const pi = 3.14159265;
    const twoPi = 6.2831853;
    const f11 = 191.49403465150002;
    const f21 = 613.90734638275;
    const f12 = 343.56282687475;
    const f31 = 1317.9295326015001;
    const f22 = 765.9761386060001;
    const f13 = 597.0108139135;
    const f41 = 2303.56059330775;
    const fr32 = 1469.9983248247502;
    const f23 = 1019.42412564475;
    const f51 = 3570.8005285015006;
    const f42 = 2455.629385531;
    const f33 = 1723.4463118635;
    const f14 = 951.8379957677499;
    const f52 = 3722.86932072475;
    const f43 = 2709.07737256975;
    const gsl_tmp_1 = (hitX + 3) / 6;
    const gsl_tmp_2 = (hitY + 5) / 10;
    const w11 = (0 * Math.sin(3.14159265 * gsl_tmp_1) + 1 * Math.sin(1.570796325 * gsl_tmp_1) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_1)) + 0 * Math.cos(3.14159265 * gsl_tmp_1)) * (0 * Math.sin(3.14159265 * gsl_tmp_2) + 1 * Math.sin(1.570796325 * gsl_tmp_2) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_2)) + 0 * Math.cos(3.14159265 * gsl_tmp_2));
    const gsl_tmp_3 = (hitX + 3) / 6;
    const gsl_tmp_4 = (hitY + 5) / 10;
    const w21 = (0 * Math.sin(6.2831853 * gsl_tmp_3) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_3) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_3)) + 0 * Math.cos(6.2831853 * gsl_tmp_3)) * (0 * Math.sin(3.14159265 * gsl_tmp_4) + 1 * Math.sin(1.570796325 * gsl_tmp_4) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_4)) + 0 * Math.cos(3.14159265 * gsl_tmp_4));
    const gsl_tmp_5 = (hitX + 3) / 6;
    const gsl_tmp_6 = (hitY + 5) / 10;
    const w12 = (0 * Math.sin(3.14159265 * gsl_tmp_5) + 1 * Math.sin(1.570796325 * gsl_tmp_5) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_5)) + 0 * Math.cos(3.14159265 * gsl_tmp_5)) * (0 * Math.sin(6.2831853 * gsl_tmp_6) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_6) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_6)) + 0 * Math.cos(6.2831853 * gsl_tmp_6));
    const gsl_tmp_7 = (hitX + 3) / 6;
    const gsl_tmp_8 = (hitY + 5) / 10;
    const w31 = (0 * Math.sin(9.424777950000001 * gsl_tmp_7) + 1 * Math.sin(7.853981625 * gsl_tmp_7) + 0 * Math.sin(7.853981625 * (1 - gsl_tmp_7)) + 0 * Math.cos(9.424777950000001 * gsl_tmp_7)) * (0 * Math.sin(3.14159265 * gsl_tmp_8) + 1 * Math.sin(1.570796325 * gsl_tmp_8) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_8)) + 0 * Math.cos(3.14159265 * gsl_tmp_8));
    const gsl_tmp_9 = (hitX + 3) / 6;
    const gsl_tmp_10 = (hitY + 5) / 10;
    const w22 = (0 * Math.sin(6.2831853 * gsl_tmp_9) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_9) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_9)) + 0 * Math.cos(6.2831853 * gsl_tmp_9)) * (0 * Math.sin(6.2831853 * gsl_tmp_10) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_10) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_10)) + 0 * Math.cos(6.2831853 * gsl_tmp_10));
    const gsl_tmp_11 = (hitX + 3) / 6;
    const gsl_tmp_12 = (hitY + 5) / 10;
    const w13 = (0 * Math.sin(3.14159265 * gsl_tmp_11) + 1 * Math.sin(1.570796325 * gsl_tmp_11) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_11)) + 0 * Math.cos(3.14159265 * gsl_tmp_11)) * (0 * Math.sin(9.424777950000001 * gsl_tmp_12) + 1 * Math.sin(7.853981625 * gsl_tmp_12) + 0 * Math.sin(7.853981625 * (1 - gsl_tmp_12)) + 0 * Math.cos(9.424777950000001 * gsl_tmp_12));
    const gsl_tmp_13 = (hitX + 3) / 6;
    const gsl_tmp_14 = (hitY + 5) / 10;
    const w41 = (0 * Math.sin(12.5663706 * gsl_tmp_13) + 1 * Math.sin(10.995574275000001 * gsl_tmp_13) + 0 * Math.sin(10.995574275000001 * (1 - gsl_tmp_13)) + 0 * Math.cos(12.5663706 * gsl_tmp_13)) * (0 * Math.sin(3.14159265 * gsl_tmp_14) + 1 * Math.sin(1.570796325 * gsl_tmp_14) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_14)) + 0 * Math.cos(3.14159265 * gsl_tmp_14));
    const gsl_tmp_15 = (hitX + 3) / 6;
    const gsl_tmp_16 = (hitY + 5) / 10;
    const w32 = (0 * Math.sin(9.424777950000001 * gsl_tmp_15) + 1 * Math.sin(7.853981625 * gsl_tmp_15) + 0 * Math.sin(7.853981625 * (1 - gsl_tmp_15)) + 0 * Math.cos(9.424777950000001 * gsl_tmp_15)) * (0 * Math.sin(6.2831853 * gsl_tmp_16) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_16) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_16)) + 0 * Math.cos(6.2831853 * gsl_tmp_16));
    const gsl_tmp_17 = (hitX + 3) / 6;
    const gsl_tmp_18 = (hitY + 5) / 10;
    const w23 = (0 * Math.sin(6.2831853 * gsl_tmp_17) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_17) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_17)) + 0 * Math.cos(6.2831853 * gsl_tmp_17)) * (0 * Math.sin(9.424777950000001 * gsl_tmp_18) + 1 * Math.sin(7.853981625 * gsl_tmp_18) + 0 * Math.sin(7.853981625 * (1 - gsl_tmp_18)) + 0 * Math.cos(9.424777950000001 * gsl_tmp_18));
    const gsl_tmp_19 = (hitX + 3) / 6;
    const gsl_tmp_20 = (hitY + 5) / 10;
    const w51 = (0 * Math.sin(15.70796325 * gsl_tmp_19) + 1 * Math.sin(14.137166925 * gsl_tmp_19) + 0 * Math.sin(14.137166925 * (1 - gsl_tmp_19)) + 0 * Math.cos(15.70796325 * gsl_tmp_19)) * (0 * Math.sin(3.14159265 * gsl_tmp_20) + 1 * Math.sin(1.570796325 * gsl_tmp_20) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_20)) + 0 * Math.cos(3.14159265 * gsl_tmp_20));
    const gsl_tmp_21 = (hitX + 3) / 6;
    const gsl_tmp_22 = (hitY + 5) / 10;
    const w42 = (0 * Math.sin(12.5663706 * gsl_tmp_21) + 1 * Math.sin(10.995574275000001 * gsl_tmp_21) + 0 * Math.sin(10.995574275000001 * (1 - gsl_tmp_21)) + 0 * Math.cos(12.5663706 * gsl_tmp_21)) * (0 * Math.sin(6.2831853 * gsl_tmp_22) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_22) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_22)) + 0 * Math.cos(6.2831853 * gsl_tmp_22));
    const gsl_tmp_23 = (hitX + 3) / 6;
    const gsl_tmp_24 = (hitY + 5) / 10;
    const w33 = (0 * Math.sin(9.424777950000001 * gsl_tmp_23) + 1 * Math.sin(7.853981625 * gsl_tmp_23) + 0 * Math.sin(7.853981625 * (1 - gsl_tmp_23)) + 0 * Math.cos(9.424777950000001 * gsl_tmp_23)) * (0 * Math.sin(9.424777950000001 * gsl_tmp_24) + 1 * Math.sin(7.853981625 * gsl_tmp_24) + 0 * Math.sin(7.853981625 * (1 - gsl_tmp_24)) + 0 * Math.cos(9.424777950000001 * gsl_tmp_24));
    const gsl_tmp_25 = (hitX + 3) / 6;
    const gsl_tmp_26 = (hitY + 5) / 10;
    const w14 = (0 * Math.sin(3.14159265 * gsl_tmp_25) + 1 * Math.sin(1.570796325 * gsl_tmp_25) + 0 * Math.sin(1.570796325 * (1 - gsl_tmp_25)) + 0 * Math.cos(3.14159265 * gsl_tmp_25)) * (0 * Math.sin(12.5663706 * gsl_tmp_26) + 1 * Math.sin(10.995574275000001 * gsl_tmp_26) + 0 * Math.sin(10.995574275000001 * (1 - gsl_tmp_26)) + 0 * Math.cos(12.5663706 * gsl_tmp_26));
    const gsl_tmp_27 = (hitX + 3) / 6;
    const gsl_tmp_28 = (hitY + 5) / 10;
    const w52 = (0 * Math.sin(15.70796325 * gsl_tmp_27) + 1 * Math.sin(14.137166925 * gsl_tmp_27) + 0 * Math.sin(14.137166925 * (1 - gsl_tmp_27)) + 0 * Math.cos(15.70796325 * gsl_tmp_27)) * (0 * Math.sin(6.2831853 * gsl_tmp_28) + 1 * Math.sin(4.7123889750000005 * gsl_tmp_28) + 0 * Math.sin(4.7123889750000005 * (1 - gsl_tmp_28)) + 0 * Math.cos(6.2831853 * gsl_tmp_28));
    const gsl_tmp_29 = (hitX + 3) / 6;
    const gsl_tmp_30 = (hitY + 5) / 10;
    const w43 = (0 * Math.sin(12.5663706 * gsl_tmp_29) + 1 * Math.sin(10.995574275000001 * gsl_tmp_29) + 0 * Math.sin(10.995574275000001 * (1 - gsl_tmp_29)) + 0 * Math.cos(12.5663706 * gsl_tmp_29)) * (0 * Math.sin(9.424777950000001 * gsl_tmp_30) + 1 * Math.sin(7.853981625 * gsl_tmp_30) + 0 * Math.sin(7.853981625 * (1 - gsl_tmp_30)) + 0 * Math.cos(9.424777950000001 * gsl_tmp_30));
    const tEnv = tLocal / 0.9;
    const tau = 0.25;
    const decay1 = Math.exp(-tEnv / 0.25);
    const decay2 = Math.exp(-tEnv / 0.15);
    const decay3 = Math.exp(-tEnv / 0.0875);
    const decay4 = Math.exp(-tEnv / 0.055);
    const decay5 = Math.exp(-tEnv / 0.0375);
    const s11 = 0.35 * w11 * decay1 * Math.sin(1203.1925035599957 * tEnv);
    const s21 = 0.28 * w21 * decay1 * Math.sin(3857.2936143541033 * tEnv);
    const s12 = 0.28 * w12 * decay1 * Math.sin(2158.668903445874 * tEnv);
    const s31 = 0.22 * w31 * decay2 * Math.sin(8280.795465677616 * tEnv);
    const s22 = 0.2 * w22 * decay2 * Math.sin(4812.770014239983 * tEnv);
    const s13 = 0.2 * w13 * decay2 * Math.sin(3751.1295699223388 * tEnv);
    const s41 = 0.15 * w41 * decay3 * Math.sin(14473.698057530533 * tEnv);
    const s32 = 0.14 * w32 * decay3 * Math.sin(9236.271865563496 * tEnv);
    const s23 = 0.14 * w23 * decay3 * Math.sin(6405.230680716447 * tEnv);
    const s51 = 0.1 * w51 * decay4 * Math.sin(22436.00138991286 * tEnv);
    const s42 = 0.1 * w42 * decay4 * Math.sin(15429.174457416413 * tEnv);
    const s33 = 0.1 * w33 * decay4 * Math.sin(10828.73253203996 * tEnv);
    const s14 = 0.08 * w14 * decay4 * Math.sin(5980.574502989389 * tEnv);
    const s52 = 0.06 * w52 * decay5 * Math.sin(23391.477789798737 * tEnv);
    const s43 = 0.06 * w43 * decay5 * Math.sin(17021.635123892876 * tEnv);
    const signal = (s11 + s21 + s12 + s31 + s22 + s13 + s41 + s32 + s23 + s51 + s42 + s33 + s14 + s52 + s43) * 0.5;
    const audioShift = 0.003;
    const t2 = gsl_time - 0.003;
    const tLocal2 = (t2 > 0 ? 1.0 : 0.0) > 0 ? gsl_fract(0.9 * t2) : 0;
    const tEnv2 = tLocal2 / 0.9;
    const decay1b = Math.exp(-tEnv2 / 0.25);
    const decay2b = Math.exp(-tEnv2 / 0.15);
    const decay3b = Math.exp(-tEnv2 / 0.0875);
    const decay4b = Math.exp(-tEnv2 / 0.055);
    const decay5b = Math.exp(-tEnv2 / 0.0375);
    const s11b = 0.35 * w11 * decay1b * Math.sin(1203.1925035599957 * tEnv2);
    const s21b = 0.28 * w21 * decay1b * Math.sin(3857.2936143541033 * tEnv2);
    const s12b = 0.28 * w12 * decay1b * Math.sin(2158.668903445874 * tEnv2);
    const s31b = 0.22 * w31 * decay2b * Math.sin(8280.795465677616 * tEnv2);
    const s22b = 0.2 * w22 * decay2b * Math.sin(4812.770014239983 * tEnv2);
    const s13b = 0.2 * w13 * decay2b * Math.sin(3751.1295699223388 * tEnv2);
    const s41b = 0.15 * w41 * decay3b * Math.sin(14473.698057530533 * tEnv2);
    const s32b = 0.14 * w32 * decay3b * Math.sin(9236.271865563496 * tEnv2);
    const s23b = 0.14 * w23 * decay3b * Math.sin(6405.230680716447 * tEnv2);
    const s51b = 0.1 * w51 * decay4b * Math.sin(22436.00138991286 * tEnv2);
    const s42b = 0.1 * w42 * decay4b * Math.sin(15429.174457416413 * tEnv2);
    const s33b = 0.1 * w33 * decay4b * Math.sin(10828.73253203996 * tEnv2);
    const s14b = 0.08 * w14 * decay4b * Math.sin(5980.574502989389 * tEnv2);
    const s52b = 0.06 * w52 * decay5b * Math.sin(23391.477789798737 * tEnv2);
    const s43b = 0.06 * w43 * decay5b * Math.sin(17021.635123892876 * tEnv2);
    const signalR = (s11b + s21b + s12b + s31b + s22b + s13b + s41b + s32b + s23b + s51b + s42b + s33b + s14b + s52b + s43b) * 0.5;
    const panL = 0.5 - hitX / 6;
    const panR = 0.5 + hitX / 6;
    return [signal * panL, signalR * panR];
}
