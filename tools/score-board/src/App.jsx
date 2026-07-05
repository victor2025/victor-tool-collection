import { useState, useEffect, useCallback } from 'react'
import ScoreBoard from './components/ScoreBoard'
import './App.css'

const LS_KEY = 'scoreboard_state'

function loadState() {
  try {
    const raw = localStorage.getItem(LS_KEY)
    if (raw) return JSON.parse(raw)
  } catch {}
  return null
}

export default function App() {
  const saved = loadState()

  const [teamA, setTeamA] = useState(saved?.teamA ?? 0)
  const [teamB, setTeamB] = useState(saved?.teamB ?? 0)
  const [nameA, setNameA] = useState(saved?.nameA ?? '主队')
  const [nameB, setNameB] = useState(saved?.nameB ?? '客队')

  // 变化时自动存到 localStorage
  const save = useCallback((a, b, nA, nB) => {
    localStorage.setItem(LS_KEY, JSON.stringify({
      teamA: a, teamB: b, nameA: nA, nameB: nB
    }))
  }, [])

  // 每次状态变化都保存
  useEffect(() => { save(teamA, teamB, nameA, nameB) }, [teamA, teamB, nameA, nameB])

  const incrementA = (n = 1) => setTeamA(prev => Math.min(prev + n, 999))
  const incrementB = (n = 1) => setTeamB(prev => Math.min(prev + n, 999))
  const decrementA = () => setTeamA(prev => Math.max(prev - 1, 0))
  const decrementB = () => setTeamB(prev => Math.max(prev - 1, 0))

  const reset = () => {
    setTeamA(0)
    setTeamB(0)
  }

  return (
    <div className="app">
      <ScoreBoard
        teamA={teamA}
        teamB={teamB}
        nameA={nameA}
        nameB={nameB}
        onNameAChange={setNameA}
        onNameBChange={setNameB}
        onIncrementA={incrementA}
        onIncrementB={incrementB}
        onDecrementA={decrementA}
        onDecrementB={decrementB}
        onReset={reset}
      />
    </div>
  )
}
