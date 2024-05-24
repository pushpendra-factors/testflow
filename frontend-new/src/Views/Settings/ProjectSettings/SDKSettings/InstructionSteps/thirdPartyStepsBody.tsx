import React, { useState } from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import { Button, Input, Tooltip, notification } from 'antd';
import { camelCase } from 'lodash';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { udpateProjectSettings } from 'Reducers/global';
import { bindActionCreators } from 'redux';
import SdkVerificationFooter from '../SdkVerificationFooter';
import style from './index.module.scss';

const ThirdPartyStepsBody = ({
  udpateProjectSettings
}: ThirdPartyStepsBodyProps) => {
  const [cdpType, setCdpType] = useState('');
  const { active_project, currentProjectSettings } = useSelector(
    (state: any) => state.global
  );
  const handleCDPTypeChangeClick = (type: string) => {
    setCdpType(type);
  };

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
  return (
    <div className='flex flex-col gap-1.5 px-4'>
      <Text
        type='paragraph'
        color='character-secondary'
        weight='bold'
        extraClass='m-0 mb-4 -ml-4'
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
              <Text type='title' level={6} color='mono-6' extraClass='m-0'>
                1. Take your API key and configure Factors as a destination in
                your {camelCase(cdpType)} Workspace.
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
              <Text type='title' level={6} color='mono-6' extraClass='m-0'>
                2. Once done, enable all the data sources inside{' '}
                {camelCase(cdpType)} that you would like to send to factors
              </Text>
            </div>
            <SdkVerificationFooter type='cdp' />
          </>
        )}
      </div>
    </div>
  );
};

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      udpateProjectSettings
    },
    dispatch
  );
const connector = connect(null, mapDispatchToProps);
type ThirdPartyStepsBodyProps = ConnectedProps<typeof connector>;

export default connector(ThirdPartyStepsBody);
