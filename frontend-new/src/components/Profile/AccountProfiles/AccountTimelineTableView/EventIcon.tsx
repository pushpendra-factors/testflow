// EventIcon.tsx
import { eventIconsColorMap } from 'Components/Profile/constants';
import { CustomStyles, EventIconProps } from 'Components/Profile/types';
import React from 'react';

function EventIcon({ icon, size = 16 }: EventIconProps): JSX.Element {
  if (!icon) return <div />;

  // exception
  const showIcon = icon === 'globe' ? 'globepointer' : icon;

  const styles: CustomStyles = {
    '--border-color': eventIconsColorMap?.[showIcon]?.borderColor,
    '--bg-color': eventIconsColorMap?.[showIcon]?.bgColor,
    '--icon-size': `${size}px`
  };

  return (
    <div className='event-icon' style={styles as React.CSSProperties}>
      <img src={`/assets/icons/${showIcon}.svg`} alt='' loading='lazy' />
    </div>
  );
}

export default EventIcon;
