import React, { useCallback, useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { noop } from 'lodash';
import cx from 'classnames';
import { Button, Dropdown, Menu, Popover } from 'antd';
import { Text, SVG } from '../factorsComponents';
import { BUTTON_SIZES, BUTTON_TYPES } from '../../constants/buttons.constants';
import styles from './dataTableFilters.module.scss';
import ValuesMenu from './components/ValuesMenu';
import { EQUALITY_OPERATOR_MENU } from './dataTableFilters.constants';
import {
  getUpdatedFiltersOnCategoryChange,
  getUpdatedFiltersOnEqualityOperatorChange,
  getUpdatedFiltersOnCategoryDelete,
  getUpdatedFiltersOnValueChange
} from './dataTableFilters.helpers';

const DataTableFilters = ({
  filters,
  appliedFilters,
  setAppliedFilters,
  setFiltersVisibility
}) => {
  const [selectedFilters, setSelectedFilters] = useState({});

  useEffect(() => {
    setSelectedFilters(appliedFilters);
  }, [appliedFilters]);

  const handleCategoryChange = (index, option) => {
    setSelectedFilters((currentFilters) => {
      const updatedFilters = getUpdatedFiltersOnCategoryChange({
        categoryIndex: index,
        selectedCategoryValue: option.key,
        currentFilters
      });
      return updatedFilters;
    });
  };

  const handleEqualityOperatorChange = (index, option) => {
    setSelectedFilters((currentFilters) => {
      const updatedFilters = getUpdatedFiltersOnEqualityOperatorChange({
        categoryIndex: index,
        selectedOperator: option.key,
        currentFilters
      });
      return updatedFilters;
    });
  };

  const handleCategoryDelete = (categoryIndex) => {
    setSelectedFilters((currentFilters) => {
      const updatedFilters = getUpdatedFiltersOnCategoryDelete({
        categoryIndex,
        currentFilters
      });
      return updatedFilters;
    });
  };

  const handleValueChange = useCallback((categoryIndex, toggledOption) => {
    setSelectedFilters((currentFilters) => {
      const updatedFilters = getUpdatedFiltersOnValueChange({
        categoryIndex,
        toggledOption,
        currentFilters
      });
      return updatedFilters;
    });
  }, []);

  const handleFiltersApply = useCallback(() => {
    setAppliedFilters(selectedFilters);
  }, [selectedFilters, setAppliedFilters]);

  const getCategoryMenu = (categoryIndex) => {
    return (
      <Menu className={styles.menu}>
        {filters.map((filter) => {
          return (
            <Menu.Item
              className={styles['dropdown-item']}
              key={filter.key}
              onClick={handleCategoryChange.bind(null, categoryIndex)}
            >
              <Text type="title" level={7} color="grey-6">
                {filter.title}
              </Text>
            </Menu.Item>
          );
        })}
      </Menu>
    );
  };

  const getEqualityOperatorMenu = (categoryIndex) => {
    return (
      <Menu className={styles.menu}>
        {EQUALITY_OPERATOR_MENU.map((option, index) => {
          return (
            <Menu.Item
              className={styles['dropdown-item']}
              key={option.key}
              onClick={handleEqualityOperatorChange.bind(null, categoryIndex)}
            >
              <Text type="title" level={7} color="grey-6">
                {option.title}
              </Text>
            </Menu.Item>
          );
        })}
      </Menu>
    );
  };

  const renderLabelButton = ({ label, leftRounded = false }) => {
    return (
      <Button
        className={cx(
          'flex col-gap-1 items-center shadow-none',
          styles['label-button'],
          {
            [styles['label-button-left-rounded']]: leftRounded
          }
        )}
        type={BUTTON_TYPES.SECONDARY}
      >
        <Text type="title" weight="medium" level={7}>
          {label}
        </Text>
      </Button>
    );
  };

  const renderCrossIcon = (categoryIndex) => {
    return (
      <Button
        className={cx(
          'flex col-gap-1 items-center shadow-none',
          styles['label-button'],
          styles['label-button-right-rounded']
        )}
        type={BUTTON_TYPES.SECONDARY}
        onClick={handleCategoryDelete.bind(null, categoryIndex)}
      >
        <SVG name="remove" />
      </Button>
    );
  };

  const categoryValuesMenu = (options, selectedOptions, categoryIndex) => {
    return (
      <ValuesMenu
        options={options}
        selectedOptions={selectedOptions}
        onChange={handleValueChange.bind(null, categoryIndex)}
      />
    );
  };

  const renderSelectedFilters = () => {
    if (selectedFilters.categories == null) {
      return null;
    }
    return (
      <div className="flex flex-col row-gap-2">
        {selectedFilters.categories.map((category, index) => {
          const filterDetail = filters.find(
            (filter) => filter.key === category.field
          );
          const equalityOperator = EQUALITY_OPERATOR_MENU.find(
            (option) => option.key === category.equalityOperator
          ).title;
          const label = filterDetail.title;
          const options = filterDetail.options;
          const selectedOptions = category.values;

          return (
            <div key={category.key} className="flex col-gap-px items-center">
              <Dropdown
                overlayClassName="rounded-lg"
                trigger="click"
                overlay={getCategoryMenu(index)}
              >
                {renderLabelButton({ label, leftRounded: true })}
              </Dropdown>
              <Dropdown
                overlayClassName="rounded-lg"
                trigger="click"
                overlay={getEqualityOperatorMenu(index)}
              >
                {renderLabelButton({ label: equalityOperator })}
              </Dropdown>
              <Popover
                overlayClassName={styles['values-popover']}
                trigger="click"
                placement="bottomRight"
                content={categoryValuesMenu.bind(
                  null,
                  options,
                  selectedOptions,
                  index
                )}
              >
                {renderLabelButton({
                  label:
                    selectedOptions.length === 0
                      ? 'Select values'
                      : selectedOptions.join(', ')
                })}
              </Popover>
              {renderCrossIcon(index)}
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="flex flex-col row-gap-3">
      <Text type="title" color="grey-2" level={7}>
        Filter if
      </Text>
      {renderSelectedFilters()}
      <Dropdown
        overlayClassName="rounded-lg"
        trigger="click"
        overlay={getCategoryMenu()}
      >
        <Button
          type={BUTTON_TYPES.PLAIN}
          className={cx(
            'flex col-gap-1 items-center',
            styles['add-filters-button']
          )}
        >
          <SVG name="plus" color="#8692A3" />
          <Text type="title" color="grey" level={7}>
            Add condition
          </Text>
        </Button>
      </Dropdown>
      <div className="flex justify-end col-gap-2">
        <Button
          onClick={setFiltersVisibility.bind(null, false)}
          size={BUTTON_SIZES.MEDIUM}
          type={BUTTON_TYPES.SECONDARY}
        >
          Cancel
        </Button>
        <Button
          onClick={handleFiltersApply}
          size={BUTTON_SIZES.MEDIUM}
          type={BUTTON_TYPES.PRIMARY}
        >
          Apply
        </Button>
      </div>
    </div>
  );
};

export default DataTableFilters;

DataTableFilters.propTypes = {
  filters: PropTypes.arrayOf(
    PropTypes.shape({
      title: PropTypes.string,
      key: PropTypes.string,
      options: PropTypes.arrayOf(PropTypes.string)
    })
  ),
  appliedFilters: PropTypes.shape({
    categories: PropTypes.arrayOf(
      PropTypes.shape({
        values: PropTypes.arrayOf(PropTypes.string),
        equalityOperator: PropTypes.oneOf([
          'equal',
          'not-equal',
          'contains',
          'does-not-contain'
        ]),
        field: PropTypes.string,
        key: PropTypes.number
      })
    ),
    categoryCombinationOperator: PropTypes.oneOf(['AND', 'OR'])
  }),
  setAppliedFilters: PropTypes.func,
  setFiltersVisibility: PropTypes.func
};

DataTableFilters.defaultProps = {
  filters: [],
  appliedFilters: {},
  setAppliedFilters: noop,
  setFiltersVisibility: noop
};
