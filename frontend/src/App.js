import React, {useCallback, useEffect, useState} from 'react';
import './App.css';
import CryptoIcon from './components/CryptoIcon';
import Timer from './components/Timer';

const App = () => {
    const [rates, setRates] = useState({});
    const [previousRates, setPreviousRates] = useState({});
    const [connected, setConnected] = useState(false);
    const [lastUpdate, setLastUpdate] = useState(null);
    const [nextUpdate, setNextUpdate] = useState(null);

    // Функция обновления курсов
    const updateRates = useCallback((newRates, timestamp, nextUpd) => {
        setRates(prevRates => {
            // console.log('Previous rates:', prevRates);
            // console.log('New rates:', newRates);
            setPreviousRates({ ...prevRates });
            return newRates;
        });
        setLastUpdate(new Date(timestamp));
        setNextUpdate(nextUpd);
    }, []);

    useEffect(() => {
        const eventSource = new EventSource('/api/rates/stream');

        eventSource.onopen = () => {
            setConnected(true);
            console.log('SSE connected via Fiber');
        };

        eventSource.onmessage = (event) => {
            // Обрабатываем события без указанного типа (legacy)
            try {
                const data = JSON.parse(event.data);
                if (data.rates) {
                    updateRates(data.rates, data.timestamp, data.next_update);
                }
            } catch (error) {
                console.error('Error parsing SSE data:', error);
            }
        };

        // Обрабатываем события rates
        eventSource.addEventListener('rates', (event) => {
            try {
                const data = JSON.parse(event.data);
                updateRates(data.rates, data.timestamp, data.next_update);
            } catch (error) {
                console.error('Error parsing rates data:', error);
            }
        });

        // Обрабатываем heartbeat (можно игнорировать или логировать)
        eventSource.addEventListener('heartbeat', (event) => {
            // Heartbeat для поддержания соединения
            // console.log('Heartbeat received');
        });

        eventSource.onerror = () => {
            setConnected(false);
            console.log('SSE error');
        };

        return () => {
            eventSource.close();
        };
    }, [updateRates]);

    // Функция для расчета процента изменения цены
    const getPriceChangePercent = (currentPrice, previousPrice) => {
        if (!previousPrice || previousPrice === currentPrice) {
            return null;
        }

        const current = parseFloat(currentPrice);
        const previous = parseFloat(previousPrice);
        const percentChange = ((current - previous) / previous) * 100;
        const absPercent = Math.abs(percentChange);

        // console.log(`Price change: ${previous} -> ${current} (${percentChange.toFixed(4)}%)`);

        // Показываем только если изменение больше 0.01%
        if (absPercent < 0.01) {
            // console.log('Change too small, not displaying');
            return null;
        }

        return {
            percent: absPercent.toFixed(2),
            direction: percentChange > 0 ? 'up' : 'down',
            arrow: percentChange > 0 ? '↗' : '↘'
        };
    };

    const groupedRates = {};
    const symbolChanges = {}; // Для хранения процентов изменения по символам

    Object.entries(rates).forEach(([key, value]) => {
        const parts = key.split('_');
        const symbol = parts[0];
        const currency = parts[1];

        if (!groupedRates[symbol]) {
            groupedRates[symbol] = {};
        }

        // Добавляем цену (всегда синяя)
        groupedRates[symbol][currency] = {
            price: value
        };

        // Рассчитываем процент изменения только для USD (базовая валюта)
        if (currency === 'USD') {
            const previousPrice = previousRates[key];
            const changeData = getPriceChangePercent(value, previousPrice);
            if (changeData) {
                symbolChanges[symbol] = changeData;
            }
        }
    });



    if (Object.keys(rates).length === 0) {
        return (
            <div className="container">
                <div className="header">
                    <h1>
                        🚀 Crypto Exchange Aggregator
                    </h1>
                    <p>Real-time cryptocurrency rates</p>
                </div>
                <div className="loading">Loading rates...</div>
            </div>
        );
    }

    return (
        <div className="container">
            <div className="header">
                <h1>
                    🚀 Crypto Exchange Aggregator
                </h1>
                <p>Real-time cryptocurrency rates</p>
                {nextUpdate && <Timer nextUpdate={nextUpdate} connected={connected} />}
            </div>
            <div className="rates-grid">
                {Object.entries(groupedRates).map(([symbol, currencies]) => (
                    <div key={symbol} className="rate-card">
                        <div className="crypto-symbol">
                            <CryptoIcon symbol={symbol} />
                            <span className="symbol-name">{symbol}</span>
                            {symbolChanges[symbol] && (
                                <span className={`change-percent change-${symbolChanges[symbol].direction}`}>
                                    {symbolChanges[symbol].percent}% {symbolChanges[symbol].arrow}
                                </span>
                            )}
                        </div>
                        {Object.entries(currencies).map(([currency, data]) => (
                            <div key={currency} className="rate-info">
                                <span className="currency">{currency}:</span>
                                <span className="price">
                                    {data.price}
                                </span>
                            </div>
                        ))}
                    </div>
                ))}
            </div>
            {lastUpdate && (
                <div className="timestamp">
                    Last updated: {lastUpdate.toLocaleString()}
                </div>
            )}
        </div>
    );
};

export default App; 