#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import copy
import json

from typing import List
from typing import Tuple
from typing import Iterable
from opcodes import x86_64

def instruction_set():
    for ins in x86_64.read_instruction_set():
        fv = []
        for form in ins.forms:
            if any(op.type in ['{sae}', '{er}'] for op in form.operands):
                new = copy.deepcopy(form)
                new.operands = [op for op in new.operands if op.type not in ['{sae}', '{er}']]
                new_evex = next(v for v in new.encodings[0].components if isinstance(v, x86_64.EVEX))
                new_evex.b = 0
                new_evex.LL = 0b10
                fv.append(new)
                old = next(v for v in form.encodings[0].components if isinstance(v, x86_64.EVEX))
                if not isinstance(old.LL, x86_64.Operand):
                    old.LL = 0
                old.b = 1
        ins.forms.extend(fv)
        yield ins

def instruction_domains():
    with open(os.path.join(os.path.dirname(__file__), 'domains_x86_64.json')) as fp:
        domains = json.load(fp)
        return { ins: dom for dom, insv in domains.items() for ins in insv }

instrs = {}
domains = instruction_domains()

for instr in instruction_set():
    for form in instr.forms:
        if all([v.type not in ('r8l', 'r16l', 'r32l', 'moffs32', 'moffs64') for v in form.operands]):
            name = form.gas_name.upper()
            if name not in instrs:
                instrs[name] = (instr.name, instr.summary, [])
            instrs[name][2].append(form)

for _, _, forms in instrs.values():
    forms.sort(key = lambda f: max([x.score for x in f.isa_extensions], default = 0))

OPCHECKS = {
    '1'             : 'isConst1(%s)',
    '3'             : 'isConst3(%s)',
    'al'            : '%s == AL',
    'ax'            : '%s == AX',
    'eax'           : '%s == EAX',
    'rax'           : '%s == RAX',
    'cl'            : '%s == CL',
    'xmm0'          : '%s == XMM0',
    'rel8'          : 'isRel8(%s)',
    'rel32'         : 'isRel32(%s)',
    'imm4'          : 'isImm4(%s)',
    'imm8'          : 'isImm8(%s)',
    'imm16'         : 'isImm16(%s)',
    'imm32'         : 'isImm32(%s)',
    'imm64'         : 'isImm64(%s)',
    'r8'            : 'isReg8(%s)',
    'r16'           : 'isReg16(%s)',
    'r32'           : 'isReg32(%s)',
    'r64'           : 'isReg64(%s)',
    'mm'            : 'isMM(%s)',
    'xmm'           : 'isXMM(%s)',
    'xmm{k}'        : 'isXMMk(%s)',
    'xmm{k}{z}'     : 'isXMMkz(%s)',
    'ymm'           : 'isYMM(%s)',
    'ymm{k}'        : 'isYMMk(%s)',
    'ymm{k}{z}'     : 'isYMMkz(%s)',
    'zmm'           : 'isZMM(%s)',
    'zmm{k}'        : 'isZMMk(%s)',
    'zmm{k}{z}'     : 'isZMMkz(%s)',
    'k'             : 'isK(%s)',
    'k{k}'          : 'isKk(%s)',
    'm'             : 'isM(%s)',
    'm8'            : 'isM8(%s)',
    'm16'           : 'isM16(%s)',
    'm16{k}{z}'     : 'isM16kz(%s)',
    'm32'           : 'isM32(%s)',
    'm32{k}'        : 'isM32k(%s)',
    'm32{k}{z}'     : 'isM32kz(%s)',
    'm64'           : 'isM64(%s)',
    'm64{k}'        : 'isM64k(%s)',
    'm64{k}{z}'     : 'isM64kz(%s)',
    'm80'           : 'isM80(%s)',
    'm128'          : 'isM128(%s)',
    'm128{k}{z}'    : 'isM128kz(%s)',
    'm256'          : 'isM256(%s)',
    'm256{k}{z}'    : 'isM256kz(%s)',
    'm512'          : 'isM512(%s)',
    'm512{k}{z}'    : 'isM512kz(%s)',
    'm64/m32bcst'   : 'isM64M32bcst(%s)',
    'm128/m32bcst'  : 'isM128M32bcst(%s)',
    'm256/m32bcst'  : 'isM256M32bcst(%s)',
    'm512/m32bcst'  : 'isM512M32bcst(%s)',
    'm128/m64bcst'  : 'isM128M64bcst(%s)',
    'm256/m64bcst'  : 'isM256M64bcst(%s)',
    'm512/m64bcst'  : 'isM512M64bcst(%s)',
    'vm32x'         : 'isVMX(%s)',
    'vm32x{k}'      : 'isVMXk(%s)',
    'vm32y'         : 'isVMY(%s)',
    'vm32y{k}'      : 'isVMYk(%s)',
    'vm32z'         : 'isVMZ(%s)',
    'vm32z{k}'      : 'isVMZk(%s)',
    'vm64x'         : 'isVMX(%s)',
    'vm64x{k}'      : 'isVMXk(%s)',
    'vm64y'         : 'isVMY(%s)',
    'vm64y{k}'      : 'isVMYk(%s)',
    'vm64z'         : 'isVMZ(%s)',
    'vm64z{k}'      : 'isVMZk(%s)',
    '{sae}'         : 'isSAE(%s)',
    '{er}'          : 'isER(%s)',
}

