/* eslint-disable */
import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';

import { Input } from 'antd';
import { SVG } from 'factorsComponents';

import { fetchEventPropertyValues } from '../../../reducers/coreQuery/services';

export default function FilterBlock({ filterProps, activeProject, event, filter, insertFilter, closeFilter }) {
  const [filterTypeState, setFilterTypeState] = useState('props');
  const [groupCollapseState, setGroupCollapse] = useState({});
  const [searchTerm, setSearchTerm] = useState('');
  const [newFilterState, setNewFilterState] = useState({
    props: [],
    operator: '',
    values: []
  });

  const [dropDownValues, setDropDownValues] = useState({});

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
        
      },
      {
        label: 'Event Properties',
        icon: 'mouseevent',
      }
    ],
    operator: {
      "categorical": [
        '=',
        '!=',
        'contains',
        'not contains'
      ],
      "numerical": [
        '=',
        '!=',
        '<',
        '<=',
        '>',
        '>='
      ]
  },
  };

  const renderFilterContent = () => {
    return (
      <div className={`${styles.filter_block__filter_content} ml-4`}>
        {filter.props[0] + ' ' + filter.operator + ' ' + filter.values.join(', ')}
      </div>
    );
  };

  const onSelectSearch = (userInput) => {
    if (!userInput.currentTarget.value.length) {
      if (userInput.keyCode === 8 || userInput.keyCode === 46) {
        removeFilter();
      }
      
    } else if (filterTypeState === 'values' 
      && userInput.keyCode === 13 
      && newFilterState.props[1] === 'numerical') { 
        const newFilter = Object.assign({}, newFilterState);
        newFilter[filterTypeState].push(userInput.currentTarget.value);
        changeFilterTypeState();
        insertFilter(newFilter);
        closeFilter();
    }
    setSearchTerm(userInput.currentTarget.value);
  };

  const removeFilter = () => {
    const filterState = Object.assign({}, newFilterState);
    filterTypeState === 'operator' ? (() => {
      filterState.props = [];
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

  useEffect(() => {
    if(newFilterState.props[1] === 'categorical' && !dropDownValues[newFilterState.props[0]]) {
      fetchEventPropertyValues(activeProject.id, event.label, newFilterState.props[0]).then(res => {
        const ddValues = Object.assign({}, dropDownValues);
        ddValues[newFilterState.props[0]] = res.data;
        setDropDownValues(ddValues);
      })
    }

  }, [newFilterState])

  const optionClick = (value) => {
    const newFilter = Object.assign({}, newFilterState);
    if (filterTypeState === 'props') {
      newFilter[filterTypeState] = value;
    }
    else if (filterTypeState === 'values') {
      newFilter[filterTypeState].push(value);
    } else {
      newFilter[filterTypeState] = value;
    }
    // One more check for props and fetch prop values;
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
          renderOptions.push(
          <div class={styles.filter_block__filter_select__option_group_container}>
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
                filterProps[['user', 'event'][grpIndex]].forEach((val) => {
                  if (val[0].toLowerCase().includes(searchTerm.toLowerCase())) {
                    valuesOptions.push(
                      <span className={styles.filter_block__filter_select__option}
                        onClick={() => optionClick(val)} >
                        {val[0]}
                      </span>
                    );
                  }
                });
                return valuesOptions;
              })()
              : null
            }
          </div>);
        });
        break;
      case 'operator':
        options[newFilterState.props[1]].forEach(opt => {
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
        if(newFilterState.props[1] === 'categorical') {
          options[newFilterState.props[0]] && options[newFilterState.props[0]].forEach(opt => {
            if (opt.toLowerCase().includes(searchTerm.toLowerCase())) {
              renderOptions.push(<span className={styles.filter_block__filter_select__option}
                onClick={() => optionClick(opt)} >
                {opt}
              </span>
              );
            }
          });
        }
        
        break;
    }

    return renderOptions;
  };

  const renderTags = () => {
    const tags = [];
    const tagClass = styles.filter_block__filter_select__tag;
    newFilterState.props
      ? tags.push(<span className={tagClass}>
        {newFilterState.props[0]}
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
          { 
          filterTypeState!== 'values'? 
            renderOptions(filterDropDownOptions[filterTypeState])
           : renderOptions(dropDownValues)
           }
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
      closeFilter();
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
