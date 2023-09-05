import { Button, Menu, Dropdown } from 'antd';
import React, { useMemo } from 'react';
import styles from './index.module.scss';
import { Text, SVG } from 'Components/factorsComponents';

interface Props {
  onDelete: (flag: boolean) => void;
  onRename: (flag: boolean) => void;
}

const MoreActionsDropdown = ({ onRename, onDelete }: Props) => {
  const moreActionsMenu = useMemo(() => {
    return (
      <Menu className={styles['dropdown-menu']}>
        <Menu.Item
          key='rename'
          className={styles['dropdown-menu-item']}
          onClick={() => onRename(true)}
        >
          <div className='flex items-center col-gap-1'>
            <SVG color='#00000073' name='edit_query' size={16} />
            <Text color='character-primary' type='title' extraClass='mb-0'>
              Rename Segment
            </Text>
          </div>
        </Menu.Item>
        <Menu.Item
          className={styles['dropdown-menu-item']}
          onClick={() => onDelete(true)}
          key='delete'
        >
          <div className='flex items-center col-gap-1'>
            <SVG color='#FF4D4F' name='trash' size={16} />
            <Text color='red' type='title' extraClass='mb-0'>
              Delete Segment
            </Text>
          </div>
        </Menu.Item>
      </Menu>
    );
  }, [onDelete, onRename]);

  return (
    <Dropdown placement='bottomRight' overlay={moreActionsMenu}>
      <Button className={styles['filter-button']}>
        <Text
          weight='medium'
          type='title'
          extraClass='mb-0'
          color='character-primary'
        >
          More Actions
        </Text>
      </Button>
    </Dropdown>
  );
};

export default MoreActionsDropdown;
