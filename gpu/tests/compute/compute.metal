#include <metal_stdlib>
using namespace metal;

typedef struct Params {
    int w_in, h_in;
    int w_out, h_out;
} Params;

int idx(int x, int y, int z, int w, int h) {
  int i = z * w * h;
  i += y * w;
  i += x;
  return i;
}

kernel void main0(
    device const Params* p,
    device const float*  input,
    device float*        output,
    uint3                gridSize[[threads_per_grid]],
    uint3                gid[[thread_position_in_grid]]) {
    if(gid.x != 0) {
        return;
    }

    int input_index = idx(gid.x, gid.y, gid.z, p->w_in, p->h_in);

    float a = input[input_index];
    float b = input[input_index+1];

    int output_index = idx(0, gid.y, 0, p->w_out, p->h_out);

    output[output_index] = a + b;
}