import React, {
  useState, useEffect, memo, useMemo
} from 'react';
import isEmpty from 'lodash/isEmpty';
import isArray from 'lodash/isArray';
import { QUERY_TYPE_EVENT, QUERY_TYPE_FUNNEL } from '../../../utils/constants';
import ComposerBlock from '../../QueryCommons/ComposerBlock';
import GroupBlock from '../GroupBlock';
import { areKpisInSameGroup } from '../../../utils/kpiQueryComposer.helpers';

const GroupByBlock = ({
  queryType,
  queries,
  selectedMainCategory,
  KPIConfigProps,
  groupBy,
  resetGroupByAction
}) => {
  const [groupBlockOpen, setGroupBlockOpen] = useState(true);
  const isSameKPIGrp = useMemo(() => {
    return areKpisInSameGroup({ kpis: queries });
  }, [queries]);

  useEffect(() => {
    if (
      ((!isSameKPIGrp && isArray(queries) && queries.length > 1) ||
        isEmpty(queries)) &&
      !isEmpty(groupBy?.global)
    ) {
      // we will not show global breakdown when kpis selected are from different groups
      resetGroupByAction();
    }
  }, [isSameKPIGrp, queries, groupBy, resetGroupByAction]);

  if (!isSameKPIGrp || isEmpty(queries)) {
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
        />
      </div>
    </ComposerBlock>
  );
};

export default memo(GroupByBlock);
