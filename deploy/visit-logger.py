#!/usr/bin/env python3
"""
Visit Logger + Admin Auth — 记录访问 + 服务端密码管理
Runs on port 8002, called via nginx proxy.
"""

import hashlib
import json
import os
import secrets
import sqlite3
import sys
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DB_PATH = os.path.join(BASE_DIR, 'data', 'visits.db')
AUTH_PATH = os.path.join(BASE_DIR, 'data', 'admin.json')

# 默认密码
DEFAULT_PASSWORD = 'mima123123'

# 会话存储（内存）：token → expiry_timestamp
_sessions = {}

TRACK_JS = """(function(){
  if (window.__vtc_logged) return;
  window.__vtc_logged = true;
  var tool = window.location.pathname.replace(/\\/$/, '').split('/').pop() || 'home';
  var data = JSON.stringify({ tool: tool });
  if (navigator.sendBeacon) {
    navigator.sendBeacon('/_log/log/visit', data);
  } else {
    fetch('/_log/log/visit', { method: 'POST', body: data, keepalive: true });
  }
})();
"""


# ─── 密码管理 ─────────────────────────────

def _load_password():
    """从 admin.json 读取密码，不存在则创建默认"""
    if not os.path.exists(AUTH_PATH):
        _save_password(DEFAULT_PASSWORD)
        return DEFAULT_PASSWORD
    try:
        with open(AUTH_PATH, 'r') as f:
            data = json.load(f)
            return data.get('password', DEFAULT_PASSWORD)
    except Exception:
        return DEFAULT_PASSWORD


def _save_password(pwd):
    os.makedirs(os.path.dirname(AUTH_PATH), exist_ok=True)
    with open(AUTH_PATH, 'w') as f:
        json.dump({'password': pwd}, f, ensure_ascii=False)


def _create_session():
    token = secrets.token_hex(32)
    _sessions[token] = time.time() + 86400  # 24小时过期
    return token


def _check_session(token):
    """验证 token 并清理过期会话"""
    now = time.time()
    expired = []
    for t, exp in _sessions.items():
        if now > exp:
            expired.append(t)
    for t in expired:
        _sessions.pop(t, None)
    return token in _sessions


def _cleanup_sessions():
    now = time.time()
    for t in list(_sessions.keys()):
        if now > _sessions[t]:
            _sessions.pop(t, None)


# ─── HTTP Handler ──────────────────────────

class LogHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        pass

    def _send_json(self, code, data):
        self.send_response(code)
        self.send_header('Content-Type', 'application/json; charset=utf-8')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(data, ensure_ascii=False).encode('utf-8'))

    def _send_js(self, code, js):
        self.send_response(code)
        self.send_header('Content-Type', 'application/javascript; charset=utf-8')
        self.send_header('Cache-Control', 'no-cache, must-revalidate')
        self.end_headers()
        self.wfile.write(js.encode('utf-8'))

    def _get_client_ip(self):
        for h in ['X-Real-IP', 'X-Forwarded-For', 'X-Client-IP']:
            val = self.headers.get(h)
            if val:
                return val.split(',')[0].strip()
        return self.client_address[0]

    def _read_body(self):
        content_len = int(self.headers.get('Content-Length', 0))
        raw = self.rfile.read(content_len) if content_len else b'{}'
        try:
            return json.loads(raw)
        except json.JSONDecodeError:
            return {}

    def _log_visit(self, ip, tool):
        try:
            conn = sqlite3.connect(DB_PATH)
            conn.execute('INSERT INTO visits (ip, tool) VALUES (?, ?)', (ip, tool))
            conn.commit()
            conn.close()
        except Exception as e:
            print(f"[logger] DB error: {e}", file=sys.stderr)

    # ─── GET routes ────────────────────────

    def do_GET(self):
        parsed = urlparse(self.path)
        path = parsed.path.rstrip('/') or '/'

        if path == '/track.js':
            self._send_js(200, TRACK_JS)
        elif path == '/stats':
            self._handle_stats(parsed)
        else:
            self._send_json(404, {'error': 'not found'})

    # ─── POST routes ────────────────────────

    def do_POST(self):
        parsed = urlparse(self.path)
        path = parsed.path.rstrip('/')

        if path == '/log/visit' or path == '/visit':
            data = self._read_body()
            ip = self._get_client_ip()
            tool = data.get('tool', 'unknown')
            self._log_visit(ip, tool)
            self._send_json(200, {'ok': True})

        elif path == '/api/login':
            data = self._read_body()
            pwd = data.get('password', '')
            if pwd == _load_password():
                token = _create_session()
                self._send_json(200, {'ok': True, 'token': token})
            else:
                self._send_json(200, {'ok': False, 'error': '密码错误'})

        elif path == '/api/change-password':
            data = self._read_body()
            token = data.get('token', '')
            if not _check_session(token):
                self._send_json(200, {'ok': False, 'error': '未登录'})
                return
            old = data.get('old_password', '')
            nu = data.get('new_password', '')
            if old != _load_password():
                self._send_json(200, {'ok': False, 'error': '当前密码错误'})
                return
            if len(nu) < 4:
                self._send_json(200, {'ok': False, 'error': '新密码至少4位'})
                return
            _save_password(nu)
            self._send_json(200, {'ok': True})

        elif path == '/api/check-session':
            data = self._read_body()
            valid = _check_session(data.get('token', ''))
            self._send_json(200, {'ok': valid})

        else:
            self._send_json(404, {'error': 'not found'})

    # ─── Stats ──────────────────────────────

    def _handle_stats(self, parsed):
        # 可选 token 验证（已登录用户查看）
        # 不强制，方便 curl 调试
        query = parse_qs(parsed.query)
        days = int(query.get('days', ['7'])[0])
        detail = query.get('detail', ['0'])[0] == '1'

        conn = sqlite3.connect(DB_PATH)
        conn.row_factory = sqlite3.Row
        c = conn.cursor()

        c.execute('SELECT COUNT(*) as total FROM visits')
        total = c.fetchone()['total']

        c.execute('SELECT COUNT(*) as cnt FROM visits WHERE visited_at >= datetime("now", ?)',
                  (f'-{days} days',))
        recent = c.fetchone()['cnt']

        c.execute('''SELECT tool, COUNT(*) as cnt FROM visits
                     WHERE visited_at >= datetime("now", ?)
                     GROUP BY tool ORDER BY cnt DESC''',
                  (f'-{days} days',))
        tool_stats = [dict(row) for row in c.fetchall()]

        c.execute('SELECT COUNT(DISTINCT ip) as cnt FROM visits WHERE visited_at >= datetime("now", ?)',
                  (f'-{days} days',))
        unique_ips = c.fetchone()['cnt']

        c.execute('''SELECT strftime("%H", visited_at) as hour, COUNT(*) as cnt
                     FROM visits WHERE visited_at >= datetime("now", ?)
                     GROUP BY hour ORDER BY hour''',
                  (f'-{days} days',))
        hourly = dict((row['hour'], row['cnt']) for row in c.fetchall())

        result = {
            'total': total,
            f'recent_{days}d': recent,
            'unique_ips': unique_ips,
            'tool_stats': tool_stats,
            'hourly': hourly,
        }

        if detail:
            ip_detail = {}
            for ts in tool_stats:
                tool = ts['tool']
                c.execute('''SELECT ip, COUNT(*) as cnt, MAX(visited_at) as last_visit
                             FROM visits WHERE tool = ? AND visited_at >= datetime("now", ?)
                             GROUP BY ip ORDER BY cnt DESC LIMIT 20''',
                          (tool, f'-{days} days'))
                ip_detail[tool] = [dict(row) for row in c.fetchall()]
            result['ip_detail'] = ip_detail

        conn.close()
        self._send_json(200, result)

    def do_OPTIONS(self):
        self.send_response(204)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        self.end_headers()


def run(port=8002):
    # 确保默认密码文件存在
    _load_password()
    server = HTTPServer(('127.0.0.1', port), LogHandler)
    print(f"[visit-logger] Listening on http://127.0.0.1:{port}", flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n[visit-logger] Shutting down...", flush=True)
        server.server_close()


if __name__ == '__main__':
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8002
    run(port)
