import React, { useState } from 'react';
import styles from './index.module.scss';

import { Input } from 'antd';
import { SVG } from 'factorsComponents';

export default function FilterBlock({ filter, insertFilter, closeFilter }) {
  const [filterTypeState, setFilterTypeState] = useState('props');
  const [groupCollapseState, setGroupCollapse] = useState({});
  const [searchTerm, setSearchTerm] = useState('');
  const [newFilterState, setNewFilterState] = useState({
    props: '',
    operator: '',
    values: []
  });

  const placeHolder = {
    props: 'Choose a property',
    operator: 'Choose an operator',
    values: 'Choose values'
  };

  const filterDropDownOptions = {
    props: [
      {
        label: 'User Properties',
        icon: 'user',
        values: [
          'Cart Updated',
          'Paid'
        ]
      },
      {
        label: 'Event Properties',
        icon: 'mouseevent',
        values: [
          'City',
          'Country'
        ]
      }
    ],
    operator: [
      'isEqual',
      'lessThan',
      'greaterThan'
    ],
    values: {
      'Cart Updated': ['cart val1', 'cart val2', 'cart val3'],
      Paid: ['paid val1', 'paid val2', 'paid val3'],
      City: ['Bangalore', 'Delhi', 'Mumbai'],
      Country: ['India', 'USA', 'France', 'UK']
    }
  };

  const renderFilterContent = () => {
    return (
      <div className={`${styles.filter_block__filter_content} ml-4`}>
        {filter.props + ' ' + filter.operator + ' ' + filter.values.join(', ')}
      </div>
    );
  };

  const onSelectSearch = (userInput) => {
    if (!userInput.currentTarget.value.length) {
      if (userInput.keyCode === 8 || userInput.keyCode === 46) {
        removeFilter();
      }
    }
    setSearchTerm(userInput.currentTarget.value);
  };

  const removeFilter = () => {
    const filterState = Object.assign({}, newFilterState);
    filterTypeState === 'operator' ? (() => {
      filterState.props = '';
      changeFilterTypeState(false);
    })()
      : null;
    if (filterTypeState === 'values') {
      filterState.values.length ? filterState.values.pop()
        : (() => {
          filterState.operator = '';
          changeFilterTypeState(false);
        })();
    }
    setNewFilterState(filterState);
  };

  const changeFilterTypeState = (next = true) => {
    if (next) {
      filterTypeState === 'props' ? setFilterTypeState('operator')
        : filterTypeState === 'operator' ? setFilterTypeState('values')
          : (() => {})();
    } else {
      filterTypeState === 'values' ? setFilterTypeState('operator')
        : filterTypeState === 'operator' ? setFilterTypeState('props')
          : (() => {})();
    }
  };

  const optionClick = (value) => {
    const newFilter = Object.assign({}, newFilterState);
    if (filterTypeState === 'values') {
      newFilter[filterTypeState].push(value);
    } else {
      newFilter[filterTypeState] = value;
    }
    changeFilterTypeState();
    setNewFilterState(newFilter);
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

  const renderOptions = (options) => {
    const renderOptions = [];
    switch (filterTypeState) {
      case 'props':
        options.forEach((group, grpIndex) => {
          const collState = groupCollapseState[grpIndex];
          renderOptions.push(<>
            <div className={styles.filter_block__filter_select__option_group}
              onClick={() => collapseGroup(grpIndex)}
            >
              <div>
                <SVG name={group.icon} extraClass={'self-center'}></SVG>
                <span className={'ml-1'}>{group.label}</span>
              </div>
              <SVG name={collState ? 'minus' : 'plus'} extraClass={'self-center'}></SVG>
            </div>
            { collState
              ? (() => {
                const valuesOptions = [];
                group.values.forEach((val, index) => {
                  if (val.toLowerCase().includes(searchTerm.toLowerCase())) {
                    valuesOptions.push(
                      <span className={styles.filter_block__filter_select__option}
                        onClick={() => optionClick(val)} >
                        {val}
                      </span>
                    );
                  }
                });
                return valuesOptions;
              })()
              : null
            }
          </>);
        });
        break;
      case 'operator':
        options.forEach(opt => {
          if (opt.toLowerCase().includes(searchTerm.toLowerCase())) {
            renderOptions.push(
              <span className={styles.filter_block__filter_select__option}
                onClick={() => optionClick(opt)} >
                {opt}
              </span>
            );
          }
        });
        break;
      case 'values':
        options[newFilterState.props].forEach(opt => {
          if (opt.toLowerCase().includes(searchTerm.toLowerCase())) {
            renderOptions.push(<span className={styles.filter_block__filter_select__option}
              onClick={() => optionClick(opt)} >
              {opt}
            </span>
            );
          }
        });
        break;
    }

    return renderOptions;
  };

  const renderTags = () => {
    const tags = [];
    const tagClass = styles.filter_block__filter_select__tag;
    newFilterState.props
      ? tags.push(<span className={tagClass}>
        {newFilterState.props}
      </span>) : (() => {})();
    newFilterState.operator
      ? tags.push(<span className={tagClass}>
        {newFilterState.operator}
      </span>) : (() => {})();

    if (newFilterState.values.length > 0) {
      newFilterState.values.slice(0, 2).forEach((val, i) => {
        tags.push(<span className={tagClass}>
          {newFilterState.values[i]}
        </span>);
      });
      newFilterState.values.length >= 3 ? tags.push(
        <span>
                    ...+{newFilterState.values.length - 2}
        </span>
      ) : (() => {})();
    }
    if (tags.length < 1) {
      tags.push(<SVG name="search" />);
    }
    return tags;
  };

  const renderFilterSelect = () => {
    return (
      <div className={`${styles.filter_block__filter_select} ml-4 fa-filter-select`}>
        <Input
          className={styles.filter_block__filter_select__input}
          placeholder={newFilterState.values.length >= 2 ? null
            : placeHolder[filterTypeState]}
          prefix={renderTags()}
          onKeyUp={onSelectSearch}
        />
        <div className={styles.filter_block__filter_select__content}>
          {renderOptions(filterDropDownOptions[filterTypeState])}
        </div>
      </div>
    );
  };

  const onClickOutside = () => {
    if (newFilterState.props.length &&
            newFilterState.operator.length &&
            newFilterState.values.length
    ) {
      insertFilter(newFilterState);
    } else {
      closeFilter();
    }
  };

  return (
    <div className={styles.filter_block}>
      <span className={`${styles.filter_block__prefix} ml-10`}>where</span>
      {filter
        ? renderFilterContent()
        : <>
          {renderFilterSelect()}
          <div className={styles.filter_block__hd_overlay} onClick={onClickOutside}></div>
        </>
      }

    </div>
  );
}
