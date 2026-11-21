#!/usr/bin/env bash
# VTC P0 安全加固一键部署
# 用法：sudo bash deploy/apply-p0.sh
# 作用：安装限流/协议映射、安全响应头片段、站点配置，校验通过后重载 nginx 并重启后端。
set -euo pipefail

REPO="${REPO:-/home/pi/projects/Frontend/victor-tool-collection}"
NGINX_DIR="$REPO/deploy/nginx"
STAMP="$(date +%Y%m%d-%H%M%S)"

if [ "$(id -u)" -ne 0 ]; then
  echo "请用 sudo 运行：sudo bash $0" >&2
  exit 1
fi

for f in "$NGINX_DIR/00-vtc-limits.conf" "$NGINX_DIR/vtc-security-headers.conf" "$NGINX_DIR/port-8001.conf"; do
  [ -f "$f" ] || { echo "缺少文件：$f" >&2; exit 1; }
done

echo "[1/5] 安装 http 上下文限流与协议映射"
install -m 644 "$NGINX_DIR/00-vtc-limits.conf" /etc/nginx/conf.d/00-vtc-limits.conf

echo "[2/5] 安装安全响应头片段"
install -d /etc/nginx/snippets
install -m 644 "$NGINX_DIR/vtc-security-headers.conf" /etc/nginx/snippets/vtc-security-headers.conf

echo "[3/5] 备份并安装站点配置"
if [ -f /etc/nginx/sites-available/port-8001 ]; then
  cp -a /etc/nginx/sites-available/port-8001 "/etc/nginx/sites-available/port-8001.bak-$STAMP"
  echo "    备份：/etc/nginx/sites-available/port-8001.bak-$STAMP"
fi
install -m 644 "$NGINX_DIR/port-8001.conf" /etc/nginx/sites-available/port-8001
ln -sfn /etc/nginx/sites-available/port-8001 /etc/nginx/sites-enabled/port-8001

echo "[4/5] 校验并重载 nginx"
/usr/sbin/nginx -t
systemctl reload nginx

echo "[5/5] 重启后端（启动时会把历史明文密码升级为 bcrypt）"
systemctl restart vtc-backend
sleep 2
systemctl is-active vtc-backend

echo
echo "完成。验证建议："
echo "  ss -lntp | grep 8003                 # 应只监听 127.0.0.1"
echo "  journalctl -u vtc-backend -n 20      # 应出现 'admin 密码已升级为 bcrypt 哈希'（首次）"
