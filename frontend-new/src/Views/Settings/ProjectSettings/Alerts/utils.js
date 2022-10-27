import _ from 'lodash';
import { formatFilterDate } from 'Utils/dataFormatter';
import { FILTER_TYPES } from '../../../CoreQuery/constants';
import { convertDateTimeObjectValuesToMilliSeconds, reverseDateOperatorMap, reverseOperatorMap } from '../../../CoreQuery/utils';

const operatorMap = {
    '=': 'equals',
    '!=': 'notEqual',
    contains: 'contains',
    'does not contain': 'notContains',
    '<': 'lesserThan',
    '<=': 'lesserThanOrEqual',
    '>': 'greaterThan',
    '>=': 'greaterThanOrEqual',
    between: 'between',
    'not between': 'notInBetween',
    'in the previous': 'inLast',
    'not in the previous': 'notInLast',
    'in the current': 'inCurrent',
    'not in the current': 'notInCurrent',
    before: 'before',
    since: 'since',
  };

export const getEventsWithPropertiesKPI = (filters, category) => {
    const filterProps = [];
    // adding fil?.extra ? fil?.extra[*] check as a hotfix for timestamp filters
    filters.forEach((fil) => {
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : `$${_.lowerCase(fil?.props[0])}`,
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              (category == 'channels' || category == 'custom_channels') ? '' : fil?.extra ? fil?.extra[3] : 'event',
            objTy:
              (category == 'channels' || category == 'custom_channels')
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val,
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : `$${_.lowerCase(fil?.props[0])}`,
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en: (category == 'channels' || category == 'custom_channels') ? '' : fil?.extra ? fil?.extra[3] : 'event',
          objTy:
            (category == 'channels' || category == 'custom_channels') ? (fil?.extra ? fil?.extra[3] : 'event') : '',
          va:
            fil.props[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values,
        });
      }
    });
    return filterProps;
  };

  export const getStateFromFilters = (rawFilters) => {
    const filters = [];

    rawFilters.forEach((pr) => {
      if (pr.lOp === 'AND') {
        const val = pr.prDaTy === FILTER_TYPES.CATEGORICAL ? [pr.va] : pr.va;

        const DNa = _.startCase(pr.prNa);

        filters.push({
          operator:
            pr.prDaTy === 'datetime'
              ? reverseDateOperatorMap[pr.co]
              : reverseOperatorMap[pr.co],
          props: [DNa, pr.prDaTy, 'filter'],
          values:
            pr.prDaTy === FILTER_TYPES.DATETIME
              ? convertDateTimeObjectValuesToMilliSeconds(val)
              : val,
          extra: [DNa, pr.prNa, pr.prDaTy],
        });
      } else if (pr.prDaTy === FILTER_TYPES.CATEGORICAL) {
        filters[filters.length - 1].values.push(pr.va);
      }
    });
    return filters;
  };
