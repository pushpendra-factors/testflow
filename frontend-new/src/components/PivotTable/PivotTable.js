import React, {
  useState,
  memo,
  useEffect,
  useCallback,
  useContext,
  useMemo
} from 'react';
import PropTypes from 'prop-types';
import { useSelector } from 'react-redux';
import PivotTableUI from 'react-pivottable/PivotTableUI';

import { EMPTY_ARRAY } from 'Utils/global';
import { QUERY_TYPE_KPI } from 'Utils/constants';

import {
  formatPivotData,
  getColumnOptions,
  getRowOptions,
  getValueOptions,
  SortRowOptions,
  getFunctionOptions,
  getMetricLabel
} from './pivotTable.helpers';
import styles from './pivotTable.module.scss';
import ControlledComponent from '../ControlledComponent';
import PivotTableControls from '../PivotTableControls';
import { PIVOT_SORT_ORDERS } from '../PivotTableControls/pivotTableControls.constants';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';

const PivotTableComponent = (props) => {
  const {
    data,
    breakdown,
    metrics,
    showControls,
    pivotConfig,
    updatePivotConfig,
    queryType
  } = props;

  const { eventNames, userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );

  const { configLoaded } = pivotConfig;

  const [pivotState, setPivotState] = useState({
    data: EMPTY_ARRAY,
    hiddenFromAggregators: EMPTY_ARRAY
  });

  const updateState = useCallback((updates) => {
    setPivotState((currState) => {
      return {
        ...currState,
        ...updates
      };
    });
  }, []);

  const handleRowAttributeRemove = useCallback(
    (attr) => {
      updatePivotConfig({
        rows: pivotConfig.rows.filter((row) => row !== attr)
      });
    },
    [pivotConfig, updatePivotConfig]
  );

  const handleColumnAttributeRemove = useCallback(() => {
    updatePivotConfig({ cols: [] });
  }, [updatePivotConfig]);

  const handleValueChange = useCallback(
    (val) => {
      updatePivotConfig({
        vals: [val]
      });
    },
    [updatePivotConfig]
  );

  const handleFunctionChange = useCallback(
    (val) => {
      updatePivotConfig({
        aggregatorName: val
      });
    },
    [updatePivotConfig]
  );

  const handleColumnChange = useCallback(
    (val) => {
      updatePivotConfig({
        cols: [val],
        rows: pivotConfig.rows.filter((row) => row !== val)
      });
    },
    [pivotConfig, updatePivotConfig]
  );

  const handleRowOptionSelect = useCallback(
    (val) => {
      updatePivotConfig({
        cols: pivotConfig.cols.filter((col) => col !== val),
        rows: SortRowOptions({
          data: [...pivotConfig.rows, val],
          metrics,
          breakdown,
          queryType,
          eventNames,
          userPropNames,
          eventPropNames
        })
      });
    },
    [
      updatePivotConfig,
      pivotConfig,
      metrics,
      queryType,
      breakdown,
      eventNames,
      userPropNames,
      eventPropNames
    ]
  );

  const handleSortChange = useCallback(() => {
    updatePivotConfig({
      rowOrder:
        pivotConfig.rowOrder === PIVOT_SORT_ORDERS.ASCEND
          ? PIVOT_SORT_ORDERS.DESCEND
          : PIVOT_SORT_ORDERS.ASCEND
    });
  }, [updatePivotConfig, pivotConfig]);

  useEffect(() => {
    const [breakdownAttributes, attributes, values] = formatPivotData({
      data,
      breakdown,
      metrics,
      queryType,
      eventNames,
      userPropNames,
      eventPropNames
    });

    if (!configLoaded) {
      updatePivotConfig({
        rows: breakdownAttributes,
        vals: [getMetricLabel({ metric: metrics[0], queryType })],
        configLoaded: true
      });
    }

    updateState({
      data: [attributes, ...values],
      hiddenFromAggregators: breakdownAttributes
    });
  }, [
    data,
    breakdown,
    metrics,
    updateState,
    configLoaded,
    updatePivotConfig,
    queryType,
    eventNames,
    userPropNames,
    eventPropNames
  ]);

  const aggregatorOptions = useMemo(() => {
    return getValueOptions({
      metrics,
      queryType,
      eventNames
    });
  }, [metrics, queryType, eventNames]);

  const rowOptions = useMemo(() => {
    return getRowOptions({
      selectedRows: pivotConfig.rows,
      metrics,
      breakdown,
      queryType,
      eventNames,
      userPropNames,
      eventPropNames
    });
  }, [pivotConfig.rows, metrics, breakdown, queryType]);

  const columnOptions = useMemo(() => {
    return getColumnOptions({
      breakdown,
      eventPropNames,
      userPropNames,
      queryType
    });
  }, [breakdown, eventPropNames, userPropNames]);

  return (
    <div className={styles.pivotTable}>
      <ControlledComponent controller={showControls}>
        <PivotTableControls
          selectedCol={pivotConfig.cols.length ? pivotConfig.cols[0] : null}
          selectedRows={pivotConfig.rows}
          selectedValue={pivotConfig.vals}
          aggregatorOptions={aggregatorOptions}
          columnOptions={columnOptions}
          rowOptions={rowOptions}
          functionOptions={getFunctionOptions()}
          aggregatorName={pivotConfig.aggregatorName}
          onRowAttributeRemove={handleRowAttributeRemove}
          onColumnAttributeRemove={handleColumnAttributeRemove}
          onValueChange={handleValueChange}
          onColumnChange={handleColumnChange}
          onRowChange={handleRowOptionSelect}
          onFunctionChange={handleFunctionChange}
          // rowOrder={pivotConfig.rowOrder}
          onSortChange={handleSortChange}
        />
      </ControlledComponent>
      <div className="w-full overflow-auto">
        <PivotTableUI
          onChange={(s) => {
            console.log(s);
            setPivotState(s);
          }}
          {...pivotState}
          {...pivotConfig}
          rowOrder="key_a_to_z" // hardcoded for now to remove sorting. Will be removed later
        />
      </div>
    </div>
  );
};

const PivotTableMemoized = memo(PivotTableComponent);

const PivotTable = (props) => {
  const {
    coreQueryState: { pivotConfig },
    updatePivotConfig
  } = useContext(CoreQueryContext);

  return (
    <PivotTableMemoized
      pivotConfig={pivotConfig}
      updatePivotConfig={updatePivotConfig}
      {...props}
    />
  );
};

export default PivotTable;

PivotTable.propTypes = {
  data: PropTypes.array,
  breakdown: PropTypes.array,
  metrics: PropTypes.array,
  showControls: PropTypes.bool,
  queryType: PropTypes.string
};

PivotTable.defaultProps = {
  data: EMPTY_ARRAY,
  breakdown: EMPTY_ARRAY,
  metrics: EMPTY_ARRAY,
  showControls: true,
  queryType: QUERY_TYPE_KPI
};
