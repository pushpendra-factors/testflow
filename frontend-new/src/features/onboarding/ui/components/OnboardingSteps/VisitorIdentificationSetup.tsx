import { SVG, Text } from 'Components/factorsComponents';
import logger from 'Utils/logger';
import EnrichFeature from 'Views/Settings/ProjectSettings/IntegrationSettings/SixSignalFactors/EnrichFeature';
import { SixSignalConfigType } from 'Views/Settings/ProjectSettings/IntegrationSettings/SixSignalFactors/types';
import {
  Button,
  Col,
  Divider,
  Radio,
  RadioChangeEvent,
  Row,
  notification
} from 'antd';
import confirm from 'antd/lib/modal/confirm';
import useMobileView from 'hooks/useMobileView';
import { isEmpty } from 'lodash';
import React, { useEffect, useState } from 'react';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { udpateProjectSettings } from 'Reducers/global';
import {
  CommonStepsProps,
  OnboardingStepsConfig,
  VISITOR_IDENTIFICATION_SETUP
} from '../../types';

const Step3 = ({
  udpateProjectSettings,
  incrementStepCount,
  decrementStepCount
}: Step3PropsType) => {
  const isMobileView = useMobileView();
  const [enrichmentType, setEnrichmentType] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(false);
  const { active_project, currentProjectSettings } = useSelector(
    (state) => state?.global
  );
  const six_signal_config: SixSignalConfigType =
    currentProjectSettings.six_signal_config;
  const onboarding_steps: OnboardingStepsConfig =
    currentProjectSettings?.onboarding_steps;

  const handleEnrichmentChange = (e: RadioChangeEvent) => {
    if (e.target.value === false) {
      if (!six_signal_config || isEmpty(six_signal_config)) {
        setEnrichmentType(false);
      } else {
        confirm({
          title: 'Confirmation',
          content: `Are you sure you want to remove the Enrichment Rules?`,
          async onOk() {
            try {
              await udpateProjectSettings(active_project?.id, {
                six_signal_config: {}
              });
            } catch (error) {
              logger.error('Error in updating project settings', error);
            }
          },
          onCancel() {
            // Reset the switch value to the previous one
          }
        });
      }
    } else {
      setEnrichmentType(e.target.value);
    }
  };

  const handleSubmission = async () => {
    try {
      setLoading(true);

      if (onboarding_steps?.[VISITOR_IDENTIFICATION_SETUP]) {
        incrementStepCount();
        setLoading(false);

        return;
      }

      let updatedOnboardingConfig = { [VISITOR_IDENTIFICATION_SETUP]: true };
      if (onboarding_steps) {
        updatedOnboardingConfig = {
          ...onboarding_steps,
          ...updatedOnboardingConfig
        };
      }

      await udpateProjectSettings(active_project.id, {
        onboarding_steps: updatedOnboardingConfig
      });
      incrementStepCount();
      setLoading(false);
    } catch (error) {
      setLoading(false);

      logger.error('Error in verifying SDK', error);
      notification.error({
        message: 'Error',
        description: 'Error in Saving settings!',
        duration: 3
      });
    }
  };

  const checkContinueButtonDisablity = () => {
    return enrichmentType === null;
  };

  useEffect(() => {
    if (!six_signal_config || isEmpty(six_signal_config)) {
      // setEnrichmentType(false);
    } else {
      setEnrichmentType(true);
    }
  }, [six_signal_config]);

  return (
    <div>
      <Row>
        <Col
          xs={24}
          sm={24}
          md={24}
          className={`${isMobileView ? 'text-center' : ''}`}
        >
          <Text
            type={'title'}
            level={3}
            color={'character-primary'}
            extraClass={'m-0'}
            weight={'bold'}
          >
            Activate Deanonymisation
          </Text>
          <Text
            type={'title'}
            level={6}
            extraClass={'m-0 mt-1'}
            color='character-secondary'
          >
            You can choose to identify all accounts that visit your website or
            set custom rules to identify some of them.
          </Text>
          <Text
            type={'title'}
            level={6}
            extraClass='m-0 mt-1 '
            color='character-secondary'
          >
            (This affects your quota of monthly accounts identified)
          </Text>
        </Col>
        <Col span={24}>
          <div className='my-8'>
            <Radio.Group
              onChange={handleEnrichmentChange}
              value={enrichmentType}
            >
              <Radio value={false}>Identify all accounts</Radio>
              <Radio value={true}>Set custom rules</Radio>
            </Radio.Group>
          </div>
          {enrichmentType === false && (
            <>
              <Text
                type={'title'}
                level={6}
                extraClass='m-0'
                color='character-secondary'
              >
                Identify all accounts that visit your website. This ensures that
                you don’t miss out on any account.
              </Text>
            </>
          )}
          {enrichmentType && (
            <>
              <div className='mt-4'>
                <EnrichFeature
                  type='page'
                  title='Identify accounts who visited specific pages'
                  subtitle='Include or exclude pages to only identify accounts that visit the pages you care about'
                  actionButtonText='Select pages'
                />
              </div>
              <div className='mt-4'>
                <EnrichFeature
                  type='country'
                  title='Identify accounts only from selected countries/region'
                  subtitle='Include or exclude countries to only identify accounts from the geography you care about'
                  actionButtonText='Select Countries '
                />
              </div>
            </>
          )}
          <div className='flex items-center my-6 gap-1'>
            <SVG name='InfoCircle' size='16' />
            <Text
              type={'title'}
              level={7}
              color='character-secondary'
              extraClass='m-0'
            >
              You can also do this later in the pricing screen inside the
              product.
            </Text>
          </div>
        </Col>
        {!isMobileView && <Divider className='mt-10 mb-6' />}
        <Col span={24}>
          <div
            className={`flex  mt-6 ${
              isMobileView
                ? 'flex-col-reverse justify-center items-center gap-2'
                : 'justify-between'
            }`}
          >
            <Button
              type='text'
              className='m-0'
              icon={<SVG name='ChevronLeft' size='16' />}
              onClick={decrementStepCount}
            >
              Previous Step
            </Button>
            <Button
              type='primary'
              className={'m-0'}
              onClick={handleSubmission}
              loading={loading}
              disabled={checkContinueButtonDisablity()}
            >
              Activate and continue
            </Button>
          </div>
        </Col>
      </Row>
    </div>
  );
};

const mapDispatchToProps = (dispatch) =>
  bindActionCreators({ udpateProjectSettings }, dispatch);
const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type Step3PropsType = ReduxProps & CommonStepsProps;

export default connector(Step3);
