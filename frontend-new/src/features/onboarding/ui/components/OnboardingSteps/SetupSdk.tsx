import { SVG, Text } from 'Components/factorsComponents';
import {
  Col,
  Row,
  Collapse,
  Tag,
  Divider,
  Button,
  notification,
  Input,
  Tooltip,
  Dropdown,
  Menu
} from 'antd';
import useMobileView from 'hooks/useMobileView';
import React, { useEffect, useState } from 'react';
import style from './index.module.scss';
import CodeBlockV2 from 'Components/CodeBlock/CodeBlockV2';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import logger from 'Utils/logger';
import { bindActionCreators } from 'redux';
import { fetchProjectSettingsV1, udpateProjectSettings } from 'Reducers/global';
import _ from 'lodash';
import ScriptHtml from 'Views/Settings/ProjectSettings/SDKSettings/ScriptHtml';
import { generateSdkScriptCode } from 'Views/Settings/ProjectSettings/SDKSettings/utils';
import { PDFDownloadLink } from '@react-pdf/renderer';
import StepsPdf from './StepsPdf';
import { generateCopyText } from '../../../utils';
import {
  CommonStepsProps,
  OnboardingStepsConfig,
  SDK_SETUP
} from '../../types';

const { Panel } = Collapse;

const Step2 = ({
  fetchProjectSettingsV1,
  udpateProjectSettings,
  incrementStepCount,
  decrementStepCount
}: OnboardingStep2Props) => {
  const [loading, setLoading] = useState(false);
  const [errorState, setErrorState] = useState<{
    gtm: boolean;
    manual: boolean;
    cdp: boolean;
  }>({ gtm: false, manual: false, cdp: false });
  const [cdpType, setCdpType] = useState('');
  const isMobileView = useMobileView();
  const { active_project, projectSettingsV1, currentProjectSettings } =
    useSelector((state: any) => state.global);
  const int_completed = projectSettingsV1?.int_completed;
  const onboarding_steps: OnboardingStepsConfig =
    currentProjectSettings?.onboarding_steps;
  const [sdkVerified, setSdkVerified] = useState(int_completed ? true : false);

  const projectToken = active_project.token;
  const assetURL = currentProjectSettings.sdk_asset_url;
  const apiURL = currentProjectSettings.sdk_api_url;

  const renderCodeBlock = () => (
    <CodeBlockV2
      collapsedViewText={
        <>
          <span style={{ color: '#2F80ED' }}>{`<script>`}</span>
          {`(function(c)d.appendCh.....func("`}
          <span style={{ color: '#EB5757' }}>{`${projectToken}`}</span>
          {`")`}
          <span style={{ color: '#2F80ED' }}>{`</script>`}</span>
        </>
      }
      fullViewText={
        <ScriptHtml
          projectToken={projectToken}
          assetURL={assetURL}
          apiURL={apiURL}
        />
      }
      textToCopy={generateSdkScriptCode(assetURL, projectToken, apiURL)}
    />
  );

  const renderSDKVerificationFooter = (type: 'gtm' | 'manual' | 'cdp') => (
    <div className='mt-4'>
      <Divider />
      {sdkVerified && (
        <div className='flex justify-between items-center'>
          <div>
            <SVG name={'CheckCircle'} extraClass={'inline'} />
            <Text
              type={'title'}
              level={6}
              color={'character-primary'}
              extraClass={'m-0 ml-2 inline'}
            >
              {type === 'cdp'
                ? 'Events recieved successfully.'
                : 'Verified. Your script is up and running.'}
            </Text>
          </div>
          <Button
            type={'text'}
            size={'small'}
            style={{ color: '#1890FF' }}
            onClick={() => handleSdkVerification(type)}
            loading={loading}
          >
            {type === 'cdp' ? 'Check again' : 'Verify again'}
          </Button>
        </div>
      )}
      {!int_completed && !errorState[type] && (
        <div className='flex gap-2 items-center'>
          <Text type='paragraph' color='mono-6' extraClass='m-0'>
            {type === 'cdp'
              ? 'No events received yet'
              : 'Have you already added the code?'}
          </Text>
          <Button onClick={() => handleSdkVerification(type)}>
            {type === 'cdp' ? 'Check for events' : 'Verify it now'}
          </Button>
        </div>
      )}
      {errorState[type] && (
        <div className='flex items-center'>
          <SVG name={'CloseCircle'} extraClass={'inline'} color='#F5222D' />
          <Text
            type={'title'}
            level={6}
            color={'character-primary'}
            extraClass={'m-0 ml-2 inline'}
          >
            {type === 'cdp'
              ? 'No events received so far.'
              : 'Couldn’t detect SDK.'}
          </Text>
          <Button
            type={'text'}
            size={'small'}
            style={{ color: '#1890FF', padding: 0 }}
            onClick={() => handleSdkVerification(type)}
            loading={loading}
          >
            Verify again
          </Button>
          <Text
            type={'title'}
            level={6}
            color={'character-primary'}
            extraClass={'m-0 ml-1 inline'}
          >
            or
          </Text>
          <Button
            type={'text'}
            size={'small'}
            style={{ color: '#1890FF', padding: 0, marginLeft: 4 }}
            onClick={() =>
              window.open('https://calendly.com/aravindhvetri', '_blank')
            }
          >
            book a call
          </Button>
        </div>
      )}
    </div>
  );

  const renderGtmContent = () => (
    <div className='flex flex-col gap-1.5 px-4'>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        1. Sign in to&nbsp;
        <span>
          <a href='https://tagmanager.google.com/' target='_blank'>
            Google Tag Manager
          </a>
        </span>
        , select “Workspace”, and “Add a new tag”
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        2. Name it “Factors tag”. Select Edit on Tag Configuration
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        3. Under custom, select custom HTML
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        4. Copy the below tracking script and paste it on the HTML field, Select
        Save
      </Text>
      <div className='py-4'>{renderCodeBlock()}</div>

      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        5. In theTriggers popup, select Add Trigger and select All Pages
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        6. The trigger has been added. Click on Publish at the top of your GTM
        window!
      </Text>
      {renderSDKVerificationFooter('gtm')}
    </div>
  );

  const renderManualSetupContent = () => (
    <div className='flex flex-col gap-1.5 px-4'>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        Add the below javascript code on every page between the &lt;head&gt; and
        &lt;/head&gt; tags.
      </Text>
      <div className='py-4'>{renderCodeBlock()}</div>
      {renderSDKVerificationFooter('manual')}
    </div>
  );

  const generateCollapseHeader = (
    title: string,
    subtitle: string,
    tag?: string
  ) => (
    <div className='w-full'>
      <div className='flex gap-2'>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          color='character-primary'
          extraClass='m-0'
        >
          {title}
        </Text>
        {tag && (
          <div>
            <Tag color='success'>{tag}</Tag>
          </div>
        )}
      </div>

      <Text
        type={'title'}
        level={6}
        color='character-secondary'
        extraClass='m-0'
      >
        {subtitle}
      </Text>
    </div>
  );

  const renderCDPContent = () => (
    <div className='flex flex-col gap-1.5 px-4'>
      <Text
        type='paragraph'
        color='character-secondary'
        weight={'bold'}
        extraClass={'m-0 mb-4 -ml-4'}
      >
        Select your CDP
      </Text>
      <div>
        <div className='flex items-center gap-3'>
          <div
            className={
              cdpType === 'segment'
                ? style.dashedButtonActive
                : style.dashedButton
            }
          >
            <Button
              type='dashed'
              onClick={() => handleCDPTypeChangeClick('segment')}
              icon={<SVG name='Segment_ads' size='24' />}
              size='large'
            >
              Segment
            </Button>
          </div>
          <div
            className={
              cdpType === 'rudderstack'
                ? style.dashedButtonActive
                : style.dashedButton
            }
          >
            <Button
              type='dashed'
              onClick={() => handleCDPTypeChangeClick('rudderstack')}
              icon={<SVG name='Rudderstack_ads' size='24' />}
              size='large'
            >
              Rudderstack
            </Button>
          </div>
        </div>
        {cdpType && (
          <>
            <div className='mt-6 flex flex-col gap-4'>
              <Text type={'title'} level={6} color='mono-6' extraClass='m-0'>
                1. Take your API key and configure Factors as a destination in
                your {_.camelCase(cdpType)} Workspace.
              </Text>
              <div>
                <Input.Group compact>
                  <Input
                    style={{
                      width: 300
                    }}
                    defaultValue={active_project?.private_token}
                    disabled
                  />
                  <Tooltip title='Copy Code'>
                    <Button
                      onClick={handleCDPCopyClick}
                      type='text'
                      className={style.outlineButton}
                    >
                      <SVG name='TextCopy' size='24' />
                    </Button>
                  </Tooltip>
                </Input.Group>
              </div>
              <Text type={'title'} level={6} color='mono-6' extraClass='m-0'>
                2. Once done, enable all the data sources inside{' '}
                {_.camelCase(cdpType)} that you would like to send to factors
              </Text>
            </div>
            {renderSDKVerificationFooter('cdp')}
          </>
        )}
      </div>
    </div>
  );

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

  const getInstructionMenu = () => {
    return (
      <Menu>
        <Menu.Item key='1'>
          <Button
            type='text'
            icon={<SVG name='TextCopy' size='24' color='#8C8C8C' />}
            onClick={copyInstruction}
          >
            <Text
              color='character-primary'
              type={'title'}
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
                  scriptCode={generateSdkScriptCode(
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
  };

  const handleSdkVerification = async (type: 'gtm' | 'manual' | 'cdp') => {
    try {
      setLoading(true);
      setErrorState({
        gtm: false,
        manual: false,
        cdp: false
      });
      const res = await fetchProjectSettingsV1(active_project.id);

      if (res?.data?.int_completed) {
        setSdkVerified(true);
        notification.success({
          message: 'Success',
          description: 'SDK Verified!',
          duration: 3
        });
      } else {
        notification.error({
          message: 'Error',
          description: 'SDK not Verified!',
          duration: 3
        });
        setErrorState({ ...errorState, [type]: true });
      }

      setLoading(false);
    } catch (error) {
      logger.error(error);
      setErrorState({ ...errorState, [type]: true });
      setLoading(false);
    }
  };

  const handleCDPTypeChangeClick = (type: string) => {
    setCdpType(type);
  };

  const renderDocumentationLink = () => (
    <div>
      <Link
        className='flex items-center font-semibold gap-2'
        target='_blank'
        to={{
          pathname:
            'https://help.factors.ai/en/collections/3953559‒getting‒started'
        }}
      >
        <SVG name='ArrowUpRightSquare' color='#40A9FF' />
        <Text type={'title'} level={6} color='brand-color' extraClass='m-0'>
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

  const handleCDPCopyClick = async () => {
    let updateProjectSettingsFlag = false;
    let updatedProjectSettings = {};
    if (cdpType === 'rudderstack' && !currentProjectSettings?.int_rudderstack) {
      updateProjectSettingsFlag = true;
      updatedProjectSettings = {
        int_rudderstack: true
      };
    }
    if (cdpType === 'segment' && !currentProjectSettings?.int_segment) {
      updateProjectSettingsFlag = true;
      updatedProjectSettings = {
        int_segment: true
      };
    }

    try {
      if (updateProjectSettingsFlag) {
        await udpateProjectSettings(active_project.id, updatedProjectSettings);
      }
      navigator?.clipboard
        ?.writeText(active_project?.private_token)
        .then(() => {
          notification.success({
            message: 'Success',
            description: 'Successfully copied!',
            duration: 3
          });
        })
        .catch(() => {
          notification.error({
            message: 'Failed!',
            description: 'Failed to copy!',
            duration: 3
          });
        });
    } catch (error) {
      notification.error({
        message: 'Failed!',
        description: 'Failed to copy!',
        duration: 3
      });
    }
  };

  const handleSDkSubmission = async () => {
    try {
      setLoading(true);
      if (!sdkVerified) {
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
                type={'title'}
                level={3}
                color={'character-primary'}
                extraClass={'m-0'}
                weight={'bold'}
              >
                Connect with your website
              </Text>
              <div className='flex flex-wrap gap-1 mt-1 w-full'>
                <Text
                  type={'title'}
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
            type={'title'}
            level={6}
            color='character-primary'
            extraClass='m-0 mb-4'
            weight={'bold'}
          >
            Use Factors SDK
          </Text>
          <Collapse
            key='gtm'
            expandIconPosition='right'
            className={style.collapse}
          >
            <Panel
              key={'gtm'}
              header={generateCollapseHeader(
                'Setup using GTM',
                'Add Factors SDK quickly using Google Tag Manager without any engineering effort',
                'Most Popular'
              )}
            >
              {renderGtmContent()}
            </Panel>
          </Collapse>
          <div className='mt-6'>
            <Collapse
              key='manual'
              expandIconPosition='right'
              className={style.collapse}
            >
              <Panel
                key={'manual'}
                header={generateCollapseHeader(
                  'Manual Setup',
                  'Add Factors SDK manually in the head section for all pages you wish to get data for'
                )}
              >
                {renderManualSetupContent()}
              </Panel>
            </Collapse>
          </div>
        </Col>
        <Col className='mt-10' span={24}>
          <Text
            type={'title'}
            level={6}
            color='character-primary'
            extraClass='m-0 mb-4'
            weight={'bold'}
          >
            Use a third party data source
          </Text>
          <Collapse
            key='gtm'
            expandIconPosition='right'
            className={style.collapse}
          >
            <Panel
              key={'gtm'}
              header={generateCollapseHeader(
                'Use a Customer Data Platform',
                'Use an existing Customer Data Platform to bring in website data and events'
              )}
            >
              {renderCDPContent()}
            </Panel>
          </Collapse>
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
              className={'m-0'}
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
};
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
