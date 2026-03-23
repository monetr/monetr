# Calc

This package contains basic math operations that monetr performs. These math operations specifically have been
implemented in assembly as well as pure go in order to leverage SIMD features.

## Euclidean Distance

### AVX512

At the end of the euclidean distance code we have to get the sum of the 8 64-bit values in the ZMM0 register. This is
done using the following sequence of instructions:

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

This can be improved, as we actually don't even need to use SIMD registers after the last VPERMF128, as the first value
in each register can be added individually at that point instead of using SIMD.

## Benchmarks

All benchmarks below were run using:

```shell
go test github.com/monetr/monetr/server/internal/calc -bench=${BENCHMARK_NAME} -benchmem -run=$\^ -benchtime=30s -v -timeout=900m
```

**System:**
- OS: `Debian 13 (Trixie)`
- Kernel: `6.12.74+deb13+1-amd64 `
- CPU: `AMD Ryzen 9 7950X`
- Memory: `64GB G.Skill F5-5600J3636D32GX2-TZ5RK`

### Euclidean Distance

```
goos: linux
goarch: amd64
pkg: github.com/monetr/monetr/server/internal/calc
cpu: AMD Ryzen 9 7950X 16-Core Processor            
BenchmarkEuclideanDistance64_AVX
BenchmarkEuclideanDistance64_AVX/16
BenchmarkEuclideanDistance64_AVX/16-32                 1000000000          2.552 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/32
BenchmarkEuclideanDistance64_AVX/32-32                 1000000000          3.365 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/64
BenchmarkEuclideanDistance64_AVX/64-32                 1000000000          6.204 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/128
BenchmarkEuclideanDistance64_AVX/128-32                1000000000          10.95 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/256
BenchmarkEuclideanDistance64_AVX/256-32                1000000000          19.91 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/512
BenchmarkEuclideanDistance64_AVX/512-32                960981878           37.53 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/1024
BenchmarkEuclideanDistance64_AVX/1024-32               490452626           73.47 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/2048
BenchmarkEuclideanDistance64_AVX/2048-32               244265064           147.4 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/4096
BenchmarkEuclideanDistance64_AVX/4096-32               92691448            387.4 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX/8192
BenchmarkEuclideanDistance64_AVX/8192-32               46915230            770.7 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA
BenchmarkEuclideanDistance64_AVX_FMA/16
BenchmarkEuclideanDistance64_AVX_FMA/16-32             1000000000          2.497 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/32
BenchmarkEuclideanDistance64_AVX_FMA/32-32             1000000000          3.312 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/64
BenchmarkEuclideanDistance64_AVX_FMA/64-32             1000000000          5.248 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/128
BenchmarkEuclideanDistance64_AVX_FMA/128-32            1000000000          10.04 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/256
BenchmarkEuclideanDistance64_AVX_FMA/256-32            1000000000          20.19 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/512
BenchmarkEuclideanDistance64_AVX_FMA/512-32            847858956           42.56 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/1024
BenchmarkEuclideanDistance64_AVX_FMA/1024-32           401934748           89.52 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/2048
BenchmarkEuclideanDistance64_AVX_FMA/2048-32           194943972           185.1 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/4096
BenchmarkEuclideanDistance64_AVX_FMA/4096-32           92174170            388.9 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX_FMA/8192
BenchmarkEuclideanDistance64_AVX_FMA/8192-32           46557018            771.7 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX
BenchmarkEuclideanDistance32_AVX/32
BenchmarkEuclideanDistance32_AVX/32-32                 1000000000          2.810 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/64
BenchmarkEuclideanDistance32_AVX/64-32                 1000000000          3.769 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/128
BenchmarkEuclideanDistance32_AVX/128-32                1000000000          6.372 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/256
BenchmarkEuclideanDistance32_AVX/256-32                1000000000          11.58 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/512
BenchmarkEuclideanDistance32_AVX/512-32                1000000000          20.86 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/1024
BenchmarkEuclideanDistance32_AVX/1024-32               940851178           38.29 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/2048
BenchmarkEuclideanDistance32_AVX/2048-32               485386914           74.10 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/4096
BenchmarkEuclideanDistance32_AVX/4096-32               246143260           146.4 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX/8192
BenchmarkEuclideanDistance32_AVX/8192-32               92926340            389.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA
BenchmarkEuclideanDistance32_AVX_FMA/32
BenchmarkEuclideanDistance32_AVX_FMA/32-32             1000000000          2.815 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/64
BenchmarkEuclideanDistance32_AVX_FMA/64-32             1000000000          3.700 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/128
BenchmarkEuclideanDistance32_AVX_FMA/128-32            1000000000          5.953 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/256
BenchmarkEuclideanDistance32_AVX_FMA/256-32            1000000000          10.63 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/512
BenchmarkEuclideanDistance32_AVX_FMA/512-32            1000000000          21.06 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/1024
BenchmarkEuclideanDistance32_AVX_FMA/1024-32           827681122           43.50 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/2048
BenchmarkEuclideanDistance32_AVX_FMA/2048-32           398271775           90.52 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/4096
BenchmarkEuclideanDistance32_AVX_FMA/4096-32           192206140           187.4 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX_FMA/8192
BenchmarkEuclideanDistance32_AVX_FMA/8192-32           92527422            388.6 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512
BenchmarkEuclideanDistance64_AVX512/16
BenchmarkEuclideanDistance64_AVX512/16-32              1000000000          2.678 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/32
BenchmarkEuclideanDistance64_AVX512/32-32              1000000000          14.07 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/64
BenchmarkEuclideanDistance64_AVX512/64-32              1000000000          5.409 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/128
BenchmarkEuclideanDistance64_AVX512/128-32             1000000000          15.48 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/256
BenchmarkEuclideanDistance64_AVX512/256-32             1000000000          15.83 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/512
BenchmarkEuclideanDistance64_AVX512/512-32             1000000000          28.46 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/1024
BenchmarkEuclideanDistance64_AVX512/1024-32            665821790           53.69 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/2048
BenchmarkEuclideanDistance64_AVX512/2048-32            324229441           110.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/4096
BenchmarkEuclideanDistance64_AVX512/4096-32            95171127            378.9 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512/8192
BenchmarkEuclideanDistance64_AVX512/8192-32            47018046            760.9 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA
BenchmarkEuclideanDistance64_AVX512_FMA/16
BenchmarkEuclideanDistance64_AVX512_FMA/16-32          1000000000          2.677 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/32
BenchmarkEuclideanDistance64_AVX512_FMA/32-32          1000000000          3.299 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/64
BenchmarkEuclideanDistance64_AVX512_FMA/64-32          1000000000          4.714 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/128
BenchmarkEuclideanDistance64_AVX512_FMA/128-32         1000000000          8.160 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/256
BenchmarkEuclideanDistance64_AVX512_FMA/256-32         1000000000          14.32 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/512
BenchmarkEuclideanDistance64_AVX512_FMA/512-32         1000000000          27.26 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/1024
BenchmarkEuclideanDistance64_AVX512_FMA/1024-32        676067337           53.28 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/2048
BenchmarkEuclideanDistance64_AVX512_FMA/2048-32        330780349           109.1 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/4096
BenchmarkEuclideanDistance64_AVX512_FMA/4096-32        95773225            376.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_AVX512_FMA/8192
BenchmarkEuclideanDistance64_AVX512_FMA/8192-32        47522134            771.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512
BenchmarkEuclideanDistance32_AVX512/32
BenchmarkEuclideanDistance32_AVX512/32-32              1000000000          2.921 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/64
BenchmarkEuclideanDistance32_AVX512/64-32              1000000000          3.700 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/128
BenchmarkEuclideanDistance32_AVX512/128-32             1000000000          5.488 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/256
BenchmarkEuclideanDistance32_AVX512/256-32             1000000000          8.976 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/512
BenchmarkEuclideanDistance32_AVX512/512-32             1000000000          16.45 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/1024
BenchmarkEuclideanDistance32_AVX512/1024-32            1000000000          28.85 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/2048
BenchmarkEuclideanDistance32_AVX512/2048-32            666612703           53.99 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/4096
BenchmarkEuclideanDistance32_AVX512/4096-32            328500315           109.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512/8192
BenchmarkEuclideanDistance32_AVX512/8192-32            95462151            377.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA
BenchmarkEuclideanDistance32_AVX512_FMA/32
BenchmarkEuclideanDistance32_AVX512_FMA/32-32          1000000000          2.868 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/64
BenchmarkEuclideanDistance32_AVX512_FMA/64-32          1000000000          3.558 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/128
BenchmarkEuclideanDistance32_AVX512_FMA/128-32         1000000000          5.184 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/256
BenchmarkEuclideanDistance32_AVX512_FMA/256-32         1000000000          8.473 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/512
BenchmarkEuclideanDistance32_AVX512_FMA/512-32         1000000000          14.79 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/1024
BenchmarkEuclideanDistance32_AVX512_FMA/1024-32        1000000000          27.61 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/2048
BenchmarkEuclideanDistance32_AVX512_FMA/2048-32        669811491           53.61 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/4096
BenchmarkEuclideanDistance32_AVX512_FMA/4096-32        323506220           111.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_AVX512_FMA/8192
BenchmarkEuclideanDistance32_AVX512_FMA/8192-32        95016112            377.0 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go
BenchmarkEuclideanDistance64_Go/16
BenchmarkEuclideanDistance64_Go/16-32                  1000000000          3.744 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/32
BenchmarkEuclideanDistance64_Go/32-32                  1000000000          6.723 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/64
BenchmarkEuclideanDistance64_Go/64-32                  1000000000          12.58 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/128
BenchmarkEuclideanDistance64_Go/128-32                 1000000000          24.30 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/256
BenchmarkEuclideanDistance64_Go/256-32                 712961875           50.08 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/512
BenchmarkEuclideanDistance64_Go/512-32                 375392373           96.56 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/1024
BenchmarkEuclideanDistance64_Go/1024-32                188710843           190.0 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/2048
BenchmarkEuclideanDistance64_Go/2048-32                96855597            369.9 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/4096
BenchmarkEuclideanDistance64_Go/4096-32                49012028            742.7 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_Go/8192
BenchmarkEuclideanDistance64_Go/8192-32                24605997            1489 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow
BenchmarkEuclideanDistance64_GoSlow/16
BenchmarkEuclideanDistance64_GoSlow/16-32              255121204           142.2 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/32
BenchmarkEuclideanDistance64_GoSlow/32-32              127402603           282.8 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/64
BenchmarkEuclideanDistance64_GoSlow/64-32              64123255            561.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/128
BenchmarkEuclideanDistance64_GoSlow/128-32             32066616            1124 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/256
BenchmarkEuclideanDistance64_GoSlow/256-32             15777607            2245 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/512
BenchmarkEuclideanDistance64_GoSlow/512-32             8024150             4488 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/1024
BenchmarkEuclideanDistance64_GoSlow/1024-32            4027728             8992 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/2048
BenchmarkEuclideanDistance64_GoSlow/2048-32            2010822             17899 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/4096
BenchmarkEuclideanDistance64_GoSlow/4096-32            806524              39304 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/8192
BenchmarkEuclideanDistance64_GoSlow/8192-32            307816              113505 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go
BenchmarkEuclideanDistance32_Go/16
BenchmarkEuclideanDistance32_Go/16-32                  1000000000          3.753 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/32
BenchmarkEuclideanDistance32_Go/32-32                  1000000000          6.766 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/64
BenchmarkEuclideanDistance32_Go/64-32                  1000000000          12.53 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/128
BenchmarkEuclideanDistance32_Go/128-32                 1000000000          24.51 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/256
BenchmarkEuclideanDistance32_Go/256-32                 712866818           50.50 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/512
BenchmarkEuclideanDistance32_Go/512-32                 369870501           97.12 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/1024
BenchmarkEuclideanDistance32_Go/1024-32                188845203           190.9 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/2048
BenchmarkEuclideanDistance32_Go/2048-32                93987160            381.8 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/4096
BenchmarkEuclideanDistance32_Go/4096-32                46151284            754.5 ns/op        0 B/op        0 allocs/op
BenchmarkEuclideanDistance32_Go/8192
BenchmarkEuclideanDistance32_Go/8192-32                23836549            1494 ns/op        0 B/op        0 allocs/op
```

