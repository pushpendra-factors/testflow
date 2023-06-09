import React, { ReactNode, useEffect, useRef, useState } from 'react';
import styles from './index.module.scss';
import { Text } from '../../factorsComponents';
import { OptionType, SingleSelectOptionClickCallbackType } from './types';
import { filterSearchFunction } from './utils';
import useKey from 'hooks/useKey';

interface SingleSelectProps {
  options: OptionType[];
  optionClickCallback?: SingleSelectOptionClickCallbackType;
  allowSearch: boolean;
  searchOption: OptionType | null;
  allowSearchTextSelection: boolean;
  searchTerm: string;
  extraClass?: string;
}
export default function SingleSelect({
  options,
  optionClickCallback,
  allowSearch,
  searchOption,
  allowSearchTextSelection,
  searchTerm,
  extraClass = ''
}: SingleSelectProps) {
  const handleOptionClick = (op: OptionType) => {
    if (optionClickCallback) optionClickCallback(op);
  };
  const dropdownRef = useRef(null);
  const filteredOptions = options.filter((op) =>
    filterSearchFunction(op, searchTerm)
  );

  const optionsLength =
    searchOption && allowSearchTextSelection
      ? filteredOptions.length + 1
      : filteredOptions.length;

  const [hoveredOptionIndex, setHoveredOptionIndex] = useState(0);

  const handleKeyArrowDown = () => {
    setHoveredOptionIndex((prevIndex) =>
      prevIndex < optionsLength - 1 ? prevIndex + 1 : 0
    );
  };
  const handleKeyArrowUp = () => {
    setHoveredOptionIndex((prevIndex) =>
      prevIndex > 0 ? prevIndex - 1 : optionsLength - 1
    );
  };
  const handleKeyEnter = () => {
    if (searchOption && allowSearchTextSelection) {
      if (hoveredOptionIndex > 0)
        handleOptionClick(filteredOptions[hoveredOptionIndex - 1]);
      else {
        handleOptionClick(searchOption);
      }
    } else {
      handleOptionClick(filteredOptions[hoveredOptionIndex]);
    }
  };
  useKey('ArrowDown', handleKeyArrowDown);
  useKey('ArrowUp', handleKeyArrowUp);
  useKey('Enter', handleKeyEnter);

  const scrollToSelectedOption = () => {
    if (dropdownRef.current) {
      const selectedOptionElement =
        dropdownRef.current.children[hoveredOptionIndex];
      if (selectedOptionElement) {
        const { offsetTop } = selectedOptionElement;
        dropdownRef.current.scrollTop = offsetTop;
      }
    }
  };
  useEffect(() => {
    scrollToSelectedOption();
  }, [hoveredOptionIndex]);

  let rendOpts: ReactNode[] = [];
  if (searchOption && allowSearchTextSelection) {
    // Adding Select Option Based On SearchTerm
    rendOpts.push(
      <div
        key={searchOption.value}
        className={`${extraClass} ${
          allowSearch ? 'fa-select-group-select--options' : 'fa-select--options'
        } ${hoveredOptionIndex === 0 ? styles.hoveredOption : ''}`}
        onClick={() => handleOptionClick(searchOption)}
      >
        <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
          Select:
        </Text>
        <span className={`ml-1 ${styles.optText}`}>{searchOption.label}</span>
      </div>
    );
  }

  filteredOptions.forEach((op, index) => {
    rendOpts.push(
      <div
        key={'op' + index}
        onClick={() => {
          handleOptionClick(op);
        }}
        className={`${extraClass} ${
          allowSearch ? 'fa-select-group-select--options' : 'fa-select--options'
        } ${
          hoveredOptionIndex ===
          (searchOption && allowSearchTextSelection ? index + 1 : index)
            ? styles.hoveredOption
            : ''
        }`}
      >
        {op.labelNode ? op.labelNode : op.label}
      </div>
    );
  });
  return (
    <div
      className='flex flex-col'
      ref={dropdownRef}
      style={{ overflowY: 'auto' }}
    >
      {rendOpts}
    </div>
  );
}