IMMCHECKS = {
    'imm8'  : 'isImm8Ext(%s, %d)',
    'imm32' : 'isImm32Ext(%s, %d)',
}

EVEXCHECKS = {
    'xmm'   : 'isEVEXXMM(%s)',
    'ymm'   : 'isEVEXYMM(%s)',
    'vm32x' : 'isEVEXVMX(%s)',
    'vm64x' : 'isEVEXVMX(%s)',
    'vm32y' : 'isEVEXVMY(%s)',
    'vm64y' : 'isEVEXVMY(%s)',
}

VEXBYTES = {
    'VEX': '0xc4',
    'XOP': '0x8f',
}

ISAMAPPING = {
    'CPUID'           : 'ISA_CPUID',
    'RDTSC'           : 'ISA_RDTSC',
    'RDTSCP'          : 'ISA_RDTSCP',
    'CMOV'            : 'ISA_CMOV',
    'MOVBE'           : 'ISA_MOVBE',
    'POPCNT'          : 'ISA_POPCNT',
    'LZCNT'           : 'ISA_LZCNT',
    'TBM'             : 'ISA_TBM',
    'BMI'             : 'ISA_BMI',
    'BMI2'            : 'ISA_BMI2',
    'ADX'             : 'ISA_ADX',
    'MMX'             : 'ISA_MMX',
    'MMX+'            : 'ISA_MMX_PLUS',
    'FEMMS'           : 'ISA_FEMMS',
    '3dnow!'          : 'ISA_3DNOW',
    '3dnow!+'         : 'ISA_3DNOW_PLUS',
    'SSE'             : 'ISA_SSE',
    'SSE2'            : 'ISA_SSE2',
    'SSE3'            : 'ISA_SSE3',
    'SSSE3'           : 'ISA_SSSE3',
    'SSE4A'           : 'ISA_SSE4A',
    'SSE4.1'          : 'ISA_SSE4_1',
    'SSE4.2'          : 'ISA_SSE4_2',
    'FMA3'            : 'ISA_FMA3',
    'FMA4'            : 'ISA_FMA4',
    'XOP'             : 'ISA_XOP',
    'F16C'            : 'ISA_F16C',
    'AVX'             : 'ISA_AVX',
    'AVX2'            : 'ISA_AVX2',
    'AVX512F'         : 'ISA_AVX512F',
    'AVX512BW'        : 'ISA_AVX512BW',
    'AVX512DQ'        : 'ISA_AVX512DQ',
    'AVX512VL'        : 'ISA_AVX512VL',
    'AVX512PF'        : 'ISA_AVX512PF',
    'AVX512ER'        : 'ISA_AVX512ER',
    'AVX512CD'        : 'ISA_AVX512CD',
    'AVX512VBMI'      : 'ISA_AVX512VBMI',
    'AVX512IFMA'      : 'ISA_AVX512IFMA',
    'AVX512VPOPCNTDQ' : 'ISA_AVX512VPOPCNTDQ',
    'AVX512_4VNNIW'   : 'ISA_AVX512_4VNNIW',
    'AVX512_4FMAPS'   : 'ISA_AVX512_4FMAPS',
    'PREFETCH'        : 'ISA_PREFETCH',
    'PREFETCHW'       : 'ISA_PREFETCHW',
    'PREFETCHWT1'     : 'ISA_PREFETCHWT1',
    'CLFLUSH'         : 'ISA_CLFLUSH',
    'CLFLUSHOPT'      : 'ISA_CLFLUSHOPT',
    'CLWB'            : 'ISA_CLWB',
    'CLZERO'          : 'ISA_CLZERO',
    'RDRAND'          : 'ISA_RDRAND',
    'RDSEED'          : 'ISA_RDSEED',
    'PCLMULQDQ'       : 'ISA_PCLMULQDQ',
    'AES'             : 'ISA_AES',
    'SHA'             : 'ISA_SHA',
    'MONITOR'         : 'ISA_MONITOR',
    'MONITORX'        : 'ISA_MONITORX',
}

