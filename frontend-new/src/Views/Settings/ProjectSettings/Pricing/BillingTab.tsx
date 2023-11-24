import React from 'react';
import ProgressBar from 'Components/GenericComponents/Progress';
import { SVG, Text } from 'Components/factorsComponents';
import { PLANS, PLANS_V0 } from 'Constants/plans.constants';
import { FeatureConfigState } from 'Reducers/featureConfig/types';
import { PathUrls } from 'Routes/pathUrls';
import { Alert, Button, Divider, Tag, Tooltip } from 'antd';
import useAgentInfo from 'hooks/useAgentInfo';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { PRICING_PAGE_TABS, showV2PricingVersion } from './utils';
import { PlansConfigState } from 'Reducers/plansConfig/types';
import moment from 'moment';
const BillingTab = () => {
  const history = useHistory();
  const { email, isAdmin } = useAgentInfo();
  const { plan, sixSignalInfo } = useSelector(
    (state: any) => state.featureConfig
  ) as FeatureConfigState;
  const { currentPlanDetail } = useSelector(
    (state: any) => state.plansConfig
  ) as PlansConfigState;
  const { active_project } = useSelector((state) => state.global);

  const sixSignalLimit = sixSignalInfo?.limit || 0;
  const sixSignalUsage = sixSignalInfo?.usage || 0;
  const isSolutionsAdmin = email === 'solutions@factors.ai';
  const isFreePlan =
    plan?.name === PLANS.PLAN_FREE || plan?.name === PLANS_V0?.PLAN_FREE;
  const showV2PricingVersionFlag = showV2PricingVersion(active_project);

  const handleUpgradePlan = () => {
    if (isSolutionsAdmin) {
      history.push(PathUrls.ConfigurePlans);
      return;
    }
    if (isAdmin && !showV2PricingVersionFlag) {
      window.open(
        `https://factors.schedulehero.io/meet/srikrishna/discovery-call`,
        '_blank'
      );
      return;
    }
    history.push(
      `${PathUrls.SettingsPricing}?activeTab=${PRICING_PAGE_TABS.UPGRADE}`
    );
  };

  return (
    <div className='py-4'>
      <div className='flex justify-between'>
        <div>
          <div className='flex items-center gap-2'>
            <SVG name='Userplus' size='28' color='#1890FF' />
            <Text
              type={'title'}
              level={3}
              weight={'bold'}
              color='character-primary'
              extraClass={'m-0 '}
              id={'fa-at-text--page-title'}
            >
              {showV2PricingVersionFlag && currentPlanDetail?.plan?.externalName
                ? currentPlanDetail.plan.externalName
                : plan?.display_name || plan?.name}
            </Text>
            {showV2PricingVersionFlag && currentPlanDetail?.period && (
              <>
                {currentPlanDetail.period === 'month' && (
                  <Tag color='orange'>Monthly</Tag>
                )}

                {currentPlanDetail.period == 'year' && (
                  <Tag color='orange'>Yearly</Tag>
                )}
              </>
            )}
          </div>
          {!showV2PricingVersionFlag && isFreePlan && (
            <div className='mt-2'>
              <Text
                type={'paragraph'}
                extraClass='m-0'
                color='character-primary'
              >
                $0.0 USD / month
              </Text>
            </div>
          )}
          {showV2PricingVersionFlag && (
            <div className='mt-2'>
              <Text
                type={'paragraph'}
                extraClass='m-0'
                color='character-primary'
              >
                ${currentPlanDetail?.plan?.amount || '0.0'}
                {' USD / '}
                {currentPlanDetail.period ? currentPlanDetail.period : ''}
              </Text>
            </div>
          )}

          <div className='mt-5'>
            <Tooltip
              title={`${
                isSolutionsAdmin
                  ? 'Configure Plans'
                  : 'Talk to our Sales team to upgrade'
              }`}
            >
              <Button
                type='primary'
                disabled={!isSolutionsAdmin && !isAdmin}
                onClick={handleUpgradePlan}
              >
                Upgrade Plan
              </Button>
            </Tooltip>
          </div>
        </div>
        {showV2PricingVersionFlag && currentPlanDetail?.renews_on && (
          <div>
            <Text
              type={'title'}
              level={5}
              extraClass={'m-0 text-right opacity-60'}
              color='character-primary'
            >
              Billing period
            </Text>
            <Text
              type={'title'}
              level={6}
              extraClass={'m-0 text-right mt-1'}
              color='brand-color'
            >
              Renews{' '}
              {moment(currentPlanDetail.renews_on).format('MMMM Do YYYY')}
            </Text>
          </div>
        )}
      </div>
      <Divider />
      <div
        className='rounded-lg border-gray-600 p-4'
        style={{ borderRadius: 8, border: '1px solid #F5F5F5' }}
      >
        <Text
          type={'paragraph'}
          extraClass='m-0'
          color='character-primary'
          weight={'bold'}
        >
          Accounts identified
        </Text>
        <Divider />
        <div>
          <div className='flex justify-between items-center'>
            <Text type={'paragraph'} mini>
              Default Monthly Quota
            </Text>
            <Text type={'paragraph'} mini>
              {`${sixSignalUsage} / ${sixSignalLimit}`}
            </Text>
          </div>
          <ProgressBar percentage={(sixSignalUsage / sixSignalLimit) * 100} />
          {false && (
            <div className='mt-5'>
              <Alert
                message={
                  <Text type={'paragraph'} mini color='character-title'>
                    Account identification stopped. Close to 250 accounts lost
                    so far.
                  </Text>
                }
                type='error'
                showIcon
              />
            </div>
          )}
          {/* {showV2PricingVersionFlag && (
            <>
              <div className='flex justify-between items-center mt-4'>
                <Text type={'paragraph'} mini>
                  Ad on - 27 Jul
                </Text>
                <Text type={'paragraph'} mini>
                  {`${sixSignalUsage} / ${sixSignalLimit}`}
                </Text>
              </div>
              <ProgressBar
                percentage={(sixSignalUsage / sixSignalLimit) * 100}
              />
            </>
          )} */}

          <Tooltip
            title={`${
              isSolutionsAdmin
                ? 'Configure Plans'
                : 'Talk to our Sales team to upgrade'
            }`}
          >
            <Button
              type='link'
              style={{ marginTop: 20 }}
              onClick={handleUpgradePlan}
              disabled={!isSolutionsAdmin && !isAdmin}
            >
              Buy Add on
            </Button>
          </Tooltip>
        </div>
      </div>
    </div>
  );
};

export default BillingTab;
