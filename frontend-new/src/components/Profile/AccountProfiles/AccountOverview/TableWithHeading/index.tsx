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
}) => {
  return (
    <div className='grid-table'>
      <div className='heading'>
        <Text
          type='title'
          level={7}
          extraClass='m-0 whitespace-no-wrap'
          weight='bold'
          color='grey-2'
        >
          {heading}
        </Text>
      </div>
      <div className='data-table'>
        <Table
          className='fa-overview--top-table'
          dataSource={data}
          columns={columns}
          pagination={false}
          scroll={{ y: yScroll }}
        />
      </div>
    </div>
  );
};
export default TableWithHeading;
