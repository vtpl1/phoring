import { useState } from 'react'

import './App.css'
import MetricsHistory from './components/MetricsHistory'

function App() {
  const [count, setCount] = useState(0)

  return (
    <MetricsHistory/>
  )
}

export default App
