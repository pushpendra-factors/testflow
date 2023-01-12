import React from 'react';
import cx from 'classnames';
import { Button, Select } from 'antd';
import { Text } from '../../factorsComponents';
import styles from './FunnelsConversionDurationBlock.module.scss';

const UNIT_OPTIONS = [
  {
    label: 'Days',
    value: 'D'
  },
  {
    label: 'Hours',
    value: 'H'
  },
  {
    label: 'Minutes',
    value: 'M'
  }
];

const FunnelsConversionDurationBlockContent = ({
  funnelConversionDurationNumber,
  funnelConversionDurationUnit,
  onNumberChange,
  onUnitChange,
  onApply
}) => {
  const handleNumberChange = (e) => {
    const value = e.target.value;
    if (value.match(/^\d+$/) || value.length === 0) {
      onNumberChange(value);
    }
  };

  const handleUnitChange = (value) => {
    onUnitChange(value);
  };

  return (
    <div className='p-4 flex flex-col gap-y-6'>
      <div className='flex flex-col gap-y-1'>
        <Text
          extraClass='mb-0 text-with-no-margin'
          level={6}
          color='grey-2'
          weight='bold'
          type='title'
        >
          Conversion Window
        </Text>
        <Text
          extraClass={cx('mb-0 text-with-no-margin', styles.subtext)}
          level={5}
          color='grey'
          type='paragraph'
        >
          The window of time a user has to complete all steps once they enter
          the funnel
        </Text>
      </div>
      <div className='flex gap-x-4'>
        <input
          className={cx(
            'py-1 px-3 border rounded w-1/3',
            styles['input-border']
          )}
          onChange={handleNumberChange}
          value={funnelConversionDurationNumber}
        />
        <Select
          onChange={handleUnitChange}
          className={cx(
            'flex-1 py-1 px-3 rounded shadow-none',
            styles['input-border'],
            styles['unit-dropdown']
          )}
          value={funnelConversionDurationUnit}
        >
          {UNIT_OPTIONS.map((option) => {
            return (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            );
          })}
        </Select>
      </div>
      <Button onClick={onApply} type='primary'>
        Apply
      </Button>
    </div>
  );
};

export default FunnelsConversionDurationBlockContent;
