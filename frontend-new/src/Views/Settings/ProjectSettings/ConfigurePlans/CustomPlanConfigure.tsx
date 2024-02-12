import React, { useEffect, useState } from 'react';
import { Form, Input, Button, Spin, notification } from 'antd';
import FaSelect from 'Components/GenericComponents/FaSelect';
import { FEATURES } from 'Constants/plans.constants';
import { FeatureConfig, SixSignalInfo } from 'Reducers/featureConfig/types';
import { SVG, Text } from 'Components/factorsComponents';
import { OptionType } from 'Components/GenericComponents/FaSelect/types';
import logger from 'Utils/logger';
import { updatePlanConfig } from 'Reducers/featureConfig/services';
import { getAllActiveFeatures } from 'Reducers/featureConfig/utils';
import style from './index.module.scss';

interface CustomplanConfigureProps {
  projectId: string;
  activeFeatures?: FeatureConfig[];
  addOns?: FeatureConfig[];
  sixSignalInfo?: SixSignalInfo;
  featureLoading: boolean;
  successCallback: () => void;
}

function CustomPlanConfigure({
  activeFeatures,
  addOns,
  sixSignalInfo,
  featureLoading,
  projectId,
  successCallback
}: CustomplanConfigureProps) {
  const [selectedFeatures, setSelectedFeature] = useState<string[]>([]);
  const [showFeatureSelection, setShowFeatureSelection] = useState(false);
  const [loading, setLoading] = useState(false);

  const sixSignalLimit = sixSignalInfo?.limit || 0;

  const getFeatureOptions = () =>
    Object.entries(FEATURES).map(([key, value]) => ({
      value,
      label: key,
      isSelected: selectedFeatures.includes(value)
    }));

  const handleApplyClick = (
    _options: OptionType[],
    selectedOption: string[]
  ) => {
    setSelectedFeature(selectedOption);
    setShowFeatureSelection(false);
  };

  const onFinish = async (values: any) => {
    try {
      setLoading(true);
      const { accountLimit, mtuLimit } = values;
      if (!accountLimit || !mtuLimit) {
        logger.error('Invalid account or mtu limit');
      }
      await updatePlanConfig(
        projectId,
        Number(accountLimit),
        Number(mtuLimit),
        selectedFeatures
      );
      successCallback();
      notification.success({
        message: 'Success!',
        description: 'Successfully Updated Plan configuration',
        duration: 3
      });
      setLoading(false);
    } catch (error) {
      setLoading(false);
      logger.error('Error in updating plan config', error);
      notification.error({
        message: 'Error',
        description:
          error?.data?.err?.display_message ||
          'Something went wrong. Could not update plan configuration',
        duration: 2
      });
    }
  };

  useEffect(() => {
    if (!activeFeatures && !addOns) return;
    const allActiveFeatures = getAllActiveFeatures(activeFeatures, addOns);
    const selectedFeaturesState = allActiveFeatures.map(
      (feature) => feature.name
    );
    setSelectedFeature(selectedFeaturesState);
  }, [activeFeatures, addOns]);

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
      <Form onFinish={onFinish}>
        <Form.Item
          name='accountLimit'
          label={
            <Text type='paragraph' mini>
              Accounts Identified Limit
            </Text>
          }
          initialValue={sixSignalLimit}
        >
          <Input
            type='number'
            size='middle'
            style={{ borderRadius: 6, width: 200 }}
          />
        </Form.Item>
        {/* <br />
        <Form.Item
          name='mtuLimit'
          label={
            <Text type={'paragraph'} mini>
              MTU's Limit
            </Text>
          }
          initialValue={5000}
        >
          <Input
            type='number'
            defaultValue={10000}
            size='middle'
            style={{ borderRadius: 6, width: 200 }}
          />
        </Form.Item> */}
        <br />
        <Form.Item
          name='features'
          labelAlign='left'
          label={
            <Text type='paragraph' mini>
              Features
            </Text>
          }
        >
          <div>
            <div className={style.filter}>
              <Button
                onClick={() => setShowFeatureSelection(true)}
                className={`${style.customButton} flex items-center gap-1`}
              >
                <Text type='title' level={7} extraClass='m-0'>
                  Configure Features
                </Text>
                <SVG size={14} name='chevronDown' />
              </Button>
              {showFeatureSelection && (
                <FaSelect
                  options={getFeatureOptions()}
                  onClickOutside={() => setShowFeatureSelection(false)}
                  applyClickCallback={handleApplyClick}
                  allowSearch
                  variant='Multi'
                  loadingState={featureLoading}
                  allowSearchTextSelection={false}
                />
              )}
            </div>
          </div>
        </Form.Item>
        <br />
        <Form.Item>
          <Button type='primary' htmlType='submit'>
            Save Changes
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
}

export default CustomPlanConfigure;
