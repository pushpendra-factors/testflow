import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { Input } from 'antd';
import { SVG } from 'factorsComponents';

function GroupSelect({
  groupedProperties, placeholder,
  optionClick, onClickOutside, extraClass
}) {
  const [groupCollapseState, setGroupCollapseState] = useState({});
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    const groupColState = Object.assign({}, groupCollapseState);
    groupColState[0] = true;
    setGroupCollapseState(groupColState);
  }, [groupedProperties]);

  const collapseGroup = (index) => {
    const groupColState = Object.assign({}, groupCollapseState);
    if (groupColState[index]) {
      groupColState[index] = !groupColState[index];
    } else {
      groupColState[index] = true;
    }
    setGroupCollapseState(groupColState);
  };

  const onInputSearch = (userInput) => {
    setSearchTerm(userInput.currentTarget.value);
  };

  const renderOptions = (options) => {
    const renderGroupedOptions = [];
    options.forEach((group, grpIndex) => {
      const collState = groupCollapseState[grpIndex] || searchTerm.length > 0;
      renderGroupedOptions.push(
            <div key={grpIndex} className={`fa-select-group-select--content`}>
              {!searchTerm.length && <div className={'fa-select-group-select--option-group'}
                onClick={() => collapseGroup(grpIndex)}
              >
                <div>
                    <SVG name={group.icon} size={16} extraClass={'self-center'}></SVG>
                    <span className={'ml-1'}>{group.label}</span>
                </div>
                <SVG color={'grey'} color={'grey'} name={collState ? 'minus' : 'plus'} extraClass={'self-center'}></SVG>
              </div>}
              <div className={styles.dropdown__filter_select__option_group_container_sec}>
                { collState
                  ? (() => {
                    const valuesOptions = [];
                    group.values.forEach((val, i) => {
                      if (val[0].toLowerCase().includes(searchTerm.toLowerCase())) {
                        valuesOptions.push(
                          <div key={i} title={val[0]} className={`fa-select-group-select--options`}
                            onClick={() => optionClick(group.label, val)} >
                              {searchTerm.length > 0 && 
                              <div>
                                <SVG name={group.icon} extraClass={'self-center'}></SVG>
                              </div>
                              }
                              <span className={'ml-1'}>{val[0]}</span>
                          </div>
                        );
                      }
                    });
                    return valuesOptions;
                  })()
                  : null
                }
              </div>
            </div>
      );
    });
    return renderGroupedOptions;
  };

  return (
    <>
        <div className={`${styles.dropdown__filter_select} ml-4 fa-select fa-select--group-select ${extraClass}`}>
          <Input
            className={styles.dropdown__filter_select__input}
            placeholder={placeholder}
            onKeyUp={onInputSearch}
            prefix={(<SVG name="search" size={16} color={'grey'} />)}
          />
          <div className={styles.dropdown__filter_select__content}>
            {renderOptions(groupedProperties)}
          </div>
        </div>
        <div className={styles.dropdown__hd_overlay} onClick={onClickOutside}></div>
    </>
  );
}

export default GroupSelect;