DOMAIN_MAP = {
    'generic' : 'DomainGeneric',
    'mmxsse'  : 'DomainMMXSSE',
    'avx'     : 'DomainAVX',
    'fma'     : 'DomainFMA',
    'crypto'  : 'DomainCrypto',
    'mask'    : 'DomainMask',
    'amd'     : 'DomainAMDSpecific',
    'misc'    : 'DomainMisc',
}

BRANCH_INSTRUCTIONS = {
    'JA'    , 'JNA',
    'JAE'   , 'JNAE',
    'JB'    , 'JNB',
    'JBE'   , 'JNBE',
    'JC'    , 'JNC',
    'JE'    , 'JNE',
    'JG'    , 'JNG',
    'JGE'   , 'JNGE',
    'JL'    , 'JNL',
    'JLE'   , 'JNLE',
    'JO'    , 'JNO',
    'JP'    , 'JNP',
    'JS'    , 'JNS',
    'JZ'    , 'JNZ',
    'JPE'   , 'JPO',
    'JECXZ' , 'JRCXZ',
    'JMP'
}

def is_avx512(form: x86_64.InstructionForm) -> bool:
    return any(v.name.startswith('AVX512') for v in form.isa_extensions)

def dump_form(form: x86_64.InstructionForm) -> str:
    if not form.operands:
        return form.gas_name.upper()
    else:
        return form.gas_name.upper() + ' ' + ', '.join(v.type for v in reversed(form.operands))

def require_isa(isa: List[x86_64.ISAExtension]) -> str:
    flags = []
    for v in isa:
        if v.name not in ISAMAPPING:
            raise RuntimeError('invalid ISA: ' + v.name)
        flags.append(ISAMAPPING[v.name])
    return ' | '.join(flags)

def operand_match(ops: List[x86_64.Operand], argc: int, avx512: bool) -> Iterable[str]:
    for i, op in enumerate(ops):
        if i < argc:
            argv = 'v%d' % i
        else:
            argv = 'vv[%d]' % (i - argc)
        if op.extended_size is not None and op.type in IMMCHECKS:
            yield IMMCHECKS[op.type] % (argv, op.extended_size)
        elif avx512 and op.type in EVEXCHECKS:
            yield EVEXCHECKS[op.type] % argv
        else:
            yield OPCHECKS[op.type] % argv

