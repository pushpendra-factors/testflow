import React from 'react';
import PropTypes from 'prop-types';
import cx from 'classnames';
import { Menu, Dropdown } from 'antd';
import map from 'lodash/map';
import noop from 'lodash/noop';

import { EMPTY_ARRAY, EMPTY_STRING } from 'Utils/global';

import SelectedItem from './SelectedItem';
import styles from './PivotTableControls.module.scss';
import ColumnsWrapper from './ColumnsWrapper';
import ControlledComponent from '../ControlledComponent';
// import { ArrowDownOutlined, ArrowUpOutlined } from '@ant-design/icons';
// import { PIVOT_SORT_ORDERS } from './pivotTableControls.constants';

const PivotTableControls = ({
  selectedRows,
  selectedCol,
  selectedValue,
  aggregatorOptions,
  columnOptions,
  rowOptions,
  functionOptions,
  aggregatorName,
  // rowOrder,
  onRowAttributeRemove,
  onColumnAttributeRemove,
  onValueChange,
  onColumnChange,
  onRowChange,
  onFunctionChange
}) => {
  const handleValueChange = (obj) => {
    onValueChange(obj.key);
  };

  const handleColumnChange = (obj) => {
    onColumnChange(obj.key);
  };

  const handleRowChange = (obj) => {
    onRowChange(obj.key);
  };

  const handleFunctionChange = (obj) => {
    onFunctionChange(obj.key);
  };

  const getMenu = ({ options, onOptionClick }) => {
    return (
      <Menu className={styles.dropdownMenu}>
        {map(options, (option) => {
          return (
            <Menu.Item onClick={onOptionClick} key={option}>
              {option}
            </Menu.Item>
          );
        })}
      </Menu>
    );
  };

  const aggregatorMenu = getMenu({
    options: aggregatorOptions,
    onOptionClick: handleValueChange
  });

  const columnsMenu = getMenu({
    options: columnOptions,
    onOptionClick: handleColumnChange
  });

  const rowsMenu = getMenu({
    options: rowOptions,
    onOptionClick: handleRowChange
  });

  const functionsMenu = getMenu({
    options: functionOptions,
    onOptionClick: handleFunctionChange
  });

  const renderDropdown = ({ dropdownMenu, label }) => {
    return (
      <Dropdown overlay={dropdownMenu}>
        <button className={styles.selectedAggregatorBtn}>{label}</button>
      </Dropdown>
    );
  };

  const renderRowsColumn = () => {
    return (
      <ColumnsWrapper heading="Rows">
        <div className="flex flex-col gap-y-2">
          {map(selectedRows, (row) => {
            return (
              <React.Fragment key={row}>
                <SelectedItem onRemove={onRowAttributeRemove} label={row} />
              </React.Fragment>
            );
          })}
          <ControlledComponent controller={rowOptions.length}>
            {renderDropdown({ dropdownMenu: rowsMenu, label: 'Select...' })}
          </ControlledComponent>
        </div>
      </ColumnsWrapper>
    );
  };

  const renderColumnsColumn = () => {
    return (
      <ColumnsWrapper heading="Columns">
        <div className="flex flex-col gap-y-2">
          <ControlledComponent controller={!!selectedCol}>
            <SelectedItem
              onRemove={onColumnAttributeRemove}
              label={selectedCol}
            />
          </ControlledComponent>

          <ControlledComponent controller={!selectedCol}>
            {renderDropdown({ dropdownMenu: columnsMenu, label: 'Select...' })}
          </ControlledComponent>
        </div>
      </ColumnsWrapper>
    );
  };

  const renderValuesColumn = () => {
    return (
      <ColumnsWrapper heading="Value">
        <div className="flex flex-col gap-y-2">
          {renderDropdown({
            dropdownMenu: aggregatorMenu,
            label: selectedValue
          })}
        </div>
      </ColumnsWrapper>
    );
  };

  return (
    <div className="flex flex-col">
      <div className={cx('flex border border-solid', styles.controls)}>
        <div className="w-1/3 py-5 px-10 border-r border-solid">
          {renderRowsColumn()}
        </div>

        <div className="w-1/3 py-5 px-10 border-r border-solid">
          {renderColumnsColumn()}
        </div>

        <div className="w-1/3 py-5 px-10">{renderValuesColumn()}</div>
      </div>

      <div className="flex border border-solid">
        <div className="w-1/3"></div>
        <div className="w-1/3"></div>
        <div className="w-1/3 py-5 px-10">
          <div className="flex gap-x-2 items-center">
            <span>Function:</span>
            {renderDropdown({
              dropdownMenu: functionsMenu,
              label: aggregatorName
            })}
          </div>
        </div>
      </div>
    </div>
  );
};

export default PivotTableControls;

PivotTableControls.propTypes = {
  selectedRows: PropTypes.array,
  selectedCol: PropTypes.string,
  selectedValue: PropTypes.string,
  aggregatorOptions: PropTypes.array,
  columnOptions: PropTypes.array,
  rowOptions: PropTypes.array,
  functionOptions: PropTypes.array,
  aggregatorName: PropTypes.string,
  onRowAttributeRemove: PropTypes.func,
  onColumnAttributeRemove: PropTypes.func,
  onColumnChange: PropTypes.func,
  onValueChange: PropTypes.func,
  onRowChange: PropTypes.func,
  onFunctionChange: PropTypes.func,
  onSortChange: PropTypes.func
};

PivotTableControls.defaultProps = {
  selectedRows: EMPTY_ARRAY,
  selectedCol: EMPTY_STRING,
  selectedValue: EMPTY_STRING,
  aggregatorOptions: EMPTY_ARRAY,
  columnOptions: EMPTY_ARRAY,
  rowOptions: EMPTY_ARRAY,
  functionOptions: EMPTY_ARRAY,
  aggregatorName: EMPTY_STRING,
  onRowAttributeRemove: noop,
  onColumnAttributeRemove: noop,
  onColumnChange: noop,
  onValueChange: noop,
  onRowChange: noop,
  onFunctionChange: noop,
  onSortChange: noop
};
