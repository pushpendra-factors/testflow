import React, { useState } from 'react';
import { Button, Tooltip } from 'antd';
import { SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import GroupSelect2 from '../../../../../components/QueryComposer/GroupSelect2';
import { getNormalizedKpi } from '../../../../../utils/kpiQueryComposer.helpers';

function SelectKPIBlock({ kpi, header, index, ev, attrConfig, setAttrConfig }) {
  const [isDDVisible, setDDVisible] = useState(false);

  const kpiEvents = kpi?.config?.map((item) => {
    return getNormalizedKpi({ kpi: item });
  });

  const onChange = (group, value) => {
    const opts = Object.assign({}, attrConfig);
    const newEv = { label: '', group: '' };
    newEv.label = value[0];
    newEv.group = group;
    opts.kpis_to_attribute[header] === null
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
      header === 'sf_kpi'
        ? kpiEvents.filter((ev) => ev.group == 'salesforce_opportunities')
        : header === 'hs_kpi'
        ? kpiEvents.filter((ev) => ev.group == 'hubspot_deals')
        : kpiEvents;
    return (
      <div className={`absolute`}>
        {isDDVisible ? (
          <GroupSelect2
            groupedProperties={groupedProps}
            placeholder="Select Event"
            optionClick={(group, val) => onChange(group, val)}
            onClickOutside={() => setDDVisible(false)}
            allowEmpty={true}
          />
        ) : null}
      </div>
    );
  };

  if (!ev) {
    return (
      <div className={`mt-4`}>
        <Button
          type="text"
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
          icon={<SVG name="mouseevent" size={16} color={'purple'} />}
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
        type="text"
        onClick={deleteItem}
        className={`fa-btn--custom ml-2`}
      >
        <SVG name="trash"></SVG>
      </Button>
    </div>
  );
}

const mapStateToProps = (state) => ({
  kpi: state.kpi
});
export default connect(mapStateToProps)(SelectKPIBlock);
