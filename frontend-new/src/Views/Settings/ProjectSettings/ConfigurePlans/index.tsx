import { FeatureConfigState } from 'Reducers/featureConfig/types';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Text } from 'Components/factorsComponents';
import { Switch, Modal, Spin, notification } from 'antd';
import useAgentInfo from 'hooks/useAgentInfo';
import { PLANS, PLANS_V0 } from 'Constants/plans.constants';
import { changePlanType } from 'Reducers/featureConfig/services';
import { fetchFeatureConfig } from 'Reducers/featureConfig/middleware';
import logger from 'Utils/logger';
import CustomPlanConfigure from './CustomPlanConfigure';
import { showV2PricingVersion } from '../Pricing/utils';

const { confirm } = Modal;

function ConfigurePlans() {
  const [switchValue, setSwitchValue] = useState(false);
  const [loading, setLoading] = useState(false);

  const {
    plan,
    loading: featureLoading,
    activeFeatures,
    addOns,
    sixSignalInfo
  } = useSelector((state) => state.featureConfig) as FeatureConfigState;
  const { active_project } = useSelector((state) => state.global);
  const showV2PricingVersionFlag = showV2PricingVersion(active_project);

  const dispatch = useDispatch();
  const { email } = useAgentInfo();
  const planName = plan?.name;

  const successCallback = () => {
    dispatch(fetchFeatureConfig(active_project?.id));
  };

  useEffect(() => {
    if (planName === PLANS_V0.PLAN_FREE) {
      setSwitchValue(false);
    } else {
      setSwitchValue(true);
    }
  }, [planName]);

  const handleSwitchChange = (value: boolean) => {
    confirm({
      title: 'Confirmation',
      content: `Are you sure you want to change the plan  to ${
        !value ? PLANS_V0.PLAN_FREE : PLANS_V0.PLAN_CUSTOM
      }?`,
      async onOk() {
        try {
          setLoading(true);
          await changePlanType(
            active_project?.id,
            !value ? PLANS_V0.PLAN_FREE : PLANS_V0.PLAN_CUSTOM
          );
          dispatch(fetchFeatureConfig(active_project?.id));
          setSwitchValue(value);
          notification.success({
            message: 'Success!',
            description: 'Successfully Updated Plan',
            duration: 3
          });
          setLoading(false);
        } catch (error) {
          setLoading(false);
          logger.error('Error in updating plan', error);
          notification.error({
            message: 'Error',
            description: 'Something went wrong. Could not update plan type',
            duration: 2
          });
        }
      },
      onCancel() {
        // Reset the switch value to the previous one
      }
    });
  };

  if (loading || featureLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
    );
  }
  return (
    <div>
      <div>
        <Text type='title' level={3} weight='bold' extraClass='m-0 m-1'>
          Plan Configuration
        </Text>
      </div>
      {showV2PricingVersionFlag ? (
        planName === PLANS.PLAN_CUSTOM || planName === PLANS_V0.PLAN_CUSTOM ? (
          <CustomPlanConfigure
            sixSignalInfo={sixSignalInfo}
            activeFeatures={activeFeatures}
            addOns={addOns}
            featureLoading={featureLoading}
            projectId={active_project?.id}
            successCallback={successCallback}
          />
        ) : (
          <Text type='paragraph' mini>
            Plan configuration is only allowed for {PLANS.PLAN_CUSTOM} plan
          </Text>
        )
      ) : null}

      {!showV2PricingVersionFlag ? (
        <>
          <div className='flex items-center gap-3 my-5'>
            <Text type='paragraph' mini>
              Switch Plan:
            </Text>

            <Switch
              checked={switchValue}
              checkedChildren={PLANS_V0.PLAN_CUSTOM}
              unCheckedChildren={PLANS_V0.PLAN_FREE}
              disabled={email !== 'solutions@factors.ai'}
              onChange={handleSwitchChange}
            />
          </div>

          {planName !== PLANS_V0.PLAN_FREE && (
            <CustomPlanConfigure
              sixSignalInfo={sixSignalInfo}
              activeFeatures={activeFeatures}
              addOns={addOns}
              featureLoading={featureLoading}
              projectId={active_project?.id}
              successCallback={successCallback}
            />
          )}
        </>
      ) : null}
    </div>
  );
}

export default ConfigurePlans;
