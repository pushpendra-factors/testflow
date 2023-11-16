import React from 'react';
import { Button } from 'antd';
import { Text } from 'Components/factorsComponents';
import styles from './index.module.scss';

const LastPlanCard = () => {
  return (
    <div className={`${styles.planDescriptionCard}  flex gap-10 items-center `}>
      <div>
        <Text
          type={'title'}
          level={3}
          weight={'bold'}
          color='character-primary'
          extraClass={'m-0 '}
        >
          These plans didn't work out for you?
        </Text>
        <Text
          type={'title'}
          level={7}
          color='character-primary'
          extraClass={'m-0 mt-1'}
        >
          Unlock the power of choice with our personalized pricing options. Find
          the ideal plan that aligns with your specific requirements and budget.
          Tailor-made solutions for your success await!
        </Text>
      </div>
      <div>
        <Button
          className={`${styles.outlineButton} w-full`}
          style={{ width: 290 }}
          onClick={() =>
            window.open(
              `https://factors.schedulehero.io/meet/srikrishna/discovery-call`,
              '_blank'
            )
          }
        >
          <Text
            type={'title'}
            level={7}
            color='character-primary'
            weight={'bold'}
            extraClass={'m-0'}
          >
            Talk to us
          </Text>
        </Button>
      </div>
    </div>
  );
};

export default LastPlanCard;
