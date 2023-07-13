import { SVG, Text } from 'Components/factorsComponents';
import React from 'react';
import ProgressBar from '../Progress';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';

const RangeNudge = ({
  title,
  amountUsed = 0,
  totalLimit = 0
}: RangeNudgeProps) => {
  const history = useHistory();
  const percentage = Number(((amountUsed / totalLimit) * 100).toFixed(2));
  let backgroundColor, borderColor;
  if (percentage < 75) {
    backgroundColor = '#FAFAFA';
    borderColor = '#D9D9D9';
  } else if (percentage < 100) {
    backgroundColor = '#FFF7E6';
    borderColor = '#FFC069';
  } else {
    backgroundColor = '#FFF1F0';
    borderColor = '#FF7875';
  }
  return (
    <div
      style={{
        background: backgroundColor,
        border: `1px solid ${borderColor}`,
        borderRadius: 6
      }}
      className='flex items-center px-4 py-2  justify-between'
    >
      <div className='flex gap-3'>
        <Text type={'paragraph'} mini color='character-title'>
          {title}
        </Text>
        <div style={{ width: 200 }}>
          <ProgressBar percentage={percentage} />
        </div>
        <Text type={'paragraph'} mini color='character-primary'>
          {`${amountUsed} of ${totalLimit} used`}
        </Text>
      </div>
      <div>
        <div
          className='flex items-center gap-2 cursor-pointer'
          onClick={() => history.push(PathUrls.SettingsPricing)}
        >
          <Text type={'paragraph'} mini color='brand-color'>
            Buy add on
          </Text>
          <SVG name='ArrowUpRightSquare' color='#40A9FF' />
        </div>
      </div>
    </div>
  );
};

type RangeNudgeProps = {
  title: string;
  nudgeLink?: string;
  nudgeText?: string;
  totalLimit: number;
  amountUsed: number;
};

export default RangeNudge;
