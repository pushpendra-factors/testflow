import React, { useState } from 'react';
import { Input } from 'antd';
import { SVG } from '../factorsComponents';
import { getUniqueItemsByKeyAndSearchTerm } from '../Profile/utils';
import CustomCheckbox from './CustomCheckbox';

export default function SearchCheckList({
  placeholder,
  searchIcon = <SVG name="search" size={16} color="grey" />,
  mapArray = [],
  titleKey,
  checkedKey,
  onChange,
  emptyListText
}) {
  const [searchTerm, setSearchTerm] = useState('');
  const handleSearch = (e) => {
    setSearchTerm(e.target.value || '');
  };

  return (
    <>
      <Input
        placeholder={placeholder}
        prefix={searchIcon}
        onChange={handleSearch}
        value={searchTerm}
      />
      <div className="fa-custom--popover-content">
        {mapArray?.length ? (
          getUniqueItemsByKeyAndSearchTerm(mapArray, searchTerm).map(
            (option) => (
              <CustomCheckbox
                key={option[titleKey]}
                name={option[titleKey]}
                checked={option[checkedKey]}
                onChange={onChange.bind(this, option)}
              />
            )
          )
        ) : (
          <div className="text-center p-2 italic">{emptyListText}</div>
        )}
      </div>
    </>
  );
}
