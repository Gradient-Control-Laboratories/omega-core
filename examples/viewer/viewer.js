/**
 * viewer.js - Zero-dependency client-side WebGPU viewer for procedural GSL shaders and CAD meshes.
 */

import { MeshLoader } from "./mesh_loader.js";

class ViewerApp {
  constructor() {
    this.canvas = document.getElementById("webgpu-canvas");
    this.device = null;
    this.context = null;
    this.presentationFormat = "bgra8unorm";

    // Camera state
    this.camera = {
      target: [0, 0, 0],
      radius: 6.0,
      theta: Math.PI / 4,   // azimuth
      phi: Math.PI / 6,     // elevation
      fov: (45 * Math.PI) / 180,
      aspect: 1.0,
      near: 0.1,
      far: 100.0,
      isDragging: false,
      isPanning: false,
      lastMouse: [0, 0],
      inMotion: false,
      motionTimer: null
    };

    // Animation & Playback state
    this.isPlaying = false;
    this.currentTime = 0.0;
    this.playbackSpeed = 1.0;
    this.lastFrameTime = performance.now();
    this.maxTime = 10.0;

    // Rendering options
    this.showSDF = true;
    this.showMesh = true;
    this.showWireframe = false;
    this.showGrid = true;
    this.clipZ = 5.0;
    this.lightingPreset = 0; // 0=Studio, 1=Warm, 2=Cool, 3=Dramatic, 4=Metallic

    // Active shader & mesh data
    this.currentWgslCode = "";
    this.currentMetadata = null;
    this.sliderValues = new Float32Array(16); // up to 16 parameter sliders
    this.loadedMesh = null;
    this.meshGpuBuffers = null;

    // WebGPU resources
    this.depthTexture = null;
    this.sdfPipeline = null;
    this.meshPipeline = null;
    this.wireframePipeline = null;
    this.gridPipeline = null;

    this.cameraUniformBuffer = null;
    this.sliderUniformBuffer = null;
    this.cameraBindGroup = null;

    // Performance diagnostics
    this.frameCount = 0;
    this.fpsTimer = performance.now();
    this.currentFps = 60;
  }

  async init() {
    if (!navigator.gpu) {
      this.showFatalError("WebGPU is not supported in this browser. Please use Chrome 113+, Edge 113+, or Safari 18+ on macOS Sequoia.");
      return;
    }

    try {
      const adapter = await navigator.gpu.requestAdapter({ powerPreference: "high-performance" });
      if (!adapter) {
        this.showFatalError("No suitable WebGPU adapter found.");
        return;
      }
      this.device = await adapter.requestDevice();
    } catch (e) {
      this.showFatalError("Failed to initialize WebGPU device: " + e.message);
      return;
    }

    this.context = this.canvas.getContext("webgpu");
    this.presentationFormat = navigator.gpu.getPreferredCanvasFormat();
    this.context.configure({
      device: this.device,
      format: this.presentationFormat,
      alphaMode: "premultiplied"
    });

    this.initUniformBuffers();
    this.setupEventListeners();
    this.resizeCanvas();

    // Load initial model from showcase
    await this.loadShowcaseModel("gear");

    // Start render loop
    requestAnimationFrame((t) => this.renderLoop(t));
  }

  showFatalError(msg) {
    const status = document.getElementById("shape-status");
    if (status) {
      status.textContent = "Error";
      status.style.color = "#f85149";
    }
    const container = document.getElementById("canvas-container");
    container.innerHTML = `
      <div style="display:flex;align-items:center;justify-content:center;height:100%;color:#f85149;padding:32px;text-align:center;font-size:16px;">
        <div>
          <h2>WebGPU Initialization Failed</h2>
          <p style="margin-top:12px;color:#8b949e;">${msg}</p>
        </div>
      </div>
    `;
  }

  initUniformBuffers() {
    // Camera uniforms:
    // viewProj: 16 floats (64 bytes)
    // invViewProj: 16 floats (64 bytes)
    // eyePos: 3 floats + 1 pad (16 bytes)
    // time, clipZ, lightingPreset, pad (16 bytes)
    // Total = 160 bytes -> align to 256 bytes
    this.cameraUniformBuffer = this.device.createBuffer({
      size: 256,
      usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST
    });

    // Slider uniform buffer: 16 floats (64 bytes) -> align to 256 bytes
    this.sliderUniformBuffer = this.device.createBuffer({
      size: 256,
      usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST
    });
  }

  resizeCanvas() {
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const width = Math.floor(this.canvas.clientWidth * dpr);
    const height = Math.floor(this.canvas.clientHeight * dpr);

    if (this.canvas.width !== width || this.canvas.height !== height || !this.depthTexture) {
      if (width > 0 && height > 0) {
        this.canvas.width = width;
        this.canvas.height = height;
        this.camera.aspect = width / height;

        // Recreate depth texture
        if (this.depthTexture) this.depthTexture.destroy();
        this.depthTexture = this.device.createTexture({
          size: [width, height],
          format: "depth24plus",
          usage: GPUTextureUsage.RENDER_ATTACHMENT
        });
      }
    }
  }

