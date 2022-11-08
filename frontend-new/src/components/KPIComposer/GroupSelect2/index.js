import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { Input, Button } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { CaretDownOutlined, CaretUpOutlined } from '@ant-design/icons';
import _ from 'lodash';
import useAutoFocus from '../../../hooks/useAutoFocus';

function GroupSelect2({
  groupedProperties,
  placeholder,
  optionClick,
  onClickOutside,
  extraClass,
  allowEmpty = false,
  hideTitle = false,
  textStartCase = false
}) {
  const [groupCollapseState, setGroupCollapseState] = useState({});
  const [searchTerm, setSearchTerm] = useState('');
  const [showFull, setShowFull] = useState([]);
  const inputComponentRef = useAutoFocus();

  useEffect(() => {
    const groupColState = Object.assign({}, groupCollapseState);
    Object.keys(groupedProperties)?.forEach((index) => {
      groupColState[index] = true;
    });
    setGroupCollapseState(groupColState);
  }, [groupedProperties]);

  const onInputSearch = (userInput) => {
    setSearchTerm(userInput.currentTarget.value);
  };

  const renderEmptyOpt = () => {
    if (!searchTerm.length) return null;
    return (
      <div key={0} className={`fa-select-group-select--content`}>
        <div
          className={styles.dropdown__filter_select__option_group_container_sec}
        >
          <div
            className={`fa-select-group-select--options`}
            onClick={() => optionClick('', '', [searchTerm])}
          >
            <div>
              <Text level={7} type={'title'} extraClass={'mr-2'}>
                Select:
              </Text>
            </div>
            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
              '{searchTerm}'
            </Text>
          </div>
        </div>
      </div>
    );
  };

  const getGroupLabel = (grp) => {
    if (grp === 'event') return 'Event Properties';
    if (grp === 'user') return 'User Properties';
    return grp;
  };

  const setShowMoreIndex = (ind, flag) => {
    const showMoreState = [...showFull];
    showMoreState[ind] = flag;
    setShowFull(showMoreState);
  };

  const renderOptions = (options) => {
    const renderGroupedOptions = [];
    options?.forEach((group, grpIndex) => {
      const collState = groupCollapseState[grpIndex] || searchTerm.length > 0;

      let hasSearchTerm = false;
      const valuesOptions = [];

      const icon = group?.icon
        ? group.icon.toLowerCase().split(' ').join('_')
        : group.icon;
      const groupItem = (
        <div key={group.label} className={`fa-select-group-select--content`}>
          {!hideTitle && (
            <div className={'fa-select-group-select--option-group'}>
              <div>
                {/* <SVG name={icon} color={'purple'} extraClass={'self-center'}></SVG> */}
                <Text
                  level={8}
                  type={'title'}
                  extraClass={'m-0 ml-2'}
                  weight={'bold'}
                >
                  {_.startCase(group.label)}
                </Text>
              </div>
            </div>
          )}

          <div
            className={
              styles.dropdown__filter_select__option_group_container_sec
            }
          >
            {collState
              ? (() => {
                  group?.values?.forEach((val, i) => {
                    if (
                      val[0].toLowerCase().includes(searchTerm.toLowerCase())
                    ) {
                      hasSearchTerm = true;
                      valuesOptions.push(
                        <div
                          key={i}
                          title={val[0]}
                          className={`fa-select-group-select--options`}
                          onClick={() =>
                            optionClick(group.label, val, group?.category)
                          }
                        >
                          {searchTerm.length > 0}
                          <Text
                            level={7}
                            type={'title'}
                            extraClass={'m-0'}
                            weight={'thin'}
                          >
                            {textStartCase ? _.startCase(val[0]) : val[0]}
                          </Text>
                        </div>
                      );
                    }
                  });
                  return showFull[grpIndex]
                    ? valuesOptions
                    : valuesOptions.slice(0, 5);
                })()
              : null}
          </div>

          {valuesOptions.length > 5 && collState ? (
            !showFull[grpIndex] ? (
              <Button
                className={styles.dropdown__filter_select__showhide}
                type='text'
                onClick={() => {
                  setShowMoreIndex(grpIndex, true);
                }}
                icon={<CaretDownOutlined />}
              >
                Show More ({valuesOptions.length - 5})
              </Button>
            ) : (
              <Button
                className={styles.dropdown__filter_select__showhide}
                type='text'
                onClick={() => {
                  setShowMoreIndex(grpIndex, false);
                }}
                icon={<CaretUpOutlined />}
              >
                Show Less
              </Button>
            )
          ) : null}
        </div>
      );
      hasSearchTerm && renderGroupedOptions.push(groupItem);
    });
    if (allowEmpty) {
      renderGroupedOptions.push(renderEmptyOpt());
    }
    return renderGroupedOptions;
  };

  return (
    <>
      <div
        className={`${styles.dropdown__filter_select} fa-select fa-select--group-select ${extraClass}`}
      >
        <div className={styles.dropdown__filter_select__input}>
          <Input
            placeholder={placeholder}
            onKeyUp={onInputSearch}
            prefix={<SVG name='search' size={16} color={'grey'} />}
            ref={inputComponentRef}
          />
        </div>
        <div className={styles.dropdown__filter_select__content}>
          {renderOptions(groupedProperties)}
        </div>
      </div>
      <div
        className={styles.dropdown__hd_overlay}
        onClick={onClickOutside}
      ></div>
    </>
  );
}

export default GroupSelect2;
