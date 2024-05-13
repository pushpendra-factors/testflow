import { SVG, Text } from 'Components/factorsComponents';
import { Col, Row, Divider, Button, notification, Dropdown, Menu } from 'antd';
import useMobileView from 'hooks/useMobileView';
import React, { useEffect, useState } from 'react';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import logger from 'Utils/logger';
import { bindActionCreators } from 'redux';
import { fetchProjectSettingsV1, udpateProjectSettings } from 'Reducers/global';
import {
  generateSdkScriptCode,
  generateSdkScriptCodeForPdf
} from 'Views/Settings/ProjectSettings/SDKSettings/utils';
import { PDFDownloadLink } from '@react-pdf/renderer';
import GtmSteps from 'Views/Settings/ProjectSettings/SDKSettings/InstructionSteps/gtmSteps';
import ManualSteps from 'Views/Settings/ProjectSettings/SDKSettings/InstructionSteps/manualSteps';
import ThirdPartySteps from 'Views/Settings/ProjectSettings/SDKSettings/InstructionSteps/thirdPartySteps';
import StepsPdf from './StepsPdf';
import { SDKDocumentation, generateCopyText } from '../../../utils';
import {
  CommonStepsProps,
  OnboardingStepsConfig,
  SDK_SETUP
} from '../../types';

import style from './index.module.scss';