### Normalize Vector

```
goos: linux
goarch: amd64
pkg: github.com/monetr/monetr/server/internal/calc
cpu: AMD Ryzen 9 7950X 16-Core Processor            
BenchmarkNormalizeVector64_AVX
BenchmarkNormalizeVector64_AVX/16
BenchmarkNormalizeVector64_AVX/16-32          1000000000         14.38 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/32
BenchmarkNormalizeVector64_AVX/32-32          1000000000         15.60 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/64
BenchmarkNormalizeVector64_AVX/64-32          1000000000         17.85 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/128
BenchmarkNormalizeVector64_AVX/128-32         1000000000         23.23 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/256
BenchmarkNormalizeVector64_AVX/256-32         1000000000         35.67 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/512
BenchmarkNormalizeVector64_AVX/512-32         541047789         66.64 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/1024
BenchmarkNormalizeVector64_AVX/1024-32        286022740        126.0 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/2048
BenchmarkNormalizeVector64_AVX/2048-32        147407353        244.0 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/4096
BenchmarkNormalizeVector64_AVX/4096-32        73944530        487.3 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX/8192
BenchmarkNormalizeVector64_AVX/8192-32        37142563        969.9 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA
BenchmarkNormalizeVector64_AVX_FMA/16
BenchmarkNormalizeVector64_AVX_FMA/16-32      1000000000         14.05 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/32
BenchmarkNormalizeVector64_AVX_FMA/32-32      1000000000         15.51 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/64
BenchmarkNormalizeVector64_AVX_FMA/64-32      1000000000         18.63 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/128
BenchmarkNormalizeVector64_AVX_FMA/128-32     1000000000         25.32 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/256
BenchmarkNormalizeVector64_AVX_FMA/256-32     736224198         48.87 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/512
BenchmarkNormalizeVector64_AVX_FMA/512-32     375531165         95.80 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/1024
BenchmarkNormalizeVector64_AVX_FMA/1024-32    189715005        189.8 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/2048
BenchmarkNormalizeVector64_AVX_FMA/2048-32    95427960        376.7 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/4096
BenchmarkNormalizeVector64_AVX_FMA/4096-32    47401306        757.7 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX_FMA/8192
BenchmarkNormalizeVector64_AVX_FMA/8192-32    23770020       1514 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512
BenchmarkNormalizeVector64_AVX512/16
BenchmarkNormalizeVector64_AVX512/16-32       1000000000         15.77 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/32
BenchmarkNormalizeVector64_AVX512/32-32       1000000000         16.63 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/64
BenchmarkNormalizeVector64_AVX512/64-32       1000000000         18.32 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/128
BenchmarkNormalizeVector64_AVX512/128-32      1000000000         22.43 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/256
BenchmarkNormalizeVector64_AVX512/256-32      1000000000         31.56 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/512
BenchmarkNormalizeVector64_AVX512/512-32      655420219         54.87 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/1024
BenchmarkNormalizeVector64_AVX512/1024-32     360875760         99.95 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/2048
BenchmarkNormalizeVector64_AVX512/2048-32     191194354        188.2 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/4096
BenchmarkNormalizeVector64_AVX512/4096-32     98468160        366.4 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512/8192
BenchmarkNormalizeVector64_AVX512/8192-32     45701610        788.6 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA
BenchmarkNormalizeVector64_AVX512_FMA/16
BenchmarkNormalizeVector64_AVX512_FMA/16-32   1000000000         15.78 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/32
BenchmarkNormalizeVector64_AVX512_FMA/32-32   1000000000         16.64 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/64
BenchmarkNormalizeVector64_AVX512_FMA/64-32   1000000000         18.65 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/128
BenchmarkNormalizeVector64_AVX512_FMA/128-32           1000000000         23.56 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/256
BenchmarkNormalizeVector64_AVX512_FMA/256-32           1000000000         34.32 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/512
BenchmarkNormalizeVector64_AVX512_FMA/512-32           622773775         57.89 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/1024
BenchmarkNormalizeVector64_AVX512_FMA/1024-32          343213983        104.8 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/2048
BenchmarkNormalizeVector64_AVX512_FMA/2048-32          180687567        199.6 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/4096
BenchmarkNormalizeVector64_AVX512_FMA/4096-32          91346121        395.2 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_AVX512_FMA/8192
BenchmarkNormalizeVector64_AVX512_FMA/8192-32          44797147        791.5 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX
BenchmarkNormalizeVector32_AVX/32
BenchmarkNormalizeVector32_AVX/32-32                   1000000000         13.47 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/64
BenchmarkNormalizeVector32_AVX/64-32                   1000000000         14.77 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/128
BenchmarkNormalizeVector32_AVX/128-32                  1000000000         17.34 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/256
BenchmarkNormalizeVector32_AVX/256-32                  1000000000         22.52 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/512
BenchmarkNormalizeVector32_AVX/512-32                  1000000000         34.56 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/1024
BenchmarkNormalizeVector32_AVX/1024-32                 551703820         65.28 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/2048
BenchmarkNormalizeVector32_AVX/2048-32                 287625228        125.2 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/4096
BenchmarkNormalizeVector32_AVX/4096-32                 147440902        244.0 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX/8192
BenchmarkNormalizeVector32_AVX/8192-32                 73800518        490.6 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA
BenchmarkNormalizeVector32_AVX_FMA/32
BenchmarkNormalizeVector32_AVX_FMA/32-32               1000000000         13.31 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/64
BenchmarkNormalizeVector32_AVX_FMA/64-32               1000000000         14.96 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/128
BenchmarkNormalizeVector32_AVX_FMA/128-32              1000000000         17.86 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/256
BenchmarkNormalizeVector32_AVX_FMA/256-32              1000000000         24.44 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/512
BenchmarkNormalizeVector32_AVX_FMA/512-32              738019984         48.77 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/1024
BenchmarkNormalizeVector32_AVX_FMA/1024-32             375357504         95.77 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/2048
BenchmarkNormalizeVector32_AVX_FMA/2048-32             188947520        190.7 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/4096
BenchmarkNormalizeVector32_AVX_FMA/4096-32             94544977        378.5 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX_FMA/8192
BenchmarkNormalizeVector32_AVX_FMA/8192-32             47267839        764.3 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512
BenchmarkNormalizeVector32_AVX512/32
BenchmarkNormalizeVector32_AVX512/32-32                1000000000         15.50 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/64
BenchmarkNormalizeVector32_AVX512/64-32                1000000000         16.44 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/128
BenchmarkNormalizeVector32_AVX512/128-32               1000000000         17.74 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/256
BenchmarkNormalizeVector32_AVX512/256-32               1000000000         21.91 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/512
BenchmarkNormalizeVector32_AVX512/512-32               1000000000         30.89 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/1024
BenchmarkNormalizeVector32_AVX512/1024-32              666891607         54.08 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/2048
BenchmarkNormalizeVector32_AVX512/2048-32              363662887         98.93 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/4096
BenchmarkNormalizeVector32_AVX512/4096-32              192229987        187.3 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512/8192
BenchmarkNormalizeVector32_AVX512/8192-32              98854729        364.5 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA
BenchmarkNormalizeVector32_AVX512_FMA/32
BenchmarkNormalizeVector32_AVX512_FMA/32-32            1000000000         14.91 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/64
BenchmarkNormalizeVector32_AVX512_FMA/64-32            1000000000         15.72 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/128
BenchmarkNormalizeVector32_AVX512_FMA/128-32           1000000000         17.43 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/256
BenchmarkNormalizeVector32_AVX512_FMA/256-32           1000000000         22.49 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/512
BenchmarkNormalizeVector32_AVX512_FMA/512-32           1000000000         32.70 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/1024
BenchmarkNormalizeVector32_AVX512_FMA/1024-32          641099390         56.19 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/2048
BenchmarkNormalizeVector32_AVX512_FMA/2048-32          346612978        103.9 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/4096
BenchmarkNormalizeVector32_AVX512_FMA/4096-32          181699172        198.0 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_AVX512_FMA/8192
BenchmarkNormalizeVector32_AVX512_FMA/8192-32          91928882        392.1 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go
BenchmarkNormalizeVector64_Go/16
BenchmarkNormalizeVector64_Go/16-32                    1000000000         22.90 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/32
BenchmarkNormalizeVector64_Go/32-32                    934608193         38.51 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/64
BenchmarkNormalizeVector64_Go/64-32                    453039915         79.61 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/128
BenchmarkNormalizeVector64_Go/128-32                   208425565        172.9 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/256
BenchmarkNormalizeVector64_Go/256-32                   100000000        359.0 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/512
BenchmarkNormalizeVector64_Go/512-32                   49152573        731.0 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/1024
BenchmarkNormalizeVector64_Go/1024-32                  24338930       1475 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/2048
BenchmarkNormalizeVector64_Go/2048-32                  12140846       2961 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/4096
BenchmarkNormalizeVector64_Go/4096-32                  6075872       5944 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector64_Go/8192
BenchmarkNormalizeVector64_Go/8192-32                  3020638      11893 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go
BenchmarkNormalizeVector32_Go/16
BenchmarkNormalizeVector32_Go/16-32                    1000000000         16.89 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/32
BenchmarkNormalizeVector32_Go/32-32                    1000000000         28.79 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/64
BenchmarkNormalizeVector32_Go/64-32                    610343859         58.99 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/128
BenchmarkNormalizeVector32_Go/128-32                   278532180        129.1 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/256
BenchmarkNormalizeVector32_Go/256-32                   133761294        268.8 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/512
BenchmarkNormalizeVector32_Go/512-32                   65514152        548.6 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/1024
BenchmarkNormalizeVector32_Go/1024-32                  32447563       1112 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/2048
BenchmarkNormalizeVector32_Go/2048-32                  16042054       2243 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/4096
BenchmarkNormalizeVector32_Go/4096-32                  8065782       4437 ns/op        0 B/op        0 allocs/op
BenchmarkNormalizeVector32_Go/8192
BenchmarkNormalizeVector32_Go/8192-32                  4000630       8952 ns/op        0 B/op        0 allocs/op
```
