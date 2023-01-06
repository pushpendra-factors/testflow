import { Skeleton } from 'antd';
import { FallBackImage } from 'Constants/templates.constants';
import React, { useEffect, useState } from 'react';
import { StartFreshImage } from 'Constants/templates.constants';
import styles from './index.module.scss';

const TemplateThumbnailImage = ({
  eachState,
  TemplatesThumbnail,
  isStartFresh
}) => {
  let [isLoaded, setIsLoaded] = useState(false);
  useEffect(() => {
    return () => {
      setTimeout(() => {
        setIsLoaded(false);
      }, 1000);
    };
  }, []);
  // Below is to render StartFresh Banner Image
  // And to Skeleton is that image not loaded
  if (isStartFresh === true) {
    return (
      <>
        {isLoaded === false ? (
          <Skeleton.Image
            className={styles['ant-skeleton']}
            loading={!isLoaded}
            active={!isLoaded}
            style={{
              padding: '5px 0px',
              margin: '0 auto',
              borderRadius: '5px',
              height: '196px',
              width: '100%'
            }}
          />
        ) : (
          ''
        )}
        <img
          onLoad={() => setIsLoaded(true)}
          style={{
            display: isLoaded === true ? 'block' : 'none',
            padding: '5px 0px',
            margin: '0 auto',
            borderRadius: '5px'
          }}
          src={StartFreshImage}
        />
      </>
    );
  }

  // To Render Normal Templates in Step1
  return (
    <>
      {isLoaded === false ? (
        <Skeleton.Image
          className={styles['ant-skeleton']}
          loading={!isLoaded}
          style={{
            height: '196px',
            width: '100%',
            padding: '5px 0px',
            margin: '0 auto',
            borderRadius: '5px',
            width: '100%'
          }}
          active={isLoaded}
        />
      ) : (
        ''
      )}
      <img
        onLoad={() => setIsLoaded(true)}
        style={{
          display: isLoaded === true ? 'block' : 'none',
          padding: '5px 0px',
          margin: '0 auto',
          borderRadius: '5px',
          width: '100%'
        }}
        src={
          TemplatesThumbnail.has(
            eachState.title.toLowerCase().replace(/\s/g, '')
          )
            ? TemplatesThumbnail.get(
                eachState.title.toLowerCase().replace(/\s/g, '')
              ).image
            : FallBackImage
        }
      />
    </>
  );
};

export default TemplateThumbnailImage;
