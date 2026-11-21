#!/usr/bin/env python3
"""
把老工具页从「硬编码深色 + theme-light.css 整页反相」迁移到 /shared/vtc.css 变量色板。

只做三件事，不改任何 JS / HTML 结构：
  1. <head> 里的 <style> 之前插入 /shared/vtc.css 链接，删掉 /theme-light.css 链接
  2. <body ...> 加 vtc-native 类（theme-light.css 的 invert 兜底会自动跳过）
  3. 只在 <style> 块内部，把已知的深色硬编码色值换成 var(--vtc-*)

用法：
  python3 shared/migrate_theme.py tools/base64/dist/index.html [...]     实跑
  python3 shared/migrate_theme.py --dry-run tools/base64/...             只看残留报告
"""

import re
import sys

SHARED_LINK = '  <link rel="stylesheet" href="/shared/vtc.css?v=1"/>\n'

# ---------- 颜色 → 变量映射（深色主题的原始色值） ----------
COLOR_MAP = {
    # 背景：最深的一层
    '#000': 'var(--vtc-bg)', '#000000': 'var(--vtc-bg)',
    '#080808': 'var(--vtc-bg)', '#0a0a0a': 'var(--vtc-bg)',
    '#0b0b0b': 'var(--vtc-bg)', '#0d0d0d': 'var(--vtc-bg)',
    '#0f0f0f': 'var(--vtc-bg)', '#101010': 'var(--vtc-bg)',
    '#111': 'var(--vtc-bg-elev)', '#111111': 'var(--vtc-bg-elev)',
    '#141414': 'var(--vtc-bg-elev)', '#151515': 'var(--vtc-bg-elev)',
    '#161616': 'var(--vtc-bg-elev)', '#1a1a1a': 'var(--vtc-bg-elev)',
    '#1b1f24': 'var(--vtc-bg-elev)', '#1c1c1c': 'var(--vtc-bg-elev)',
    '#1d1d1d': 'var(--vtc-bg-elev)', '#1e1e2e': 'var(--vtc-bg-elev)',
    '#24292f': 'var(--vtc-bg-elev)', '#1f2328': 'var(--vtc-bg-elev)',
    '#111418': 'var(--vtc-bg-elev)',
    # 背景：面板 / 抬升层
    '#222': 'var(--vtc-bg-elev-2)', '#222222': 'var(--vtc-bg-elev-2)',
    '#2a2a2a': 'var(--vtc-bg-elev-2)', '#2b2b2b': 'var(--vtc-bg-elev-2)',
    '#0d0f0f': 'var(--vtc-bg-elev-2)', '#0d1212': 'var(--vtc-bg-elev-2)',
    '#101a18': 'var(--vtc-bg-elev-2)', '#16221f': 'var(--vtc-bg-elev-2)',
    '#1a2626': 'var(--vtc-bg-elev-2)', '#1a282a': 'var(--vtc-bg-elev-2)',
    '#1a2a28': 'var(--vtc-bg-elev-2)', '#1a3a36': 'var(--vtc-bg-elev-2)',
    '#142a26': 'var(--vtc-bg-elev-2)', '#142b29': 'var(--vtc-bg-elev-2)',
    '#102b28': 'var(--vtc-bg-elev-2)',
    '#173f3a': 'var(--vtc-bg-elev-2)', '#1d3b38': 'var(--vtc-bg-elev-2)',
    '#24403b': 'var(--vtc-bg-elev-2)', '#23514d': 'var(--vtc-bg-elev-2)',
    '#0d1a18': 'var(--vtc-bg-elev-2)', '#241414': 'var(--vtc-bg-elev-2)',
    '#221010': 'var(--vtc-bg-elev-2)', '#2b1010': 'var(--vtc-bg-elev-2)',
    '#3f1717': 'var(--vtc-bg-elev-2)', '#4a2020': 'var(--vtc-bg-elev-2)',
    '#4a2323': 'var(--vtc-bg-elev-2)', '#662222': 'var(--vtc-bg-elev-2)',
    '#1a1010': 'var(--vtc-bg-elev-2)', '#2a1818': 'var(--vtc-bg-elev-2)',
    '#1a1a10': 'var(--vtc-bg-elev-2)', '#1a1a0a': 'var(--vtc-bg-elev-2)',
    '#2a281a': 'var(--vtc-bg-elev-2)', '#2a2820': 'var(--vtc-bg-elev-2)',
    '#3a3820': 'var(--vtc-bg-elev-2)', '#1e3228': 'var(--vtc-bg-elev-2)',
    '#2a3a38': 'var(--vtc-bg-elev-2)', '#24292e': 'var(--vtc-bg-elev-2)',
    # 边框
    '#333': 'var(--vtc-border)', '#333333': 'var(--vtc-border)',
    # 文字
    '#e8e8e8': 'var(--vtc-fg)', '#e0e0e0': 'var(--vtc-fg)',
    '#e6e9ee': 'var(--vtc-fg)', '#fff': 'var(--vtc-fg)', '#ffffff': 'var(--vtc-fg)',
    '#eceff3': 'var(--vtc-fg)', '#eef1f4': 'var(--vtc-fg)', '#f0f2f5': 'var(--vtc-fg)',
    '#ccc': 'var(--vtc-fg-2)', '#cccccc': 'var(--vtc-fg-2)',
    '#ddd': 'var(--vtc-fg-2)', '#dddddd': 'var(--vtc-fg-2)',
    '#d4d4d4': 'var(--vtc-fg-2)', '#b5b5b5': 'var(--vtc-fg-2)',
    '#aaa': 'var(--vtc-fg-3)', '#aaaaaa': 'var(--vtc-fg-3)',
    '#bbb': 'var(--vtc-fg-3)', '#9aa3af': 'var(--vtc-fg-3)',
    '#949494': 'var(--vtc-fg-3)', '#9e9e9e': 'var(--vtc-fg-3)',
    '#888': 'var(--vtc-fg-3)', '#888888': 'var(--vtc-fg-3)',
    '#8a8a8a': 'var(--vtc-fg-3)', '#8b93a1': 'var(--vtc-fg-4)',
    '#777': 'var(--vtc-fg-4)', '#808080': 'var(--vtc-fg-4)',
    '#6b7280': 'var(--vtc-fg-4)', '#7b8494': 'var(--vtc-fg-4)',
    '#a8b0bb': 'var(--vtc-fg-4)', '#b3bac4': 'var(--vtc-fg-4)',
    '#555': 'var(--vtc-fg-4)', '#666': 'var(--vtc-fg-4)',
    '#444': 'var(--vtc-fg-4)',
    # 主色
    '#4ecdc4': 'var(--vtc-accent)', '#00cec9': 'var(--vtc-accent)',
    '#0f9d94': 'var(--vtc-accent)', '#45b8b0': 'var(--vtc-accent)',
    '#2aa79e': 'var(--vtc-accent)', '#0b7a73': 'var(--vtc-accent)',
    '#7fe3da': 'var(--vtc-accent)', '#7fe0d8': 'var(--vtc-accent)',
    '#8fe8df': 'var(--vtc-accent)', '#66ccee': 'var(--vtc-accent)',
    '#a8ddd7': 'var(--vtc-accent)', '#b9e6e2': 'var(--vtc-accent)',
    '#d8f0ed': 'var(--vtc-accent)', '#e6f7f5': 'var(--vtc-accent)',
    '#e6f7f4': 'var(--vtc-accent)',
    # 警告 / 危险 / 成功
    '#ffd93d': 'var(--vtc-warn)', '#fdcb6e': 'var(--vtc-warn)',
    '#f0b400': 'var(--vtc-warn)', '#e0a800': 'var(--vtc-warn)',
    '#ff6b6b': 'var(--vtc-danger)', '#ef5350': 'var(--vtc-danger)',
    '#e04040': 'var(--vtc-danger)', '#d64545': 'var(--vtc-danger)',
    '#e07a7a': 'var(--vtc-danger)', '#ff9d9d': 'var(--vtc-danger)',
    '#ff9a9a': 'var(--vtc-danger)', '#b52d2d': 'var(--vtc-danger)',
    '#dc2626': 'var(--vtc-danger)', '#e17055': 'var(--vtc-danger)',
    '#ffab91': 'var(--vtc-danger)', '#f0c9c9': 'var(--vtc-danger)',
    '#66bb6a': 'var(--vtc-ok)', '#a5d6a7': 'var(--vtc-ok)',
    '#00b894': 'var(--vtc-ok)',
    # 语法高亮
    '#569cd6': 'var(--vtc-syn-key)', '#ce9178': 'var(--vtc-syn-str)',
    '#b5cea8': 'var(--vtc-syn-num)', '#dcdcaa': 'var(--vtc-syn-fn)',
    '#6a9955': 'var(--vtc-syn-comment)',
    '#9cdcfe': 'var(--vtc-syn-key)', '#7ec8e3': 'var(--vtc-syn-key)',
    '#ce93d8': 'var(--vtc-syn-alt)', '#c9a0dc': 'var(--vtc-syn-alt)',
    '#22665e': 'var(--vtc-accent)', '#8a7a3a': 'var(--vtc-warn)',
    '#4ecdc444': 'var(--vtc-accent-border)',
    '#4ecdc411': 'var(--vtc-accent-weak)',
}