def generate_encoding(enc: x86_64.Encoding, ops: List[x86_64.Operand], gen_branch: bool = False) -> Tuple[str, List[str]]:
    buf = []
    flags = []
    disp8v = None
    for item in enc.components:
        if isinstance(item, x86_64.Prefix):
            buf.append('m.emit(0x%02x)' % item.byte)
        elif isinstance(item, x86_64.REX):
            item.set_ignored()
            if item.is_mandatory:
                if isinstance(item.X, x86_64.Operand):
                    args = [str(item.W)]
                    if isinstance(item.R, x86_64.Operand):
                        args.append('hcode(v[%d])' % ops.index(item.R))
                    else:
                        args.append(str(item.R))
                    args.append('addr(v[%d])' % ops.index(item.X))
                    buf.append('m.rexm(%s)' % ', '.join(args))
                else:
                    rex = 0x40 | (item.W << 3)
                    args = []
                    if isinstance(item.R, x86_64.Operand):
                        args.append('hcode(v[%d]) << 2' % ops.index(item.R))
                    else:
                        rex |= item.R << 2
                    if isinstance(item.B, x86_64.Operand):
                        args.append('hcode(v[%d])' % ops.index(item.B))
                    else:
                        rex |= item.B
                    rex |= item.X << 1
                    buf.append('m.emit(%s)' % ' | '.join(['0x%02x' % rex] + args))
            else:
                args = []
                if isinstance(item.R, x86_64.Operand):
                    args.append('hcode(v[%d])' % ops.index(item.R))
                else:
                    args.append(str(item.R))
                if isinstance(item.X, x86_64.Operand):
                    args.append('addr(v[%d])' % ops.index(item.X))
                else:
                    args.append('v[%d]' % ops.index(item.B))
                rexv = []
                for i, op in enumerate(ops):
                    if op.type == 'r8':
                        rexv.append('isReg8REX(v[%d])' % i)
                if not rexv:
                    args.append('false')
                else:
                    args.append(' || '.join(rexv))
                buf.append('m.rexo(%s)' % ', '.join(args))
        elif isinstance(item, x86_64.VEX):
            item.set_ignored()
            if item.type == 'VEX' and item.mmmmm == 0b00001 and item.W == 0:
                if item.R == 1 and item.X == 1 and item.B == 1:
                    buf.append('m.emit(0xc5)')
                    buf.append('m.emit(0x%02x)' % (0xf8 | (item.L << 2) | int(item.pp)))
                else:
                    args = [str(int(item.L << 2) | item.pp)]
                    if isinstance(item.R, x86_64.Operand):
                        args.append('hcode(v[%d])' % ops.index(item.R))
                    else:
                        args.append('0')
                    if isinstance(item.X, x86_64.Operand):
                        args.append('addr(v[%d])' % ops.index(item.X))
                    elif isinstance(item.B, x86_64.Operand):
                        args.append('v[%d]' % ops.index(item.B))
                    else:
                        args.append('nil')
                    if isinstance(item.vvvv, x86_64.Operand):
                        args.append('hlcode(v[%d])' % ops.index(item.vvvv))
                    else:
                        args.append('0')
                    buf.append('m.vex2(%s)' % ', '.join(args))
            else:
                if isinstance(item.X, x86_64.Operand):
                    args = [
                        VEXBYTES[item.type],
                        bin(item.mmmmm),
                        '0x%02x' % ((item.W << 7) | (item.L << 2) | int(item.pp)),
                    ]
                    if isinstance(item.R, x86_64.Operand):
                        args.append('hcode(v[%d])' % ops.index(item.R))
                    else:
                        args.append('0')
                    args.append('addr(v[%d])' % ops.index(item.X))
                    if isinstance(item.vvvv, x86_64.Operand):
                        args.append('hlcode(v[%d])' % ops.index(item.vvvv))
                    else:
                        args.append('0')
                    buf.append('m.vex3(%s)' % ', '.join(args))
                else:
                    buf.append('m.emit(%s)' % VEXBYTES[item.type])
                    v0 = '0x%02x' % (0xe0 | item.mmmmm)
                    if isinstance(item.R, x86_64.Operand):
                        v0 += ' ^ (hcode(v[%d]) << 7)' % ops.index(item.R)
                    if isinstance(item.B, x86_64.Operand):
                        v0 += ' ^ (hcode(v[%d]) << 5)' % ops.index(item.B)
                    buf.append('m.emit(%s)' % v0)
                    vex = 0x78 | (item.W << 7) | (item.L << 2) | int(item.pp)
                    if isinstance(item.vvvv, x86_64.Operand):
                        buf.append('m.emit(0x%02x ^ (hlcode(v[%d]) << 3))' % (vex, ops.index(item.vvvv)))
                    else:
                        buf.append('m.emit(0x%02x)' % vex)
        elif isinstance(item, x86_64.EVEX):
            disp8v = item.disp8xN
            item.set_ignored()
            if item.X.is_memory:
                args = ['0b' + format(item.mm, '02b'), '0x%02x' % (item.W << 7 | int(item.pp) | 0b100)]
                if isinstance(item.LL, x86_64.Operand):
                    args.append('vcode(v[%d])' % ops.index(item.LL))
                else:
                    args.append('0b' + format(item.LL, '02b'))
                if isinstance(item.RR, x86_64.Operand):
                    args.append('ehcode(v[%d])' % ops.index(item.RR))
                else:
                    args.append(str(item.RR))
                args.append('addr(v[%d])' % ops.index(item.X))
                if item.vvvv != 0:
                    args.append('vcode(v[%d])' % ops.index(item.vvvv))
                else:
                    args.append('0')
                if item.aaa != 0:
                    args.append('kcode(v[%d])' % ops.index(item.aaa))
                else:
                    args.append('0')
                if item.z != 0:
                    args.append('zcode(v[%d])' % ops.index(item.z))
                else:
                    args.append('0')
                if isinstance(item.b, x86_64.Operand):
                    args.append('bcode(v[%d])' % ops.index(item.b))
                elif item.b != 0:
                    args.append(str(item.b))
                else:
                    args.append('0')
                buf.append('m.evex(%s)' % ', '.join(args))
            else:
                buf.append('m.emit(0x62)')
                if isinstance(item.RR, x86_64.Operand):
                    v0, v1, v2, v3 = 0xf0 | item.mm, ops.index(item.RR), ops.index(item.B), ops.index(item.RR)
                    buf.append('m.emit(0x%02x ^ ((hcode(v[%d]) << 7) | (ehcode(v[%d]) << 5) | (ecode(v[%d]) << 4)))' % (v0, v1, v2, v3))
                else:
                    r0 = item.RR & 1
                    r1 = (item.RR >> 1) & 1
                    byte = (item.mm | (r0 << 7) | (r1 << 4)) ^ 0xf0
                    if byte == 0:
                        buf.append('m.emit(ehcode(v[%d]) << 5)' % ops.index(item.B))
                    else:
                        buf.append('m.emit(0x%02x ^ (ehcode(v[%d]) << 5))' % (byte, ops.index(item.B)))
                vvvv = item.W << 7 | int(item.pp) | 0b01111100
                if isinstance(item.vvvv, x86_64.Operand):
                    buf.append('m.emit(0x%02x ^ (hlcode(v[%d]) << 3))' % (vvvv, ops.index(item.vvvv)))
                else:
                    buf.append('m.emit(0x%02x)' % vvvv)
                byte = item.b << 4
                parts = []
                if isinstance(item.z, x86_64.Operand):
                    parts.append('(zcode(v[%d]) << 7)' % ops.index(item.z))
                else:
                    byte |= item.z << 7
                if isinstance(item.LL, x86_64.Operand):
                    parts.append('(vcode(v[%d]) << 5)' % ops.index(item.LL))
                else:
                    byte |= item.LL << 5
                if isinstance(item.V, x86_64.Operand):
                    parts.append('(0x08 ^ (ecode(v[%d]) << 3))' % ops.index(item.V))
                else:
                    byte |= (item.V ^ 1) << 3
                if isinstance(item.aaa, x86_64.Operand):
                    parts.append('kcode(v[%d])' % ops.index(item.aaa))
                parts.append('0x%02x' % byte)
                buf.append('m.emit(%s)' % ' | '.join(parts))
        elif isinstance(item, x86_64.Opcode):
            if not item.addend:
                buf.append('m.emit(0x%02x)' % item.byte)
            else:
                buf.append('m.emit(0x%02x | lcode(v[%d]))' % (item.byte, ops.index(item.addend)))
        elif isinstance(item, x86_64.ModRM):
            if isinstance(item.mode, x86_64.Operand):
                if isinstance(item.reg, x86_64.Operand):
                    reg = 'lcode(v[%d])' % ops.index(item.reg)
                else:
                    reg = str(item.reg)
                if disp8v is None:
                    disp = 1
                else:
                    disp = disp8v
                buf.append('m.mrsd(%s, addr(v[%d]), %d)' % (reg, ops.index(item.rm), disp))
            else:
                mod = item.mode << 6
                parts = []
                if isinstance(item.reg, x86_64.Operand):
                    parts.append('lcode(v[%d]) << 3' % ops.index(item.reg))
                elif item.reg:
                    mod |= item.reg << 3
                parts.append('lcode(v[%d])' % ops.index(item.rm))
                buf.append('m.emit(%s)' % ' | '.join(['0x%02x' % mod] + parts))
        elif isinstance(item, x86_64.Immediate):
            if isinstance(item.value, x86_64.Operand):
                if item.size == 1:
                    buf.append('m.imm1(toImmAny(v[%d]))' % ops.index(item.value))
                elif item.size == 2:
                    buf.append('m.imm2(toImmAny(v[%d]))' % ops.index(item.value))
                elif item.size == 4:
                    buf.append('m.imm4(toImmAny(v[%d]))' % ops.index(item.value))
                elif item.size == 8:
                    buf.append('m.imm8(toImmAny(v[%d]))' % ops.index(item.value))
                else:
                    raise RuntimeError('invalid imm size: ' + str(item.size))
            else:
                if item.size == 1:
                    buf.append('m.imm1(0x%02x)' % item.value)
                elif item.size == 2:
                    buf.append('m.imm2(0x%04x)' % item.value)
                elif item.size == 4:
                    buf.append('m.imm4(0x%08x)' % item.value)
                elif item.size == 8:
                    buf.append('m.imm8(0x%016x)' % item.value)
                else:
                    raise RuntimeError('invalid imm size: ' + str(item.size))
        elif isinstance(item, x86_64.RegisterByte):
            ibr = 'hlcode(v[%d]) << 4' % ops.index(item.register)
            if item.payload is not None:
                ibr = '(%s) | imml(v[%d])' % (ibr, ops.index(item.payload))
            buf.append('m.emit(%s)' % ibr)
        elif isinstance(item, x86_64.CodeOffset):
            if item.size == 1:
                buf.append('m.imm1(relv(v[%d]))' % ops.index(item.value))
                if gen_branch:
                    flags.append('_F_rel1')
            elif item.size == 4:
                buf.append('m.imm4(relv(v[%d]))' % ops.index(item.value))
                if gen_branch:
                    flags.append('_F_rel4')
            else:
                raise RuntimeError('invalid code offset size: ' + repr(item.size))
        else:
            raise RuntimeError('unknown encoding component: ' + repr(item))
    if not flags:
        return '0', buf
    else:
        return ' | '.join(flags), buf

