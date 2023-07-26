import React, { useEffect, useState } from 'react';
import { Button, Input } from 'antd';
import { SVG, Text } from '../factorsComponents';
import { getUniqueItemsByKeyAndSearchTerm } from '../Profile/utils';
import CustomCheckbox from './CustomCheckbox';
import { PropTextFormat } from 'Utils/dataFormatter';

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

  return (
    <>
      <Input
        placeholder={placeholder}
        prefix={searchIcon}
        onChange={handleSearch}
        value={searchTerm}
      />
      <div className='fa-custom--popover-content'>
        <div className={`${showApply ? 'apply_active' : ''}`}>
          {showDisabledOption && (
            <>
              {disabledOptions &&
                disabledOptions?.length > 0 &&
                disabledOptions?.map((option) => (
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
                      {option}
                    </Text>
                    <SVG size={16} name='Lock' />
                  </div>
                ))}
            </>
          )}
          {mapArray?.length ? (
            getUniqueItemsByKeyAndSearchTerm(mapArray, searchTerm).map(
              (option) => (
                <CustomCheckbox
                  key={option[titleKey]}
                  name={PropTextFormat(option[titleKey])}
                  checked={option[checkedKey]}
                  onChange={onChange.bind(this, option)}
                />
              )
            )
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
