import { useState, useRef } from 'react'
import FlipDigit from './FlipDigit'
import './ScoreBoard.css'

function EditableName({ name, onChange, className }) {
  const [editing, setEditing] = useState(false)
  const [tmp, setTmp] = useState(name)
  const inputRef = useRef(null)

  const startEdit = () => {
    setTmp(name)
    setEditing(true)
    setTimeout(() => inputRef.current?.select(), 0)
  }

  const done = () => {
    const val = tmp.trim()
    if (val) {
      onChange(val)
    } else {
      setTmp(name)
    }
    setEditing(false)
  }

  const onKeyDown = (e) => {
    if (e.key === 'Enter') done()
    if (e.key === 'Escape') { setTmp(name); setEditing(false) }
  }

  if (editing) {
    return (
      <input
        ref={inputRef}
        className={`team-name-input ${className}`}
        value={tmp}
        onChange={e => setTmp(e.target.value)}
        onBlur={done}
        onKeyDown={onKeyDown}
        maxLength={8}
        autoFocus
      />
    )
  }

  return (
    <div className={`team-name ${className}`} onClick={startEdit} title="点击编辑队名">
      <span className="team-name-text">{name}</span>
      <span className="team-name-edit-icon">✎</span>
    </div>
  )
}

export default function ScoreBoard({
  teamA, teamB, nameA, nameB,
  onNameAChange, onNameBChange,
  onIncrementA, onIncrementB,
  onDecrementA, onDecrementB,
  onReset
}) {
  const digitsA = teamA.toString().padStart(3, '0').split('').map(Number)
  const digitsB = teamB.toString().padStart(3, '0').split('').map(Number)

  return (
    <div className="scoreboard">
      <div className="scoreboard-title">
        <span className="title-text">SCORE BOARD</span>
        <span className="title-deco">⚡</span>
      </div>

      <div className="scoreboard-body">
        <div className="team-section">
          <EditableName name={nameA} onChange={onNameAChange} className="team-a-name" />
          <div className="digit-group">
            {digitsA.map((d, i) => (
              <FlipDigit key={i} digit={d} />
            ))}
          </div>
          <div className="team-controls">
            <button className="ctrl-btn ctrl-plus" onClick={() => onIncrementA(1)}>+1</button>
            <button className="ctrl-btn ctrl-plus ctrl-plus-3" onClick={() => onIncrementA(3)}>+3</button>
            <button className="ctrl-btn ctrl-minus" onClick={() => onDecrementA()}>-1</button>
          </div>
        </div>

        <div className="vs-divider">
          <span className="vs-text">VS</span>
        </div>

        <div className="team-section">
          <EditableName name={nameB} onChange={onNameBChange} className="team-b-name" />
          <div className="digit-group">
            {digitsB.map((d, i) => (
              <FlipDigit key={i} digit={d} />
            ))}
          </div>
          <div className="team-controls">
            <button className="ctrl-btn ctrl-plus" onClick={() => onIncrementB(1)}>+1</button>
            <button className="ctrl-btn ctrl-plus ctrl-plus-3" onClick={() => onIncrementB(3)}>+3</button>
            <button className="ctrl-btn ctrl-minus" onClick={() => onDecrementB()}>-1</button>
          </div>
        </div>
      </div>

      <div className="scoreboard-footer">
        <button className="ctrl-btn ctrl-reset" onClick={onReset}>🔄 重置</button>
      </div>
    </div>
  )
}
