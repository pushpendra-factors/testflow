import React, { useEffect, useState } from 'react';
import { Text, SVG } from 'factorsComponents';
import { Button, Table, Avatar, Menu, Dropdown, Modal, message, Badge, Input } from 'antd';
import { useHistory } from 'react-router-dom';
import { fetchSavedPathAnalysis, removeSavedQuery, fetchPathAnalysisInsights,setActiveInsightQuery } from 'Reducers/pathAnalysis';
import { connect } from 'react-redux';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import moment from 'moment';


const { confirm } = Modal;




const SavedGoals = ({ savedQuery, setShowReport, SetfetchingIngishts,setActiveInsightQuery, removeSavedQuery, fetchSavedPathAnalysis, activeProject, fetchPathAnalysisInsights }) => {

  const [loadingTable, SetLoadingTable] = useState(true);
  const [dataSource, setdataSource] = useState(null);
  const [showSearch, setShowSearch] = useState(false);
  const [tableData, setTableData] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');

  const history = useHistory();

  const searchReport = (e) => {
    let term = e.target.value;
    setSearchTerm(term);
    let searchResults = dataSource.filter((item) => {
      return item?.title?.toLowerCase().includes(term.toLowerCase());
    }); 
    setTableData(searchResults);
  };

  const menu = (values) => {
    return (
      <Menu>
        <Menu.Item key="0" onClick={() => confirmRemove(values)}>
          <a>Delete Report</a>
        </Menu.Item>
      </Menu>
    );
  };

  const getInsights = (data) => { 
    SetfetchingIngishts(true);
    setActiveInsightQuery(data?.query)
    const getData = async () => {
      await fetchPathAnalysisInsights(activeProject?.id, data?.id);
    };
    getData().then(() => {
      history.push('/path-analysis/insights');
      SetfetchingIngishts(false)
      setShowReport(true);
    });
  };

  const statusRefreshColumn = () => {
    return <div className='flex items-center mr-1'>
      <Button
        loading={loadingTable}
        icon={<SVG name={'syncAlt'} size={18} color={'grey'} />}
        type="text"
        ghost={true}
        shape="square"
        onClick={() => {
          SetLoadingTable(true)
          fetchSavedPathAnalysis(activeProject?.id).then(() => {
            SetLoadingTable(false)
          })
        }} />
    </div>
  }

  const columns = [
    {
      title: 'Title',
      dataIndex: 'data',
      key: 'data',
      width: '350px',
      render: (data) => <Text type={'title'} level={7} truncate={true}
      extraClass={`${(data?.status == 'saved' || data?.status == 'building') ? "" : "cursor-pointer"} m-0`} 
      onClick={(data?.status == 'saved' || data?.status == 'building') ? "" : () => getInsights(data)}
      // onClick={() => getInsights(data)}
      >{data?.title}</Text>
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => <div className="flex items-center"> {(status == 'saved' || status == 'building') ? <Badge className={'fa-custom-badge fa-custom-badge--orange'} status="processing" text={'Building'} /> : <Badge className={'fa-custom-badge fa-custom-badge--green'} status="success" text={'Ready'} />}</div>
    },
    {
      title: 'Created By',
      dataIndex: 'author',
      key: 'author',
      render: (text) => <div className="flex items-center">
        <Avatar src="assets/avatar/avatar.png" className={'mr-2'} size={24} /><Text type={'title'} level={7} extraClass={'cursor-pointer m-0 ml-2'} >{text}</Text>
      </div>
    },
    {
      title: 'Date',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => <Text type={'title'} level={7} extraClass={'m-0'}>{moment(date).format('MMM DD, YYYY')}</Text>
    },
    {
      title: '',
      dataIndex: 'data',
      key: 'data',
      render: (values) => (
        <Dropdown overlay={() => menu(values)} trigger={['hover']}>
          <Button type="text" icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];


  const confirmRemove = (values) => {

    confirm({
      title: 'Are you sure you want to remove this saved report?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        removeSavedQuery(activeProject?.id, values?.id).then(() => {
          message.success('Saved report removed!');
          fetchSavedPathAnalysis(activeProject?.id)
        }).catch((err) => {
          message.error(err);
        });
      }
    });

  };



  useEffect(() => {
    SetLoadingTable(true);
    setdataSource(null);
    if (savedQuery) {
      const formattedArray = [];
      savedQuery.map((data, index) => {
        let createdUser = '';
        formattedArray.push({
          key: index,
          title: data?.title,
          status: data?.status,
          author: data?.created_by,
          created_at: data?.date,
          data: data
        });
        setdataSource(formattedArray);
      });
      SetLoadingTable(false);
    }
    else{
      setdataSource([]);
      SetLoadingTable(false);
    }
  }, [savedQuery]);


  //   useEffect(()=>{
  //     fetchSavedPathAnalysis(activeProject?.id)
  //   },[]);

  return (<div>
    <div className='flex items-end justify-between mt-10 mb-2'>
      <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>{'Saved Paths'}</Text>

      <div className='flex justify-end'>
        {statusRefreshColumn()}
        <div className={'flex items-center justify-between'}>
          {showSearch ? (
            <Input
              autoFocus
              onChange={(e)=>searchReport(e)} 
              placeholder={'Search reports'}
              style={{ width: '220px', 'border-radius': '5px' }}
              prefix={
                <SVG name="search" size={16} color={'grey'} />
              }
            />
          ) : null}
          <Button
            type="text"
            ghost={true}
            shape="circle"
            className={'p-2 bg-white'}
            onClick={() => {
              setShowSearch(!showSearch);
              if (showSearch) {
                setSearchTerm('');
              }
            }}
          >
            <SVG
              name={!showSearch ? 'search' : 'close'}
              size={20}
              color={'grey'}
            />
          </Button>
        </div>
      </div>
    </div>
    <Table loading={loadingTable} className="fa-table--basic"
      columns={columns}
      dataSource={searchTerm ? tableData : dataSource}
      pagination={true}
    />
  </div>
  );
};


const mapStateToProps = (state) => {
  return {
    savedQuery: state.pathAnalysis.savedQuery,
    activeProject: state.global.active_project,
  };
};


export default connect(mapStateToProps, { fetchSavedPathAnalysis, removeSavedQuery, fetchPathAnalysisInsights, setActiveInsightQuery })(SavedGoals);
