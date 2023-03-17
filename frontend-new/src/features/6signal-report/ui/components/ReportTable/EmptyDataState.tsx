import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import React from 'react';

const EmptyDataState = ({ title, subtitle, icon, action }: Props) => {
  return (
    <div className='w-100 flex flex-col  align-middle gap-4 py-10'>
      <div className='flex justify-center align-middle'>
        <SVG
          name={icon.name}
          size={icon.size}
          color={icon.color}
          height={icon.height}
          width={icon.width}
        />
      </div>
      <div>
        <Text type={'title'} level={6} weight={'bold'} color='grey-2'>
          {title}
        </Text>
        <Text type={'title'} level={7} color='grey-2' extraClass='mt-2'>
          {subtitle}
        </Text>
      </div>
      {action && action?.name && (
        <div>
          <Button size='large' type='primary' onClick={action.handleClick}>
            {action.name}
          </Button>
        </div>
      )}
    </div>
  );
};

type Props = {
  title: string;
  subtitle: string;
  icon: {
    name: string;
    size?: number;
    color: string;
    height?: number;
    width?: number;
  };
  action?: {
    name: string;
    handleClick: () => void;
  };
};

export default EmptyDataState;
