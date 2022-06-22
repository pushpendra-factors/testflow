import React from 'react';

function LegendsCircle({ color, extraClass }) {
  return (
    <div
      className={`w-4 h-4 rounded-lg ${extraClass}`}
      style={{ backgroundColor: color }}
    ></div>
  );
}

export default LegendsCircle;
