// EventIcon.tsx
import { eventIconsColorMap } from 'Components/Profile/constants';
import { CustomStyles, EventIconProps } from 'Components/Profile/types';
import React from 'react';

function EventIcon({ icon, size = 16 }: EventIconProps): JSX.Element {
  if (!icon) return <></>;

  // exception
  const showIcon = icon === 'globe' ? 'globepointer' : icon;

  const styles: CustomStyles = {
    '--border-color': eventIconsColorMap?.[showIcon]?.borderColor,
    '--bg-color': eventIconsColorMap?.[showIcon]?.bgColor,
    '--icon-size': `${size}px`
  };

  const handleImageError = (
    e: React.SyntheticEvent<HTMLImageElement, Event>
  ) => {
    const defaultSrc = `/assets/icons/${showIcon}.svg`;
    if (e.currentTarget.src !== defaultSrc) {
      e.currentTarget.src = defaultSrc;
    }
  };

  return (
    <div className='event-icon' style={styles as React.CSSProperties}>
      <img
        src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${showIcon}.svg`}
        onError={handleImageError}
        alt=''
        loading='lazy'
      />
    </div>
  );
}

export default EventIcon;
