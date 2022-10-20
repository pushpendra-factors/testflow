import React, { Fragment, useCallback, useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { find, get, noop } from 'lodash';
import cx from 'classnames';
import { Button, Dropdown, Menu, Popover, Tooltip } from 'antd';
import { Text, SVG } from '../factorsComponents';
import { BUTTON_SIZES, BUTTON_TYPES } from '../../constants/buttons.constants';
import styles from './dataTableFilters.module.scss';
import ValuesMenu from './components/ValuesMenu';
import {
  EQUALITY_OPERATOR_MENU,
  CATEGORY_COMBINATION_OPERATOR_MENU,
  EQUALITY_OPERATOR_KEYS,
  CATEGORY_COMBINATION_OPERATOR_KEYS
} from './dataTableFilters.constants';
import {
  getUpdatedFiltersOnCategoryChange,
  getUpdatedFiltersOnEqualityOperatorChange,
  getUpdatedFiltersOnCategoryDelete,
  getUpdatedFiltersOnValueChange
} from './dataTableFilters.helpers';
import ControlledComponent from '../ControlledComponent/ControlledComponent';
import { isNumeric } from '../../utils/global';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';

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

  const handleValueChange = useCallback((categoryIndex, value) => {
    setSelectedFilters((currentFilters) => {
      const updatedFilters = getUpdatedFiltersOnValueChange({
        categoryIndex,
        value,
        currentFilters
      });
      return updatedFilters;
    });
  }, []);

  const handleNumericalFilterValueChange = useCallback(
    (categoryIndex, e) => {
      const value = e.target.value;
      if (!isNumeric(value) && value !== '') {
        return false;
      }
      setSelectedFilters((currentFilters) => {
        const selectedCategoryField = get(
          selectedFilters,
          `categories.${categoryIndex}.field`
        );
        const selectedCategoryFieldValueType = get(
          find(filters, (filter) => filter.key === selectedCategoryField),
          'valueType',
          'numerical'
        );
        const updatedFilters = getUpdatedFiltersOnValueChange({
          categoryIndex,
          value,
          currentFilters,
          categoryFieldType: selectedCategoryFieldValueType
        });
        return updatedFilters;
      });
    },
    [selectedFilters, filters]
  );

  const handleCategoryCombinationOperatorChange = useCallback((option) => {
    setSelectedFilters((currentFilters) => {
      return {
        ...currentFilters,
        categoryCombinationOperator: option.key
      };
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

  const getCategoryCombinationOperatorMenu = () => {
    return (
      <Menu className={styles.menu}>
        {CATEGORY_COMBINATION_OPERATOR_MENU.map((option) => {
          return (
            <Menu.Item
              className={styles['dropdown-item']}
              key={option.key}
              onClick={handleCategoryCombinationOperatorChange}
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

  const getEqualityOperatorMenu = (categoryIndex) => {
    const selectedCategoryField = get(
      selectedFilters,
      `categories.${categoryIndex}.field`
    );
    const selectedCategoryFieldValueType = get(
      find(filters, (filter) => filter.key === selectedCategoryField),
      'valueType',
      'categorical'
    );
    return (
      <Menu className={styles.menu}>
        {EQUALITY_OPERATOR_MENU.filter((option) =>
          option.valueType.includes(selectedCategoryFieldValueType)
        ).map((option) => {
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

  const getCategoryValuesMenu = (
    options,
    selectedOptions,
    equalityOperator,
    categoryIndex
  ) => {
    return (
      <ValuesMenu
        options={options}
        selectedOptions={selectedOptions}
        onChange={handleValueChange.bind(null, categoryIndex)}
        equalityOperator={equalityOperator}
      />
    );
  };

  const renderCategoryCombinationDropdown = (index) => {
    const showDropdown =
      selectedFilters.categories != null &&
      selectedFilters.categories.length > 1 &&
      index === 0;

    const showPlainText =
      selectedFilters.categories != null &&
      selectedFilters.categories.length > 1 &&
      index > 0 &&
      index < selectedFilters.categories.length - 1;

    return (
      <Fragment>
        <ControlledComponent controller={showDropdown}>
          <Dropdown
            overlayClassName="rounded-lg w-20"
            trigger="click"
            overlay={getCategoryCombinationOperatorMenu()}
          >
            <Button
              className="flex items-center"
              disabled
              type={BUTTON_TYPES.PLAIN}
            >
              <Text type="title" level={7}>
                {selectedFilters.categoryCombinationOperator}
              </Text>
              <SVG size={14} name="chevronDown" />
            </Button>
          </Dropdown>
        </ControlledComponent>
        <ControlledComponent controller={showPlainText}>
          <Button
            className={styles['disabled-button']}
            disabled
            type={BUTTON_TYPES.PLAIN}
          >
            <Text type="title" level={7}>
              {selectedFilters.categoryCombinationOperator}
            </Text>
          </Button>
        </ControlledComponent>
      </Fragment>
    );
  };

  const renderSelectedFilters = () => {
    if (selectedFilters.categories == null) {
      return null;
    }
    return (
      <div className="flex flex-col row-gap-2">
        {selectedFilters.categories.map((category, index) => {
          const selectedCategoryField = get(category, `field`);
          const selectedCategoryFieldValueType = get(
            find(filters, (filter) => filter.key === selectedCategoryField),
            'valueType',
            'categorical'
          );
          const filterDetail = filters.find(
            (filter) => filter.key === category.field
          );
          const equalityOperator = EQUALITY_OPERATOR_MENU.find(
            (option) => option.key === category.equalityOperator
          ).title;
          const label = filterDetail.title;
          const options = filterDetail.options;
          const filterValue = category.values;

          const valuesLabel =
            Array.isArray(filterValue) && filterValue.length === 0
              ? 'Select values'
              : Array.isArray(filterValue)
              ? filterValue.join(', ')
              : '';

          return (
            <div className="flex col-gap-1 items-center">
              <div key={category.key} className="flex col-gap-1 items-center">
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
                {selectedCategoryFieldValueType === 'categorical' && (
                  <Popover
                    overlayClassName={styles['values-popover']}
                    trigger="click"
                    placement="bottomRight"
                    content={getCategoryValuesMenu.bind(
                      null,
                      options,
                      filterValue,
                      category.equalityOperator,
                      index
                    )}
                  >
                    <Tooltip title={valuesLabel}>
                      {renderLabelButton({
                        label: valuesLabel
                      })}
                      color={TOOLTIP_CONSTANTS.DARK}
                    </Tooltip>
                  </Popover>
                )}
                {(selectedCategoryFieldValueType === 'numerical' ||
                  selectedCategoryFieldValueType === 'percentage') && (
                  <>
                    <input
                      onChange={handleNumericalFilterValueChange.bind(
                        null,
                        index
                      )}
                      value={filterValue}
                      className={styles['value-input-box']}
                      type="text"
                    />
                    {selectedCategoryFieldValueType === 'percentage' && (
                      <span>%</span>
                    )}
                  </>
                )}
                {renderCrossIcon(index)}
              </div>
              {renderCategoryCombinationDropdown(index)}
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
        equalityOperator: PropTypes.oneOf(
          Object.values(EQUALITY_OPERATOR_KEYS)
        ),
        field: PropTypes.string,
        key: PropTypes.number
      })
    ),
    categoryCombinationOperator: PropTypes.oneOf(
      Object.values(CATEGORY_COMBINATION_OPERATOR_KEYS)
    )
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
