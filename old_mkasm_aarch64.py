#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import operator
import itertools

from typing import Iterator
from typing import NamedTuple

from xml.etree import ElementTree
from xml.etree.ElementTree import Element

HAND_ENCODED = {
    'DSB',
    'FCVTZS_advsimd_fix',
    'FCVTZU_advsimd_fix',
}

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
            if None in self.bits[p:p + n]:
                char += 1
                bits[p:p + n] = ['\x1b[3%dmx\x1b[0m' % (char % 5 + 1)] * n

        kfns = operator.itemgetter(1)
        vals = [(v, (p, n)) for v, (p, n) in self.refs.items() if None in self.bits[p:p + n]]

        for i, (v, (p, n)) in enumerate(sorted(vals, key = kfns, reverse = True)):
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

class TokenStream:
    name: str
    vals: list['str | Token']

    class Token(NamedTuple):
        name: str
        text: str

        def __repr__(self) -> str:
            return '\x1b[31m{%s}\x1b[0m' % self.name

        @classmethod
        def parse(cls, item: Element) -> 'TokenStream.Token':
            return cls(
                text = item.text or '',
                name = item.attrib['link'],
            )

    def __init__(self, name: str, vals: list[str | Token]):
        self.name = name
        self.vals = vals

    def __iter__(self) -> Iterator[str | Token]:
        for item in self.vals:
            if isinstance(item, self.Token):
                yield item
            else:
                for val in item.split():
                    if all(c.isalnum() or c in {'.', '_'} for c in val):
                        yield val
                    else:
                        yield from val

    @classmethod
    def parse(cls, name: str, items: list[Element]):
        return cls(name, list(map(cls.parse_item, items)))

    @classmethod
    def parse_item(cls, item: Element) -> str | Token:
        match item.tag:
            case 'a'    : return cls.Token.parse(item)
            case 'text' : return item.text or ''
            case _      : raise RuntimeError('unexpected tag in assembly template: ' + repr(item.tag))

# data = ElementTree.parse(os.path.join(sys.argv[1], 'ld1_advsimd_mult.xml'))
# instrs = [data.getroot()]

data = ElementTree.parse(os.path.join(sys.argv[1], 'onebigfile.xml'))
instrs = data.findall('.//sect1[@id="iformpages"]//file/instructionsection[@type="instruction"]')

def parse_bit(bit: str) -> str:
    match bit:
        case '0': return '0'
        case '1': return '1'
        case 'x': return '01'
        case _  : raise RuntimeError('invalid bit value: ' + repr(bit))

def parse_props(out: dict[str, str], p: Element):
    for v in p.findall('docvars/docvar'):
        out[v.attrib['key']] = v.attrib['value']

def parse_boxes(ins: Instruction, boxes: list[Element]):
    for box in boxes:
        hibit = int(box.attrib['hibit'])
        width = int(box.attrib.get('width', 1))

        for i, item in enumerate(box.findall('c')):
            match item.text:
                case None  : pass
                case ''    : pass
                case 'x'   : pass
                case 'N'   : pass
                case 'z'   : ins.bits[hibit - i] = 0
                case '0'   : ins.bits[hibit - i] = 0
                case '1'   : ins.bits[hibit - i] = 1
                case '(0)' : ins.bits[hibit - i] = 0
                case '(1)' : ins.bits[hibit - i] = 1
                case v     : raise RuntimeError('invalid cell value: ' + repr(v))

        if box.attrib.get('usename') == '1':
            ins.refs[box.attrib['name']] = (hibit - width + 1, width)

def parse_symdef(defs: Element) -> dict[str, set[str]]:
    rets = {}
    tabs = defs.findall('.//tgroup')
    rows = defs.findall('.//tbody/row')
    dest = defs.attrib['encodedin']

    if len(tabs) != 1:
        raise RuntimeError('expect exactly 1 table: encodedin="%s"' % dest)

    for row in rows:
        bits = []
        syms = row.find('entry[@class="symbol"]')

        for bv in row.findall('entry[@class="bitfield"]'):
            if not bv.text:
                raise RuntimeError('missing symbol or bitfield: encodedin="%s"' % dest)
            else:
                bits.append(bv.text)

        if not bits or syms is None or not syms.text:
            raise RuntimeError('missing symbol or bitfield: encodedin="%s"' % dest)

        for sym in map(str.strip, syms.text.split('|')):
            prod = itertools.product(*map(parse_bit, ''.join(bits)))
            rets.setdefault(sym, set()).update(''.join(v) for v in prod)

    if not rets:
        raise RuntimeError('missing definitions: encodedin="%s"' % dest)
    else:
        return rets

for instr in instrs:
    name = instr.attrib['id']
    title = instr.attrib['title']

    if name in HAND_ENCODED:
        continue

    opts = {}
    fields = {}
    print('>>>>> %s: %s' % (name, title))
    parse_props(opts, instr)

    for expl in instr.findall('explanations/explanation'):
        symbol = expl.find('symbol')
        symacc = expl.find('account')
        symdef = expl.find('definition')

        if symbol is None or (symacc is None) is (symdef is None):
            raise RuntimeError('invalid explanation')

        desc = symbol.text or ''
        name = symbol.attrib['link']

        if symacc is not None:
            assert symdef is None
            defs = Account(symacc.attrib['encodedin'], desc)
        else:
            assert isinstance(symdef, Element)
            defs = Definition(symdef.attrib['encodedin'], desc, parse_symdef(symdef))

        for enc in expl.attrib['enclist'].split(','):
            tab = fields.setdefault(enc.strip(), {})
            tab[name] = defs

    for iclass in instr.findall('classes/iclass'):
        attrs = dict(opts)
        proto = Instruction()

        parse_props(attrs, iclass)
        parse_boxes(proto, iclass.findall('regdiagram/box'))

        for encoding in iclass.findall('encoding'):
            props = dict(attrs)
            instr = Instruction(proto)
            lexer = TokenStream.parse(encoding.attrib['name'], encoding.findall('asmtemplate/*'))

            parse_props(props, encoding)
            parse_boxes(instr, encoding.findall('box'))

            print('-------- %s --------' % lexer.name)
            print('Mnemonic :', props['mnemonic'])
            print('Syntax   :', ' '.join(map(repr, lexer)))
            import pprint
            print('Fields   :', end = ' ')
            pprint.pprint(fields.get(lexer.name, {}), compact = True, sort_dicts = True)
            print('Encoding :')
            print(instr)
            print()
