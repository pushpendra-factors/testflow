import { Tooltip } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import { CustomStyles } from 'Components/Profile/types';
import { TextWithTooltipProps } from './types';

function TextWithOverflowTooltip({
  text,
  tooltipText,
  extraClass,
  disabled = false,
  maxLines = 1
}: TextWithTooltipProps) {
  const styles: CustomStyles = {
    '--max-lines': maxLines
  };

  const tooltipRef = useRef<HTMLDivElement>(null);
  const [showTooltip, setShowTooltip] = useState(false);

  useEffect(() => {
    const element = tooltipRef.current;
    if (element) {
      const isOverflowing =
        element.offsetWidth < element.scrollWidth ||
        element.offsetHeight < element.scrollHeight;
      const hasTooltipText = tooltipText ? text !== tooltipText : false;
      setShowTooltip(isOverflowing || hasTooltipText);
    }
  }, [text, tooltipText]);

  return (
    <Tooltip
      title={tooltipText || text}
      trigger={showTooltip && !disabled ? 'hover' : []}
    >
      <div
        style={maxLines > 1 ? (styles as React.CSSProperties) : undefined}
        ref={tooltipRef}
        className={`${
          maxLines > 1 ? 'text-with-tooltip--multiline' : 'text-with-tooltip'
        } ${extraClass || ''}`}
      >
        {text}
      </div>
    </Tooltip>
  );
}

export default TextWithOverflowTooltip;
