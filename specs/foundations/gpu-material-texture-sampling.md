---
title: "GPU material texture sampling + seam option B (forward->deferred, no CPU round-trip)"
status: NOT STARTED -- successor brick to gpu-forward-raster.md; deferred by the user 2026-07-01
depends_on:
  - foundations/gpu-forward-raster.md
affects:
  - render
  - gpu
  - buffer
created: 2026-07-01
author: changkun
dispatched_task_id: null
---

# GPU material texture sampling + seam option B

## Why this exists as its own brick

`gpu-forward-raster.md` (brick 3b) sketched a "step 3": *remove the round-trip (seam
option B) -- keep the G-buffer on GPU textures into the deferred pass, no CPU
FragmentBuffer round-trip.* That step cannot live in the rasterizer brick because it
depends on an item that brick lists as **out of scope**: GPU-side material texture
sampling. It is split out here.

The rasterizer brick is complete: the GPU forward pass rasterizes the full G-buffer
(world position, normal, uv + du/dv, material id, depth) on the GPU by default on both
GL and Metal, gated by measured parity. What remains is the *seam* between forward and
deferred.

## The round-trip today (seam A)

1. `gpuForwardPass` rasterizes the G-buffer into GPU **textures** (RGBA32F MRT + depth).
2. It **reads those textures back** to the CPU `*buffer.FragmentBuffer` (`buf.Set`).
3. `passDeferred` -> `gpuDeferredShade(dev, buf, ...)` reads the CPU buffer and
   **re-uploads** normals/worldpos/basecol/matidx to GPU **storage buffers**, then runs
   the Blinn-Phong lighting compute kernel.

So the G-buffer makes a GPU-textures -> CPU -> GPU-buffers round-trip.

## Why seam B is blocked on GPU texture sampling

The "GPU deferred" pass is a **hybrid**, not fully on the GPU:

- Material `basecol` is sampled on the **CPU**: `gpudeferred.go`,
  `bp.Texture.Query(lod, info.U, 1-info.V)`, using the FragmentBuffer's `U/V/Du/Dv`.
- Only the Blinn-Phong lighting math runs in the compute kernel, and that kernel reads
  storage **buffers**, not texture samplers.

To keep the G-buffer on the GPU into the deferred pass (seam B), the deferred pass must
stop reading the CPU FragmentBuffer. But `basecol` needs `U/V/Du/Dv` on the CPU to call
`Query`. So for any **textured** material, seam B forces material texture sampling onto
the GPU: upload each material's mipmap pyramid, and sample it in a kernel with LOD +
bilinear filtering that matches `buffer.Texture.Query` for parity. That is a whole
subsystem -- exactly the "texture-sampled materials in the GPU G-buffer" item that
`gpu-forward-raster.md` lists as out of scope.

For **flat** (non-textured) materials `basecol = bp.Diffuse` (already a constant in the
`materials[]` table), so seam B is achievable without any texture sampling. But the
main test scenes (bunny etc.) are textured, so a flat-only seam B leaves the round-trip
in place for the scenes that actually matter.

## Scope of this brick

1. **GPU material texture sampling with parity.** Upload material texture mipmaps to the
   GPU (textures or storage buffers); add sampler/texture support to the kernel DSL
   (`shader` / `gpukernel.go`); sample with LOD + bilinear filtering matched to
   `buffer.Texture.Query` (same mipmap generation, same LOD formula
   `log2(texSize * sqrt(max(Du,Dv)))`, same `1-V` flip). Gate parity by measurement,
   like the deferred/forward bricks (this is parity-sensitive -- the forward arc's
   hardest bugs were all convention/parity mismatches).
2. **Seam B.** Add a deferred entry that consumes the forward pass's GPU G-buffer
   (textures or buffers) directly -- normals, worldpos, uv, matid -- with `basecol`
   sampled on the GPU (step 1), removing the CPU FragmentBuffer round-trip. Keep the
   CPU path (seam A) as the fallback.
3. **GPU render -> present with no CPU round-trip.** With seam B, the whole forward ->
   deferred -> present pipeline stays on the GPU, the arc's north star.

## Parity risk (why this is a multi-session arc, not a quick step)

Matching `buffer.Texture.Query` bit-closely on the GPU is the crux. The CPU builds the
mipmap pyramid a specific way (`imageutil.Resize` per level) and selects LOD + filters
with specific rounding. The forward-raster arc showed that GPU-vs-CPU convention
mismatches (Y-flip, clip-z range, winding, perspective-vs-linear interp) each produce
large, confusing divergences that only a measure-then-attribute loop untangles. Expect
the same here: measure the GPU-vs-CPU sampled-color delta first, attribute it, then
gate with a measured tolerance -- do not assume exactness.

## Out of scope

- MSAA on the GPU raster (inherited from the rasterizer brick).
- Anisotropic filtering (the CPU path is isotropic trilinear-ish via `Query`).
