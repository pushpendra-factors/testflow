import React from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import { settingsCategorisedMap } from 'Components/FaHeader/FaHeader';
import { WhiteListedAccounts } from 'Routes/constants';
import { useSelector } from 'react-redux';
import { PathUrls } from 'Routes/pathUrls';
import { SVG, Text } from 'Components/factorsComponents';
import SidebarMenuItem from './SidebarMenuItem';
import { checkMatchPath } from './appSidebar.helpers';
import styles from './index.module.scss';

const SettingItem = ({ item, isMainCategory }) => {
  const location = useLocation();
  const history = useHistory();
  const { pathname } = location;

  const handleItemClick = () => {
    history.push(item.url);
  };
  const isActive =
    item.url === PathUrls.SettingsIntegration
      ? pathname === item.url ||
        checkMatchPath(pathname, PathUrls.SettingsIntegrationURLID)
      : pathname === item.url;

  if (isMainCategory) {
    return (
      <div className='flex items-center gap-1'>
        <SVG name={item.icon} size={16} color='#8C8C8C' />
        <Text
          type='title'
          level={7}
          extraClass='mb-0 text-with-ellipsis w-40'
          weight='bold'
          color={`${isActive ? 'brand-color-6' : 'character-primary'}`}
        >
          {item.label}
        </Text>
      </div>
    );
  }
  return (
    <SidebarMenuItem
      text={item.label}
      isActive={isActive}
      icon={item.icon}
      onClick={handleItemClick}
      iconSize={16}
      hoverable={item.hoverable}
    />
  );
};

const SettingsSidebar = () => {
  const agentState = useSelector((state) => state.agent);
  const activeAgent = agentState?.agent_details?.email;

  return (
    <div className={`flex flex-col gap-2 px-2 ${styles['settings-sidebar']}`}>
      {settingsCategorisedMap(activeAgent).map((item) => {
        if (item?.whitelisted && !WhiteListedAccounts.includes(activeAgent)) {
          return null;
        }
        return (
          <div>
            <SettingItem item={item} isMainCategory />
            {item.items.map((subItem) => (
              <SettingItem item={subItem} />
            ))}
          </div>
        );
      })}
    </div>
  );
};

export default SettingsSidebar;
