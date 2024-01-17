import { Tooltip } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import { TextWithTooltipProps } from './types';

function TextWithOverflowTooltip({
  text,
  tooltipText,
  extraClass,
  disabled = false
}: TextWithTooltipProps) {
  const tooltipRef = useRef<HTMLDivElement>(null);
  const [isOverflowing, setIsOverflowing] = useState(false);

  useEffect(() => {
    const element = tooltipRef.current;
    if (element) {
      setIsOverflowing(element.offsetWidth < element.scrollWidth);
    }
  }, [text]);

  return (
    <Tooltip
      title={tooltipText || text}
      trigger={isOverflowing && !disabled ? 'hover' : []}
    >
      <div ref={tooltipRef} className={`text-with-tooltip ${extraClass || ''}`}>
        {text}
      </div>
    </Tooltip>
  );
}

export default TextWithOverflowTooltip;
