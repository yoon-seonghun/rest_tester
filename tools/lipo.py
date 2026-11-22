#!/usr/bin/env python3
"""Mach-O 실행 파일 여러 개를 유니버설(fat) 바이너리 하나로 묶는다. (lipo -create 대체)
사용법: lipo.py <출력> <입력1> <입력2> ..."""
import struct, sys
out, ins = sys.argv[1], sys.argv[2:]
ALIGN = 14  # 2^14 = 16KB (arm64 페이지 크기)
slices = []
for p in ins:
    d = open(p, "rb").read()
    magic, cputype, cpusub = struct.unpack_from("<III", d, 0)
    assert magic == 0xFEEDFACF, p + " is not a 64-bit Mach-O"
    slices.append((cputype, cpusub, d))
off = 4096
hdr = struct.pack(">II", 0xCAFEBABE, len(slices))
body = b""; entries = b""
for cputype, cpusub, d in slices:
    off = (off + (1 << ALIGN) - 1) & ~((1 << ALIGN) - 1)
    entries += struct.pack(">iiIII", cputype, cpusub, off, len(d), ALIGN)
    slices_off = off; off += len(d)
    body_pos = slices_off
    slices[slices.index((cputype, cpusub, d))] = (cputype, cpusub, d, slices_off)
buf = bytearray(off)
buf[:len(hdr + entries)] = hdr + entries
for cputype, cpusub, d, o in slices:
    buf[o:o + len(d)] = d
open(out, "wb").write(buf)
