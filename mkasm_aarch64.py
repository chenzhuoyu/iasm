#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import re
import sys
import textwrap

from typing import cast
from typing import Literal
from typing import Iterator
from typing import NamedTuple

from enum import StrEnum
from functools import cached_property
from collections import OrderedDict

from xml.etree import ElementTree
from xml.etree.ElementTree import Element

DOC_FILE       = 'isa_docs/ISA_A64_xml_A_profile-2022-12_OPT/onebigfile.xml'
MAX_TEXT_WIDTH = 80
MAX_LINE_WIDTH = 120

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
    defv: str

    def __init__(self, name: str, desc: str, defv: str):
        self.name = name
        self.desc = desc
        self.defv = defv

    def __repr__(self) -> str:
        if not self.defv:
            return '%s %s' % (self.name, self.desc)
        else:
            return '%s %s (default = %s)' % (self.name, self.desc, self.defv)

class Definition:
    name: str
    desc: str
    defv: str
    refs: list[str]
    bits: dict[str, set[str]]

    def __init__(self, name: str, desc: str, defv: str, refs: list[str], bits: dict[str, set[str]]):
        self.name = name
        self.desc = desc
        self.defv = defv
        self.refs = refs
        self.bits = bits

    def __repr__(self) -> str:
        return '%s %s (=%s) [%s]%s' % (
            self.name,
            self.desc,
            ':'.join(self.refs[::-1]),
            ' '.join('%s={%s}' % (k, ':'.join(sorted(v))) for k, v in self.bits.items()),
            '' if not self.defv else ' (default = %s)' % self.defv
        )

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

    def update(self, boxes: list[Element]):
        for box in boxes:
            bits = box.findall('c')
            name = box.attrib.get('name')
            hibit = int(box.attrib['hibit'])
            width = int(box.attrib.get('width', 1))

            if name is not None:
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
    args: list[str]
    bits: Instruction
    link: str | None = None

    def __repr__(self) -> str:
        return '\n'.join([
            'Encoding %s (%s) {' % (self.name, ', '.join(self.args)),
            '    ' + self.desc if self.link is None else '    %s (alias of %s)' % (self.desc, self.link),
            '    form = ' + self.form,
            '    file = ' + self.file,
            '    func = ' + self.func,
            *('    ' + v for v in repr(self.bits).splitlines()),
            '}',
        ])

    def derive(self, name: str, desc: str, file: str) -> 'EncodingTabEntry':
        return EncodingTabEntry(
            name = name,
            desc = desc,
            form = self.form,
            file = file,
            func = self.func,
            args = self.args[:],
            bits = Instruction(self.bits),
            link = self.name,
        )

class InstructionTabEntry(NamedTuple):
    name: str
    desc: str
    data: Element
    args: list[str]
    base: Instruction

    def __repr__(self) -> str:
        return '\n'.join([
            'Instruction %s (%s) {' % (self.name, ', '.join(self.args)),
            '    ' + self.desc,
            *('    ' + v for v in repr(self.base).splitlines()),
            '}',
        ])

class DescriptionTabEntry(NamedTuple):
    brief    : str
    notes    : list[str]
    authored : list[str]

TEXT_ITEMS = {
    'arm-defined-word',
    'asm-code',
    'binarynumber',
    'hexnumber',
    'instruction',
    'para',
    'syntax',
    'value',
    'xref',
}

def parse_bit(bit: str) -> int | None:
    match bit:
        case '0': return 0
        case '1': return 1
        case 'x': return None
        case _  : raise AssertionError('invalid bit value: ' + repr(bit))

def break_line(text: str, indent: int, prefix: str = '') -> Iterator[str]:
    yield from textwrap.wrap(
        text              = text,
        width             = MAX_TEXT_WIDTH - indent,
        initial_indent    = ' ' * indent + prefix,
        subsequent_indent = ' ' * (indent + len(prefix)),
    )

def layout_line(node: Element, indent: int, prefix: str = '') -> Iterator[str]:
    text = []
    head = node.text

    if head:
        text.extend(head.split())

    for elem in node:
        if elem.tag in TEXT_ITEMS:
            if elem.text:
                text.extend(elem.text.split())

        elif elem.tag == 'sub':
            assert elem.text, 'invalid subscript'
            text.append('_')
            text.extend(elem.text.split())

        elif elem.tag == 'sup':
            assert elem.text, 'invalid superscript'
            text.append('^')
            text.extend(elem.text.split())

        elif elem.tag == 'list':
            yield from break_line(' '.join(text), indent, prefix)
            yield ''
            yield from layout_element(elem, indent)
            text.clear()

        elif elem.tag == 'image':
            yield from break_line(' '.join(text), indent, prefix)
            yield ' ' * indent + '[image:%s]' % os.path.join(os.path.dirname(DOC_FILE), elem.attrib['file'])
            yield ' ' * indent + elem.attrib['label']
            text.clear()

        else:
            raise RuntimeError('not a renderable element: ' + repr(elem.tag))

        if elem.tail:
            text.extend(elem.tail.split())

    if text:
        yield from break_line(' '.join(text), indent, prefix)

def layout_element(node: Element, indent: int = 0) -> Iterator[str]:
    match node.tag:
        case 'para':
            yield from layout_line(node, indent)
            yield ''

        case 'list':
            tail = False
            kind = node.attrib['type']

            if kind != 'unordered':
                raise RuntimeError('unsupported list type: ' + repr(kind))

            for item in node.findall('listitem'):
                content = item.find('content')
                assert content is not None, 'missing list item content'

                for line in layout_line(content, indent + 4, '* '):
                    tail = bool(line.strip())
                    yield line

            if tail:
                yield ''

        case 'note':
            tail = False
            size = len(node)

            if size:
                yield ' ' * indent + 'NOTE: '

            for elem in node:
                for line in layout_element(elem, 4):
                    tail = bool(line.strip())
                    yield line

            if tail:
                yield ''

        case 'table':
            tgroup = node.find('tgroup')
            assert tgroup is not None, 'invalid table'

            cols, rows = int(tgroup.attrib['cols']), []
            thead, tbody = tgroup.find('thead'), tgroup.find('tbody')
            assert thead is not None and tbody is not None, 'invalid table element'

            rr = thead.findall('row')
            assert len(rr) == 1, 'invalid table rows'

            hdr = rr[0].findall('entry')
            rows.append([x.text or '' for x in hdr])
            assert len(hdr) == cols, 'invalid table header'

            for row in tbody.findall('row'):
                ents = row.findall('entry')
                rows.append([x.text or '' for x in ents])
                assert len(ents) == cols, 'invalid table body'

            maxw = list(max(map(len, v)) for v in zip(*rows))
            rows = [[(c.ljust if i else c.center)(w, ' ') for c, w in zip(r, maxw)] for i, r in enumerate(rows)]

            rows.append([])
            rows.insert(0, [])
            rows.insert(len(rr) + 1, [])

            for row in rows:
                if row:
                    yield ' ' * indent + '    | %s |' % ' | '.join(row)
                else:
                    yield ' ' * indent + '    +-%s-+' % '-+-'.join('-' * w for w in maxw)
            else:
                yield ''

        case tag:
            raise RuntimeError('unrecognized tag ' + repr(tag))

def layout_content(node: Element) -> Iterator[str]:
    if node.text:
        yield from break_line(' '.join(node.text.split()), 0)

    for elem in node:
        yield from layout_element(elem)

cc = CodeGen()
cc.line('// Code generated by "mkasm_aarch64.py", DO NOT EDIT.')
cc.line()
cc.line('package aarch64')
cc.line()

iidtab = dict[str, str]()
enctab = dict[str, EncodingTabEntry]()
destab = dict[str, DescriptionTabEntry]()
instab = dict[str, InstructionTabEntry]()

isadocs = ElementTree.parse(os.path.join(os.path.dirname(__file__), DOC_FILE)).getroot()
parent_tab = { c: p for p in isadocs.iter() for c in p }

for isect in isadocs.findall('.//instructionsection'):
    iid  = isect.attrib['id']
    desc = isect.find('desc')
    assert desc is not None, 'missing description for ' + repr(iid)
    status('* Document:', iid)

    brief  = desc.find('brief')
    notes  = desc.find('encodingnotes')
    detail = desc.find('authored') or desc.find('description')
    assert brief is not None, 'missing brief for ' + repr(iid)

    destab[iid] = DescriptionTabEntry(
        brief    = ' '.join(layout_content(brief)),
        notes    = list(layout_content(notes)) if notes else [],
        authored = list(layout_content(detail)) if detail else [],
    )

    for enc in isect.findall('.//encoding'):
        key = enc.attrib['name']
        val = iidtab.get(key)

        if val and val != iid:
            raise RuntimeError('encoding IID confliction: %s=%s:%s' % (key, val, iid))
        else:
            iidtab[key] = iid

for iclass in sorted(isadocs.findall('.//encodingindex/iclass_sect'), key = lambda x: x.attrib['id']):
    name = iclass.attrib['id']
    desc = iclass.attrib['title']
    itab = iclass.find('instructiontable')

    # TODO: support SVE and SME instructions sometime in the future
    if name.startswith('sve') or name.startswith('mortlach'):
        continue

    assert itab is not None, 'missing instruction table for ' + name
    assert itab.attrib['iclass'] == name, 'wrong instruction table iclass for ' + name
    status('* Instruction Class:', name)

    bits = Instruction()
    bits.update(iclass.findall('regdiagram/box'))

    cond = set()
    vals = bits.bits[:]
    argv = sorted(bits.refs.items(), key = lambda x: x[1], reverse = True)
    args = [v for v, _ in argv]

    for dc in iclass.findall('decode_constraints/decode_constraint'):
        cond.add((dc.attrib['name'], dc.attrib['op'], dc.attrib['val']))

    for p, n in bits.refs.values():
        vals[p:p + n] = [0] * n

    instab[name] = InstructionTabEntry(
        name = name,
        desc = desc,
        data = itab,
        args = args,
        base = bits,
    )

    cc.line('// %s: %s' % (name, desc))
    cc.line('func %s(%s uint32) uint32 {' % (name, ', '.join(args)))
    cc.indent()

    for th, (_, n) in argv:
        cc.line('if %s &^ %s != 0 {' % (th, hex((1 << n) - 1)))
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

    assert None not in vals, 'unset bit in %s: %r' % (name, vals)
    cc.line('ret := uint32(0x%08x)' % int(''.join(map(str, vals))[::-1], 2))

    for th, (p, _) in argv:
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
                args = entry.args,
                bits = instr,
            )

            assert encname not in enctab, 'duplicated encoding name %s.%s' % (name, encname)
            status('* Encoding Table Entry: %s.%s' % (name, encname))
            enctab[encname] = enc

for node in sorted(isadocs.findall('.//instructionsection[@type="alias"]'), key = lambda x: x.attrib['id']):
    to = node.find('aliasto')
    hdr = node.find('heading')

    assert to is not None, 'invalid alias section'
    assert hdr is not None and hdr.text, 'invalid alias section'

    iclass = ''
    refiform = to.attrib['refiform']
    iformfile = parent_tab[node].attrib['file']

    for vv in node.findall('docvars/docvar'):
        if vv.attrib['key'] == 'instr-class':
            iclass = vv.attrib['value']
            break

    # TODO: support SVE and SME instructions sometime in the future
    if iclass.startswith('sve') or iclass.startswith('mortlach'):
        continue

    for cls in node.findall('classes/iclass'):
        encs = cls.findall('encoding')
        bits = cls.findall('regdiagram/box')

        for enc in encs:
            parent = None
            encname = enc.attrib['name']
            asmtmpl = enc.findall('equivalent_to/asmtemplate/a')

            for equ in asmtmpl:
                href = equ.attrib.get('href', '')
                vals = href.split('#')

                if len(vals) == 2 and vals[0] == refiform:
                    parent = enctab.get(vals[1])
                    break

            if parent is not None:
                form = parent.derive(encname, hdr.text, iformfile)
                form.bits.update(bits)
                enctab[encname] = form

### ---------- Instruction Assembly Template ---------- ###

