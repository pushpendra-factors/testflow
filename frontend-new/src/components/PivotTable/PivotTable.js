import React, { useState, memo, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import PivotTableUI from 'react-pivottable/PivotTableUI';

import { EMPTY_ARRAY } from 'Utils/global';

import {
  formatPivotData,
  getColumnOptions,
  getRowOptions,
  getValueOptions,
  SortRowOptions,
  getFunctionOptions,
} from './pivotTable.helpers';
import styles from './pivotTable.module.scss';
import ControlledComponent from '../ControlledComponent';
import PivotTableControls from '../PivotTableControls';
import { SORT_ORDERS } from '../PivotTableControls/pivotTableControls.constants';

const PivotTable = ({ data, breakdown, kpis, showControls }) => {
  const [pivotState, setPivotState] = useState({
    aggregatorName: 'Sum',
    vals: [kpis[0]],
    data: EMPTY_ARRAY,
    rows: EMPTY_ARRAY,
    cols: EMPTY_ARRAY,
    rowOrder: SORT_ORDERS.ASCEND,
  });

  const updateState = useCallback((updates) => {
    setPivotState((currState) => {
      return {
        ...currState,
        ...updates,
      };
    });
  }, []);

  const handleRowAttributeRemove = useCallback((attr) => {
    setPivotState((currState) => {
      return {
        ...currState,
        rows: currState.rows.filter((row) => row !== attr),
      };
    });
  }, []);

  const handleColumnAttributeRemove = useCallback(() => {
    updateState({ cols: [] });
  }, []);

  const handleValueChange = useCallback(
    (val) => {
      updateState({
        vals: [val],
      });
    },
    [updateState]
  );

  const handleFunctionChange = useCallback(
    (val) => {
      updateState({
        aggregatorName: val,
      });
    },
    [updateState]
  );

  const handleColumnChange = useCallback(
    (val) => {
      setPivotState((currState) => {
        return {
          ...currState,
          cols: [val],
          rows: currState.rows.filter((row) => row !== val),
        };
      });
    },
    [updateState]
  );

  const handleRowOptionSelect = useCallback((val) => {
    setPivotState((currState) => {
      return {
        ...currState,
        cols: currState.cols.filter((col) => col !== val),
        rows: SortRowOptions({
          data: [...currState.rows, val],
          kpis,
          breakdown,
        }),
      };
    });
  }, []);

  const handleSortChange = useCallback(() => {
    setPivotState((currState) => {
      return {
        ...currState,
        rowOrder:
          currState.rowOrder === SORT_ORDERS.ASCEND
            ? SORT_ORDERS.DESCEND
            : SORT_ORDERS.ASCEND,
      };
    });
  }, []);

  useEffect(() => {
    const [breakdownAttributes, attributes, values] = formatPivotData({
      data,
      breakdown,
      kpis,
    });

    updateState({
      rows: attributes,
      data: [attributes, ...values],
      hiddenFromAggregators: breakdownAttributes,
    });
  }, [data, breakdown, kpis, updateState]);

  return (
    <div className={styles.pivotTable}>
      <ControlledComponent controller={showControls}>
        <PivotTableControls
          selectedCol={
            pivotState.cols && pivotState.cols.length
              ? pivotState.cols[0]
              : null
          }
          selectedRows={pivotState.rows ? pivotState.rows : EMPTY_ARRAY}
          selectedValue={pivotState.vals[0]}
          aggregatorOptions={getValueOptions({ kpis })}
          columnOptions={getColumnOptions({ breakdown })}
          rowOptions={getRowOptions({
            selectedRows: pivotState.rows,
            kpis,
            breakdown,
          })}
          functionOptions={getFunctionOptions()}
          aggregatorName={pivotState.aggregatorName}
          onRowAttributeRemove={handleRowAttributeRemove}
          onColumnAttributeRemove={handleColumnAttributeRemove}
          onValueChange={handleValueChange}
          onColumnChange={handleColumnChange}
          onRowChange={handleRowOptionSelect}
          onFunctionChange={handleFunctionChange}
          rowOrder={pivotState.rowOrder}
          onSortChange={handleSortChange}
        />
      </ControlledComponent>

      <PivotTableUI
        onChange={(s) => {
          console.log(s.rowOrder);
          setPivotState(s);
        }}
        {...pivotState}
      />
    </div>
  );
};

export default memo(PivotTable);

PivotTable.propTypes = {
  data: PropTypes.array,
  breakdown: PropTypes.array,
  kpis: PropTypes.array,
  showControls: PropTypes.bool,
};

PivotTable.defaultProps = {
  data: EMPTY_ARRAY,
  breakdown: EMPTY_ARRAY,
  kpis: EMPTY_ARRAY,
  showControls: true,
};
