import { ArrowLeftOutlined, LoadingOutlined, MacCommandOutlined, SearchOutlined, VerticalRightOutlined } from "@ant-design/icons";
import { Button, Input } from "antd";
import styles from './index.module.scss';
import React, { useEffect, useState } from "react"
import { SVG, Text } from "Components/factorsComponents";
import { useDispatch, useSelector } from "react-redux";
import VirtualList from 'rc-virtual-list';
import { getQueryType } from "Utils/dataFormatter";
import {generatePath, Link, useHistory} from 'react-router-dom'
import { TOGGLE_GLOBAL_SEARCH } from "Reducers/types";
import { QUERY_TYPE_ATTRIBUTION, QUERY_TYPE_EVENT, QUERY_TYPE_FUNNEL, QUERY_TYPE_KPI, QUERY_TYPE_PROFILE } from "Utils/constants";

const itemHeight = 47;
const ContainerHeight = 443;

const Part1GlobalSearch = ({items,setStep, showAllCreateNew, showAllReports})=>{
    
    const history = useHistory()
    const dispatch = useDispatch()
    const queries = useSelector(state=>state?.queries?.data?.slice(0,5));
    const queriesLoading = useSelector(state=>state?.queries?.loading);
    
    useEffect(()=>{
        return ()=>{
            
        }
    },[])

    const moveToPath = (tpath)=>{
        dispatch({type: TOGGLE_GLOBAL_SEARCH})
        history.push({pathname: tpath })
    }
    return <React.Fragment>
        {/* Below is for Create New Container */}
        <div>
        <div className={styles["each-container"]}>
            <div className={styles["create-new-title"]} style={{padding: '0 10px'}}><Text color='grey' level={6} type={'title'} weight={'bold'}>Create New Report</Text></div>
            <div className={styles["create-new-container"]}>
                {items.create.data.slice(0,6).map((eachItem, eachIndex)=>{
                    return <div  
                            key={eachIndex} 
                            className={styles["create-new-items-container"]} 
                            onClick={()=>moveToPath(eachItem.path)} 
                            onKeyUp={(e)=>e.key === 'Enter' ? moveToPath(eachItem.path) : ''}
                            >
                    <div className={styles["create-new-items"]} tabIndex={0}> 
                        <div className={styles["create-new-item-icon"]}>{eachItem.icon}</div> 
                        {eachItem.name}
                        
                    </div>
                </div>
                    })}
            </div>
            <span className={styles['globalSearchShowBtn']} 
                onClick={showAllCreateNew} 
                onKeyUp={(e)=>e.key === 'Enter' ? showAllCreateNew() : ''} 
                tabIndex={0}>Show all</span>
        </div>
        {/* Below is for Reports */}
        <div className={styles["each-container"]}>
            <div className={styles["create-new-title"]} style={{padding: '0 10px'}}><Text color='grey' level={6} type={'title'} weight={'bold'}>Recent reports/Dashboards</Text></div>
            {queriesLoading === false ? <React.Fragment> <div className={styles["reports-new-container"]}>
            {queries.map((eachItem, eachIndex)=>{
                    const queryType = getQueryType(eachItem.type)
                    const queryTypeName = {
                        events: 'events_cq',
                        funnel: 'funnels_cq',
                        channel_v1: 'campaigns_cq',
                        attribution: 'attributions_cq',
                        profiles: 'profiles_cq',
                        kpi: 'KPI_cq'
                      };
                      let svgName = '';
                      Object.entries(queryTypeName).forEach(([k, v]) => {
                        if (queryType === k) {
                          svgName = v;
                        }
                      });
                    return <div key={eachIndex} className={styles["reports-new-items-container"]}>
                        <div className={styles["reports-new-items"]} tabIndex={0}> 
                       <div className={styles["reports-new-item-icon"]}><SVG name={svgName} size={20} color='blue' /></div> 
                        {eachItem.title}
                </div>
                    </div>
                })}
            </div> <span onClick={showAllReports} onKeyUp={(e)=>e.key === 'Enter' ? showAllReports() : ''} className={styles['globalSearchShowBtn']} tabIndex={0}>Show all</span> </React.Fragment>
            : 
            <div style={{alignItems:'center', display:'flex', justifyContent:"center"}}><LoadingOutlined size={'20px'} style={{margin: '0 10px'}} /> Loading Reports ...</div>}
            
            
        </div>
        </div>
    </React.Fragment>
}   
const Part2GlobalSearch = ({data, moveBackStep1, step2Type})=>{
    const [state, setState] = useState({});
    const [type, setType] = useState(null)
    const history = useHistory()
    const dispatch = useDispatch()
    useEffect(()=>{
        setType(step2Type)
    },[step2Type])
    useEffect(()=>{
        setState(data)
    },[data])
    const moveToPath = (tpath)=>{
        dispatch({type: TOGGLE_GLOBAL_SEARCH})
        history.push({pathname: tpath })
    }
    return <React.Fragment>
        { type && type === 1 ? <div className={styles['globalsearch-step2-container']}>
            <div className={styles['globalsearch-step2-title']}> 
                <div>
                    <Button
                        size='large'
                        type='text'
                        icon={<ArrowLeftOutlined />}
                        onClick={moveBackStep1}
                        onKeyUp={(e)=>e.key === 'Enter' ? moveBackStep1() : ''}
                    />
                </div>  {state.title}
            </div>
            <div className={styles['globalsearch-item-list']}>
                {data.data.map((eachItem, eachIndex)=>{
                    return <div 
                                key={eachItem.fullName+eachIndex} 
                                className={styles['globalsearch-item-list-item']} 
                                onClick={()=>moveToPath(eachItem.path)} 
                                onKeyUp={(e)=>e.key === 'Enter' ? moveToPath(eachItem.path) : ''}
                                tabIndex={0}> 
                                {eachItem.icon}
                                <div>
                                    <Text level={6} type={'paragraph'} weight={'normal'} color='#0E2647'>{eachItem.fullName}</Text>
                                    <div className={styles['globalsearch-item-list-item-desc']}>{eachItem.description}</div>
                                </div>
                            </div>
                })}
            </div>
        </div> : ''}
        {type && type === 2 ? <div className={styles['globalsearch-step2-container']}>
            <div className={styles['globalsearch-step2-title']}> 
                <div>
                    <Button
                        size='large'
                        type='text'
                        icon={<ArrowLeftOutlined />}
                        onClick={moveBackStep1}

                        onKeyUp={(e)=>e.key === 'Enter' ? moveBackStep1() : ''}
                    />
                </div>  Recent reports/Dashboards
            </div>
            <div className={styles['globalsearch-item-list']}>
            <VirtualList
                data={data}
                height={ContainerHeight}
                itemHeight={itemHeight}
                itemKey='id'
              >
                {(eachItem, eachIndex)=>{
                    const queryType = getQueryType(eachItem.query)
                    const queryTypeName = {
                        events: 'events_cq',
                        funnel: 'funnels_cq',
                        channel_v1: 'campaigns_cq',
                        attribution: 'attributions_cq',
                        profiles: 'profiles_cq',
                        kpi: 'KPI_cq'
                      };
                      let svgName = '';
                      Object.entries(queryTypeName).forEach(([k, v]) => {
                        if (queryType === k) {
                          svgName = v;
                        }
                      });
                    return <div key={eachItem?.id+eachIndex} className={styles['globalsearch-item-list-item']} tabIndex={0}>
                        <SVG name={svgName} size={20} color='blue' />
                        <Text level={6} type={'paragraph'} weight={'normal'} color='#0E2647'>{eachItem.title}</Text>
                        
                        </div>
                }}
            </VirtualList>
                
            </div>
        </div> : ''}
    </React.Fragment>
}
const SearchResults = ({searchString})=>{
    const [searchResults, setSearchResults] = useState([])
    const [finalResults, setFinalResults] = useState([])
    const allRoutes = useSelector(state=>state.allRoutes.data)
    const dispatch = useDispatch()
    const history = useHistory()
    const SEARCH_TYPES = {
        'settings': 'settings',
        'configure':'configure',
        'explain':'explain',
        'explainV2':'explain',
        'template':'',
        'welcome':'',
        'analyse':'analysis',
        'components':'',
        'profiles':'profile',
        'attribution':'attribution',
        'reports':'reports',
        'path-analysis':'pathAnalysis',
        'project-setup':'',
        '':'dashboardFilled'
    }
    const ommitRoutes = new Set(["components","explainV2","project-setup", "template","welcome"]);
    const checkRoute = (eachRoute)=>{
        return ommitRoutes.has(eachRoute) 
    }
    useEffect(()=>{
        let ss = searchString.trim();
        if(ss.length>0 && ss[0] === ' '){
            ss= ss.slice(1, searchString.length)
        }
        let filtered = searchResults.filter((eachPath, eachIndex)=>{
            if(eachPath.toLowerCase().includes(ss.toLowerCase())) return true; else return false;
        })
        let sss = 'dashboard';
        if(sss.includes(ss.toLowerCase())){
            filtered.push('/');
        }
        setFinalResults(Array.from(new Set(filtered)))
    },[searchString, searchResults])
    useEffect(()=>{
    
        let filteredResults = allRoutes && allRoutes?.filter((eachEle)=>{
            if(eachEle.includes(':')) return false;
            let tmparr1 = eachEle.split('/');
            let n = tmparr1.length;

            if(n <= 1) return false;

            tmparr1 = tmparr1.slice(1,n);
            
            if(tmparr1[0].length===0){
                return false;
            }
            return true;
        });
        setFinalResults(filteredResults)
        setSearchResults(Array.from(new Set(filteredResults)));
        
    },[allRoutes])
    const getSearchType = (route)=>{
        let arr = route.split('/');
        let type = arr[1];

        return type;
    }
    const renderRoute = (route)=>{
        let arr = route.split('/');
        let n = arr.length;
        let selectedPaths = [];
        let ans = '';
        for(let i=0; i < n; i++){
            if(arr[i].length > 0){
                ans += arr[i];
                
                if(i < n-1){
                    ans += ' > ';
                }
                selectedPaths.push(arr[i]);
            }
        }
        

        return ans;
    }
    const moveToRoute = (path)=>{
        history.push(path)
        dispatch({type: TOGGLE_GLOBAL_SEARCH})
    }
        return <div className={styles["searchresults-container"]}>
            <div className={styles["searchresults-container-result-item-container"]}>
                {Array.isArray(finalResults) && finalResults?.map((eachRoute, eachIndex)=>{
                    return <React.Fragment key={eachIndex}>
                        <div 
                            className={styles["searchresults-container-result-item"]} 
                            onClick={()=>moveToRoute(eachRoute)}
                            onKeyUp={(e)=>e.key === 'Enter' ? moveToRoute(eachRoute) : ''}
                            tabIndex={0}>
                            
                        {checkRoute(eachRoute) ? '': <SVG name={SEARCH_TYPES[getSearchType(eachRoute)]} size={20} color={'blue'} /> }
                        
                        <Text level={6} type={'paragraph'} weight={'normal'} color='#0E2647'>{eachRoute !== '/' ? renderRoute(eachRoute) : 'Dashboard'}</Text>
                        </div>


                    </React.Fragment>
                })}
            </div>
        </div>

}
const GlobalSearch = ()=>{
    const items = {
        "create":{
            title: "Create New Report",
            data:[
                {name:"KPIs", fullName: "KPI Report", description: "Measure performance over time", icon: <SVG name={`KPI_cq`} size={20} color={'blue'} />, path: '/analyse/'+QUERY_TYPE_KPI},
                {name:"Funnels", fullName: "Funnel Report", description: "Track how users navigate", icon: <SVG name={`Funnels_cq`} size={20} color={'blue'} />, path: '/analyse/'+QUERY_TYPE_FUNNEL},
                {name:"Events", fullName: "Event Report", description: "Track and chart events", icon: <SVG name={`Events_cq`} size={20} color={'blue'} />, path: '/analyse/'+QUERY_TYPE_EVENT},
                {name:"Attribution", fullName: "Attribution Report", description: "Track and chart events", icon: <SVG name={`Attributions_cq`} size={20} color={'blue'} />, path: '/analyse/'+QUERY_TYPE_ATTRIBUTION},
                {name:"Profiles", fullName: "Profiles Report", description: "Slice and dice your visitors", icon: <SVG name={`Profiles_cq`} size={20} color={'blue'} />, path: '/analyse/'+QUERY_TYPE_PROFILE},
                {name:"Path Analysis", fullName: "Path Analysis Report", description: "Track and chart events", icon: <SVG name={`PathAnalysis`} size={20} color={'blue'} />, path: '/path-analysis'},
                {name:"Explain", fullName: "Explain", description: "Track and chart events", icon: <SVG name={`Explain`} size={20} color={'blue'} />, path: '/explain'},
                {name:"Website visitors identification", fullName: "Website visitors identification", description: "Track and chart events", icon: <SVG name={`Funnel_cq`} size={20} color={'blue'} />, path: '/'},
            ]
        },
        "reports":{
            title:"Recent reports/Dashboards",
            data:[
                
            ]
        }
    }
    const [step, setStep] = useState(1);
    const [step2Content, setStep2Content] = useState([])
    const [searchString, setSearchString] = useState('')
    const queries = useSelector(state=>state.queries.data)
    const [step2Type, setStep2Type] = useState(null)
    const showAllCreateNew = ()=>{
        setStep(2);
        setStep2Type(1)
        setStep2Content(items.create);
    }

    const showAllReports = ()=>{
        setStep(2);
        setStep2Type(2)
        setStep2Content(queries);
    }
    const moveBackStep1 = ()=>{
        setStep(1);
        setStep2Content([]);
    }
    const onChangeInput = (e)=>{
        setSearchString(e.target.value)
    }
    const moveToStep2 = (value)=>{
        setStep(2);
    }
    useEffect(()=>{
        if(step == 2){
        }else{

        }
    },[step])
    
    return <div className={styles["globalsearch-container"]} style={{transitionDuration:"1s"}}>
        <div style={{display:'flex', alignItems:"center", padding:"5px 10px 0 10px"}}>
            <Input 
                onChange={onChangeInput}
                className={styles['input-globalSearch']}
                placeholder="Search or jump to" 
                prefix={<SearchOutlined style={{color: '#B7BEC8'}} />} 
                style={{
                    borderRadius:"12px", 
                    border:'none', 
                    height:'56px', 
                    color: '#B7BEC8',
                    boxShadow:'none',
                    
                    }} 
                /> 
            <div style={{padding: '0 5px'}}><SVG name='command' width='32px' height='32px' /></div>
            <div style={{padding: '0 5px'}}><SVG name='letterk' width='32px' height='32px' /></div>
        </div>
        {searchString.length !== 0 ? 
            <SearchResults searchString={searchString} /> : 
            step === 1 ? <Part1GlobalSearch items={items} setStep={moveToStep2} showAllCreateNew={showAllCreateNew} showAllReports={showAllReports} /> 
                :
                <Part2GlobalSearch moveBackStep1={moveBackStep1} setStep={setStep} data={step2Content} step2Type={step2Type} />
        }
    </div>
}
export default GlobalSearch;