import React, { useEffect, useState } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text } from 'Components/factorsComponents';
import {
  EnrichPageData,
  EnrichTypes,
  FeatureModes,
  SixSignalConfigType
} from './types';
import { Button, Input, notification, Radio, Select, Tooltip } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { udpateProjectSettings } from 'Reducers/global';
import style from './index.module.scss';

const defaultPageData: EnrichPageData = {
  type: 'equals',
  value: ''
};
const EnrichPages = ({
  mode,
  setMode,
  sixSignalConfig,
  projectId,
  udpateProjectSettings
}: EnrichPagesProps) => {
  const [enrichType, setEnrichType] = useState<EnrichTypes | null>(null);
  const [data, setData] = useState<EnrichPageData[]>([defaultPageData]);
  const [errors, setErrors] = useState<number[] | null>(null);
  const [errorType, setErrorType] = useState<string>('');

  const updateDataAtIndex = (
    value: string,
    index: number,
    key: keyof EnrichPageData
  ) => {
    const updatedObj = {
      [key]: value?.trim()
    };
    setErrors(null);
    setErrorType('');
    setData([
      ...data.slice(0, index),
      { ...data[index], ...updatedObj },
      ...data.slice(index + 1)
    ]);
  };

  const handleDeleteClick = (index: number) => {
    setData([...data.slice(0, index), ...data.slice(index + 1)]);
  };

  const handleAddNew = () => {
    setData([...data, defaultPageData]);
  };

  const handleCancel = () => {
    if (!sixSignalConfig?.pages_exclude && !sixSignalConfig?.pages_include) {
      setMode('configure');
    } else {
      setMode('view');
    }
  };

  const renderData = () => {
    const rData = data.map((d, index) => (
      <div
        className={`flex w-100 items-center gap-2 ${index !== 0 ? 'mt-3' : ''}`}
        key={index}
      >
        <Select
          style={{
            width: 180,
            borderRadius: 6
          }}
          value={d.type}
          onChange={(value) => updateDataAtIndex(value, index, 'type')}
          options={[
            {
              value: 'equals',
              label: 'If the URL equals'
            },
            {
              value: 'contains',
              label: 'If the URL’s contains'
            }
          ]}
        />
        <Input
          size='middle'
          placeholder=''
          value={d.value}
          style={{ borderRadius: 6 }}
          onChange={(e) => updateDataAtIndex(e.target.value, index, 'value')}
          status={errors?.includes(index) ? 'error' : undefined}
        />

        <Button
          size='middle'
          shape='circle'
          type='text'
          onClick={() => handleDeleteClick(index)}
          icon={<MinusCircleOutlined style={{ color: '#8692A3' }} />}
        />
      </div>
    ));
    return rData;
  };

  const handleSaveClick = async () => {
    try {
      // verify data
      if (!projectId) return;

      let errorIndexes = [];
      const urlRegex =
        /^[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&//=]*)$/;
      let errorMessage = '';
      for (let i = 0; i < data.length; i++) {
        const currentData = data[i];
        if (currentData.type === 'contains') {
          if (typeof currentData.value !== 'string' || !currentData.value) {
            errorIndexes.push(i);
            errorMessage = 'Please enter a valid value';
          }
        }
        if (currentData.type === 'equals') {
          if (!currentData.value || !urlRegex.test(currentData.value)) {
            errorIndexes.push(i);
          }
          if (currentData.value.includes('http')) {
            errorMessage = 'Please enter URL without https';
          }
        }
      }
      if (errorIndexes.length > 0) {
        setErrors(errorIndexes);
        setErrorType(errorMessage || 'Please enter a valid URL');
        return;
      }
      // update local state
      let state: SixSignalConfigType = {};
      if (sixSignalConfig) state = { ...sixSignalConfig };
      if (enrichType === 'include') {
        state.pages_include = data;
        state.pages_exclude = undefined;
      } else if (enrichType === 'exclude') {
        state.pages_include = undefined;
        state.pages_exclude = data;
      }
      await udpateProjectSettings(projectId, {
        six_signal_config: state
      });

      setMode('view');
      notification.success({
        message: 'Success',
        description: `Successfully updated settings`,
        duration: 3
      });
    } catch (error) {
      console.error('Error in save changes', error);
    }
  };

  useEffect(() => {
    let data = null;
    if (
      sixSignalConfig?.pages_exclude &&
      sixSignalConfig.pages_exclude?.length > 0
    ) {
      setEnrichType('exclude');
      data = sixSignalConfig.pages_exclude;
    } else if (
      sixSignalConfig?.pages_include &&
      sixSignalConfig.pages_include?.length > 0
    ) {
      setEnrichType('include');
      data = sixSignalConfig.pages_include;
    }
    if (data) {
      setData(data);
    }
  }, [sixSignalConfig.pages_exclude, sixSignalConfig.pages_include, mode]);

  return (
    <div>
      {/* for edit mode */}
      {mode === 'edit' && (
        <>
          <div className={`mt-3 ${style.customRadioGroup}`}>
            <Radio.Group
              value={enrichType}
              onChange={(e) => setEnrichType(e.target.value)}
            >
              <Tooltip
                placement='topLeft'
                title='Enrich only for specific pages selected'
                color='#0B1E39'
              >
                <Radio.Button
                  value={'include'}
                  key={'include'}
                  disabled={enrichType === 'exclude' && data.length > 1}
                >
                  Include
                </Radio.Button>
              </Tooltip>
              <Tooltip
                placement='topLeft'
                title='Enrich for all pages except the selected ones'
                color='#0B1E39'
              >
                <Radio.Button
                  value={'exclude'}
                  key={'exclude'}
                  disabled={enrichType === 'include' && data.length > 1}
                >
                  Exclude
                </Radio.Button>
              </Tooltip>
            </Radio.Group>
          </div>
          <div className={`mt-5 ${style.customSelect}`}>
            {data && data?.length > 0 && renderData()}
          </div>
          {data.length < 50 && (
            <div className='mt-5'>
              <Button
                type='text'
                icon={<PlusOutlined style={{ color: '#8692A3' }} />}
                onClick={handleAddNew}
              >
                Add new
              </Button>
            </div>
          )}

          {errorType && (
            <div className={style.errorMessage}>
              <Text type={'paragraph'} mini>
                {errorType}
              </Text>
            </div>
          )}

          <div className=' flex items-center gap-2 mt-6'>
            <Button onClick={handleCancel}>Cancel</Button>
            <Button
              type='primary'
              disabled={!enrichType || !!errorType || !data.length}
              onClick={handleSaveClick}
            >
              Save changes
            </Button>
          </div>
        </>
      )}
      {/* for view mode  */}
      {mode === 'view' && (
        <>
          <div className='mt-3'>
            <Text type={'paragraph'} mini color='grey'>
              {enrichType === 'exclude' ? 'Exclude' : 'Include'}
            </Text>
          </div>
          {data?.length > 0 && (
            <div className='mt-5'>
              {data.map((d, i) => (
                <div
                  key={i}
                  className={`flex items-center gap-5  ${
                    i !== 0 ? 'mt-3' : ''
                  }`}
                >
                  <div style={{ width: 125 }}>
                    <Text type={'paragraph'} mini color='grey-2'>
                      {d.type === 'equals'
                        ? 'If the url equals'
                        : 'If the url contains'}
                    </Text>
                  </div>

                  <Text type={'paragraph'} mini weight={'bold'}>
                    {d.value}
                  </Text>
                </div>
              ))}
            </div>
          )}
        </>
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
type EnrichPage = {
  mode: FeatureModes;
  setMode: (value: FeatureModes) => void;
  sixSignalConfig: SixSignalConfigType;
  projectId: string;
};

type EnrichPagesProps = EnrichPage & ReduxProps;

export default connector(EnrichPages);
