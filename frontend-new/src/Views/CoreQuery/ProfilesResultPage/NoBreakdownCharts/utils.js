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
import { ProfileUsersMapper } from '../../../../utils/constants';

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

export const getTableColumns = (currentSorter, handleSorting) => {
  const userCol = {
    title: getClickableTitleSorter(
      'Users',
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
      handleSorting
    ),
    dataIndex: 'value',
    render: (d) => {
      return <NumFormat number={d} />;
    },
  };
  return [userCol, valCol];
};

export const getTableData = (data, queries, currentSorter, searchText) => {
  try {
    const result = data.result_group.map((rg) => {
      const index = rg.rows[0][0];
      const query = queries[index];
      return {
        index,
        Users: `${toLetters(index)}. ${ProfileUsersMapper[query]}`,
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

export const getHorizontalBarChartColumns = () => {
  const row = {
    title: 'Users',
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
        ProfileUsersMapper[queries[queryIndex]]
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
