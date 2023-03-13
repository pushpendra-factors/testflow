import React, { useEffect, useState } from 'react';
import { Button } from 'antd';
import { useSelector } from 'react-redux';
import { Text, SVG } from 'Components/factorsComponents';
import EnrichPages from './EnrichPages';
import EnrichCountries from './EnrichCountries';
import { FeatureModes, SixSignalConfigType } from './types';

const EnrichFeature = ({ title, type, subtitle }: EnrichFeatureProps) => {
  const [mode, setMode] = useState<FeatureModes>('configure');
  //   @ts-ignore
  const six_signal_config: SixSignalConfigType = useSelector(
    (state) => state?.global?.currentProjectSettings?.six_signal_config
  );

  const active_project = useSelector((state) => state.global.active_project);
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
  }, [six_signal_config, type]);
  return (
    <div className='flex flex-col border-bottom--thin py-4'>
      <div
        className={`flex items-center ${
          mode === 'view' ? 'justify-between' : 'justify-start'
        }`}
      >
        <div>
          <Text type='title' level={6} weight='bold' extraClass='m-0 mb-1.5'>
            {title}
          </Text>
          {subtitle && (
            <Text type='title' level={8} color='grey' extraClass='m-0 mb-3'>
              {subtitle}
            </Text>
          )}
        </div>
        {mode === 'view' && (
          <div>
            <Button
              size='middle'
              shape='circle'
              onClick={() => setMode('edit')}
              icon={<SVG name={'Edit'} size={18} color='#8692A3' />}
            />
          </div>
        )}
      </div>

      {mode === 'configure' && (
        <div>
          <Button type='link' onClick={() => setMode('edit')}>
            Configure rules
          </Button>
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

type EnrichFeatureProps = {
  type: 'country' | 'page';
  title: string;
  subtitle?: string;
};

export default EnrichFeature;
