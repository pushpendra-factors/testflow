import { Carousel } from 'antd';
import React, { useEffect, useRef } from 'react';
import AccountLimitNudge from './AccountLimitNudge';
import { UpgradeNudges } from './config';
import FeatureNudgeComponent from './FeatureNudgeComponent';
import styles from './index.module.scss';

const CarouselNudge = ({ percentage, limit, usage }: CarouselNudgeProps) => {
  const randomIndex = Math.floor(Math.random() * UpgradeNudges.length);
  const carouselRef = useRef(null);

  const selectedNudge = UpgradeNudges[randomIndex];

  useEffect(() => {
    setTimeout(() => {
      if (carouselRef?.current) {
        carouselRef.current.next();
      }
    }, 3000);
  }, []);
  return (
    <Carousel
      ref={carouselRef}
      dotPosition='right'
      style={{ height: 90 }}
      className={styles.upgradeNudgeCarousel}
    >
      <div>
        <AccountLimitNudge
          percentage={percentage}
          limit={limit}
          usage={usage}
        />
      </div>
      {selectedNudge && (
        <div>
          <FeatureNudgeComponent config={selectedNudge} />
        </div>
      )}
    </Carousel>
  );
};

interface CarouselNudgeProps {
  percentage: number;
  limit: number;
  usage: number;
}

export default CarouselNudge;
