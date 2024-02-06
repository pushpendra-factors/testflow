import React, { useCallback, useEffect, useRef, useState } from 'react';
import { SVG, Text } from '../../factorsComponents';
import { OptionType, PlacementType } from '../FaSelect/types';
import useAutoFocus from 'hooks/useAutoFocus';
import { Button, Input, Spin } from 'antd';
import styles from './index.module.scss';
import SingleSelect from '../FaSelect/SingleSelect';
import {
  GroupSelectOptionClickCallbackType,
  GroupSelectOptionType
} from './types';
import { getIcon } from './utils';
import useKey from 'hooks/useKey';
import { HighlightSearchText } from 'Utils/dataFormatter';
import useDynamicPosition from 'hooks/useDynamicPosition';
import { debounce } from 'lodash';
interface GroupSelectProps {
  options: GroupSelectOptionType[];
  optionClickCallback: GroupSelectOptionClickCallbackType;
  loadingState?: boolean;
  onClickOutside: any;
  allowSearch?: boolean;
  extraClass?: string;
  placement?: PlacementType;
  searchPlaceHolder?: string;
  // for allowing to select the search option
  allowSearchTextSelection?: boolean;
}
export default function GroupSelect({
  options,
  optionClickCallback,
  loadingState = false,
  onClickOutside,
  extraClass = '',
  placement = 'Bottom',
  allowSearch = false,
  searchPlaceHolder = 'Search',
  allowSearchTextSelection = true
}: GroupSelectProps) {
  const [groupSelectorOpen, setGroupSelectorOpen] = useState(true);
  useEffect(() => {
    if (options && options.length === 1) setGroupSelectorOpen(false);
    else setGroupSelectorOpen(true);
  }, [options]);
  const [selectedGroupIndex, setSelectedGroupIndex] = useState(0);

  const [searchTerm, setSearchTerm] = useState('');
  const inputComponentRef = useAutoFocus(allowSearch);
  const dropdownRef = useRef(null);
  const relativeRef = useRef(null);

  const position = useDynamicPosition(relativeRef, dropdownRef, placement, 400);
  const renderSearchInput = () => {
    return (
      <div className={`fa-filter-select fa-search-select pb-0`}>
        <Input
          style={{ overflow: 'hidden' }}
          prefix={<SVG name={'search'} />}
          size='large'
          placeholder={searchPlaceHolder}
          onChange={(event) => {
            updateSearchText(event.target.value);
          }}
          ref={inputComponentRef}
          autoFocus={true}
        ></Input>
      </div>
    );
  };
  const updateSearchText = debounce((value) => {
    setSearchTerm(value);
  }, 200);

  const handleGroupSelectClickWithSearchEmpty = (group: OptionType) => {
    //When Search is Empty in Group Dropdown.
    setSelectedGroupIndex(options?.map((op) => op.label)?.indexOf(group.label));
    setGroupSelectorOpen(false);
  };
  const handleGroupSelectClickWithSearch = (valueOption: OptionType) => {
    //When Search is not Empty in Group Dropdown.
    const index = options
      ?.map((op) => op.label)
      ?.indexOf(valueOption.extraProps.groupLabel);
    optionClickCallback(
      valueOption,
      index > -1
        ? options[index]
        : { label: valueOption.label, values: [valueOption] }
    );
  };
  const handleOptionSelectClick = (option: OptionType) => {
    optionClickCallback(option, options[selectedGroupIndex]);
  };
  const handleOptionBackClick = useCallback(() => {
    setGroupSelectorOpen(true);
    setSelectedGroupIndex(-1);
  }, []);

  const generateOptionHeader = () => {
    const selectedGroup = options[selectedGroupIndex];
    return (
      <div className='flex flex-row justify-between items-center w-full'>
        <div className='flex flex-row justify-between items-center'>
          <SVG
            name={getIcon(selectedGroup?.iconName || '')}
            extraClass={'self-center'}
            size={18}
          ></SVG>
          <Text
            level={7}
            type={'title'}
            extraClass={'m-0 ml-2'}
            weight={'bold'}
            size={14}
          >
            {selectedGroup?.label}
          </Text>
          <div className={`${styles.numberTag} ml-1`}>
            {selectedGroup?.values?.length}
          </div>
        </div>
        <Button
          icon={
            <SVG name={'chevronLeft'} extraClass={'self-center'} size={16} />
          }
          type='text'
          onClick={handleOptionBackClick}
        >
          Back
        </Button>
      </div>
    );
  };
  const renderGroupFaSelect = () => {
    if (searchTerm.length) {
      let groupValueOptions: OptionType[] = [];
      //The Value is Modified, to extract the group information when Value-Labels are same.
      options.forEach((group) => {
        if (group.label !== 'Most Recent') {
          group.values?.forEach((groupValue) => {
            groupValueOptions.push({
              value: groupValue?.value,
              label: groupValue?.label,
              labelNode: (
                <div
                  className='flex flex-row items-center'
                  title={groupValue?.label}
                >
                  <div className='flex'>
                    <SVG
                      name={getIcon(group?.iconName || '')}
                      extraClass={'self-center'}
                      size={18}
                    ></SVG>
                  </div>
                  <div className='flex'>
                    <Text
                      level={7}
                      type={'title'}
                      extraClass={'m-0 ml-2'}
                      weight={'medium'}
                    >
                      <HighlightSearchText
                        text={groupValue?.label}
                        highlight={searchTerm}
                      />
                    </Text>
                  </div>
                </div>
              ),
              extraProps: {
                groupLabel: group?.label,
                ...groupValue.extraProps
              }
            });
          });
        }
      });

      const searchOption: OptionType = {
        label: searchTerm,
        value: searchTerm,
        extraProps: {
          groupLabel: searchTerm
        }
      };
      return (
        <SingleSelect
          key={'group-with-search'}
          options={groupValueOptions}
          optionClickCallback={(group) =>
            handleGroupSelectClickWithSearch(group)
          }
          allowSearch={true}
          searchOption={searchOption}
          searchTerm={searchTerm}
          allowSearchTextSelection={allowSearchTextSelection}
          extraClass={styles.dropdown__select__content__options}
        />
      );
    }
    const groupOptions: OptionType[] = options.map((group) => {
      return {
        value: group?.label,
        label: group?.label,
        labelNode: (
          <div
            className='flex flex-row justify-between w-full items-center	'
            title={group?.label}
          >
            <div className='flex flex-row items-center'>
              <div className='flex'>
                <SVG
                  name={getIcon(group?.iconName || '')}
                  extraClass={'self-center'}
                  size={18}
                ></SVG>
              </div>
              <div className='flex justify-between items-center'>
                <Text
                  level={7}
                  type={'title'}
                  extraClass={'m-0 ml-2'}
                  weight={'medium'}
                >
                  {group?.label}
                </Text>
                <div className={`${styles.numberTag} ml-1`}>
                  {group?.values?.length}
                </div>
              </div>
            </div>
            <div className='flex flex-row'>
              <SVG
                name={'chevronRight'}
                extraClass={'self-center'}
                size={'16'}
              ></SVG>
            </div>
          </div>
        )
      };
    });
    return (
      <SingleSelect
        key={'group-without-search'}
        options={groupOptions}
        optionClickCallback={handleGroupSelectClickWithSearchEmpty}
        allowSearch={true}
        searchOption={null}
        searchTerm={searchTerm}
        allowSearchTextSelection={false}
        extraClass={styles.dropdown__select__content__options}
      />
    );
  };
  const renderOptionFaSelect = () => {
    const selectedGroup = options[selectedGroupIndex];
    return (
      <SingleSelect
        key={'group-values-options'}
        options={selectedGroup?.values || []}
        optionClickCallback={handleOptionSelectClick}
        allowSearch={true}
        searchOption={null}
        searchTerm={searchTerm}
        allowSearchTextSelection={false}
      />
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
    return groupSelectorOpen ? renderGroupFaSelect() : renderOptionFaSelect();
  };

  useKey(['Escape'], handleOptionBackClick);

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
          {!groupSelectorOpen && options.length > 1 && (
            <div className={`${styles.dropdown__select__header}`}>
              {generateOptionHeader()}
            </div>
          )}
          {allowSearch && renderSearchInput()}

          <div
            className={`fa-select-dropdown ${styles.dropdown__select__content}`}
          >
            {renderOptions()}
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
