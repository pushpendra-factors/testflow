import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { SVG, Text } from '../factorsComponents';
import { Input, Button } from 'antd';
import { displayName } from '../FaFilterSelect/utils';

const FaSelect = ({
  options,
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
  disabled=false
}) => {
  const [optClickArr, setOptClickArr] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    if (multiSelect && selectedOpts && selectedOpts.length) {
      const arr = selectedOpts.map((op) => {
        return JSON.stringify([op]);
      });
      setOptClickArr(arr);
    }
  }, [selectedOpts]);

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

    if (searchTerm?.length) {
      isSelected = isSelectedCheck([searchTerm]);
      rendOpts.push(
        <div
          key={'custom_value'}
          className={`fa-select-group-select--options ${
            isSelectedCheck([searchTerm]) ? styles.fa_selected : null
          }`}
          onClick={() =>
            optClick(() => optionClick([searchTerm]), [searchTerm])
          }
        >
          <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
            {' '}
            Select:{' '}
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
        if (op[0].toLowerCase().includes(searchTerm.toLowerCase())) {
          rendOpts.push(
            <div
              key={index}
              title={displayName[op[0]] ? displayName[op[0]] : op[0]}
              className={`fa-select-group-select--options ${
                isSelected ? styles.fa_selected : null
              }`}
              onClick={() => optClick(() => optionClick(op), op)}
            >
              {op[1] && !multiSelect ? (
                <SVG name={op[1]} extraClass={'self-center'}></SVG>
              ) : null}
              <span className={`ml-1 ${styles.optText}`}>
                {displayName[op[0]] ? displayName[op[0]] : op[0]}
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
      options.forEach((op, index) => {
        isSelected = isSelectedCheck(op);
        rendOpts.push(
          <div
            key={index}
            title={displayName[op[0]] ? displayName[op[0]] : op[0]}
            className={`fa-select-group-select--options ${
              isSelected ? styles.fa_selected : null
            }`}
            onClick={() => optClick(() => optionClick(op), op)}
          >
            {op[1] && !multiSelect ? (
              <SVG name={op[1]} extraClass={'self-center'}></SVG>
            ) : null}
            <span className={`ml-1 ${styles.optText}`}>
              {displayName[op[0]] ? displayName[op[0]] : op[0]}
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
          className={`fa-select-group-select--options ${styles.dropdown__del_opt}`}
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
            className={`fa-select-group-select--options ${styles.dropdown__empty_opt}`}
          ></div>
        );
      }
      rendOpts.push(
        <div
          key={'apply_opt'}
          className={`fa-select-group-select--options ${styles.dropdown__apply_opt}`}
          onClick={applyClick}
        >
          <Button disabled={!optClickArr.length} type='primary'>
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
      <div className={`${styles.selectInput} fa-filter-select`}>
        <Input
          prefix={<SVG name={'search'} />}
          size='large'
          placeholder={'Search'}
          onChange={search}
        ></Input>
      </div>
    );
  };

  return (
    <>
      <div
        className={`${extraClass} ${
          posRight ? styles.dropdown__select_rt : styles.dropdown__select_lt
        } fa-select ${
          posRight ? `fa-select--group-select-sm` : `fa-select--group-select`
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
};

export default FaSelect;
