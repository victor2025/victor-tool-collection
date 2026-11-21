#!/usr/bin/env bash
# VTC 一键部署：测试 → 编译后端 → 同步 nginx 配置 → 重启服务
# 用法：bash deploy/deploy.sh [--scoreboard]
#   --scoreboard  顺带重新构建记分板（React/Vite）
set -euo pipefail

REPO="${REPO:-/home/pi/projects/Frontend/victor-tool-collection}"
cd "$REPO"

echo "[1/5] 后端单元测试"
(cd tools/backend && GO111MODULE=on go test ./handlers)

echo "[2/5] 编译后端"
(cd tools/backend && GO111MODULE=on go build -o backend .)

echo "[3/5] 同步 nginx 片段"
changed=0
for pair in "deploy/nginx/00-vtc-limits.conf:/etc/nginx/conf.d/00-vtc-limits.conf" \
            "deploy/nginx/10-vtc-gzip.conf:/etc/nginx/conf.d/10-vtc-gzip.conf" \
            "deploy/nginx/vtc-security-headers.conf:/etc/nginx/snippets/vtc-security-headers.conf"; do
  src="${pair%%:*}"; dst="${pair##*:}"
  [ -f "$src" ] || { echo "  缺少 $src" >&2; exit 1; }
  if ! cmp -s "$src" "$dst" 2>/dev/null; then
    install -d "$(dirname "$dst")"
    sudo -n install -m 644 "$src" "$dst"
    echo "  更新 $dst"
    changed=1
  fi
done

echo "[4/5] 同步站点配置"
if ! cmp -s deploy/nginx/port-8001.conf /etc/nginx/sites-available/port-8001; then
  sudo -n cp -a /etc/nginx/sites-available/port-8001 "/etc/nginx/sites-available/port-8001.bak-$(date +%Y%m%d-%H%M%S)"
  sudo -n install -m 644 deploy/nginx/port-8001.conf /etc/nginx/sites-available/port-8001
  sudo -n ln -sfn /etc/nginx/sites-available/port-8001 /etc/nginx/sites-enabled/port-8001
  echo "  站点配置已更新"
  changed=1
fi

if [ "$changed" -eq 1 ]; then
  sudo -n /usr/sbin/nginx -t
  sudo -n systemctl reload nginx
  echo "  nginx 已重载"
else
  echo "  nginx 配置无变化"
fi

echo "[5/5] 重启后端"
sudo -n systemctl restart vtc-backend
sleep 2
systemctl is-active vtc-backend

if [ "${1:-}" = "--scoreboard" ]; then
  echo "[附加] 构建记分板"
  (cd tools/score-board && npm ci && npm run build)
fi

echo "完成。"
