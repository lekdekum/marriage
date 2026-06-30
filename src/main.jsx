import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

const weddingDate = new Date('2026-11-28T09:00:00-03:00');

function getTimeLeft(targetDate) {
  const difference = targetDate.getTime() - Date.now();

  if (difference <= 0) {
    return {
      days: 0,
      hours: 0,
      minutes: 0,
      seconds: 0,
      isComplete: true,
    };
  }

  return {
    days: Math.floor(difference / (1000 * 60 * 60 * 24)),
    hours: Math.floor((difference / (1000 * 60 * 60)) % 24),
    minutes: Math.floor((difference / (1000 * 60)) % 60),
    seconds: Math.floor((difference / 1000) % 60),
    isComplete: false,
  };
}

function CountdownBlock({ label, value }) {
  return (
    <div className="countdown-block">
      <span className="countdown-value">{String(value).padStart(2, '0')}</span>
      <span className="countdown-label">{label}</span>
    </div>
  );
}

function App() {
  const targetLabel = useMemo(
    () =>
      new Intl.DateTimeFormat('en', {
        dateStyle: 'full',
        timeStyle: 'short',
        timeZone: 'America/Sao_Paulo',
      }).format(weddingDate),
    [],
  );

  const [timeLeft, setTimeLeft] = useState(() => getTimeLeft(weddingDate));

  useEffect(() => {
    const timer = window.setInterval(() => {
      setTimeLeft(getTimeLeft(weddingDate));
    }, 1000);

    return () => window.clearInterval(timer);
  }, []);

  return (
    <main className="page-shell">
      <section className="hero" aria-labelledby="page-title">
        <p className="eyebrow">Save the date</p>
        <h1 id="page-title">Save the Date</h1>
        <p className="date-line">{targetLabel} UTC-3</p>

        {timeLeft.isComplete ? (
          <p className="complete-message">Today is the day.</p>
        ) : (
          <div className="countdown" aria-label="Countdown to the wedding">
            <CountdownBlock label="Days" value={timeLeft.days} />
            <CountdownBlock label="Hours" value={timeLeft.hours} />
            <CountdownBlock label="Minutes" value={timeLeft.minutes} />
            <CountdownBlock label="Seconds" value={timeLeft.seconds} />
          </div>
        )}
      </section>
    </main>
  );
}

createRoot(document.getElementById('root')).render(<App />);
