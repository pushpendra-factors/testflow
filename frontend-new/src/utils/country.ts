import { AVAILABLE_FLAGS, COUNTRY_LIST } from '../constants/country.list';
import _ from 'lodash';

export const isCountryFlagAvailable = (counrtyName: string): boolean => {
  const iso_code = getCountryCode(counrtyName);
  if (iso_code) return AVAILABLE_FLAGS.includes(iso_code);
  return false;
};

export const getCountryCode = (countryName: string): string | null => {
  if (!countryName || typeof countryName !== 'string') return null;
  const countryCode = COUNTRY_LIST.find((country) =>
    country.name.includes(countryName.toLowerCase())
  );
  if (countryCode) return countryCode.iso_code;
  return null;
};

export const getCountryNameFromIsoCode = (isoCode: string): string => {
  if (!isoCode || typeof isoCode !== 'string') return '';
  const countryName = COUNTRY_LIST.find(
    (country) => country.iso_code === isoCode
  );
  if (countryName) return _.startCase(countryName?.name?.[0]);
  return '';
};

export const getAllCountryIsoCodes = () => {
  return COUNTRY_LIST.map((country) => country.iso_code);
};
