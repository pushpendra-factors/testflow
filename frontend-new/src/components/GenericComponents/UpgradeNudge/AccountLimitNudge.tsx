import { SVG, Text } from 'Components/factorsComponents';
import { Button } from 'antd';
import React from 'react';
import usePlanUpgrade from 'hooks/usePlanUpgrade';
import { getStatusColors } from './utils';
import ProgressBar from '../Progress';

const AccountLimitNudge = ({
  percentage,
  limit,
  usage
}: AccountLimitNudgeProps) => {
  const { backgroundColor, progressBarBackgroundColor, progressBarColor } =
    getStatusColors(percentage);
  const { handlePlanUpgradeClick } = usePlanUpgrade();
  return (
    <div
      className='px-6 py-5 flex items-center justify-between '
      style={{
        borderRadius: 12,
        height: 90,
        backgroundColor,
        border: '1px solid #f0f0f0'
      }}
    >
      <div className='flex items-center gap-4'>
        <SVG
          name='AccountIdentificationIllustration'
          color={progressBarColor}
          color2={backgroundColor}
        />
        <div className='flex justify-center flex-col '>
          <Text
            type='title'
            level={6}
            weight='bold'
            color='character-primary'
            extraClass='m-0'
          >
            Identified Accounts
          </Text>
          <div className='flex justify-start gap-3 items-center'>
            <div style={{ width: 500 }}>
              <ProgressBar
                percentage={percentage}
                trailColor={progressBarBackgroundColor}
                strokeColor={progressBarColor}
              />
            </div>

            <div>
              <Text
                type='paragraph'
                mini
                color='character-primary'
                extraClass='m-0'
              >
                {`${usage} of ${limit} used${
                  percentage >= 100 ? '. Enrichment paused' : ''
                }`}
              </Text>
            </div>
          </div>
        </div>
      </div>
      <Button
        onClick={() =>
          handlePlanUpgradeClick('ACCOUNT_LIMIT_ADDON_CLICK', 'addonClick')
        }
        icon={<SVG name='ArrowBottomUp' color='#595959' size={16} />}
      >
        Buy Add-on
      </Button>
    </div>
  );
};

interface AccountLimitNudgeProps {
  percentage: number;
  limit: number;
  usage: number;
}

export default AccountLimitNudge;
