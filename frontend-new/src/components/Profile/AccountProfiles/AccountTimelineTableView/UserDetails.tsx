import React, { useEffect, useMemo, useState } from 'react';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { SVG, Text } from 'Components/factorsComponents';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { ALPHANUMSTR, iconColors } from 'Components/Profile/constants';
import { propValueFormat } from 'Components/Profile/utils';
import { PropTextFormat } from 'Utils/dataFormatter';
import { UserDetailsProps } from 'Components/Profile/types';
import { Avatar, Button } from 'antd';
import { bindActionCreators } from 'redux';
import { getConfiguredUserProperties } from 'Reducers/timelines/middleware';

function UserDetails({
  user,
  userPropsType,
  onUpdate,
  getConfiguredUserProperties
}: ComponentProps) {
  const [sortableItems, setSortableItems] = useState<string[]>([]);

  const { userPropNames } = useSelector((state: any) => state.coreQuery);
  const { active_project: activeProject, currentProjectSettings } = useSelector(
    (state: any) => state.global
  );
  const { userConfigProperties } = useSelector((state: any) => state.timelines);

  const userProperties = useMemo(() => {
    if (!user) return {};
    return userConfigProperties[user.id] || {};
  }, [user, userConfigProperties]);

  useEffect(() => {
    if (!user) return;
    if (!userConfigProperties[user?.id])
      getConfiguredUserProperties(activeProject.id, user.id, user.isAnonymous);
  }, [activeProject, user, userConfigProperties]);

  const renderUsername = (userName: string, isAnon: boolean) => {
    if (isAnon) {
      return 'Anonymous User';
    }
    return userName;
  };

  useEffect(() => {
    if (
      user &&
      currentProjectSettings?.timelines_config?.user_config?.table_props
    ) {
      setSortableItems(
        currentProjectSettings.timelines_config.user_config?.table_props
      );
    }
  }, [user, currentProjectSettings]);

  const handleDelete = (index: number) => {
    const updatedItems = [...sortableItems];
    updatedItems.splice(index, 1);
    setSortableItems(updatedItems);
    onUpdate(updatedItems);
  };

  if (!user) return null;

  return (
    <div className='py-4'>
      <div className='top-section mb-4'>
        <div className='flex items-center w-full'>
          {user.isAnonymous ? (
            <SVG
              name={`TrackedUser${user.name.match(/\d/g)?.[0] || 0}`}
              size={28}
            />
          ) : (
            <Avatar
              size={28}
              className='avatar'
              style={{
                backgroundColor: `${
                  iconColors[
                    ALPHANUMSTR.indexOf(user.name.charAt(0).toUpperCase()) % 8
                  ]
                }`,
                fontSize: '20px'
              }}
            >
              {user.name.charAt(0).toUpperCase()}
            </Avatar>
          )}
          <TextWithOverflowTooltip
            text={renderUsername(user.name, user.isAnonymous)}
            extraClass='heading ml-2'
          />
        </div>
      </div>
      <div>
        {sortableItems.map((property, index) => {
          const propType = userPropsType[property] || 'categorical';
          return (
            <div className='leftpane-prop justify-between'>
              <div className='flex items-center justify-start'>
                <div className='flex flex-col items-start truncate ml-6'>
                  <Text
                    type='title'
                    level={8}
                    color='grey'
                    truncate
                    charLimit={40}
                    extraClass='m-0'
                  >
                    {userPropNames[property] || PropTextFormat(property)}
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    truncate
                    charLimit={36}
                    extraClass='m-0'
                  >
                    {propValueFormat(
                      property,
                      userProperties?.[property],
                      propType
                    ) || '-'}
                  </Text>
                </div>
              </div>

              {sortableItems.length > 1 && (
                <Button
                  type='text'
                  className='del-button'
                  onClick={() => handleDelete(index)}
                  icon={<SVG name='delete' />}
                />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

const mapDispatchToProps = (dispatch: any) =>
  bindActionCreators(
    {
      getConfiguredUserProperties
    },
    dispatch
  );

const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type ComponentProps = ReduxProps & UserDetailsProps;

export default connector(UserDetails);
