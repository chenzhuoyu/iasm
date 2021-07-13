#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import copy

from typing import List
from typing import Tuple
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

instrs = {}
for instr in instruction_set():
    for form in instr.forms:
        # TODO: remove this
        if not (form.gas_name.startswith('j') or form.gas_name.startswith('add') or form.gas_name == 'vzeroupper'): # 'vpperm'):
            continue
        if all([v.type not in ('r8l', 'r16l', 'r32l', 'moffs32', 'moffs64') for v in form.operands]):
            name = form.gas_name.upper()
            if name not in instrs:
                instrs[name] = (instr.name, instr.summary, [])
            instrs[name][2].append(form)

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

def operand_match(ops: List[x86_64.Operand], avx512: bool) -> str:
    checks = []
    for i, op in enumerate(ops):
        if op.extended_size is not None and op.type in IMMCHECKS:
            checks.append(IMMCHECKS[op.type] % ('args[%d]' % i, op.extended_size))
        elif avx512 and op.type in EVEXCHECKS:
            checks.append(EVEXCHECKS[op.type] % 'args[%d]' % i)
        else:
            checks.append(OPCHECKS[op.type] % 'args[%d]' % i)
    return ' && '.join(checks)

def generate_encoding(enc: x86_64.Encoding, ops: List[x86_64.Operand]) -> Tuple[str, List[str]]:
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
                flags.append('_OREX')
        elif isinstance(item, x86_64.VEX):
            item.set_ignored()
            if item.type == 'VEX' and item.mmmmm == 0b00001 and item.W == 0:
                if item.R == 1 and item.X == 1 and item.B == 1:
                    buf.append('m.emit(0xc5)')
                    buf.append('m.emit(0x%02x)' % (0xf8 | int(item.L << 2) | item.pp))
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
                    flags.append('_VEX2')
            else:
                pass
        elif isinstance(item, x86_64.EVEX):
            # TODO: this
            buf.append('EVEX')
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
                flags.append('_MRSD')
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
            # TODO: this
            buf.append('RegByte')
        elif isinstance(item, x86_64.CodeOffset):
            if item.size == 1:
                buf.append('m.imm1(offs(v[%d]))' % ops.index(item.value))
            elif item.size == 4:
                buf.append('m.imm4(offs(v[%d]))' % ops.index(item.value))
            else:
                raise RuntimeError('invalid code offset size: ' + repr(item.size))
        else:
            raise RuntimeError('unknown encoding component: ' + repr(item))
    if not flags:
        return '0', buf
    else:
        return ' | '.join(flags), buf

src = [
    '// Code generated by "mkasm_amd64.py", DO NOT EDIT.',
    '',
    'package x86_64',
    '',
    '// Instructions maps all the instruction name to it\'s encoder function.',
    'var Instructions = map[string]func(*Program, ...interface{}) *Instruction {',
]

width = max(
    len(x)
    for x in instrs
)

for name in sorted(instrs):
    key = '"%s"' % name.lower()
    src.append('    %s: (*Program).%s,' % (key.ljust(width + 3), name))

src.append('}')
src.append('')

for name, (ins, desc, forms) in sorted(instrs.items()):
    src.append('// %s performs "%s".' % (name, desc))
    src.append('//')
    src.append('// Mnemonic        : ' + ins)
    src.append('// Supported forms : (%d form%s)' % (len(forms), '' if len(forms) == 1 else 's'))
    src.append('//')
    for form in forms:
        src.append('//    * ' + dump_form(form))
    src.append('//')
    src.append('func (self *Program) %s(args ...interface{}) *Instruction {' % name)
    src.append('    p := self.alloc(args)')
    for form in forms:
        ops = list(reversed(form.operands))
        src.append('    // ' + dump_form(form))
        if not ops:
            src.append('    if len(args) == 0 {')
        else:
            src.append('    if len(args) == %d && %s {' % (len(ops), operand_match(ops, is_avx512(form))))
        if form.isa_extensions:
            src.append('        self.require(%s)' % require_isa(form.isa_extensions))
        for enc in form.encodings:
            flags, instr = generate_encoding(enc, ops)
            src.append('        p.add(%s, func(m *_Encoding, v []interface{}) {' % flags)
            for line in instr:
                src.append('            ' + line)
            src.append('        })')
        src.append('    }')
    src.append('    if len(p.buf) == 0 {')
    src.append('        panic("invalid operands for %s")' % name)
    src.append('    }')
    src.append('    return p')
    src.append('}')
    src.append('')

with open('x86_64/instrs.go', 'w') as fp:
    fp.write('\n'.join(src))
