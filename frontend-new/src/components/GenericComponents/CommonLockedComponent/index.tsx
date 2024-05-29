import { SVG, Text } from 'Components/factorsComponents';
import React from 'react';
import { Button } from 'antd';
import usePlanUpgrade from 'hooks/usePlanUpgrade';
import { Link } from 'react-router-dom';
import LockedScreenImage from '../../../assets/images/LockedScreen.png';
import LockedTabImage from '../../../assets/images/LockedTab.png';

function CommonLockedComponent({
  title,
  description,
  learnMoreLink,
  featureName,
  variant = 'default'
}: CommonLockedComponentPropType) {
  const { handlePlanUpgradeClick } = usePlanUpgrade();
  return (
    <div>
      {variant === 'default' && title && (
        <div>
          <Text
            type='title'
            level={3}
            weight='bold'
            color='character-title'
            id='fa-at-text--page-title'
          >
            {title}
          </Text>
          {description && (
            <div className='flex items-baseline flex-wrap'>
              <Text
                type='paragraph'
                mini
                color='character-primary'
                extraClass='inline-block'
              >
                {description}
                {learnMoreLink && (
                  <Link
                    className='inline-block ml-1'
                    target='_blank'
                    to={{
                      pathname: learnMoreLink
                    }}
                  >
                    <Text
                      type='paragraph'
                      mini
                      weight='bold'
                      color='brand-color-6'
                    >
                      {'  '} Learn more
                    </Text>
                  </Link>
                )}
              </Text>
            </div>
          )}
        </div>
      )}

      <div className='flex flex-col items-center mt-12 gap-6'>
        <img
          src={variant === 'default' ? LockedScreenImage : LockedTabImage}
          alt='locked screen'
          style={{ width: 250, height: 172 }}
        />
        <div>
          <Text
            type='title'
            level={3}
            extraClass='m-0 text-center'
            color='character-primary'
            id='fa-at-text--page-locked'
          >
            Your plan doesn’t have this feature
          </Text>
          <Text
            type='paragraph'
            mini
            color='character-secondary'
            extraClass='text-center'
          >
            This feature is not included in your current plan. Please upgrade to
            use this feature
          </Text>
        </div>
        <div>
          {variant === 'tab' && learnMoreLink && (
            <Button
              onClick={() => {
                window.open(learnMoreLink, '_blank');
              }}
              icon={<SVG name='NewTab' />}
            >
              Learn more
            </Button>
          )}
          <Button
            type='primary'
            onClick={() => {
              handlePlanUpgradeClick(featureName);
            }}
          >
            Upgrade now
          </Button>
        </div>
      </div>
    </div>
  );
}

type CommonLockedComponentPropType = {
  title?: string;
  description?: string;
  learnMoreLink?: string;
  featureName: string;
  variant?: 'default' | 'tab';
};

export default CommonLockedComponent;
