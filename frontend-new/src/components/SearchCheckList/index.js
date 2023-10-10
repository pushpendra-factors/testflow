import React, { useEffect, useState } from 'react';
import { Button, Input } from 'antd';
import { SVG, Text } from '../factorsComponents';
import { getUniqueItemsByKeyAndSearchTerm } from '../Profile/utils';
import CustomCheckbox from './CustomCheckbox';
import { PropTextFormat } from 'Utils/dataFormatter';
import VirtualList from 'rc-virtual-list';

export default function SearchCheckList({
  placeholder,
  searchIcon = <SVG name='search' size={16} color='grey' />,
  mapArray = [],
  titleKey,
  checkedKey,
  onChange,
  emptyListText,
  showApply = false,
  onApply,
  showDisabledOption = false,
  disabledOptions = null,
  handleDisableOptionClick
}) {
  const [searchTerm, setSearchTerm] = useState('');
  const handleSearch = (e) => {
    setSearchTerm(e.target.value || '');
  };

  useEffect(() => {
    setSearchTerm('');
  }, [mapArray]);

  const getListWithDisabledOptions = (list) => {
    if (!showDisabledOption) return list;
    if (!disabledOptions?.length) return list;
    const disabledOptionsList =
      disabledOptions
        .map((option) => {
          return { name: option, isDisabled: true };
        })
        .filter((option) =>
          option?.name?.toLowerCase()?.includes(searchTerm?.toLowerCase())
        ) || [];
    return [...disabledOptionsList, ...list];
  };

  return (
    <>
      <Input
        placeholder={placeholder}
        prefix={searchIcon}
        onChange={handleSearch}
        value={searchTerm}
      />
      <div>
        <div className={`${showApply ? 'apply_active' : ''}`}>
          {mapArray?.length ? (
            <VirtualList
              data={getListWithDisabledOptions(
                getUniqueItemsByKeyAndSearchTerm(mapArray, searchTerm)
              )}
              height={showApply ? 348 : 392}
              itemHeight={38}
              itemKey={titleKey}
            >
              {(item, index) => {
                if (item.isDisabled) {
                  return (
                    <div
                      className='flex justify-between items-center py-2 px-3 cursor-not-allowed'
                      onClick={(option) =>
                        handleDisableOptionClick &&
                        handleDisableOptionClick(option)
                      }
                    >
                      <Text
                        type='title'
                        level={7}
                        extraClass='mb-0 truncate'
                        truncate
                        charLimit={25}
                      >
                        {item.name}
                      </Text>
                      <SVG size={16} name='Lock' />
                    </div>
                  );
                }
                return (
                  <CustomCheckbox
                    key={item[titleKey]}
                    name={PropTextFormat(item[titleKey])}
                    checked={item[checkedKey]}
                    onChange={onChange.bind(this, item)}
                  />
                );
              }}
            </VirtualList>
          ) : (
            <div className='text-center p-2 italic'>{emptyListText}</div>
          )}
        </div>
        {showApply ? (
          <Button type='primary' className='w-full' onClick={onApply}>
            Apply
          </Button>
        ) : null}
      </div>
    </>
  );
}