  // --- Model and Asset Ingestion ---

  async loadShowcaseModel(name) {
    const status = document.getElementById("shape-status");
    if (status) status.textContent = "Loading...";

    try {
      // 1. Fetch WGSL shader
      const shaderResp = await fetch(`shaders/${name}.wgsl`);
      if (!shaderResp.ok) throw new Error(`Failed to load shaders/${name}.wgsl`);
      const wgslCode = await shaderResp.text();

      // 2. Fetch companion JSON metadata
      let meta = null;
      try {
        const metaResp = await fetch(`shaders/${name}.json`);
        if (metaResp.ok) meta = await metaResp.json();
      } catch (_) {
        // metadata is optional
      }

      // 3. Load hybrid mesh if hybrid_bracket
      if (name === "hybrid_bracket") {
        try {
          const meshResp = await fetch("../assets/bracket.stl");
          if (meshResp.ok) {
            const buf = await meshResp.arrayBuffer();
            this.setLoadedMesh(MeshLoader.parse(buf, "bracket.stl"));
          }
        } catch (_) {}
      } else {
        this.clearLoadedMesh();
      }

      await this.setShaderCode(wgslCode, meta);
      if (status) status.textContent = "Ready";
    } catch (e) {
      if (status) {
        status.textContent = "Error";
        status.style.color = "#f85149";
      }
      console.error("Failed to load model:", e);
    }
  }

  async setShaderCode(wgslCode, metadata = null) {
    this.currentWgslCode = wgslCode;
    this.currentMetadata = metadata;

    this.rebuildDynamicSliders(metadata);
    await this.compileSdfPipeline();
  }

  setLoadedMesh(meshData) {
    this.loadedMesh = meshData;

    // Create GPU buffers for triangle mesh
    const vBuf = this.device.createBuffer({
      size: meshData.positions.byteLength,
      usage: GPUBufferUsage.VERTEX | GPUBufferUsage.COPY_DST,
      mappedAtCreation: true
    });
    new Float32Array(vBuf.getMappedRange()).set(meshData.positions);
    vBuf.unmap();

    const nBuf = this.device.createBuffer({
      size: meshData.normals.byteLength,
      usage: GPUBufferUsage.VERTEX | GPUBufferUsage.COPY_DST,
      mappedAtCreation: true
    });
    new Float32Array(nBuf.getMappedRange()).set(meshData.normals);
    nBuf.unmap();

    const iBuf = this.device.createBuffer({
      size: meshData.indices.byteLength,
      usage: GPUBufferUsage.INDEX | GPUBufferUsage.COPY_DST,
      mappedAtCreation: true
    });
    new Uint32Array(iBuf.getMappedRange()).set(meshData.indices);
    iBuf.unmap();

    this.meshGpuBuffers = {
      vertexBuffer: vBuf,
      normalBuffer: nBuf,
      indexBuffer: iBuf,
      indexCount: meshData.indices.length
    };

    this.compileMeshPipelines();
  }

  clearLoadedMesh() {
    this.loadedMesh = null;
    this.meshGpuBuffers = null;
  }

  rebuildDynamicSliders(metadata) {
    const container = document.getElementById("dynamic-sliders-container");
    if (!container) return;
    container.innerHTML = "";

    const sliders = metadata?.sliders || [];
    if (sliders.length === 0) {
      container.innerHTML = '<span style="color:var(--text-muted);font-size:12px;">No parametric sliders for this model.</span>';
      return;
    }

    sliders.forEach((s, idx) => {
      if (idx >= 16) return;
      const initial = s.initialValue ?? (s.min + s.max) * 0.5;
      this.sliderValues[idx] = initial;

      const control = document.createElement("div");
      control.className = "slider-control";

      const labelRow = document.createElement("div");
      labelRow.className = "slider-label-row";

      const nameSpan = document.createElement("span");
      nameSpan.className = "slider-name";
      nameSpan.textContent = s.variableName || `Param ${idx + 1}`;

      const valSpan = document.createElement("span");
      valSpan.className = "slider-val";
      valSpan.id = `val-slider-${idx}`;
      valSpan.textContent = initial.toFixed(s.decimals || 2);

      labelRow.appendChild(nameSpan);
      labelRow.appendChild(valSpan);

      const input = document.createElement("input");
      input.type = "range";
      input.min = s.min;
      input.max = s.max;
      input.step = Math.pow(10, -(s.decimals || 2));
      input.value = initial;

      input.addEventListener("input", (e) => {
        const v = parseFloat(e.target.value);
        this.sliderValues[idx] = v;
        valSpan.textContent = v.toFixed(s.decimals || 2);
        this.markMotion();
      });

      control.appendChild(labelRow);
      control.appendChild(input);
      container.appendChild(control);
    });

    this.updateSliderUniformBuffer();
  }

  // --- Shader Pipelines ---

