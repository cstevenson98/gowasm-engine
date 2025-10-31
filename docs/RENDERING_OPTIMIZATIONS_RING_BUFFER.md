# Rendering “Easy Wins” with a Ring Buffer

This guide outlines practical, low-risk optimizations to improve sprite rendering throughput using a preallocated ring buffer, without changing the current single textured pipeline architecture.

## Goals
- Reduce per-frame allocations and driver overhead
- Avoid GPU/CPU stalls from overwriting in-use buffer ranges
- Keep the current batching-by-texture flow intact
- Maintain simple APIs and minimal invasive changes

## Terminology
- Ring Buffer: A fixed-size GPU buffer written in a circular fashion. When you reach the end, wrap to 0, ensuring you don’t overwrite data still in flight.
- Batch: Group of quads using the same texture, drawn with one bind+draw.

## Current Approach (Summary)
- CPU accumulates quads per texture during a frame
- At EndBatch, uploads per-batch vertex data (WriteBuffer) and issues one draw per batch
- Uses non-indexed quads (4 vertices per sprite)

## Easy Wins

### 1) Preallocate a Large GPU Vertex Ring Buffer
Allocate once, reuse every frame; write batched vertex data at a moving offset.

```go
// Pseudocode additions to your Canvas manager

// Constants – tune for your game scale
const (
    maxSpritesPerFrame   = 10_000
    bytesPerVertex       = 32 // pos.xy, uv.xy, color.rgba as float32 -> ~32 bytes
    vertsPerQuad         = 4
    ringBufferFrames     = 3   // triple-buffering headroom
)

// Derived size
var ringBufferSize = uint64(maxSpritesPerFrame * vertsPerQuad * bytesPerVertex * ringBufferFrames)

type WebGPUCanvasManager struct {
    // ... existing fields ...

    vertexRingBuffer   *wgpu.Buffer
    writeOffset        uint64 // advancing offset in bytes
    frameCursor        int    // 0..ringBufferFrames-1
}

func (w *WebGPUCanvasManager) Initialize(canvasID string) error {
    // ... existing init ...

    w.vertexRingBuffer = w.device.CreateBuffer(&wgpu.BufferDescriptor{
        Size:  ringBufferSize,
        Usage: wgpu.BufferUsage_CopyDst | wgpu.BufferUsage_Vertex,
        MappedAtCreation: false,
        Label: "sprite-vertex-ring",
    })

    // start at frame 0
    w.frameCursor = 0
    w.writeOffset = 0
    return nil
}
```

Notes:
- Overprovision: aim for peak per-frame vertex bytes × 2–3.
- Keep writeOffset 256-byte aligned (WebGPU best practice).

### 2) Upload Per-Texture Batches at Increasing Offsets
Instead of creating temporary buffers per batch, copy each batch’s vertex data into the ring at writeOffset, track metadata, then draw.

```go
// During EndBatch() pseudocode

// 1) finalize CPU batches (per texture) – you already have this
//    batches := []Batch { {texturePath, vertices []byte}, ... }

var playback []struct{
    offset      uint64
    vertexCount uint32
    bindGroup   *wgpu.BindGroup
}

for _, b := range batches {
    // Ensure 256-byte alignment
    aligned := (w.writeOffset + 255) & ^uint64(255)
    w.writeOffset = aligned

    // Copy CPU vertex bytes into the ring buffer at w.writeOffset
    w.queue.WriteBuffer(w.vertexRingBuffer, w.writeOffset, b.vertices)

    playback = append(playback, struct{
        offset      uint64
        vertexCount uint32
        bindGroup   *wgpu.BindGroup
    }{
        offset:      w.writeOffset,
        vertexCount: uint32(len(b.vertices) / bytesPerVertex),
        bindGroup:   w.bindGroupForTexture(b.texturePath),
    })

    w.writeOffset += uint64(len(b.vertices))

    // Wrap if near end; advance frameCursor to avoid overlap
    if w.writeOffset+4096 > ringBufferSize { // small guard margin
        w.frameCursor = (w.frameCursor + 1) % ringBufferFrames
        w.writeOffset = 0
    }
}

// 2) Playback draws
pass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{ /* ... */ })
for _, p := range playback {
    pass.SetBindGroup(0, p.bindGroup, nil)
    pass.SetVertexBuffer(0, w.vertexRingBuffer, int64(p.offset), int64(p.vertexCount*bytesPerVertex))
    pass.Draw(p.vertexCount, 1, 0, 0)
}
pass.End()
```

### 3) Optional: Static Index Buffer for Quads
Keep using non-indexed draws if you like. If you prefer indexed draws, create one static index buffer once (6 indices per quad) sized for your max sprites, and reuse it every frame.

```go
// Build once at init
totalIndices := maxSpritesPerFrame * 6
indices := make([]uint16, totalIndices)
for i := 0; i < maxSpritesPerFrame; i++ {
    baseV := uint16(i * 4)
    baseI := i * 6
    indices[baseI+0] = baseV + 0
    indices[baseI+1] = baseV + 1
    indices[baseI+2] = baseV + 2
    indices[baseI+3] = baseV + 2
    indices[baseI+4] = baseV + 3
    indices[baseI+5] = baseV + 0
}

w.indexBuffer = w.device.CreateBufferInit(&wgpu.BufferInitDescriptor{
    Label:    "quad-index",
    Contents: bytesFrom(indices),
    Usage:    wgpu.BufferUsage_Index,
})
```

### 4) Triple Buffer Headroom
The ringBufferFrames constant reserves space for ~3 frames in flight. This reduces the chance of overwriting regions still being read by the GPU.

### 5) Texture Preloading (Avoid First-Use Hitches)
Call LoadTexture(path) for all assets needed by a scene during Scene.Initialize() or expose them via SceneTextureProvider. This avoids mid-frame lazy loads and unpredictable stalls.

### 6) Simple Stats (Visibility + Tuning)
Expose counters each frame:
- numBatches, numDrawCalls, numTexturesBound, vertexBytesUploaded
- Log or render in a small overlay to spot regressions quickly

### 7) Checklist for Correctness
- Align vertex writes to 256 bytes
- Wrap safely (guard + advance frame cursor)
- Don’t exceed maxSpritesPerFrame; detect and log if overflow would occur
- Reuse bind groups per texture path (cache once)

## FAQ

- Can one large vertex buffer handle many textures?
  Yes. The vertex data is texture-agnostic. You bind a texture per batch when drawing specific offset ranges.

- Do I need indices?
  Not strictly. Non-indexed is fine. Indices reduce vertex duplication slightly and can help cache efficiency; choose based on your pipeline and complexity tolerance.

- How big should the ring be?
  Compute peak vertices per frame × vertex stride × 2–3 frames. Start generous; log peak usage and adjust once.

## Minimal API Impacts
- No change to scene APIs
- Canvas keeps BeginBatch/EndBatch
- Internals switch to a ring buffer instead of per-batch ad-hoc buffers

## Example: Integration Steps
1) Add vertexRingBuffer, writeOffset, and frameCursor fields to the canvas manager.
2) Create the ring buffer in Initialize().
3) Keep batching on CPU per texture as today.
4) In EndBatch(), write each batch to the ring at writeOffset, record playback metadata, then draw.
5) Align writes, wrap safely, and bump the frame cursor when wrapping.
6) Add simple counters and texture preload during scene init.

That’s it—minimal code churn, better throughput, and fewer stalls.
