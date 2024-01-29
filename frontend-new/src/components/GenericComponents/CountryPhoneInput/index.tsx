import { Input } from 'antd';
import React, { useCallback, useState } from 'react';
import { InputProps } from 'antd/es/input';
import { getCountryDialCodeFromIsoCode } from 'Utils/country';
import styles from './index.module.scss';
import CountrySelect from './CountrySelect';
import { CountryPhoneInputValue } from './types';

interface CountryPhoneInputProps
  extends Omit<InputProps, 'defaultValue' | 'value' | 'onChange'> {
  defaultValue: CountryPhoneInputValue;
  value?: CountryPhoneInputValue;
  onChange?: (value: CountryPhoneInputValue) => void;
  className?: string;
}

function CountryPhoneInput({
  defaultValue,
  onChange,
  className,
  ...inputProps
}: CountryPhoneInputProps) {
  const [countryCode, setCountryCode] = useState<string>(
    defaultValue?.code || 'US'
  );
  const [phone, setPhone] = useState<string>(defaultValue?.phone || '');
  const { value } = inputProps;

  const triggerChange = useCallback(
    (_phone: string, _countryCode: string) => {
      const dialCode = getCountryDialCodeFromIsoCode(_countryCode);
      if (!dialCode) return;
      const result: CountryPhoneInputValue = {
        phone: _phone,
        code: dialCode
      };
      onChange?.(result);
    },
    [onChange]
  );

  const handlePhoneChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const currentValue = e.target.value;
      setPhone(currentValue);
      triggerChange(currentValue, countryCode);
    },
    [setPhone, countryCode]
  );

  const handleCountryChange = useCallback(
    (_value: string) => {
      setCountryCode(_value);
      triggerChange(phone, _value);
    },
    [setCountryCode, triggerChange, phone]
  );

  return (
    <Input
      {...inputProps}
      className={`${styles.countryPhoneInput} ${className || ''}`}
      value={value?.phone}
      onChange={handlePhoneChange}
      addonBefore={
        <CountrySelect
          handleCountryChange={handleCountryChange}
          countryCode={countryCode}
        />
      }
    />
  );
}

export default CountryPhoneInput;
