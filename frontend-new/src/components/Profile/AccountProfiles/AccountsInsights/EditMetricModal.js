import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Input } from 'antd';
import cx from 'classnames';
import AppModal from 'Components/AppModal';
import { Text } from 'Components/factorsComponents';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { useSelector } from 'react-redux';
import { cloneDeep, startCase } from 'lodash';
import { processProperties } from 'Utils/dataFormatter';
import { getNormalizedKpi } from 'Utils/kpiQueryComposer.helpers';
import styles from './index.module.scss';

function EditMetricModal({
  visible,
  onCancel,
  isLoading,
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
    if (kpiEvents != null) {
      kpiEvents.forEach((kpiEvent) => {
        const metricIndex = kpiEvent.values.findIndex(
          (elem) => elem.value === savedMetric
        );
        if (metricIndex > -1) {
          setSelectedMetric(cloneDeep(kpiEvent.values[metricIndex]));
        }
      });
    }
  }, [savedMetric, kpiEvents]);

  return (
    <AppModal
      okText='Update'
      visible={visible}
      onOk={handleOk}
      onCancel={onCancel}
      width={834}
      isLoading={isLoading}
    >
      <div className='flex flex-col gap-y-6'>
        <div className='flex flex-col gap-y-2'>
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
            Choose the KPI to map to this metric.
          </Text>
        </div>
        <div className='flex flex-col gap-y-5'>
          <div className='flex flex-col gap-y-2'>
            <Text type='title' extraClass='mb-0' color='character-primary'>
              Signal name
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

          <div className='border p-4'>
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
      </div>
    </AppModal>
  );
}

export default EditMetricModal;
