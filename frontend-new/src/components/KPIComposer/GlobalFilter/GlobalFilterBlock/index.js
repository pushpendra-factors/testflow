import React, { useState, useEffect } from 'react';
import { useSelector, connect } from 'react-redux';
import styles from './index.module.scss';
import { DateRangePicker } from 'react-date-range';
import { Input, Button } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { DEFAULT_DATE_RANGE } from 'Components/QueryComposer/DateRangeSelector/utils';
import MomentTz from 'Components/MomentTz';

import GlobalFilterSelect from '../GlobalFilterSelect';
import { DEFAULT_OPERATOR_PROPS } from '../../../FaFilterSelect/utils';
import { getKPIPropertyValues } from 'Reducers/coreQuery/middleware';
import _ from 'lodash';
import { groupKPIPropertiesOnCategory } from 'Utils/dataFormatter';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;

function GlobalFilterBlock({
  index,
  refValue,
  blockType = 'event',
  filterType = 'analytics',
  typeProps,
  filterProps,
  activeProject,
  operatorProps = defaultOpProps,
  event,
  filter,
  delIcon = 'remove',
  propsConstants = ['user', 'event'],
  extraClass,
  delBtnClass,
  deleteFilter,
  insertFilter,
  closeFilter,
  getKPIPropertyValues,
  selectedMainCategory,
  showOr,
  viewMode = false,
  isSameKPIGrp
}) {
  const propertyValuesMap = useSelector(
    (state) => state.coreQuery.propertyValuesMap
  );
  const [filterTypeState, setFilterTypeState] = useState('props');
  const [groupCollapseState, setGroupCollapse] = useState({});
  const [searchTerm, setSearchTerm] = useState('');
  const [newFilterState, setNewFilterState] = useState({
    props: [],
    operator: '',
    values: []
  });

  const [dropDownValues, setDropDownValues] = useState({});
  const [valueOptsLoading, setvalueOptsLoading] = useState(false);

  const [selectedRngState, setSelectedRngState] = useState([
    { ...DEFAULT_DATE_RANGE }
  ]);

  const placeHolder = {
    props: 'Choose a property',
    operator: 'Choose an operator',
    values: 'Choose values'
  };

  const [filterDropDownOptions, setFiltDD] = useState({
    props: [
      {
        label: '',
        icon: 'mouseclick'
      }
    ],
    operator: operatorProps
  });

  useEffect(() => {
    if (filter) {
      setValuesByProps(filter.props);
      setNewFilterState(filter);

      if (filter && filter?.extra) {
        let filterData = {};
        if (!isSameKPIGrp) {
          filterData = {
            category: '',
            display_category: '',
            object_type: '',
            property_name: filter?.extra[1],
            entity: '',
            me: '',
            is_property_mapping: true
          };
        } else {
          if (
            selectedMainCategory?.category == 'channels' ||
            selectedMainCategory?.category == 'custom_channels'
          ) {
            filterData = {
              category: selectedMainCategory?.category,
              object_type: filter?.extra[3]
                ? filter?.extra[3]
                : selectedMainCategory?.group,
              property_name: filter?.extra[1],
              display_category: selectedMainCategory?.group,
              entity: 'event'
            };
          } else {
            filterData = {
              category: selectedMainCategory?.category,
              object_type: selectedMainCategory?.group, // depreciated! object_type to display_category key change
              display_category: selectedMainCategory?.group, // object_type to display_category key change
              property_name: filter?.extra[1],
              entity: filter?.extra[3] ? filter?.extra[3] : filter?.extra[2]
            };
          }
        }
        setvalueOptsLoading(true);
        if (propertyValuesMap[filterData?.property_name]) {
          getKPIPropertyValues(activeProject.id, filterData)
            .then((res) => {
              setvalueOptsLoading(false);
            })
            .catch((err) => {
              setvalueOptsLoading(false);
            });
        }
      }
    }
  }, [filter, isSameKPIGrp]);

  useEffect(() => {
    const filterDD = Object.assign({}, filterDropDownOptions);
    const propertyArrays = [];
    Object.keys(filterProps).forEach((propertyType) => {
      const kpiItemsgroupedByCategoryProperty = groupKPIPropertiesOnCategory(
        filterProps?.[propertyType],
        propertyType,
        selectedMainCategory?.group
      );
      propertyArrays.push(...Object.values(kpiItemsgroupedByCategoryProperty));
    });
    filterDD.props = propertyArrays;
    setFiltDD(filterDD);
  }, [filterProps]);

  const parseDateRangeFilter = (fr, to) => {
    return (
      MomentTz(fr).format('MMM DD, YYYY') +
      ' - ' +
      MomentTz(to).format('MMM DD, YYYY')
    );
  };

  const renderFilterContent = () => {
    return (
      <GlobalFilterSelect
        propOpts={filterDropDownOptions.props}
        operatorOpts={filterDropDownOptions.operator}
        valueOpts={propertyValuesMap.data}
        valueOptsLoading={propertyValuesMap.loading}
        applyFilter={applyFilter}
        setValuesByProps={setValuesByProps}
        filter={filter}
        refValue={refValue}
        viewMode={viewMode}
      />
    );
  };

  const onSelectSearch = (userInput) => {
    if (!userInput.currentTarget.value.length) {
      if (userInput.keyCode === 8 || userInput.keyCode === 46) {
        removeFilter();
        return;
      }
    } else if (
      filterTypeState === 'values' &&
      userInput.keyCode === 13 &&
      newFilterState.props[1] === 'numerical'
    ) {
      const newFilter = Object.assign({}, newFilterState);
      newFilter[filterTypeState].push(userInput.currentTarget.value);
      changeFilterTypeState();
      insertFilter(newFilter);
      closeFilter();
    }
    setSearchTerm(userInput.currentTarget.value);

    if (
      (newFilterState.operator === 'contains' ||
        newFilterState.operator === 'does not contain') &&
      filterTypeState === 'values'
    ) {
      const newFilter = Object.assign({}, newFilterState);
      newFilter[filterTypeState][0]
        ? (newFilter[filterTypeState][0] =
            newFilter[filterTypeState][0] + userInput.currentTarget.value)
        : (newFilter[filterTypeState][0] = userInput.currentTarget.value);
      setNewFilterState(newFilter);
      setSearchTerm('');
    }
  };

  const removeFilter = () => {
    const filterState = Object.assign({}, newFilterState);
    if (filterTypeState === 'operator') {
      filterState.props = [];
      changeFilterTypeState(false);
    }
    if (filterTypeState === 'values') {
      filterState.values.length
        ? filterState.values.pop()
        : (() => {
            filterState.operator = '';
            changeFilterTypeState(false);
          })();
    }
    setNewFilterState(filterState);
  };

  const changeFilterTypeState = (next = true) => {
    if (next) {
      filterTypeState === 'props'
        ? setFilterTypeState('operator')
        : filterTypeState === 'operator'
        ? setFilterTypeState('values')
        : (() => {})();
    } else {
      filterTypeState === 'values'
        ? setFilterTypeState('operator')
        : filterTypeState === 'operator'
        ? setFilterTypeState('props')
        : (() => {})();
    }
  };

  const delFilter = () => {
    deleteFilter(index);
  };

  const optionClick = (value) => {
    const newFilter = Object.assign({}, newFilterState);
    if (filterTypeState === 'props') {
      newFilter[filterTypeState] = value;
    } else if (filterTypeState === 'values') {
      newFilter[filterTypeState].push(value);
    } else {
      newFilter[filterTypeState] = value;
    }
    // One more check for props and fetch prop values;

    changeFilterTypeState();
    setNewFilterState(newFilter);
    setSearchTerm('');
  };

  const collapseGroup = (index) => {
    const groupColState = Object.assign({}, groupCollapseState);
    if (groupColState[index]) {
      groupColState[index] = !groupColState[index];
    } else {
      groupColState[index] = true;
    }
    setGroupCollapse(groupColState);
  };

  const addInput = ($event) => {
    const newFilter = Object.assign({}, newFilterState);
    newFilter[filterTypeState] = $event.target.value;
    setNewFilterState(newFilter);
  };

  const onDateSelect = (range) => {
    const newRange = [...selectedRngState];
    const newFilter = Object.assign({}, newFilterState);
    newRange[0] = range.selected;
    const endRange = MomentTz(newRange[0].endDate)
      .endOf('day')
      .toDate()
      .getTime();
    setSelectedRngState(newRange);
    const rangeValue = {
      fr: newRange[0].startDate.getTime(),
      to: endRange,
      ovp: false
    };
    newFilter[filterTypeState] = JSON.stringify(rangeValue);
    setNewFilterState(newFilter);
  };

  const renderOptions = (options) => {
    const renderOptions = [];
    switch (filterTypeState) {
      case 'props':
        let groupOpts;
        if (blockType === 'event') {
          groupOpts = options;
        } else {
          groupOpts = filterType === 'analytics' ? [options[0]] : options;
        }

        groupOpts?.forEach((group, grpIndex) => {
          const collState =
            groupCollapseState[grpIndex] || searchTerm.length > 0;
          renderOptions.push(
            <div className={`fa-select-group-select--content`}>
              {!searchTerm.length && (
                <div
                  className={'fa-select-group-select--option-group'}
                  onClick={() => collapseGroup(grpIndex)}
                >
                  <div>
                    <SVG
                      color={'purple'}
                      name={group.icon}
                      extraClass={'self-center'}
                    ></SVG>
                    <Text
                      level={8}
                      type={'title'}
                      extraClass={'m-0 ml-2 uppercase'}
                      weight={'bold'}
                    >
                      {group.label}
                    </Text>
                  </div>
                  <SVG
                    name={collState ? 'minus' : 'plus'}
                    extraClass={'self-center'}
                  ></SVG>
                </div>
              )}
              <div
                className={
                  styles.filter_block__filter_select__option_group_container_sec
                }
              >
                {collState
                  ? (() => {
                      const valuesOptions = [];
                      filterProps[propsConstants[grpIndex]].forEach((val) => {
                        if (
                          val[0]
                            .toLowerCase()
                            .includes(searchTerm.toLowerCase())
                        ) {
                          valuesOptions.push(
                            <div
                              title={val[0]}
                              className={`fa-select-group-select--options`}
                              onClick={() =>
                                optionClick([...val, propsConstants[grpIndex]])
                              }
                            >
                              {searchTerm.length > 0 && (
                                <div>
                                  <SVG
                                    color={'purple'}
                                    name={group.icon}
                                    extraClass={'self-center'}
                                  ></SVG>
                                </div>
                              )}
                              <Text
                                level={7}
                                type={'title'}
                                extraClass={'m-0'}
                                weight={'thin'}
                              >
                                {val[0]}
                              </Text>
                            </div>
                          );
                        }
                      });
                      return valuesOptions;
                    })()
                  : null}
              </div>
            </div>
          );
        });
        break;
      case 'operator':
        options[newFilterState.props[1]].forEach((opt) => {
          if (opt.toLowerCase().includes(searchTerm.toLowerCase())) {
            renderOptions.push(
              <span
                className={styles.filter_block__filter_select__option}
                onClick={() => optionClick(opt)}
              >
                {opt}
              </span>
            );
          }
        });
        break;
      case 'values':
        if (newFilterState.props[1] === 'categorical') {
          if (!searchTerm.length) {
            renderOptions.push(
              <span
                className={styles.filter_block__filter_select__option}
                onClick={() => optionClick('$none')}
              >
                {'$none'}
              </span>
            );
          } else {
            renderOptions.push(
              <span
                className={styles.filter_block__filter_select__option}
                onClick={() => optionClick(searchTerm)}
              >
                {searchTerm}
              </span>
            );
          }

          if (
            newFilterState.operator !== 'contains' &&
            newFilterState.operator !== 'does not contain' &&
            options[newFilterState.props[0]] &&
            options[newFilterState.props[0]].length
          ) {
            options[newFilterState.props[0]].forEach((opt) => {
              if (opt?.toLowerCase()?.includes(searchTerm.toLowerCase())) {
                renderOptions.push(
                  <span
                    className={styles.filter_block__filter_select__option}
                    onClick={() => optionClick(opt)}
                  >
                    {opt}
                  </span>
                );
              }
            });
          }

          if (!renderOptions.length && searchTerm.length) {
            renderOptions.push(
              <span
                className={styles.filter_block__filter_select__option_nomatch}
              >
                Sorry! No matches
              </span>
            );
          }
        } else if (newFilterState.props[1] === 'numerical') {
          renderOptions.push(
            <span
              className={styles.filter_block__filter_select__option_numerical}
            >
              <Input
                size='large'
                placeholder={'Enter a value'}
                onChange={addInput}
              ></Input>
            </span>
          );
        } else if (newFilterState.props[1] === 'datetime') {
          renderOptions.push(
            <span
              className={`${styles.filter_block__filter_select__date} fa_date_filter`}
            >
              <DateRangePicker
                ranges={selectedRngState}
                onChange={onDateSelect}
                minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
                maxDate={MomentTz(new Date())
                  .subtract(1, 'days')
                  .endOf('day')
                  .toDate()}
              />
            </span>
          );
        }

        break;
    }

    return renderOptions;
  };

  const renderTags = () => {
    const tags = [];
    const tagClass = styles.filter_block__filter_select__tag;
    newFilterState.props?.length
      ? tags.push(<span className={tagClass}>{newFilterState.props[0]}</span>)
      : (() => {})();
    newFilterState.operator?.length
      ? tags.push(<span className={tagClass}>{newFilterState.operator}</span>)
      : (() => {})();

    if (newFilterState.values.length > 0) {
      if (newFilterState.props[1] === 'categorical') {
        newFilterState.values.slice(0, 2).forEach((val, i) => {
          tags.push(
            <span className={tagClass}>{newFilterState.values[i]}</span>
          );
        });
        newFilterState.values.length >= 3
          ? tags.push(<span>...+{newFilterState.values.length - 2}</span>)
          : (() => {})();
      } else if (newFilterState.props[1] === 'datetime') {
        const parsedValues = JSON.parse(newFilterState.values);
        const parsedDatetimeValue = parseDateRangeFilter(
          parsedValues.fr,
          parsedValues.to
        );
        tags.push(<span className={tagClass}>{parsedDatetimeValue}</span>);
      } else if (newFilterState.props[1] === 'numerical') {
        tags.push(<span className={tagClass}>{newFilterState.values}</span>);
      }
    }

    if (tags.length < 1) {
      tags.push(<SVG name={'search'} />);
    }
    return tags;
  };

  const renderApplyFilter = () => {
    if (filterTypeState === 'values') {
      return (
        <span
          className={styles.filter_block__filter_select__apply}
          onClick={() => applyFilter()}
        >
          <Button
            block
            disabled={!newFilterState.values.length}
            className={styles.filter_block__filter_select__apply_btn}
            type='primary'
            onClick={() => applyFilter()}
          >
            Apply Filter
          </Button>
        </span>
      );
    }
  };

  const renderFilterSelect = () => {
    return (
      <div
        className={`absolute ml-4 fa-select fa-filter-select fa-select--group-select top-0 left-0`}
      >
        <Input
          id='fai-filter-input'
          className={styles.filter_block__filter_select__input}
          placeholder={
            newFilterState.values.length >= 2
              ? null
              : placeHolder[filterTypeState]
          }
          prefix={renderTags()}
          onChange={onSelectSearch}
          onKeyDown={onSelectSearch}
          value={searchTerm}
        />
        <div className={'border-top--thin-2 '}>
          <div
            className={`${styles.filter_block__filter_select__options} 
            ${
              filterTypeState === 'values' &&
              styles.filter_block__filter_select__values__options
            }`}
          >
            {filterTypeState !== 'values'
              ? renderOptions(filterDropDownOptions[filterTypeState])
              : renderOptions(dropDownValues)}
          </div>
          {renderApplyFilter()}
        </div>
      </div>
    );
  };

  const applyFilter = (filterState) => {
    if (filterState) {
      insertFilter(filterState, index);
      closeFilter();
    }
  };

  const onClickOutside = () => {
    closeFilter();
  };

  const propOpByPayload = (props, index) => {
    if (props.length === 4) {
      return props[index + 1];
    } else {
      return props[index];
    }
  };

  const setValuesByProps = (props) => {
    if (props && props.length > 3) {
      let filterData = {};

      if (!isSameKPIGrp) {
        filterData = {
          category: '',
          display_category: '',
          object_type: '',
          property_name: props[1],
          entity: '',
          me: '',
          is_property_mapping: true
        };
      } else {
        if (
          selectedMainCategory?.category == 'channels' ||
          selectedMainCategory?.category == 'custom_channels'
        ) {
          filterData = {
            category: selectedMainCategory?.category,
            object_type: props[3] ? props[3] : selectedMainCategory?.group,
            property_name: props[1],
            display_category: selectedMainCategory?.group,
            entity: 'event'
          };
        } else {
          filterData = {
            category: selectedMainCategory?.category,
            object_type: selectedMainCategory?.group, // depreciated! object_type to display_category key change
            display_category: selectedMainCategory?.group, // object_type to display_category key change
            property_name: props[1],
            entity: props[3] ? props[3] : props[2]
          };
        }
      }

      setvalueOptsLoading(true);

      getKPIPropertyValues(activeProject.id, filterData)
        .then((res) => {
          setvalueOptsLoading(false);
        })
        .catch((err) => {
          setvalueOptsLoading(false);
        });
    }
  };

  const filterSelComp = () => {
    return (
      <>
        <GlobalFilterSelect
          propOpts={filterDropDownOptions.props}
          operatorOpts={filterDropDownOptions.operator}
          valueOpts={propertyValuesMap.data}
          valueOptsLoading={propertyValuesMap.loading}
          applyFilter={applyFilter}
          setValuesByProps={setValuesByProps}
          filter={filter}
          refValue={refValue}
          viewMode={viewMode}
        />
      </>
    );
  };

  return (
    <div className={'flex items-center relative'}>
      {showOr && (
        <Text
          level={8}
          type={'title'}
          extraClass={'m-0 mr-2 ml-2'}
          weight={'thin'}
        >
          or
        </Text>
      )}
      <div className={`relative flex`}>
        {filter ? renderFilterContent() : filterSelComp()}
      </div>
      {!viewMode && delFilter && (
        <Button
          type='text'
          onClick={delFilter}
          size={'small'}
          className={`fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button`}
        >
          <SVG name={delIcon} />
        </Button>
      )}
    </div>
  );
}

export default connect(null, { getKPIPropertyValues })(GlobalFilterBlock);
