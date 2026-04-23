import { useEffect, useState } from 'react'

function App() {
  const [msg, setMsg] = useState("Loading...")

  useEffect(() => {
    fetch("http://localhost:8080/api/hello")
      .then(res => res.json())
      .then(data => setMsg(data.message))
      .catch(() => setMsg("Error connecting to API"))
  }, [])

  return (
    <div>
      <h1>University Assignment App</h1>
      <p>Backend says: {msg}</p>
    </div>
  )
}

export default App
