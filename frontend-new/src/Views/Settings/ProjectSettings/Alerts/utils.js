import _ from 'lodash';
import { formatFilterDate } from 'Utils/dataFormatter';
import { groupFilters } from 'Utils/global';
import { FILTER_TYPES } from '../../../CoreQuery/constants';
import {
  convertDateTimeObjectValuesToMilliSeconds,
  reverseDateOperatorMap,
  reverseOperatorMap
} from '../../../CoreQuery/utils';

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
  since: 'since'
};

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
  const filtersGroupedByRef = Object.values(groupFilters(filters, 'ref'));
  filtersGroupedByRef.forEach((filtersGr) => {
    if (filtersGr.length === 1) {
      const fil = filtersGr[0];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : 'event',
            objTy:
              category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : 'event',
          objTy:
            category === 'channels' || category === 'custom_channels'
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
    } else {
      let fil = filtersGr[0];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : 'event',
            objTy:
              category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : 'event',
          objTy:
            category === 'channels' || category === 'custom_channels'
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
      fil = filtersGr[1];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : 'event',
            objTy:
              category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'OR',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : 'event',
          objTy:
            category === 'channels' || category === 'custom_channels'
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
    }
  });
  return filterProps;
};

// export const getStateFromFiltersEvent = (rawFilters) => {
//   const filters = [];

//   rawFilters.forEach((pr) => {
//     if (pr.lOp === 'AND') {
//       const val = pr.prDaTy === FILTER_TYPES.CATEGORICAL ? [pr.va] : pr.va;

//       const DNa = _.startCase(pr.prNa);

//       filters.push({
//         operator:
//           pr.prDaTy === 'datetime'
//             ? reverseDateOperatorMap[pr.co]
//             : reverseOperatorMap[pr.co],
//         props: [DNa, pr.prDaTy, 'filter'],
//         values:
//           pr.prDaTy === FILTER_TYPES.DATETIME
//             ? convertDateTimeObjectValuesToMilliSeconds(val)
//             : val,
//         extra: [DNa, pr.prNa, pr.prDaTy],
//       });
//     } else if (pr.prDaTy === FILTER_TYPES.CATEGORICAL) {
//       filters[filters.length - 1].values.push(pr.va);
//     }
//   });
//   return filters;
// };

export const getStateFromFilters = (rawFilters) => {
  const eventFilters = [];

  let ref = -1,
    lastProp = '',
    lastOp = '';
  rawFilters.forEach((pr) => {
    if (pr.lOp === 'AND') {
      ref += 1;
      const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
      const DNa = _.startCase(pr.prNa);
      const isCamp =
        pr?.ca === 'channels' || pr?.ca === 'custom_channels'
          ? pr.objTy
          : pr.en;
      eventFilters.push({
        operator:
          pr.prDaTy === 'datetime'
            ? reverseDateOperatorMap[pr.co]
            : reverseOperatorMap[pr.co],
        props: [DNa, pr.prDaTy, isCamp],
        values:
          pr.prDaTy === FILTER_TYPES.DATETIME
            ? convertDateTimeObjectValuesToMilliSeconds(val)
            : val,
        extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
        ref
      });
      lastProp = pr.prNa;
      lastOp = pr.co;
    } else if (lastProp === pr.prNa && lastOp === pr.co) {
      eventFilters[eventFilters.length - 1].values.push(pr.va);
    } else {
      const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
      const DNa = _.lowerCase(pr.prNa);
      const isCamp =
        pr?.ca === 'channels' || pr?.ca === 'custom_channels'
          ? pr.objTy
          : pr.en;
      eventFilters.push({
        operator:
          pr.prDaTy === 'datetime'
            ? reverseDateOperatorMap[pr.co]
            : reverseOperatorMap[pr.co],
        props: [DNa, pr.prDaTy, isCamp],
        values:
          pr.prDaTy === FILTER_TYPES.DATETIME
            ? convertDateTimeObjectValuesToMilliSeconds(val)
            : val,
        extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
        ref
      });
      lastProp = pr.prNa;
      lastOp = pr.co;
    }
  });
  return eventFilters;
};

export const getStateFromFiltersEvent = (rawFilter) => {
  const filters = [];
  let ref = -1,
    lastProp = '',
    lastOp = '';
  rawFilter.forEach((pr) => {
    if (pr.lop === 'AND') {
      ref += 1;
      filters.push({
        operator:
          pr.ty === 'datetime'
            ? reverseDateOperatorMap[pr.op]
            : reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: [pr.va],
        ref
      });
      lastProp = pr.pr;
      lastOp = pr.op;
    } else if (lastProp === pr.pr && lastOp === pr.op) {
      filters[filters.length - 1].values.push(pr.va);
    } else {
      filters.push({
        operator:
          pr.ty === 'datetime'
            ? reverseDateOperatorMap[pr.op]
            : reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: [pr.va],
        ref
      });
      lastProp = pr.pr;
      lastOp = pr.op;
    }
  });
  return filters;
};