class CodeGen:
    def __init__(self):
        self.buf = []
        self.level = 0

    @property
    def src(self) -> str:
        return '\n'.join(self.buf)

    def line(self, src: str = ''):
        self.buf.append(' ' * (self.level * 4) + src)

    def dedent(self):
        self.level -= 1

    def indent(self):
        self.level += 1

class CodeBlock:
    def __init__(self, gen: CodeGen):
        self.gen = gen

    def __exit__(self, *_):
        self.gen.dedent()

    def __enter__(self):
        self.gen.indent()
        return self

cc = CodeGen()
cc.line('// Code generated by "mkasm_amd64.py", DO NOT EDIT.')
cc.line()
cc.line('package x86_64')
cc.line()

nargs = 0
nforms = 0
argsmap = {}
for name, (_, _, forms) in instrs.items():
    fcnt = 0
    for form in forms:
        acnt = len(form.operands)
        fcnt += len(form.encodings)
        argsmap.setdefault(name, set()).add(acnt)
        if nargs < acnt:
            nargs = acnt
    if nforms < fcnt:
        nforms = fcnt

cc.line('const (')
with CodeBlock(cc):
    cc.line('_N_args  = %d' % nargs)
    cc.line('_N_forms = %d' % nforms)
cc.line(')')
cc.line()
cc.line('// Instructions maps all the instruction name to it\'s encoder function.')
cc.line('var Instructions = map[string]_InstructionEncoder {')

