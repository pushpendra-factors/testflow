import React, { useState } from 'react';
import DataTableFilters from '../DataTableFilters';
import { TEST_FILTER_OPTIONS } from '../dataTableFilters.constants';

export default {
  title: 'Components/DataTableFilters',
  component: DataTableFilters
};

export const NoFiltersSelected = () => {
  const [selectedFilters, setSelectedFilters] = useState({});

  return (
    <div style={{ width: '500px' }}>
      <DataTableFilters
        filters={TEST_FILTER_OPTIONS}
        appliedFilters={selectedFilters}
        setAppliedFilters={setSelectedFilters}
      />
    </div>
  );
};

export const WithFiltersSelected = () => {
  const [selectedFilters, setSelectedFilters] = useState({
    categoryCombinationOperator: 'AND',
    categories: [
      {
        values: [],
        equalityOperator: 'equal',
        field: 'Campaign_Name',
        key: 1657998866521
      }
    ]
  });

  return (
    <div style={{ width: '500px' }}>
      <DataTableFilters
        filters={TEST_FILTER_OPTIONS}
        appliedFilters={selectedFilters}
        setAppliedFilters={setSelectedFilters}
      />
    </div>
  );
};