function Step2({
  fetchProjectSettingsV1,
  udpateProjectSettings,
  incrementStepCount,
  decrementStepCount
}: OnboardingStep2Props) {
  const [loading, setLoading] = useState<boolean>(false);

  const isMobileView = useMobileView();
  const { active_project, projectSettingsV1, currentProjectSettings } =
    useSelector((state: any) => state.global);
  const int_completed = projectSettingsV1?.int_completed;
  const onboarding_steps: OnboardingStepsConfig =
    currentProjectSettings?.onboarding_steps;
  const [sdkVerified, setSdkVerified] = useState(!!int_completed);

  const projectToken = active_project.token;
  const assetURL = currentProjectSettings.sdk_asset_url;
  const apiURL = currentProjectSettings.sdk_api_url;

  const copyInstruction = () => {
    try {
      navigator?.clipboard
        ?.writeText(
          generateCopyText(
            generateSdkScriptCode(assetURL, projectToken, apiURL)
          )
        )
        .then(() => {
          notification.success({
            message: 'Success',
            description: 'Successfully copied!',
            duration: 3
          });
        });
    } catch (error) {
      logger.error(error);
      notification.success({
        message: 'Error',
        description: 'Error in copying code',
        duration: 3
      });
    }
  };

  const getInstructionMenu = () => (
    <Menu>
      <Menu.Item key='1'>
        <Button
          type='text'
          icon={<SVG name='TextCopy' size='24' color='#8C8C8C' />}
          onClick={copyInstruction}
        >
          <Text
            color='character-primary'
            type='title'
            level={7}
            extraClass='m-0 inline'
          >
            {' '}
            Copy instructions
          </Text>
        </Button>
      </Menu.Item>
      <Menu.Item key='2'>
        <Button
          type='text'
          icon={<SVG name='DownloadOutline' size='24' color='#8C8C8C' />}
        >
          <PDFDownloadLink
            document={
              <StepsPdf
                scriptCode={generateSdkScriptCodeForPdf(
                  assetURL,
                  projectToken,
                  apiURL,
                  false
                )}
              />
            }
            style={{
              color: 'rgba(0, 0, 0, 0.65)',
              marginLeft: 5,
              fontWeight: 400,
              fontSize: 14
            }}
            fileName='onboarding_steps.pdf'
          >
            {({ blob, url, _loading, error }) =>
              _loading ? 'Loading...' : 'Download as PDF'
            }
          </PDFDownloadLink>
        </Button>
      </Menu.Item>
    </Menu>
  );

  const renderDocumentationLink = () => (
    <div>
      <Link
        className='flex items-center font-semibold gap-2'
        target='_blank'
        to={{
          pathname: SDKDocumentation
        }}
      >
        <SVG name='ArrowUpRightSquare' color='#40A9FF' />
        <Text type='title' level={6} color='brand-color' extraClass='m-0'>
          Documentation
        </Text>
      </Link>
    </div>
  );

  const renderInstructionDropdown = () => (
    <div>
      <Dropdown
        placement='bottomRight'
        overlay={getInstructionMenu()}
        trigger={['click']}
      >
        <Button className={`${isMobileView ? '' : `${style.outlineButton}`}`}>
          {isMobileView ? <SVG name='PaperPlane' col /> : 'Send instructions'}
        </Button>
      </Dropdown>
    </div>
  );

  const handleSDkSubmission = async () => {
    try {
      setLoading(true);
      const res = await fetchProjectSettingsV1(active_project.id);

      if (!res?.data?.int_completed) {
        notification.warn({
          message: 'Warning',
          description:
            'Please Verify SDK or CDP events before proceeding to next step',
          duration: 3
        });
        setLoading(false);
        return;
      }

      if (onboarding_steps?.[SDK_SETUP]) {
        incrementStepCount();
        setLoading(false);

        return;
      }

      let updatedOnboardingConfig = { [SDK_SETUP]: true };
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
        description: 'Error in verifying SDK!',
        duration: 3
      });
    }
  };

  useEffect(() => {
    if (int_completed && !sdkVerified) {
      setSdkVerified(true);
    }
    if (!int_completed && sdkVerified) {
      setSdkVerified(false);
    }
  }, [int_completed, sdkVerified]);
  return (
    <div>
      <Row>
        <Col
          xs={24}
          sm={24}
          md={24}
          className={`${isMobileView ? 'text-center' : ''}`}
        >
          <div
            className={`flex ${
              isMobileView ? 'w-full' : 'items-center justify-between'
            }`}
          >
            <div className='w-full'>
              <Text
                type='title'
                level={3}
                color='character-primary'
                extraClass='m-0'
                weight='bold'
              >
                Connect with your website
              </Text>
              <div className='flex flex-wrap gap-1 mt-1 w-full'>
                <Text
                  type='title'
                  level={6}
                  extraClass={`m-0 ${isMobileView ? 'w-full' : ''}`}
                  color='character-secondary'
                >
                  Bring in data from your website for comprehensive analysis.
                </Text>
                {!isMobileView && <>{renderDocumentationLink()}</>}
              </div>
            </div>
            {!isMobileView && <>{renderInstructionDropdown()}</>}
          </div>
        </Col>
        {isMobileView && (
          <Col span={24} className='mt-8'>
            <div className='flex items-center justify-between'>
              {renderDocumentationLink()}
              {renderInstructionDropdown()}
            </div>
          </Col>
        )}
        <Col className='mt-8' span={24}>
          <Text
            type='title'
            level={6}
            color='character-primary'
            extraClass='m-0 mb-4'
            weight='bold'
          >
            Use Factors SDK
          </Text>
          <GtmSteps
            projectToken={projectToken}
            assetURL={assetURL}
            apiURL={apiURL}
            isOnboardingFlow
          />
          <div className='mt-6'>
            <ManualSteps
              projectToken={projectToken}
              assetURL={assetURL}
              apiURL={apiURL}
              isOnboardingFlow
            />
          </div>
        </Col>
        <Col className='mt-10' span={24}>
          <Text
            type='title'
            level={6}
            color='character-primary'
            extraClass='m-0 mb-4'
            weight='bold'
          >
            Use a third party data source
          </Text>
          <ThirdPartySteps />
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
              loading={loading}
              className='m-0'
              onClick={handleSDkSubmission}
              disabled={!sdkVerified}
            >
              Connect and continue
            </Button>
          </div>
        </Col>
      </Row>
    </div>
  );
}
const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjectSettingsV1,
      udpateProjectSettings
    },
    dispatch
  );
const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type OnboardingStep2Props = ReduxProps & CommonStepsProps;

export default connector(Step2);
