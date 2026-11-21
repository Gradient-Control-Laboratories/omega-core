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


function gsl_noiseHash3(c, octaves) {
    if (typeof gsl_noiseTile === 'undefined' || !gsl_noiseTile) return [0, 0, 0];
    var tile = gsl_noiseTile;
    var i = ((((Math.floor(c[2]) & 63) * 64 + (Math.floor(c[1]) & 63)) * 64) + (Math.floor(c[0]) & 63)) * 4;
    var r = Math.fround(tile[i] / 255), g = Math.fround(tile[i + 1] / 255), b = Math.fround(tile[i + 2] / 255);
    if (octaves < 2) {
        return [r - Math.floor(r), g - Math.floor(g), b - Math.floor(b)];
    }
    var j = ((((Math.floor(c[2] / 64) & 63) * 64 + (Math.floor(c[1] / 64) & 63)) * 64) + (Math.floor(c[0] / 64) & 63)) * 4;
    r = Math.fround(r + Math.fround(tile[j] / 255));
    g = Math.fround(g + Math.fround(tile[j + 1] / 255));
    b = Math.fround(b + Math.fround(tile[j + 2] / 255));
    return [r - Math.floor(r), g - Math.floor(g), b - Math.floor(b)];
}

function computeAudio(gsl_time) {
const gsl_p = [0, 0, 0];
const gsl_pi = Math.PI;
const gsl_e = Math.E;

const n = Math.floor(gsl_time * 96000);
    const cell = [n - 4096 * Math.floor(n / 4096), Math.floor(n / 4096), 0];
    const white = gsl_noiseHash3(cell, 2)[0] * 2 - 1;
    const coarse = gsl_noiseHash3([Math.floor(gsl_time * 8), 0, 0], 1)[1];
    return [white * 0.1, white * coarse * 0.1];
}
