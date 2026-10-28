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

const mixed = 0 + Math.sin(gsl_time) / 1 + Math.sin(gsl_time * 2) / 2 + Math.sin(gsl_time * 3) / 3 + Math.sin(gsl_time * 4) / 4;
    const quieter = mixed * 0.5;
    return [mixed, quieter];
}
