import React, { useState } from 'react';
import styles from './index.module.scss';
import { Input } from 'antd';
import { SVG } from 'factorsComponents';

function GroupSelect({
  groupedProperties, placeholder,
  optionClick, onClickOutside
}) {
  const [groupCollapseState, setGroupCollapseState] = useState({});
  const [searchTerm, setSearchTerm] = useState('');

  const collapseGroup = (index) => {
    const groupColState = Object.assign({}, groupCollapseState);
    if (groupColState[index]) {
      groupColState[index] = !groupColState[index];
    } else {
      groupColState[index] = true;
    }
    setGroupCollapseState(groupColState);
  };

  const renderOptions = (options) => {
    const renderGroupedOptions = [];
    options.forEach((group, grpIndex) => {
      const collState = groupCollapseState[grpIndex];
      renderGroupedOptions.push(
            <div className={styles.dropdown__filter_select__option_group_container}>
              <div className={styles.dropdown__filter_select__option_group}
                onClick={() => collapseGroup(grpIndex)}
              >
                <div>
                    <SVG name={group.icon} extraClass={'self-center'}></SVG>
                    <span className={'ml-1'}>{group.label}</span>
                </div>
                <SVG name={collState ? 'minus' : 'plus'} extraClass={'self-center'}></SVG>
              </div>
              <div className={styles.dropdown__filter_select__option_group_container_sec}>
                { collState
                  ? (() => {
                    const valuesOptions = [];
                    group.values.forEach((val) => {
                      if (val[0].toLowerCase().includes(searchTerm.toLowerCase())) {
                        valuesOptions.push(
                            <span className={styles.dropdown__filter_select__option}
                            onClick={() => optionClick(group.label, val)} >
                            {val[0]}
                            </span>
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
        <div className={`${styles.dropdown__filter_select} ml-4 fa-filter-select`}>
          <Input
            className={styles.dropdown__filter_select__input}
            placeholder={placeholder}
            onKeyUp={(userInput) => setSearchTerm(userInput.currentTarget.value)}
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
