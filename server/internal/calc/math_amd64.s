#include "textflag.h"

// Coeficients used by the sine calculations.
DATA  CalcSineCoef+0(SB)/8,  $0x3de5d8fd1fd19ccd
DATA  CalcSineCoef+8(SB)/8,  $0xbe5ae5e5a9291f5d
DATA  CalcSineCoef+16(SB)/8, $0x3ec71de3567d48a1
DATA  CalcSineCoef+24(SB)/8, $0xbf2a01a019bfdf03
DATA  CalcSineCoef+32(SB)/8, $0x3f8111111110f7d0
DATA  CalcSineCoef+40(SB)/8, $0xbfc5555555555548
GLOBL CalcSineCoef(SB),      RODATA, $48

// Coeficients used by the cosine calculations.
DATA  CalcCosineCoef+0(SB)/8,  $0xbda8fa49a0861a9b
DATA  CalcCosineCoef+8(SB)/8,  $0x3e21ee9d7b4e3f05
DATA  CalcCosineCoef+16(SB)/8, $0xbe927e4f7eac4bc6
DATA  CalcCosineCoef+24(SB)/8, $0x3efa01a019c844f5
DATA  CalcCosineCoef+32(SB)/8, $0xbf56c16c16c14f91
DATA  CalcCosineCoef+40(SB)/8, $0x3fa555555555554b
GLOBL CalcCosineCoef(SB),      RODATA, $48

// Pi/4 split into three parts
DATA  CalcCosinePI4+0(SB)/8,  $0x3fe921fb40000000 // A
DATA  CalcCosinePI4+8(SB)/8,  $0x3e64442d00000000 // B
DATA  CalcCosinePI4+16(SB)/8, $0x3ce8469898cc5170 // C
GLOBL CalcCosinePI4(SB),      RODATA, $24

DATA Float64ABSMask+0(SB)/8, $0x7FFFFFFFFFFFFFFF
GLOBL Float64ABSMask(SB),      RODATA, $8

// func __cosine64(input float64) float64
TEXT ·__cosine64(SB), NOSPLIT, $16-8
  MOVQ input+0(FP), AX // Get our input argument
  ANDQ Float64ABSMask+0(SB), AX // And get the absolute value of our input argument.


  MOVQ AX, return+8(FP)
  RET

// func __absFloat64(input float64) float64
TEXT ·__absFloat64(SB), NOSPLIT, $16-8
  MOVQ input+0(FP), AX
  ANDQ Float64ABSMask+0(SB), AX
  MOVQ AX, return+8(FP)        // Return the absolute value of the float.
  RET

  
