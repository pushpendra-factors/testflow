import { SVG, Text } from 'Components/factorsComponents';
import React from 'react';
import { Link } from 'react-router-dom';

interface IntegrationInstructionProps {
  title?: string;
  kbLink?: string;
  description?: string;
}

const IntegrationInstruction = ({
  title = 'Integration Details',
  kbLink,
  description
}: IntegrationInstructionProps) => (
  <div>
    <div className='flex items-center justify-between'>
      <Text
        type='title'
        level={4}
        weight='bold'
        extraClass='m-0'
        color='character-primary'
      >
        {title}
      </Text>
      {kbLink && (
        <Link
          className='inline-block ml-1'
          target='_blank'
          to={{
            pathname: kbLink
          }}
        >
          <div className='flex items-center gap-2'>
            <Text type='paragraph' mini weight='bold' color='brand-color-6'>
              View Documentation
            </Text>
            <SVG name='ArrowUpRightSquare' size={14} color='#1890ff' />
          </div>
        </Link>
      )}
    </div>
    <Text
      type='title'
      level={7}
      color='character-secondary'
      extraClass='m-0 mt-1'
    >
      {description}
    </Text>
  </div>
);

export default IntegrationInstruction;
