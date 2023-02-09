import React from 'react';
import { Radio } from 'antd';

const QuickFilter = ({
  filters,
  onFilterChange,
  selectedFilter
}: QuickFilterProps) => {
  return (
    <div>
      <Radio.Group
        value={selectedFilter}
        onChange={(e) => onFilterChange(e.target.value)}
      >
        {filters.map((filter) => (
          <Radio.Button value={filter.id} key={filter.id}>
            {filter.label}
          </Radio.Button>
        ))}
      </Radio.Group>
    </div>
  );
};

type QuickFilterProps = {
  filters: { id: string; label: string }[];
  onFilterChange: (id: string) => void;
  selectedFilter?: string;
};

export default QuickFilter;
