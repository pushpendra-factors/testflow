import React, { ReactNode, useMemo, useState } from 'react';
import styles from './index.module.scss';
import { Button } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import { ApplyClickCallbackType, OptionType } from './types';
import { filterSearchFunction, moveSelectedOptionsToTop } from './utils';
import { HighlightSearchText } from 'Utils/dataFormatter';
interface MultiSelectProps {
  options: OptionType[];
  applyClickCallback?: ApplyClickCallbackType;
  allowSearch: boolean;
  searchOption: OptionType | null;
  maxAllowedSelection?: number;
  allowSearchTextSelection: boolean;
  searchTerm: string;
}

export default function MultiSelect({
  options,
  applyClickCallback,
  allowSearch,
  searchOption,
  maxAllowedSelection = 0,
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
        else count -= 1;
        updatedItem.isSelected = !updatedItem?.isSelected;

        updatedLocalOptions = [
          ...updatedLocalOptions.slice(0, index),
          updatedItem,
          ...updatedLocalOptions.slice(index + 1)
        ];
        updateFlag = true;
      }
    } else if (allowSearchTextSelection) {
      //For Custom Value.
      if (
        maxAllowedSelection &&
        localSelectedOptionCount >= maxAllowedSelection
      ) {
        updateFlag = false;
      } else {
        updatedLocalOptions = [
          { ...op, isSelected: true },
          ...updatedLocalOptions
        ];
        updateFlag = true;
      }
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
    } else if (maxAllowedSelection && updateFlag) {
      //enabling all the disabled options
      updatedLocalOptions = updatedLocalOptions.map((option) => {
        return { ...option, isDisabled: false };
      });
    }
    if (updateFlag) setLocalOptions(updatedLocalOptions);
  };

  const localSelectedOptionCount = useMemo(() => {
    return localOptions.filter((option) => option?.isSelected).length;
  }, [localOptions]);

  const propsOptionsSelectedCount = useMemo(() => {
    return options.filter((option) => option?.isSelected).length;
  }, [options]);

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
    let updatedLocalOptions = [...localOptions];

    updatedLocalOptions = updatedLocalOptions.map((option) => {
      return { ...option, isSelected: false, isDisabled: false };
    });
    setLocalOptions(updatedLocalOptions);
  };

  let rendOpts: ReactNode[] = [];

  //Clear Button.
  rendOpts.push(
    <div key={'clear_opt'} className={`${styles.dropdown__clear_opt}`}>
      <Button
        disabled={localSelectedOptionCount === 0}
        onClick={clearAllClick}
        type='link'
      >
        Clear all
      </Button>
      {maxAllowedSelection !== 0 && (
        <Text
          level={7}
          type={'title'}
          extraClass={'ml-4 mb-0'}
          weight={'thin'}
          color={'grey'}
        >
          {localSelectedOptionCount}/{maxAllowedSelection}
        </Text>
      )}
    </div>
  );

  if (searchOption && allowSearchTextSelection) {
    // Adding Select Option Based On SearchTerm
    rendOpts.push(
      <div
        tabIndex={0}
        key={'search' + searchOption.value}
        className={`${
          allowSearch ? 'fa-select-group-select--options' : 'fa-select--options'
        } ${
          searchOption?.isSelected
            ? `${styles.fa_selected}`
            : maxAllowedSelection !== 0 &&
                localSelectedOptionCount >= maxAllowedSelection
              ? `${styles.dropdown__disabled_opt}`
              : ''
        } `}
        onClick={() => handleMultipleOptionClick(searchOption)}
        onKeyDown={(e) => {
          e.key === 'Enter' && handleMultipleOptionClick(searchOption);
        }}
      >
        <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
          Select:
        </Text>
        <span className={`ml-1 ${styles.optText}`}>{searchOption.label}</span>
        {searchOption?.isSelected ? (
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
          onKeyDown={(e) => {
            e.key === 'Enter' && handleMultipleOptionClick(option);
          }}
          tabIndex={0}
          key={option.value}
          title={option.label}
          onClick={() => {
            handleMultipleOptionClick(option);
          }}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          } ${
            option?.isSelected
              ? `${styles.fa_selected}`
              : option?.isDisabled
                ? `${styles.dropdown__disabled_opt}`
                : ''
          } `}
        >
          <div className='w-11/12'>
            {option.labelNode ? (
              option.labelNode
            ) : searchTerm.length > 0 ? (
              <HighlightSearchText text={option.label} highlight={searchTerm} />
            ) : (
              <Text
                level={7}
                type={'title'}
                weight={'regular'}
                extraClass={'self-center m-0'}
              >
                {option.label}
              </Text>
            )}
          </div>
          {option?.isSelected && (
            <SVG
              name='checkmark'
              extraClass={'self-center w-1/12'}
              size={17}
              color={'purple'}
            />
          )}
        </div>
      );
    });

  //Apply Button.
  rendOpts.push(
    <div className={`${styles.dropdown__apply_opt}`} key={'actions'}>
      <div key={'apply_opt'} className={`fa-select--buttons `}>
        <Button
          tabIndex={0}
          disabled={
            localSelectedOptionCount === 0 && propsOptionsSelectedCount === 0
          }
          type='primary'
          onClick={applyClick}
          className={'w-full'}
        >
          Apply
        </Button>
      </div>
    </div>
  );
  return <>{rendOpts}</>;
}
