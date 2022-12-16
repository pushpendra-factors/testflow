import React, { Fragment, memo, useState } from 'react';
import PropTypes from 'prop-types';
import { DatePicker, Menu, Dropdown, Button, Radio } from 'antd';
import { noop } from 'lodash';
import classNames from 'classnames';
import MomentTz from '../MomentTz';
import ControlledComponent from '../ControlledComponent';
import { BUTTON_TYPES } from '../../constants/buttons.constants';
import { SVG, Text } from '../factorsComponents';
import styles from './comparisonDatePicker.module.scss';
import {
  COMPARISON_DATE_RANGE_TYPE,
  SELECTABLE_OPTIONS,
  SELECTABLE_OPTIONS_KEYS
} from './comparisonDatePicker.constants';

function ComparisonDatePicker({
  comparisonLabel,
  placement,
  value,
  onChange,
  onRemoveClick
}) {
  const [visible, setVisible] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [customRangeType, setCustomRangeType] = useState(
    COMPARISON_DATE_RANGE_TYPE.START_DATE
  );

  const removeClicked = (event) => {
    onRemoveClick();
    event.stopPropagation();
  };

  const handleVisibleChange = () => {
    if (!showDatePicker) {
      setVisible((curr) => !curr);
    }
  };

  const handleShowCustomRangePickerClick = () => {
    setShowDatePicker(true);
    handleVisibleChange();
  };

  const displayRange = () => {
    if (value == null) {
      return null;
    }

    const rangeText = `${MomentTz(value.from).format(
      'MMM DD, YYYY'
    )} - ${MomentTz(value.to).format('MMM DD, YYYY')}`;

    return (
      <Text
        extraClass={`${styles.label}`}
        type='title'
        level={6}
        color='grey-2'
      >
        {rangeText}
      </Text>
    );
  };

  const getMenu = () => {
    return (
      <Menu>
        <Menu.Item key={'compare label'} className={styles.noHoverEffect}>
          <Text
            extraClass={`${styles.label}`}
            weight='bold'
            color='grey-6'
            level={6}
            type='title'
          >
            Compare with:
          </Text>
        </Menu.Item>
        <Menu.Divider />
        {SELECTABLE_OPTIONS.map((option) => (
          <Menu.Item
            onClick={() => {
              onChange({ value: option.value, isPreset: true });
              setVisible(false);
            }}
            key={option.value}
          >
            <Text type='title' color='grey-2' level={6}>
              {option.label}
            </Text>
          </Menu.Item>
        ))}
        <Menu.Divider />
        <Menu.Item
          onClick={handleShowCustomRangePickerClick}
          className={classNames(
            styles.selectCustomDateListItem,
            styles.noHoverEffect
          )}
          key={'date-picker'}
        >
          <div
            className={classNames(
              'flex justify-between items-center',
              styles.customDateListItemWrapper
            )}
            role='button'
          >
            <Text type='title' color='grey-3' level={6}>
              Select Date
            </Text>
            <SVG name='calendar' color='#B7BEC8' size={16} />
          </div>
        </Menu.Item>
      </Menu>
    );
  };

  const handleDateChange = (value) => {
    setShowDatePicker(false);
    onChange({ value, isPreset: false, customRangeType });
  };

  return (
    <div className='fa-custom-datepicker'>
      <Dropdown
        overlayClassName='fa-custom-datepicker--dropdown'
        overlay={getMenu()}
        placement={placement}
        trigger={['click']}
        visible={visible}
        onVisibleChange={handleVisibleChange}
      >
        <Button
          type={BUTTON_TYPES.SECONDARY}
          className='flex items-center col-gap-1'
        >
          <SVG name='compare' size={16} />
          <ControlledComponent controller={!showDatePicker && !!value}>
            <Fragment>
              {displayRange()}
              <div onClick={removeClicked}>
                <SVG name='removeOutlined' size={16} />
              </div>
            </Fragment>
          </ControlledComponent>
          <ControlledComponent controller={!showDatePicker && !value}>
            <Text
              weight='medium'
              extraClass={`${styles.label}`}
              color='grey-2'
              level={7}
              type='title'
            >
              {comparisonLabel}
            </Text>
          </ControlledComponent>
          {showDatePicker && (
            <DatePicker
              disabledDate={(d) => !d || d.isAfter(MomentTz())}
              autoFocus
              open
              size='small'
              suffixIcon={null}
              showToday={false}
              bordered={false}
              allowClear
              onChange={handleDateChange}
              panelRender={(panelNode) => {
                return (
                  <div className='py-4 flex flex-col row-gap-2'>
                    <div className='px-3 flex flex-col row-gap-2'>
                      <Text type='title' weight='bold' level={6} color='grey-2'>
                        Compare to date
                      </Text>
                      <Radio.Group
                        className={styles['custom-range-radio-group']}
                        value={customRangeType}
                        onChange={(e) => setCustomRangeType(e.target.value)}
                      >
                        <Radio value={COMPARISON_DATE_RANGE_TYPE.START_DATE}>
                          Starts on
                        </Radio>
                        <Radio value={COMPARISON_DATE_RANGE_TYPE.END_DATE}>
                          Ends on
                        </Radio>
                      </Radio.Group>
                    </div>
                    {panelNode}
                  </div>
                );
              }}
            />
          )}
        </Button>
      </Dropdown>
    </div>
  );
}

export { SELECTABLE_OPTIONS_KEYS };

export default memo(ComparisonDatePicker);

ComparisonDatePicker.propTypes = {
  comparisonLabel: PropTypes.string,
  placement: PropTypes.oneOf([
    'bottomLeft',
    'bottom',
    'bottomRight',
    'topLeft',
    'top',
    'topRight'
  ]),
  value: PropTypes.object,
  onChange: PropTypes.func,
  onRemoveClick: PropTypes.func
};

ComparisonDatePicker.defaultProps = {
  comparisonLabel: 'Compare',
  placement: 'bottom',
  value: null,
  onChange: noop,
  onRemoveClick: noop
};
