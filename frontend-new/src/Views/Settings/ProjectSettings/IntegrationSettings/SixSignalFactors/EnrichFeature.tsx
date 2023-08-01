import React, { useEffect, useState } from 'react';
import { Button, Modal, notification } from 'antd';
import { connect, ConnectedProps, useSelector } from 'react-redux';
import { Text, SVG } from 'Components/factorsComponents';
import EnrichPages from './EnrichPages';
import EnrichCountries from './EnrichCountries';
import { FeatureModes, SixSignalConfigType } from './types';
import { udpateProjectSettings } from 'Reducers/global';
import { bindActionCreators } from 'redux';

const EnrichFeature = ({
  title,
  type,
  subtitle,
  udpateProjectSettings,
  actionButtonText = 'Configure rules'
}: EnrichFeatureProps) => {
  const [mode, setMode] = useState<FeatureModes>('configure');
  //   @ts-ignore
  const six_signal_config: SixSignalConfigType = useSelector(
    (state) => state?.global?.currentProjectSettings?.six_signal_config
  );

  const active_project = useSelector((state) => state.global.active_project);

  const handleDelete = async () => {
    Modal.confirm({
      title: 'Are you sure you want to delete this Rule?',
      content:
        'You are about to delete this Rule. Factors will no longer filter 6 signal data as per this Rule.',
      okText: 'Delete',
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          if (!active_project?.id) return '';
          let updatedSixSignalConfig: SixSignalConfigType = {
            ...six_signal_config
          };
          if (type === 'country') {
            updatedSixSignalConfig.country_exclude = undefined;
            updatedSixSignalConfig.country_include = undefined;
          } else if (type === 'page') {
            updatedSixSignalConfig.pages_exclude = undefined;
            updatedSixSignalConfig.pages_include = undefined;
          }
          await udpateProjectSettings(active_project?.id, {
            six_signal_config: updatedSixSignalConfig
          });
          setMode('configure');
          notification.success({
            message: 'Success',
            description: `Successfully updated settings`,
            duration: 3
          });
        } catch (error) {
          console.error('Error in deleting settings', error);
        }
      },
      onCancel: () => {}
    });
  };

  useEffect(() => {
    //checking for country type
    if (type === 'country') {
      if (
        (six_signal_config?.country_exclude &&
          six_signal_config.country_exclude?.length > 0) ||
        (six_signal_config?.country_include &&
          six_signal_config.country_include?.length > 0)
      ) {
        setMode('view');
      }
    }
  }, [
    six_signal_config?.country_exclude,
    six_signal_config?.country_include,
    type
  ]);

  useEffect(() => {
    //checking for page type
    if (type === 'page') {
      if (
        (six_signal_config?.pages_exclude &&
          six_signal_config.pages_exclude?.length > 0) ||
        (six_signal_config?.pages_include &&
          six_signal_config.pages_include?.length > 0)
      ) {
        setMode('view');
      }
    }
  }, [
    type,
    six_signal_config?.pages_exclude,
    six_signal_config?.pages_include
  ]);
  return (
    <div className={`flex flex-col py-4`}>
      <div
        className={`flex items-center ${
          mode === 'view' ? 'justify-between' : 'justify-start'
        }`}
      >
        <div>
          <Text
            type='title'
            level={6}
            color='character-primary'
            extraClass='m-0 mb-1.5'
          >
            {title}
          </Text>
          {subtitle && (
            <Text
              type='title'
              level={8}
              color='character-secondary'
              extraClass='m-0 mb-3'
            >
              {subtitle}
            </Text>
          )}
        </div>
        {mode === 'view' && (
          <div className='flex gap-2'>
            <Button
              size='middle'
              onClick={() => setMode('edit')}
              icon={<SVG name={'Edit'} size={18} color='#8692A3' />}
              type='text'
            />
            <Button
              size='middle'
              onClick={handleDelete}
              icon={<SVG name={'Delete'} size={18} color='#8692A3' />}
              type='text'
            />
          </div>
        )}
      </div>

      {mode === 'configure' && (
        <div>
          <Button onClick={() => setMode('edit')}>{actionButtonText}</Button>
        </div>
      )}
      {/* Rendering page enrich component  */}
      {mode !== 'configure' && type === 'page' && (
        <EnrichPages
          mode={mode}
          setMode={setMode}
          sixSignalConfig={six_signal_config}
          projectId={active_project?.id || ''}
        />
      )}
      {/* Rendering countries enrich component */}
      {mode !== 'configure' && type === 'country' && (
        <EnrichCountries
          mode={mode}
          setMode={setMode}
          sixSignalConfig={six_signal_config}
          projectId={active_project?.id || ''}
        />
      )}
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
type ReduxProps = ConnectedProps<typeof connector>;
type EnrichFeatureType = {
  type: 'country' | 'page';
  title: string;
  subtitle?: string;
  actionButtonText?: string;
};

type EnrichFeatureProps = EnrichFeatureType & ReduxProps;

export default connector(EnrichFeature);
