import React from 'react';

export default function ({ index }: { index: number }) {
  if (index % 4 < 2) {
    return (
      <div
        style={{
          height: '200px',
          background: '#e4e4e426',
          borderRadius: '10px',
          flex: `1 1 ${index % 2 == 0 ? '60%' : '25%'}`
        }}
      ></div>
    );
  }
  return (
    <div
      style={{
        height: '200px',
        background: '#e4e4e426',
        borderRadius: '10px',
        flex: `1 1 ${index % 2 ? '60%' : '25%'}`
      }}
    ></div>
  );
}
