#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import itertools

from typing import cast
from typing import Literal
from typing import Callable
from typing import Iterator
from typing import NamedTuple

from enum import Enum
from functools import cached_property
from collections import OrderedDict

from xml.etree import ElementTree
from xml.etree.ElementTree import Element

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

def status(*args):
    sys.stdout.write('\x1b[#F\x1b[K')
    print(*args)

print()
status('* Preparing ...')

### ---------- Instruction Encoding Classes ---------- ###

class Account:
    name: str
    desc: str

    def __init__(self, name: str, desc: str):
        self.name = name
        self.desc = desc

    def __repr__(self) -> str:
        return '%s %s' % (self.name, self.desc)

class Definition:
    name: str
    desc: str
    bits: dict[str, set[str]]

    def __init__(self, name: str, desc: str, bits: dict[str, set[str]]):
        self.name = name
        self.desc = desc
        self.bits = bits

    def __repr__(self) -> str:
        return '%s %s [%s]' % (self.name, self.desc, ' '.join(
            '%s=%s%s%s' % (
                key,
                '' if len(vals) == 1 else '{',
                ','.join(sorted(vals)),
                '' if len(vals) == 1 else '}'
            ) for key, vals in self.bits.items()
        ))

class Instruction:
    bits: list[int | None]
    refs: dict[str, tuple[int, int]]

    def __init__(self, other: 'Instruction | None' = None):
        if other is None:
            self.refs = {}
            self.bits = [None] * 32
        else:
            self.bits = other.bits[:]
            self.refs = dict(other.refs)

    def __repr__(self) -> str:
        char = 0
        mlen = 0
        refs = []
        bits = ['x' if v is None else str(v) for v in self.bits]

        for v in self.refs:
            mlen = max(mlen, len(v))

        bits.extend(' ' * mlen)
        bits.extend('  ')

        for p, n in self.refs.values():
            char += 1
            bits[p:p + n] = [
                '\x1b[3%dm%s\x1b[0m' % (char % 5 + 1, 'x' if v is None else v)
                for v in self.bits[p:p + n]
            ]

        vals = list(self.refs.items())
        vals.sort(key = lambda x: x[1], reverse = True)

        for i, (v, (p, n)) in enumerate(vals):
            line = list(v[::-1].rjust(mlen + 34))
            line[p + n // 2] = '┘'
            line[p + n // 2 + 1:33] = '─' * (32 - p - n // 2)

            for i in range(len(refs)):
                refs[i][p + n // 2] = '│'
            else:
                refs.append(line)

        return '\n'.join([
            ''.join(bits[::-1]),
            *(''.join(v[::-1]) for v in refs),
        ])

    @staticmethod
    def _intersects(a1: int, b1: int, a2: int, b2: int) -> bool:
        return a1 <= a2 <= b1 <= b2 or a2 <= a1 <= b2 <= b1

    def _remove_bits(self, hibit: int, width: int) -> Iterator[str | None]:
        for k, (i, w) in self.refs.items():
            if w == width and i == hibit - width + 1:
                yield k
            elif self._intersects(i, i + w - 1, hibit - width + 1, hibit):
                yield None

    def update(self, boxes: list[Element], *, no_add: bool = False):
        for box in boxes:
            bits = box.findall('c')
            name = box.attrib.get('name')
            hibit = int(box.attrib['hibit'])
            width = int(box.attrib.get('width', 1))

            if not no_add and name is not None:
                keys = self._remove_bits(hibit, width)
                keys = set(keys)

                for key in keys:
                    if key is not None:
                        del self.refs[key]

                if None not in keys:
                    self.refs[name] = (hibit - width + 1, width)

            for i, item in enumerate(bits):
                if item.text in {'0', '(0)', 'z', 'Z'}:
                    assert self.bits[hibit - i] != 1, 'bit confiction: 0'
                    self.bits[hibit - i] = 0
                elif item.text in {'1', '(1)'}:
                    assert self.bits[hibit - i] != 0, 'bit confiction: 1'
                    self.bits[hibit - i] = 1
                else:
                    assert box.attrib.get('usename') == '1' or item.text in {None, '', 'x', 'N'}, \
                        'invalid cell value: ' + repr(item.text)

class EncodingTabEntry(NamedTuple):
    name: str
    desc: str
    form: str
    file: str
    func: str
    bits: Instruction

    def __repr__(self) -> str:
        return '\n'.join([
            'Encoding %s {' % self.name,
            '    ' + self.desc,
            '    form = ' + self.form,
            '    file = ' + self.file,
            '    func = ' + self.func,
            *('    ' + v for v in repr(self.bits).splitlines()),
            '}',
        ])

class InstructionTabEntry(NamedTuple):
    name: str
    desc: str
    data: Element
    base: Instruction

    def __repr__(self) -> str:
        return '\n'.join([
            'Instruction %s {' % self.name,
            '    ' + self.desc,
            *('    ' + v for v in repr(self.base).splitlines()),
            '}',
        ])

def parse_bit(bit: str) -> int | None:
    match bit:
        case '0': return 0
        case '1': return 1
        case 'x': return None
        case _  : raise AssertionError('invalid bit value: ' + repr(bit))

cc = CodeGen()
cc.line('// Code generated by "mkasm_aarch64.py", DO NOT EDIT.')
cc.line()
cc.line('package aarch64')
cc.line()

enctab = dict[str, EncodingTabEntry]()
instab = dict[str, InstructionTabEntry]()

isadocs = ElementTree.parse(os.path.join(sys.argv[1], 'onebigfile.xml')).getroot()
parent_tab = { c: p for p in isadocs.iter() for c in p }

encindex = isadocs.find('.//encodingindex')
assert encindex is not None, 'invalid encoding index file'

for iclass in sorted(encindex.findall('iclass_sect'), key = lambda x: x.attrib['id']):
    name = iclass.attrib['id']
    desc = iclass.attrib['title']
    itab = iclass.find('instructiontable')

    # TODO: support SVE instructions sometime in the future
    if name.startswith('sve_'):
        continue

    assert itab is not None, 'missing instruction table for ' + name
    assert itab.attrib['iclass'] == name, 'wrong instruction table iclass for ' + name
    status('* Instruction Class:', name)

    bits = Instruction()
    bits.update(iclass.findall('regdiagram/box'))

    cond = set()
    req = bits.bits[:]
    args = sorted(bits.refs.items(), key = lambda x: x[1], reverse = True)

    for dc in iclass.findall('decode_constraints/decode_constraint'):
        cond.add((
            dc.attrib['name'],
            dc.attrib['op'],
            dc.attrib['val'],
        ))

    for p, n in bits.refs.values():
        req[p:p + n] = [0] * n

    instab[name] = InstructionTabEntry(
        name = name,
        desc = desc,
        data = itab,
        base = bits,
    )

    cc.line('// %s: %s' % (name, desc))
    cc.line('func %s(%s uint32) uint32 {' % (name, ', '.join(v for v, _ in args)))
    cc.indent()

    for th, (_, n) in args:
        cc.line('if %s &^ 0b%s != 0 {' % (th, '1' * n))
        cc.indent()
        cc.line('panic("%s: invalid %s")' % (name, th))
        cc.dedent()
        cc.line('}')

    for key, op, val in sorted(cond):
        assert op == '!=', 'decode constraint is not implemented: %s %s %s' % (key, op, val)
        cc.line('if %s == 0b%s {' % (key, val))
        cc.indent()
        cc.line('panic("%s: decode constraint is not satisfied: %s %s %s")' % (name, key, op, val))
        cc.dedent()
        cc.line('}')

    assert None not in req, 'unset bit in %s: %r' % (name, req)
    cc.line('ret := uint32(0x%08x)' % int(''.join(map(str, req))[::-1], 2))

    for th, (p, _) in args:
        if p == 0:
            cc.line('ret |= ' + th)
        else:
            cc.line('ret |= %s << %d' % (th, p))

    cc.line('return ret;')
    cc.dedent()
    cc.line('}')
    cc.line()

with open('arch/aarch64/encodings.go', 'w') as fp:
    fp.write('\n'.join(cc.buf))

for name, entry in sorted(instab.items(), key = lambda v: v[0]):
    keys = []
    itab = entry.data

    for th in itab.findall('thead/tr[@id="heading2"]/th[@class="bitfields"]'):
        assert th.text, 'missing field name for ' + name
        keys.append(th.text)

    for th in itab.findall('tbody/tr[@class="instructiontable"]'):
        if th.attrib.get('undef') != '1':
            encname   = th.attrib['encname']
            iformfile = th.attrib['iformfile']
            iformname = th.find('td[@class="iformname"]')
            bitfields = th.findall('td[@class="bitfield"]')

            # TODO: remove this
            # if 'advsimd' in iformfile or iformfile in {'b_cond.xml', 'casp.xml'}:
            #     continue
            if not (
                # iformfile.startswith('bti') or
                # iformfile.startswith('aesd_') or
                # iformfile.startswith('uqxtn_') or
                # iformfile.startswith('ldr_') or
                # iformfile.startswith('and_') or
                # iformfile.startswith('eor_') or
                # iformfile.startswith('orn_') or
                # iformfile.startswith('orr_') or
                # iformfile.startswith('addg') or
                # iformfile.startswith('blra') or
                # iformfile.startswith('bra') or
                # iformfile.startswith('ccmn') or
                # iformfile.startswith('autia') or
                # iformfile.startswith('clrex') or
                # iformfile.startswith('dmb') or
                # iformfile.startswith('dsb') or
                # iformfile.startswith('fcmp') or
                iformfile.startswith('fcvtzs')
            ):
                continue

            assert iformfile, 'missing iform files for ' + name
            assert iformname is not None, 'missing iform names for ' + name
            assert len(bitfields) == len(keys), 'mismatched bitfields for %s.%s' % (name, encname)

            desc = iformname.text
            form = iformname.attrib['iformid']
            assert desc and form, 'missing form or description for %s.%s' % (name, encname)

            bits = []
            instr = Instruction(entry.base)

            for field in bitfields:
                if not field.text or field.text.startswith('!='):
                    bits.append(None)
                else:
                    bits.append(list(map(parse_bit, field.text[::-1])))

            for key, req in zip(keys, bits):
                if req is not None:
                    p, n = instr.refs[key]
                    assert len(req) == n, 'mismatched bits for %s.%s.%s' % (name, encname, key)
                    instr.bits[p:p + n] = req

            enc = EncodingTabEntry(
                name = encname,
                desc = desc,
                form = form,
                file = iformfile,
                func = name,
                bits = instr,
            )

            assert encname not in enctab, 'duplicated encoding name %s.%s' % (name, encname)
            status('* Encoding Table Entry: %s.%s' % (name, encname))
            enctab[encname] = enc

### ---------- Instruction Assembly Template ---------- ###

class Sop(Enum):
    LSL  = 'lsl'
    LSR  = 'lsr'
    ASR  = 'asr'
    ROR  = 'ror'
    MSL  = 'msl'

class Xop(Enum):
    UXTB = 'uxtb'
    UXTH = 'uxth'
    UXTW = 'uxtw'
    UXTX = 'uxtx'
    SXTB = 'sxtb'
    SXTH = 'sxth'
    SXTW = 'sxtw'
    SXTX = 'sxtx'

class Sym(Enum):
    CSYNC   = 'CSYNC'
    PRFOP   = 'prfop'
    OPTION  = 'option'
    SYSREG  = 'systemreg'
    TARGETS = 'targets'

class Tag(str):
    @cached_property
    def name(self) -> str:
        return self + ''

    def __str__(self) -> str:
        return '=' + self

    def __repr__(self) -> str:
        return 'Tag(%s)' % super().__repr__()

class Imm(str):
    @cached_property
    def name(self) -> str:
        return self + ''

    def __str__(self) -> str:
        return '#' + self

    def __repr__(self) -> str:
        return 'Imm(%s)' % super().__repr__()

class Lit:
    ty  : type
    val : object

    def __init__(self, v: object):
        if type(v) not in {int, float}:
            raise TypeError('invalid literal valu: ' + repr(v))
        else:
            self.ty, self.val = type(v), v

    def __str__(self) -> str:
        return '#%s' % self.val

    def __repr__(self) -> str:
        return 'Lit(%s: %s)' % (self.val, self.ty)

class Reg(NamedTuple):
    name: str
    altr: str | None = None
    size: str | None = None
    mode: str | Tag | None = None
    vidx: Imm | Lit | None = None

    def __str__(self) -> str:
        if self.size is not None:
            assert self.altr is None
            assert self.mode is None
            assert self.vidx is None
            return '%s[%s]' % (self.size, self.name)
        elif self.mode is not None and self.vidx is not None:
            assert self.altr is None
            return '%s.%s[%s]' % (self.name, self.mode, self.vidx)
        elif self.mode is not None:
            assert self.altr is None
            return '%s.%s' % (self.name, self.mode)
        elif self.altr is not None:
            assert self.vidx is None
            return '(%s|%s)' % (self.name, self.altr)
        else:
            return self.name

class Vec(NamedTuple):
    mode: str | Tag
    regs: list[str]
    vidx: Imm | Lit | None = None

    def __str__(self) -> str:
        return '{ %s }.%s%s' % (
            ', '.join(self.regs),
            self.mode,
            '' if self.vidx is None else '[%s]' % self.vidx
        )

class Mod(NamedTuple):
    mod: str | Sop | Xop
    imm: tuple[Imm | Lit, bool] | None = None

    def name(self) -> str:
        if isinstance(self.mod, str):
            return self.mod
        else:
            return self.mod.name

    def __str__(self) -> str:
        if self.imm is None:
            return self.name()
        elif not self.imm[1]:
            return '%s %s' % (self.name(), self.imm[0])
        else:
            return '%s {%s}' % (self.name(), self.imm[0])

class Mem(NamedTuple):
    base   : Reg
    offs   : tuple[Imm | Reg | Lit, bool] | None = None
    index  : Literal['pre', 'post'] | None       = None
    extend : tuple[Mod, bool] | None             = None

    def __str__(self) -> str:
        ret = []
        ret.append('[')
        ret.append(self.base)

        if self.offs and self.index != 'post':
            v, o = self.offs
            ret.append('%s, %s%s' % ('{' if o else '', str(v), '}' if o else ''))

        if self.extend:
            v, o = self.extend
            ret.append('%s, %s%s' % ('{' if o else '', str(v), '}' if o else ''))

        match self.index:
            case 'pre':
                assert self.offs, 'missing index for pre index'
                ret.append(']!')

            case 'post':
                assert self.offs, 'missing index for post index'
                v, o = self.offs
                ret.append(']%s+%s%s' % ('{' if o else '', str(v), '}' if o else ''))

            case v:
                ret.append(']')
                assert v is None

        return ''.join(
            str(v)
            for v in ret
        )

class Seq(NamedTuple):
    req: list[Reg | Vec | Mem | Mod | Imm | Lit | Sym]
    opt: Reg | Vec | Mem | Mod | Imm | Lit | Sym | None = None

    def __str__(self) -> str:
        return ''.join([
            *('%s%s' % (', ' if i else '', str(v)) for i, v in enumerate(self.req)),
            (f'{{, {self.opt}}}' if self.req else f'{{{self.opt}}}') if self.opt else ''
        ])

class Instr(NamedTuple):
    mnemonic: str
    operands: Seq
    modifier: str | None

    def __str__(self) -> str:
        return ''.join([
            self.mnemonic,
            '{%s}' % self.modifier if self.modifier else '',
            ' ' if self.operands.req or self.operands.opt is not None else '',
            str(self.operands)
        ])

class Token(NamedTuple):
    name: str
    text: str

    def __repr__(self) -> str:
        return '\x1b[31m{%s}\x1b[0m' % self.name

    @classmethod
    def parse(cls, item: Element) -> 'Token':
        return cls(
            text = item.text or '',
            name = item.attrib['link'],
        )

class AsmTemplate:
    pos: int
    buf: list[str | Token]

    sops = { v.name for v in Sop }
    xops = { v.name for v in Xop }

    shifts  = { 'shift' }
    extends = { 'extend', 'extend_1' }

    amounts = {
        'amount',
        'amount_1',
        'amount_2',
        'amount_3',
        'amount_4',
        'shift',
        'shift_1',
        'shift_2',
        'shift_3',
    }

    registers = {
        'd',
        'm',
        'n',
        't',
    }

    immediates = {
        'imm',
        'imm_1',
        'imm_2',
        'imm_3',
    }

    predefined = {
        'CSYNC': (
            Sym.CSYNC,
            'CSYNC option',
            []
        ),
        'prfop': (
            Sym.PRFOP,
            'prefetch option',
            ['|', '#', 'imm5']
        ),
        'option': (
            Sym.OPTION,
            'barrier option',
            ['|', '#', 'imm']
        ),
        'systemreg': (
            Sym.SYSREG,
            'system register',
            ['|', 'S', 'op0', '_', 'op1', '_', 'cn', '_', 'cm', '_', 'op2']
        ),
        'targets': (
            Sym.TARGETS,
            'branch targets option',
            []
        ),
    }

    for v in Sym:
        assert v.value in predefined

    def __init__(self, tok: list[str | Token]):
        self.pos = 0
        self.buf = tok

    @property
    def eof(self) -> bool:
        return self.pos >= len(self.buf)

    @property
    def tok(self) -> str | Token:
        if self.eof:
            raise SyntaxError('unexpected EOF')
        else:
            return self.buf[self.pos]

    def next(self) -> str | Token:
        ret = self.tok
        self.pos += 1
        return ret

    def must(self, tok: str):
        if self.next() != tok:
            raise SyntaxError('"%s" expected' % tok)

    def skip(self, tok: str) -> bool:
        if self.eof or self.buf[self.pos] != tok:
            return False
        else:
            self.pos += 1
            return True

    def dsym(self, name: str) -> Sym:
        tok = []
        sym, msg, mat = self.predefined[name]

        for _ in mat:
            v = self.next()
            tok.append(v.name if isinstance(v, Token) else v)

        if tok != mat:
            raise SyntaxError('invalid %s: %r' % (msg, name))
        else:
            return sym

    def value(self) -> Reg | Vec | Mem | Mod | Imm | Lit | Sym:
        match self.next():
            case '#':
                match self.next():
                    case '0.0'                                   : return Lit(0.0)
                    case v if isinstance(v, str) and v.isdigit() : return Lit(int(v))
                    case v if isinstance(v, Token)               : return Imm(v.name)
                    case v                                       : raise SyntaxError('integer or token expected')

            case '(':
                reg = self.next()
                tab = self.predefined

                if not isinstance(reg, Token):
                    raise SyntaxError('register token expected')

                if reg.name in tab:
                    ret = self.dsym(reg.name)
                    self.must(')')
                    return ret

                self.must('|')
                tok = self.next()

                if not isinstance(tok, Token):
                    raise SyntaxError('register token expected')

                ret = Reg(reg.name, altr = tok.name)
                self.must(')')
                return ret

            case '[':
                buf = self.vlist()
                self.must(']')

                if not buf.req:
                    raise SyntaxError('invalid memory operand')

                idx = 'pre' if self.skip('!') else None
                args = [(v, bool(False)) for v in buf.req]

                if buf.opt is not None:
                    args.append((buf.opt, True))

                exts = None
                offs = None
                base = args[0][0]

                if not isinstance(base, Reg):
                    raise SyntaxError('memory base must be a register')

                if len(args) >= 2:
                    if not isinstance(args[1][0], (Reg, Imm, Lit)):
                        raise SyntaxError('memory offset must be a register or an immediate value')
                    else:
                        offs = (args[1][0], args[1][1])

                if len(args) >= 3:
                    if not isinstance(args[2][0], Mod):
                        raise SyntaxError('memory offset extension must be an extension')
                    else:
                        exts = (args[2][0], args[2][1])

                if len(args) >= 4:
                    raise SyntaxError('too many argumnets for memory operand')

                return Mem(
                    base   = base,
                    offs   = offs,
                    index  = idx,
                    extend = exts,
                )

            case '{{':
                regs = []
                mode = None
                vidx = None

                while True:
                    reg = self.next()
                    self.must('.')

                    if not isinstance(reg, Token):
                        raise SyntaxError('invalid vector element')

                    tok = self.next()
                    tag = Tag(tok) if isinstance(tok, str) else tok.name

                    if mode != tag and mode is not None:
                        raise SyntaxError('mode confliction within vector')

                    mode = tag
                    regs.append(reg.name)

                    match self.next():
                        case ','  : pass
                        case '}}' : break
                        case _    : raise SyntaxError('"," or "}" expected')

                if self.skip('['):
                    vidx = self.next()
                    vidx = Imm(vidx.name) if isinstance(vidx, Token) else Lit(int(vidx))
                    self.must(']')

                return Vec(
                    mode = mode,
                    regs = regs,
                    vidx = vidx,
                )

            case v if isinstance(v, str) and v in self.predefined:
                return self.dsym(v)

            case v if isinstance(v, Token) and v.name in self.immediates:
                return Imm(v.name)

            case v if isinstance(v, Token) and v.name in self.predefined:
                return self.dsym(v.name)

            case v if isinstance(v, Token) and v.name not in self.shifts and v.name not in self.extends:
                mode = None
                size = None
                vidx = None
                name = v.name

                if not self.eof:
                    if self.skip('.'):
                        mode = self.next()
                        mode = Tag(mode) if isinstance(mode, str) else mode.name

                        if self.skip('['):
                            vidx = self.next()
                            vidx = Imm(vidx.name) if isinstance(vidx, Token) else Lit(int(vidx))
                            self.must(']')

                    elif isinstance(self.tok, Token) and self.tok.name in self.registers:
                        size = v.name
                        name = self.tok.name
                        self.next()

                return Reg(
                    name = name,
                    mode = mode,
                    size = size,
                    vidx = vidx,
                )

            case v:
                opt = False
                mod = v.name if isinstance(v, Token) else v

                if isinstance(v, str):
                    if mod not in self.sops and mod not in self.xops:
                        raise SyntaxError('unexpected token: ' + repr(v))
                else:
                    if mod not in self.shifts and mod not in self.extends:
                        raise SyntaxError('unexpected token: ' + repr(v))

                if self.tok == '{':
                    opt = True
                    self.next()

                if self.tok == '#':
                    imm = self.value()
                elif isinstance(self.tok, Token):
                    imm = Imm(cast(Token, self.next()).name)
                else:
                    imm = None

                if opt and self.next() != '}':
                    raise SyntaxError('"}" expected')

                if imm is not None and not isinstance(imm, Lit):
                    if not isinstance(imm, Imm) or not imm in self.amounts:
                        raise SyntaxError('invalid extension immediate')

                if mod in self.sops:
                    mod = Sop(mod.lower())
                elif mod in self.xops:
                    mod = Xop(mod.lower())

                if imm is None:
                    return Mod(mod)
                else:
                    return Mod(mod, (imm, opt))

    def vlist(self) -> Seq:
        req = []
        opt = None

        if not self.eof and self.tok != '{':
            val = self.value()
            req.append(val)

            while self.skip(','):
                val = self.value()
                mem = req[-1] if req else None

                if not isinstance(mem, Mem):
                    req.append(val)
                elif not isinstance(val, (Reg, Imm, Lit)):
                    req.append(val)
                elif mem.offs is not None or mem.index is not None:
                    req.append(val)
                else:
                    req[-1] = Mem(mem.base, (val, False), 'post', mem.extend)

        if self.skip('{'):
            if req:
                self.must(',')

            opt = self.value()
            self.must('}')

        return Seq(
            req = req,
            opt = opt,
        )

    def instr(self) -> Instr:
        mods = None
        args = Seq([])
        name = self.next()

        if not isinstance(name, str):
            raise SyntaxError('mnemonic expected')

        if not self.eof:
            match self.tok:
                case '.' if name == 'B':
                    self.next()
                    cond = self.next()

                    if not isinstance(cond, Token):
                        raise SyntaxError('branch condition expected')
                    else:
                        mods = cond.name

                case v if isinstance(v, Token) and v.name == '2':
                    mods = '2'
                    self.next()

                case v if isinstance(v, Token) and v.name == 'bt' and name == 'BFMLAL':
                    mods = 'bt'
                    self.next()

        if not self.eof:
            args = self.vlist()

        if not self.eof:
            raise SyntaxError('junk after instruction: ' + str(self.next()))

        return Instr(
            mnemonic = name,
            operands = args,
            modifier = mods,
        )

    class Lexer:
        buf   : str
        sbuf  : str
        items : list[Element]

        def __init__(self, items: list[Element]):
            self.buf   = ''
            self.sbuf  = ''
            self.items = items

        def __iter__(self) -> Iterator[str | Token]:
            for item in self.items:
                match item.tag:
                    case 'a':
                        yield from self._parse_text()
                        yield Token.parse(item)
                        self.buf = ''

                    case 'text':
                        if item.text:
                            self.buf += item.text

                    case _:
                        raise AssertionError('unexpected tag in assembly template: ' + repr(item.tag))

            yield from self._parse_text()
            self.buf = ''

        def _parse_dots(self) -> Iterator[str]:
            if self.sbuf:
                if self.sbuf == '.' or self.sbuf[-1] != '.':
                    yield self.sbuf
                else:
                    yield self.sbuf[:-1]
                    yield self.sbuf[-1]

        def _parse_text(self) -> Iterator[str]:
            for i, ch in enumerate(self.buf):
                if ch.isalnum() or ch == '_' or (self.sbuf and ch == '.'):
                    self.sbuf += ch
                    continue

                yield from self._parse_dots()
                self.sbuf = ''

                if not ch.isspace():
                    if ch == '}' and i != 0 and self.buf[i - 1].isspace():
                        yield '}}'
                    elif ch == '{' and i != len(self.buf) - 1 and self.buf[i + 1].isspace():
                        yield '{{'
                    else:
                        yield ch

            yield from self._parse_dots()
            self.sbuf = ''

    @classmethod
    def parse(cls, items: list[Element]) -> Instr:
        return cls(list(cls.Lexer(items))).instr()

class InstrForm(NamedTuple):
    text   : str
    inst   : Instr
    bits   : Instruction
    opts   : dict[str, str]
    args   : list[str | int]
    enctab : EncodingTabEntry
    fields : dict[str, Account | Definition]

def permu_bits(bit: str) -> str:
    match bit:
        case '0': return '0'
        case '1': return '1'
        case 'x': return '01'
        case _  : raise AssertionError('invalid bit value: ' + repr(bit))

def parse_props(out: dict[str, str], p: Element):
    for v in p.findall('docvars/docvar'):
        out[v.attrib['key']] = v.attrib['value']

def parse_symdef(defs: Element) -> dict[str, set[str]]:
    rets = {}
    tabs = defs.findall('.//tgroup')
    rows = defs.findall('.//tbody/row')
    dest = defs.attrib['encodedin']

    if len(tabs) != 1:
        raise AssertionError('expect exactly 1 table: encodedin="%s"' % dest)

    for row in rows:
        bits = []
        syms = row.find('entry[@class="symbol"]')

        for bv in row.findall('entry[@class="bitfield"]'):
            if not bv.text:
                raise AssertionError('missing symbol or bitfield: encodedin="%s"' % dest)
            else:
                bits.append(bv.text)

        if not bits or syms is None or not syms.text:
            raise AssertionError('missing symbol or bitfield: encodedin="%s"' % dest)

        for sym in map(str.strip, syms.text.split('|')):
            prod = itertools.product(*map(permu_bits, ''.join(bits)))
            rets.setdefault(sym, set()).update(''.join(v) for v in prod)

    if not rets:
        raise AssertionError('missing definitions: encodedin="%s"' % dest)
    else:
        return rets

maxargs = 0
formtab = dict[str, list[InstrForm]]()
fieldtab = dict[str, dict[str, Account | Definition]]()

for expl in isadocs.findall('.//explanation'):
    symbol = expl.find('symbol')
    symacc = expl.find('account')
    symdef = expl.find('definition')

    if symbol is None or (symacc is None) is (symdef is None):
        raise AssertionError('invalid explanation')

    desc = symbol.text or ''
    name = symbol.attrib['link']
    status('* Field Explanation:', name)

    if symacc is not None:
        assert symdef is None
        defs = Account(symacc.attrib['encodedin'], desc)
    else:
        assert isinstance(symdef, Element)
        defs = Definition(symdef.attrib['encodedin'], desc, parse_symdef(symdef))

    for enc in expl.attrib['enclist'].split(','):
        tab = fieldtab.setdefault(enc.strip(), {})
        tab[name] = defs

for encdata in sorted(enctab.values(), key = lambda x: x.name):
    node = isadocs.find('.//iclass/encoding[@name="%s"]' % encdata.name)
    assert node is not None, 'encoding %s does not exists' % repr(encdata.name)

    tokens = node.findall('asmtemplate/*')
    assert tokens, 'encoding %s does not have an assembly syntax' % repr(encdata.name)

    text = ''.join(v.text or '' for v in tokens)
    status('* Assembly Template:', text)

    args = []
    opts = {}
    bits = Instruction(encdata.bits)
    inst = AsmTemplate.parse(tokens)

    parse_props(opts, node)
    bits.update(node.findall('box'), no_add = True)
    bits.update(parent_tab[node].findall('regdiagram/box'), no_add = True)
    assert inst.mnemonic == opts['mnemonic']

    if inst.operands.opt is None:
        maxargs = max(maxargs, len(inst.operands.req))
    else:
        maxargs = max(maxargs, len(inst.operands.req) + 1)

    req = list(bits.refs.items())
    req.sort(key = lambda x: x[1], reverse = True)

    for v, (p, n) in req:
        if None in bits.bits[p:p + n]:
            args.append(v)
        else:
            args.append(int(''.join(map(str, bits.bits[p:p + n][::-1])), 2))

    formtab.setdefault(inst.mnemonic, []).append(InstrForm(
        text   = text,
        inst   = inst,
        bits   = bits,
        opts   = opts,
        args   = args,
        enctab = encdata,
        fields = fieldtab.get(encdata.name, {})
    ))

### ---------- Per-instruction Encoding ---------- ###

class Or(list['And | Or | str']):
    def __init__(self, *terms: 'And | Or | str'):
        super().__init__(terms)

    def __str__(self) -> str:
        return ' || '.join(str(v) for v in self)

class And(list['And | Or | str']):
    def __init__(self, *terms: 'And | Or | str'):
        super().__init__(terms)

    def __str__(self) -> str:
        if len(self) == 1 and isinstance(self[0], Or):
            return str(self[0])
        else:
            return ' && '.join('(%s)' % v if isinstance(v, Or) else str(v) for v in self)

class OnceDict(OrderedDict):
    def __setitem__(self, k, v) -> None:
        if k not in self:
            super().__setitem__(k, v)

SYM_CHECKS = {
    Sym.CSYNC   : 'isCSync(%s)',
    Sym.PRFOP   : 'isPrefetch(%s)',
    Sym.OPTION  : 'isOptions(%s)',
    Sym.SYSREG  : 'isSysReg(%s)',
    Sym.TARGETS : 'isTargets(%s)',
}

IMM_CHECKS = {
    'CRm'             : 'isUimm4(%s)',
    'N:immr:imms'     : 'isMask64(%s)',
    'a:b:c:d:e:f:g:h' : 'isUimm8(%s)',
    'imm5'            : 'isUimm5(%s)',
    'imm9'            : 'isImm9(%s)',
    'imm12'           : 'isImm12(%s)',
    'imm16'           : 'isUimm16(%s)',
    'immr'            : 'isUimm6(%s)',
    'immr:imms'       : 'isMask32(%s)',
    'imms'            : 'isUimm6(%s)',
    'nzcv'            : 'isUimm4(%s)',
    'scale'           : 'isUimm6(%s)',
    'uimm4'           : 'isUimm4(%s)',
    'uimm6'           : 'isUimm6(%s)',
}

REG_CHECKS = {
    'cond'   : 'isBrCond(%s)',
    'bt'     : 'isBr(%s)',
    'dd'     : 'isDr(%s)',
    'dm'     : 'isDr(%s)',
    'dn'     : 'isDr(%s)',
    'dn_1'   : 'isDr(%s)',
    'dt'     : 'isDr(%s)',
    'hd'     : 'isHr(%s)',
    'hm'     : 'isHr(%s)',
    'hn'     : 'isHr(%s)',
    'hn_1'   : 'isHr(%s)',
    'ht'     : 'isHr(%s)',
    'qt'     : 'isQr(%s)',
    'sd'     : 'isSr(%s)',
    'sm'     : 'isSr(%s)',
    'sn'     : 'isSr(%s)',
    'sn_1'   : 'isSr(%s)',
    'st'     : 'isSr(%s)',
    'vd'     : 'isVr(%s)',
    'vm'     : 'isVr(%s)',
    'vn'     : 'isVr(%s)',
    'wd'     : 'isWr(%s)',
    'wd_wsp' : 'isWrOrWSP(%s)',
    'wm'     : 'isWr(%s)',
    'wn'     : 'isWr(%s)',
    'wn_wsp' : 'isWrOrWSP(%s)',
    'ws'     : 'isWr(%s)',
    'wt'     : 'isWr(%s)',
    'xd'     : 'isXr(%s)',
    'xd_sp'  : 'isXrOrSP(%s)',
    'xm'     : 'isXr(%s)',
    'xm_sp'  : 'isXrOrSP(%s)',
    'xn'     : 'isXr(%s)',
    'xn_sp'  : 'isXrOrSP(%s)',
    'xs'     : 'isXr(%s)',
    'xt'     : 'isXr(%s)',
}

REG_CHECKS_MERGED = {
    'isWr(%s) || isXr(%s)' : 'isWrOrXr(%s)',
    'isXr(%s) || isSP(%s)' : 'isXrOrSP(%s)',
    'isWr(%s) || isWSP(%s)': 'isWrOrWSP(%s)',
}

COMBREG_CHECKS = {
    'd': 'isWrOrXr(%s)',
    'm': 'isWrOrXr(%s)',
    'n': 'isWrOrXr(%s)',
}

FIXEDVEC_CHECKS = {
    'd': 'isAdvSIMD(%s)',
    'm': 'isAdvSIMD(%s)',
    'n': 'isAdvSIMD(%s)',
}

SIGNED_IMM = {
    'simm',
    'pimm',
    'pimm_1',
    'pimm_2',
    'pimm_3',
    'pimm_4',
}

UNSIGNED_IMM = {
    'uimm',
}

SCALAR_TYPES = {
    'B': 'SIMDRegister8',
    'H': 'SIMDRegister16',
    'S': 'SIMDRegister32',
    'D': 'SIMDRegister64',
    'Q': 'SIMDRegister128',
}

VECTOR_TYPES = {
    '8B'  : 'Vec8B',
    '16B' : 'Vec16B',
    '4H'  : 'Vec4H',
    '8H'  : 'Vec8H',
    '2S'  : 'Vec2S',
    '4S'  : 'Vec4S',
    '1D'  : 'Vec1D',
    '2D'  : 'Vec2D',
}

def match_modifier(name: str, mod: Mod, optional: bool, *extra_cond: str) -> Iterator[Or | str]:
    if not optional:
        if isinstance(mod.mod, Sop):
            yield 'isSameMod(%s, %s(0))' % (name, mod.mod.name)
        elif isinstance(mod.mod, Xop):
            yield 'isSameMod(%s, %s(0))' % (name, mod.mod.name)
        elif mod.mod in AsmTemplate.shifts or mod.mod in AsmTemplate.extends:
            yield 'isMod(%s)' % name
        else:
            raise RuntimeError('invalid extension')

        if mod.imm is None:
            yield 'modn(%s) == 0' % name

        else:
            imm = mod.imm[0]
            opm = mod.imm[1]

            if isinstance(imm, Lit):
                if imm.ty is not int:
                    raise RuntimeError('invalid literal type for amounts')
                elif not opm or imm.val == 0:
                    yield 'modn(%s) == %d' % (name, imm.val)
                else:
                    yield Or('modn(%s) == 0' % name, 'modn(%s) == %d' % (name, imm.val))

    else:
        if isinstance(mod.mod, (Sop, Xop)):
            base = 'isSameMod(%s, %s(0))' % (name, mod.mod.name)
        elif mod.mod in AsmTemplate.shifts:
            base = 'isShift(%s)' % name
        elif mod.mod in AsmTemplate.extends:
            base = 'isExtend(%s)' % name
        else:
            raise RuntimeError('invalid extension')

        if mod.imm is None:
            yield Or(*extra_cond, And(base, 'modn(%s) == 0' % name))

        else:
            imm = mod.imm[0]
            opm = mod.imm[1]

            if not isinstance(imm, Lit):
                yield Or(*extra_cond, base)
            elif imm.ty is not int:
                raise RuntimeError('invalid literal type for amounts')
            elif not opm or imm.val == 0:
                yield Or(*extra_cond, And(base, 'modn(%s) == %d' % (name, imm.val)))
            else:
                yield Or(*extra_cond, And(base, Or('modn(%s) == 0' % name, 'modn(%s) == %d' % (name, imm.val))))

def match_operands(form: InstrForm, argc: int) -> Iterator['And | Or | str']:
    argv = form.inst.operands.req[:]
    dynvec = {}
    fixedvec = {}
    # TODO: remove this
    print('------- match operand -------')
    print(form.inst)

    if form.inst.operands.opt is not None:
        argv.append(form.inst.operands.opt)

    if len(argv) > argc:
        if form.inst.operands.opt is None:
            yield 'len(vv) == %d' % (len(argv) - argc)
        else:
            yield 'len(vv) <= %d' % (len(argv) - argc)
            yield 'len(vv) >= %d' % (len(argv) - argc - 1)

    for i, val in enumerate(argv):
        name = 'v%d' % i if i < argc else 'vv[%d]' % (i - argc)
        optcond = []

        if i >= len(form.inst.operands.req):
            if i == argc:
                optcond.append('len(vv) == 0')
            else:
                optcond.append('len(vv) <= %d' % (i - argc))

        if isinstance(val, Reg):
            if optcond:
                raise RuntimeError('optional reg operand is not supported')

            elif val.size is not None:
                if val.size == 'r':
                    yield COMBREG_CHECKS[val.name] % name
                else:
                    yield FIXEDVEC_CHECKS[val.name] % name
                    fixedvec.setdefault(val.size, []).append(name)

            elif val.mode is not None and val.vidx is not None:
                # TODO: this
                print('%s.%s[%s]' % (val.name, val.mode, val.vidx))
                raise NotImplementedError('indexed vector')

            elif val.mode is not None:
                if isinstance(val.mode, Tag):
                    yield 'isVr(%s)' % name
                    yield 'vfmt(%s) == %s' % (name, VECTOR_TYPES[val.mode.name])

                else:
                    yield 'isVr(%s)' % name
                    field = form.fields[val.mode]
                    dynvec.setdefault(val.mode, []).append(name)

                    if isinstance(field, Definition):
                        if not field.bits:
                            raise RuntimeError('no definitions')
                        elif len(field.bits) == 1:
                            yield 'vfmt(%s) == %s' % (name, VECTOR_TYPES[list(field.bits)[0]])
                        else:
                            yield 'isVfmt(%s, %s)' % (name, ', '.join(VECTOR_TYPES[x] for x in field.bits if x != 'RESERVED'))

            elif val.altr is not None:
                c1 = REG_CHECKS[val.name]
                c2 = REG_CHECKS[val.altr]
                v1 = REG_CHECKS_MERGED.get('%s || %s' % (c1, c2))
                v2 = REG_CHECKS_MERGED.get('%s || %s' % (c2, c1))

                if v1 is not None:
                    yield v1 % name
                elif v2 is not None:
                    yield v2 % name
                else:
                    yield Or(c1 % name, c2 % name)

            elif val.name == 'label':
                yield 'isLabel(%s)' % name

            else:
                yield REG_CHECKS[val.name] % name

        elif isinstance(val, Vec):
            if optcond:
                raise RuntimeError('optional vector operand is not supported')

            # TODO: handle vector
            print(val)
            raise NotImplementedError('vec operand')

        elif isinstance(val, Mem):
            if optcond:
                raise RuntimeError('optional mem operand is not supported')

            yield 'isMem(%s)' % name
            yield REG_CHECKS[val.base.name] % ('mbase(%s)' % name)

            if val.offs is None:
                yield 'midx(%s) == nil' % name
                yield 'moffs(%s) == 0' % name

            else:
                off = val.offs[0]
                opt = val.offs[1]

                if isinstance(off, Imm):
                    if off in SIGNED_IMM:
                        yield 'midx(%s) == nil' % name
                    elif off in UNSIGNED_IMM:
                        yield from ['midx(%s) == nil' % name, 'moffs(%s) >= 0' % name]
                    else:
                        raise RuntimeError('invalid offset type ' + repr(off))

                elif isinstance(off, Reg):
                    key = 'midx(%s)' % name
                    yield 'moffs(%s) == 0' % name

                    if off.size is not None:
                        raise RuntimeError('invalid offset register: vector')

                    if off.mode is not None:
                        raise RuntimeError('invalid offset register: dyn or indexed vector')

                    if not opt:
                        if off.altr is None:
                            yield REG_CHECKS[off.name] % key

                        else:
                            c1 = REG_CHECKS[off.name]
                            c2 = REG_CHECKS[off.altr]
                            v1 = REG_CHECKS_MERGED.get('%s || %s' % (c1, c2))
                            v2 = REG_CHECKS_MERGED.get('%s || %s' % (c2, c1))

                            if v1 is not None:
                                yield v1 % key
                            elif v2 is not None:
                                yield v2 % key
                            else:
                                yield Or(c1 % key, c2 % key)

                else:
                    if off.ty is not int:
                        raise RuntimeError('invalid literal type for offsets')
                    elif not opt:
                        yield from ['midx(%s) == nil' % name, 'moffs(%s) == %d' % (name, off.val)]
                    else:
                        yield from ['midx(%s) == nil' % name, Or('moffs(%s) == 0' % name, 'moffs(%s) == %d' % (name, off.val))]

            if val.extend is None:
                match val.index:
                    case None   : yield 'mext(%s) == nil' % name
                    case 'pre'  : yield 'mext(%s) == PreIndex' % name
                    case 'post' : yield 'mext(%s) == PostIndex' % name
                    case _      : raise RuntimeError('invalid memory index')

            else:
                if val.index is not None:
                    raise RuntimeError('extension conflits with indexing')
                else:
                    yield from match_modifier('mext(%s)' % name, val.extend[0], val.extend[1], 'mext(%s) == nil' % name)

        elif isinstance(val, Mod):
            yield from match_modifier(name, val, bool(optcond), *optcond)

        elif isinstance(val, Imm):
            yield Or(*optcond, IMM_CHECKS[form.fields[val].name] % name)

        elif isinstance(val, Lit):
            if optcond:
                raise RuntimeError('optional lit operand is not supported')
            else:
                yield '%s == %s' % (name, val.val)

        elif isinstance(val, Sym):
            yield Or(*optcond, SYM_CHECKS[val] % name)

        else:
            raise RuntimeError('invalid operand type')

    yield from (
        'isSameSize(%s, %s)' % (v[i], v[i + 1])
        for v in dynvec.values()
        for i in range(len(v) - 1)
    )

    yield from (
        'isSameType(%s, %s)' % (v[i], v[i + 1])
        for v in fixedvec.values()
        for i in range(len(v) - 1)
    )

IMM_ENCODER = {
    'CRm'             : 'asUimm4',
    'N:immr:imms'     : 'asMaskImm',
    'a:b:c:d:e:f:g:h' : 'asUimm8',
    'imm5'            : 'asUimm5',
    'imm12'           : 'asImm12',
    'imm16'           : 'asUimm16',
    'immr'            : 'asUimm6',
    'immr:imms'       : 'asMaskImm',
    'imms'            : 'asUimm6',
    'nzcv'            : 'asUimm4',
    'scale'           : 'asUimm6',
    'uimm4'           : 'asUimm4',
    'uimm6'           : 'asUimm6',
}

OPT_DEFAULTS = {
    'CLREX_BN_barriers': {
        'imm': 'uint32(0b1111)',
    }
}

SPECIAL_ENCODERS = {
    'cond'  : 'uint32(%s.(BranchCondition))',
    'label' : '%s.(*asm.Label)',
}

def encode_defs(
    form     : InstrForm,
    sw_cond  : str,
    var_name : str,
    name_tab : dict[str, str],
    vals     : dict[str, str | tuple[str, str, dict[str, str]]],
    opts     : dict[str, str | list],
    err_msg  : str,
):
    cond = []
    field = form.fields[var_name]

    if not isinstance(field, Definition):
        raise RuntimeError('field should be definitions')

    opts[var_name] = cond
    vals[var_name] = sw_cond

    for k, v in field.bits.items():
        if k != 'RESERVED':
            if len(v) != 1:
                raise RuntimeError('ambiguous encoding')
            else:
                cond.append('case %s: %s = 0b%s' % (name_tab[k], var_name, next(iter(v))))

    if not cond:
        raise RuntimeError('not encodable: ' + form.inst.mnemonic)
    else:
        cond.append('default: panic("aarch64: %s")' % err_msg)

def encode_modifier(
    name        : str,
    mod         : Mod,
    vals        : dict[str, str | tuple[str, str, dict[str, str]]],
    opts        : dict[str, str | list],
    optional    : bool,
    *extra_cond : str,
):
    if optional and not isinstance(mod.mod, (Sop, Xop)):
        opts[mod.mod] = str(And(*extra_cond))

    if mod.mod in AsmTemplate.shifts:
        vals[mod.mod] = 'uint32(%s.(ShiftType).ShiftType())' % name
    elif mod.mod in AsmTemplate.extends:
        vals[mod.mod] = 'uint32(%s.(Extension).Extension())' % name

    if mod.imm is not None and isinstance(mod.imm[0], Imm):
        args = mod.imm[0].name
        vals[args] = 'uint32(%s.(Modifier).Amount())' % name

        if optional:
            opts[args] = str(And(*extra_cond))

def encode_operand(
    form     : InstrForm,
    name     : str,
    val      : Reg | Vec | Mem | Mod | Imm | Lit | Sym,
    vals     : dict[str, str | tuple[str, str, dict[str, str]]],
    opts     : dict[str, str | list],
    *optcond : str,
):
    if isinstance(val, Reg):
        if optcond:
            raise RuntimeError('optional reg operand is not supported')

        elif val.mode is not None and val.vidx is not None:
            # TODO: this
            print('%s.%s[%s]' % (val.name, val.mode, val.vidx))
            raise NotImplementedError('indexed vector')

        elif val.mode is not None:
            if isinstance(val.mode, Tag):
                vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name
            else:
                vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name
                encode_defs(form, 'vfmt(%s)' % name, val.mode, VECTOR_TYPES, vals, opts, 'unreachable')

        elif val.size is not None and val.size != 'r':
            if not val.size.startswith('v'):
                raise RuntimeError('invalid fixed vec size')

            err = 'invalid scalar operand size for ' + form.inst.mnemonic
            vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name
            encode_defs(form, '%s.(type)' % name, val.size, SCALAR_TYPES, vals, opts, err)

        elif val.name in SPECIAL_ENCODERS:
            vals[val.name] = SPECIAL_ENCODERS[val.name] % name

        else:
            vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name

    elif isinstance(val, Vec):
        if optcond:
            raise RuntimeError('optional vector operand is not supported')

        # TODO: handle vector
        print(val)
        raise NotImplementedError('vec operand')

    elif isinstance(val, Mem):
        if optcond:
            raise RuntimeError('optional mem operand is not supported')
        elif val.base.mode is not None or val.base.vidx is not None:
            raise RuntimeError('vector base is not supported')
        else:
            vals[val.base.name] = 'uint32(mbase(%s).ID())' % name

        if val.offs is not None:
            off = val.offs[0]
            opt = val.offs[1]

            if isinstance(off, Imm):
                vals[off.name] = 'uint32(moffs(%s))' % name

            elif isinstance(off, Reg):
                if opt:
                    raise RuntimeError('optional offset register is not supported')
                elif off.size is not None:
                    raise RuntimeError('invalid offset register: vector')
                elif off.mode is not None:
                    raise RuntimeError('invalid offset register: dyn or indexed vector')
                elif off.altr is not None:
                    vals[off.altr] = 'uint32(midx(%s).ID())' % name
                else:
                    vals[off.name] = 'uint32(midx(%s).ID())' % name

        if val.extend is not None:
            ext = val.extend[0]
            opx = val.extend[1]

            if val.index is not None:
                raise RuntimeError('extension conflits with indexing')
            else:
                encode_modifier('mext(%s)' % name, ext, vals, opts, opx, 'isMod(mext(%s))' % name)

    elif isinstance(val, Mod):
        encode_modifier(name, val, vals, opts, bool(optcond), *optcond)

    elif isinstance(val, Imm):
        ref = form.fields[val.name]
        enc = IMM_ENCODER[ref.name]

        if not optcond:
            vals[val.name] = '%s(%s)' % (enc, name)

        else:
            defv = OPT_DEFAULTS.get(form.enctab.name, {}).get(val.name)
            opts[val.name] = str(And(*optcond))

            if defv is None:
                vals[val.name] = '%s(%s)' % (enc, name)
            else:
                vals[val.name] = ('%s(%s)' % (enc, name), defv, {})

    elif isinstance(val, Lit):
        if optcond:
            raise RuntimeError('optional lit operand is not supported')

    elif isinstance(val, Sym):
        match val:
            case Sym.CSYNC:
                if optcond:
                    raise RuntimeError('optional CSYNC is not supported')

            case Sym.PRFOP:
                raise NotImplementedError('prfop sym')

            case Sym.OPTION:
                vals['option'] = ('%s.(BarrierOption)' % name, 'SY', {})
                vals['imm'] = 'uint32(option)'

                if optcond:
                    opts['option'] = str(And(*optcond))

            case Sym.SYSREG:
                raise NotImplementedError('sysreg sym')

            case Sym.TARGETS:
                vals['targets'] = ('%s.(BranchTarget)' % name, '_BrOmitted', {
                    '(omitted)' : '_BrOmitted',
                    'c'         : 'BrC',
                    'j'         : 'BrJ',
                    'jc'        : 'BrJC',
                })

                if optcond:
                    opts['targets'] = str(And(*optcond))

            case _:
                raise RuntimeError('invalid symbol')

    else:
        raise RuntimeError('invalid operand type')

FIELD_ALIASES = {
    'ADDG_64_addsub_immtags': {
        'Rn': 'Xn',
        'Rd': 'Xd',
    },
    'BLRAA_64P_branch_reg': {
        'op4': 'Rm',
    },
    'BLRAB_64P_branch_reg': {
        'op4': 'Rm',
    },
    'BRAA_64P_branch_reg': {
        'op4': 'Rm',
    },
    'BRAB_64P_branch_reg': {
        'op4': 'Rm',
    },
    'FCMP_DZ_floatcmp': {
        'Rn': 'dn',
    },
}

cc = CodeGen()
cc.line('// Code generated by "mkasm_aarch64.py", DO NOT EDIT.')
cc.line()
cc.line('package aarch64')
cc.line()
cc.line('import (')
cc.indent()
cc.line('`github.com/chenzhuoyu/iasm/asm`')
cc.dedent()
cc.line(')')
cc.line()
cc.line('const (')
cc.indent()
cc.line('_N_args = %d' % maxargs)
cc.dedent()
cc.line(')')
cc.line()

for mnemonic, forms in sorted(formtab.items(), key = lambda x: x[0]):
    nops = set()
    always = False
    status('* Instruction:', mnemonic)

    if len(forms) == 1:
        cc.line('// %s instruction have one single form:' % mnemonic)
        cc.line('//')
    else:
        cc.line('// %s instruction have %d forms:' % (mnemonic, len(forms)))
        cc.line('//')

    for form in forms:
        nop = len(form.inst.operands.req)
        nops.add(nop)
        cc.line('//   * %s' % form.text)

        if form.inst.operands.opt is not None:
            nops.add(nop + 1)

    cc.line('//')
    nfix = min(nops)
    base = ['v%d' % i for i in range(nfix)]

    if nfix == max(nops):
        if not nfix:
            cc.line('func (self *Program) %s() *Instruction {' % mnemonic)
            cc.indent()
            cc.line('p := self.alloc("%s", 0, Operands {})' % mnemonic)
        else:
            cc.line('func (self *Program) %s(%s interface{}) *Instruction {' % (mnemonic, ', '.join(base)))
            cc.indent()
            cc.line('p := self.alloc("%s", %d, Operands { %s })' % (mnemonic, nfix, ', '.join(base)))

    else:
        if not nfix:
            cc.line('func (self *Program) %s(vv ...interface{}) *Instruction {' % mnemonic)
        else:
            cc.line('func (self *Program) %s(%s interface{}, vv ...interface{}) *Instruction {' % (mnemonic, ', '.join(base)))

        cc.indent()
        cc.line('var p *Instruction')
        cc.line('switch len(vv) {')
        cc.indent()

        for argc in sorted(nops):
            if not argc:
                assert nfix == 0
                cc.line('case 0  : p = self.alloc("%s", 0, Operands {})' % mnemonic)
            else:
                args = base[:] + ['vv[%d]' % i for i in range(argc - nfix)]
                cc.line('case %d  : p = self.alloc("%s", %d, Operands { %s })' % (argc - nfix, mnemonic, argc, ', '.join(args)))

        cc.line('default : panic("instruction %s takes %s operands")' % (mnemonic, ' or '.join(map(str, sorted(nops)))))
        cc.dedent()
        cc.line('}')

    for form in forms:
        mt = And(*match_operands(form, nfix))
        mts = str(mt)

        if len(forms) != 1:
            cc.line('// ' + form.text)

        if not mt:
            if len(forms) == 1:
                always = True
            else:
                raise RuntimeError('multiple instruction forms while no matching condition')

        elif len(mt) == 1 or len(mts) + cc.level * 4 + 5 <= 120:
            cc.line('if %s {' % mts)
            cc.indent()

        else:
            if not isinstance(mt[0], Or):
                cc.line('if %s &&' % mt[0])
            else:
                cc.line('if (%s) &&' % mt[0])

            for i, mm in enumerate(mt[1:], 1):
                if i == len(mt) - 1:
                    cc.line('   (%s) {' % mm if isinstance(mm, Or) else '   %s {' % mm)
                else:
                    cc.line('   (%s) &&' % mm if isinstance(mm, Or) else '   %s &&' % mm)
            else:
                cc.indent()

        fmap = {}
        args = []

        if form.inst.modifier not in {None, '2'}:
            raise RuntimeError('invalid instruction modifier')

        print(form)
        for key, fv in form.fields.items():
            nb = 0
            fn = []
            st = False

            for s in fv.name.split(':'):
                if st:
                    st = '>' not in s
                    fn[-1] += ':' + s
                else:
                    st = '<' in s
                    fn.append(s)

            if fv.name == form.inst.modifier:
                continue

            if len(fn) == 1:
                fmap.setdefault(fv.name, []).append((key, key))
                continue

            for x in fn:
                _, n = form.bits.refs[x]
                nb += n

            for x in fn:
                _, n = form.bits.refs[x]
                nb -= n

                if nb == 0:
                    fmap.setdefault(x, []).append((key, '%s & 0b%s' % (key, '1' * n)))
                else:
                    fmap.setdefault(x, []).append((key, '(%s >> %d) & 0b%s' % (key, nb, '1' * n)))

        for i, v in enumerate(form.args):
            if not isinstance(v, str):
                args.append(str(v))
                continue

            if v not in fmap and form.enctab.name in FIELD_ALIASES and v in FIELD_ALIASES[form.enctab.name]:
                v = FIELD_ALIASES[form.enctab.name][v]
                form.args[i] = v

            if v not in fmap:
                args.append(v)
            elif all(x[1] != 'label' for x in fmap[v]):
                args.append(fmap[v][-1][1])
            elif len(fmap[v]) == 1:
                args.append('uint32(label.RelativeTo(pc))')
            else:
                raise RuntimeError('multiple lables')

        vals = OnceDict()
        opts = OnceDict()
        lastcond = None

        for i, val in enumerate(form.inst.operands.req):
            if i < nfix:
                encode_operand(form, 'v%d' % i, val, vals, opts)
            else:
                encode_operand(form, 'vv[%d]' % (i - nfix), val, vals, opts)

        if form.inst.operands.opt is not None:
            if not isinstance(form.inst.operands.opt, (Reg, Mod, Imm, Sym)):
                raise RuntimeError('invalid optional operand')
            else:
                idx = len(form.inst.operands.req) - nfix
                encode_operand(form, 'vv[%d]' % idx, form.inst.operands.opt, vals, opts, 'len(vv) > %d' % idx)

        for var in sorted(opts):
            if not isinstance(vals[var], tuple):
                cc.line('var %s uint32' % var)
            else:
                cc.line('%s := %s' % (var, vals[var][1]))

        for k, v in vals.items():
            if isinstance(v, tuple):
                v = v[0]

            if k not in opts:
                if lastcond is not None:
                    cc.dedent()
                    cc.line('}')

                cc.line('%s := %s' % (k, v))
                lastcond = None

            elif opts[k] == lastcond:
                cc.line('%s = %s' % (k, v))

            else:
                if lastcond is not None:
                    cc.dedent()
                    cc.line('}')
                    lastcond = None

                if isinstance(opts[k], str):
                    lastcond = opts[k]
                    cc.line('if %s {' % opts[k])
                    cc.indent()
                    cc.line('%s = %s' % (k, v))

                elif isinstance(opts[k], list):
                    cc.line('switch %s {' % v)
                    cc.indent()

                    for line in opts[k]:
                        cc.line(line)

                    cc.dedent()
                    cc.line('}')

                else:
                    raise RuntimeError('invalid optional cond')

        if lastcond is not None:
            cc.dedent()
            cc.line('}')

        for arg in form.args:
            if isinstance(arg, str) and arg not in fmap:
                ok = False
                i, n = form.bits.refs[arg]
                bits = form.bits.bits[i:i + n]
                cc.line('%s := uint32(0b%s)' % (arg, ''.join('1' if v else '0' for v in bits)))

                for bn, fv in fmap.items():
                    if bn.endswith('>') and bn.startswith(arg + '<'):
                        if len(fv) != 1 or fv[0][0] != fv[0][1]:
                            raise RuntimeError('composite field encoding confliction: ' + repr(bn))

                        ok = True
                        fn = fv[0][0]
                        fv = form.fields[fn]
                        bv = bn[len(arg) + 1:-1]

                        if ':' not in bv:
                            be, bs = int(bv), int(bv)
                        else:
                            be, bs = map(int, bn[len(arg) + 1:-1].split(':'))

                        if not isinstance(fv, Definition):
                            raise RuntimeError('field definition required for composite field ' + repr(bn))

                        cc.line('switch %s {' % fn)
                        cc.indent()

                        for sel, bits in fv.bits.items():
                            if len(bits) != 1:
                                raise RuntimeError('multiple definitions for compositie field ' + repr(bn))

                            bc = be - bs + 1
                            bx = list(bits)[0]

                            if len(bx) != bc:
                                raise RuntimeError('bit count mismatch for composite field ' + repr(bn))

                            if not isinstance(vals[fn], tuple):
                                cc.line('case %s: %s |= 0b%s << %d' % (sel, arg, bx, bs))
                            else:
                                cc.line('case %s: %s |= 0b%s << %d' % (vals[fn][2][sel], arg, bx, bs))

                        cc.line('default: panic("aarch64: invalid combination of operands for %s")' % mnemonic)
                        cc.dedent()
                        cc.line('}')

                if not ok:
                    print(form)  # TODO: remove this
                    raise RuntimeError('invalid field ' + repr(arg))

        for fv in fmap.values():
            terms = []
            exprs = [v[1] for v in sorted(fv, key = lambda v: v[0]) if v[0] in opts or v[0] in vals]

            if len(exprs) > 1:
                for i, v in enumerate(exprs[1:]):
                    terms.append('%s != %s' % (exprs[i], v))

                cc.line('if %s {' % ' || '.join(terms))
                cc.indent()
                cc.line('panic("aarch64: invalid combination of operands for %s")' % mnemonic)
                cc.dedent()
                cc.line('}')

        encoding = '%s(%s)' % (form.enctab.func, ', '.join(args))
        deferred = any(v == Reg('label') for v in form.inst.operands.req)

        if not deferred:
            if len(encoding) + cc.level * 4 + 10 <= 120:
                cc.line('p.setins(%s)' % encoding)
            else:
                cc.line('p.setins(%s(' % form.enctab.func)
                cc.indent()

                for arg in args:
                    cc.line(arg + ',')

                cc.dedent()
                cc.line('))')

        else:
            if len(encoding) + cc.level * 4 + 45 <= 120:
                cc.line('p.setenc(func(pc uintptr) uint32 { return %s })' % encoding)
            else:
                cc.line('p.setenc(func(pc uintptr) uint32 {')
                cc.indent()
                cc.line('return %s(' % form.enctab.func)
                cc.indent()

                for arg in args:
                    cc.line(arg + ',')

                cc.dedent()
                cc.line(')')
                cc.dedent()
                cc.line('})')

        if mt:
            cc.dedent()
            cc.line('}')

    if not always:
        cc.line('if !p.isvalid {')
        cc.indent()
        cc.line('panic("aarch64: invalid combination of operands for %s")' % mnemonic)
        cc.dedent()
        cc.line('}')

    cc.line('return p')
    cc.dedent()
    cc.line('}')
    cc.line()

with open('arch/aarch64/instructions.go', 'w') as fp:
    fp.write('\n'.join(cc.buf))
