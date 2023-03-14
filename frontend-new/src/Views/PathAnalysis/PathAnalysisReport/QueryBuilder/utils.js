import { EMPTY_ARRAY, groupFilters } from 'Utils/global';
import {
  operatorMap,
  reverseOperatorMap,
  reverseDateOperatorMap
} from 'Utils/operatorMapping';

export const getGlobalFilters = (globalFilters = []) => {
  const filterProps = [];
  
  const filtersGroupedByRef = Object.values(groupFilters(globalFilters, 'ref')); 

  filtersGroupedByRef.forEach((filtersGr) => {
    if (filtersGr.length === 1) {
      const fil = filtersGr[0];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            en: fil.props[2],
            lop: !index ? 'AND' : 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: fil.props[1] === 'datetime' ? val : val
          });
        });
      } else {
        filterProps.push({
          en: fil.props[2],
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
            en: fil.props[2],
            lop: !index ? 'AND' : 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: fil.props[1] === 'datetime' ? val : val
          });
        });
      } else {
        filterProps.push({
          en: fil.props[2],
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
            en: fil.props[2],
            lop: 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: fil.props[1] === 'datetime' ? val : val
          });
        });
      } else {
        filterProps.push({
          en: fil.props[2],
          lop: 'OR',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.props[1] === 'datetime' ? fil.values : fil.values
        });
      }
    }
  });
  return filterProps;
};

export const getGlobalFiltersfromSavedState = (savedFilterArr = []) => {
  let globalFilters = [];
  if (Array.isArray(savedFilterArr)) {
    let ref = -1;
    let lastProp = '';
    let lastOp = '';
    savedFilterArr.forEach((pr) => {
      if (pr.lop === 'AND') {
        ref += 1;
        globalFilters.push({
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
        globalFilters[globalFilters.length - 1].values.push(pr.va);
      } else {
        globalFilters.push({
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
      return globalFilters;
    });
  }
  return globalFilters;
};
