// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

#include <metal_stdlib>
using namespace metal;

kernel void add0(device const float* inA,
                 device const float* inB,
                 device float* out,
                 uint index [[thread_position_in_grid]]) {
    out[index] = inA[index] + inB[index];
}

kernel void sub0(device const float* inA,
                 device const float* inB,
                 device float* out,
                 uint index [[thread_position_in_grid]]) {
    out[index] = inA[index] - inB[index];
}

kernel void sqrt0(device const float* inA,
                 device float* out,
                 uint index [[thread_position_in_grid]]) {
    out[index] = sqrt(inA[index]);
}
