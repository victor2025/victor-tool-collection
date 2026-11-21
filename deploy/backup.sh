#!/usr/bin/env bash
# VTC 每日备份：PostgreSQL 全库 + 乐谱图片
# 用法：bash deploy/backup.sh
# 默认输出到 ~/backups/vtc，保留最近 7 份；可用 OUT / KEEP 环境变量覆盖。
set -euo pipefail

REPO="${REPO:-/home/pi/projects/Frontend/victor-tool-collection}"
OUT="${OUT:-/home/pi/backups/vtc}"
KEEP="${KEEP:-7}"
STAMP="$(date +%Y%m%d-%H%M%S)"

mkdir -p "$OUT"

# 从 config.json 读取连接信息（口令不硬编码在脚本里，也不打印）
eval "$(python3 - "$REPO/tools/backend/config.json" <<'PY'
import json, re, shlex, sys
cfg = json.load(open(sys.argv[1], encoding='utf-8'))
kv = dict(re.findall(r'(\w+)=([^\s]+)', cfg.get('dsn', '')))
for key, var in (('host', 'DSN_HOST'), ('user', 'DSN_USER'), ('password', 'DSN_PASSWORD'), ('dbname', 'DSN_DBNAME')):
    if key in kv:
        print(f"{var}={shlex.quote(kv[key])}")
PY
)"

export PGPASSWORD="${DSN_PASSWORD:-}"

echo "[1/3] 导出数据库"
pg_dump -h "${DSN_HOST:-/var/run/postgresql}" -U "${DSN_USER:-vtc}" "${DSN_DBNAME:-vtc}" \
  -Fc -f "$OUT/vtc-db-$STAMP.dump"

# 立刻校验 dump 可被 pg_restore 解析，避免产出"看起来成功"的坏备份
if ! pg_restore -l "$OUT/vtc-db-$STAMP.dump" >/dev/null 2>&1; then
  echo "备份校验失败：$OUT/vtc-db-$STAMP.dump" >&2
  exit 1
fi

echo "[2/3] 打包乐谱文件"
if [ -d "$REPO/data/scores" ]; then
  tar czf "$OUT/vtc-scores-$STAMP.tar.gz" -C "$REPO/data" scores
else
  echo "  跳过：未找到 $REPO/data/scores"
fi

echo "[3/3] 清理超过 $KEEP 份的旧备份"
ls -1t "$OUT"/vtc-db-*.dump 2>/dev/null | tail -n +$((KEEP + 1)) | xargs -r rm -f
ls -1t "$OUT"/vtc-scores-*.tar.gz 2>/dev/null | tail -n +$((KEEP + 1)) | xargs -r rm -f

echo "完成：$(ls -lh "$OUT" | tail -n +2 | awk '{print $9" ("$5")"}' | paste -sd ' ' -)"
