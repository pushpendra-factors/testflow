import React from 'react';
import { Button, Divider, Tag, Tooltip } from 'antd';
import { Number, SVG, Text } from 'Components/factorsComponents';
import { PlansDetailStateInterface } from 'Reducers/plansConfig/types';
import { PLANS } from 'Constants/plans.constants';
import { PRICING_HELP_LINK } from 'Views/Settings/ProjectSettings/Pricing/utils';
import styles from './index.module.scss';

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
  isAdditionalAccountsAddonPurchased,
  seats,
  icons
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
        <div className='flex gap-2 items-center mt-2'>
          <SVG name={planIcon} size='28' color={planIconColor} />
          <Text
            type='title'
            level={3}
            weight='bold'
            color='character-primary'
            extraClass='m-0 '
          >
            {planName}
          </Text>
        </div>
        <Text
          type='title'
          level={6}
          color='character-primary'
          extraClass='m-0 mt-2'
        >
          {planDescription}
        </Text>
        <div className='mt-4 flex-col'>
          {planFeatures?.map((feature, i) => (
            <div
              key={i}
              className='flex gap-2 items-center '
              style={{ marginTop: i === 0 ? 0 : 6 }}
            >
              <SVG name='CheckCircleOutline' size='18' color='#52C41A' />
              <Text type='title' level={7} extraClass='m-0'>
                {feature}
              </Text>
            </div>
          ))}
        </div>
        {icons && icons?.length && (
          <div className='flex items-center gap-2 px-6 mt-2'>
            {icons.map((icon) => (
              <SVG name={icon} size='16' />
            ))}
          </div>
        )}
      </div>
      <div className={`${styles.planAmountContainer} h-auto flex  gap-12`}>
        <Divider type='vertical' style={{ height: '100%' }} />
        <div className='flex flex-col justify-between w-full'>
          {plan.name !== PLANS.PLAN_CUSTOM && (
            <div>
              <Text
                type='title'
                level={6}
                color='character-secondary'
                extraClass='m-0'
              >
                Starts for
              </Text>
              {yearlyPlan?.id && plan.name !== PLANS.PLAN_FREE && (
                <>
                  <Text
                    type='title'
                    level={3}
                    weight='bold'
                    color='character-primary'
                    extraClass='m-0 '
                  >
                    $
                    <Number
                      number={yearlyPlan?.price ? yearlyPlan.price / 12 : 0}
                    />
                    /Month
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    color='character-secondary'
                    extraClass='m-0 '
                  >
                    billed annually
                  </Text>
                </>
              )}
              {monthlyPlan?.id && plan.name !== PLANS.PLAN_FREE && (
                <Text
                  type='title'
                  level={7}
                  weight='bold'
                  color='character-secondary'
                  extraClass='m-0'
                >
                  or $
                  <Number number={monthlyPlan?.price || 0} /> monthly
                </Text>
              )}
              {plan.name === PLANS.PLAN_FREE && (
                <>
                  <Text
                    type='title'
                    level={3}
                    weight='bold'
                    color='character-primary'
                    extraClass='m-0 '
                  >
                    $0
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    weight='bold'
                    color='character-secondary'
                    extraClass='m-0 '
                  >
                    Can be upgraded
                  </Text>
                </>
              )}

              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0 mt-6'
                weight='bold'
              >
                Includes
              </Text>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0 mt-1.5'
              >
                Up to{' '}
                <b>
                  <Number number={accountIdentifiedLimit} />
                </b>{' '}
                Companies Identified/Mo
              </Text>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0'
              >
                Up to{' '}
                <b>
                  <Number number={mtuLimit} />
                </b>{' '}
                Monthly Visitors
              </Text>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0'
              >
                {seats ? (
                  <>
                    Up to <b>{seats}</b> seats
                  </>
                ) : (
                  'Unlimited Seats'
                )}
              </Text>
            </div>
          )}

          {plan.name === PLANS.PLAN_CUSTOM && (
            <div>
              <Text
                type='title'
                level={3}
                weight='bold'
                color='character-primary'
                extraClass='m-0 '
              >
                Talk to us
              </Text>
              <Text
                type='title'
                level={6}
                color='character-secondary'
                extraClass='m-0 '
              >
                Bespoke plans for agencies, enterprises, and established teams
                looking to scale GTM
              </Text>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0 mt-6'
                weight='bold'
              >
                Includes
              </Text>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0 mt-1.5'
              >
                Custom Companies Identified/Mo
              </Text>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0'
              >
                Custom Monthly Visitors
              </Text>
              <Text
                type='title'
                level={7}
                color='character-primary'
                extraClass='m-0'
              >
                Unlimited Seats
              </Text>
            </div>
          )}

          <div>
            {plan.name === PLANS.PLAN_CUSTOM ? (
              <Button
                className={`${styles.outlineButton} m-0 w-full`}
                style={{ width: 290 }}
                onClick={() => window.open(PRICING_HELP_LINK, '_blank')}
              >
                Talk to us
              </Button>
            ) : isPlanActive && plan.name === PLANS.PLAN_FREE ? null : (
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
                    isUserBillingAdmin && !isRecommendedPlan
                      ? styles.outlineButton
                      : ''
                  } mt-6`}
                  disabled={!isUserBillingAdmin}
                  onClick={() => handleBuyButtonClick(planName, isPlanActive)}
                  loading={isPlanActive ? isButtonLoading : false}
                  type={isRecommendedPlan ? 'primary' : 'default'}
                  block
                >
                  {isPlanActive
                    ? isAdditionalAccountsAddonPurchased
                      ? 'Edit Add-ons'
                      : 'Buy Add-ons'
                    : 'Buy this Plan'}
                </Button>
              </Tooltip>
            )}
          </div>
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
  seats: string;
  icons?: string[];
}

export default PlanDescriptionCard;
