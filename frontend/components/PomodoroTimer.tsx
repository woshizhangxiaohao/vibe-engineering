import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Timer } from 'lucide-react'

export default function PomodoroTimer() {
  const [timeLeft, setTimeLeft] = useState(25 * 60) // 25 minutes in seconds
  const [isRunning, setIsRunning] = useState(false)
  const [duration, setDuration] = useState(25) // default duration in minutes

  useEffect(() => {
    let timer: NodeJS.Timeout

    if (isRunning && timeLeft > 0) {
      timer = setInterval(() => {
        setTimeLeft((prev) => prev - 1)
      }, 1000)
    } else if (timeLeft === 0) {
      handleComplete()
    }

    return () => clearInterval(timer)
  }, [isRunning, timeLeft])

  const handleStart = () => {
    setIsRunning(true)
    // Record start time to backend
    fetch('/api/pomodoros', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        startTime: new Date(),
        duration: duration
      })
    })
  }

  const handleComplete = async () => {
    setIsRunning(false)
    // Record completion to backend
    await fetch('/api/pomodoros/complete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        endTime: new Date(),
        isCompleted: true
      })
    })
    
    if ('Notification' in window) {
      Notification.requestPermission().then(permission => {
        if (permission === 'granted') {
          new Notification('Pomodoro Complete!', {
            body: 'Time to take a break!'
          })
        }
      })
    }
  }

  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
  }

  return (
    <div className="flex flex-col items-center space-y-4">
      <div className="text-6xl font-bold">{formatTime(timeLeft)}</div>
      <Button
        onClick={() => isRunning ? setIsRunning(false) : handleStart()}
        className="flex items-center gap-2"
      >
        <Timer className="w-4 h-4" />
        {isRunning ? 'Pause' : 'Start'}
      </Button>
    </div>
  )
}