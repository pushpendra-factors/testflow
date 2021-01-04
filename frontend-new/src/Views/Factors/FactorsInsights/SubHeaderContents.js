import React, { useEffect } from 'react';
import {  Button , Badge} from 'antd';
import { SVG, Text} from 'factorsComponents';
import { Link } from 'react-router-dom'; 
import {connect} from 'react-redux'; 
import _, { isEmpty } from 'lodash';

function Header({factors_insight_rules}) {
  

//   useEffect(() => {

//   }, [factors_insight_rules]);

  if(factors_insight_rules){
      return ( 
        <div className={'fa-container'}>
             <div className="flex flex-col justify-between border-bottom--thin-2 pb-4" style={{borderBottomWidth:'3px'}}> 
                    <Text type={'title'} level={2} color={'grey'} weight={'bold'} color={'grey-3'} extraClass={'m-0 '}>{_.isEmpty(factors_insight_rules.name) ? 'Untitled Name' : factors_insight_rules.name }</Text>
                    <div className="flex items-center">
                        <Badge count={'Goal'} className={'fa-custom-badge'} />
                        {factors_insight_rules.rule.st_en ? <>
                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 ml-2'}>{factors_insight_rules.rule.st_en}</Text> 
                            <Text type={'title'} level={4} color={'grey'} extraClass={'m-0 ml-2'}>and</Text>
                        </> : null}
                        <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 ml-2'}>{factors_insight_rules.rule.en_en}</Text>
                        {!_.isEmpty(factors_insight_rules?.rule?.rule?.ft) ? <>
                            <Text type={'title'} level={4} color={'grey'} extraClass={'m-0 ml-2'}>where</Text>
                            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 ml-2'}>Untitled</Text>
                        </> : null
                        }
                    </div>
            </div> 
        </div> 
      );
  }
  else return null
}

const mapStateToProps = (state) => {
  return { 
    factors_insight_rules: state.factors.factors_insight_rules
  };
};
export default connect(mapStateToProps, null)(Header);