# rgba() 形式的固定色（深色底上的半透明层）
RGBA_MAP = {
    'rgba(255,255,255,0.015)': 'var(--vtc-bg-elev-2)',
    'rgba(255,255,255,.015)': 'var(--vtc-bg-elev-2)',
    'rgba(78,205,196,0.04)': 'var(--vtc-accent-weak)',
    'rgba(78, 205, 196, 0.04)': 'var(--vtc-accent-weak)',
    'rgba(78, 205, 196, 0.1)': 'var(--vtc-accent-weak)',
    'rgba(78,205,196,0.2)': 'var(--vtc-accent-weak)',
    'rgba(78, 205, 196, 0.28)': 'var(--vtc-accent-weak)',
    'rgba(78,205,196,0.3)': 'var(--vtc-accent-weak)',
    'rgba(78,205,196,0.6)': 'var(--vtc-accent)',
    'rgba(255,217,61,0.1)': 'var(--vtc-accent-weak)',
    'rgba(255, 217, 61, 0.1)': 'var(--vtc-accent-weak)',
    'rgba(255,217,61,0.18)': 'var(--vtc-accent-weak)',
    'rgba(255, 217, 61, 0.18)': 'var(--vtc-accent-weak)',
    'rgba(255,217,61,0.2)': 'var(--vtc-accent-weak)',
    'rgba(255, 217, 61, 0.35)': 'var(--vtc-accent-weak)',
}

