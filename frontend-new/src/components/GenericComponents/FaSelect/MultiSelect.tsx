import React, { ReactNode, useEffect, useMemo, useState } from 'react';
import styles from './index.module.scss';
import { Button } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import { ApplyClickCallbackType, OptionType } from './types';
import { filterSearchFunction, moveSelectedOptionsToTop } from './utils';
interface MultiSelectProps {
  options: OptionType[];
  applyClickCallback?: ApplyClickCallbackType;
  allowSearch: boolean;
  searchOption: OptionType | null;
  maxAllowedSelection?: number;
  setShowMaxLimitWarning: (state: boolean) => void;
  allowSearchTextSelection: boolean;
  searchTerm: string;
}

export default function MultiSelect({
  options,
  applyClickCallback,
  allowSearch,
  searchOption,
  maxAllowedSelection = 0,
  setShowMaxLimitWarning,
  allowSearchTextSelection,
  searchTerm
}: MultiSelectProps) {
  const [localOptions, setLocalOptions] = useState<OptionType[]>(
    moveSelectedOptionsToTop(options)
  );

  const handleMultipleOptionClick = (op: OptionType) => {
    let updatedLocalOptions = [...localOptions];
    let count = localSelectedOptionCount;
    let updateFlag = false;
    const index = localOptions.findIndex(
      (option: OptionType) => option.value === op.value
    );
    if (index > -1) {
      let updatedItem = updatedLocalOptions[index];
      if (
        maxAllowedSelection &&
        !updatedItem?.isSelected &&
        localSelectedOptionCount >= maxAllowedSelection
      ) {
        updateFlag = false;
      } else {
        if (!updatedItem.isSelected) count += 1;
        updatedItem.isSelected = !updatedItem?.isSelected;

        updatedLocalOptions = [
          ...updatedLocalOptions.slice(0, index),
          updatedItem,
          ...updatedLocalOptions.slice(index + 1)
        ];
        updateFlag = true;
      }
    } else {
      updatedLocalOptions = [
        { ...op, isSelected: true },
        ...updatedLocalOptions
      ];
      updateFlag = true;
    }
    //finding count for selected options
    if (maxAllowedSelection && count >= maxAllowedSelection) {
      //diabling all not selected options
      updatedLocalOptions = updatedLocalOptions.map((option) => {
        if (!option?.isSelected) {
          return { ...option, isDisabled: true };
        }
        return option;
      });
      updateFlag = true;
    }
    if (updateFlag) setLocalOptions(updatedLocalOptions);
  };

  const localSelectedOptionCount = useMemo(() => {
    return localOptions.filter((option) => option?.isSelected).length;
  }, [localOptions]);

  const applyClick = () => {
    if (applyClickCallback)
      applyClickCallback(
        localOptions,
        localOptions
          .filter((option) => option?.isSelected)
          .map((op) => op.value)
      );
  };

  const clearAllClick = () => {
    if (applyClickCallback) applyClickCallback(options, []);
  };

  // For showing or hiding max limit warning
  useEffect(() => {
    if (
      maxAllowedSelection &&
      localSelectedOptionCount >= maxAllowedSelection
    ) {
      setShowMaxLimitWarning(true);
    } else {
      setShowMaxLimitWarning(false);
    }
  }, [
    maxAllowedSelection,
    localOptions,
    setShowMaxLimitWarning,
    localSelectedOptionCount
  ]);

  let rendOpts: ReactNode[] = [];
  if (searchOption && allowSearchTextSelection) {
    // Adding Select Option Based On SearchTerm
    let isSearchTermSelected = searchOption?.isSelected;
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
  localOptions
    .filter((op) => filterSearchFunction(op, searchTerm))
    .forEach((option) => {
      rendOpts.push(
        <div
          key={option.value}
          onClick={() => {
            handleMultipleOptionClick(option);
          }}
          style={{ cursor: option?.isDisabled ? 'not-allowed' : 'pointer' }}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          } ${option?.isSelected ? `${styles.fa_selected}` : ''} `}
        >
          <span className={`ml-1 ${styles.optText}`}>
            {option.labelNode ? option.labelNode : option.label}
          </span>
          {option?.isSelected && (
            <SVG
              name='checkmark'
              extraClass={'self-center'}
              size={17}
              color={'purple'}
            />
          )}
        </div>
      );
    });

  //Apply and Clear Button.
  rendOpts.push(
    <div className={`${styles.dropdown__apply_opt}`} key={'actions'}>
      <div key={'apply_opt'} className={`fa-select--buttons `}>
        <Button
          disabled={
            localSelectedOptionCount === 0 &&
            options.filter((op) => op?.isSelected).length === 0
          }
          type='primary'
          onClick={applyClick}
          className={'w-full'}
        >
          Apply
        </Button>
      </div>
      <div key={'clear_opt'} className={`fa-select--buttons`}>
        <Button
          disabled={localSelectedOptionCount === 0}
          onClick={clearAllClick}
          className={'w-full'}
        >
          <SVG
            name='times'
            size={17}
            color={
              localSelectedOptionCount === 0 ? 'rgba(0, 0, 0, 0.251)' : 'grey'
            }
          />
          Clear All
        </Button>
      </div>
    </div>
  );
  return <>{rendOpts}</>;
}
