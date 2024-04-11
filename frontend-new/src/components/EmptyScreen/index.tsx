import React from 'react';
import EmptyScreenDefaultIllustration from './../../assets/images/illustrations/EmptyScreenDefaultIllustration.png';
import styles from './index.module.scss';
import { Text } from 'Components/factorsComponents';
import { Button, Empty } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import SVG from 'Components/factorsComponents/SVG';
type EmptyScreenProps = {
  image: JSX.Element | null;
  title: JSX.Element | string | null;
  learnMore?: null | string; // If we have any URL
  topTitle?: JSX.Element | null;
  showTop?: boolean;
  imageStyle?: React.CSSProperties;
  ActionButton?: {
    icon?: JSX.Element;
    text?: string | JSX.Element;
    onClick?: () => void | null;
  } | null;
  upgradeScreen?: boolean;
};
export default function ({
  image,
  imageStyle = { width: 216, height: 216 },
  title,
  topTitle,
  ActionButton,
  showTop,
  learnMore,
  upgradeScreen = false
}: EmptyScreenProps) {
  return (
    <div className={styles.parent}>
      {showTop && (
        <div className={styles['top-action']}>
          <div> {topTitle} </div>{' '}
          {ActionButton && (
            <Button
              type='primary'
              icon={<PlusOutlined color='white' />}
              onClick={ActionButton.onClick}
            >
              {' '}
              {ActionButton.text || 'Add New'}{' '}
            </Button>
          )}
        </div>
      )}
      <Empty
        imageStyle={{
          margin: '0 auto',
          ...imageStyle
        }}
        image={image || EmptyScreenDefaultIllustration}
        style={{
          width: upgradeScreen ? '100%' : '60%',
          margin: '0 auto',
          padding: 5,
          textAlign: 'center'
        }}
        description={
          <Text
            type={'title'}
            level={6}
            color={'grey-2'}
            extraClass={'m-0 mt-2'}
          >
            {title}
          </Text>
        }
      >
        <div className='flex justify-center gap-2'>
          {learnMore && (
            <a href={learnMore} target='_blank' tabIndex={0}>
              <Button
                className='dropdown-btn'
                type='text'
                icon={<SVG name={'NewTab'} />}
              >
                Learn More
              </Button>
            </a>
          )}
          {!showTop && ActionButton && (
            <Button
              type='primary'
              icon={!upgradeScreen && <PlusOutlined />}
              onClick={ActionButton.onClick}
            >
              {ActionButton.text || 'Add New'}{' '}
            </Button>
          )}
        </div>
      </Empty>
    </div>
  );
}
