import { CoreQueryContext } from 'Context/CoreQueryContext';
import React, { memo, useCallback, useContext, useMemo } from 'react';
import OptionsPopover from '../AttributionsResult/OptionsPopover';
import GroupedChart from './GroupedChart';
import UngroupedChart from './UngroupedChart';

const getTableConfig = (funnelTableConfig) => {
  return funnelTableConfig.reduce((prev, curr) => {
    return {
      ...prev,
      [curr.key]: curr.enabled
    };
  }, {});
};

function ResultantChartComponent({
  queries,
  resultState,
  breakdown,
  isWidgetModal,
  arrayMapper,
  section,
  durationObj,
  chartType,
  renderedCompRef,
  tableConfig,
  tableConfigPopoverContent
}) {
  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        queries={queries}
        isWidgetModal={isWidgetModal}
        arrayMapper={arrayMapper}
        section={section}
        durationObj={durationObj}
        ref={renderedCompRef}
        tableConfig={tableConfig}
        tableConfigPopoverContent={tableConfigPopoverContent}
        chartType={chartType}
      />
    );
  } else {
    return (
      <GroupedChart
        queries={queries}
        resultState={resultState}
        breakdown={breakdown}
        isWidgetModal={isWidgetModal}
        arrayMapper={arrayMapper}
        section={section}
        renderedCompRef={renderedCompRef}
        chartType={chartType}
        tableConfig={tableConfig}
        tableConfigPopoverContent={tableConfigPopoverContent}
      />
    );
  }
}

const ResultantChartMemoized = memo(ResultantChartComponent);

function ResultantChart(props) {
  const {
    coreQueryState: { funnelTableConfig },
    updateFunnelTableConfig
  } = useContext(CoreQueryContext);
  
  const handleTableConfigChange = useCallback(
    (option) => {
      const updatedConfig = funnelTableConfig.map((config) => {
        if (config.key === option.key) {
          return {
            ...config,
            enabled: !config.enabled
          };
        }
        return config;
      });
      updateFunnelTableConfig(updatedConfig);
    },
    [funnelTableConfig, updateFunnelTableConfig]
  );

  const tableConfigPopoverContent = useMemo(() => {
    return (
      <OptionsPopover
        options={funnelTableConfig}
        onChange={handleTableConfigChange}
      />
    );
  }, [funnelTableConfig, handleTableConfigChange]);

  const tableConfig = useMemo(() => {
    return getTableConfig(funnelTableConfig);
  }, [funnelTableConfig]);

  return (
    <ResultantChartMemoized
      tableConfigPopoverContent={tableConfigPopoverContent}
      tableConfig={tableConfig}
      {...props}
    />
  );
}

export default ResultantChart;
