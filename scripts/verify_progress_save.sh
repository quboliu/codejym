#!/bin/bash

# 验证实时保存进度功能
# 检查API调用频率和数据库更新

echo "========================================="
echo "实时保存进度功能验证"
echo "========================================="
echo

# 测试参数
TEST_DURATION=15  # 测试时长（秒）
INTERVAL=1.2     # 预期调用间隔（秒）
EXPECTED_CALLS=$(echo "$TEST_DURATION / $INTERVAL" | bc | cut -d. -f1)
EXPECTED_CALLS=$((EXPECTED_CALLS + 2))  # 允许一些误差

echo "测试参数:"
echo "  预期调用间隔: ${INTERVAL}s"
echo "  测试时长: ${TEST_DURATION}s"
echo "  预期调用次数: 约 ${EXPECTED_CALLS} 次"
echo

# 检查当前会话数据
echo "当前数据库状态:"
docker exec codejym-postgres-1 psql -U codecopy -d codecopybook -t \
  -c "SELECT 'Session ID: ' || id || ', Cursor: ' || cursor || ', Errors: ' || errors || ', Duration: ' || duration_seconds || 's, Updated: ' || updated_at FROM typing_sessions ORDER BY updated_at DESC LIMIT 1;" 2>&1 | grep -v "could not translate"
echo

echo "等待 ${TEST_DURATION} 秒，监控 PATCH 调用..."
echo "预期: 每 ${INTERVAL} 秒调用一次 PATCH API"
echo "---"

# 记录初始的PATCH调用次数
INITIAL_COUNT=$(docker compose -f docker-compose.proxy.yml logs --tail=1000 codecopybook 2>/dev/null | grep -c "PATCH /api/sessions" || echo "0")

echo "初始PATCH调用次数: ${INITIAL_COUNT}"

# 等待并监控
for i in $(seq 1 $TEST_DURATION); do
  sleep 1
  CURRENT_COUNT=$(docker compose -f docker-compose.proxy.yml logs --tail=1000 codecopybook 2>/dev/null | grep -c "PATCH /api/sessions" || echo "0")
  DIFF=$((CURRENT_COUNT - INITIAL_COUNT))

  # 每5秒显示一次进度
  if [ $((i % 5)) -eq 0 ]; then
    echo "[${i}s] 已执行 ${DIFF} 次PATCH调用"
  fi
done

echo "---"
echo

# 最终统计
FINAL_COUNT=$(docker compose -f docker-compose.proxy.yml logs --tail=1000 codecopybook 2>/dev/null | grep -c "PATCH /api/sessions" || echo "0")
ACTUAL_CALLS=$((FINAL_COUNT - INITIAL_COUNT))

echo "========================================="
echo "测试结果"
echo "========================================="
echo "总PATCH调用次数: ${ACTUAL_CALLS}"
echo "预期调用次数: 约 ${EXPECTED_CALLS}"
echo

# 计算调用间隔
if [ $ACTUAL_CALLS -gt 1 ]; then
  AVG_INTERVAL=$(echo "scale=2; $TEST_DURATION / $ACTUAL_CALLS" | bc)
  echo "平均调用间隔: ${AVG_INTERVAL}s"
else
  echo "❌ 错误: PATCH调用次数过少"
  echo
  echo "可能原因:"
  echo "  1. 用户未进入练习模式"
  echo "  2. 前端JavaScript错误"
  echo "  3. session未正确创建"
  echo
  exit 1
fi

# 验证结果
echo
echo "验证:"
if (( $(echo "$AVG_INTERVAL < 2.0" | bc -l) )); then
  echo "  ✅ 调用间隔正常 (< 2秒)"
else
  echo "  ❌ 调用间隔异常 (>= 2秒)"
fi

if [ $ACTUAL_CALLS -ge $((EXPECTED_CALLS / 2)) ]; then
  echo "  ✅ 调用次数充足"
else
  echo "  ⚠️  调用次数较少，但可能是正常现象（用户未输入）"
fi

echo
echo "========================================="
echo "数据库验证"
echo "========================================="
docker exec codejym-postgres-1 psql -U codecopy -d codecopybook -c \
  "SELECT '最新会话: Cursor=' || cursor || ', Errors=' || errors || ', Duration=' || duration_seconds || 's, Updated=' || updated_at FROM typing_sessions ORDER BY updated_at DESC LIMIT 1;" 2>&1 | grep -v "could not translate"

echo
echo "========================================="
echo "✅ 验证完成"
echo "========================================="
echo
echo "提示: 如果调用间隔仍不正确，请检查:"
echo "  1. 浏览器控制台是否有JavaScript错误"
echo "  2. 是否在练习模式下输入了字符"
echo "  3. session是否正确创建"
echo