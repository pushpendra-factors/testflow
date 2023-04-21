import React, { ReactNode, useState } from 'react';
import styles from './index.module.scss';
import { Button } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import { OptionType, handleOptionFunctionType } from './types';
import { selectedOptionsMapper } from './utils';
interface MultiSelectProps {
  options: OptionType[];
  optionClick: handleOptionFunctionType;
  selectedOptions: string[];
  allowSearch: boolean;
  searchOption: OptionType | null;
}

export default function MultiSelect({
  options,
  optionClick,
  selectedOptions,
  allowSearch,
  searchOption
}: MultiSelectProps) {
  const [optionsClicked, setOptionsClicked] = useState(selectedOptions);

  const handleMultipleOptionClick = (op: OptionType) => {
    let newoptionsClicked = [...optionsClicked];
    let index = optionsClicked.indexOf(op.value);
    if (index > -1) {
      //Removing Option From Selected Options
      newoptionsClicked.splice(index, 1);
    } else {
      //Adding Option To Selected Options
      newoptionsClicked.push(op.value);
    }
    setOptionsClicked(newoptionsClicked);
  };

  const checkIsOptionSelected = (value: string) => {
    return optionsClicked.includes(value);
  };

  const applyClick = () => {
    optionClick(selectedOptionsMapper(options, optionsClicked));
  };

  const clearAllClick = () => {
    optionClick([]);
  };
  let rendOpts: ReactNode[] = [];
  if (searchOption) {
    // Adding Select Option Based On SearchTerm
    let isSearchTermSelected = checkIsOptionSelected(searchOption.value);
    rendOpts.push(
      <div
        key={'search' + searchOption.value}
        className={`${
          allowSearch ? 'fa-select-group-select--options' : 'fa-select--options'
        } ${isSearchTermSelected ? styles.fa_selected : ''}`}
        onClick={() => handleMultipleOptionClick(searchOption)}
      >
        <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
          Select:
        </Text>
        <span className={`ml-1 ${styles.optText}`}>{searchOption.label}</span>
        {isSearchTermSelected ? (
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
  const optionsClickedMapped = selectedOptionsMapper(options, optionsClicked);
  //Selected Options That Should be shown on Top.[These Options are present in both optionsClicked and SelectedOptions].
  optionsClickedMapped.forEach((op) => {
    if (selectedOptions.includes(op.value)) {
      rendOpts.push(
        <div
          key={op.value}
          onClick={() => {
            handleMultipleOptionClick(op);
          }}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          } ${styles.fa_selected} `}
        >
          <span className={`ml-1 ${styles.optText}`}>
            {op.labelNode ? op.labelNode : op.label}
          </span>
          <SVG
            name='checkmark'
            extraClass={'self-center'}
            size={17}
            color={'purple'}
          />
        </div>
      );
    }
  });
  options.forEach((op) => {
    const isSelectedInClickedOptions: boolean = checkIsOptionSelected(op.value);
    const isSelectedInSelectedOptions: boolean = selectedOptions.includes(
      op.value
    );
    if (
      (!isSelectedInClickedOptions && !isSelectedInSelectedOptions) ||
      (!isSelectedInClickedOptions && isSelectedInSelectedOptions)
    ) {
      //UnSelected Options.
      rendOpts.push(
        <div
          key={op.value}
          onClick={() => {
            handleMultipleOptionClick(op);
          }}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          }`}
        >
          <span className={`ml-1 ${styles.optText}`}>
            {op.labelNode ? op.labelNode : op.label}
          </span>
        </div>
      );
    } else if (isSelectedInClickedOptions && !isSelectedInSelectedOptions) {
      //Selected Options that should not be moved to Top.
      rendOpts.push(
        <div
          key={op.value}
          onClick={() => {
            handleMultipleOptionClick(op);
          }}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          } ${styles.fa_selected} `}
        >
          <span className={`ml-1 ${styles.optText}`}>
            {op.labelNode ? op.labelNode : op.label}
          </span>
          <SVG
            name='checkmark'
            extraClass={'self-center'}
            size={17}
            color={'purple'}
          />
        </div>
      );
    }
  });
  //Apply and Clear Button.
  rendOpts.push(
    <div className={`${styles.dropdown__apply_opt}`}>
      <div key={'apply_opt'} className={`fa-select--buttons `}>
        <Button
          disabled={optionsClicked.length === 0 && selectedOptions.length === 0}
          type='primary'
          onClick={applyClick}
          className={'w-full'}
        >
          Apply
        </Button>
      </div>
      <div key={'clear_opt'} className={`fa-select--buttons`}>
        <Button
          disabled={optionsClicked.length === 0}
          onClick={clearAllClick}
          className={'w-full'}
        >
          <SVG
            name='times'
            size={17}
            color={
              optionsClicked.length === 0 ? 'rgba(0, 0, 0, 0.251)' : 'grey'
            }
          />
          Clear All
        </Button>
      </div>
    </div>
  );
  return <>{rendOpts}</>;
}
