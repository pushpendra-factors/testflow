import { AVAILABLE_FLAGS, COUNTRY_LIST } from 'Constants/country.list';
import {
  getCountryDialCodeFromIsoCode,
  getCountryNameFromIsoCode
} from 'Utils/country';
import { Select } from 'antd';
import React, { useMemo } from 'react';

interface CountrySelectProps {
  countryCode: string;
  handleCountryChange: (value: string) => void;
}

function CountrySelect({
  countryCode,
  handleCountryChange
}: CountrySelectProps) {
  const AvailableCountries = useMemo(
    () =>
      COUNTRY_LIST.filter((country) => !!country?.dial_code).filter((country) =>
        AVAILABLE_FLAGS.includes(country.iso_code)
      ),
    []
  );

  return (
    <Select
      dropdownStyle={{ width: 300 }}
      filterOption={(input, option) =>
        (option?.value
          ? getCountryNameFromIsoCode(option?.value).toLowerCase()
          : ''
        ).includes(input.toLowerCase())
      }
      value={countryCode}
      showSearch
      showArrow
      onChange={handleCountryChange}
      optionLabelProp='label'
      dropdownMatchSelectWidth={false}
    >
      {AvailableCountries.map((country) => (
        <Select.Option
          key={country.iso_code}
          value={country.iso_code}
          label={
            <>
              <div className={`fflag fflag-${countryCode} ff-sm`} />{' '}
              <span>{getCountryDialCodeFromIsoCode(countryCode)}</span>
            </>
          }
        >
          <div className='flex items-center gap-2 justify-start '>
            <div className={`fflag fflag-${country.iso_code} ff-sm`} />
            {getCountryNameFromIsoCode(country.iso_code)}{' '}
            {getCountryDialCodeFromIsoCode(country.iso_code)}
          </div>
        </Select.Option>
      ))}
    </Select>
  );
}

export default CountrySelect;
