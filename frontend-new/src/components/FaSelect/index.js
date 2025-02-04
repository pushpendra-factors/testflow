import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { SVG, Text } from '../factorsComponents';
import { Input, Button, Spin } from 'antd';
import { DISPLAY_PROP } from '../../utils/constants';
import useAutoFocus from '../../hooks/useAutoFocus';
import { generateRandomKey } from 'Utils/global';
import isEqual from 'lodash/isEqual';
import sortBy from 'lodash/sortBy';
import { filterURLValue } from 'Utils/filterURLValue';

function FaSelect({
  options,
  displayNames = DISPLAY_PROP,
  delOption,
  optionClick,
  delOptionClick,
  applClick,
  multiSelect = false,
  onClickOutside,
  selectedOpts = [],
  allowSearch = false,
  posRight = false,
  children,
  extraClass = '',
  disabled = false,
  showIcon = false,
  placement = 'bottom',
  style,
  placeholder = 'Search'
}) {
  const [optClickArr, setOptClickArr] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const inputComponentRef = useAutoFocus(allowSearch);

  useEffect(() => {
    if (multiSelect && selectedOpts && selectedOpts.length) {
      const arr = selectedOpts.map((op) => {
        return JSON.stringify([op]);
      });
      setOptClickArr(arr);
    }
  }, [selectedOpts]);

  const checkSelectedAndOptClickEquality = () => {
    const parsedArray = optClickArr
      .map((arr) => JSON.parse(arr)?.[0])
      .filter((v) => !!v);
    return isEqual(sortBy(selectedOpts), sortBy(parsedArray));
  };

  const optClick = (clickFunc, option) => {
    if (!multiSelect) {
      clickFunc();
    } else {
      const stringedOpt = JSON.stringify(option);
      const clckInd = optClickArr.findIndex((opt) => opt === stringedOpt);
      let opts;
      if (clckInd < 0) {
        opts = [...optClickArr];
        opts.push(stringedOpt);
      } else {
        opts = [...optClickArr.filter((op, i) => i !== clckInd)];
      }
      setOptClickArr(opts);
    }
  };

  const applyClick = () => {
    applClick([...optClickArr]);
    onClickOutside();
  };

  const isSelectedCheck = (op) => {
    return optClickArr.includes(JSON.stringify(op)) ? true : false;
  };

  const renderOptions = () => {
    let rendOpts = [];
    let isSelected = false;
    if (!options?.length) {
      rendOpts.push(
        <div className='flex justify-center items-center my-2'>
          <Spin size='small' />
          <Text
            level={7}
            type={'title'}
            extraClass={'ml-2'}
            weight={'thin'}
            color={'grey'}
          >
            Loading data...
          </Text>
        </div>
      );
    } else if (searchTerm?.length) {
      isSelected = isSelectedCheck([searchTerm]);
      rendOpts.push(
        <div
          key={'custom_value'}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          } ${isSelectedCheck([searchTerm]) ? styles.fa_selected : null}`}
          onClick={() =>
            optClick(() => optionClick([searchTerm]), [searchTerm])
          }
        >
          <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
            Select:
          </Text>
          <span className={`ml-1 ${styles.optText}`}>{searchTerm}</span>
          {isSelected ? (
            <SVG
              name='checkmark'
              extraClass={'self-center'}
              size={17}
              color={'purple'}
            />
          ) : null}
        </div>
      );

      options.forEach((op, index) => {
        isSelected = isSelectedCheck(op);
        let st = filterURLValue(searchTerm);
        if (
          op[0].toLowerCase().includes(st) ||
          (op[0] === '$none' && displayNames[op[0]].toLowerCase().includes(st))
        ) {
          rendOpts.push(
            <div
              key={index}
              title={displayNames[op[0]] ? displayNames[op[0]] : op[0]}
              className={`${
                allowSearch
                  ? 'fa-select-group-select--options'
                  : 'fa-select--options'
              } ${isSelected ? styles.fa_selected : null}`}
              onClick={() => optClick(() => optionClick(op), op)}
            >
              {op[1] && showIcon && !multiSelect ? (
                <SVG name={op[1]} extraClass={'self-center'}></SVG>
              ) : null}
              <span className={`ml-1 ${styles.optText}`}>
                {displayNames[op[0]] ? displayNames[op[0]] : op[0]}
              </span>
              {isSelected ? (
                <SVG
                  name='checkmark'
                  extraClass={'self-center'}
                  size={17}
                  color={'purple'}
                />
              ) : null}
            </div>
          );
        }
      });
    } else {
      multiSelect &&
        selectedOpts.forEach((op, index) => {
          let is = isSelectedCheck([op]);
          if (!is) {
            return;
          }
          rendOpts.push(
            <div
              key={op ? op : generateRandomKey()}
              title={displayNames[op] ? displayNames[op] : op}
              className={`${
                allowSearch
                  ? 'fa-select-group-select--options'
                  : 'fa-select--options'
              } ${styles.fa_selected}`}
              onClick={() => optClick(() => optionClick([op]), [op])}
            >
              <span className={`ml-1 ${styles.optText}`}>
                {displayNames[op] ? displayNames[op] : op}
              </span>

              <SVG
                name='checkmark'
                extraClass={'self-center'}
                size={17}
                color={'purple'}
              />
            </div>
          );
        });
      options.forEach((op, index) => {
        isSelected = isSelectedCheck(op);

        let is = selectedOpts.includes(op[0]) && isSelected;
        if (is) return;
        rendOpts.push(
          <div
            key={op ? op[0] : generateRandomKey()}
            title={displayNames[op[0]] ? displayNames[op[0]] : op[0]}
            style={{
              color: op[2] === 'disabled' ? '#B7BEC8' : '#0E2647',
              cursor: op[2] === 'disabled' ? 'not-allowed' : 'pointer'
            }}
            className={`${
              allowSearch
                ? 'fa-select-group-select--options'
                : 'fa-select--options'
            } ${isSelected ? styles.fa_selected : null}`}
            onClick={() => optClick(() => optionClick(op), op)}
          >
            {op[1] && showIcon && !multiSelect ? (
              <SVG
                name={op[1]}
                color={op[2] === 'disabled' ? '#B7BEC8' : '#0E2647'}
                style={{
                  cursor: op[2] === 'disabled' ? 'not-allowed' : 'pointer'
                }}
                extraClass={'self-center'}
              ></SVG>
            ) : null}
            <span className={`ml-1 ${styles.optText}`}>
              {displayNames[op[0]] ? displayNames[op[0]] : op[0]}
            </span>
            {isSelected ? (
              <SVG
                name='checkmark'
                extraClass={'self-center'}
                size={17}
                color={'purple'}
              />
            ) : null}
          </div>
        );
      });
    }

    if (delOption) {
      rendOpts.push(
        <div
          key={'del_opt'}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          } ${styles.dropdown__del_opt}`}
          onClick={delOptionClick}
        >
          <SVG name={'remove'} extraClass={'self-center'}></SVG>
          <span className={'ml-1'}>{delOption}</span>
        </div>
      );
    }

    if (multiSelect) {
      if (rendOpts.length >= 4) {
        rendOpts.push(
          <div
            key={'empty_opt'}
            className={`${
              allowSearch
                ? 'fa-select-group-select--options'
                : 'fa-select--options'
            } ${styles.dropdown__empty_opt}`}
          ></div>
        );
      }
      rendOpts.push(
        <div
          key={'apply_opt'}
          className={`fa-select--buttons ${styles.dropdown__apply_opt} p `}
        >
          <Button
            disabled={checkSelectedAndOptClickEquality()}
            type='primary'
            onClick={applyClick}
            className={'w-full'}
          >
            Apply
          </Button>
        </div>
      );
    }

    return rendOpts;
  };

  const search = (val) => {
    setSearchTerm(val.currentTarget.value);
  };

  const renderSearchInput = () => {
    return (
      <div
        className={`${styles.selectInput} fa-filter-select fa-search-select`}
      >
        <Input
          style={{ overflow: 'hidden' }}
          prefix={<SVG name={'search'} />}
          size='large'
          placeholder={placeholder}
          onChange={search}
          ref={inputComponentRef}
          autoFocus={true}
        ></Input>
      </div>
    );
  };

  return (
    <>
      <div
        style={style}
        className={`${extraClass}  ${styles.dropdown__select}
          ${
            posRight
              ? styles.dropdown__select_right_0
              : styles.dropdown__select_left_0
          } fa-select ${
          posRight ? `fa-select--group-select-sm` : `fa-select--group-select`
        } ${
          allowSearch
            ? `fa-select--group-select-sm`
            : `fa-select--group-select-mini`
        } ${
          placement === 'top'
            ? styles.dropdown__select_placement_top
            : styles.dropdown__select_placement_bottom
        }`}
      >
        {allowSearch && renderSearchInput()}
        <div
          className={`fa-select-dropdown ${styles.dropdown__select__content}`}
        >
          {children || renderOptions()}
        </div>
      </div>
      <div
        className={styles.dropdown__hd_overlay}
        onClick={onClickOutside}
      ></div>
    </>
  );
}

export default FaSelect;
