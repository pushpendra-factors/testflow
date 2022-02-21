import _ from 'lodash'

export const EMPTY_FUNCTION = () => { }

export const EMPTY_STRING = '';

export const EMPTY_OBJECT = {};

export const EMPTY_ARRAY = [];

export const isStringLengthValid = (str, length = 1) => {
  return _.size(_.trim(str)) >= length
}