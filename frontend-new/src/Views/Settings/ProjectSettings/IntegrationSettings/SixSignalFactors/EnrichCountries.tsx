import React, { useEffect, useState, useRef } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  FeatureModes,
  EnrichTypes,
  SixSignalConfigType,
  EnrichCountryData,
  CountryLabel
} from './types';

import { Button, notification, Radio, Select, Tooltip } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { useSelector } from 'react-redux';
import { fetchEventPropertyValues } from 'Reducers/coreQuery/services';
import { getCountryCode, isCountryFlagAvailable } from 'Utils/country';
import { Text } from 'Components/factorsComponents';
import { udpateProjectSettings } from 'Reducers/global';
import style from './index.module.scss';

const EnrichCountries = ({
  mode,
  setMode,
  sixSignalConfig,
  projectId,
  udpateProjectSettings
}: EnrichCountriesProps) => {
  const [enrichType, setEnrichType] = useState<EnrichTypes | null>(null);
  const [countryOptions, setCountryOptions] = useState<CountryLabel[]>([]);
  const [data, setData] = useState<CountryLabel[]>([]);
  const countriesSet = useRef(false);
  const { active_project } = useSelector((state) => state.global);

  const handleAddNew = () => {
    if (countryOptions && countryOptions?.length > 0) {
      const countryOption = countryOptions[0];
      setData([
        ...data,
        { value: countryOption.value, label: countryOption.label }
      ]);
    }
  };

  const handleDeleteClick = (index: number) => {
    setData([...data.slice(0, index), ...data.slice(index + 1)]);
  };

  const handleSelectChange = (value: CountryLabel, index: number) => {
    setData([...data.slice(0, index), value, ...data.slice(index + 1)]);
  };

  const renderOption = (country: string) => {
    const isFlagAvailable = isCountryFlagAvailable(country);
    return (
      <div className='flex items-center gap-2 justify-start mt-1'>
        {isFlagAvailable && (
          <div className={`fflag fflag-${getCountryCode(country)} ff-md`}></div>
        )}
        <div className='flex-1 whitespace-nowrap overflow-hidden text-ellipsis'>
          <Text type={'paragraph'} mini ellipsis>
            {' '}
            {country}
          </Text>
        </div>
      </div>
    );
  };

  const renderData = () => {
    return data.map((country, index) => {
      return (
        <div
          className={`flex w-100 items-center gap-2 ${
            index !== 0 ? 'mt-3' : ''
          }`}
          key={index}
        >
          <Select
            style={{
              borderRadius: 6,
              width: 'fix-content',
              minWidth: 215
            }}
            className={style.countrySelectContainer}
            labelInValue
            value={country}
            showSearch
            onSelect={(labelInValue: CountryLabel) =>
              handleSelectChange(labelInValue, index)
            }
            options={countryOptions}
          ></Select>
          <Button
            size='middle'
            shape='circle'
            type='text'
            onClick={() => handleDeleteClick(index)}
            icon={<MinusCircleOutlined style={{ color: '#8692A3' }} />}
          />
        </div>
      );
    });
  };

  const handleCancel = () => {
    if (sixSignalConfig?.country_exclude || sixSignalConfig?.country_include) {
      setMode('view');
    } else {
      setMode('configure');
    }
  };

  const handleSaveClick = async () => {
    try {
      if (!projectId) return;
      // update local state
      let state: SixSignalConfigType = {};
      if (sixSignalConfig) state = { ...sixSignalConfig };
      const updatedData: EnrichCountryData[] = data.map((d) => {
        return {
          value: d.value,
          type: 'equals'
        };
      });
      if (
        new Set(updatedData?.map((d) => d.value)).size !== updatedData.length
      ) {
        notification.error({
          message: 'Error',
          description: `Please remove duplicate countries`,
          duration: 3
        });
        return;
      }
      if (enrichType === 'include') {
        state.country_include = updatedData;
        state.country_exclude = undefined;
      } else if (enrichType === 'exclude') {
        state.country_include = undefined;
        state.country_exclude = updatedData;
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
    let _data = null;
    if (
      sixSignalConfig?.country_exclude &&
      sixSignalConfig.country_exclude.length > 0
    ) {
      setEnrichType('exclude');
      _data = sixSignalConfig?.country_exclude;
    } else if (
      sixSignalConfig?.country_include &&
      sixSignalConfig.country_include.length > 0
    ) {
      setEnrichType('include');
      _data = sixSignalConfig?.country_include;
    }
    if (_data) {
      const selectedValues = _data.map((d) => {
        return {
          value: d?.value,
          label: renderOption(d?.value)
        };
      });
      setData(selectedValues);
      countriesSet.current = true;
    }
  }, [
    sixSignalConfig?.country_exclude,
    sixSignalConfig?.country_include,
    mode
  ]);

  useEffect(() => {
    //fetching country list
    const fetchCountries = async () => {
      const res = await fetchEventPropertyValues(
        active_project?.id,
        '$session',
        '$country'
      );
      if (res.ok && res?.data) {
        const data = res?.data;
        if (typeof data !== 'object' || !data) return;
        const countryList = Object.keys(data);
        const countryListWithLabels = countryList.map((country: string) => ({
          value: country,
          label: renderOption(data[country])
        }));
        setCountryOptions(countryListWithLabels);
        if (!countriesSet.current) {
          const firstCountry = countryList[0];
          if (firstCountry)
            setData([
              { value: firstCountry, label: renderOption(data[firstCountry]) }
            ]);
        }
      }
    };
    if (active_project?.id) fetchCountries();
  }, [active_project?.id]);

  return (
    <div className={style.customSelect}>
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
                title='Enrich only for specific countries selected'
                color='#0B1E39'
              >
                <Radio.Button value={'include'} key={'include'}>
                  Include
                </Radio.Button>
              </Tooltip>
              <Tooltip
                placement='topLeft'
                title='Enrich for all countries except the selected countries'
                color='#0B1E39'
              >
                <Radio.Button value={'exclude'} key={'exclude'}>
                  Exclude
                </Radio.Button>
              </Tooltip>
            </Radio.Group>
          </div>
          <div className='mt-5'>{data && data?.length > 0 && renderData()}</div>
          {data.length < 50 && data.length < countryOptions.length && (
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

          <div className=' flex items-center gap-2 mt-6'>
            <Button onClick={handleCancel}>Cancel</Button>
            <Button
              type='primary'
              disabled={!enrichType || !data.length}
              onClick={handleSaveClick}
            >
              Save changes
            </Button>
          </div>
        </>
      )}

      {/* for view mode */}
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
                  {renderOption(d?.value)}
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
type EnrichCountry = {
  mode: FeatureModes;
  setMode: (value: FeatureModes) => void;
  sixSignalConfig: SixSignalConfigType | null;
  projectId: string;
};

type EnrichCountriesProps = EnrichCountry & ReduxProps;

export default connector(EnrichCountries);
