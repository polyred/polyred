// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

#version 310 es

layout(local_size_x = 128) in;
layout(std430) buffer;
layout(binding = 0) writeonly buffer Output {
    vec4 data[];
} outC;
layout(binding = 1) readonly buffer Input0 {
    vec4 data[];
} inA;
layout(binding = 2) readonly buffer Input1 {
    vec4 data[];
} inB;

void main() {
    uint ident = gl_GlobalInvocationID.x;
    outC.data[ident] = inA.data[ident] + inB.data[ident];
}