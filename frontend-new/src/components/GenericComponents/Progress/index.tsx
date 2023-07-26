import React from 'react';
import { Progress } from 'antd';

const ProgressBar = ({
  percentage,
  showInfo = false,
  trailColor = '#D9D9D9'
}: ProgressProps) => {
  let strokeColor;
  if (percentage < 75) {
    strokeColor = '#1890FF';
  } else if (percentage < 100) {
    strokeColor = '#FAAD14';
  } else {
    strokeColor = '#EA6262';
  }
  return (
    <Progress
      percent={percentage}
      strokeColor={strokeColor}
      showInfo={showInfo}
      trailColor={trailColor}
    />
  );
};

type ProgressProps = {
  percentage: number;
  showInfo?: boolean;
  trailColor?: string;
};

export default ProgressBar;
