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
- OS: `Debian 12 (Bookworm)`
- Kernel: `6.1.0-14-amd64`
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
BenchmarkEuclideanDistance64_AVX/16-32         1000000000    2.537  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/32
BenchmarkEuclideanDistance64_AVX/32-32         1000000000    3.996  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/64
BenchmarkEuclideanDistance64_AVX/64-32         1000000000    7.067  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/128
BenchmarkEuclideanDistance64_AVX/128-32        1000000000    13.67  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/256
BenchmarkEuclideanDistance64_AVX/256-32        1000000000    28.02  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/512
BenchmarkEuclideanDistance64_AVX/512-32        549605646     65.24  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/1024
BenchmarkEuclideanDistance64_AVX/1024-32       261180774     136.0  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/2048
BenchmarkEuclideanDistance64_AVX/2048-32       127792701     284.7  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/4096
BenchmarkEuclideanDistance64_AVX/4096-32       59748481      585.7  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX/8192
BenchmarkEuclideanDistance64_AVX/8192-32       30999266      1151   ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512
BenchmarkEuclideanDistance64_AVX512/16
BenchmarkEuclideanDistance64_AVX512/16-32      1000000000    2.700  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/32
BenchmarkEuclideanDistance64_AVX512/32-32      1000000000    3.325  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/64
BenchmarkEuclideanDistance64_AVX512/64-32      1000000000    5.181  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/128
BenchmarkEuclideanDistance64_AVX512/128-32     1000000000    9.084  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/256
BenchmarkEuclideanDistance64_AVX512/256-32     1000000000    18.31  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/512
BenchmarkEuclideanDistance64_AVX512/512-32     1000000000    35.66  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/1024
BenchmarkEuclideanDistance64_AVX512/1024-32    483242805     76.16  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/2048
BenchmarkEuclideanDistance64_AVX512/2048-32    227822972     158.4  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/4096
BenchmarkEuclideanDistance64_AVX512/4096-32    86723244      392.6  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_AVX512/8192
BenchmarkEuclideanDistance64_AVX512/8192-32    44334475      775.3  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go
BenchmarkEuclideanDistance64_Go/16
BenchmarkEuclideanDistance64_Go/16-32          1000000000    4.141  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/32
BenchmarkEuclideanDistance64_Go/32-32          1000000000    8.574  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/64
BenchmarkEuclideanDistance64_Go/64-32          1000000000    16.84  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/128
BenchmarkEuclideanDistance64_Go/128-32         1000000000    28.91  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/256
BenchmarkEuclideanDistance64_Go/256-32         695258588     51.56  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/512
BenchmarkEuclideanDistance64_Go/512-32         357411787     100.8  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/1024
BenchmarkEuclideanDistance64_Go/1024-32        183526327     194.1  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/2048
BenchmarkEuclideanDistance64_Go/2048-32        89894823      387.7  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/4096
BenchmarkEuclideanDistance64_Go/4096-32        47002590      756.3  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_Go/8192
BenchmarkEuclideanDistance64_Go/8192-32        23957892      1532   ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow
BenchmarkEuclideanDistance64_GoSlow/16
BenchmarkEuclideanDistance64_GoSlow/16-32      215302400     169.8  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/32
BenchmarkEuclideanDistance64_GoSlow/32-32      100000000     337.5  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/64
BenchmarkEuclideanDistance64_GoSlow/64-32      50568763      670.9  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/128
BenchmarkEuclideanDistance64_GoSlow/128-32     26037364      1358   ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/256
BenchmarkEuclideanDistance64_GoSlow/256-32     13196414      2702   ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/512
BenchmarkEuclideanDistance64_GoSlow/512-32     6568803       5446   ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/1024
BenchmarkEuclideanDistance64_GoSlow/1024-32    3280534       10755  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/2048
BenchmarkEuclideanDistance64_GoSlow/2048-32    1645164       22083  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/4096
BenchmarkEuclideanDistance64_GoSlow/4096-32    778020        51135  ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance64_GoSlow/8192
BenchmarkEuclideanDistance64_GoSlow/8192-32    266604        138283 ns/op    0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX
BenchmarkEuclideanDistance32_AVX/16
BenchmarkEuclideanDistance32_AVX/16-32         1000000000    2.148 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/32
BenchmarkEuclideanDistance32_AVX/32-32         1000000000    2.741 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/64
BenchmarkEuclideanDistance32_AVX/64-32         1000000000    3.827 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/128
BenchmarkEuclideanDistance32_AVX/128-32        1000000000    7.012 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/256
BenchmarkEuclideanDistance32_AVX/256-32        1000000000    14.90 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/512
BenchmarkEuclideanDistance32_AVX/512-32        1000000000    27.96 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/1024
BenchmarkEuclideanDistance32_AVX/1024-32       562765047     65.52 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/2048
BenchmarkEuclideanDistance32_AVX/2048-32       264678050     135.7 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/4096
BenchmarkEuclideanDistance32_AVX/4096-32       128494790     285.0 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX/8192
BenchmarkEuclideanDistance32_AVX/8192-32       63154987      585.1 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512
BenchmarkEuclideanDistance32_AVX512/16
BenchmarkEuclideanDistance32_AVX512/16-32      1000000000    2.391 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/32
BenchmarkEuclideanDistance32_AVX512/32-32      1000000000    2.587 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/64
BenchmarkEuclideanDistance32_AVX512/64-32      1000000000    3.337 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/128
BenchmarkEuclideanDistance32_AVX512/128-32     1000000000    5.500 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/256
BenchmarkEuclideanDistance32_AVX512/256-32     1000000000    9.143 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/512
BenchmarkEuclideanDistance32_AVX512/512-32     1000000000    18.19 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/1024
BenchmarkEuclideanDistance32_AVX512/1024-32    1000000000    35.08 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/2048
BenchmarkEuclideanDistance32_AVX512/2048-32    492597610     74.31 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/4096
BenchmarkEuclideanDistance32_AVX512/4096-32    230532822     154.7 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_AVX512/8192
BenchmarkEuclideanDistance32_AVX512/8192-32    86509042      388.7 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go
BenchmarkEuclideanDistance32_Go/16
BenchmarkEuclideanDistance32_Go/16-32          1000000000    6.468 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/32
BenchmarkEuclideanDistance32_Go/32-32          1000000000    12.23 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/64
BenchmarkEuclideanDistance32_Go/64-32          1000000000    28.06 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/128
BenchmarkEuclideanDistance32_Go/128-32         729570974     47.46 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/256
BenchmarkEuclideanDistance32_Go/256-32         493884171     70.06 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/512
BenchmarkEuclideanDistance32_Go/512-32         306095514     117.0 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/1024
BenchmarkEuclideanDistance32_Go/1024-32        167661460     214.0 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/2048
BenchmarkEuclideanDistance32_Go/2048-32        85926724      403.1 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/4096
BenchmarkEuclideanDistance32_Go/4096-32        45109837      802.8 ns/op     0 B/op    0 allocs/op
BenchmarkEuclideanDistance32_Go/8192
BenchmarkEuclideanDistance32_Go/8192-32        22735119      1531  ns/op     0 B/op    0 allocs/op
```

### Normalize Vector

```
goos: linux
goarch: amd64
pkg: github.com/monetr/monetr/server/internal/calc
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkNormalizeVector64_AVX
BenchmarkNormalizeVector64_AVX/16
BenchmarkNormalizeVector64_AVX/16-32         1000000000    13.93 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/32
BenchmarkNormalizeVector64_AVX/32-32         1000000000    17.30 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/64
BenchmarkNormalizeVector64_AVX/64-32         1000000000    25.67 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/128
BenchmarkNormalizeVector64_AVX/128-32        874503805     42.15 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/256
BenchmarkNormalizeVector64_AVX/256-32        423851012     84.12 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/512
BenchmarkNormalizeVector64_AVX/512-32        203888568     179.1 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/1024
BenchmarkNormalizeVector64_AVX/1024-32       91693118      368.2 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/2048
BenchmarkNormalizeVector64_AVX/2048-32       46273400      758.4 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/4096
BenchmarkNormalizeVector64_AVX/4096-32       23155038      1513  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX/8192
BenchmarkNormalizeVector64_AVX/8192-32       11608892      3012  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512
BenchmarkNormalizeVector64_AVX512/16
BenchmarkNormalizeVector64_AVX512/16-32      1000000000    14.88 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/32
BenchmarkNormalizeVector64_AVX512/32-32      1000000000    18.37 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/64
BenchmarkNormalizeVector64_AVX512/64-32      1000000000    24.87 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/128
BenchmarkNormalizeVector64_AVX512/128-32     926581855     39.15 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/256
BenchmarkNormalizeVector64_AVX512/256-32     536417649     65.83 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/512
BenchmarkNormalizeVector64_AVX512/512-32     270329512     133.8 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/1024
BenchmarkNormalizeVector64_AVX512/1024-32    129711741     274.2 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/2048
BenchmarkNormalizeVector64_AVX512/2048-32    65699619      558.4 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/4096
BenchmarkNormalizeVector64_AVX512/4096-32    31166433      1121  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_AVX512/8192
BenchmarkNormalizeVector64_AVX512/8192-32    15749985      2232  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go
BenchmarkNormalizeVector64_Go/16
BenchmarkNormalizeVector64_Go/16-32          1000000000    23.37 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/32
BenchmarkNormalizeVector64_Go/32-32          932246338     38.60 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/64
BenchmarkNormalizeVector64_Go/64-32          463032544     76.74 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/128
BenchmarkNormalizeVector64_Go/128-32         213150487     172.0 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/256
BenchmarkNormalizeVector64_Go/256-32         98082843      354.7 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/512
BenchmarkNormalizeVector64_Go/512-32         47007858      730.8 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/1024
BenchmarkNormalizeVector64_Go/1024-32        23980852      1506  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/2048
BenchmarkNormalizeVector64_Go/2048-32        12121506      3020  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/4096
BenchmarkNormalizeVector64_Go/4096-32        5914047       6013  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector64_Go/8192
BenchmarkNormalizeVector64_Go/8192-32        3015862       12105 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX
BenchmarkNormalizeVector32_AVX/16
BenchmarkNormalizeVector32_AVX/16-32         1000000000    10.93 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/32
BenchmarkNormalizeVector32_AVX/32-32         1000000000    12.18 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/64
BenchmarkNormalizeVector32_AVX/64-32         1000000000    14.88 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/128
BenchmarkNormalizeVector32_AVX/128-32        1000000000    20.50 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/256
BenchmarkNormalizeVector32_AVX/256-32        1000000000    30.02 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/512
BenchmarkNormalizeVector32_AVX/512-32        574852094     62.36 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/1024
BenchmarkNormalizeVector32_AVX/1024-32       268132621     132.2 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/2048
BenchmarkNormalizeVector32_AVX/2048-32       129321574     279.9 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/4096
BenchmarkNormalizeVector32_AVX/4096-32       61041141      565.5 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX/8192
BenchmarkNormalizeVector32_AVX/8192-32       31572360      1133  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512
BenchmarkNormalizeVector32_AVX512/16
BenchmarkNormalizeVector32_AVX512/16-32      1000000000    12.60 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/32
BenchmarkNormalizeVector32_AVX512/32-32      1000000000    13.96 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/64
BenchmarkNormalizeVector32_AVX512/64-32      1000000000    16.33 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/128
BenchmarkNormalizeVector32_AVX512/128-32     1000000000    20.34 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/256
BenchmarkNormalizeVector32_AVX512/256-32     1000000000    29.61 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/512
BenchmarkNormalizeVector32_AVX512/512-32     721470657     49.91 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/1024
BenchmarkNormalizeVector32_AVX512/1024-32    388907968     92.68 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/2048
BenchmarkNormalizeVector32_AVX512/2048-32    177967101     204.4 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/4096
BenchmarkNormalizeVector32_AVX512/4096-32    84550104      422.0 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_AVX512/8192
BenchmarkNormalizeVector32_AVX512/8192-32    41956521      854.0 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go
BenchmarkNormalizeVector32_Go/16
BenchmarkNormalizeVector32_Go/16-32          1000000000    17.04 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/32
BenchmarkNormalizeVector32_Go/32-32          1000000000    27.36 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/64
BenchmarkNormalizeVector32_Go/64-32          606857464     59.23 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/128
BenchmarkNormalizeVector32_Go/128-32         273112796     129.8 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/256
BenchmarkNormalizeVector32_Go/256-32         130480930     269.8 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/512
BenchmarkNormalizeVector32_Go/512-32         64203439      552.0 ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/1024
BenchmarkNormalizeVector32_Go/1024-32        31811756      1116  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/2048
BenchmarkNormalizeVector32_Go/2048-32        15772390      2274  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/4096
BenchmarkNormalizeVector32_Go/4096-32        7835325       4527  ns/op    0 B/op    0 allocs/op
BenchmarkNormalizeVector32_Go/8192
BenchmarkNormalizeVector32_Go/8192-32        3936650       9002  ns/op    0 B/op    0 allocs/op
```
