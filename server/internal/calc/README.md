# Calc

This package contains basic math operations that monetr performs. These math operations specifically have been implemented in assembly as well as pure go in order to leverage SIMD features.

## Euclidean Distance

### AVX512

At the end of the euclidean distance code we have to get the sum of the 8 64-bit values in the ZMM0 register. This is done using the following sequence of instructions:

|               | 64 bit | 64 bit | 64 bit | 64 bit | 64 bit | 64 bit | 64 bit | 64 bit |
| ------------- | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ |
| ZMM0          | 1      | 2      | 3      | 4      | 5      | 6      | 7      | 8      |
| VEXTRACTF64X4 |        |        |        |        |        |        |        |        |
| YMM0          | 5      | 6      | 7      | 8      |        |        |        |        |
| VHADDPD       | 1+2    | 1+2    | 3+4    | 3+4    |        |        |        |        |
| YMM0          | 3      | 3      | 7      | 7      |        |        |        |        |
| VPERM2F128    |        |        |        |        |        |        |        |        |
| YMM2          | 7      | 7      | ?      | ?      |        |        |        |        |
| VHADDPD       | 5+6    | 5+6    | 7+8    | 7+8    |        |        |        |        |
| YMM1          | 11     | 11     | 15     | 15     |        |        |        |        |
| VPERM2F128    |        |        |        |        |        |        |        |        |
| YMM3          | 15     | 15     |        |        |        |        |        |        |
|               |        |        |        |        |        |        |        |        |
| YMM0          | 3      | 3      |        |        |        |        |        |        |
| YMM1          | 11     | 11     |        |        |        |        |        |        |
| YMM2          | 7      | 7      |        |        |        |        |        |        |
| YMM3          | 15     | 15     |        |        |        |        |        |        |
| VADDPD        | 3+11   | 3+11   |        |        |        |        |        |        |
| YMM0          | 14     | 14     |        |        |        |        |        |        |
| VADDPD        | 14+7   | 14+7   |        |        |        |        |        |        |
| YMM0          | 21     | 21     |        |        |        |        |        |        |
| VADDPD        | 15+21  | 15+21  |        |        |        |        |        |        |
| YMM0          | 36     | 36     |        |        |        |        |        |        |
| RET           | 36     |        |        |        |        |        |        |        |

This can be improved, as we actually don't even need to use SIMD registers after the last VPERMF128, as the first value in each register can be added individually at that point instead of using SIMD.
