import { useEffect, useRef, useState } from 'react'
import './FlipDigit.css'

export default function FlipDigit({ digit }) {
  const [current, setCurrent] = useState(digit)
  const [flipping, setFlipping] = useState(false)
  const oldRef = useRef(digit)
  const timerRef = useRef(null)
  const queueRef = useRef(null)

  // 监听 digit 变化 + 动画状态变化，处理排队
  useEffect(() => {
    if (flipping) {
      // 正在翻页，把最新的数字存到队列
      if (digit !== current) {
        queueRef.current = digit
      }
      return
    }

    // 不翻页 → 检查是否需要开始翻页
    const target = queueRef.current !== null ? queueRef.current : digit
    queueRef.current = null

    if (target === current) return

    // 开始翻页动画
    const oldVal = current
    oldRef.current = oldVal
    setFlipping(true)

    // 第一阶段：上半翻页卡带旧数字翻下去
    timerRef.current = setTimeout(() => {
      // 上半翻完，静态区更新为新数字
      setCurrent(target)

      // 第二阶段：下半翻页卡带新数字翻上来
      timerRef.current = setTimeout(() => {
        setFlipping(false)
        // 翻完会自动触发 useEffect，检查队列中是否有新数字
      }, 220)
    }, 220)
  })

  // 组件卸载时清理定时器
  useEffect(() => {
    return () => {
      clearTimeout(timerRef.current)
    }
  }, [])

  const display = current.toString().padStart(1, '0')
  const oldDisplay = oldRef.current

  return (
    <div className="flip-digit">
      <div className="flip-digit-inner">
        {/* 上半部分 */}
        <div className="dh dh-top">
          <span className="dt">{display}</span>
        </div>

        {/* 下半部分 */}
        <div className="dh dh-bottom">
          <span className="dt">{display}</span>
        </div>

        {/* 翻页动画 */}
        {flipping && (
          <div className="flip-anim-wrap">
            {/* 上半翻页卡 (旧数字翻下去) */}
            <div className="flip-card flip-card-top" key="top">
              <span className="dt">{oldDisplay}</span>
            </div>

            {/* 下半翻页卡 (新数字翻上来) */}
            <div className="flip-card flip-card-bottom" key="bottom">
              <span className="dt">{digit}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
