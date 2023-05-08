import React, { ReactNode, useState } from 'react';
import styles from './index.module.scss';
import { Input, Spin } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import useAutoFocus from 'hooks/useAutoFocus';
import { DISPLAY_PROP } from 'Utils/constants';

import {
  OptionType,
  PlacementType,
  Variant,
  handleOptionFunctionType
} from './types';
import SingleSelect from './SingleSelect';
import MultiSelect from './MultiSelect';

interface FaSelectProps {
  options: OptionType[];
  optionClick: handleOptionFunctionType;
  selectType?: Variant;
  onClickOutside: any;
  selectedOptions?: string[];
  allowSearch?: boolean;
  loadingState?: boolean;
  children?: ReactNode;
  extraClass?: string;
  placement?: PlacementType;
}

export default function FaSelect({
  options,
  optionClick,
  selectType = 'Single',
  onClickOutside,
  selectedOptions = [],
  allowSearch = false,
  loadingState = true,
  children,
  extraClass = '',
  placement = 'BottomLeft'
}: FaSelectProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const inputComponentRef = useAutoFocus(allowSearch);
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
      //Reducing Options Based On Search.
      options = options.filter((op) => {
        let searchTermLowerCase = searchTerm.toLowerCase();
        // Regex to detect https/http is there or not as a protocol
        let testURLRegex =
          /^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$/;
        if (testURLRegex.test(searchTermLowerCase)) {
          searchTermLowerCase = searchTermLowerCase.split('://')[1];
        }
        searchTermLowerCase = searchTermLowerCase.replace(/\/$/, '');
        return (
          op.label.toLowerCase().includes(searchTermLowerCase) ||
          (op.label === '$none' &&
            DISPLAY_PROP[op.label].toLowerCase().includes(searchTermLowerCase))
        );
      });
      searchOption = { value: searchTerm, label: searchTerm };
    }
    if (selectType === 'Multi') {
      return (
        <MultiSelect
          options={options}
          selectedOptions={selectedOptions}
          optionClick={optionClick}
          allowSearch={allowSearch}
          searchOption={searchOption}
        />
      );
    }
    return (
      <SingleSelect
        options={options}
        optionClick={optionClick}
        allowSearch={allowSearch}
        searchOption={searchOption}
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
