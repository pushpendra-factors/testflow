import React from 'react';
import { Text } from 'Components/factorsComponents';
import NewReportButton from '../NewReportButton';
import SavedQueriesTable from 'Views/Dashboard/Drafts/SavedQueriesTable';

const Drafts = () => {
  return (
    <div className='flex flex-col row-gap-12'>
      <div className='flex justify-between items-center'>
        <div className='flex row-gap-2 flex-col'>
          <Text
            color='character-primary'
            level={4}
            weight='bold'
            extraClass='mb-0'
            type='title'
          >
            Drafts
          </Text>
          <Text color='character-secondary' extraClass='mb-0' type='title'>
            This is a list of all reports across dashboards with in your current
            project.
          </Text>
        </div>
        <div className='flex items-center'>
          <NewReportButton showSavedReport={false} />
        </div>
      </div>
      <SavedQueriesTable />
    </div>
  );
};

export default Drafts;
