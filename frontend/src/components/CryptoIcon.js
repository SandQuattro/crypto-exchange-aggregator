import React from 'react';

const CryptoIcon = ({ symbol }) => {
    const icons = {
        'BTC': '₿',
        'ETH': 'Ξ',
        'LTC': 'Ł',
        'DOGE': 'Ð',
        'USDT': '₮',
        'USDC': '$'
    };

    return (
        <span className="crypto-icon">
            {icons[symbol] || '💰'}
        </span>
    );
};

export default CryptoIcon; 