  async compileSdfPipeline() {
    // Inspect sdShape parameters to bind slider uniforms dynamically
    const match = this.currentWgslCode.match(/fn\s+sdShape\s*\(([^)]*)\)/);
    let extraArgsStr = "";
    if (match) {
      const params = match[1].split(",").map(s => s.trim()).filter(Boolean);
      // First two parameters are p: vec3<f32>, time: f32
      const extraCount = Math.max(0, params.length - 2);
      const sliderAccessors = [
        "sliders[0].x", "sliders[0].y", "sliders[0].z", "sliders[0].w",
        "sliders[1].x", "sliders[1].y", "sliders[1].z", "sliders[1].w",
        "sliders[2].x", "sliders[2].y", "sliders[2].z", "sliders[2].w",
        "sliders[3].x", "sliders[3].y", "sliders[3].z", "sliders[3].w",
      ];
      const args = ["pt", "t"];
      for (let i = 0; i < extraCount; i++) {
        args.push(sliderAccessors[i] || "0.0");
      }
      extraArgsStr = args.join(", ");
    } else {
      extraArgsStr = "pt, t";
    }

    // Generate SDF raymarching shader with camera uniforms and clip plane
    const fullShader = `
struct CameraUniforms {
  viewProj: mat4x4<f32>,
  invViewProj: mat4x4<f32>,
  eyePos: vec3<f32>,
  time: f32,
  clipZ: f32,
  lightingPreset: f32,
  motionDraft: f32,
};

@group(0) @binding(0) var<uniform> camera: CameraUniforms;
@group(0) @binding(1) var<uniform> sliders: array<vec4<f32>, 4>;

struct gsl_shape {
  d: f32,
  shapeIndex: u32,
};

fn gsl_min(a: gsl_shape, b: gsl_shape) -> gsl_shape {
  if (abs(a.d) < abs(b.d)) {
    return a;
  } else {
    return b;
  }
}

struct VertexOutput {
  @builtin(position) position: vec4<f32>,
  @location(0) uv: vec2<f32>,
};

@vertex
fn vs_main(@builtin(vertex_index) vertexIndex: u32) -> VertexOutput {
  var pos = array<vec2<f32>, 3>(
    vec2<f32>(-1.0, -1.0),
    vec2<f32>( 3.0, -1.0),
    vec2<f32>(-1.0,  3.0)
  );
  var out: VertexOutput;
  out.position = vec4<f32>(pos[vertexIndex], 0.0, 1.0);
  out.uv = pos[vertexIndex] * 0.5 + vec2<f32>(0.5, 0.5);
  return out;
}

// Emitted GSL code containing sdShape
${this.currentWgslCode}

// Dynamic dispatch wrapper injecting active slider uniform values
fn evalSDF(pt: vec3<f32>, t: f32) -> gsl_shape {
  return sdShape(${extraArgsStr});
}

struct FragmentOutput {
  @location(0) color: vec4<f32>,
  @builtin(frag_depth) depth: f32,
};

@fragment
fn fs_main(in: VertexOutput) -> FragmentOutput {
  let ndc = vec4<f32>(in.uv * 2.0 - 1.0, 1.0, 1.0);
  let worldTarget = camera.invViewProj * ndc;
  let rayDir = normalize(worldTarget.xyz / worldTarget.w - camera.eyePos);

  var t: f32 = 0.01;
  let maxDist: f32 = 40.0;
  let maxSteps: i32 = select(128, 64, camera.motionDraft > 0.5);

  var out: FragmentOutput;
  out.color = vec4<f32>(0.0, 0.0, 0.0, 0.0);
  out.depth = 1.0;

  for (var i: i32 = 0; i < maxSteps; i = i + 1) {
    let p = camera.eyePos + rayDir * t;

    // Cross-section clipping plane (active if clipZ < 4.9)
    if (camera.clipZ < 4.9 && p.z > camera.clipZ) {
      if (rayDir.z < -0.001) {
        let skip = (p.z - camera.clipZ) / (-rayDir.z);
        t = t + max(skip, 0.02);
      } else {
        t = t + 0.1;
      }
      continue;
    }

    let res = evalSDF(p, camera.time);
    let dist = res.d;

    if (dist < 0.001) {
      // Calculate normal via gradient
      let eps = 0.001;
      let nx = evalSDF(p + vec3<f32>(eps, 0.0, 0.0), camera.time).d - evalSDF(p - vec3<f32>(eps, 0.0, 0.0), camera.time).d;
      let ny = evalSDF(p + vec3<f32>(0.0, eps, 0.0), camera.time).d - evalSDF(p - vec3<f32>(0.0, eps, 0.0), camera.time).d;
      let nz = evalSDF(p + vec3<f32>(0.0, 0.0, eps), camera.time).d - evalSDF(p - vec3<f32>(0.0, 0.0, eps), camera.time).d;
      let normal = normalize(vec3<f32>(nx, ny, nz));

      // Environment lighting models
      var baseColor = vec3<f32>(0.85, 0.88, 0.95);
      var light1 = normalize(vec3<f32>(1.0, 2.0, 3.0));
      var light2 = normalize(vec3<f32>(-1.0, -1.0, 1.0));

      if (camera.lightingPreset > 0.5 && camera.lightingPreset < 1.5) {
        // Warm sunset
        baseColor = vec3<f32>(1.0, 0.75, 0.6);
        light1 = normalize(vec3<f32>(2.0, 1.0, 0.5));
      } else if (camera.lightingPreset > 1.5 && camera.lightingPreset < 2.5) {
        // Cool industrial
        baseColor = vec3<f32>(0.65, 0.8, 1.0);
      } else if (camera.lightingPreset > 3.5) {
        // High-contrast metallic
        baseColor = vec3<f32>(0.95, 0.95, 1.0);
      }

      let diff1 = max(dot(normal, light1), 0.0);
      let diff2 = max(dot(normal, light2), 0.0) * 0.3;
      let ambient = 0.15;

      let litColor = baseColor * (diff1 + diff2 + ambient);

      // Hardware depth buffer projection (from WebDM)
      let clipPos = camera.viewProj * vec4<f32>(p, 1.0);
      out.depth = clamp(clipPos.z / clipPos.w, 0.0, 1.0);
      out.color = vec4<f32>(litColor, 1.0);
      return out;
    }

    t = t + dist;
    if (t > maxDist) {
      break;
    }
  }

  discard;
  return out;
}
`;

