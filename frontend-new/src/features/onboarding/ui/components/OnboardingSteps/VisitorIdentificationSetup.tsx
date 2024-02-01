import React, { useEffect, useState } from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import logger from 'Utils/logger';
import EnrichFeature from 'Views/Settings/ProjectSettings/IntegrationSettings/SixSignalFactors/EnrichFeature';
import {
  Button,
  Col,
  Divider,
  Radio,
  RadioChangeEvent,
  Row,
  Tooltip,
  notification
} from 'antd';
import confirm from 'antd/lib/modal/confirm';
import useMobileView from 'hooks/useMobileView';
import { isEmpty } from 'lodash';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { udpateProjectSettings } from 'Reducers/global';
import {
  CommonStepsProps,
  FactorsDeAnonymisationProvider,
  OnboardingStepsConfig,
  VISITOR_IDENTIFICATION_SETUP
} from '../../types';
import styles from './index.module.scss';
import CheckListIllustration from '../../../../../assets/images/checklist_Illustration.png';
import { setFactorsDeAnonymisationProvider } from '../../../utils/service';

function Step3({
  udpateProjectSettings,
  incrementStepCount,
  decrementStepCount
}: Step3PropsType) {
  const isMobileView = useMobileView();
  const [enrichmentType, setEnrichmentType] = useState<boolean | null>(null);
  const [provider, setProvider] =
    useState<FactorsDeAnonymisationProvider>('factors_clearbit');
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

  const handleSubmission = async () => {
    try {
      setLoading(true);

      await setFactorsDeAnonymisationProvider(active_project.id, provider);
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

  const handleProviderChange = (_provider: FactorsDeAnonymisationProvider) => {
    if (provider !== _provider) {
      setProvider(_provider);
    }
  };

  const checkContinueButtonDisablity = () => enrichmentType === null;

  useEffect(() => {
    if (!six_signal_config || isEmpty(six_signal_config)) {
      // setEnrichmentType(false);
    } else {
      setEnrichmentType(true);
    }
  }, [six_signal_config]);

  useEffect(() => {
    if (
      currentProjectSettings?.factors_deanon_config?.clearbit
        ?.traffic_fraction === 1
    ) {
      setProvider('factors_clearbit');
    } else if (
      currentProjectSettings?.factors_deanon_config?.['6signal']
        ?.traffic_fraction === 1
    ) {
      setProvider('factors_6Signal');
    }
  }, [currentProjectSettings]);

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
            Identify accounts that visit your website and track their activity
            with one of these providers
          </Text>
          <div className='mt-6 flex gap-4'>
            <div
              className={`${styles.providerCard} ${
                provider === 'factors_clearbit' ? styles.selectedProvider : ''
              }`}
              onClick={() => handleProviderChange('factors_clearbit')}
            >
              <SVG name='ClearbitLogo' size={48} color='purple' />
              <div>
                <Text
                  type='title'
                  level={6}
                  weight='bold'
                  extraClass='m-0'
                  color='character-primary'
                >
                  Clearbit Reveal
                </Text>
                <Text
                  type='title'
                  level={8}
                  extraClass='m-0'
                  color='character-secondary'
                >
                  Use Clearbit Reveal to identify accounts
                </Text>
              </div>
              <div className={styles.providerCheckContainer}>
                {provider === 'factors_clearbit' && (
                  <SVG name='Check_circle' size={24} />
                )}
              </div>
            </div>
            <div
              className={`${styles.providerCard} ${
                provider === 'factors_6Signal' ? styles.selectedProvider : ''
              }`}
              onClick={() => handleProviderChange('factors_6Signal')}
            >
              <SVG name='SixSignalLogo' size={48} color='purple' />
              <div>
                <Text
                  type='title'
                  level={6}
                  weight='bold'
                  extraClass='m-0'
                  color='character-primary'
                >
                  6Signal by 6Sense
                </Text>
                <Text
                  type='title'
                  level={8}
                  extraClass='m-0'
                  color='character-secondary'
                >
                  Use 6Signal by 6Sense to identify accounts
                </Text>
              </div>
              <div className={styles.providerCheckContainer}>
                {provider === 'factors_6Signal' && (
                  <SVG name='Check_circle' size={24} />
                )}
              </div>
            </div>
          </div>
          <Text
            type='title'
            level={6}
            extraClass='m-0 mt-10'
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
            </div>
          )}
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
