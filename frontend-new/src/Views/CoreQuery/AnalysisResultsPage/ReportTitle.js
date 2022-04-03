import React, { useCallback } from 'react';
import styles from './index.module.scss';
import moment from 'moment';
import { Button, Tooltip } from 'antd';
import { SVG, Text } from '../../../components/factorsComponents';
import { useSelector } from 'react-redux';
import {
  REPORT_SECTION,
  DASHBOARD_MODAL,
  QUERY_TYPE_WEB,
} from '../../../utils/constants';

function ReportTitle({
  title,
  setDrawerVisible,
  queryDetail,
  section,
  onReportClose,
  queryType,
  apiCallStatus,
}) {
  // const [apiCallStatusMsgVisible, setApiCallStatusMsgVisible] = useState(true);
  const handleClick = useCallback(() => {
    if (section === REPORT_SECTION) {
      setDrawerVisible(true);
    }
    if (section === DASHBOARD_MODAL) {
      setDrawerVisible();
    }
  }, [section, setDrawerVisible]);

  const { eventNames } = useSelector((state) => state.coreQuery);

  const displayQueryName = (q) => {
    const names = q.split(',');
    const sanitisedNames = names.map((nam) => {
      return eventNames[nam.trim()] ? eventNames[nam.trim()] : nam;
    });
    return sanitisedNames.join(', ');
  };

  // const handleApiStatusVisibilityChange = useCallback(() => {
  //   setApiCallStatusMsgVisible((currState) => {
  //     return !currState;
  //   });
  // }, []);

  // useEffect(() => {
  //   setApiCallStatusMsgVisible(true);
  // }, [apiCallStatus]);

  return (
    <div className='pb-2 border-bottom--thin-2'>
      <div className='flex justify-between items-center'>
        <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0 mt-6'}>
          {title || `Untitled Analysis ${moment().format('DD/MM/YYYY')}`}
        </Text>
        {section === DASHBOARD_MODAL ? (
          <Button
            type={'text'}
            onClick={onReportClose.bind(this, false)}
            icon={<SVG name='Remove' />}
          />
        ) : null}
      </div>
      <div className='flex items-center justify-between'>
        <div
          className={'fa-title--editable flex items-center cursor-pointer '}
          onClick={queryType !== QUERY_TYPE_WEB ? handleClick : null}
        >
          <Text type={'title'} level={6} color={'grey'} extraClass={'m-0 mr-2'}>
            {displayQueryName(queryDetail)}
          </Text>
          <SVG name='edit' color={'grey'} />
        </div>
        {apiCallStatus && apiCallStatus.required && apiCallStatus.message ? (
          <Tooltip
            mouseEnterDelay={0.2}
            title={apiCallStatus.message}
            placement='topLeft'
            overlayClassName={`${styles.apiCallStatusMsgTooltip}`}
          >
            <div className='cursor-pointer'>
              <SVG color='#dea069' name={'warning'} />
            </div>
          </Tooltip>
        ) : null}
      </div>
    </div>
  );
}

export default ReportTitle;