width = max(
    len(x)
    for x in instrs
)

with CodeBlock(cc):
    for name in sorted(instrs):
        key = '"%s"' % name.lower()
        cc.line('%s: __asm_proxy_%s__,' % (key.ljust(width + 3), name))

cc.line('}')
cc.line()

for name in sorted(instrs):
    cc.line('func __asm_proxy_%s__(p *Program, v ...interface{}) *Instruction {' % name)
    with CodeBlock(cc):
        args = argsmap[name]
        if len(args) == 1:
            argc = next(iter(args))
            cc.line('if len(v) == %d {' % argc)
            with CodeBlock(cc):
                argv = ['v[%d]' % i for i in range(argc)]
                cc.line('return p.%s(%s)' % (name, ', '.join(argv)))
            cc.line('} else {')
            with CodeBlock(cc):
                if argc == 0:
                    cc.line('panic("instruction %s takes no operands")' % name)
                elif argc == 1:
                    cc.line('panic("instruction %s takes exactly 1 operand")' % name)
                else:
                    cc.line('panic("instruction %s takes exactly %d operands")' % (name, argc))
            cc.line('}')
        else:
            cc.line('switch len(v) {')
            with CodeBlock(cc):
                for argc in sorted(args):
                    argv = ['v[%d]' % i for i in range(argc)]
                    cc.line('case %d  : return p.%s(%s)' % (argc, name, ', '.join(argv)))
                cc.line('default : panic("instruction %s takes %s operands")' % (name, ' or '.join(map(str, sorted(args)))))
            cc.line('}')
    cc.line('}')
    cc.line()

