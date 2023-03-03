import React, {
  useState, useEffect, memo, useMemo
} from 'react';
import isEmpty from 'lodash/isEmpty';
import isArray from 'lodash/isArray';
import { QUERY_TYPE_EVENT, QUERY_TYPE_FUNNEL } from '../../../utils/constants';
import ComposerBlock from '../../QueryCommons/ComposerBlock';
import GroupBlock from '../GroupBlock';
import { areKpisInSameGroup } from '../../../utils/kpiQueryComposer.helpers'; 
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

const GroupByBlock = ({
  queryType,
  queries,
  selectedMainCategory,
  KPIConfigProps,
  groupBy,
  resetGroupByAction, 
  activeProject,
  propertyMaps
}) => {
  const [groupBlockOpen, setGroupBlockOpen] = useState(true);
  const isSameKPIGrp = useMemo(() => {
    return areKpisInSameGroup({ kpis: queries });
  }, [queries]);

  // useEffect(() => {
  //   if (
  //     ((!isSameKPIGrp && isArray(queries) && queries.length > 1) ||
  //       isEmpty(queries)) &&
  //     !isEmpty(groupBy?.global)
  //   ) {
  //     // we will not show global breakdown when kpis selected are from different groups
  //     resetGroupByAction();
  //   }
  // }, [isSameKPIGrp, queries, groupBy, resetGroupByAction]);

  // useEffect(() => { 
  //     // we will not show global breakdown when kpis selected are from different groups
  //     if (!isEmpty(queries)) { 
  //       resetGroupByAction();  
  //     }
  // }, [queries]); 
  
  if (isEmpty(queries)) {
    return null;
  }

  if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
    return null;
  }
  if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
    return null;
  }

  return (
    <ComposerBlock
      blockTitle={'BREAKDOWN'}
      isOpen={groupBlockOpen}
      showIcon={true}
      onClick={() => setGroupBlockOpen(!groupBlockOpen)}
      extraClass={'no-padding-l'}
    >
      <div key={0} className={'fa--query_block borderless no-padding '}>
        <GroupBlock
          queryType={queryType}
          events={queries}
          selectedMainCategory={selectedMainCategory}
          KPIConfigProps={KPIConfigProps}
          isSameKPIGrp={isSameKPIGrp}
          propertyMaps={propertyMaps}
        />
      </div>
    </ComposerBlock>
  );
};


const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  propertyMaps: state.kpi.kpi_property_mapping,
});
 
export default memo(
  connect(mapStateToProps, null)(GroupByBlock)
);
