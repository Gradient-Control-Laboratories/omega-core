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

function nearest(xs, s) {
const gsl_fold = (xs[0] > 0 ? 1.0 : 0.0) > 0 ? Math.min(1e+09, Math.abs(s - 0)) : 1e+09;
    const gsl_fold_2 = (xs[1] > 0 ? 1.0 : 0.0) > 0 ? Math.min(gsl_fold, Math.abs(s - 1)) : gsl_fold;
    const gsl_fold_3 = (xs[2] > 0 ? 1.0 : 0.0) > 0 ? Math.min(gsl_fold_2, Math.abs(s - 2)) : gsl_fold_2;
    const gsl_fold_4 = (xs[3] > 0 ? 1.0 : 0.0) > 0 ? Math.min(gsl_fold_3, Math.abs(s - 3)) : gsl_fold_3;
    const gsl_fold_5 = (xs[4] > 0 ? 1.0 : 0.0) > 0 ? Math.min(gsl_fold_4, Math.abs(s - 4)) : gsl_fold_4;
    const gsl_fold_6 = (xs[5] > 0 ? 1.0 : 0.0) > 0 ? Math.min(gsl_fold_5, Math.abs(s - 5)) : gsl_fold_5;
    const gsl_fold_7 = (xs[6] > 0 ? 1.0 : 0.0) > 0 ? Math.min(gsl_fold_6, Math.abs(s - 6)) : gsl_fold_6;
    return (xs[7] > 0 ? 1.0 : 0.0) > 0 ? Math.min(gsl_fold_7, Math.abs(s - 7)) : gsl_fold_7;
}

    return [Math.sin(gsl_time * 220) * 0.1 / (1 + (nearest([1, 0, 0, 1, 0, 1, 1, 0], gsl_fract(gsl_time) * 8))), Math.sin(gsl_time * 220) * 0.1 / (1 + (nearest([1, 0, 0, 1, 0, 1, 1, 0], gsl_fract(gsl_time) * 8)))];
}
