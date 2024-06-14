import React from 'react';
import { Text } from 'Components/factorsComponents';
import { Link } from 'react-router-dom';
import { Divider } from 'antd';

interface CommonSettingsHeaderProps {
  title: string;
  description?: string;
  learnMoreLink?: string;
  actionsNode?: React.ReactNode;
  hasNoBottomPadding?: boolean;
}

const CommonSettingsHeader = ({
  title,
  description,
  learnMoreLink,
  actionsNode,
  hasNoBottomPadding = false
}: CommonSettingsHeaderProps) => (
  <div>
    <div className='flex items-center justify-between gap-4'>
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
              extraClass='inline-block  m-0 }'
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
    <Divider
      style={{
        margin: `${hasNoBottomPadding ? '16px 0px 0px 0px' : '16px 0px'}`
      }}
    />
  </div>
);

export default CommonSettingsHeader;
