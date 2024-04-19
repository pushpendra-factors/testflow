import { Table } from 'antd';
import { Text } from 'Components/factorsComponents';
import React from 'react';

interface TableWithHeadingProps {
  heading: string;
  data: any;
  columns: any;
  yScroll: number;
}

const TableWithHeading: React.FC<TableWithHeadingProps> = ({
  heading,
  data,
  columns,
  yScroll = 200
}) => (
  <div className='top-table-container'>
    <div className='heading'>
      <Text
        type='title'
        level={7}
        extraClass='m-0 whitespace-nowrap'
        weight='bold'
        color='grey-2'
      >
        {heading}
      </Text>
    </div>
    <Table
      className='content'
      dataSource={data}
      columns={columns}
      pagination={false}
      scroll={{ y: yScroll }}
    />
  </div>
);
export default TableWithHeading;
