import React from 'react';
import { Text } from 'Components/factorsComponents';
import SavedQueriesTable from 'Views/Dashboard/Drafts/SavedQueriesTable';
import NewReportButton from '../NewReportButton';

const Drafts = () => (
  <div className='flex flex-col gap-y-12'>
    <div className='flex justify-between items-center'>
      <div className='flex gap-y-2 flex-col'>
        <Text
          color='character-primary'
          level={4}
          weight='bold'
          extraClass='mb-0'
          type='title'
          id='fa-at-text--draft-title'
        >
          Drafts
        </Text>
        <Text
          color='character-secondary'
          extraClass='mb-0'
          type='title'
          id='fa-at-text--draft-desc'
        >
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

export default Drafts;