with open('x86_64/instructions_table.go', 'w') as fp:
    fp.write(cc.src)

cc = CodeGen()
cc.line('// Code generated by "mkasm_amd64.py", DO NOT EDIT.')
cc.line()
cc.line('package x86_64')
cc.line()

for name, (ins, desc, forms) in sorted(instrs.items()):
    cc.line('// %s performs "%s".' % (name, desc))
    cc.line('//')
    cc.line('// Mnemonic        : ' + ins)
    cc.line('// Supported forms : (%d form%s)' % (len(forms), '' if len(forms) == 1 else 's'))
    cc.line('//')
    nops = set()
    fwidth = max(map(len, map(dump_form, forms)))
    for form in forms:
        nops.add(len(form.operands))
        if not form.isa_extensions:
            cc.line('//    * ' + dump_form(form))
        else:
            cc.line('//    * %-*s    [%s]' % (fwidth, dump_form(form), ','.join(sorted(v.name for v in form.isa_extensions))))
    nfix = min(nops)
    args = ['v%d interface{}' % i for i in range(nfix)]
    if len(nops) != 1:
        args.append('vv ...interface{}')
    cc.line('//')
    cc.line('func (self *Program) %s(%s) *Instruction {' % (name, ', '.join(args)))
    with CodeBlock(cc):
        base = ['v%d' % i for i in range(nfix)]
        if len(nops) == 1:
            cc.line('p := self.alloc("%s", %d, Operands { %s })' % (name, nfix, ', '.join(base)))
        else:
            cc.line('var p *Instruction')
            cc.line('switch len(vv) {')
            with CodeBlock(cc):
                for argc in sorted(nops):
                    args = base[:] + ['vv[%d]' % i for i in range(argc - nfix)]
                    cc.line('case %d  : p = self.alloc("%s", %d, Operands { %s })' % (argc - nfix, name, argc, ', '.join(args)))
                cc.line('default : panic("instruction %s takes %s operands")' % (name, ' or '.join(map(str, sorted(nops)))))
            cc.line('}')
        if name == 'JMP':
            cc.line('p.branch = _B_unconditional')
        elif name in BRANCH_INSTRUCTIONS:
            cc.line('p.branch = _B_conditional')
        is_labeled = False
        must_success = False
        for form in forms:
            ops = list(reversed(form.operands))
            if len(ops) == 1 and ops[0].type in ('rel8', 'rel32'):
                is_labeled = True
            conds = []
            cc.line('// ' + dump_form(form))
            if len(nops) != 1:
                conds.append('len(vv) == %d' % (len(ops) - nfix))
            conds.extend(operand_match(ops, nfix, is_avx512(form)))
            if conds:
                cc.line('if %s {' % ' && '.join(conds))
                cc.indent()
            else:
                must_success = True
            if form.isa_extensions:
                cc.line('self.require(%s)' % require_isa(form.isa_extensions))
            cc.line('p.domain = ' + DOMAIN_MAP[domains.get(form.name, 'misc')])
            for enc in form.encodings:
                flags, instr = generate_encoding(enc, ops, gen_branch = False)
                cc.line('p.add(%s, func(m *_Encoding, v []interface{}) {' % flags)
                with CodeBlock(cc):
                    for line in instr:
                        cc.line(line)
                cc.line('})')
            if conds:
                cc.dedent()
                cc.line('}')
        if is_labeled:
            cc.line('// %s label' % name)
            cc.line('if isLabel(v0) {')
            with CodeBlock(cc):
                for form in forms:
                    ops = list(reversed(form.operands))
                    for enc in form.encodings:
                        flags, instr = generate_encoding(enc, ops, gen_branch = True)
                        cc.line('p.add(%s, func(m *_Encoding, v []interface{}) {' % flags)
                        with CodeBlock(cc):
                            for line in instr:
                                cc.line(line)
                        cc.line('})')
            cc.line('}')
        if not must_success:
            cc.line('if p.len == 0 {')
            with CodeBlock(cc):
                cc.line('panic("invalid operands for %s")' % name)
            cc.line('}')
        cc.line('return p')
    cc.line('}')
    cc.line()

with open('x86_64/instructions.go', 'w') as fp:
    fp.write(cc.src)
