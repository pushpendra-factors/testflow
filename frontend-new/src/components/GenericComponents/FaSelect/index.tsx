import React, { ReactNode, useState } from 'react';
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
import { InfoCircleOutlined } from '@ant-design/icons';

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
  const inputComponentRef = useAutoFocus(allowSearch);
  const [showMaxLimitWarning, setShowMaxLimitWarning] =
    useState<boolean>(false);
  const renderSearchInput = () => {
    return (
      <div
        className={`${styles.selectInput} fa-filter-select fa-search-select`}
      >
        <Input
          style={{ overflow: 'hidden' }}
          prefix={<SVG name={'search'} />}
          size='large'
          placeholder={'Search'}
          onChange={(val) => {
            setSearchTerm(val.target.value);
          }}
          ref={inputComponentRef}
        ></Input>
        {showMaxLimitWarning && (
          <div className='flex gap-2 my-2 items-center'>
            <InfoCircleOutlined style={{ color: '#8C8C8C', fontSize: 14 }} />
            <Text type={'paragraph'} mini extraClass='m-0' color='grey'>
              You can only add up to {maxAllowedSelection} items at a time
            </Text>
          </div>
        )}
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
          setShowMaxLimitWarning={setShowMaxLimitWarning}
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
      <div
        className={`${extraClass}  ${styles.dropdown__select}
          ${
            placement === 'Right' ||
            placement === 'TopRight' ||
            placement === 'BottomRight'
              ? styles.dropdown__select_right_0
              : styles.dropdown__select_left_0
          } fa-select  fa-select--group-select
         ${
           allowSearch
             ? `fa-select--group-select-sm`
             : `fa-select--group-select-mini`
         } ${
          placement === 'Top' ||
          placement === 'TopLeft' ||
          placement === 'TopRight'
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