    const shaderModule = this.device.createShaderModule({ code: fullShader });
    const compInfo = await shaderModule.getCompilationInfo();
    let hasError = false;
    for (const msg of compInfo.messages) {
      if (msg.type === "error") {
        hasError = true;
        console.error(`WGSL error [${msg.lineNum}:${msg.linePos}]: ${msg.message}`);
      } else if (msg.type === "warning") {
        console.warn(`WGSL warning [${msg.lineNum}:${msg.linePos}]: ${msg.message}`);
      }
    }
    if (hasError) {
      throw new Error("Shader module compilation failed. See console for details.");
    }

    const bindGroupLayout = this.device.createBindGroupLayout({
      entries: [
        { binding: 0, visibility: GPUShaderStage.VERTEX | GPUShaderStage.FRAGMENT, buffer: { type: "uniform" } },
        { binding: 1, visibility: GPUShaderStage.FRAGMENT, buffer: { type: "uniform" } }
      ]
    });

    this.cameraBindGroup = this.device.createBindGroup({
      layout: bindGroupLayout,
      entries: [
        { binding: 0, resource: { buffer: this.cameraUniformBuffer } },
        { binding: 1, resource: { buffer: this.sliderUniformBuffer } }
      ]
    });

    const pipelineLayout = this.device.createPipelineLayout({
      bindGroupLayouts: [bindGroupLayout]
    });

