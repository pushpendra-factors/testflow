// EventIcon.tsx
import React from 'react';
import { eventIconsColorMap } from '../../utils';
import { CustomStyles, EventIconProps } from './types';

const EventIcon: React.FC<EventIconProps> = ({ icon, size = 16 }) => {
  if (!icon) return null;

  const styles: CustomStyles = {
    '--border-color': eventIconsColorMap?.[icon]?.borderColor,
    '--bg-color': eventIconsColorMap?.[icon]?.borderColor
  };

  const handleImageError = (
    e: React.SyntheticEvent<HTMLImageElement, Event>
  ) => {
    const defaultSrc = `/assets/icons/${icon}.svg`;
    if (e.currentTarget.src !== defaultSrc) {
      e.currentTarget.src = defaultSrc;
    }
  };

  return (
    <div className='event-icon' style={styles as React.CSSProperties}>
      <img
        src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${icon}.svg`}
        onError={handleImageError}
        alt=''
        height={size}
        width={size}
        loading='lazy'
      />
    </div>
  );
};

export default EventIcon;
