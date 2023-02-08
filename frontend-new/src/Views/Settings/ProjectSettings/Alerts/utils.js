import _ from 'lodash';
import { formatFilterDate } from 'Utils/dataFormatter';
import { groupFilters } from 'Utils/global';
import { FILTER_TYPES } from '../../../CoreQuery/constants';
import { convertDateTimeObjectValuesToMilliSeconds } from '../../../CoreQuery/utils';
import {
  operatorMap,
  reverseOperatorMap,
  reverseDateOperatorMap
} from 'Utils/operatorMapping';

export const getEventsWithProperties = (queries) => {
  const ewps = [];
  queries.forEach((ev) => {
    const filterProps = [];

    const filtersGroupedByRef = Object.values(groupFilters(ev.filters, 'ref'));
    filtersGroupedByRef.forEach((filtersGr) => {
      if (filtersGr.length === 1) {
        const fil = filtersGr[0];
        if (Array.isArray(fil.values)) {
          fil.values.forEach((val, index) => {
            filterProps.push({
              en: fil.props[2] === 'group' ? 'user' : fil.props[2],
              lop: !index ? 'AND' : 'OR',
              op: operatorMap[fil.operator],
              pr: fil.props[0],
              ty: fil.props[1],
              va: fil.props[1] === 'datetime' ? val : val
            });
          });
        } else {
          filterProps.push({
            en: fil.props[2] === 'group' ? 'user' : fil.props[2],
            lop: 'AND',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: fil.props[1] === 'datetime' ? fil.values : fil.values
          });
        }
      } else {
        let fil = filtersGr[0];
        if (Array.isArray(fil.values)) {
          fil.values.forEach((val, index) => {
            filterProps.push({
              en: fil.props[2] === 'group' ? 'user' : fil.props[2],
              lop: !index ? 'AND' : 'OR',
              op: operatorMap[fil.operator],
              pr: fil.props[0],
              ty: fil.props[1],
              va: fil.props[1] === 'datetime' ? val : val
            });
          });
        } else {
          filterProps.push({
            en: fil.props[2] === 'group' ? 'user' : fil.props[2],
            lop: 'AND',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: fil.props[1] === 'datetime' ? fil.values : fil.values
          });
        }
        fil = filtersGr[1];
        if (Array.isArray(fil.values)) {
          fil.values.forEach((val) => {
            filterProps.push({
              en: fil.props[2] === 'group' ? 'user' : fil.props[2],
              lop: 'OR',
              op: operatorMap[fil.operator],
              pr: fil.props[0],
              ty: fil.props[1],
              va: fil.props[1] === 'datetime' ? val : val
            });
          });
        } else {
          filterProps.push({
            en: fil.props[2] === 'group' ? 'user' : fil.props[2],
            lop: 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: fil.props[1] === 'datetime' ? fil.values : fil.values
          });
        }
      }
    });

    ewps.push({
      an: ev.alias,
      na: ev.label,
      grpa: ev.group,
      pr: filterProps
    });
  });
  return ewps[0]?.pr;
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
            category == 'channels' || category == 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : 'event',
          objTy:
            category == 'channels' || category == 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3]
                : 'event'
              : '',
          va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
        });
      });
    } else {
      filterProps.push({
        prNa: fil?.extra ? fil?.extra[1] : `$${_.lowerCase(fil?.props[0])}`,
        prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
        co: operatorMap[fil.operator],
        lOp: 'AND',
        en:
          category == 'channels' || category == 'custom_channels'
            ? ''
            : fil?.extra
            ? fil?.extra[3]
            : 'event',
        objTy:
          category == 'channels' || category == 'custom_channels'
            ? fil?.extra
              ? fil?.extra[3]
              : 'event'
            : '',
        va:
          fil.props[1] === 'datetime'
            ? formatFilterDate(fil.values)
            : fil.values
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
        extra: [DNa, pr.prNa, pr.prDaTy]
      });
    } else if (pr.prDaTy === FILTER_TYPES.CATEGORICAL) {
      filters[filters.length - 1].values.push(pr.va);
    }
  });
  return filters;
};

export const getStateFromFiltersEvent = (rawFilters = []) => {
  const filters = [];
  rawFilters.forEach((pr) => {
    if (pr.lop === 'AND') {
      filters.push({
        operator:
          pr.ty === 'datetime'
            ? reverseDateOperatorMap[pr.op]
            : reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: [pr.va]
      });
    } else {
      filters[filters.length - 1].values.push(pr.va);
    }
  });
  return filters;
};