    this.sdfPipeline = this.device.createRenderPipeline({
      layout: pipelineLayout,
      vertex: { module: shaderModule, entryPoint: "vs_main" },
      fragment: {
        module: shaderModule,
        entryPoint: "fs_main",
        targets: [{ format: this.presentationFormat }]
      },
      primitive: { topology: "triangle-list" },
      depthStencil: {
        depthWriteEnabled: true,
        depthCompare: "always",
        format: "depth24plus"
      }
    });
  }

  compileMeshPipelines() {
    const meshShader = `
struct CameraUniforms {
  viewProj: mat4x4<f32>,
  invViewProj: mat4x4<f32>,
  eyePos: vec3<f32>,
  time: f32,
  clipZ: f32,
  lightingPreset: f32,
  motionDraft: f32,
};

@group(0) @binding(0) var<uniform> camera: CameraUniforms;

struct VertexInput {
  @location(0) position: vec3<f32>,
  @location(1) normal: vec3<f32>,
};

struct VertexOutput {
  @builtin(position) position: vec4<f32>,
  @location(0) worldPos: vec3<f32>,
  @location(1) normal: vec3<f32>,
};

@vertex
fn vs_mesh(input: VertexInput) -> VertexOutput {
  var out: VertexOutput;
  out.worldPos = input.position;
  out.normal = input.normal;
  out.position = camera.viewProj * vec4<f32>(input.position, 1.0);
  return out;
}

@fragment
fn fs_mesh(input: VertexOutput) -> @location(0) vec4<f32> {
  if (input.worldPos.z > camera.clipZ) {
    discard;
  }
  let lightDir = normalize(vec3<f32>(1.0, 2.0, 3.0));
  let diff = max(dot(normalize(input.normal), lightDir), 0.15);
  let color = vec3<f32>(0.4, 0.65, 0.9) * diff + vec3<f32>(0.05);
  return vec4<f32>(color, 0.85);
}

@fragment
fn fs_wireframe(input: VertexOutput) -> @location(0) vec4<f32> {
  return vec4<f32>(0.2, 0.8, 1.0, 0.9);
}
`;

    const module = this.device.createShaderModule({ code: meshShader });

    const bindGroupLayout = this.device.createBindGroupLayout({
      entries: [
        { binding: 0, visibility: GPUShaderStage.VERTEX | GPUShaderStage.FRAGMENT, buffer: { type: "uniform" } }
      ]
    });

    const meshBindGroup = this.device.createBindGroup({
      layout: bindGroupLayout,
      entries: [{ binding: 0, resource: { buffer: this.cameraUniformBuffer } }]
    });
    this.meshBindGroup = meshBindGroup;

    const pipelineLayout = this.device.createPipelineLayout({ bindGroupLayouts: [bindGroupLayout] });

    this.meshPipeline = this.device.createRenderPipeline({
      layout: pipelineLayout,
      vertex: {
        module,
        entryPoint: "vs_mesh",
        buffers: [
          { arrayStride: 12, attributes: [{ shaderLocation: 0, offset: 0, format: "float32x3" }] },
          { arrayStride: 12, attributes: [{ shaderLocation: 1, offset: 0, format: "float32x3" }] }
        ]
      },
      fragment: {
        module,
        entryPoint: "fs_mesh",
        targets: [{
          format: this.presentationFormat,
          blend: {
            color: { srcFactor: "src-alpha", dstFactor: "one-minus-src-alpha", operation: "add" },
            alpha: { srcFactor: "one", dstFactor: "one-minus-src-alpha", operation: "add" }
          }
        }]
      },
      primitive: { topology: "triangle-list", cullMode: "none" },
      depthStencil: { depthWriteEnabled: true, depthCompare: "less-equal", format: "depth24plus" }
    });

    this.wireframePipeline = this.device.createRenderPipeline({
      layout: pipelineLayout,
      vertex: {
        module,
        entryPoint: "vs_mesh",
        buffers: [
          { arrayStride: 12, attributes: [{ shaderLocation: 0, offset: 0, format: "float32x3" }] },
          { arrayStride: 12, attributes: [{ shaderLocation: 1, offset: 0, format: "float32x3" }] }
        ]
      },
      fragment: {
        module,
        entryPoint: "fs_wireframe",
        targets: [{
          format: this.presentationFormat,
          blend: {
            color: { srcFactor: "src-alpha", dstFactor: "one-minus-src-alpha", operation: "add" },
            alpha: { srcFactor: "one", dstFactor: "one-minus-src-alpha", operation: "add" }
          }
        }]
      },
      primitive: { topology: "line-strip", cullMode: "none" },
      depthStencil: { depthWriteEnabled: false, depthCompare: "less-equal", format: "depth24plus" }
    });
  }

  // --- Camera & Matrix Calculations ---

  updateCameraMatrices() {
    const eye = [
      this.camera.radius * Math.cos(this.camera.phi) * Math.cos(this.camera.theta) + this.camera.target[0],
      this.camera.radius * Math.cos(this.camera.phi) * Math.sin(this.camera.theta) + this.camera.target[1],
      this.camera.radius * Math.sin(this.camera.phi) + this.camera.target[2]
    ];

    const viewProj = this.createViewProjectionMatrix(eye, this.camera.target, [0, 0, 1]);
    const invViewProj = this.invertMatrix4(viewProj);

    const uniformData = new Float32Array(40);
    uniformData.set(viewProj, 0);       // 0-15
    uniformData.set(invViewProj, 16);    // 16-31
    uniformData[32] = eye[0];           // 32
    uniformData[33] = eye[1];           // 33
    uniformData[34] = eye[2];           // 34
    uniformData[35] = this.currentTime; // 35: time
    uniformData[36] = this.clipZ;       // 36: clipZ
    uniformData[37] = this.lightingPreset; // 37: lightingPreset
    uniformData[38] = this.camera.inMotion ? 1.0 : 0.0; // 38: motionDraft
    uniformData[39] = 0.0;              // 39: pad

    this.device.queue.writeBuffer(this.cameraUniformBuffer, 0, uniformData.buffer);
  }

  updateSliderUniformBuffer() {
    this.device.queue.writeBuffer(this.sliderUniformBuffer, 0, this.sliderValues.buffer);
  }

  markMotion() {
    this.camera.inMotion = true;
    if (this.camera.motionTimer) clearTimeout(this.camera.motionTimer);
    this.camera.motionTimer = setTimeout(() => {
      this.camera.inMotion = false;
    }, 120);
  }

  createViewProjectionMatrix(eye, target, up) {
    // LookAt
    const zAxis = this.normalize([eye[0] - target[0], eye[1] - target[1], eye[2] - target[2]]);
    const xAxis = this.normalize(this.cross(up, zAxis));
    const yAxis = this.cross(zAxis, xAxis);

    const view = [
      xAxis[0], yAxis[0], zAxis[0], 0,
      xAxis[1], yAxis[1], zAxis[1], 0,
      xAxis[2], yAxis[2], zAxis[2], 0,
      -this.dot(xAxis, eye), -this.dot(yAxis, eye), -this.dot(zAxis, eye), 1
    ];

    // Perspective
    const f = 1.0 / Math.tan(this.camera.fov / 2);
    const rangeInv = 1.0 / (this.camera.near - this.camera.far);

    const proj = [
      f / this.camera.aspect, 0, 0, 0,
      0, f, 0, 0,
      0, 0, this.camera.far * rangeInv, -1,
      0, 0, this.camera.near * this.camera.far * rangeInv, 0
    ];

    return this.multiplyMatrices4(proj, view);
  }

  multiplyMatrices4(a, b) {
    const out = new Float32Array(16);
    for (let i = 0; i < 4; i++) {
      for (let j = 0; j < 4; j++) {
        let sum = 0;
        for (let k = 0; k < 4; k++) {
          sum += a[k * 4 + i] * b[j * 4 + k];
        }
        out[j * 4 + i] = sum;
      }
    }
    return out;
  }

  invertMatrix4(m) {
    const inv = new Float32Array(16);
    inv[0] = m[5] * m[10] * m[15] - m[5] * m[11] * m[14] - m[9] * m[6] * m[15] + m[9] * m[7] * m[14] + m[13] * m[6] * m[11] - m[13] * m[7] * m[10];
    inv[4] = -m[4] * m[10] * m[15] + m[4] * m[11] * m[14] + m[8] * m[6] * m[15] - m[8] * m[7] * m[14] - m[12] * m[6] * m[11] + m[12] * m[7] * m[10];
    inv[8] = m[4] * m[9] * m[15] - m[4] * m[11] * m[13] - m[8] * m[5] * m[15] + m[8] * m[7] * m[13] + m[12] * m[5] * m[11] - m[12] * m[7] * m[9];
    inv[12] = -m[4] * m[9] * m[14] + m[4] * m[10] * m[13] + m[8] * m[5] * m[14] - m[8] * m[6] * m[13] - m[12] * m[5] * m[10] + m[12] * m[6] * m[9];

    let det = m[0] * inv[0] + m[1] * inv[4] + m[2] * inv[8] + m[3] * inv[12];
    if (det === 0) return new Float32Array(16);
    det = 1.0 / det;

    inv[1] = -m[1] * m[10] * m[15] + m[1] * m[11] * m[14] + m[9] * m[2] * m[15] - m[9] * m[3] * m[14] - m[13] * m[2] * m[11] + m[13] * m[3] * m[10];
    inv[5] = m[0] * m[10] * m[15] - m[0] * m[11] * m[14] - m[8] * m[2] * m[15] + m[8] * m[3] * m[14] + m[12] * m[2] * m[11] - m[12] * m[3] * m[10];
    inv[9] = -m[0] * m[9] * m[15] + m[0] * m[11] * m[13] + m[8] * m[1] * m[15] - m[8] * m[3] * m[13] - m[12] * m[1] * m[11] + m[12] * m[3] * m[9];
    inv[13] = m[0] * m[9] * m[14] - m[0] * m[10] * m[13] - m[8] * m[1] * m[14] + m[8] * m[2] * m[13] + m[12] * m[1] * m[10] - m[12] * m[2] * m[9];

    inv[2] = m[1] * m[6] * m[15] - m[1] * m[7] * m[14] - m[5] * m[2] * m[15] + m[5] * m[3] * m[14] + m[13] * m[2] * m[7] - m[13] * m[3] * m[6];
    inv[6] = -m[0] * m[6] * m[15] + m[0] * m[7] * m[14] + m[4] * m[2] * m[15] - m[4] * m[3] * m[14] - m[12] * m[2] * m[7] + m[12] * m[3] * m[6];
    inv[10] = m[0] * m[5] * m[15] - m[0] * m[7] * m[13] - m[4] * m[1] * m[15] + m[4] * m[3] * m[13] + m[12] * m[1] * m[7] - m[12] * m[3] * m[5];
    inv[14] = -m[0] * m[5] * m[14] + m[0] * m[6] * m[13] + m[4] * m[1] * m[14] - m[4] * m[2] * m[13] - m[12] * m[1] * m[6] + m[12] * m[2] * m[5];

    inv[3] = -m[1] * m[6] * m[11] + m[1] * m[7] * m[10] + m[5] * m[2] * m[11] - m[5] * m[3] * m[10] - m[9] * m[2] * m[7] + m[9] * m[3] * m[6];
    inv[7] = m[0] * m[6] * m[11] - m[0] * m[7] * m[10] - m[4] * m[2] * m[11] + m[4] * m[3] * m[10] + m[8] * m[2] * m[7] - m[8] * m[3] * m[6];
    inv[11] = -m[0] * m[5] * m[11] + m[0] * m[7] * m[9] + m[4] * m[1] * m[11] - m[4] * m[3] * m[9] - m[8] * m[1] * m[7] + m[8] * m[3] * m[5];
    inv[15] = m[0] * m[5] * m[10] - m[0] * m[6] * m[9] - m[4] * m[1] * m[10] + m[4] * m[2] * m[9] + m[8] * m[1] * m[6] - m[8] * m[2] * m[5];

    for (let i = 0; i < 16; i++) inv[i] *= det;
    return inv;
  }

  normalize(v) {
    const len = Math.sqrt(v[0] * v[0] + v[1] * v[1] + v[2] * v[2]);
    return len > 0 ? [v[0] / len, v[1] / len, v[2] / len] : [0, 0, 0];
  }

  cross(a, b) {
    return [
      a[1] * b[2] - a[2] * b[1],
      a[2] * b[0] - a[0] * b[2],
      a[0] * b[1] - a[1] * b[0]
    ];
  }

  dot(a, b) {
    return a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
  }

  // --- Rendering Loop ---

  renderLoop(now) {
    const delta = (now - this.lastFrameTime) / 1000;
    this.lastFrameTime = now;

    // Advance timeline if playing
    if (this.isPlaying) {
      this.currentTime += delta * this.playbackSpeed;
      if (this.currentTime > this.maxTime) {
        this.currentTime = 0;
      }
      this.updateTimelineUI();
      this.markMotion();
    }

    // FPS counter
    this.frameCount++;
    if (now - this.fpsTimer >= 1000) {
      this.currentFps = this.frameCount;
      this.frameCount = 0;
      this.fpsTimer = now;
      const fpsEl = document.getElementById("diag-fps");
      if (fpsEl) fpsEl.textContent = this.currentFps;
    }

    this.resizeCanvas();
    this.updateCameraMatrices();
    this.updateSliderUniformBuffer();

    const commandEncoder = this.device.createCommandEncoder();
    const textureView = this.context.getCurrentTexture().createView();

    const renderPassDescriptor = {
      colorAttachments: [{
        view: textureView,
        clearValue: { r: 0.05, g: 0.06, b: 0.09, a: 1.0 },
        loadOp: "clear",
        storeOp: "store"
      }],
      depthStencilAttachment: {
        view: this.depthTexture.createView(),
        depthClearValue: 1.0,
        depthLoadOp: "clear",
        depthStoreOp: "store"
      }
    };

    const pass = commandEncoder.beginRenderPass(renderPassDescriptor);

    // Pass 1: Procedural SDF quad
    if (this.showSDF && this.sdfPipeline && this.cameraBindGroup) {
      pass.setPipeline(this.sdfPipeline);
      pass.setBindGroup(0, this.cameraBindGroup);
      pass.draw(3, 1, 0, 0);
    }

    // Pass 2: Traditional polygon CAD mesh
    if (this.showMesh && this.meshGpuBuffers && this.meshPipeline) {
      pass.setPipeline(this.showWireframe ? this.wireframePipeline : this.meshPipeline);
      pass.setBindGroup(0, this.meshBindGroup);
      pass.setVertexBuffer(0, this.meshGpuBuffers.vertexBuffer);
      pass.setVertexBuffer(1, this.meshGpuBuffers.normalBuffer);
      pass.setIndexBuffer(this.meshGpuBuffers.indexBuffer, "uint32");
      pass.drawIndexed(this.meshGpuBuffers.indexCount);
    }

    pass.end();
    this.device.queue.submit([commandEncoder.finish()]);

    requestAnimationFrame((t) => this.renderLoop(t));
  }

  updateTimelineUI() {
    const slider = document.getElementById("timeline-slider");
    const label = document.getElementById("time-current");
    if (slider) slider.value = this.currentTime;
    if (label) label.textContent = `${this.currentTime.toFixed(2)}s`;
  }

  // --- Interactive Event Handlers ---

  setupEventListeners() {
    window.addEventListener("resize", () => this.resizeCanvas());

    // Canvas Orbit and Pan
    this.canvas.addEventListener("mousedown", (e) => {
      this.camera.isDragging = e.button === 0;
      this.camera.isPanning = e.button === 1 || (e.button === 0 && e.shiftKey);
      this.camera.lastMouse = [e.clientX, e.clientY];
    });

    window.addEventListener("mousemove", (e) => {
      if (!this.camera.isDragging && !this.camera.isPanning) return;
      const dx = e.clientX - this.camera.lastMouse[0];
      const dy = e.clientY - this.camera.lastMouse[1];
      this.camera.lastMouse = [e.clientX, e.clientY];
      this.markMotion();

      if (this.camera.isPanning) {
        const factor = this.camera.radius * 0.002;
        this.camera.target[0] -= dx * factor;
        this.camera.target[1] += dy * factor;
      } else if (this.camera.isDragging) {
        this.camera.theta -= dx * 0.008;
        this.camera.phi = Math.max(-Math.PI / 2 + 0.05, Math.min(Math.PI / 2 - 0.05, this.camera.phi + dy * 0.008));
      }
    });

    window.addEventListener("mouseup", () => {
      this.camera.isDragging = false;
      this.camera.isPanning = false;
    });

    this.canvas.addEventListener("wheel", (e) => {
      e.preventDefault();
      this.camera.radius = Math.max(1.0, Math.min(30.0, this.camera.radius + e.deltaY * 0.01));
      this.markMotion();
    }, { passive: false });

    // Model Selector
    const modelSelect = document.getElementById("model-select");
    if (modelSelect) {
      modelSelect.addEventListener("change", (e) => {
        this.loadShowcaseModel(e.target.value);
      });
    }

    // Camera Views
    const viewSelect = document.getElementById("camera-view-select");
    if (viewSelect) {
      viewSelect.addEventListener("change", (e) => {
        switch (e.target.value) {
          case "iso": this.camera.theta = Math.PI / 4; this.camera.phi = Math.PI / 6; break;
          case "top": this.camera.theta = 0; this.camera.phi = Math.PI / 2 - 0.01; break;
          case "front": this.camera.theta = 0; this.camera.phi = 0; break;
          case "right": this.camera.theta = Math.PI / 2; this.camera.phi = 0; break;
        }
        this.markMotion();
      });
    }

    const resetBtn = document.getElementById("btn-reset-cam");
    if (resetBtn) {
      resetBtn.addEventListener("click", () => {
        this.camera.target = [0, 0, 0];
        this.camera.radius = 6.0;
        this.camera.theta = Math.PI / 4;
        this.camera.phi = Math.PI / 6;
        this.markMotion();
      });
    }

    // Screenshot
    const snapBtn = document.getElementById("btn-screenshot");
    if (snapBtn) {
      snapBtn.addEventListener("click", () => {
        const link = document.createElement("a");
        link.download = "gsl_view.png";
        link.href = this.canvas.toDataURL("image/png");
        link.click();
      });
    }

    // Visibility Checkboxes
    document.getElementById("chk-show-sdf")?.addEventListener("change", (e) => {
      this.showSDF = e.target.checked;
    });
    document.getElementById("chk-show-mesh")?.addEventListener("change", (e) => {
      this.showMesh = e.target.checked;
    });
    document.getElementById("chk-show-wireframe")?.addEventListener("change", (e) => {
      this.showWireframe = e.target.checked;
    });

    // Cross-Section Clip Plane
    const clipSlider = document.getElementById("slider-clip-z");
    const clipVal = document.getElementById("val-clip-z");
    if (clipSlider) {
      clipSlider.addEventListener("input", (e) => {
        const v = parseFloat(e.target.value);
        this.clipZ = v;
        if (clipVal) clipVal.textContent = v >= 5.0 ? "None" : `${v.toFixed(2)}`;
        this.markMotion();
      });
    }

    // Lighting Preset
    document.getElementById("lighting-preset-select")?.addEventListener("change", (e) => {
      const presets = { studio: 0, warm: 1, cool: 2, dramatic: 3, metallic: 4 };
      this.lightingPreset = presets[e.target.value] ?? 0;
      this.markMotion();
    });

    // Animation Toolbar Controls
    const playBtn = document.getElementById("btn-play-pause");
    const playIcon = document.getElementById("play-icon");
    if (playBtn) {
      playBtn.addEventListener("click", () => {
        this.isPlaying = !this.isPlaying;
        if (playIcon) playIcon.textContent = this.isPlaying ? "⏸ Pause" : "▶ Play";
      });
    }

    window.addEventListener("keydown", (e) => {
      if (e.code === "Space" && e.target.tagName !== "INPUT") {
        e.preventDefault();
        this.isPlaying = !this.isPlaying;
        if (playIcon) playIcon.textContent = this.isPlaying ? "⏸ Pause" : "▶ Play";
      }
    });

    const timeline = document.getElementById("timeline-slider");
    if (timeline) {
      timeline.addEventListener("input", (e) => {
        this.currentTime = parseFloat(e.target.value);
        this.updateTimelineUI();
        this.markMotion();
      });
    }

    document.getElementById("playback-speed-select")?.addEventListener("change", (e) => {
      this.playbackSpeed = parseFloat(e.target.value);
    });

    // Drag and Drop Hot-reloading
    const overlay = document.getElementById("drag-overlay");
    window.addEventListener("dragenter", (e) => {
      e.preventDefault();
      if (overlay) overlay.style.display = "flex";
    });

    window.addEventListener("dragover", (e) => {
      e.preventDefault();
    });

    window.addEventListener("dragleave", (e) => {
      if (e.relatedTarget === null && overlay) {
        overlay.style.display = "none";
      }
    });

    window.addEventListener("drop", async (e) => {
      e.preventDefault();
      if (overlay) overlay.style.display = "none";

      const files = Array.from(e.dataTransfer.files);
      for (const file of files) {
        const name = file.name.toLowerCase();
        if (name.endsWith(".wgsl")) {
          const text = await file.text();
          await this.setShaderCode(text);
        } else if (name.endsWith(".stl") || name.endsWith(".obj")) {
          const buf = await file.arrayBuffer();
          this.setLoadedMesh(MeshLoader.parse(buf, file.name));
        } else if (name.endsWith(".json")) {
          const jsonText = await file.text();
          try {
            const meta = JSON.parse(jsonText);
            this.rebuildDynamicSliders(meta);
          } catch (_) {}
        }
      }
    });
  }
}

// Bootstrap application on DOM ready or immediately if already loaded
function startApp() {
  const app = new ViewerApp();
  app.init();
  window.viewerInstance = app;
}

if (document.readyState === "loading") {
  window.addEventListener("DOMContentLoaded", startApp);
} else {
  startApp();
}
