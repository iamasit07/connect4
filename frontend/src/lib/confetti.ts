import confetti from 'canvas-confetti';
import { CONFETTI_DURATION } from './config';

export const fireWinConfetti = () => {
  const duration = CONFETTI_DURATION;
  const animationEnd = Date.now() + duration;
  
  const colors = ['#ef4444', '#eab308', '#3b82f6', '#22c55e', '#a855f7'];

  const randomInRange = (min: number, max: number) => {
    return Math.random() * (max - min) + min;
  };

  const interval = setInterval(() => {
    const timeLeft = animationEnd - Date.now();

    if (timeLeft <= 0) {
      clearInterval(interval);
      return;
    }

    const particleCount = 20 * (timeLeft / duration);

    // Left side burst
    confetti({
      particleCount,
      startVelocity: 30,
      spread: 360,
      origin: {
        x: randomInRange(0.1, 0.3),
        y: Math.random() - 0.2,
      },
      colors,
    });

    // Right side burst
    confetti({
      particleCount,
      startVelocity: 30,
      spread: 360,
      origin: {
        x: randomInRange(0.7, 0.9),
        y: Math.random() - 0.2,
      },
      colors,
    });
  }, 250);
};

export const fireDrawConfetti = () => {
  confetti({
    particleCount: 100,
    spread: 70,
    origin: { y: 0.6 },
    colors: ['#6b7280', '#9ca3af', '#d1d5db'],
  });
};