HEX_RE = re.compile(r'#[0-9a-fA-F]{3,8}\b')
STYLE_RE = re.compile(r'(<style>)(.*?)(</style>)', re.S)
BODY_RE = re.compile(r'<body(\s*[^>]*)>')
LIGHT_LINK_RE = re.compile(r'\n?\s*<link rel="stylesheet" href="/theme-light\.css[^"]*"\s*/?>')


def migrate(path, dry=False):
    src = open(path, encoding='utf-8').read()
    orig = src

    # ---- 1. head：加共享样式，去 theme-light ----
    src, n_light = LIGHT_LINK_RE.subn('', src)
    if '<link rel="stylesheet" href="/shared/vtc.css' not in src:
        m_indent = re.search(r'\n([ \t]*)<style>', src)
        indent = m_indent.group(1) if m_indent else '  '
        src = re.sub(r'\n([ \t]*)<style>', lambda m: f'\n{m.group(1)}<link rel="stylesheet" href="/shared/vtc.css?v=1"/>\n{m.group(1)}<style>',
                     src, count=1)

    # ---- 2. body 加 vtc-native ----
    def _body(m):
        attrs = m.group(1)
        if 'class=' in attrs:
            return re.sub(r'class="([^"]*)"', lambda c: f'class="vtc-native {c.group(1)}"', m.group(0))
        return f'<body{attrs} class="vtc-native">'
    src, n_body = BODY_RE.subn(_body, src, count=1)

    # ---- 3. 只在 <style> 内换色 ----
    replaced = []

    def _style(m):
        head, body_css, tail = m.group(1), m.group(2), m.group(3)

        def _hex(mm):
            key = mm.group(0).lower()
            if key in COLOR_MAP:
                replaced.append(key)
                return COLOR_MAP[key]
            return mm.group(0)

        css = HEX_RE.sub(_hex, body_css)
        for k, v in RGBA_MAP.items():
            if k in css:
                css = css.replace(k, v)
                replaced.append(k)
        return head + css + tail

    src = STYLE_RE.sub(_style, src, count=1)

    # ---- 残留报告 ----
    style = STYLE_RE.search(src)
    residual = sorted(set(m.group(0).lower() for m in HEX_RE.finditer(style.group(2)))) if style else []

    changed = src != orig
    if not dry and changed:
        open(path, 'w', encoding='utf-8').write(src)

    return {
        'path': path,
        'theme_light_removed': n_light,
        'body_class': n_body,
        'colors': len(replaced),
        'residual': residual,
        'changed': changed,
    }


def main():
    args = [a for a in sys.argv[1:] if not a.startswith('--')]
    dry = '--dry-run' in sys.argv
    bad = 0
    for path in args:
        r = migrate(path, dry=dry)
        flag = '⚠ ' if r['residual'] else '✓ '
        print(f"{flag}{r['path'].split('/')[1]:<18} 换色{r['colors']:>4}  "
              f"去theme-light={r['theme_light_removed']}  body={r['body_class']}  "
              f"残留={len(r['residual'])}")
        if r['residual']:
            bad += 1
            print('     ', ' '.join(r['residual']))
    print(f"\n{'[dry-run] ' if dry else ''}有残留的页面：{bad}/{len(args)}")


if __name__ == '__main__':
    main()
