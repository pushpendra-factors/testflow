import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { cloneDeep, startCase } from 'lodash';
import { Button, Input } from 'antd';
import cx from 'classnames';
import AppModal from 'Components/AppModal';
import { SVG, Text } from 'Components/factorsComponents';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { processProperties } from 'Utils/dataFormatter';
import { getNormalizedKpi } from 'Utils/kpiQueryComposer.helpers';
import { PathUrls } from 'Routes/pathUrls';
import styles from './index.module.scss';

const getDataTypeNote = (metricDataType) => {
  if (metricDataType === 'currency') {
    return '* this metric will be shown in USD currency format. Sample data - $2.4k';
  }
  if (metricDataType === 'percentage') {
    return '* this metric will be shown in % format. Sample data - 24%';
  }
  if (metricDataType === 'duration') {
    return '* this metric will be shown in time format. Sample data - 29d 12h';
  }
  return '* this metric will be shown in number format. Sample data - 15';
};

function EditMetricModal({
  visible,
  onCancel,
  isLoading,
  metricDataType,
  savedMetricName,
  savedMetric,
  onSave
}) {
  const [metricName, setMetricName] = useState('');
  const [selectedMetric, setSelectedMetric] = useState(null);
  const [isDDVisible, setDDVisible] = useState(false);
  const eventNames = useSelector((state) => state.coreQuery.eventNames);
  const kpi = useSelector((state) => state.kpi);

  const kpiEvents = useMemo(
    () =>
      kpi?.config
        ?.filter(
          (item) =>
            item.display_category.toLowerCase().includes('hubspot') ||
            item.display_category.toLowerCase().includes('salesforce')
        )
        ?.map((item) => getNormalizedKpi({ kpi: item }))
        ?.map((groupOpt) => ({
          iconName: groupOpt?.icon,
          label: startCase(groupOpt?.label),
          value: groupOpt?.label,
          extraProps: {
            category: groupOpt?.category
          },
          values: processProperties(groupOpt?.values)
        })),
    [kpi?.config]
  );

  const handleOk = () => {
    onSave(selectedMetric, metricName);
  };

  const onChange = useCallback((...props) => {
    setSelectedMetric(props[0]);
    setDDVisible(false);
  }, []);

  const handleNameChange = useCallback((e) => {
    setMetricName(e.target.value);
  }, []);

  useEffect(() => {
    setMetricName(savedMetricName);
  }, [savedMetricName]);

  useEffect(() => {
    if (kpiEvents != null && visible === true) {
      kpiEvents.forEach((kpiEvent) => {
        const metricIndex = kpiEvent.values.findIndex(
          (elem) => elem.value === savedMetric
        );
        if (metricIndex > -1) {
          setSelectedMetric(cloneDeep(kpiEvent.values[metricIndex]));
        }
      });
    }
  }, [savedMetric, kpiEvents, visible]);

  return (
    <AppModal visible={visible} onCancel={onCancel} width={600} footer={null}>
      <div className='flex flex-col gap-y-6'>
        <div className='flex flex-col'>
          <Text
            type='title'
            extraClass='mb-0'
            weight='bold'
            level={4}
            color='character-primary'
          >
            Manage Mapping
          </Text>
          <Text type='title' extraClass='mb-0' color='character-secondary'>
            Choose an existing KPI definition to map to {savedMetricName}.
          </Text>
        </div>
        <div className='flex gap-x-6 items-center'>
          <div className='flex flex-col gap-y-2'>
            <Text type='title' extraClass='mb-0' color='character-secondary'>
              Metric
            </Text>
            <Input
              onChange={handleNameChange}
              value={metricName}
              className={cx('fa-input', styles.input)}
              size='large'
              placeholder='Name'
              disabled
            />
          </div>
          <div className={cx('relative', styles['top-3'])}>
            <SVG name='arrowsLeftRight' color='#8692A3' size={16} />
          </div>

          <div className='flex flex-col gap-y-2'>
            <Text type='title' extraClass='mb-0' color='character-secondary'>
              KPI Definition
            </Text>
            <Button
              className='fa-button--truncate fa-button--truncate-lg btn-total-round'
              type='link'
              onClick={() => setDDVisible(true)}
            >
              {selectedMetric != null
                ? eventNames[selectedMetric.label] ?? selectedMetric.label
                : ''}
            </Button>
            {isDDVisible ? (
              <div className={styles['kpi-dropdown-container']}>
                <GroupSelect
                  options={kpiEvents ?? []}
                  optionClickCallback={onChange}
                  onClickOutside={() => setDDVisible(false)}
                  allowSearch
                  extraClass={styles['kpi-dropdown']}
                  allowSearchTextSelection={false}
                />
              </div>
            ) : null}
          </div>
        </div>
        <Text
          type='title'
          extraClass='mb-0'
          weight='medium'
          color='character-secondary'
        >
          {getDataTypeNote(metricDataType)}
        </Text>
        <div className={cx('pt-4', styles['border-t'])}>
          <div className='flex justify-between items-center'>
            <Link
              className='flex items-center gap-x-1'
              to={PathUrls.ConfigureCustomKpi}
            >
              <SVG size={16} color='#1890FF' name='arrowUpRightSquare' />
              <Text
                type='title'
                extraClass='mb-0'
                color='brand-color-6'
                weight='medium'
              >
                Create a new definition
              </Text>
            </Link>
            <div className='flex gap-x-2 items-center'>
              <Button onClick={onCancel}>Cancel</Button>
              <Button
                disabled={savedMetric === selectedMetric?.value}
                loading={isLoading}
                onClick={handleOk}
              >
                Update
              </Button>
            </div>
          </div>
        </div>
      </div>
    </AppModal>
  );
}

export default EditMetricModal;
