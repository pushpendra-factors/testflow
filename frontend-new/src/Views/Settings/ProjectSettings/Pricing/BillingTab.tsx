import React from 'react';
import ProgressBar from 'Components/GenericComponents/Progress';
import { SVG, Text } from 'Components/factorsComponents';
import { PLANS } from 'Constants/plans.constants';
import { FeatureConfigState } from 'Reducers/featureConfig/types';
import { PathUrls } from 'Routes/pathUrls';
import { Alert, Button, Divider, Tooltip } from 'antd';
import useAgentInfo from 'hooks/useAgentInfo';
import moment from 'moment';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
const BillingTab = () => {
  const history = useHistory();
  const { email, isAdmin } = useAgentInfo();
  const { plan, sixSignalInfo } = useSelector(
    (state: any) => state.featureConfig
  ) as FeatureConfigState;
  const sixSignalLimit = sixSignalInfo?.limit || 0;
  const sixSignalUsage = sixSignalInfo?.usage || 0;
  const isSolutionsAdmin = email === 'solutions@factors.ai';
  const isFreePlan = plan?.name === PLANS.PLAN_FREE;

  const handleUpgradePlan = () => {
    if (isSolutionsAdmin) {
      history.push(PathUrls.ConfigurePlans);
      return;
    }
    if (isAdmin) {
      window.open(
        `https://calendly.com/srikrishna-s/30-min-demo-call?month=${moment().format(
          'yyyy-MM'
        )}`,
        '_blank'
      );
      return;
    }
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
            >
              {plan?.display_name || plan?.name}
            </Text>

            {/* <Tag color='orange'>Monthly</Tag> */}
          </div>
          {isFreePlan && (
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
        {/* <div>
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
                    weight={'bold'}
                    extraClass={'m-0 text-right'}
                    color='brand-color'
                  >
                    Renews August 16th 2023
                  </Text>
                </div> */}
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
