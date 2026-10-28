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

const cyclePeriod = 4;
    const phase = gsl_fract(gsl_time / 4);
    const baseFreq = 8;
    const f0 = 8 * gsl_exp2(0 + phase);
    const a0 = 0.04393693362340742;
    const f1 = 8 * gsl_exp2(1 + phase);
    const a1 = 0.1353352832366127;
    const f2 = 8 * gsl_exp2(2 + phase);
    const a2 = 0.32465246735834974;
    const f3 = 8 * gsl_exp2(3 + phase);
    const a3 = 0.6065306597126334;
    const f4 = 8 * gsl_exp2(4 + phase);
    const a4 = 0.8824969025845955;
    const f5 = 8 * gsl_exp2(5 + phase);
    const a5 = 1;
    const f6 = 8 * gsl_exp2(6 + phase);
    const a6 = 0.8824969025845955;
    const f7 = 8 * gsl_exp2(7 + phase);
    const a7 = 0.6065306597126334;
    const f8 = 8 * gsl_exp2(8 + phase);
    const a8 = 0.32465246735834974;
    const f9 = 8 * gsl_exp2(9 + phase);
    const a9 = 0.1353352832366127;
    const tau = 6.283185307;
    const t = gsl_time;
    const stereoOff = 0.0003;
    const outL = 0.09 * (0.04393693362340742 * Math.sin(6.283185307 * f0 * t) + 0.1353352832366127 * Math.sin(6.283185307 * f1 * t) + 0.32465246735834974 * Math.sin(6.283185307 * f2 * t) + 0.6065306597126334 * Math.sin(6.283185307 * f3 * t) + 0.8824969025845955 * Math.sin(6.283185307 * f4 * t) + 1 * Math.sin(6.283185307 * f5 * t) + 0.8824969025845955 * Math.sin(6.283185307 * f6 * t) + 0.6065306597126334 * Math.sin(6.283185307 * f7 * t) + 0.32465246735834974 * Math.sin(6.283185307 * f8 * t) + 0.1353352832366127 * Math.sin(6.283185307 * f9 * t));
    const tr = gsl_time + 0.0003;
    const outR = 0.09 * (0.04393693362340742 * Math.sin(6.283185307 * f0 * tr) + 0.1353352832366127 * Math.sin(6.283185307 * f1 * tr) + 0.32465246735834974 * Math.sin(6.283185307 * f2 * tr) + 0.6065306597126334 * Math.sin(6.283185307 * f3 * tr) + 0.8824969025845955 * Math.sin(6.283185307 * f4 * tr) + 1 * Math.sin(6.283185307 * f5 * tr) + 0.8824969025845955 * Math.sin(6.283185307 * f6 * tr) + 0.6065306597126334 * Math.sin(6.283185307 * f7 * tr) + 0.32465246735834974 * Math.sin(6.283185307 * f8 * tr) + 0.1353352832366127 * Math.sin(6.283185307 * f9 * tr));
    return [outL, outR];
}
