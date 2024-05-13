import { SVG, Text } from 'Components/factorsComponents';
import { Avatar, Button } from 'antd';
import React from 'react';

interface IntegrationHeaderProps {
  title: string;
  description: string;
  iconText: string;
  handleBackClick: () => void;
  lastSyncDetail: string;
}

const IntegrationHeader = ({
  title,
  description,
  iconText,
  handleBackClick,
  lastSyncDetail
}: IntegrationHeaderProps) => (
  <div>
    <Button
      type='text'
      icon={<SVG name='GoBack' size='16' />}
      onClick={handleBackClick}
    >
      Back
    </Button>
    <div className='flex justify-between mt-2'>
      <div className='flex items-center justify-center '>
        <Avatar
          size={60}
          shape='square'
          icon={<SVG name={iconText} size={40} color='purple' />}
          style={{
            backgroundColor: '#fafafa',
            borderRadius: 10,
            border: '1px solid #fafafa',
            display: 'flex'
          }}
          className='flex items-center justify-center'
        />
      </div>
      <div className='flex flex-col justify-start items-start ml-4 w-full'>
        <div className='flex flex-row items-center justify-start w-full'>
          <div className='flex justify-between items-center w-full'>
            <Text
              type='title'
              level={4}
              weight='bold'
              extraClass='m-0'
              color='character-primary'
            >
              {title}
            </Text>
            {lastSyncDetail && (
              <div className='flex gap-2 items-center'>
                <SVG name='SyncAlt' size='20' color='#EA6262' />
                <Text
                  type='title'
                  level={7}
                  color='character-primary'
                  extraClass='m-0'
                >
                  {lastSyncDetail}
                </Text>
              </div>
            )}
          </div>
        </div>

        <Text
          type='title'
          level={7}
          extraClass='m-0 '
          color='character-secondary'
        >
          {description}
        </Text>
      </div>
    </div>
  </div>
);

export default IntegrationHeader;
