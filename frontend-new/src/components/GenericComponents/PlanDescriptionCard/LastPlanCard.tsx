import React from 'react';
import { Button } from 'antd';
import { Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { PRICING_HELP_LINK } from 'Views/Settings/ProjectSettings/Pricing/utils';

function LastPlanCard() {
  return (
    <div
      className={`${styles.planDescriptionCard}  flex justify-between gap-10 items-center `}
    >
      <div>
        <Text
          type={'title'}
          level={3}
          weight={'bold'}
          color='character-primary'
          extraClass={'m-0 '}
        >
          Not sure which plan is best for you?
        </Text>
        <Text
          type={'title'}
          level={7}
          color='character-primary'
          extraClass={'m-0 mt-1'}
        >
          Get a customised product demo and identify the plan that best fits
          your needs.
        </Text>
      </div>
      <div>
        <Button
          className={`${styles.outlineButton} m-0 w-full`}
          style={{ width: 290 }}
          onClick={() => window.open(PRICING_HELP_LINK, '_blank')}
        >
          Talk to us
        </Button>
      </div>
    </div>
  );
}

export default LastPlanCard;
