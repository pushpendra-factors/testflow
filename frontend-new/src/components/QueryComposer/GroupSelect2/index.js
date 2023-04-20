import React, { useState, useEffect, useMemo } from 'react';
import { Input, Button } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { CaretDownOutlined, CaretUpOutlined } from '@ant-design/icons';
import { HighlightSearchText } from '../../../utils/dataFormatter';
import useAutoFocus from '../../../hooks/useAutoFocus';

function GroupSelect2({
  groupedProperties,
  placeholder,
  optionClick,
  onClickOutside,
  extraClass,
  allowEmpty = false,
  iconColor = 'purple',
  placement = 'bottom',
  height = 576,
  additionalActions,
  useCollapseView
}) {
  const [groupCollapseState, setGroupCollapseState] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [showAll, setShowAll] = useState([]);
  const inputComponentRef = useAutoFocus();

  const groupedProps = useMemo(() => {
    return groupedProperties?.filter((group) => group?.values?.length > 0);
  }, [groupedProperties]);

  useEffect(() => {
    const flag =
      groupedProps?.length === 1 || searchTerm.length || !useCollapseView
        ? false
        : true;
    const groupColState = new Array(groupedProps.length).fill(flag);
    setGroupCollapseState(groupColState);
  }, [groupedProps, searchTerm]);

  const collapseGroup = (index) => {
    const groupColState = [...groupCollapseState];
    groupColState[index] = !groupColState[index];
    const showMoreState = [...showAll];
    showMoreState[index] = false;
    setGroupCollapseState(groupColState);
    setShowAll(showMoreState);
  };

  const setShowMoreIndex = (ind, flag) => {
    const showMoreState = [...showAll];
    showMoreState[ind] = flag;
    setShowAll(showMoreState);
  };

  const onInputSearch = (userInput) => {
    setSearchTerm(userInput.currentTarget.value);
  };

  const renderEmptyOpt = () => {
    if (!searchTerm.length) return null;
    return (
      <div key={0} className={`group`}>
        <div className={`group__options`}>
          <div
            className={`option flex items-center`}
            onClick={() => optionClick('', [searchTerm])}
          >
            <Text level={7} type={'title'} extraClass={'m-0 mr-2'}>
              Select:
            </Text>
            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
              '{searchTerm}'
            </Text>
          </div>
        </div>
      </div>
    );
  };

  const getGroupLabel = (grp) => {
    if (grp === 'event') return 'Event Properties';
    if (grp === 'user') return 'User Properties';
    if (grp === 'group') return 'Group Properties';
    if (!grp) return 'Properties';
    return grp;
  };

  const getIcon = (icon) => {
    const checkIcon = icon?.toLowerCase().split(' ').join('_');
    if (checkIcon?.includes('salesforce')) {
      return 'salesforce_ads';
    }
    if (checkIcon?.includes('hubspot')) {
      return 'hubspot_ads';
    }
    if (checkIcon?.includes('marketo')) {
      return 'marketo';
    }
    if (checkIcon?.includes('leadsquared')) {
      return 'leadSquared';
    }
    if (checkIcon?.includes('group')) {
      return 'profile';
    }
    return icon;
  };

  const renderOptions = (options) => {
    const renderGroupedOptions = [];
    if (allowEmpty) {
      renderGroupedOptions.push(renderEmptyOpt());
    }
    options?.forEach((group, grpIndex) => {
      const valuesOptions = [];

      const groupValues = group?.values?.filter((val) =>
        val[0]?.toLowerCase()?.includes(searchTerm.toLowerCase())
      );

      const groupItem = (
        <div key={group.label} className={`group`}>
          <div
            className={`group__header ${
              useCollapseView ? 'cursor-pointer' : 'cursor-default'
            }`}
            onClick={() => (useCollapseView ? collapseGroup(grpIndex) : null)}
          >
            <div>
              <SVG
                name={getIcon(group?.icon)}
                color={iconColor}
                extraClass={'self-center'}
              ></SVG>
              <Text
                level={8}
                type={'title'}
                extraClass={'m-0 ml-2'}
                weight={'bold'}
              >
                {`${getGroupLabel(group.label)} (${groupValues?.length})`}
              </Text>
            </div>
            {useCollapseView ? (
              <SVG
                color={'grey'}
                name={!groupCollapseState[grpIndex] ? 'minus' : 'plus'}
                extraClass={'self-center'}
              ></SVG>
            ) : null}
          </div>

          <div className={`group__options`}>
            {!groupCollapseState[grpIndex]
              ? (() => {
                  groupValues?.forEach((val, i) =>
                    valuesOptions.push(
                      <div
                        key={i}
                        title={val[0]}
                        className={`option`}
                        onClick={() =>
                          optionClick(
                            group.label ? group.label : group.icon,
                            val,
                            group.category,
                            group
                          )
                        }
                      >
                        {searchTerm.length > 0}
                        <Text
                          level={7}
                          type={'title'}
                          extraClass={'m-0 truncate'}
                          weight={'thin'}
                        >
                          <HighlightSearchText
                            text={val[0]}
                            highlight={searchTerm}
                          />
                        </Text>
                      </div>
                    )
                  );
                  return showAll[grpIndex]
                    ? valuesOptions
                    : valuesOptions.slice(0, 5);
                })()
              : null}
          </div>

          {valuesOptions.length > 5 && !groupCollapseState[grpIndex] ? (
            !showAll[grpIndex] ? (
              <Button
                className={`show-hide-btn`}
                type='text'
                onClick={() => {
                  setShowMoreIndex(grpIndex, true);
                }}
                icon={<CaretDownOutlined />}
              >
                Show More ({valuesOptions.length - 5})
              </Button>
            ) : (
              <Button
                className={`show-hide-btn`}
                type='text'
                onClick={() => {
                  setShowMoreIndex(grpIndex, false);
                }}
                icon={<CaretUpOutlined />}
              >
                Show Less
              </Button>
            )
          ) : null}
        </div>
      );
      groupValues?.length
        ? renderGroupedOptions.push(groupItem)
        : renderGroupedOptions.push(null);
    });
    return renderGroupedOptions;
  };

  return (
    <>
      <div
        style={{ '--height-var': `${height}px` }}
        className={`group-select height-full placement-${placement}`}
      >
        <div className={`group-select__input`}>
          <Input
            placeholder={placeholder}
            onKeyUp={onInputSearch}
            prefix={<SVG name='search' size={16} color={'grey'} />}
            ref={inputComponentRef}
          />
        </div>
        <div className={`group-select__content`}>
          {renderOptions(groupedProps)}
        </div>
        <div className={`group-select__additionalAction`}>
          {additionalActions}
        </div>
      </div>
      <div
        className={`group-select__hd_overlay`}
        onClick={onClickOutside}
      ></div>
    </>
  );
}

export default GroupSelect2;
