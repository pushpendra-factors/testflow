import React from 'react';
import { Text } from 'Components/factorsComponents';
import { Link } from 'react-router-dom';
import { Divider } from 'antd';

interface CommonSettingsHeaderProps {
  title: string;
  description?: string;
  learnMoreLink?: string;
  actionsNode?: React.ReactNode;
}

const CommonSettingsHeader = ({
  title,
  description,
  learnMoreLink,
  actionsNode
}: CommonSettingsHeaderProps) => (
  <div>
    <div className='flex items-center justify-between'>
      <div>
        <Text
          type='title'
          level={4}
          color='character-primary'
          weight='bold'
          extraClass='m-0'
        >
          {title}
        </Text>
        {description && (
          <div className='flex items-baseline flex-wrap'>
            <Text
              type='paragraph'
              mini
              color='character-primary'
              extraClass={`inline-block  m-0 ${!actionsNode ? 'w-3/4' : ''}`}
            >
              {description}
              {learnMoreLink && (
                <Link
                  className='inline-block ml-1'
                  target='_blank'
                  to={{
                    pathname: learnMoreLink
                  }}
                >
                  <Text
                    type='paragraph'
                    mini
                    weight='bold'
                    color='brand-color-6'
                  >
                    {'  '} Learn more
                  </Text>
                </Link>
              )}
            </Text>
          </div>
        )}
      </div>

      {actionsNode && <>{actionsNode}</>}
    </div>
    <Divider />
  </div>
);

export default CommonSettingsHeader;
