import React, { useState } from 'react';
import { Button, Tooltip } from 'antd';
import { SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { getNormalizedKpiWithConfigs } from '../../../../../utils/kpiQueryComposer.helpers';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import styles from '../index.module.scss';
import { processProperties } from 'Utils/dataFormatter';

function SelectKPIBlock({ kpi, header, index, ev, attrConfig, setAttrConfig }) {
  const [isDDVisible, setDDVisible] = useState(false);

  const kpiEvents = kpi?.config?.map((item) => {
    return getNormalizedKpiWithConfigs({ kpi: item });
  });

  const onChange = (option, group) => {
    const opts = Object.assign({}, attrConfig);
    const newEv = {
      label: '',
      group: '',
      value: '',
      category: '',
      kpi_query_type: ''
    };
    newEv.label = option?.label;
    newEv.group = group?.value;
    newEv.value = option?.value;
    newEv.kpi_query_type = option?.extraProps?.queryType;
    newEv.category = group.extraProps?.category;
    !opts.kpis_to_attribute[header]
      ? (opts.kpis_to_attribute[header] = [])
      : opts.kpis_to_attribute[header];
    index > opts.kpis_to_attribute[header].length
      ? opts.kpis_to_attribute[header].push(newEv)
      : (opts.kpis_to_attribute[header][index] = newEv);
    setAttrConfig(opts);
    setDDVisible(false);
  };

  const deleteItem = () => {
    const opts = Object.assign({}, attrConfig);
    opts.kpis_to_attribute[header] = attrConfig.kpis_to_attribute[header];
    opts.kpis_to_attribute[header].splice(index, 1);
    setAttrConfig(opts);
  };

  const selectKPI = () => {
    const groupedProps =
      (header === 'sf_kpi'
        ? kpiEvents.filter((ev) => ev.group == 'salesforce_opportunities')
        : header === 'hs_kpi'
        ? kpiEvents.filter((ev) => ev.group == 'hubspot_deals')
        : kpiEvents
      )?.map((groupOpt) => {
        return {
          iconName: groupOpt?.icon,
          label: _.startCase(groupOpt?.label),
          value: groupOpt?.label,
          extraProps: {
            category: groupOpt?.category
          },
          values: processProperties(groupOpt?.values)
        };
      }) || [];
    return (
      <div className={styles.filter__event_selector}>
        {isDDVisible ? (
          <GroupSelect
            options={groupedProps}
            searchPlaceHolder='Select Event'
            optionClickCallback={onChange}
            onClickOutside={() => setDDVisible(false)}
            allowSearch={true}
            allowSearchTextSelection={false}
            extraClass={styles.filter__event_selector__select}
          />
        ) : null}
      </div>
    );
  };

  if (!ev) {
    return (
      <div className={`mt-4`}>
        <Button
          type='text'
          onClick={() => {
            setDDVisible(true);
          }}
          icon={<SVG name={'plus'} color={'grey'} />}
        >
          Add New
        </Button>
        {selectKPI()}
      </div>
    );
  }

  return (
    <div className={`flex items-center mt-4`}>
      <Tooltip title={ev?.label ? ev.label : ev}>
        <Button
          icon={
            <SVG name={getGroupIcon(ev?.group)} size={16} color={'purple'} />
          }
          type={'link'}
          onClick={() => {
            setDDVisible(true);
          }}
          block={true}
        >
          {ev?.label ? ev.label : ev}
        </Button>
        {selectKPI()}
      </Tooltip>
      <Button
        size={'large'}
        type='text'
        onClick={deleteItem}
        className={`fa-btn--custom ml-2`}
      >
        <SVG name='trash'></SVG>
      </Button>
    </div>
  );
}

const mapStateToProps = (state) => ({
  kpi: state.kpi
});
export default connect(mapStateToProps)(SelectKPIBlock);
