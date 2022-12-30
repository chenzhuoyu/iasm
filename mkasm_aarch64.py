#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys

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

        vals = [(v, (p, n)) for v, (p, n) in self.refs.items() if None in self.bits[p:p + n]]
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

    def update(self, boxes: list[Element]):
        for box in boxes:
            hibit = int(box.attrib['hibit'])
            width = int(box.attrib.get('width', 1))

            if box.attrib.get('usename') == '1':
                self.refs[box.attrib['name']] = (hibit - width + 1, width)
                continue

            for i, item in enumerate(box.findall('c')):
                match item.text:
                    case '0' : self.bits[hibit - i] = 0
                    case '1' : self.bits[hibit - i] = 1
                    case _   : raise RuntimeError('invalid cell value: ' + repr(item.text))

class InstructionTabEntry(NamedTuple):
    name: str
    desc: str
    itab: Element
    bits: Instruction

    def __repr__(self) -> str:
        return '\n'.join([
            'Instruction %s {' % self.name,
            '    ' + self.desc,
            *('    ' + v for v in repr(self.bits).splitlines()),
            '}',
        ])

cc = CodeGen()
cc.line('// Code generated by "mkasm_aarch64.py", DO NOT EDIT.')
cc.line()
cc.line('package aarch64')
cc.line()

root = ElementTree.parse(os.path.join(sys.argv[1], 'encodingindex.xml')).getroot()
instab = dict[str, InstructionTabEntry]()

if root.tag != 'encodingindex':
    raise RuntimeError('invalid encoding index file')

for iclass in sorted(root.findall('iclass_sect'), key = lambda x: x.attrib['id']):
    name = iclass.attrib['id']
    desc = iclass.attrib['title']
    itab = iclass.find('instructiontable')

    if itab is None:
        raise RuntimeError('missing instruction table for ' + name)

    if itab.attrib['iclass'] != name:
        raise RuntimeError('wrong instruction table iclass for ' + name)

    bits = Instruction()
    bits.update(iclass.findall('regdiagram/box'))

    vals = bits.bits[:]
    args = sorted(bits.refs.items(), key = lambda x: x[1], reverse = True)

    for p, n in bits.refs.values():
        vals[p:p + n] = [0] * n

    if None in vals:
        raise RuntimeError('unset bit in %s: %r' % (name, vals))

    instab[name] = InstructionTabEntry(
        name = name,
        desc = desc,
        itab = itab,
        bits = bits,
    )

    cc.line('// %s: %s' % (name, desc))
    cc.line('func %s(%s uint32) Instruction {' % (name, ', '.join(v for v, _ in args)))
    cc.indent()

    for v, (_, n) in args:
        cc.line('if %s &^ 0b%s != 0 {' % (v, '1' * n))
        cc.indent()
        cc.line('panic("%s: invalid %s")' % (name, v))
        cc.dedent()
        cc.line('}')

    iv = ''.join(map(str, vals))[::-1]
    cc.line('ret := uint32(0x%08x)' % int(iv, 2))

    for v, (p, _) in args:
        if p == 0:
            cc.line('ret |= ' + v)
        else:
            cc.line('ret |= %s << %d' % (v, p))

    cc.line('return Instruction(ret);')
    cc.dedent()
    cc.line('}')
    cc.line()

with open('arch/aarch64/encodings.go', 'w') as fp:
    fp.write('\n'.join(cc.buf))

cc = CodeGen()
cc.line('// Code generated by "mkasm_aarch64.py", DO NOT EDIT.')
cc.line()
cc.line('package aarch64')
cc.line()

for name, entry in sorted(instab.items(), key = lambda v: v[0]):
    print(name, entry)
