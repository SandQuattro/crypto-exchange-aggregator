import React, {useEffect, useState} from 'react';

const Timer = ({ nextUpdate, connected }) => {
    const [timeLeft, setTimeLeft] = useState(0);

    useEffect(() => {
        const interval = setInterval(() => {
            if (nextUpdate) {
                const now = new Date().getTime();
                const target = new Date(nextUpdate).getTime();
                const difference = target - now;

                if (difference > 0) {
                    setTimeLeft(Math.ceil(difference / 1000));
                } else {
                    setTimeLeft(0);
                }
            }
        }, 1000);

        return () => clearInterval(interval);
    }, [nextUpdate]);

    const minutes = Math.floor(timeLeft / 60);
    const seconds = timeLeft % 60;

    return (
        <div className={`timer ${connected ? 'timer-connected' : 'timer-disconnected'}`}>
            <div className="timer-text">
                {connected ? (
                    `🟢 Next update in: ${minutes}:${seconds.toString().padStart(2, '0')}`
                ) : (
                    '🔴 Connection lost'
                )}
            </div>
        </div>
    );
};

export default Timer; 