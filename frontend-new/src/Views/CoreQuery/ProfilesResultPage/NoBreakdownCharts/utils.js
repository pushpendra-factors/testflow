import React from 'react';
import {
  generateColors,
  SortResults,
  getClickableTitleSorter,
  toLetters,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import tableStyles from '../../../../components/DataTable/index.module.scss';
import HorizontalBarChartCell from '../../EventsAnalytics/SingleEventMultipleBreakdown/HorizontalBarChartCell';
import {
  ReverseProfileMapper,
  revProfileGroupMapper,
} from '../../../../utils/constants';
import NonClickableTableHeader from '../../../../components/NonClickableTableHeader';

export const defaultSortProp = () => {
  return [
    {
      order: 'descend',
      key: 'value',
      type: 'numerical',
      subtype: null,
    },
  ];
};

export const getTableColumns = (
  currentSorter,
  handleSorting,
  groupAnalysis
) => {
  const userCol = {
    title: getClickableTitleSorter(
      revProfileGroupMapper[groupAnalysis],
      { key: 'Users', type: 'categorical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'Users',
  };
  const valCol = {
    title: getClickableTitleSorter(
      'Value',
      { key: 'value', type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    className: 'text-right',
    dataIndex: 'value',
    render: (d) => {
      return <NumFormat number={d} />;
    },
  };
  return [userCol, valCol];
};

export const getTableData = (
  data,
  queries,
  groupAnalysis,
  currentSorter,
  searchText
) => {
  try {
    const result = data.result_group.map((rg) => {
      const index = rg.rows[0][0];
      const query = queries[index];
      return {
        index,
        Users: `${toLetters(index)}. ${
          ReverseProfileMapper[query][groupAnalysis]
        }`,
        value: rg.rows[0][1],
      };
    });
    const filteredResults = result.filter((r) =>
      r.Users.toLowerCase().includes(searchText.toLowerCase())
    );
    return SortResults(filteredResults, currentSorter);
  } catch (err) {
    console.log(err);
    return [];
  }
};

export const getHorizontalBarChartColumns = (groupAnalysis) => {
  const row = {
    title: (
      <NonClickableTableHeader title={revProfileGroupMapper[groupAnalysis]} />
    ),
    dataIndex: `users`,
    className: tableStyles.horizontalBarTableHeader,
    render: (d) => {
      const obj = {
        children: <div className='h-full p-6'>{d}</div>,
      };
      return obj;
    },
  };
  return [row];
};

export const getDataInHorizontalBarChartFormat = (
  data,
  queries,
  groupAnalysis,
  cardSize,
  isDashboardWidget
) => {
  try {
    const row = {};
    row.index = 0;
    const series = [
      {
        data: [],
      },
    ];

    const colors = generateColors(10);
    const values = data.result_group.map((rg) => {
      return {
        index: rg.rows[0][0],
        value: rg.rows[0][1],
      };
    });
    const sortedData = SortResults(values, {
      key: 'value',
      order: 'descend',
    });

    const categories = sortedData.map((elem, index) => {
      const queryIndex = elem.index;
      series[0].data.push({
        y: elem.value,
        color: colors[index % 10],
      });
      return `${toLetters(queryIndex)}. ${
        ReverseProfileMapper[queries[queryIndex]][groupAnalysis]
      }`;
    });

    row['users'] = (
      <HorizontalBarChartCell
        series={series}
        categories={categories}
        cardSize={cardSize}
        isDashboardWidget={isDashboardWidget}
      />
    );

    const result = [row];
    return result;
  } catch (err) {
    console.log(err);
    return [];
  }
};
