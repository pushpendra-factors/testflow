import React from 'react';
import moment from 'moment';
import { calcChangePerc } from './utils';
import DataTable from '../../../components/DataTable';
import { DASHBOARD_WIDGET_SECTION } from '../../../utils/constants';

function AttributionTable({
  comparison_data,
  isWidgetModal,
  setVisibleIndices,
  visibleIndices,
  maxAllowedVisibleProperties,
  reportTitle = 'Attributions',
  durationObj,
  cmprDuration,
  attributionMetrics,
  section = null,
  columns,
  tableData,
  setSearchText,
  searchText,
  metricsOptionsPopover
}) {
  const getCSVData = () => {
    const dt = tableData;
    const enabledAttributionMetricKeys = attributionMetrics
      .filter((d) => d.enabled)
      .map((d) => d.title);
    const mappedData = dt.map(({ index, category, ...rest }) => {
      if (!comparison_data) {
        return rest;
      }
      const fromDate = moment(durationObj.from).format('MMM DD');
      const toDate = moment(durationObj.to).format('MMM DD');
      const compareFromDate = moment(cmprDuration.from).format('MMM DD');
      const compareToDate = moment(cmprDuration.to).format('MMM DD');
      const result = {};
      Object.keys(rest).forEach((key) => {
        if (
          !enabledAttributionMetricKeys.includes(key) &&
          key !== 'Conversion' &&
          key !== 'Cost per Conversion' &&
          key !== 'Conversion Rate' &&
          !key.includes('Linked Event')
        ) {
          result[key] = rest[key];
        } else {
          const changePercent = calcChangePerc(
            rest[key].value,
            rest[key].compare_value
          );
          result[`${key} (${fromDate} - ${toDate})`] = rest[key].value;
          result[`${key} (${compareFromDate} - ${compareToDate})`] =
            rest[key].compare_value;
          result[`${key} change`] = isNaN(changePercent)
            ? '0%'
            : changePercent === 'Infinity' || changePercent === '-Infinity'
            ? 'Infinity'
            : changePercent + '%';
        }
      });
      return result;
    });

    return {
      fileName: `${reportTitle}.csv`,
      data: mappedData,
    };
  };

  const onSelectionChange = (selectedIncices) => {
    if (selectedIncices.length > maxAllowedVisibleProperties) {
      return false;
    }
    if (!selectedIncices.length) {
      return false;
    }
    selectedIncices.sort();
    setVisibleIndices(selectedIncices);
  };

  const rowSelection = {
    selectedRowKeys: visibleIndices,
    onChange: onSelectionChange,
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      rowSelection={rowSelection}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
      ignoreDocumentClick={section === DASHBOARD_WIDGET_SECTION}
      controlsPopover={metricsOptionsPopover}
    />
  );
}

export default AttributionTable;
