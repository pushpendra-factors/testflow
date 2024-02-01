import React, {
  ReactNode,
  useCallback,
  useEffect,
  useRef,
  useState
} from 'react';
import styles from './index.module.scss';
import { Input, Spin } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import useAutoFocus from 'hooks/useAutoFocus';

import {
  ApplyClickCallbackType,
  OptionType,
  PlacementType,
  SingleSelectOptionClickCallbackType,
  Variant
} from './types';
import SingleSelect from './SingleSelect';
import MultiSelect from './MultiSelect';
import useDynamicPosition from 'hooks/useDynamicPosition';
import useKeyboardNavigation from 'hooks/useKeyboardNavigation';
interface FaSelectProps {
  options: OptionType[];
  optionClickCallback?: SingleSelectOptionClickCallbackType;
  applyClickCallback?: ApplyClickCallbackType;
  variant?: Variant;
  onClickOutside: any;
  allowSearch?: boolean;
  loadingState?: boolean;
  children?: ReactNode;
  extraClass?: string;
  placement?: PlacementType;
  // for multi select feature
  maxAllowedSelection?: number;
  // for allowing to select the search option
  allowSearchTextSelection?: boolean;
}

export default function FaSelect({
  options,
  optionClickCallback,
  applyClickCallback,
  variant = 'Single',
  onClickOutside,
  allowSearch = false,
  loadingState = false,
  children,
  extraClass = '',
  placement = 'BottomLeft',
  maxAllowedSelection = 0,
  allowSearchTextSelection = true
}: FaSelectProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [autoFocus, setAutofocus] = useState(false);
  const inputComponentRef = useAutoFocus(autoFocus);
  const dropdownRef = useRef(null);
  const relativeRef = useRef(null);

  const position = useDynamicPosition(relativeRef, dropdownRef, placement, 250);

  useEffect(() => {
    setAutofocus(true);

    return () => {
      setAutofocus(false);
    };
  }, []);
  const OnKeyDownEvent = useCallback(
    (e) => useKeyboardNavigation(dropdownRef, e),
    []
  );
  const renderSearchInput = () => {
    return (
      <div
        className={`${styles.selectInput} fa-filter-select fa-search-select`}
        onKeyDown={OnKeyDownEvent}
      >
        <Input
          tabIndex={0}
          style={{ overflow: 'hidden' }}
          prefix={<SVG name={'search'} />}
          size='large'
          placeholder={'Search'}
          onChange={(val) => {
            setSearchTerm(val.target.value);
          }}
          ref={inputComponentRef}
        ></Input>
      </div>
    );
  };
  const renderOptions = () => {
    //Options Loading
    if (loadingState) {
      return (
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
    }
    //No Data
    if (options.length === 0) {
      return (
        <div className='flex justify-center items-center my-2'>
          <Text
            level={7}
            type={'title'}
            extraClass={'ml-2'}
            weight={'thin'}
            color={'grey'}
          >
            No Options Available!
          </Text>
        </div>
      );
    }
    let searchOption: OptionType | null = null;
    if (searchTerm.length) {
      searchOption = { value: searchTerm, label: searchTerm };
    }
    if (variant === 'Multi') {
      return (
        <MultiSelect
          options={options}
          applyClickCallback={applyClickCallback}
          allowSearch={allowSearch}
          searchOption={searchOption}
          maxAllowedSelection={maxAllowedSelection}
          allowSearchTextSelection={allowSearchTextSelection}
          searchTerm={searchTerm}
        />
      );
    }
    return (
      <SingleSelect
        options={options}
        optionClickCallback={optionClickCallback}
        allowSearch={allowSearch}
        searchOption={searchOption}
        allowSearchTextSelection={allowSearchTextSelection}
        searchTerm={searchTerm}
      />
    );
  };
  return (
    <>
      <div ref={relativeRef}></div>
      {position && (
        <div
          className={`${extraClass}  ${styles.dropdown__select}
          ${
            position === 'TopRight' || position === 'BottomRight'
              ? styles.dropdown__select_right_0
              : styles.dropdown__select_left_0
          } fa-select  fa-select--group-select
         ${
           allowSearch
             ? `fa-select--group-select-sm`
             : `fa-select--group-select-mini`
         } ${
           position === 'Top' ||
           position === 'TopLeft' ||
           position === 'TopRight'
             ? styles.dropdown__select_placement_top
             : styles.dropdown__select_placement_bottom
         }`}
          ref={dropdownRef}
        >
          {allowSearch && renderSearchInput()}

          <div
            className={`fa-select-dropdown ${styles.dropdown__select__content}`}
            onKeyDown={OnKeyDownEvent}
          >
            {children || renderOptions()}
          </div>
        </div>
      )}
      <div
        className={styles.dropdown__hd_overlay}
        onClick={onClickOutside}
      ></div>
    </>
  );
}
