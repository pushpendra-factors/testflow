import React from 'react';
import { Progress } from 'antd';

const ProgressBar = ({
  percentage,
  showInfo = false,
  trailColor = '#D9D9D9',
  strokeColor
}: ProgressProps) => {
  let localStrokeColor = '';
  if (!strokeColor) {
    if (percentage < 75) {
      localStrokeColor = '#597EF7';
    } else if (percentage < 100) {
      localStrokeColor = '#FFA940';
    } else {
      localStrokeColor = '##F5222D';
    }
  }

  return (
    <Progress
      percent={percentage}
      strokeColor={strokeColor || localStrokeColor}
      showInfo={showInfo}
      trailColor={trailColor}
    />
  );
};

type ProgressProps = {
  percentage: number;
  showInfo?: boolean;
  trailColor?: string;
  strokeColor?: string;
};

export default ProgressBar;
