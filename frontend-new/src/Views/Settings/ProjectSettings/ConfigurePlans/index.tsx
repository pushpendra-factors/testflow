import { FeatureConfigState } from 'Reducers/featureConfig/types';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import CustomPlanConfigure from './CustomPlanConfigure';
import { Text } from 'Components/factorsComponents';
import { Switch, Modal, Spin, notification } from 'antd';
import useAgentInfo from 'hooks/useAgentInfo';
import { PLANS } from 'Constants/plans.constants';
import { changePlanType } from 'Reducers/featureConfig/services';
import { fetchFeatureConfig } from 'Reducers/featureConfig/middleware';
import logger from 'Utils/logger';
const { confirm } = Modal;

const ConfigurePlans = () => {
  const [switchValue, setSwitchValue] = useState(false);
  const [loading, setLoading] = useState(false);
  const { plan, loading: featureLoading } = useSelector(
    (state) => state.featureConfig
  ) as FeatureConfigState;
  const { active_project } = useSelector((state) => state.global);
  const dispatch = useDispatch();
  const { email } = useAgentInfo();
  const planName = plan?.name;
  useEffect(() => {
    if (planName === PLANS.PLAN_FREE) {
      setSwitchValue(false);
    } else {
      setSwitchValue(true);
    }
  }, [planName]);

  const handleSwitchChange = (value: boolean) => {
    confirm({
      title: 'Confirmation',
      content: `Are you sure you want to change the plan  to ${
        !value ? PLANS.PLAN_FREE : PLANS.PLAN_CUSTOM
      }?`,
      async onOk() {
        try {
          setLoading(true);
          await changePlanType(
            active_project?.id,
            !value ? PLANS.PLAN_FREE : PLANS.PLAN_CUSTOM
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
        <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0 m-1'}>
          Plan Configuration
        </Text>
      </div>
      <div className='flex items-center gap-3 my-5'>
        <Text type={'paragraph'} mini>
          Switch Plan:
        </Text>

        <Switch
          checked={switchValue}
          checkedChildren='CUSTOM'
          unCheckedChildren='FREE'
          disabled={email !== 'solutions@factors.ai'}
          onChange={handleSwitchChange}
        />
      </div>

      {planName !== 'FREE' && <CustomPlanConfigure />}
    </div>
  );
};

export default ConfigurePlans;
