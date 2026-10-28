/**
 * mesh_loader.js - Zero-dependency client-side parser for STL (binary and ASCII) and OBJ polygon meshes.
 * Converts 3D mesh files into WebGPU vertex buffers, normal buffers, and index buffers in browser memory.
 */

export class MeshLoader {
  /**
   * Parse an ArrayBuffer or string into a structured 3D Mesh object.
   * @param {ArrayBuffer|string} bufferOrText 
   * @param {string} filename 
   * @returns {{ positions: Float32Array, normals: Float32Array, indices: Uint32Array, triangleCount: number, bbox: number[] }}
   */
  static parse(bufferOrText, filename = "") {
    const isObj = filename.toLowerCase().endsWith(".obj");
    if (isObj || (typeof bufferOrText === "string" && (bufferOrText.startsWith("v ") || bufferOrText.startsWith("#")))) {
      const text = typeof bufferOrText === "string" ? bufferOrText : new TextDecoder().decode(bufferOrText);
      return this.parseOBJ(text);
    }

    if (typeof bufferOrText === "string") {
      return this.parseASCIIStl(bufferOrText);
    }

    // Binary or ASCII STL check
    const dv = new DataView(bufferOrText);
    if (bufferOrText.byteLength > 84) {
      // Check if header starts with 'solid'
      const headerStr = new TextDecoder().decode(new Uint8Array(bufferOrText, 0, Math.min(80, bufferOrText.byteLength)));
      const declaredTris = dv.getUint32(80, true);
      const expectedSize = 84 + declaredTris * 50;

      if (headerStr.startsWith("solid") && expectedSize !== bufferOrText.byteLength) {
        // Likely ASCII STL
        const text = new TextDecoder().decode(bufferOrText);
        return this.parseASCIIStl(text);
      }
      return this.parseBinarySTL(bufferOrText);
    }

    const text = new TextDecoder().decode(bufferOrText);
    return this.parseASCIIStl(text);
  }

  /**
   * Parse binary STL buffer.
   */
  static parseBinarySTL(buffer) {
    const dv = new DataView(buffer);
    const triangleCount = dv.getUint32(80, true);

    const positions = new Float32Array(triangleCount * 9);
    const normals = new Float32Array(triangleCount * 9);
    const indices = new Uint32Array(triangleCount * 3);

    let offset = 84;
    let posIdx = 0;
    let idx = 0;

    let minX = Infinity, minY = Infinity, minZ = Infinity;
    let maxX = -Infinity, maxY = -Infinity, maxZ = -Infinity;

    for (let i = 0; i < triangleCount; i++) {
      if (offset + 50 > buffer.byteLength) break;

      const nx = dv.getFloat32(offset, true);
      const ny = dv.getFloat32(offset + 4, true);
      const nz = dv.getFloat32(offset + 8, true);
      offset += 12;

      for (let v = 0; v < 3; v++) {
        const x = dv.getFloat32(offset, true);
        const y = dv.getFloat32(offset + 4, true);
        const z = dv.getFloat32(offset + 8, true);
        offset += 12;

        positions[posIdx] = x;
        positions[posIdx + 1] = y;
        positions[posIdx + 2] = z;

        normals[posIdx] = nx;
        normals[posIdx + 1] = ny;
        normals[posIdx + 2] = nz;

        indices[idx] = idx;

        if (x < minX) minX = x; if (x > maxX) maxX = x;
        if (y < minY) minY = y; if (y > maxY) maxY = y;
        if (z < minZ) minZ = z; if (z > maxZ) maxZ = z;

        posIdx += 3;
        idx++;
      }

      offset += 2; // attribute byte count
    }

    return {
      positions,
      normals,
      indices,
      triangleCount,
      bbox: [minX, minY, minZ, maxX, maxY, maxZ]
    };
  }

  /**
   * Parse ASCII STL text.
   */
  static parseASCIIStl(text) {
    const lines = text.split("\n");
    const posList = [];
    const normList = [];

    let curNormal = [0, 0, 1];
    let minX = Infinity, minY = Infinity, minZ = Infinity;
    let maxX = -Infinity, maxY = -Infinity, maxZ = -Infinity;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (line.startsWith("facet normal")) {
        const parts = line.split(/\s+/);
        curNormal = [parseFloat(parts[2]), parseFloat(parts[3]), parseFloat(parts[4])];
      } else if (line.startsWith("vertex")) {
        const parts = line.split(/\s+/);
        const x = parseFloat(parts[1]);
        const y = parseFloat(parts[2]);
        const z = parseFloat(parts[3]);

        posList.push(x, y, z);
        normList.push(curNormal[0], curNormal[1], curNormal[2]);

        if (x < minX) minX = x; if (x > maxX) maxX = x;
        if (y < minY) minY = y; if (y > maxY) maxY = y;
        if (z < minZ) minZ = z; if (z > maxZ) maxZ = z;
      }
    }

    const count = Math.floor(posList.length / 3);
    const indices = new Uint32Array(count);
    for (let i = 0; i < count; i++) indices[i] = i;

    return {
      positions: new Float32Array(posList),
      normals: new Float32Array(normList),
      indices,
      triangleCount: Math.floor(count / 3),
      bbox: [minX, minY, minZ, maxX, maxY, maxZ]
    };
  }

  /**
   * Parse Wavefront OBJ text.
   */
  static parseOBJ(text) {
    const lines = text.split("\n");
    const vertices = [];
    const vNormals = [];

    const outPositions = [];
    const outNormals = [];

    let minX = Infinity, minY = Infinity, minZ = Infinity;
    let maxX = -Infinity, maxY = -Infinity, maxZ = -Infinity;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (line.startsWith("v ")) {
        const p = line.split(/\s+/);
        vertices.push([parseFloat(p[1]), parseFloat(p[2]), parseFloat(p[3])]);
      } else if (line.startsWith("vn ")) {
        const p = line.split(/\s+/);
        vNormals.push([parseFloat(p[1]), parseFloat(p[2]), parseFloat(p[3])]);
      } else if (line.startsWith("f ")) {
        const p = line.split(/\s+/).slice(1);
        const faceIndices = p.map(elem => {
          const parts = elem.split("/");
          return {
            v: parseInt(parts[0], 10) - 1,
            vn: parts[2] ? parseInt(parts[2], 10) - 1 : -1
          };
        });

        // Fan triangulation for convex polygons
        for (let j = 1; j < faceIndices.length - 1; j++) {
          const tri = [faceIndices[0], faceIndices[j], faceIndices[j + 1]];
          for (const item of tri) {
            const v = vertices[item.v];
            outPositions.push(v[0], v[1], v[2]);

            if (v[0] < minX) minX = v[0]; if (v[0] > maxX) maxX = v[0];
            if (v[1] < minY) minY = v[1]; if (v[1] > maxY) maxY = v[1];
            if (v[2] < minZ) minZ = v[2]; if (v[2] > maxZ) maxZ = v[2];

            if (item.vn >= 0 && vNormals[item.vn]) {
              const n = vNormals[item.vn];
              outNormals.push(n[0], n[1], n[2]);
            } else {
              outNormals.push(0, 0, 1);
            }
          }
        }
      }
    }

    const count = Math.floor(outPositions.length / 3);
    const indices = new Uint32Array(count);
    for (let i = 0; i < count; i++) indices[i] = i;

    return {
      positions: new Float32Array(outPositions),
      normals: new Float32Array(outNormals),
      indices,
      triangleCount: Math.floor(count / 3),
      bbox: [minX, minY, minZ, maxX, maxY, maxZ]
    };
  }
}