class Mop(StrEnum):
    LSL  = 'lsl'
    LSR  = 'lsr'
    ASR  = 'asr'
    ROR  = 'ror'
    MSL  = 'msl'
    UXTB = 'uxtb'
    UXTH = 'uxth'
    UXTW = 'uxtw'
    UXTX = 'uxtx'
    SXTB = 'sxtb'
    SXTH = 'sxth'
    SXTW = 'sxtw'
    SXTX = 'sxtx'

class Sym(StrEnum):
    CSYNC        = 'CSYNC'
    DSYNC        = 'DSYNC'
    PRFOP        = 'sa_prfop'
    RCTX         = 'RCTX'
    RPRFOP       = 'sa_rprfop'
    OPTION       = 'sa_option'
    OPTION_NXS   = 'sa_option_1'
    AT_OPTION    = 'sa_at_op'
    BRB_OPTION   = 'sa_brb_op'
    DC_OPTION    = 'sa_dc_op'
    IC_OPTION    = 'sa_ic_op'
    SME_OPTION   = 'sa_sme_option'
    TLBI_OPTION  = 'sa_tlbi_op'
    TLBIP_OPTION = 'sa_tlbip_op'
    PSTATE_FIELD = 'sa_pstatefield'
    SYSREG       = 'sa_systemreg'
    TARGETS      = 'sa_targets'

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
    incr: bool = False
    altr: str | None = None
    size: str | None = None
    mode: str | Tag | None = None
    vidx: Imm | Lit | None = None

    def __str__(self) -> str:
        if self.size is not None:
            assert not self.incr
            assert self.altr is None
            assert self.mode is None
            assert self.vidx is None
            return '%s[%s]' % (self.size, self.name)
        elif self.mode is not None and self.vidx is not None:
            assert not self.incr
            assert self.altr is None
            return '%s.%s[%s]' % (self.name, self.mode, self.vidx)
        elif self.mode is not None:
            assert not self.incr
            assert self.altr is None
            return '%s.%s' % (self.name, self.mode)
        elif self.altr is not None:
            assert not self.incr
            assert self.vidx is None
            return '(%s|%s)' % (self.name, self.altr)
        elif self.incr:
            return self.name + '!'
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
    mod: str | Mop
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
    opt: list[Reg | Vec | Mem | Mod | Imm | Lit | Sym]

    def __str__(self) -> str:
        ret = []
        pfx = '{'

        if self.req:
            pfx = '{, '
            ret.extend('%s%s' % (', ' if i else '', str(v)) for i, v in enumerate(self.req))

        if self.opt:
            ret.append(pfx)
            ret.extend('%s%s' % (', ' if i else '', str(v)) for i, v in enumerate(self.opt))
            ret.append('}')

        if not ret:
            return ''
        else:
            return ''.join(ret)

