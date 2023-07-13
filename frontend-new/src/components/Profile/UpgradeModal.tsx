import { Button } from 'antd';
import Modal from 'antd/lib/modal/Modal';
import React from 'react';
import style from './index.module.scss';
import cx from 'classnames';
import { ProfileUpgradeModalType } from 'Context/types';
import { SVG, Text } from 'Components/factorsComponents';
import AccountTableImage from '../../assets/images/account_table.png';
import TimelineTableImage from '../../assets/images/timeline_table.png';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';

const UpgradeModal = ({ visible, onCancel, variant }: UpgradeModalProps) => {
  const history = useHistory();
  return (
    <Modal
      visible={visible}
      onCancel={onCancel}
      footer={null}
      className={cx(style['upgrade-modal'])}
      width={800}
      closable={false}
      centered
    >
      <div className={cx(style['background-images'])}>
        <div
          className={cx({
            [style['background-image-top-left-account']]: variant === 'account',
            [style['background-image-top-left-timeline']]:
              variant === 'timeline'
          })}
        ></div>
        <div
          className={cx({
            [style['background-image-bottom-right-account']]:
              variant === 'account',
            [style['background-image-bottom-right-timeline']]:
              variant === 'timeline'
          })}
        ></div>
      </div>
      <Button
        type='text'
        shape='circle'
        icon={<SVG name='remove' />}
        className={cx(style['cancel-button'])}
        onClick={onCancel}
      />
      <div>
        <Text
          type={'title'}
          level={2}
          weight={'bold'}
          extraClass={'m-0 text-center'}
        >
          Upgrade to explore our other features{' '}
        </Text>
        <Text
          type={'paragraph'}
          extraClass={'m-0 mt-2 text-center'}
          color='grey'
        >
          Upgrade to explore our other features Upgrade to explore our other
          features Upgrade to explore our other features Upgrade to explore our
          other features Upgrade to explore our other features
        </Text>
        <img
          src={variant === 'account' ? AccountTableImage : TimelineTableImage}
          alt='table'
        />
        <div className='flex items-center justify-center mt-4'>
          <Button
            type='primary'
            className='text-center mt-4'
            onClick={() => history.push(PathUrls.SettingsPricing)}
          >
            Upgrade
          </Button>
        </div>
      </div>
    </Modal>
  );
};

type UpgradeModalProps = {
  visible: boolean;
  onCancel: () => void;
  variant: ProfileUpgradeModalType;
};

export default UpgradeModal;
