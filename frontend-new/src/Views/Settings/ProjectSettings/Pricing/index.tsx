import ProgressBar from 'Components/GenericComponents/Progress';
import { SVG, Text } from 'Components/factorsComponents';
import { FeatureConfigState } from 'Reducers/featureConfig/types';
import { PathUrls } from 'Routes/pathUrls';
import { Alert, Breadcrumb, Button, Divider, Tabs, Tag, Tooltip } from 'antd';
import useAgentInfo from 'hooks/useAgentInfo';
import React from 'react';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';

const Pricing = () => {
  const history = useHistory();
  const { email } = useAgentInfo();
  const { plan, sixSignalInfo } = useSelector(
    (state: any) => state.featureConfig
  ) as FeatureConfigState;
  const sixSignalLimit = sixSignalInfo?.limit || 0;
  const sixSignalUsage = sixSignalInfo?.usage || 0;
  const isAdmin = email === 'solutions@factors.ai';
  return (
    <div>
      <div className='flex gap-3 items-center'>
        <div className='cursor-pointer' onClick={() => history.goBack()}>
          <SVG name='ArrowLeft' size='16' />
        </div>
        <div>
          <Breadcrumb>
            <Breadcrumb.Item>Settings</Breadcrumb.Item>
            <Breadcrumb.Item>Pricing</Breadcrumb.Item>
            <Breadcrumb.Item>Billing</Breadcrumb.Item>
          </Breadcrumb>
        </div>
      </div>
      <div className='mt-6'>
        <Tabs defaultActiveKey='1'>
          <Tabs.TabPane tab='Billing' key='1'>
            <div className='py-8'>
              <div className='flex justify-between'>
                <div>
                  <div className='flex items-center gap-2'>
                    <SVG name='Userplus' size='28' color='#1890FF' />
                    <Text
                      type={'title'}
                      level={2}
                      weight={'bold'}
                      color='character-primary'
                      extraClass={'m-0 '}
                    >
                      {plan?.name}
                    </Text>

                    {/* <Tag color='orange'>Monthly</Tag> */}
                  </div>
                  {/* <div className='mt-2'>
                    <Text
                      type={'paragraph'}
                      extraClass='m-0'
                      color='character-primary'
                    >
                      $0.0 USD / month
                    </Text>
                  </div> */}
                  <div className='mt-5'>
                    <Tooltip
                      title={`${
                        isAdmin
                          ? 'Configure Plans'
                          : 'Talk to our Sales team to upgrade'
                      }`}
                    >
                      <Button
                        type='primary'
                        disabled={!isAdmin}
                        onClick={() => {
                          history.push(PathUrls.ConfigurePlans);
                        }}
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
                  <ProgressBar
                    percentage={(sixSignalUsage / sixSignalLimit) * 100}
                  />
                  {false && (
                    <div className='mt-5'>
                      <Alert
                        message={
                          <Text type={'paragraph'} mini color='character-title'>
                            Account identification stopped. Close to 250
                            accounts lost so far.
                          </Text>
                        }
                        type='error'
                        showIcon
                      />
                    </div>
                  )}
                  <Tooltip
                    title={`${
                      isAdmin
                        ? 'Configure Plans'
                        : 'Talk to our Sales team to upgrade'
                    }`}
                  >
                    <Button
                      type='link'
                      style={{ marginTop: 20 }}
                      onClick={() => {
                        history.push(PathUrls.ConfigurePlans);
                      }}
                      disabled={!isAdmin}
                    >
                      Buy Add on
                    </Button>
                  </Tooltip>
                </div>
              </div>
            </div>
          </Tabs.TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default Pricing;
