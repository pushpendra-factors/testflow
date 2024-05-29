import React from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import { settingsCategorisedMap } from 'Components/FaHeader/FaHeader';
import { WhiteListedAccounts } from 'Routes/constants';
import { useSelector } from 'react-redux';
import { PathUrls } from 'Routes/pathUrls';
import SidebarMenuItem from './SidebarMenuItem';
import { checkMatchPath } from './appSidebar.helpers';

const SettingItem = ({ item }) => {
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
    <div className='flex flex-col gap-y-1 px-2'>
      {settingsCategorisedMap(activeAgent).map((item) => {
        if (item?.whitelisted && !WhiteListedAccounts.includes(activeAgent)) {
          return null;
        }
        return (
          <>
            <SettingItem item={item} />
            <div className='border-bottom--thin-2' />
            {item.items.map((subItem) => (
              <SettingItem item={subItem} />
            ))}
          </>
        );
      })}
    </div>
  );
};

export default SettingsSidebar;
