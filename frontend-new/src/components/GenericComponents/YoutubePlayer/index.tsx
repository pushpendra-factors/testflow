import React from 'react';

interface YouTubePlayerProps {
  embeddedLink: string;
  extraClass?: string;
  title: string;
  width?: string;
  height?: string;
}

const YouTubePlayer: React.FC<YouTubePlayerProps> = ({
  embeddedLink,
  extraClass,
  title,
  width = '100%',
  height = '100%'
}) => {
  return (
    <>
      <iframe
        className={extraClass}
        title={title}
        width={width}
        height={height}
        src={embeddedLink}
        frameBorder='0'
        allow='accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture'
        allowFullScreen
      ></iframe>
    </>
  );
};

export default YouTubePlayer;
