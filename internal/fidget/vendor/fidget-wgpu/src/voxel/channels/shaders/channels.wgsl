//! Pass to evaluate each shape's N channel tapes at every pixel's hit voxel
//! (the color pass, generalized to N unclamped floats per pixel)

struct Config {
    /// Screen-to-model transform matrix, converting pixels to model space
    mat: mat4x4f,

    /// Mapping from X, Y, Z to input indices
    axes: vec3u,

    // Explicit padding
    _pad1: u32,

    /// Image size, in pixels
    image_size: vec2u,

    /// Channel tapes per shape, and floats written per pixel
    channels: u32,

    // Explicit padding
    _pad2: u32,
}

/// Config (stored in the channels workspace)
@group(0) @binding(0) var<uniform> config: Config;

/// Shape info (stored in the shape object)
@group(1) @binding(0) var<storage, read> shape_tape: array<TapeWord>;
@group(1) @binding(1) var<storage, read> shape_start: array<u32>;
@group(1) @binding(2) var<storage, read> var_values: array<f32>;

/// Output of voxel evaluation (stored in the pixel buffers)
@group(2) @binding(0) var<storage, read> image: array<PackedVoxel>;

/// Output of channel evaluation (stored in the channel buffers)
@group(3) @binding(0) var<storage, read_write> out: array<f32>;

@compute @workgroup_size(8, 8)
fn channels_main(
    @builtin(global_invocation_id) global_id: vec3u
) {
    // Clamp to image size
    if global_id.x >= config.image_size.x ||
       global_id.y >= config.image_size.y
    {
        return;
    }

    let i = global_id.x + config.image_size.x * global_id.y;
    let n = config.channels;

    let p = unpack(image[i]);
    if p.pixel.depth == 0 || p.index >= arrayLength(&shape_start) {
        // empty or corrupt, fill with zeros
        for (var c = 0u; c < n; c++) {
            out[i * n + c] = 0.0;
        }
        return;
    }

    // Compute input values
    let corner_pos = vec3(global_id.x, global_id.y, p.pixel.depth);
    let m = transformed_inputs(
        Value(f32(corner_pos.x)),
        Value(f32(corner_pos.y)),
        Value(f32(corner_pos.z)),
    );

    // Channel tapes are packed together, one after another
    var pos = shape_start[p.index];
    var stack = Stack(); // dummy value
    for (var c = 0u; c < n; c++) {
        let r = run_tape(pos, m, &stack);
        out[i * n + c] = r.value.v;
        pos = r.pos;
    }
}

fn get_tape_word(i: u32) -> TapeWord {
    return shape_tape[i];
}
