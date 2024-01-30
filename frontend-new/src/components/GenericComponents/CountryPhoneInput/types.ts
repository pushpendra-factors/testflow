import { AVAILABLE_FLAGS } from 'Constants/country.list';

export type CountryPhoneInputValue = {
  code?: (typeof AVAILABLE_FLAGS)[number];
  phone?: string;
};
