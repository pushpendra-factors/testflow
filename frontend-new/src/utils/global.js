import _ from 'lodash';

export const EMPTY_FUNCTION = () => {};

export const EMPTY_STRING = '';

export const EMPTY_OBJECT = {};

export const EMPTY_ARRAY = [];

export const isStringLengthValid = (str, length = 1) =>
  _.size(_.trim(str)) >= length;

export const groupFilters = (array, key) =>
  // Return the end result
  array.reduce((result, currentValue) => {
    // If an array already present for key, push it to the array. Else create an array and push the object
    (result[currentValue[key]] = result[currentValue[key]] || []).push(
      currentValue
    );
    // Return the current iteration `result` value, this will be taken as next iteration `result` value and accumulate
    return result;
  }, {}); // empty object is the initial value for result object

export const compareFilters = (a, b) => {
  if (a.ref < b.ref) return -1;
  if (a.ref > b.ref) return 1;
  return 0;
};

export const toCapitalCase = (str) => {
  const lower = str.toLowerCase();
  return lower.charAt(0).toUpperCase() + lower.slice(1);
};

export function isNumeric(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

export const abbreviateNumber = (n) => {
  if (n < 1e3) return n;
  if (n >= 1e3 && n < 1e6) return `${+(n / 1e3).toFixed(1)}K`;
  if (n >= 1e6 && n < 1e9) return `${+(n / 1e6).toFixed(1)}M`;
  if (n >= 1e9 && n < 1e12) return `${+(n / 1e9).toFixed(1)}B`;
  if (n >= 1e12) return `${+(n / 1e12).toFixed(1)}T`;
  return null;
};
