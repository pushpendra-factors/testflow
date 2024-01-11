import React from 'react';
import styles from './index.module.scss';
import { Button, Divider, Tag, Tooltip } from 'antd';
import { Number, SVG, Text } from 'Components/factorsComponents';
import { PlansDetailStateInterface } from 'Reducers/plansConfig/types';
import { PLANS } from 'Constants/plans.constants';

function PlanDescriptionCard({
  plan,
  isPlanActive,
  isRecommendedPlan,
  planName,
  planIcon,
  planIconColor,
  planDescription,
  planFeatures,
  accountIdentifiedLimit,
  mtuLimit,
  isUserBillingAdmin,
  isButtonLoading,
  handleBuyButtonClick,
  isAdditionalAccountsAddonPurchased
}: PlanDescriptionCardProps) {
  const monthlyPlan = plan.terms.find((p) => p.period === 'month');
  const yearlyPlan = plan.terms.find((p) => p.period === 'year');

  return (
    <div className={`${styles.planDescriptionCard} flex justify-between`}>
      <div>
        {isPlanActive && (
          <div className={styles.activePlanTag}>
            <Tag color='orange'>Current Plan</Tag>
          </div>
        )}
        {!isPlanActive && isRecommendedPlan && (
          <div className={styles.reccommendationTag}>
            <Tag color='orange'>Recommended plan</Tag>
          </div>
        )}
        <div className='flex gap-2 items-center mt-2'>
          <SVG name={planIcon} size='28' color={planIconColor} />
          <Text
            type={'title'}
            level={3}
            weight={'bold'}
            color='character-primary'
            extraClass={'m-0 '}
          >
            {planName}
          </Text>
        </div>
        <Text
          type={'title'}
          level={6}
          color='character-primary'
          extraClass={'m-0 mt-2'}
        >
          {planDescription}
        </Text>
        <div className='mt-6 flex-col gap-1.5'>
          {planFeatures?.map((feature, i) => {
            return (
              <div key={i} className='flex gap-2 items-center'>
                <SVG name={'CheckCircleOutline'} size='18' color={'#52C41A'} />
                <Text type={'title'} level={7} extraClass={'m-0'}>
                  {feature}
                </Text>
              </div>
            );
          })}
        </div>
      </div>
      <div className={`${styles.planAmountContainer} h-auto flex  gap-12`}>
        <Divider type='vertical' style={{ height: '100%' }} />
        <div>
          <Text
            type={'title'}
            level={5}
            color='character-secondary'
            extraClass='m-0'
          >
            Starts for
          </Text>
          {monthlyPlan?.id && (
            <Text
              type={'title'}
              level={3}
              weight={'bold'}
              color='character-primary'
              extraClass={'m-0 '}
            >
              ${monthlyPlan.price}
              {monthlyPlan.price !== 0 && '/mo'}
            </Text>
          )}

          {monthlyPlan?.id && yearlyPlan?.id && (
            <Text
              type={'title'}
              level={7}
              color='character-secondary'
              extraClass={'m-0 '}
            >
              {monthlyPlan?.price === 0 && 'Can be upgraded'}

              {monthlyPlan?.price !== 0 && (
                <>
                  billed at <s>${monthlyPlan?.price * 12}</s>{' '}
                  <span style={{ fontWeight: 600 }}>
                    ${yearlyPlan?.price}/yr
                  </span>
                </>
              )}
            </Text>
          )}

          <Text
            type={'title'}
            level={7}
            color='character-primary'
            extraClass={'m-0 mt-8'}
            weight={'bold'}
          >
            Includes
          </Text>
          <Text
            type={'title'}
            level={7}
            color='character-primary'
            extraClass={'m-0 mt-1.5'}
          >
            <Number number={accountIdentifiedLimit} /> Accounts Identification
          </Text>
          <Text
            type={'title'}
            level={7}
            color='character-primary'
            extraClass={'m-0'}
          >
            <Number number={mtuLimit} /> Monthly tracked users
          </Text>

          {isPlanActive && plan.name === PLANS.PLAN_FREE ? null : (
            <Tooltip
              placement='top'
              title={`${
                isUserBillingAdmin
                  ? ''
                  : 'Please talk to your Billing Admin for upgrading plans'
              }`}
            >
              <Button
                className={`${
                  isUserBillingAdmin ? styles.outlineButton : 'mt-6'
                }`}
                disabled={!isUserBillingAdmin}
                onClick={() => handleBuyButtonClick(planName, isPlanActive)}
                loading={isPlanActive ? isButtonLoading : false}
              >
                <Text
                  type={'title'}
                  level={7}
                  color='character-primary'
                  weight={'bold'}
                  extraClass={'m-0'}
                >
                  {isPlanActive
                    ? isAdditionalAccountsAddonPurchased
                      ? 'Edit Add-ons'
                      : 'Buy Add-ons'
                    : 'Buy this Plan'}
                </Text>
              </Button>
            </Tooltip>
          )}
        </div>
      </div>
    </div>
  );
}

interface PlanDescriptionCardProps {
  isPlanActive: boolean;
  isRecommendedPlan: boolean;
  planName: string;
  planIcon: string;
  planIconColor: string;
  planDescription: string;
  planFeatures: string[];
  accountIdentifiedLimit: number;
  mtuLimit: number;
  plan: PlansDetailStateInterface;
  isUserBillingAdmin: boolean;
  handleBuyButtonClick: (planName: string, isPlanActive: boolean) => void;
  isButtonLoading: boolean;
  isAdditionalAccountsAddonPurchased: boolean;
}

export default PlanDescriptionCard;
