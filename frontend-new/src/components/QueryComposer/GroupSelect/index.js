import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { Input } from 'antd';
import { SVG, Text } from 'factorsComponents';

function GroupSelect({
  groupedProperties, placeholder,
  optionClick, onClickOutside, extraClass,
  allowEmpty=false
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

  const searchTermExists = (opts) => {
    let termExists = false;

    opts.forEach((grp) => {
      grp.values.forEach((val) => {
        if (val[0].toLowerCase().includes(searchTerm.toLowerCase())) {
          termExists = true;
        }
      })
    })
    return termExists;
  }

  const renderEmptyOpt = () => {
    if(!searchTerm.length) return null;
    return (<div key={0} className={`fa-select-group-select--content`}>
      <div className={styles.dropdown__filter_select__option_group_container_sec}>
        <div className={`fa-select-group-select--options`}
                        onClick={() => optionClick('', [searchTerm])} >
                          <div>
                            <Text level={7} type={'title'} extraClass={'mr-2'}>Select:</Text>
                          </div> 
                          <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>'{searchTerm}'</Text>
                      </div>
      </div>
    </div>)
  }

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
                    <SVG name={group.icon} color={'purple'} extraClass={'self-center'}></SVG>
                    <Text level={8} type={'title'} extraClass={'m-0 ml-2 uppercase'} weight={'bold'}>{group.label}</Text>
                </div>
                <SVG color={'grey'}  name={collState ? 'minus' : 'plus'} extraClass={'self-center'}></SVG>
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
                                <SVG name={group.icon} color={'purple'} extraClass={'self-center'}></SVG>
                              </div>
                              } 
                              <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>{val[0]}</Text>
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
    if(allowEmpty) {
      renderGroupedOptions.push(renderEmptyOpt());
    }
    return renderGroupedOptions;
  };

  return (
    <>
        <div className={`${styles.dropdown__filter_select} fa-select fa-select--group-select ${extraClass}`}>
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
