import React, { useEffect, useState } from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import logger from 'Utils/logger';
import EnrichFeature from 'Views/Settings/ProjectSettings/IntegrationSettings/SixSignalFactors/EnrichFeature';
import {
  Button,
  Checkbox,
  Col,
  Divider,
  Radio,
  RadioChangeEvent,
  Row,
  Tooltip,
  notification
} from 'antd';
import type { CheckboxChangeEvent } from 'antd/es/checkbox';
import confirm from 'antd/lib/modal/confirm';
import useMobileView from 'hooks/useMobileView';
import { isEmpty } from 'lodash';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { udpateProjectSettings } from 'Reducers/global';
import { Link } from 'react-router-dom';
import {
  CommonStepsProps,
  OnboardingStepsConfig,
  VISITOR_IDENTIFICATION_SETUP
} from '../../types';
import styles from './index.module.scss';
import CheckListIllustration from '../../../../../assets/images/checklist_Illustration.png';
import { setFactorsDeAnonymisationProvider } from '../../../utils/service';
import { ClearbitTermsOfUseLink } from '../../../utils';

function Step3({
  udpateProjectSettings,
  incrementStepCount,
  decrementStepCount
}: Step3PropsType) {
  const isMobileView = useMobileView();
  const [enrichmentType, setEnrichmentType] = useState<boolean | null>(null);
  const [isClearbitSelected, setIsClearbitSelected] = useState<boolean>(true);
  const [loading, setLoading] = useState(false);
  const { active_project, currentProjectSettings } = useSelector(
    (state) => state?.global
  );

  const { six_signal_config } = currentProjectSettings;
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

  const handleClearbitTermsChange = (e: CheckboxChangeEvent) => {
    setIsClearbitSelected(e.target.checked);
  };

  const handleSubmission = async () => {
    try {
      setLoading(true);

      await setFactorsDeAnonymisationProvider(
        active_project.id,
        'factors_clearbit'
      );
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

  const checkContinueButtonDisablity = () =>
    enrichmentType === null || isClearbitSelected === false;

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
            type='title'
            level={3}
            color='character-primary'
            extraClass='m-0'
            weight='bold'
          >
            Activate Deanonymisation
          </Text>
          <Text
            type='title'
            level={6}
            extraClass='m-0 mt-1'
            color='character-secondary'
          >
            You can choose to identify all accounts that visit your website or
            set custom rules to identify some of them. This affects your monthly
            quota of accounts identified.
          </Text>
        </Col>
        <Col span={24}>
          <div className='my-8 flex justify-between items-center'>
            <Radio.Group
              onChange={handleEnrichmentChange}
              value={enrichmentType}
            >
              <Radio value={false}>Identify all accounts</Radio>
              <Radio value>Set custom rules</Radio>
            </Radio.Group>
            <Tooltip
              title='You can also change this later in 
the pricing page inside the product.'
            >
              <div className='flex items-center  gap-1'>
                <SVG name='InfoCircle' size='16' />

                <Text
                  type='title'
                  level={7}
                  color='character-title'
                  extraClass='m-0'
                >
                  Learn more
                </Text>
              </div>
            </Tooltip>
          </div>
          {enrichmentType === false && (
            <div className={styles.allAccountInfoContainer}>
              <img
                style={{ width: 100, height: 100 }}
                src={CheckListIllustration}
                alt='illustration'
              />
              <Text
                type='title'
                level={6}
                extraClass='m-0'
                color='character-secondary'
              >
                Identify all accounts that visit your website. This ensures that
                you don’t miss out on any account.
              </Text>
            </div>
          )}
          {enrichmentType && (
            <div className={styles.customRulesContainer}>
              <div>
                <EnrichFeature
                  type='page'
                  title='Identify accounts who visited specific pages'
                  subtitle={
                    <Text
                      type='title'
                      level={8}
                      color='character-secondary'
                      extraClass='m-0 mb-3'
                    >
                      Include or exclude pages to only identify accounts that
                      visit the pages you care about.{' '}
                      <span className='font-bold'>Note-</span> Do not include{' '}
                      <span className='font-bold'>https://</span> in the URL
                    </Text>
                  }
                  actionButtonText='Select pages'
                />
              </div>
              <div className='mt-4'>
                <EnrichFeature
                  type='country'
                  title='Identify accounts only from selected countries/region'
                  subtitle={
                    <Text
                      type='title'
                      level={8}
                      color='character-secondary'
                      extraClass='m-0 mb-3'
                    >
                      Include or exclude countries to only identify accounts
                      from the geography you care about
                    </Text>
                  }
                  actionButtonText='Select Countries '
                />
              </div>
            </div>
          )}
        </Col>
        <div className='flex gap-3 mt-8'>
          <Checkbox
            checked={isClearbitSelected}
            onChange={handleClearbitTermsChange}
          />
          <Text
            type='paragraph'
            mini
            color='character-primary'
            extraClass='inline-block'
          >
            I agree to use Clearbit for identifying accounts visiting the
            website as per
            <Link
              className='inline-block ml-1'
              target='_blank'
              to={{
                pathname: ClearbitTermsOfUseLink
              }}
            >
              <Text type='paragraph' mini weight='bold' color='brand-color-6'>
                {'  '} Terms of Use
              </Text>
            </Link>
          </Text>
        </div>
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
              className='m-0'
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
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators({ udpateProjectSettings }, dispatch);
const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type Step3PropsType = ReduxProps & CommonStepsProps;

export default connector(Step3);
