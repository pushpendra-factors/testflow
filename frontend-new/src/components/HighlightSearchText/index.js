import React from 'react';

export const HighlightSearchText = ({ text = '', highlight = '' }) => {
  if (!highlight.trim()) {
    return <span>{text}</span>;
  }
  const regex = new RegExp(`(${_.escapeRegExp(highlight)})`, 'gi');
  const parts = text.split(regex);
  return (
    <span className={'truncate'}>
      {parts.map((part, i) =>
        regex.test(part) ? <b key={i}>{part}</b> : <span key={i}>{part}</span>
      )}
    </span>
  );
};
