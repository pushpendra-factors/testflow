import { AVAILABLE_FLAGS, COUNTRY_LIST } from '../constants/country.list';

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
