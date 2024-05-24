import { LoadingOutlined } from '@ant-design/icons';
import { SVG, Text } from 'Components/factorsComponents';
import { delay } from 'Utils/global';
import logger from 'Utils/logger';
import { Button, Divider, Spin, notification } from 'antd';
import React, { useState } from 'react';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { fetchProjectSettingsV1 } from 'Reducers/global';
import { bindActionCreators } from 'redux';
import { OnboardingSupportLink } from 'Onboarding/utils';

type SdkVerificationProps = {
  type: 'gtm' | 'manual' | 'cdp';
};

const LoadingIcon = <LoadingOutlined style={{ fontSize: 20 }} spin />;

const SdkVerificationFooter = ({
  type,
  fetchProjectSettingsV1
}: SdkVerificationFooterProps) => {
  const [verificationLoading, setVerificationLoading] = useState(false);
  const [errorState, setErrorState] = useState(false);
  const { active_project, projectSettingsV1 } = useSelector(
    (state: any) => state.global
  );
  const int_completed = projectSettingsV1?.int_completed;
  const [sdkVerified, setSdkVerified] = useState(!!int_completed);

  const handleSdkVerification = async () => {
    try {
      if (verificationLoading) {
        notification.warning({
          message: 'Processing',
          description: 'SDK Verification already in process!',
          duration: 2
        });
        return;
      }

      setVerificationLoading(true);
      setErrorState(false);
      await delay(5000);

      const res = await fetchProjectSettingsV1(active_project.id);
      if (res?.data?.int_completed) {
        setSdkVerified(true);
        notification.success({
          message: 'Success',
          description: 'We are receiving data from your website',
          duration: 3
        });
      } else {
        notification.error({
          message: 'Error',
          description: 'We are not receiving data from your website',
          duration: 3
        });
        setErrorState(true);
      }

      setVerificationLoading(false);
    } catch (error) {
      logger.error(error);
      setErrorState(true);
      setVerificationLoading(false);
    }
  };

  return (
    <div className='mt-4'>
      <Divider />
      {verificationLoading && (
        <div className='flex gap-2 items-center'>
          <div className='flex items-center justify-center'>
            <Spin indicator={LoadingIcon} />
          </div>

          <Text
            type='title'
            level={6}
            color='character-primary'
            extraClass='m-0 '
          >
            Checking for website data. It may take some time
          </Text>
        </div>
      )}

      {sdkVerified && !verificationLoading && (
        <div className='flex justify-between items-center'>
          <div className='flex  items-center'>
            <SVG name='CheckCircle' extraClass='inline' color='#52C41A' />
            <Text
              type='title'
              level={6}
              color='character-primary'
              extraClass='m-0 ml-2 inline'
            >
              {type === 'cdp'
                ? 'Events recieved successfully. 🎉'
                : 'Verified. Your script is up and running. 🎉'}
            </Text>
          </div>
          <Button
            type='text'
            size='small'
            style={{ color: '#1890FF' }}
            onClick={handleSdkVerification}
            loading={verificationLoading}
          >
            {type === 'cdp' ? 'Check again' : 'Verify again'}
          </Button>
        </div>
      )}
      {!int_completed && !errorState && !verificationLoading && (
        <div className='flex gap-2 items-center'>
          <Text type='paragraph' color='mono-6' extraClass='m-0'>
            {type === 'cdp'
              ? 'No events received yet'
              : 'Have you already added the code?'}
          </Text>
          <Button onClick={handleSdkVerification} loading={verificationLoading}>
            {type === 'cdp' ? 'Check for events' : 'Verify it now'}
          </Button>
        </div>
      )}
      {errorState && !verificationLoading && (
        <div className='flex items-center'>
          <SVG name='CloseCircle' extraClass='inline' color='#F5222D' />
          <Text
            type='title'
            level={6}
            color='character-primary'
            extraClass='m-0 ml-2 inline'
          >
            {type === 'cdp'
              ? 'No events received so far.'
              : 'Couldn’t detect SDK.'}
          </Text>
          <Button
            type='text'
            size='small'
            style={{ color: '#1890FF', padding: 0 }}
            onClick={handleSdkVerification}
            loading={verificationLoading}
          >
            Verify again
          </Button>
          <Text
            type='title'
            level={6}
            color='character-primary'
            extraClass='m-0 ml-1 inline'
          >
            or
          </Text>
          <Button
            type='text'
            size='small'
            style={{ color: '#1890FF', padding: 0, marginLeft: 4 }}
            onClick={() => window.open(OnboardingSupportLink, '_blank')}
          >
            book a call
          </Button>
        </div>
      )}
    </div>
  );
};

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjectSettingsV1
    },
    dispatch
  );

const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type SdkVerificationFooterProps = ReduxProps & SdkVerificationProps;

export default connector(SdkVerificationFooter);
