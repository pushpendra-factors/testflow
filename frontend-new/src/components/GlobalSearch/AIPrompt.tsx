import { Button, Tag, notification } from 'antd';
import React, { useState } from 'react';
import { debounce } from 'lodash';
import { useDispatch, useSelector } from 'react-redux';
import logger from 'Utils/logger';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import { TOGGLE_GLOBAL_SEARCH } from 'Reducers/types';
import KpiSchema from 'Schema/kpi_schema';
import { TextPromptAPIResponse, getQueryFromTextPrompt } from './service';

const AIPrompt = ({ searchkey }: AIPromptProps) => {
  const [loading, setLoading] = useState(false);
  const { active_project } = useSelector((state) => state.global);
  const { config } = useSelector((state) => state.kpi);
  const dispatch = useDispatch();
  const history = useHistory();
  const handleclick = debounce(async () => {
    try {
      setLoading(true);
      if (!searchkey) {
        logger.error('No searchKey passed');
      }

      const res = (await getQueryFromTextPrompt(
        active_project?.id,
        searchkey,
        config
      )) as TextPromptAPIResponse;

      if (res?.data?.payload) {
        const query = { query: res.data.payload };
        // validating query with KPI schema
        const isQueryValid = KpiSchema.isValidSync(res.data.payload);
        if (isQueryValid) {
          history.push({
            pathname: PathUrls.Analyse2,
            state: { query, navigatedFromAIChartPrompt: true }
          });
          dispatch({ type: TOGGLE_GLOBAL_SEARCH });
          setLoading(false);
          return;
        }
        logger.error('Query payload received is not valid', query);
      }
      notification.error({
        message: 'Failed!',
        description: "Couldn't create query using the text prompt",
        duration: 3
      });

      setLoading(false);
    } catch (error) {
      logger.error('error in ai prompt', error);
      setLoading(false);
      notification.error({
        message: 'Failed!',
        description: "Couldn't create query using the text prompt",
        duration: 3
      });
    }
  }, 300);
  return (
    <div
      tabIndex={0}
      className='flex gap-2 items-center px-4 py-1'
      onClick={handleclick}
      onKeyUp={(e) => {
        if (e.key === 'Enter') handleclick();
      }}
    >
      <Button type='text' size='large' loading={loading}>
        Enter a meaningful text prompt to search with AI
      </Button>
      <Tag color='processing'>Experimental</Tag>
    </div>
  );
};

interface AIPromptProps {
  searchkey: string;
}
export default AIPrompt;
