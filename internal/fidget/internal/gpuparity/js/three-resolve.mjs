// The viewers' import map, as a Node resolve hook (see register-three.mjs).
const vendor = new URL("../../../../../examples/viewer/vendor/three/", import.meta.url);

const MAP = {
  "three": "three.webgpu.js",
  "three/webgpu": "three.webgpu.js",
  "three/tsl": "three.tsl.js",
};

export async function resolve(specifier, context, next) {
  if (specifier in MAP) {
    return { url: new URL(MAP[specifier], vendor).href, shortCircuit: true };
  }
  return next(specifier, context);
}
