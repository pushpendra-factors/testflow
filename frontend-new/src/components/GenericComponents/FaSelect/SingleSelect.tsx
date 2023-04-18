import React, { ReactNode } from 'react';
import styles from './index.module.scss';
import { Text } from '../../factorsComponents';
import { OptionType, handleOptionFunctionType } from './types';

interface SingleSelectProps {
  options: OptionType[];
  optionClick: handleOptionFunctionType;
  allowSearch: boolean;
  searchOption: OptionType | null;
}
export default function SingleSelect({
  options,
  optionClick,
  allowSearch,
  searchOption
}: SingleSelectProps) {
  const handleOptionClick = (op: OptionType) => {
    optionClick(op);
  };
  let rendOpts: ReactNode[] = [];
  if (searchOption) {
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
  options.forEach((op, index) => {
    rendOpts.push(
      <div
        key={'op' + index}
        onClick={() => {
          handleOptionClick(op);
        }}
        className={`${
          allowSearch ? 'fa-select-group-select--options' : 'fa-select--options'
        }`}
      >
        {op.labelNode ? op.labelNode : op.label}
      </div>
    );
  });
  return <>{rendOpts}</>;
}
