import {
  CaretDownFilled,
  CaretRightFilled,
  DeleteOutlined,
  EditOutlined,
  RightOutlined
} from '@ant-design/icons';
import React, { useContext, useState } from 'react';
import { Button, Popover, Tooltip } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import FolderItem from './FolderItem';
import styles from './index.module.scss';
import { FolderContext } from './FolderContext';
import { FolderPropType } from './type';

function Folder(props: FolderPropType) {
  const { id, name, items, isAllBoard, folders } = props;
  const contextValue = useContext(FolderContext);
  const [showItems, setShowItems] = useState<boolean>(true);
  const folderOptionsPopover = (
    <div onClick={(e) => e.stopPropagation()}>
      <div className={styles.popover_list}>
        <div
          className='flex items-left'
          onClick={() => {
            if (contextValue.setFolderModalState)
              contextValue.setFolderModalState((prev: any) => ({
                ...prev,
                visible: true,
                action: 'rename',
                unit: { ...props }
              }));
          }}
        >
          <EditOutlined /> Edit Folder
        </div>
        <div
          className='flex items-left'
          onClick={() =>
            contextValue.setFolderModalState &&
            contextValue.setFolderModalState((prev: any) => ({
              ...prev,
              visible: true,
              action: 'delete',
              unit: { ...props }
            }))
          }
        >
          <DeleteOutlined /> Delete Folder
        </div>
      </div>
    </div>
  );

  return (
    <div className={styles.folder}>
      <div
        onClick={() => setShowItems((prev) => !prev)}
        className={styles.folder_header}
      >
        <div>
          {showItems ? <CaretDownFilled /> : <CaretRightFilled />}

          <Tooltip title={name}>
            <div>{name} </div>
          </Tooltip>
        </div>{' '}
        {!isAllBoard && (
          <span className={styles.folder_actions}>
            <Popover
              content={folderOptionsPopover}
              placement='right'
              trigger='hover'
              arrowContent={<RightOutlined />}
              overlayClassName={styles.popover_list_container}
            >
              {' '}
              <Button
                icon={<SVG size={16} color='#8C8C8C' name='more' />}
                className={styles.folder_actions_button}
              />
            </Popover>
          </span>
        )}
      </div>

      <div style={{ display: showItems ? 'inherit' : 'none' }}>
        {items.length > 0 ? (
          items.map((eachFolderItem) => (
            <FolderItem
              key={eachFolderItem.id}
              data={eachFolderItem}
              id={eachFolderItem.id}
              folders={folders}
              folder_id={id}
            />
          ))
        ) : (
          <Text
            level={8}
            color='character-secondary'
            type='title'
            extraClass='mb-0 text-center'
          >
            No {contextValue.unit}s in this Folder
          </Text>
        )}
      </div>
    </div>
  );
}
export default Folder;
