#!/usr/bin/env python3
"""P2 第一层：可读性与排版微调
1) 只替换 color: 声明里的低对比灰（背景/边框不动），把 1.9:1 ~ 3.9:1 提到 5:1 以上
2) 中文标题/副标题去掉过宽字距（3px/2px -> 1px）
"""
import glob
import re
import sys

FILES = ['nav/index.html'] + sorted(glob.glob('tools/*/dist/index.html'))
TOOL_FILES = sorted(glob.glob('tools/*/dist/index.html'))

# 旧值 -> 新值（在 #0a0a0a 背景上的对比度）
#   #333 1.9:1 -> #808080 5.0:1
#   #444 2.6:1 -> #8a8a8a 5.5:1
#   #555 3.4:1 -> #949494 5.9:1
#   #666 3.9:1 -> #9e9e9e 6.2:1
COLOR_MAP = {'333': '808080', '444': '8a8a8a', '555': '949494', '666': '9e9e9e'}

# 负向断言：排除 border-color / background-color / -webkit-text-fill-color
color_re = re.compile(r'(?<![A-Za-z-])color\s*:\s*#(333|444|555|666)\b', re.I)


def fix_colors(text):
    return color_re.sub(lambda m: 'color: #' + COLOR_MAP[m.group(1).lower()], text), len(color_re.findall(text))


def fix_tracking(text):
    n = 0
    text, k = re.subn(r'(h1\s*\{[^}]*?)letter-spacing\s*:\s*3px', r'\1letter-spacing: 1px', text, flags=re.S)
    n += k
    text, k = re.subn(r'((?:\.header\s+)?\.sub\s*\{[^}]*?)letter-spacing\s*:\s*2px', r'\1letter-spacing: 1px', text, flags=re.S)
    n += k
    return text, n


total_c = total_t = 0
for f in FILES:
    src = open(f, encoding='utf-8').read()
    out, c = fix_colors(src)
    t = 0
    if f in TOOL_FILES:
        out, t = fix_tracking(out)
    if out != src:
        open(f, 'w', encoding='utf-8').write(out)
    total_c += c
    total_t += t
    print(f'{f}: 颜色 {c} 处, 字距 {t} 处')

print(f'合计：颜色 {total_c} 处，字距 {total_t} 处')
