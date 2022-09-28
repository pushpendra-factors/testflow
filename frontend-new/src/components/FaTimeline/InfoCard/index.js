import React from 'react';
import { Text } from 'Components/factorsComponents';
import { Popover } from 'antd';
import {
  formatDurationIntoString,
  PropTextFormat,
} from '../../../utils/dataFormatter';
import MomentTz from '../../MomentTz';
import { TimelineHoverPropDisplayNames } from '../../Profile/utils';

function InfoCard({ title, event_name, properties = {}, trigger, children }) {
  const popoverPropValueFormat = (key, value) => {
    if (
      key.includes('timestamp') ||
      key.includes('starttime') ||
      key.includes('endtime')
    ) {
      return MomentTz(value * 1000).format('DD MMMM YYYY, hh:mm A');
    } else if (key.includes('_time')) {
      formatDurationIntoString(value);
    } else if (key.includes('durationmilliseconds')) {
      formatDurationIntoString(parseInt(value / 1000));
    } else return value;
  };
  const popoverContent = () => {
    return (
      <div className='fa-popupcard'>
        <Text
          extraClass='m-0 mb-3'
          type={'title'}
          level={6}
          weight={'bold'}
          color={'grey-2'}
        >
          {title}
        </Text>
        {Object.entries(properties).map(([key, value]) => {
          if (key === '$is_page_view' && value === true)
            return (
              <div className='flex justify-between py-2'>
                <Text
                  mini
                  type={'title'}
                  color={'grey'}
                  extraClass={'whitespace-no-wrap mr-2'}
                >
                  Page URL
                </Text>

                <Text
                  mini
                  type={'title'}
                  color={'grey-2'}
                  weight={'medium'}
                  extraClass={`break-all text-right`}
                  truncate={true}
                  charLimit={40}
                >
                  {
                    'https://studio.memsql.com/cluster/eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJuYW0iOiJmYWN0b3JzLXByb2R1Y3Rpb24iLCJ1c3IiOiJhZG1pbiIsImVuZCI6InN2Yy1iZWUwMzRiOC0yYWQyLTRmNzMtODE0NC0wYzljY2IzZWI3OWItZGRsLmdjcC1vcmVnb24tMS5zdmMuc2luZ2xlc3RvcmUuY29tIiwiZW52IjoicCIsImNpZCI6ImJlZTAzNGI4LTJhZDItNGY3My04MTQ0LTBjOWNjYjNlYjc5YiJ9.ury2WAUaJg-YW2JLKYsVNepn0oK8MhVrFyInNAR-cwwdeHe_4KdEOZ8UlIym8CYnRHS3TdOAsZ8_YdrtUc--dA/editor'
                  }
                </Text>
              </div>
            );
          else
            return (
              <div className='flex justify-between py-2'>
                <Text
                  mini
                  type={'title'}
                  color={'grey'}
                  extraClass={`${
                    key.length > 20 ? 'break-words' : 'whitespace-no-wrap'
                  } max-w-xs mr-2`}
                >
                  {TimelineHoverPropDisplayNames[key] || PropTextFormat(key)}
                </Text>
                <Text
                  mini
                  type={'title'}
                  color={'grey-2'}
                  weight={'medium'}
                  extraClass={`${
                    value?.length > 30 ? 'break-words' : 'whitespace-no-wrap'
                  }  text-right`}
                  truncate={true}
                  charLimit={40}
                >
                  {popoverPropValueFormat(key, value) || '-'}
                </Text>
              </div>
            );
        })}
      </div>
    );
  };
  return (
    <Popover
      content={popoverContent}
      overlayClassName='fa-infocard--wrapper'
      placement='rightBottom'
      trigger={trigger}
    >
      {children}
    </Popover>
  );
}
export default InfoCard;
