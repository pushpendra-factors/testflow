import { Tooltip, Image } from 'antd';
import { Text } from 'Components/factorsComponents';
import { getHost } from 'Components/Profile/utils';
import React from 'react';
import { formatDuration } from 'Utils/dataFormatter';
import {
  COMPANY_KEY,
  DOMAIN_KEY,
  INDUSTRY_KEY,
  PAGE_COUNT_KEY,
  SESSION_SPENT_TIME
} from '../../../const';
import { StringObject } from '../../../types';
import fallbackImage from '../../../../../assets/icons/fallbackImage.svg';

const TableCell = ({ text, record, header }: TableCellProps) => {
  let title = text;
  const domain = record?.[DOMAIN_KEY];
  const showCursor = header === COMPANY_KEY && domain;
  if (header === SESSION_SPENT_TIME) {
    if (isNaN(Number(text))) {
      title = 'NA';
    } else if (Number(text) < 1800) {
      title = formatDuration(text);
    } else {
      title = '> 30mins';
    }
  } else if (header === PAGE_COUNT_KEY) {
    title = `${text} ${Number(text) > 1 ? 'Pages' : 'Page'}`;
  }

  const openUrlInNewTab = (_domain: string) => {
    try {
      const url = new URL(_domain);
      window.open(url, '_blank');
    } catch (error) {
      const url = 'https://' + _domain;
      window.open(url, '_blank');
    }
  };

  return (
    <>
      <div
        className={`flex gap-2 items-center ${
          showCursor ? 'pl-1 cursor-pointer' : ''
        }`}
        onClick={showCursor ? () => openUrlInNewTab(domain) : undefined}
      >
        {domain && header === COMPANY_KEY && (
          <div className='w-6 h-6 flex justify-center items-center'>
            <Image
              width={24}
              height={24}
              className='w-100 h-100 rounded '
              src={`https://logo.uplead.com/${getHost(domain)}`}
              fallback={fallbackImage}
              preview={false}
            />
          </div>
        )}
        <div className='flex-1 whitespace-nowrap overflow-hidden text-ellipsis'>
          {[COMPANY_KEY, INDUSTRY_KEY].includes(header) ? (
            <Tooltip title={title} color='#0B1E39'>
              <Text type='title' level={7} extraClass='m-0' ellipsis truncate>
                {title}
              </Text>
            </Tooltip>
          ) : (
            <Text type='title' level={7} extraClass='m-0' ellipsis truncate>
              {title}
            </Text>
          )}
        </div>
      </div>
    </>
  );
};

type TableCellProps = {
  text: string;
  record: StringObject;
  header: string;
};

export default TableCell;
