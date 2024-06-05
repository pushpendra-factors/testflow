import { Tooltip } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import { CustomStyles } from 'Components/Profile/types';
import { Link } from 'react-router-dom';
import { TextWithTooltipProps } from './types';

function TextWithOverflowTooltip({
  text,
  tooltipText,
  extraClass,
  disabled = false,
  maxLines = 1,
  hasLink = false,
  linkTo = '',
  linkState = {},
  onClick,
  active,
  activeClass = '',
  alwaysShowTooltip = false
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
      setShowTooltip(alwaysShowTooltip || isOverflowing || hasTooltipText);
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
        } ${extraClass || ''} ${active ? activeClass : ''}`}
      >
        {hasLink ? (
          <Link
            onClick={onClick}
            to={{
              pathname: linkTo,
              state: linkState
            }}
          >
            {text}
          </Link>
        ) : (
          <span onClick={onClick}>{text}</span>
        )}
      </div>
    </Tooltip>
  );
}

export default TextWithOverflowTooltip;
