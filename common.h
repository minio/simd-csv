#define CREATE_MASK(_Y1, _Y2, _R1, _R2) \
	VPMOVMSKB _Y1, _R1 \
	VPMOVMSKB _Y2, _R2 \
	SHLQ      $32, _R2 \
	ORQ       _R1, _R2

#define MASK_TRAILING_BYTES(MAX, SCRATCH1, SCRATCH2, RPOS, Y_SCRATCH, _Y) \
	LEAQ    MASKTABLE<>(SB), SCRATCH1         \
	MOVQ    $MAX, SCRATCH2                    \
	SUBQ    CX, SCRATCH2                      \
	VMOVDQU (SCRATCH1)(SCRATCH2*1), Y_SCRATCH \ // Load mask
	VPAND   Y_SCRATCH, _Y, _Y                 // Mask message

DATA MASKTABLE<>+0x000(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x008(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x010(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x018(SB)/8, $0x00ffffffffffffff
DATA MASKTABLE<>+0x020(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x028(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x030(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x038(SB)/8, $0x0000000000000000
GLOBL MASKTABLE<>(SB), 8, $64
