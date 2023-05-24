import React, { ReactNode } from 'react';
import styles from './index.module.scss';
import { Text } from '../../factorsComponents';
import { OptionType, SingleSelectOptionClickCallbackType } from './types';
import { filterSearchFunction } from './utils';

interface SingleSelectProps {
  options: OptionType[];
  optionClickCallback?: SingleSelectOptionClickCallbackType;
  allowSearch: boolean;
  searchOption: OptionType | null;
  allowSearchTextSelection: boolean;
  searchTerm: string;
}
export default function SingleSelect({
  options,
  optionClickCallback,
  allowSearch,
  searchOption,
  allowSearchTextSelection,
  searchTerm
}: SingleSelectProps) {
  const handleOptionClick = (op: OptionType) => {
    if (optionClickCallback) optionClickCallback(op);
  };
  let rendOpts: ReactNode[] = [];
  if (searchOption && allowSearchTextSelection) {
    // Adding Select Option Based On SearchTerm
    rendOpts.push(
      <div
        key={searchOption.value}
        className={`${
          allowSearch ? 'fa-select-group-select--options' : 'fa-select--options'
        }`}
        onClick={() => handleOptionClick(searchOption)}
      >
        <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
          Select:
        </Text>
        <span className={`ml-1 ${styles.optText}`}>{searchOption.label}</span>
      </div>
    );
  }
  options
    .filter((op) => filterSearchFunction(op, searchTerm))
    .forEach((op, index) => {
      rendOpts.push(
        <div
          key={'op' + index}
          onClick={() => {
            handleOptionClick(op);
          }}
          className={`${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          } ${op.labelNode ? 'w-full' : ''}`}
        >
          {op.labelNode ? op.labelNode : op.label}
        </div>
      );
    });
  return <>{rendOpts}</>;
}