class Instr(NamedTuple):
    mnemonic: str
    operands: Seq
    modifier: str | None

    def __str__(self) -> str:
        return ''.join([
            self.mnemonic,
            '{%s}' % self.modifier if self.modifier else '',
            ' ' if self.operands.req or self.operands.opt else '',
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

    modops = {
        v.name
        for v in Mop
    }

    symtab = {
        Sym.CSYNC: (
            'CSYNC option',
            ['CSYNC'],
        ),
        Sym.DSYNC: (
            'DSYNC option',
            ['DSYNC'],
        ),
        Sym.RCTX: (
            'RCTX option',
            ['RCTX'],
        ),
        Sym.PRFOP: (
            'prefetch option',
            ['sa_prfop', '|', '#', 'sa_imm5']
        ),
        Sym.RPRFOP: (
            'range prefetch option',
            ['sa_rprfop', '|', '#', 'sa_imm6'],
        ),
        Sym.OPTION: (
            'barrier option',
            ['sa_option', '|', '#', 'sa_imm'],
        ),
        Sym.OPTION_NXS: (
            'barrier option nXS',
            ['sa_option_1', 'nXS'],
        ),
        Sym.AT_OPTION: (
            'address translation options',
            ['sa_at_op'],
        ),
        Sym.BRB_OPTION: (
            'branch record buffer options',
            ['sa_brb_op'],
        ),
        Sym.DC_OPTION: (
            'data cache options',
            ['sa_dc_op'],
        ),
        Sym.IC_OPTION: (
            'instruction cache options',
            ['sa_ic_op'],
        ),
        Sym.SME_OPTION: (
            'streaming SVE and SME options',
            ['sa_option'],
        ),
        Sym.TLBI_OPTION: (
            'TLB invalidate options',
            ['sa_tlbi_op'],
        ),
        Sym.TLBIP_OPTION: (
            'TLB invalidate pair options',
            ['sa_tlbip_op'],
        ),
        Sym.PSTATE_FIELD: (
            'MSR PState fields',
            ['sa_pstatefield'],
        ),
        Sym.SYSREG: (
            'system register',
            ['sa_systemreg', '|', 'S', 'sa_op0', '_', 'sa_op1', '_', 'sa_cn', '_', 'sa_cm', '_', 'sa_op2'],
        ),
        Sym.TARGETS: (
            'branch targets option',
            ['sa_targets'],
        ),
    }

    for v in Sym:
        assert v in symtab, 'undefined symbol: ' + str(v)

    amounts = {
        'sa_amount',
        'sa_amount_1',
        'sa_amount_2',
        'sa_amount_3',
        'sa_amount_4',
        'sa_shift',
        'sa_shift_1',
        'sa_shift_2',
        'sa_shift_3',
    }

    modifiers = {
        'sa_shift',
        'sa_extend',
        'sa_extend_1',
    }

    registers = {
        'sa_d',
        'sa_m',
        'sa_n',
        'sa_t',
    }

    immediates = {
        'sa_imm',
        'sa_imm_1',
        'sa_imm_2',
        'sa_imm_3',
    }

    predefined = {
        tok[0]
        for _, tok in symtab.values()
    }

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

    @staticmethod
    def _isint(v: str) -> bool:
        try:
            int(v)
        except ValueError:
            return False
        else:
            return True

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
        pos = self.pos

        for sym, (msg, mat) in self.symtab.items():
            if name == mat[0]:
                tok.clear()
                self.pos = pos

                for _ in mat[1:]:
                    if self.eof:
                        break
                    else:
                        v = self.next()
                        tok.append(v.name if isinstance(v, Token) else v)

                else:
                    if tok != mat[1:]:
                        raise SyntaxError('invalid %s: %r' % (msg, name))
                    else:
                        return sym

        else:
            raise RuntimeError('no matching symbol for ' + repr(name))

    def value(self) -> Reg | Vec | Mem | Mod | Imm | Lit | Sym:
        match self.next():
            case '#':
                match self.next():
                    case '0.0'                                      : return Lit(0.0)
                    case v if isinstance(v, str) and self._isint(v) : return Lit(int(v))
                    case v if isinstance(v, Token)                  : return Imm(v.name)
                    case v                                          : raise SyntaxError('integer or token expected, got ' + repr(v))

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

                if buf.opt:
                    if len(buf.opt) == 1:
                        args.append((buf.opt[0], True))
                    else:
                        raise RuntimeError('multiple optional value for memory operand')

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

            case v if isinstance(v, Token) and v.name not in self.modifiers:
                mode = None
                size = None
                vidx = None
                incr = False
                name = v.name

                if not self.eof:
                    if self.skip('!'):
                        incr = True

                    elif self.skip('.'):
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
                    incr = incr,
                    mode = mode,
                    size = size,
                    vidx = vidx,
                )

            case v:
                opt = False
                mod = v.name if isinstance(v, Token) else v

                if isinstance(v, str):
                    if mod not in self.modops:
                        raise SyntaxError('unexpected token: ' + repr(v))
                else:
                    if mod not in self.modifiers:
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

                if mod in self.modops:
                    mod = Mop(mod.lower())

                if imm is None:
                    return Mod(mod)
                else:
                    return Mod(mod, (imm, opt))

    def vlist(self) -> Seq:
        req = []
        opt = []

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

            val = self.value()
            opt.append(val)

            while self.skip(','):
                opt.append(self.value())
            else:
                self.must('}')

        return Seq(
            req = req,
            opt = opt,
        )

    def instr(self) -> Instr:
        mods = None
        args = Seq([], [])
        name = self.next()

        if not isinstance(name, str):
            raise SyntaxError('mnemonic expected')

        if not self.eof:
            match self.tok:
                case '.' if name in {'B', 'BC'}:
                    self.next()
                    cond = self.next()

                    if not isinstance(cond, Token):
                        raise SyntaxError('branch condition expected')
                    else:
                        mods = cond.name

                case v if isinstance(v, Token) and (v.name == 'sa_2' or v.name == 'sa_bt' and name == 'BFMLAL'):
                    mods = v.name
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

                if ch == '-' and i < len(self.buf) - 1 and self.buf[i + 1].isdigit():
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
    feat   : list[str]
    bits   : Instruction
    opts   : dict[str, str]
    args   : dict[str, int]
    enctab : EncodingTabEntry
    fields : dict[str, Account | Definition]

def parse_props(out: dict[str, str], p: Element):
    for v in p.findall('docvars/docvar'):
        out[v.attrib['key']] = v.attrib['value']

def parse_symdef(defs: Element) -> tuple[list[str], dict[str, set[str]]]:
    rets = {}
    tabs = defs.findall('.//tgroup')
    rows = defs.findall('.//tbody/row')
    refs = defs.findall('.//thead/row/entry[@class="bitfield"]')
    dest = defs.attrib['encodedin']

    assert refs, 'no bitfield header name found: encodedin="%s"' % dest
    assert len(tabs) == 1, 'expect exactly 1 table: encodedin="%s"' % dest

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
            rets.setdefault(sym, set()).add(''.join(bits))

    refs = [v.text or '' for v in refs]
    refs.reverse()

    if not rets:
        raise AssertionError('missing definitions: encodedin="%s"' % dest)
    else:
        return refs, rets

MISSING_ENCODING_IN = {
    'DSB_BOn_barriers': {
        'sa_option_1': 'imm2',
    }
}

formtab = dict[str, list[InstrForm]]()
fieldtab = dict[str, dict[str, Account | Definition]]()

for expl in isadocs.findall('.//explanation'):
    symbol = expl.find('symbol')
    symacc = expl.find('account')
    symdef = expl.find('definition')

    if symbol is None or (symacc is None) is (symdef is None):
        raise AssertionError('invalid explanation')

    defv = ''
    desc = symbol.text or ''
    name = symbol.attrib['link']
    status('* Field Explanation:', name)

    if symacc is not None:
        assert symdef is None
        dest, refs, bits, text = symacc.attrib['encodedin'], None, None, ' '.join(symacc.itertext())
    else:
        assert isinstance(symdef, Element)
        dest, (refs, bits), text = symdef.attrib['encodedin'], parse_symdef(symdef), ' '.join(symdef.itertext())

    if mt := re.search('[Dd]efaults to #?([^., ]+)', text):
        defv = mt.group(1)
    elif mt := re.search('[Dd]efaulting to ([^, ]+)', text):
        defv = mt.group(1)

    for enc in map(str.strip, expl.attrib['enclist'].split(',')):
        tab = MISSING_ENCODING_IN.get(enc, {})
        sym = tab.get(name)

        if sym is not None:
            assert not dest, 'replace of non-empty field'
            dest = sym

        if refs is None or bits is None:
            defs = Account(dest, desc, defv)
        else:
            defs = Definition(dest, desc, defv, refs, bits)

        tab = fieldtab.setdefault(enc, {})
        tab[name] = defs

for encdata in sorted(enctab.values(), key = lambda x: x.name):
    node = isadocs.find('.//iclass/encoding[@name="%s"]' % encdata.name)
    assert node is not None, 'encoding %s does not exists' % repr(encdata.name)

    tokens = node.findall('asmtemplate/*')
    assert tokens, 'encoding %s does not have an assembly syntax' % repr(encdata.name)

    text = ''.join(v.text or '' for v in tokens)
    status('* Assembly Template:', text)

    args = {}
    opts = {}
    feat = set()
    bits = Instruction(encdata.bits)
    inst = AsmTemplate.parse(tokens)

    for var in parent_tab[node].findall('arch_variants/arch_variant'):
        for ft in var.attrib['feature'].split('&&'):
            feat.add(ft.strip())

    parse_props(opts, node)
    bits.update(node.findall('box'))
    bits.update(parent_tab[node].findall('regdiagram/box'))

    fields   = fieldtab.get(encdata.name, {})
    mnemonic = opts['mnemonic']

    if encdata.link is not None:
        fields   = dict(fieldtab[encdata.link], **fields)
        mnemonic = opts['alias_mnemonic']

    if mnemonic != inst.mnemonic:
        raise RuntimeError('mnemonic mismatch: %s != %s' % (mnemonic, inst.mnemonic))

    req = list(bits.refs.items())
    req.sort(key = lambda x: x[1], reverse = True)

    for fn in encdata.args:
        p, n = bits.refs[fn]
        fval = bits.bits[p:p + n][::-1]

        if None not in fval:
            args[fn] = int(''.join(map(str, fval)), 2)

    formtab.setdefault(inst.mnemonic, []).append(InstrForm(
        text   = text,
        inst   = inst,
        feat   = sorted(feat),
        bits   = bits,
        opts   = opts,
        args   = args,
        enctab = encdata,
        fields = fields,
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
            return ' && '.join('(%s)' % v if isinstance(v, Or) and len(v) > 1 else str(v) for v in self)

class OnceDict(OrderedDict):
    def __setitem__(self, k, v) -> None:
        if k not in self:
            super().__setitem__(k, v)

SYM_CHECKS = {
    Sym.CSYNC        : '%s == CSYNC',
    Sym.DSYNC        : '%s == DSYNC',
    Sym.PRFOP        : 'isBasicPrf(%s)',
    Sym.RCTX         : '%s == RCTX',
    Sym.RPRFOP       : 'isRangePrf(%s)',
    Sym.OPTION       : 'isOption(%s)',
    Sym.OPTION_NXS   : 'isOptionNXS(%s)',
    Sym.AT_OPTION    : 'isATOption(%s)',
    Sym.BRB_OPTION   : 'isBRBOption(%s)',
    Sym.DC_OPTION    : 'isDCOption(%s)',
    Sym.IC_OPTION    : 'isICOption(%s)',
    Sym.SME_OPTION   : 'isSMEOption(%s)',
    Sym.TLBI_OPTION  : 'isTLBIOption(%s)',
    Sym.TLBIP_OPTION : 'isTLBIPOption(%s)',
    Sym.PSTATE_FIELD : 'isPState(%s)',
    Sym.SYSREG       : 'isSysReg(%s)',
    Sym.TARGETS      : 'isTargets(%s)',
}

IMM_CHECKS = {
    'CRm'                             : 'isUimm4(%s)',
    'CRm:Encoding:Hints:Index:by:op2' : 'isUimm7(%s)',
    'N:immr:imms'                     : 'isMask64(%s)',
    'Q:imm4'                          : 'isExtIndex(v0, %s)',
    'a:b:c:d:e:f:g:h'                 : 'isUimm8(%s)',
    'b40:b5'                          : 'isUimm6(%s)',
    'imm5'                            : 'isUimm5(%s)',
    'imm6'                            : 'isUimm6(%s)',
    'imm8'                            : 'isFpImm8(%s)',
    'imm12'                           : 'isImm12(%s)',
    'imm16'                           : 'isUimm16(%s)',
    'immh:immb'                       : 'isFpBits(%s)',
    'immr'                            : 'isUimm6(%s)',
    'immr:imms'                       : 'isMask32(%s)',
    'imms'                            : 'isUimm6(%s)',
    'mask'                            : 'isUimm4(%s)',
    'nzcv'                            : 'isUimm4(%s)',
    'op1'                             : 'isUimm3(%s)',
    'op2'                             : 'isUimm3(%s)',
    'scale'                           : 'isFpBits(%s)',
    'uimm4'                           : 'isUimm4(%s)',
    'uimm6'                           : 'isUimm6(%s)',
}

IMM_SPECIAL_CHECKS = {
    'MOV_MOVN_32_movewide': { 'hw:imm16': 'isMOVxImm(%s, 32, true)' },
    'MOV_MOVN_64_movewide': { 'hw:imm16': 'isMOVxImm(%s, 64, true)' },
    'MOV_MOVZ_32_movewide': { 'hw:imm16': 'isMOVxImm(%s, 32, false)' },
    'MOV_MOVZ_64_movewide': { 'hw:imm16': 'isMOVxImm(%s, 64, false)' },
}

REG_CHECKS = {
    'sa_cm'          : 'isUimm4(%s)',
    'sa_cn'          : 'isUimm4(%s)',
    'sa_cond'        : 'isBrCond(%s)',
    'sa_bt'          : 'isBr(%s)',
    'sa_da'          : 'isDr(%s)',
    'sa_dd'          : 'isDr(%s)',
    'sa_dm'          : 'isDr(%s)',
    'sa_dn'          : 'isDr(%s)',
    'sa_dn_1'        : 'isDr(%s)',
    'sa_dt'          : 'isDr(%s)',
    'sa_dt1'         : 'isDr(%s)',
    'sa_dt2'         : 'isDr(%s)',
    'sa_ha'          : 'isHr(%s)',
    'sa_hd'          : 'isHr(%s)',
    'sa_hm'          : 'isHr(%s)',
    'sa_hn'          : 'isHr(%s)',
    'sa_hn_1'        : 'isHr(%s)',
    'sa_ht'          : 'isHr(%s)',
    'sa_label'       : 'isLabel(%s)',
    'sa_pstatefield' : 'isPState(%s)',
    'sa_qd'          : 'isQr(%s)',
    'sa_qn'          : 'isQr(%s)',
    'sa_qt'          : 'isQr(%s)',
    'sa_qt1'         : 'isQr(%s)',
    'sa_qt2'         : 'isQr(%s)',
    'sa_sa'          : 'isSr(%s)',
    'sa_sd'          : 'isSr(%s)',
    'sa_sm'          : 'isSr(%s)',
    'sa_sn'          : 'isSr(%s)',
    'sa_sn_1'        : 'isSr(%s)',
    'sa_st'          : 'isSr(%s)',
    'sa_st1'         : 'isSr(%s)',
    'sa_st2'         : 'isSr(%s)',
    'sa_vd'          : 'isVr(%s)',
    'sa_vm'          : 'isVr(%s)',
    'sa_vn'          : 'isVr(%s)',
    'sa_wa'          : 'isWr(%s)',
    'sa_wd'          : 'isWr(%s)',
    'sa_wd_wsp'      : 'isWrOrWSP(%s)',
    'sa_wm'          : 'isWr(%s)',
    'sa_wm_1'        : 'isWr(%s)',
    'sa_wn'          : 'isWr(%s)',
    'sa_wn_1'        : 'isWr(%s)',
    'sa_wn_wsp'      : 'isWrOrWSP(%s)',
    'sa_ws'          : 'isWr(%s)',
    'sa_wt'          : 'isWr(%s)',
    'sa_wt1'         : 'isWr(%s)',
    'sa_wt2'         : 'isWr(%s)',
    'sa_xa'          : 'isXr(%s)',
    'sa_xd'          : 'isXr(%s)',
    'sa_xd_1'        : 'isXr(%s)',
    'sa_xd_sp'       : 'isXrOrSP(%s)',
    'sa_xm'          : 'isXr(%s)',
    'sa_xm_1'        : 'isXr(%s)',
    'sa_xm_sp'       : 'isXrOrSP(%s)',
    'sa_xn'          : 'isXr(%s)',
    'sa_xn_1'        : 'isXr(%s)',
    'sa_xn_2'        : 'isXr(%s)',
    'sa_xn_sp'       : 'isXrOrSP(%s)',
    'sa_xs'          : 'isXr(%s)',
    'sa_xs_1'        : 'isXr(%s)',
    'sa_xt'          : 'isXr(%s)',
    'sa_xt_1'        : 'isXr(%s)',
    'sa_xt1'         : 'isXr(%s)',
    'sa_xt2'         : 'isXr(%s)',
    'sa_xt_sp'       : 'isXrOrSP(%s)',
}

REG_SPECIAL_CHECKS = {
    'CINC_CSINC_32_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
    'CINC_CSINC_64_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
    'CSET_CSINC_32_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
    'CSET_CSINC_64_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
    'CINV_CSINV_32_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
    'CINV_CSINV_64_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
    'CSETM_CSINV_32_condsel' : { 'sa_cond_1': 'isBrCond(%s)' },
    'CSETM_CSINV_64_condsel' : { 'sa_cond_1': 'isBrCond(%s)' },
    'CNEG_CSNEG_32_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
    'CNEG_CSNEG_64_condsel'  : { 'sa_cond_1': 'isBrCond(%s)' },
}

REG_PLUS_NAMES = {
    'CASP_CP32_comswappr'         : { 'sa_w_s': 'sa_ws', 'sa_w_t': 'sa_wt' },
    'CASP_CP64_comswappr'         : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'CASPA_CP32_comswappr'        : { 'sa_w_s': 'sa_ws', 'sa_w_t': 'sa_wt' },
    'CASPA_CP64_comswappr'        : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'CASPAL_CP32_comswappr'       : { 'sa_w_s': 'sa_ws', 'sa_w_t': 'sa_wt' },
    'CASPAL_CP64_comswappr'       : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'CASPL_CP32_comswappr'        : { 'sa_w_s': 'sa_ws', 'sa_w_t': 'sa_wt' },
    'CASPL_CP64_comswappr'        : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWCASP_C64_rcwcomswappr'    : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWCASPA_C64_rcwcomswappr'   : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWCASPAL_C64_rcwcomswappr'  : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWCASPL_C64_rcwcomswappr'   : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWSCASP_C64_rcwcomswappr'   : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWSCASPA_C64_rcwcomswappr'  : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWSCASPAL_C64_rcwcomswappr' : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
    'RCWSCASPL_C64_rcwcomswappr'  : { 'sa_x_s': 'sa_xs', 'sa_x_t': 'sa_xt' },
}

REG_CHECKS_MERGED = {
    'isWr(%s) || isXr(%s)'  : 'isWrOrXr(%s)',
    'isXr(%s) || isSP(%s)'  : 'isXrOrSP(%s)',
    'isWr(%s) || isWSP(%s)' : 'isWrOrWSP(%s)',
}

COMBREG_CHECKS = {
    'sa_d': 'isWrOrXr(%s)',
    'sa_m': 'isWrOrXr(%s)',
    'sa_n': 'isWrOrXr(%s)',
    'sa_t': 'isWrOrXr(%s)',
}

FIXEDVEC_CHECKS = {
    'sa_d': 'isAdvSIMD(%s)',
    'sa_m': 'isAdvSIMD(%s)',
    'sa_n': 'isAdvSIMD(%s)',
}

XBFM_CHECKS = {
    'BFC_BFM_32M_bitfield': {
        'sa_lsb'   : 'isUimm5(%s)',
        'sa_width' : 'isBFxWidth(v1, %s, 32)',
    },
    'BFC_BFM_64M_bitfield': {
        'sa_lsb_2'   : 'isUimm6(%s)',
        'sa_width_1' : 'isBFxWidth(v1, %s, 64)',
    },
    'BFI_BFM_32M_bitfield': {
        'sa_lsb'   : 'isUimm5(%s)',
        'sa_width' : 'isBFxWidth(v2, %s, 32)',
    },
    'BFI_BFM_64M_bitfield': {
        'sa_lsb_2'   : 'isUimm6(%s)',
        'sa_width_1' : 'isBFxWidth(v2, %s, 64)',
    },
    'BFXIL_BFM_32M_bitfield': {
        'sa_lsb_1' : 'isUimm5(%s)',
        'sa_width' : 'isBFxWidth(v2, %s, 32)',
    },
    'BFXIL_BFM_64M_bitfield': {
        'sa_lsb_3'   : 'isUimm6(%s)',
        'sa_width_1' : 'isBFxWidth(v2, %s, 64)',
    },
    'LSL_UBFM_32M_bitfield': {
        'sa_shift_1': 'isUimm5(%s)',
    },
    'LSL_UBFM_64M_bitfield': {
        'sa_shift_3': 'isUimm6(%s)',
    },
    'SBFIZ_SBFM_32M_bitfield': {
        'sa_lsb'   : 'isUimm5(%s)',
        'sa_width' : 'isBFxWidth(v2, %s, 32)',
    },
    'SBFIZ_SBFM_64M_bitfield': {
        'sa_lsb_2'   : 'isUimm6(%s)',
        'sa_width_1' : 'isBFxWidth(v2, %s, 64)',
    },
    'SBFX_SBFM_32M_bitfield': {
        'sa_lsb_1' : 'isUimm5(%s)',
        'sa_width' : 'isBFxWidth(v2, %s, 32)',
    },
    'SBFX_SBFM_64M_bitfield': {
        'sa_lsb_3'   : 'isUimm6(%s)',
        'sa_width_1' : 'isBFxWidth(v2, %s, 64)',
    },
    'UBFIZ_UBFM_32M_bitfield': {
        'sa_lsb'   : 'isUimm5(%s)',
        'sa_width' : 'isBFxWidth(v2, %s, 32)',
    },
    'UBFIZ_UBFM_64M_bitfield': {
        'sa_lsb_2'   : 'isUimm6(%s)',
        'sa_width_1' : 'isBFxWidth(v2, %s, 64)',
    },
    'UBFX_UBFM_32M_bitfield': {
        'sa_lsb_1' : 'isUimm5(%s)',
        'sa_width' : 'isBFxWidth(v2, %s, 32)',
    },
    'UBFX_UBFM_64M_bitfield': {
        'sa_lsb_3'   : 'isUimm6(%s)',
        'sa_width_1' : 'isBFxWidth(v2, %s, 64)',
    },
}

SIGNED_IMM = {
    'sa_imm',
    'sa_imm_1',
    'sa_imm_2',
    'sa_imm_3',
    'sa_imm_4',
    'sa_imm_5',
    'sa_simm',
    'sa_pimm',
    'sa_pimm_1',
    'sa_pimm_2',
    'sa_pimm_3',
    'sa_pimm_4',
}

UNSIGNED_IMM = {
    'sa_uimm',
}

SCALAR_TYPES = {
    'B': 'BRegister',
    'H': 'HRegister',
    'S': 'SRegister',
    'D': 'DRegister',
    'Q': 'QRegister',
}

VECTOR_TYPES = {
    '8B'  : 'Vec8B',
    '16B' : 'Vec16B',
    '2H'  : 'Vec2H',
    '4H'  : 'Vec4H',
    '8H'  : 'Vec8H',
    '2S'  : 'Vec2S',
    '4S'  : 'Vec4S',
    '1D'  : 'Vec1D',
    '2D'  : 'Vec2D',
    '1Q'  : 'Vec1Q',
}

VECTOR_MODES = {
    'B': 'ModeB',
    'H': 'ModeH',
    'S': 'ModeS',
    'D': 'ModeD',
}

VECTOR_SIZE_CHECKS = {
    'FMAXNMV_asimdall_only_H': {
        'sa_v': 'isHr(%s)',
    },
    'FMINNMV_asimdall_only_H': {
        'sa_v': 'isHr(%s)',
    },
    'FMAXV_asimdall_only_H': {
        'sa_v': 'isHr(%s)',
    },
    'FMINV_asimdall_only_H': {
        'sa_v': 'isHr(%s)',
    },
}

MODIFIER_TYPES = {
    'MSL'  : 'ModMSL',
    'LSL'  : 'ModLSL',
    'LSR'  : 'ModLSR',
    'ASR'  : 'ModASR',
    'ROR'  : 'ModROR',
    'UXTB' : 'ModUXTB',
    'UXTH' : 'ModUXTH',
    'UXTW' : 'ModUXTW',
    'UXTX' : 'ModUXTX',
    'SXTB' : 'ModSXTB',
    'SXTH' : 'ModSXTH',
    'SXTW' : 'ModSXTW',
    'SXTX' : 'ModSXTX',
}

def match_modifier(form: InstrForm, name: str, mod: Mod, optional: bool, *extra_cond: str) -> Iterator[Or | str]:
    if not optional:
        if isinstance(mod.mod, Mop):
            yield 'modt(%s) == %s' % (name, MODIFIER_TYPES[mod.mod.name])
        elif mod.mod in AsmTemplate.modifiers:
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
        mods = set()
        fixed = set()

        if isinstance(mod.mod, Mop):
            base = 'modt(%s) == %s' % (name, MODIFIER_TYPES[mod.mod.name])

        elif mod.mod in AsmTemplate.modifiers:
            field = form.fields[mod.mod]
            assert isinstance(field, Definition), 'invalid modifier field matching'

            for v in field.bits:
                if v != 'RESERVED' and not v.startswith('SEE'):
                    if v in MODIFIER_TYPES:
                        mods.add(MODIFIER_TYPES[v])
                        continue

                    vals = v.split()
                    kind, amount = vals

                    mods.add(MODIFIER_TYPES[kind])
                    fixed.add(int(amount[1:]))

            if not mods:
                raise RuntimeError('empty modifier flags')
            else:
                base = 'isMods(%s, %s)' % (name, ', '.join(sorted(mods)))

        else:
            raise RuntimeError('invalid extension')

        if mod.imm is None:
            if not fixed:
                yield Or(*extra_cond, And(base, 'modn(%s) == 0' % name))
            else:
                yield Or(*extra_cond, And(base, 'isIntLit(modn(%s), %s)' % (name, ', '.join(map(str, sorted(fixed))))))

        else:
            imm = mod.imm[0]
            opm = mod.imm[1]

            if fixed:
                raise RuntimeError('fixed modifier with imm is not allowed')
            elif not isinstance(imm, Lit):
                yield Or(*extra_cond, base)
            elif imm.ty is not int:
                raise RuntimeError('invalid literal type for amounts')
            elif not opm or imm.val == 0:
                yield Or(*extra_cond, And(base, 'modn(%s) == %d' % (name, imm.val)))
            else:
                yield Or(*extra_cond, And(base, Or('modn(%s) == 0' % name, 'modn(%s) == %d' % (name, imm.val))))

def match_operands(form: InstrForm, argc: int) -> Iterator['And | Or | str']:
    argv = form.inst.operands.req[:]
    regtab = {}
    dynvec = {}
    fixedvec = {}

    if form.inst.operands.opt:
        argv.extend(form.inst.operands.opt)

    if len(argv) > argc:
        narg = len(argv) - argc
        nopt = len(form.inst.operands.opt)

        if not nopt:
            yield 'len(vv) == %d' % narg
        else:
            yield Or('len(vv) == %d' % (narg - nopt), 'len(vv) == %d' % narg)

    for i, val in enumerate(argv):
        name = 'v%d' % i if i < argc else 'vv[%d]' % (i - argc)
        optcond = []

        if i >= len(form.inst.operands.req):
            optcond.append('len(vv) == 0')

        if isinstance(val, Reg):
            if optcond:
                if val.altr is not None or \
                   val.size is not None or \
                   val.mode is not None or \
                   val.vidx is not None:
                    raise RuntimeError('optional complex reg operand is not supported')
                else:
                    yield Or(*optcond, REG_CHECKS[val.name] % name)

            elif val.size is not None:
                if val.size == 'sa_r':
                    yield COMBREG_CHECKS[val.name] % name
                elif val.size not in VECTOR_SIZE_CHECKS.get(form.enctab.name, {}):
                    yield FIXEDVEC_CHECKS[val.name] % name
                    fixedvec.setdefault(val.size, []).append(name)
                else:
                    yield VECTOR_SIZE_CHECKS[form.enctab.name][val.size] % name
                    fixedvec.setdefault(val.size, []).append(name)

            elif val.mode is not None and val.vidx is not None:
                if not val.name.startswith('sa_v'):
                    raise RuntimeError('invalid indexing on non-vector registers: ' + str(val))
                else:
                    yield 'isVri(%s)' % name

                if isinstance(val.mode, Tag):
                    yield 'vmoder(%s) == Mode%s' % (name, val.mode.name)

                if isinstance(val.vidx, Lit):
                    if val.vidx.ty is not int:
                        raise RuntimeError('non-integer indexing on vector register: ' + str(val))
                    else:
                        yield 'vidxr(%s) == %d' % (name, val.vidx.val)

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

                        sels = [
                            VECTOR_TYPES[x]
                            for x in field.bits
                            if x != 'RESERVED' and not x.startswith('SEE')
                        ]

                        match len(sels):
                            case 0: pass
                            case 1: yield 'vfmt(%s) == %s' % (name, sels[0])
                            case _: yield 'isVfmt(%s, %s)' % (name, ', '.join(sels))

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

            elif '_plus_' in val.name:
                vals = val.name.split('_plus_')
                real, incr = vals[0], int(vals[1])

                if len(vals) != 2:
                    raise RuntimeError('invalid register name: ' + str(val))

                rtab = REG_PLUS_NAMES.get(form.enctab.name, {})
                real = rtab.get(real, real)

                yield REG_CHECKS[real] % name
                yield 'isNextReg(%s, %s, %d)' % (name, regtab[real], incr)

            else:
                rc = REG_SPECIAL_CHECKS.get(form.enctab.name, {}).get(val.name)
                regtab[val.name] = name

                if rc is None:
                    rc = REG_CHECKS.get(val.name)

                if rc is not None:
                    yield rc % name
                else:
                    raise RuntimeError("cannot match register '%s.%s'" % (form.enctab.name, val.name))

        elif isinstance(val, Vec):
            if optcond:
                raise RuntimeError('optional vector operand is not supported')

            pfx = 'Vec'
            elm = 'velm'
            vec = 'isVec'

            if val.vidx is not None:
                pfx = 'Mode'
                elm = 'vmodei'
                vec = 'isIdxVec'

            nelm = len(val.regs)
            yield '%s%d(%s)' % (vec, nelm, name)

            if isinstance(val.mode, Tag):
                yield '%s(%s) == %s%s' % (elm, name, pfx, val.mode.name)

            if isinstance(val.vidx, Lit):
                raise NotImplementedError('match_operands: literal indexed vec')

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
                    yield from match_modifier(form, 'mext(%s)' % name, val.extend[0], val.extend[1], 'mext(%s) == nil' % name)

        elif isinstance(val, Mod):
            yield from match_modifier(form, name, val, bool(optcond), *optcond)

        elif isinstance(val, Imm):
            fv = form.fields[val.name]
            bm = XBFM_CHECKS.get(form.enctab.name, {}).get(val.name)
            fc = IMM_SPECIAL_CHECKS.get(form.enctab.name, {}).get(fv.name)

            if fc is None:
                fc = IMM_CHECKS.get(fv.name)

            if fc is not None:
                yield Or(*optcond, fc % name)
            elif isinstance(fv, Definition):
                bits = [x for x in sorted(fv.bits) if x != 'RESERVED' and not x.startswith('SEE')]
                yield Or(*optcond, 'isIntLit(%s, %s)' % (name, ', '.join(map(str, sorted(map(int, bits))))))
            elif not fv.name and bm is not None:
                yield Or(*optcond, bm % name)
            else:
                raise RuntimeError("cannot match immediate field '%s.%s (%s)'" % (form.enctab.name, val.name, fv.name))

        elif isinstance(val, Lit):
            if optcond:
                raise RuntimeError('optional lit operand is not supported')
            elif val.ty is int:
                yield 'isIntLit(%s, %s)' % (name, val.val)
            elif val.ty is float:
                yield 'isFloatLit(%s, %s)' % (name, val.val)
            else:
                raise RuntimeError('invalid literal type: ' + repr(val.ty))

        elif isinstance(val, Sym):
            yield Or(*optcond, SYM_CHECKS[val] % name)

        else:
            raise RuntimeError('invalid operand type')

    yield from (
        'vfmt(%s) == vfmt(%s)' % (v[i], v[i + 1])
        for v in dynvec.values()
        for i in range(len(v) - 1)
    )

    yield from (
        'isSameType(%s, %s)' % (v[i], v[i + 1])
        for v in fixedvec.values()
        for i in range(len(v) - 1)
    )

BM_FORMAT   = '%s__bit_mask'
REL_VARNAME = 'delta'
REL_VARINIT = 'uint32(sa_label.RelativeTo(pc))'

BR_ENUMS = {
    '(omitted)' : '_BrOmitted',
    'c'         : 'BrC',
    'j'         : 'BrJ',
    'jc'        : 'BrJC',
}

IMM_ENCODER = {
    'CRm'                             : 'asUimm4(%s)',
    'CRm:Encoding:Hints:Index:by:op2' : 'asUimm7(%s)',
    'N:immr:imms'                     : 'asMaskOp(%s)',
    'Q:imm4'                          : 'asExtIndex(v0, %s)',
    'a:b:c:d:e:f:g:h'                 : 'asUimm8(%s)',
    'b40:b5'                          : 'asUimm6(%s)',
    'imm5'                            : 'asUimm5(%s)',
    'imm6'                            : 'asUimm6(%s)',
    'imm8'                            : 'asFpImm8(%s)',
    'imm12'                           : 'asImm12(%s)',
    'imm16'                           : 'asUimm16(%s)',
    'immr'                            : 'asUimm6(%s)',
    'immr:imms'                       : 'asMaskOp(%s)',
    'imms'                            : 'asUimm6(%s)',
    'mask'                            : 'asUimm4(%s)',
    'nzcv'                            : 'asUimm4(%s)',
    'op1'                             : 'asUimm3(%s)',
    'op2'                             : 'asUimm3(%s)',
    'scale'                           : 'asFpScale(%s)',
    'uimm4'                           : 'asUimm4(%s)',
    'uimm6'                           : 'asUimm6(%s)',
}

IMM_SPECIAL_ENCODER = {
    'MOV_MOVN_32_movewide': { 'hw:imm16': 'asMOVxImm(%s, 32, true)' },
    'MOV_MOVN_64_movewide': { 'hw:imm16': 'asMOVxImm(%s, 64, true)' },
    'MOV_MOVZ_32_movewide': { 'hw:imm16': 'asMOVxImm(%s, 32, false)' },
    'MOV_MOVZ_64_movewide': { 'hw:imm16': 'asMOVxImm(%s, 64, false)' },
}

XBFM_ENCODER = {
    'BFC_BFM_32M_bitfield': {
        'sa_lsb'   : ('immr', '-asUimm5(%s) %% 32'),
        'sa_width' : ('imms', 'asUimm5(%s) - 1'),
    },
    'BFC_BFM_64M_bitfield': {
        'sa_lsb_2'   : ('immr', '-asUimm6(%s) %% 64'),
        'sa_width_1' : ('imms', 'asUimm6(%s) - 1'),
    },
    'BFI_BFM_32M_bitfield': {
        'sa_lsb'   : ('immr', '-asUimm5(%s) %% 32'),
        'sa_width' : ('imms', 'asUimm5(%s) - 1'),
    },
    'BFI_BFM_64M_bitfield': {
        'sa_lsb_2'   : ('immr', '-asUimm6(%s) %% 64'),
        'sa_width_1' : ('imms', 'asUimm6(%s) - 1'),
    },
    'BFXIL_BFM_32M_bitfield': {
        'sa_lsb_1' : ('immr', 'asUimm5(%s)'),
        'sa_width' : ('imms', 'sa_lsb_1 + asUimm5(%s) - 1'),
    },
    'BFXIL_BFM_64M_bitfield': {
        'sa_lsb_3'   : ('immr', 'asUimm6(%s)'),
        'sa_width_1' : ('imms', 'sa_lsb_3 + asUimm6(%s) - 1'),
    },
    'LSL_UBFM_32M_bitfield': {
        'sa_shift_1': ('immr:imms', 'asLSLShift(%s, 32)'),
    },
    'LSL_UBFM_64M_bitfield': {
        'sa_shift_3': ('immr:imms', 'asLSLShift(%s, 64)'),
    },
    'SBFIZ_SBFM_32M_bitfield': {
        'sa_lsb'   : ('immr', '-asUimm5(%s) %% 32'),
        'sa_width' : ('imms', 'asUimm5(%s) - 1'),
    },
    'SBFIZ_SBFM_64M_bitfield': {
        'sa_lsb_2'   : ('immr', '-asUimm6(%s) %% 64'),
        'sa_width_1' : ('imms', 'asUimm6(%s) - 1'),
    },
    'SBFX_SBFM_32M_bitfield': {
        'sa_lsb_1' : ('immr', 'asUimm5(%s)'),
        'sa_width' : ('imms', 'sa_lsb_1 + asUimm5(%s) - 1'),
    },
    'SBFX_SBFM_64M_bitfield': {
        'sa_lsb_3'   : ('immr', 'asUimm6(%s)'),
        'sa_width_1' : ('imms', 'sa_lsb_3 + asUimm6(%s) - 1'),
    },
    'UBFIZ_UBFM_32M_bitfield': {
        'sa_lsb'   : ('immr', '-asUimm5(%s) %% 32'),
        'sa_width' : ('imms', 'asUimm5(%s) - 1'),
    },
    'UBFIZ_UBFM_64M_bitfield': {
        'sa_lsb_2'   : ('immr', '-asUimm6(%s) %% 64'),
        'sa_width_1' : ('imms', 'asUimm6(%s) - 1'),
    },
    'UBFX_UBFM_32M_bitfield': {
        'sa_lsb_1' : ('immr', 'asUimm5(%s)'),
        'sa_width' : ('imms', 'sa_lsb_1 + asUimm5(%s) - 1'),
    },
    'UBFX_UBFM_64M_bitfield': {
        'sa_lsb_3'   : ('immr', 'asUimm6(%s)'),
        'sa_width_1' : ('imms', 'sa_lsb_3 + asUimm6(%s) - 1'),
    },
}

SPECIAL_REGS = {
    'sa_cm'     : 'asUimm4(%s)',
    'sa_cn'     : 'asUimm4(%s)',
    'sa_cond'   : 'uint32(%s.(ConditionCode))',
    'sa_cond_1' : 'uint32(%s.(ConditionCode) ^ 1)',
    'sa_label'  : '%s.(*asm.Label)',
}

REWRITE_FIELDS = {
    'AT_SYS_CR_systeminstrs': {
        'op1:CRm<0>:op2': 'op1:CRm:op2',
    },
    'HINT_HM_hints': {
        'CRm:Encoding:Hints:Index:by:op2': 'CRm:op2',
    },
    'MOV_ORR_asimdsame_only': {
        'Rn': 'Rm:Rn',
    }
}

MISSING_FIELDS = {
    'CPYE_CPY_memcms'       : { 'size': 0b00 },
    'CPYEN_CPY_memcms'      : { 'size': 0b00 },
    'CPYERN_CPY_memcms'     : { 'size': 0b00 },
    'CPYERTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYERTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYERTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYERT_CPY_memcms'     : { 'size': 0b00 },
    'CPYETN_CPY_memcms'     : { 'size': 0b00 },
    'CPYETRN_CPY_memcms'    : { 'size': 0b00 },
    'CPYETWN_CPY_memcms'    : { 'size': 0b00 },
    'CPYET_CPY_memcms'      : { 'size': 0b00 },
    'CPYEWN_CPY_memcms'     : { 'size': 0b00 },
    'CPYEWTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYEWTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYEWTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYEWT_CPY_memcms'     : { 'size': 0b00 },
    'CPYE_CPY_memcms'       : { 'size': 0b00 },
    'CPYFEN_CPY_memcms'     : { 'size': 0b00 },
    'CPYFERN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFERTN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFERTRN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFERTWN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFERT_CPY_memcms'    : { 'size': 0b00 },
    'CPYFETN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFETRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFETWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFET_CPY_memcms'     : { 'size': 0b00 },
    'CPYFEWN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFEWTN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFEWTRN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFEWTWN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFEWT_CPY_memcms'    : { 'size': 0b00 },
    'CPYFE_CPY_memcms'      : { 'size': 0b00 },
    'CPYFMN_CPY_memcms'     : { 'size': 0b00 },
    'CPYFMRN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFMRTN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFMRTRN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFMRTWN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFMRT_CPY_memcms'    : { 'size': 0b00 },
    'CPYFMTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFMTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFMTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFMT_CPY_memcms'     : { 'size': 0b00 },
    'CPYFMWN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFMWTN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFMWTRN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFMWTWN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFMWT_CPY_memcms'    : { 'size': 0b00 },
    'CPYFM_CPY_memcms'      : { 'size': 0b00 },
    'CPYFPN_CPY_memcms'     : { 'size': 0b00 },
    'CPYFPRN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFPRTN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFPRTRN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFPRTWN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFPRT_CPY_memcms'    : { 'size': 0b00 },
    'CPYFPTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFPTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFPTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFPT_CPY_memcms'     : { 'size': 0b00 },
    'CPYFPWN_CPY_memcms'    : { 'size': 0b00 },
    'CPYFPWTN_CPY_memcms'   : { 'size': 0b00 },
    'CPYFPWTRN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFPWTWN_CPY_memcms'  : { 'size': 0b00 },
    'CPYFPWT_CPY_memcms'    : { 'size': 0b00 },
    'CPYFP_CPY_memcms'      : { 'size': 0b00 },
    'CPYMN_CPY_memcms'      : { 'size': 0b00 },
    'CPYMRN_CPY_memcms'     : { 'size': 0b00 },
    'CPYMRTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYMRTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYMRTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYMRT_CPY_memcms'     : { 'size': 0b00 },
    'CPYMTN_CPY_memcms'     : { 'size': 0b00 },
    'CPYMTRN_CPY_memcms'    : { 'size': 0b00 },
    'CPYMTWN_CPY_memcms'    : { 'size': 0b00 },
    'CPYMT_CPY_memcms'      : { 'size': 0b00 },
    'CPYMWN_CPY_memcms'     : { 'size': 0b00 },
    'CPYMWTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYMWTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYMWTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYMWT_CPY_memcms'     : { 'size': 0b00 },
    'CPYM_CPY_memcms'       : { 'size': 0b00 },
    'CPYPN_CPY_memcms'      : { 'size': 0b00 },
    'CPYPRN_CPY_memcms'     : { 'size': 0b00 },
    'CPYPRTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYPRTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYPRTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYPRT_CPY_memcms'     : { 'size': 0b00 },
    'CPYPTN_CPY_memcms'     : { 'size': 0b00 },
    'CPYPTRN_CPY_memcms'    : { 'size': 0b00 },
    'CPYPTWN_CPY_memcms'    : { 'size': 0b00 },
    'CPYPT_CPY_memcms'      : { 'size': 0b00 },
    'CPYPWN_CPY_memcms'     : { 'size': 0b00 },
    'CPYPWTN_CPY_memcms'    : { 'size': 0b00 },
    'CPYPWTRN_CPY_memcms'   : { 'size': 0b00 },
    'CPYPWTWN_CPY_memcms'   : { 'size': 0b00 },
    'CPYPWT_CPY_memcms'     : { 'size': 0b00 },
    'CPYP_CPY_memcms'       : { 'size': 0b00 },

    'SDOT_asimdelem_D'      : { 'size': 0b10 },
    'SDOT_asimdsame2_D'     : { 'size': 0b10 },

    'SETEN_SET_memcms'      : { 'size': 0b00 },
    'SETETN_SET_memcms'     : { 'size': 0b00 },
    'SETET_SET_memcms'      : { 'size': 0b00 },
    'SETE_SET_memcms'       : { 'size': 0b00 },
    'SETGEN_SET_memcms'     : { 'size': 0b00 },
    'SETGETN_SET_memcms'    : { 'size': 0b00 },
    'SETGET_SET_memcms'     : { 'size': 0b00 },
    'SETGE_SET_memcms'      : { 'size': 0b00 },
    'SETGMN_SET_memcms'     : { 'size': 0b00 },
    'SETGMTN_SET_memcms'    : { 'size': 0b00 },
    'SETGMT_SET_memcms'     : { 'size': 0b00 },
    'SETGM_SET_memcms'      : { 'size': 0b00 },
    'SETGPN_SET_memcms'     : { 'size': 0b00 },
    'SETGPTN_SET_memcms'    : { 'size': 0b00 },
    'SETGPT_SET_memcms'     : { 'size': 0b00 },
    'SETGP_SET_memcms'      : { 'size': 0b00 },
    'SETMN_SET_memcms'      : { 'size': 0b00 },
    'SETMTN_SET_memcms'     : { 'size': 0b00 },
    'SETMT_SET_memcms'      : { 'size': 0b00 },
    'SETM_SET_memcms'       : { 'size': 0b00 },
    'SETPN_SET_memcms'      : { 'size': 0b00 },
    'SETPTN_SET_memcms'     : { 'size': 0b00 },
    'SETPT_SET_memcms'      : { 'size': 0b00 },
    'SETP_SET_memcms'       : { 'size': 0b00 },

    'UDOT_asimdelem_D'      : { 'size': 0b10 },
    'UDOT_asimdsame2_D'     : { 'size': 0b10 },
}

NOCHECK_FIELDS = {
    'MSR_SI_pstate': [
        ('sa_imm', 'sa_pstatefield'),
    ],
}

class SwitchLit(dict):
    def __getitem__(self, key: str) -> str:
        return str(key[key.startswith('#'):])

    def __contains__(self, key: str) -> bool:
        try:
            int(key[key.startswith('#'):])
        except ValueError:
            return False
        else:
            return True

def make_mask(v: str) -> str:
    return ''.join(str(int(c != 'x')) for c in v)

def encode_defs(
    form     : InstrForm,
    sw_cond  : str,
    var_name : str,
    name_tab : dict[str, str] | SwitchLit,
    vals     : dict[str, str | tuple[str, str, dict[str, str]]],
    opts     : dict[str, str | list[str] | tuple[str, list[str]]],
    err_msg  : str,
):
    initv = None
    basic = True
    multi = False
    field = form.fields[var_name]

    if not isinstance(field, Definition):
        if var_name not in VECTOR_SIZE_CHECKS.get(form.enctab.name, {}):
            raise RuntimeError('field should be definitions: ' + repr(var_name))
        else:
            return

    for k, v in field.bits.items():
        if v and k != 'RESERVED' and not k.startswith('SEE'):
            if k not in name_tab:
                basic = False
                break

    if field.defv:
        if field.defv not in name_tab:
            raise RuntimeError('undefined default value')

        if field.defv in field.bits:
            defv = field.bits[field.defv]
            defv = ['uint32(0b%s)' % v.replace('x', '0') for v in sorted(defv)]

            if len(defv) != 1:
                raise RuntimeError('multiple default value')
            else:
                initv = defv[0]

    if not basic:
        refs = ''
        defs = None

        if len(field.refs) != 1:
            raise RuntimeError('multiple field references: ' + repr(field.name))

        for k, v in form.fields.items():
            if isinstance(v, Definition) and field.refs[0] in v.refs:
                refs, defs = k, v
                break
        else:
            raise RuntimeError('cannot find reference field: ' + repr(field.refs))

        if len(defs.refs) != 1:
            vidx = defs.refs.index(field.refs[0])
            nbit = form.bits.refs[field.refs[0]][1]
            nshr = sum([form.bits.refs[x][1] for x in defs.refs[:vidx]])

            if refs not in REG_CHECKS:
                if not nshr:
                    refs = 'mask(%s, %d)' % (refs, nshr, nbit)
                else:
                    refs = 'ubfx(%s, %d, %d)' % (refs, nshr, nbit)

        if initv is not None:
            refs = (refs, initv, {})

        cond = []
        opts[var_name] = cond
        vals[var_name] = refs

        for k, vv in field.bits.items():
            if vv and k != 'RESERVED' and not k.startswith('SEE'):
                if len(vv) != 1:
                    raise RuntimeError('ambiguous bit pattern')

                for v in vv:
                    if pat := re.match(r'\((\d+)-U[Ii]nt\(%s\)\)' % field.name, k):
                        cond.append('case 0b%s: %s = %s - uint32(%s)' % (v.replace('x', '0'), var_name, pat.group(1), sw_cond))
                    elif pat := re.match(r'\(U[Ii]nt\(%s\)-(\d+)\)' % field.name, k):
                        cond.append('case 0b%s: %s = uint32(%s) + %s' % (v.replace('x', '0'), var_name, sw_cond, pat.group(1)))
                    else:
                        raise RuntimeError('unrecognized pattern: ' + repr(k))

        else:
            cond.append('default: panic("aarch64: %s")' % err_msg)

    else:
        bitp = 1
        cond = []

        for k, vv in field.bits.items():
            if vv and k != 'RESERVED' and not k.startswith('SEE'):
                if len(vv) != 1:
                    bitp = len(vv)
                    break

        if bitp != 1:
            assert initv is None, 'default value conflicts with bitmap'
            initv = 'bm:%d' % bitp

        opts[var_name] = cond
        vals[var_name] = sw_cond if initv is None else (sw_cond, initv, {})

        for k, vv in field.bits.items():
            if vv and k != 'RESERVED' and not k.startswith('SEE'):
                swc = name_tab[k]
                bvs = ['0b' + v.replace('x', '0') for v in sorted(vv)]

                if bitp != 1 or 'x' in list(vv)[0]:
                    multi = True

                if bitp == 1:
                    cond.append('case %s: %s = %s' % (swc, var_name, bvs[0]))
                else:
                    cond.append('case %s: %s = [%d]uint32{%s}' % (swc, var_name, bitp, ', '.join(bvs)))

        if multi:
            mask = []
            bvar = BM_FORMAT % var_name

            opts[bvar] = mask
            vals[bvar] = (sw_cond, 'bm:%d' % bitp, {}) if bitp != 1 else sw_cond

            for k, vv in field.bits.items():
                if vv and k != 'RESERVED' and not k.startswith('SEE'):
                    swc = name_tab[k]
                    bvs = ['0b' + make_mask(v) for v in sorted(vv)]

                    if bitp == 1:
                        mask.append('case %s: %s = %s' % (swc, bvar, bvs[0]))
                    else:
                        mask.append('case %s: %s = [%d]uint32{%s}' % (swc, bvar, bitp, ', '.join(bvs)))

            else:
                mask.append('default: panic("aarch64: %s")' % err_msg)

        if not cond:
            raise RuntimeError('not encodable: ' + form.inst.mnemonic)
        else:
            cond.append('default: panic("aarch64: %s")' % err_msg)

def encode_modifier(
    form        : InstrForm,
    name        : str,
    mod         : Mod,
    vals        : dict[str, str | tuple[str, str, dict[str, str]]],
    opts        : dict[str, str | list[str] | tuple[str, list[str]]],
    optional    : bool,
    *extra_cond : str,
):
    if mod.mod in AsmTemplate.modifiers:
        field = form.fields[mod.mod]
        sw_cond = '%s.(Modifier).Type()' % name
        assert isinstance(field, Definition), 'invalid modifier field encoding'

        if all(k in MODIFIER_TYPES for k in field.bits if k != 'RESERVED' and not k.startswith('SEE')):
            encode_defs(form, sw_cond, mod.mod, MODIFIER_TYPES, vals, opts, "invalid modifier flags")

        else:
            cond = []
            opts[mod.mod] = cond
            vals[mod.mod] = ('', 'ty:uint32', {})

            for refs, defs in field.bits.items():
                if refs != 'RESERVED' and not refs.startswith('SEE'):
                    item = refs.split()
                    kind, amount = item

                    if len(defs) != 1:
                        raise RuntimeError('invalid modifier bits')

                    cond.append('case modt(%s) == %s && modn(%s) == %d: %s = 0b%s' % (
                        name,
                        MODIFIER_TYPES[kind],
                        name,
                        int(amount[1:]),
                        mod.mod,
                        next(iter(defs)),
                    ))

            if not cond:
                raise RuntimeError('invalid modifier flags')
            else:
                cond.append('default: panic("aarch64: invalid modifier flags")')

        if optional:
            conds = opts.pop(mod.mod)
            assert isinstance(conds, list), 'invalid opts'
            opts[mod.mod] = (str(And(*extra_cond)), conds)

    if mod.imm is not None and isinstance(mod.imm[0], Imm):
        args = mod.imm[0].name
        field = form.fields[args]
        value = 'uint32(%s.(Modifier).Amount())' % name

        if isinstance(field, Definition):
            encode_defs(form, value, args, SwitchLit(), vals, opts, 'invalid modifier amount')

        else:
            if not field.defv:
                vals[args] = value
            elif field.defv.isdigit():
                vals[args] = (value, 'uint32(%s)' % field.defv, {})
            else:
                raise RuntimeError('invalid modifier amount')

        if optional:
            opts[args] = str(And(*extra_cond))

def encode_operand(
    form     : InstrForm,
    name     : str,
    val      : Reg | Vec | Mem | Mod | Imm | Lit | Sym,
    vals     : dict[str, str | tuple[str, str, dict[str, str]]],
    opts     : dict[str, str | list[str] | tuple[str, list[str]]],
    *optcond : str,
):
    if isinstance(val, Reg):
        if optcond:
            if val.altr is not None or \
                val.size is not None or \
                val.mode is not None or \
                val.vidx is not None:
                raise RuntimeError('optional complex reg operand is not supported')

            else:
                defr = form.fields[val.name].defv
                expr = 'uint32(%s.(asm.Register).ID())' % name
                opts[val.name] = str(And(*optcond))

                if not defr:
                    vals[val.name] = expr
                elif defr.isidentifier():
                    vals[val.name] = (expr, 'uint32(%s.ID())' % defr, {})
                elif defr[0] == defr[-1] == "'":
                    vals[val.name] = (expr, 'uint32(0b%s)' % defr[1:-1], {})
                else:
                    raise RuntimeError('invalid register default value: ' + defr)

        elif val.mode is not None and val.vidx is not None:
            if not val.name.startswith('sa_v'):
                raise RuntimeError('invalid indexing on non-vector registers: ' + str(val))

            mode, vidx = val.mode, val.vidx
            vals[val.name] = 'uint32(%s.(VidxRegister).ID())' % name

            if not isinstance(mode, Tag):
                encode_defs(form, 'vmoder(%s)' % name, mode, VECTOR_MODES, vals, opts, 'unreachable')

            if isinstance(vidx, Imm):
                vals[vidx.name] = 'uint32(vidxr(%s))' % name

        elif val.mode is not None:
            mode = val.mode
            vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name

            if not isinstance(mode, Tag):
                encode_defs(form, 'vfmt(%s)' % name, mode, VECTOR_TYPES, vals, opts, 'unreachable')

        elif val.size is not None and val.size == 'sa_r':
            vmap = {}
            vmap['W'] = 'isWr(%s)' % name
            vmap['X'] = 'isXr(%s)' % name
            vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name
            encode_defs(form, 'true', val.size, vmap, vals, opts, 'unreachable')

        elif val.size is not None:
            if not val.size.startswith('sa_v'):
                raise RuntimeError('invalid fixed vec size')

            err = 'invalid scalar operand size for ' + form.inst.mnemonic
            vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name
            encode_defs(form, '%s.(type)' % name, val.size, SCALAR_TYPES, vals, opts, err)

        else:
            if '_plus_' not in val.name:
                if val.name in SPECIAL_REGS:
                    vals[val.name] = SPECIAL_REGS[val.name] % name
                else:
                    vals[val.name] = 'uint32(%s.(asm.Register).ID())' % name

    elif isinstance(val, Vec):
        if optcond:
            raise RuntimeError('optional vector operand is not supported')

        elm = 'velm'
        vec = 'Vector'

        if val.vidx is not None:
            elm = 'vmodei'
            vec = 'IndexedVector'

        reg0 = val.regs[0]
        vals[reg0] = 'uint32(%s.(%s).ID())' % (name, vec)

        if not isinstance(val.mode, Tag):
            err = 'invalid vector arrangement for ' + form.inst.mnemonic
            encode_defs(form, '%s(%s)' % (elm, name), val.mode, VECTOR_TYPES, vals, opts, err)

        if isinstance(val.vidx, Imm):
            vals[val.vidx.name] = 'uint32(%s.(IndexedVector).Index())' % name

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
                # TODO: default value for extend options ?
                encode_modifier(form, 'mext(%s)' % name, ext, vals, opts, opx, 'isMod(mext(%s))' % name)

    elif isinstance(val, Mod):
        # TODO: default value for extend options ?
        encode_modifier(form, name, val, vals, opts, bool(optcond), *optcond)

    elif isinstance(val, Imm):
        ref = form.fields[val.name]
        bfm = XBFM_ENCODER.get(form.enctab.name, {}).get(val.name)
        ime = IMM_SPECIAL_ENCODER.get(form.enctab.name, {}).get(ref.name)

        if ime is None:
            ime = IMM_ENCODER.get(ref.name)

        if ime is not None:
            if not optcond:
                vals[val.name] = ime % name

            else:
                defv = ref.defv
                opts[val.name] = str(And(*optcond))

                if not defv:
                    vals[val.name] = ime % name
                elif defv.isdigit():
                    vals[val.name] = (ime % name, 'uint32(%s)' % defv, {})
                else:
                    raise RuntimeError('invalid immediate default value: ' + defv)

        elif isinstance(ref, Definition):
            err = "invalid operand '%s' for %s" % (val.name, form.inst.mnemonic)
            encode_defs(form, 'asLit(%s)' % name, val.name, SwitchLit(), vals, opts, err)

        elif not ref.name and bfm is not None:
            if optcond:
                raise RuntimeError('xBFM bit field cannot be optional')

            ref.name, enc = bfm
            vals[val.name] = enc % name

        else:
            raise RuntimeError("cannot encode immediate field '%s.%s'" % (form.enctab.name, val.name))

    elif isinstance(val, Lit):
        if optcond:
            raise RuntimeError('optional lit operand is not supported')

    elif isinstance(val, Sym):
        match val:
            case Sym.CSYNC:
                if optcond:
                    raise RuntimeError('optional CSYNC is not supported')

            case Sym.DSYNC:
                if optcond:
                    raise RuntimeError('optional DSYNC is not supported')

            case Sym.PRFOP:
                if optcond:
                    raise RuntimeError('optional prefetch is not supported')
                else:
                    vals['sa_prfop'] = 'uint32(%s.(PrefetchOp))' % name

            case Sym.RCTX:
                if optcond:
                    raise RuntimeError('optional RCTX is not supported')

            case Sym.RPRFOP:
                if optcond:
                    raise RuntimeError('optional range prefetch is not supported')
                else:
                    vals['sa_rprfop'] = '%s.(RangePrefetchOp).encode()' % name

            case Sym.OPTION:
                vals['sa_option'] = ('%s.(BarrierOption)' % name, 'SY', {})
                vals['sa_imm'] = 'uint32(sa_option)'

                if optcond:
                    opts['sa_option'] = str(And(*optcond))

            case Sym.OPTION_NXS:
                if optcond:
                    raise RuntimeError('optional nXS option is not supported')
                else:
                    vals['sa_option_1'] = '%s.(BarrierOption).nxs()' % name

            case Sym.AT_OPTION:
                if optcond:
                    raise RuntimeError('optional AT option is not supported')
                else:
                    vals['sa_at_op'] = 'uint32(%s.(ATOption))' % name

            case Sym.BRB_OPTION:
                if optcond:
                    raise RuntimeError('optional BRB option is not supported')
                else:
                    vals['sa_brb_op'] = 'uint32(%s.(BRBOption))' % name

            case Sym.DC_OPTION:
                if optcond:
                    raise RuntimeError('optional DC option is not supported')
                else:
                    vals['sa_dc_op'] = 'uint32(%s.(DCOption))' % name

            case Sym.IC_OPTION:
                if optcond:
                    raise RuntimeError('optional IC option is not supported')
                else:
                    vals['sa_ic_op'] = 'uint32(%s.(ICOption))' % name

            case Sym.SME_OPTION:
                if not optcond:
                    vals['sa_pstatefield'] = 'uint32(%s.(SMEOption))' % name
                else:
                    opts['sa_pstatefield'] = str(And(*optcond))
                    vals['sa_pstatefield'] = ('uint32(%s.(SMEOption))' % name, 'uint32(0b0111)', {})

            case Sym.TLBI_OPTION:
                if optcond:
                    raise RuntimeError('optional TLBI option is not supported')
                else:
                    vals['sa_tlbi_op'] = 'uint32(%s.(TLBIOption))' % name

            case Sym.TLBIP_OPTION:
                if optcond:
                    raise RuntimeError('optional TLBIP option is not supported')
                else:
                    vals['sa_tlbip_op'] = 'uint32(%s.(TLBIOption))' % name

            case Sym.PSTATE_FIELD:
                if optcond:
                    raise RuntimeError('optional pstatefield is not supported')
                else:
                    vals['sa_pstatefield'] = 'uint32(%s.(PStateField)) << 4' % name

            case Sym.SYSREG:
                if optcond:
                    raise RuntimeError('optional sysreg is not supported')
                else:
                    vals['sa_systemreg'] = 'uint32(%s.(SystemRegister))' % name

            case Sym.TARGETS:
                expr = '%s.(BranchTarget)' % name
                vals['sa_targets'] = (expr, '_BrOmitted', BR_ENUMS)

                if optcond:
                    opts['sa_targets'] = str(And(*optcond))

            case _:
                raise RuntimeError('invalid symbol: ' + repr(val))

    else:
        raise RuntimeError('invalid operand type')

INSTR_CLASSES = {
    'advsimd'   : 'DomainAdvSimd',
    'float'     : 'DomainFloat',
    'fpsimd'    : 'DomainFpSimd',
    'general'   : 'asm.DomainGeneric',
    'mortlach'  : 'DomainSME',
    'mortlach2' : 'DomainSME2',
    'sve'       : 'DomainSVE',
    'sve2'      : 'DomainSVE2',
    'system'    : 'DomainSystem',
}

BRANCH_CONDITIONS = [
    ('EQ', 0b0000),
    ('NE', 0b0001),
    ('CS', 0b0010),
    ('HS', 0b0010),
    ('CC', 0b0011),
    ('LO', 0b0011),
    ('MI', 0b0100),
    ('PL', 0b0101),
    ('VS', 0b0110),
    ('VC', 0b0111),
    ('HI', 0b1000),
    ('LS', 0b1001),
    ('GE', 0b1010),
    ('LT', 0b1011),
    ('GT', 0b1100),
    ('LE', 0b1101),
    ('AL', 0b1110),
]

def rebuild_bcc(cond: str, bits: int, forms: list[InstrForm]) -> list[InstrForm]:
    return [
        InstrForm(
            text   = f.text.replace('<cond>', cond.lower()),
            inst   = Instr(
                mnemonic = '%s.%s' % (f.inst.mnemonic, cond),
                operands = f.inst.operands,
                modifier = None,
            ),
            feat   = f.feat[:],
            bits   = f.bits,
            opts   = f.opts,
            args   = { **f.args, 'cond': bits },
            enctab = f.enctab,
            fields = f.fields,
        )
        for f in forms
    ]

def rebuild_upper_half(pat: str, sfx: str, forms: list[InstrForm], **args: int) -> list[InstrForm]:
    return [
        InstrForm(
            text   = f.text.replace(pat, sfx),
            inst   = Instr(
                mnemonic = f.inst.mnemonic + sfx,
                operands = f.inst.operands,
                modifier = None,
            ),
            feat   = f.feat[:],
            bits   = f.bits,
            opts   = f.opts,
            args   = { **f.args, **args },
            enctab = f.enctab,
            fields = f.fields,
        )
        for f in forms
    ]

def preprocess_instr_forms(ftab: dict[str, list[InstrForm]]) -> Iterator[tuple[str, list[InstrForm]]]:
    for name, forms in sorted(ftab.items(), key = lambda x: x[0]):
        mod = None
        nomod = []
        withmod = []

        for form in forms:
            if form.inst.modifier is None:
                nomod.append(form)
            elif mod is not None and mod != form.inst.modifier:
                raise RuntimeError('multiple modifiers')
            else:
                mod = form.inst.modifier
                withmod.append(form)

        match mod:
            case None:
                if nomod:
                    yield name, nomod
                else:
                    raise RuntimeError('non-encodable instruction ' + repr(name))

            case 'sa_2':
                yield name, nomod + rebuild_upper_half('{2}', '', withmod, __ensure__Q = 0)
                yield name + '2', rebuild_upper_half('{2}', '2', withmod, __ensure__Q = 1)

            case 'sa_bt':
                if nomod:
                    yield name, nomod

                yield name + 'B', rebuild_upper_half('<bt>', 'B', withmod, Q = 0)
                yield name + 'T', rebuild_upper_half('<bt>', 'T', withmod, Q = 1)

            case 'sa_cond':
                if nomod:
                    yield name, nomod

                for cond, bits in BRANCH_CONDITIONS:
                    yield name + cond, rebuild_bcc(cond, bits, withmod)

            case _:
                raise RuntimeError('unrecognized instruction modifier for %s: %s' % (name, mod))

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

for mnemonic, forms in preprocess_instr_forms(formtab):
    nops = set()
    always = False
    status('* Instruction:', mnemonic)
    assert forms, 'instruction %s have no form' % mnemonic

    cid = None
    iid = None
    iids = sorted(set(iidtab[f.enctab.name] for f in forms))
    forms.sort(key = lambda f: iids.index(iidtab[f.enctab.name]))

    if len(iids) > 1:
        cat_name = '%d categories' % len(iids)
    else:
        cat_name = 'one single category'

    if len(forms) > 1:
        form_name = '%d forms' % len(forms)
    else:
        form_name = 'one single form'

    cc.line('// %s instruction have %s from %s:' % (mnemonic, form_name, cat_name))
    cc.line('//')

    for form in forms:
        cid = iidtab[form.enctab.name]
        nop = len(form.inst.operands.req)

        if cid != iid:
            if iid is not None:
                des = destab[iid]
                cc.line('//')

                for line in des.authored:
                    if not line:
                        cc.line('//')
                    else:
                        cc.line('// ' + line)

                for line in des.notes:
                    if not line:
                        cc.line('//')
                    else:
                        cc.line('// ' + line)

            iid = cid
            idx = iids.index(cid)

            for line in break_line(destab[cid].brief.rstrip(), 0, '%d. ' % (idx + 1)):
                if not line:
                    cc.line('//')
                else:
                    cc.line('// ' + line)
            else:
                cc.line('//')

        cc.line('//    %s' % form.text)
        nops.add(nop)

        if form.inst.operands.opt:
            nops.add(nop + len(form.inst.operands.opt))

    if iid is not None:
        des = destab[iid]
        cc.line('//')

        for line in des.authored:
            if not line:
                cc.line('//')
            else:
                cc.line('// ' + line)

        for line in des.notes:
            if not line:
                cc.line('//')
            else:
                cc.line('// ' + line)

    nfix = min(nops)
    base = ['v%d' % i for i in range(nfix)]

    if nfix == max(nops):
        if not nfix:
            cc.line('func (self *Program) %s() *Instruction {' % mnemonic)
            cc.indent()
            cc.line('p := self.alloc("%s", 0, asm.Operands {})' % mnemonic)
        else:
            cc.line('func (self *Program) %s(%s interface{}) *Instruction {' % (mnemonic, ', '.join(base)))
            cc.indent()
            cc.line('p := self.alloc("%s", %d, asm.Operands { %s })' % (mnemonic, nfix, ', '.join(base)))

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
                cc.line('case 0  : p = self.alloc("%s", 0, asm.Operands {})' % mnemonic)
            else:
                args = base[:] + ['vv[%d]' % i for i in range(argc - nfix)]
                cc.line('case %d  : p = self.alloc("%s", %d, asm.Operands { %s })' % (argc - nfix, mnemonic, argc, ', '.join(args)))

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

        elif len(mt) == 1 or len(mts) + cc.level * 4 + 5 <= MAX_LINE_WIDTH:
            cc.line('if %s {' % mts)
            cc.indent()

        else:
            if not isinstance(mt[0], Or) or len(mt[0]) == 1:
                cc.line('if %s &&' % mt[0])
            else:
                cc.line('if (%s) &&' % mt[0])

            for i, mm in enumerate(mt[1:], 1):
                if i == len(mt) - 1:
                    cc.line('   (%s) {' % mm if isinstance(mm, Or) and len(mm) > 1 else '   %s {' % mm)
                else:
                    cc.line('   (%s) &&' % mm if isinstance(mm, Or) and len(mm) > 1 else '   %s &&' % mm)
            else:
                cc.indent()

        for ft in form.feat:
            cc.line('self.Arch.Require(%s)' % ft)

        fmap = {}
        args = []
        cc.line('p.Domain = %s' % INSTR_CLASSES[form.opts.get('instr-class', 'general')])

        if form.inst.modifier is not None:
            raise RuntimeError('instruction modifiers should have been expanded')

        vals = OnceDict()
        opts = OnceDict()
        lastcond = None

        for i, val in enumerate(form.inst.operands.req):
            if i < nfix:
                encode_operand(form, 'v%d' % i, val, vals, opts)
            else:
                encode_operand(form, 'vv[%d]' % (i - nfix), val, vals, opts)

        for i, val in enumerate(form.inst.operands.opt):
            if not isinstance(val, (Reg, Mod, Imm, Sym)):
                raise RuntimeError('invalid optional operand')
            else:
                idx = len(form.inst.operands.req) - nfix + i
                encode_operand(form, 'vv[%d]' % idx, val, vals, opts, 'len(vv) == %d' % len(form.inst.operands.opt))

        for var in sorted(opts):
            if not isinstance(vals[var], tuple):
                cc.line('var %s uint32' % var)
            elif vals[var][1].startswith('bm:'):
                cc.line('var %s [%d]uint32' % (var, int(vals[var][1][3:])))
            elif vals[var][1].startswith('ty:'):
                cc.line('var %s %s' % (var, vals[var][1][3:]))
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
                continue

            term = opts[k]
            swconds = None
            optcond = None

            if isinstance(term, str):
                optcond = term
            elif isinstance(term, list):
                swconds = term
            elif isinstance(term, tuple):
                optcond, swconds = term
            else:
                raise RuntimeError('invalid optional cond')

            if optcond != lastcond:
                if lastcond is not None:
                    cc.dedent()
                    cc.line('}')
                    lastcond = None

                if optcond is not None:
                    cc.line('if %s {' % optcond)
                    cc.indent()
                    lastcond = optcond

            if swconds is None:
                cc.line('%s = %s' % (k, v))
                continue

            if not v:
                cc.line('switch {')
                cc.indent()
            else:
                cc.line('switch %s {' % v)
                cc.indent()

            for line in swconds:
                cc.line(line)

            cc.dedent()
            cc.line('}')

        if lastcond is not None:
            cc.dedent()
            cc.line('}')

        for key, fv in form.fields.items():
            nb = 0
            fn = []
            st = False
            fx = REWRITE_FIELDS.get(form.enctab.name, {}).get(fv.name, fv.name)

            for s in fx.split(':'):
                if st:
                    st = '>' not in s
                    fn[-1] += ':' + s
                else:
                    st = '<' in s and '>' not in s
                    fn.append(s)

            if fx == form.inst.modifier:
                continue

            if len(fn) == 1:
                fmap.setdefault(fx, []).append((key, key, BM_FORMAT % key))
                continue

            for x in fn:
                _, n = form.bits.refs[x]
                nb += n

            for x in fn:
                _, n = form.bits.refs[x]
                nb -= n

                if key in REG_CHECKS:
                    fmt = '%s'
                elif not nb:
                    fmt = 'mask(%%s, %d)' % n
                else:
                    fmt = 'ubfx(%%s, %d, %d)' % (nb, n)

                fmap.setdefault(x, []).append((
                    key,
                    fmt % key,
                    fmt % (BM_FORMAT % key),
                ))

        for arg in form.enctab.args:
            fvs = fmap.get(arg)
            val = form.args.get(arg)

            if val is None:
                val = MISSING_FIELDS.get(form.enctab.name, {}).get(arg)

            if val is not None:
                args.append(str(val))
                continue

            if fvs is not None:
                vx = None
                fvs.sort(key = lambda v: (v[0] in vals and vals[v[0]][1].startswith('bm:'), v[0]))

                for r, v, m in fvs:
                    if (r in opts or r in vals) and not vals[r][1].startswith('bm:'):
                        vx = v
                        break

                if vx is None:
                    raise RuntimeError('missing field ' + repr(arg))

                args.append(vx)
                continue

            ok = False
            i, n = form.bits.refs[arg]
            bits = form.bits.bits[i:i + n]

            cc.line('%s := uint32(0b%s)' % (arg, ''.join('1' if v else '0' for v in bits[::-1])))
            args.append(arg)

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

                    for sel, bx in fv.bits.items():
                        if len(bx) != 1:
                            raise RuntimeError('ambiguous compositie field ' + repr(bn))
                        else:
                            for bits in bx:
                                if 'x' in bits:
                                    raise RuntimeError('ambiguous compositie field ' + repr(bn))
                                elif len(bits) != be - bs + 1:
                                    raise RuntimeError('bit count mismatch for composite field ' + repr(bn))
                                elif not isinstance(vals[fn], tuple) or not vals[fn][2]:
                                    cc.line('case %s: %s |= 0b%s << %d' % (sel, arg, bits, bs))
                                else:
                                    cc.line('case %s: %s |= 0b%s << %d' % (vals[fn][2][sel], arg, bits, bs))

                    cc.line('default: panic("aarch64: invalid combination of operands for %s")' % mnemonic)
                    cc.dedent()
                    cc.line('}')

                elif bn in form.bits.refs:
                    j, k = form.bits.refs[bn]
                    x, y = i + n - 1, j + k - 1

                    if i <= j and y <= x:
                        ok = True
                        bv = fv[0][1]

                        if i == j:
                            cc.line('%s |= %s' % (arg, bv))
                        else:
                            cc.line('%s |= %s << %d' % (arg, bv, j - i))

            if not ok:
                print(form)
                raise RuntimeError('invalid field ' + repr(arg))

        for fv in fmap.values():
            terms = []
            exprs = [v for v in fv if v[0] in opts or v[0] in vals]

            for (r0, v0, m0), (r1, v1, m1) in zip(exprs, exprs[1:]):
                do_check = True
                bm0, bm1 = 0, 0
                rm0, rm1 = BM_FORMAT % r0, BM_FORMAT % r1

                if form.enctab.name in NOCHECK_FIELDS:
                    for f0, f1 in NOCHECK_FIELDS[form.enctab.name]:
                        if (r0, r1) in ((f0, f1), (f1, f0)):
                            do_check = False
                            break

                if not do_check:
                    continue

                if isinstance(vals[r0], tuple) and vals[r0][1].startswith('bm:'):
                    bm0 = int(vals[r0][1][3:])

                if isinstance(vals[r1], tuple) and vals[r1][1].startswith('bm:'):
                    bm1 = int(vals[r1][1][3:])

                if bm1:
                    if bm0:
                        raise RuntimeError('multiple bit masks: bm0')
                    elif rm1 not in vals:
                        raise RuntimeError('bitmask confliction: bm0')
                    elif rm0 not in vals:
                        terms.append('!matchany(%s, &%s[0], &%s[0], %d)' % (v0, v1, m1, bm1))
                    else:
                        terms.append('!matchany(%s & %s, &%s[0], &%s[0], %d)' % (v0, m0, v1, m1, bm1))

                elif bm0:
                    if bm1:
                        raise RuntimeError('multiple bit masks: bm1')
                    elif rm0 not in vals:
                        raise RuntimeError('bitmask confliction: bm1')
                    elif rm1 not in vals:
                        terms.append('!matchany(%s, &%s[0], &%s[0], %d)' % (v1, v0, m0, bm0))
                    else:
                        terms.append('!matchany(%s & %s, &%s[0], &%s[0], %d)' % (v1, m1, v0, m0, bm0))

                else:
                    if rm0 in vals and rm1 in vals:
                        terms.append('%s & %s != %s & %s' % (v0, m0, v1, m1))
                    elif rm0 in vals:
                        terms.append('%s & %s != %s' % (v0, m0, v1))
                    elif rm1 in vals:
                        terms.append('%s != %s & %s' % (v0, v1, m1))
                    else:
                        terms.append('%s != %s' % (v0, v1))

            if terms:
                conds = ' || '.join(terms)
                width = len(conds) + cc.level * 4 + 5

                if len(terms) == 1 or width <= MAX_LINE_WIDTH:
                    cc.line('if %s {' % ' || '.join(terms))
                    cc.indent()

                else:
                    hd, *rem, tr = terms
                    cc.line('if %s ||' % hd)

                    for v in rem:
                        cc.line('   %s ||' % v)

                    cc.line('   %s {' % tr)
                    cc.indent()

                cc.line('panic("aarch64: invalid combination of operands for %s")' % mnemonic)
                cc.dedent()
                cc.line('}')

        for arg, exp in form.args.items():
            if arg.startswith('__ensure__'):
                cc.line('if %s != %d {' % (args[form.enctab.args.index(arg[10:])], exp))
                cc.indent()
                cc.line('panic("aarch64: invalid combination of operands for %s")' % mnemonic)
                cc.dedent()
                cc.line('}')

        for arg, exp in form.args.items():
            if arg.startswith('__ensure__'):
                args[form.enctab.args.index(arg[10:])] = str(exp)

        deferred = ['sa_label' in v for v in args].count(True)
        encoding = '%s(%s)' % (form.enctab.func, ', '.join(
            v.replace('sa_label', REL_VARINIT if deferred == 1 else REL_VARNAME)
            for v in args
        ))

        if not deferred:
            if len(encoding) + cc.level * 4 + 10 <= MAX_LINE_WIDTH:
                cc.line('return p.setins(%s)' % encoding)

            else:
                cc.line('return p.setins(%s(' % form.enctab.func)
                cc.indent()

                for arg in args:
                    cc.line(arg + ',')

                cc.dedent()
                cc.line('))')

        elif deferred == 1:
            if len(encoding) + cc.level * 4 + 45 <= MAX_LINE_WIDTH:
                cc.line('return p.setenc(func(pc uintptr) uint32 { return %s })' % encoding)

            else:
                cc.line('return p.setenc(func(pc uintptr) uint32 {')
                cc.indent()
                cc.line('return %s(' % form.enctab.func)
                cc.indent()

                for arg in args:
                    cc.line(arg.replace('sa_label', REL_VARINIT) + ',')

                cc.dedent()
                cc.line(')')
                cc.dedent()
                cc.line('})')

        else:
            cc.line('return p.setenc(func(pc uintptr) uint32 {')
            cc.indent()
            cc.line('%s := %s' % (REL_VARNAME, REL_VARINIT))

            if len(encoding) + cc.level * 4 + 7 <= MAX_LINE_WIDTH:
                cc.line('return %s' % encoding)

            else:
                cc.line('return %s(' % form.enctab.func)
                cc.indent()

                for arg in args:
                    cc.line(arg.replace('sa_label', REL_VARNAME) + ',')

                cc.dedent()
                cc.line(')')

            cc.dedent()
            cc.line('})')

        if mt:
            cc.dedent()
            cc.line('}')

    if not always:
        if len(forms) != 1:
            cc.line('// none of above')

        cc.line('p.Free()')
        cc.line('panic("aarch64: invalid combination of operands for %s")' % mnemonic)

    cc.dedent()
    cc.line('}')
    cc.line()

with open('arch/aarch64/instructions.go', 'w') as fp:
    fp.write('\n'.join(cc.buf))