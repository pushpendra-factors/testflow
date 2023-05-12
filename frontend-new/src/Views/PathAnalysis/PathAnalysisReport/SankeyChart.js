import React, { useEffect, useState, useCallback } from 'react';
import Highcharts from 'highcharts'; 
import * as highchartsSankey from 'highcharts/modules/sankey';
import { Timeline, Empty, Button} from 'antd'; 
import ReactDOMServer from 'react-dom/server'; 
import _ from 'lodash';
import { SVG, Text } from 'factorsComponents'; 
import { useHistory } from 'react-router-dom';

const StepArraowGenerator = (activeQuery) => {
    let eventName = activeQuery?.query?.event?.label;
    let count = Number(activeQuery?.steps)
    let isReverse = activeQuery?.query?.event_type == "startswith" ? true : false;

    let url = StripUrl(eventName);
    let finalUrl = truncateString(url, 25);
    return (
        <div className={`custom-table-headers--sankey ${isReverse ? 'reverse' : ''}`}>
            {[...Array(count).keys()].map((item, index) => {
                return <div className='table-column'>
                    <p>{index == 0 ? `${finalUrl}` : `Step ${index}`}</p>
                    <div class={`header-arrow ${isReverse ? 'reverse' : ''}`} />
                    <div class={`header-arrow--white ${isReverse ? 'reverse' : ''}`} />
                </div>
            })}
        </div>
    )
}

const DataBuildMessage = () => {

    const history = useHistory();
    const routeChange = (url) => {
      history.push(url);
    };

    return <div className='flex flex-col items-center justify-center mt-20'>
        <img style={{maxWidth: '200px',height: 'auto' }} src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/report-building.png' />
        <Text type={'title'} weight={'bold'} extraClass={'mt-4'} level={6}>Your report is being built</Text>
        <Text type={'title'} weight={'thin'} level={7}>This might take a while.</Text> 
    </div>
}


const truncateString = (str, num) => {
    if (str.length <= num) {
        return str;
    }
    return str.slice(0, num) + '...';
}

const StripUrl = (url) => {
    let finalUrl = url ? url.replace(/^[0-9]:/, '').replace(/(^\w+:|^)\/\//, '').toString() : "";
    return finalUrl
}

const CustomTooltip = ({ data }) => {
    // console.log('inside CustomTooltip-->',data)
    const fromName = data?.point?.from
    const toName = data?.point?.to
    const value = data?.point?.weight
    if (fromName && toName) {
        return (
            <div className='fa-sankey--tooltip'>
                <div className='fa-sankey--tooltip-wrapper'>
                    <Timeline>
                        <Timeline.Item><p>{StripUrl(fromName)}</p></Timeline.Item>
                        <Timeline.Item><p>{StripUrl(toName)}</p></Timeline.Item>
                        <p style={{ fontWeight: 'bold', fontSize: '14px' }}>{value}</p>
                    </Timeline>
                </div>
            </div>)
    }
    else {
        return null
    }
};


function Sankey({
    sankeyData,
    activeQuery
}) {
    const [chartData, setChartData] = useState(false);
    const [reverseChart, setReverseChart] = useState(true);

    useEffect(() => {
        let isReverse = activeQuery?.query?.event_type == "startswith" ? false : true
        setReverseChart(isReverse)
        setChartData(transformDataFn(sankeyData, isReverse, 2)) 
    }, [activeQuery, sankeyData])

    if (typeof Highcharts === 'object') {
        highchartsSankey(Highcharts);
    }

    const getChartTitleFn = data => {
        if (data) {
            let results = data;
            for (const index of Object.keys(results)) {
                if (index == '1') {
                    return StripUrl(Object.keys(results[index])[0])
                }
            }
        }
    }
    const transformDataFn = (data, isReverse, version = 1) => { 
        // console.log('input chart data-->', data); 
        if (data) {
            let results = data;
            let finalArr = [];
            let final = [];
            let title = '';

            if(version==2){

                data?.map((item,index)=>{
                    if(index!=0){
                        let arr = item?.Key.split(',');  
                        if (isReverse) {
                            final = [arr[0], arr[1], item?.Count]
                           finalArr = [...finalArr, [...final]];
                        } else {
                            let last2count = arr.length - 2;
                            let last2El = arr.slice(last2count); 
                            final = [...last2El, item?.Count];
                           finalArr = [...finalArr, [...final]];
                        }
                    }
                });
            }
            else{

                for (const index of Object.keys(results)) {
                    if (index != '1') {
                        if (results[index]) {
                            for (const key of Object.keys(results[index])) {
                                let arr = key.split(',');
                                if (isReverse) {
                                    final = [arr[0], arr[1], results[index][key]]
                                } else {
                                    let last2count = arr.length - 2;
                                    let last2El = arr.slice(last2count);
                                    final = [...last2El, results[index][key]];
                                }
                                finalArr = [...finalArr, [...final]];
                                //   console.log("finalArr",finalArr)
                            }
                        }
                    }
                }
            }

            // console.log('transformed chart data-->', finalArr); 
            return finalArr
        }
        return null;
    };

    const options = {
        title: {
            // text: `${reverseChart ? 'Ends with' : 'Starts with'} ${getChartTitleFn(chartData)}`,
            text: ""
        },

        accessibility: {
            point: {
                valueDescriptionFormat:
                    '{index}. {point.from} to {point.to}, {point.weight}.',
            },
        },
        chart:{
            overflowY: 'scroll',
            height: chartData ? (chartData.length*8 < 500 ? 500 :  chartData.length* 8) : 500,
        },

        plotOptions: {
            sankey: {
                nodeWidth: 60,
                colorByPoint: false,
                curveFactor: 0.5,
                nodePadding: 40,
                borderRadius: 4,
                centerInCategory: true,
                backgroundColor: 'transparent',
                color: '#82AEE0',
                minLinkWidth: 2,
                linkOpacity: 0.2,
                allowPointSelect: true,
                relativeXValue: true,
                useHTML: true,
                clip: false,
                states: {
                    hover: {
                        linkOpacity: 0.5,
                    },
                },
                dataLabels: {
                    backgroundColor: 'transparent',
                    align: 'left',
                    inside: false,
                    useHTML: true,
                    nodeFormatter() {
                        let url = StripUrl(this.key);
                        let finalUrl = truncateString(url, 25);
                        // let checkIfNode1 = this.point.column == 0 ? TotalCount : this.point.sum;
                        let checkIfNode1 = this.point.sum;
                        return ReactDOMServer.renderToString(
                            <div
                                style={{
                                    display: 'flex',
                                    flexDirection: 'column',
                                    marginTop: '0px',
                                }}
                            >
                                <h1
                                    style={{
                                        fontSize: '10px',
                                        fontWeight: 'bold',
                                        marginBottom: '4px',
                                    }}
                                >
                                    {finalUrl}
                                </h1>
                                <h1 style={{ fontSize: '8px', margin: '0px' }}>
                                    {checkIfNode1}
                                </h1>
                            </div>
                        );
                    },
                },
                point: {
                    events: {
                        mouseOver: false,
                    },
                },
  

                //   label:{
                //       useHTML: true,
                //       color: 'red',
                //       nodeFormatter() {
                //         console.log("thisssss label->>>",this);
                //         let url = this.key.replace(/^[0-9]:/,'').replace(/(^\w+:|^)\/\//, '')
                //         return ReactDOMServer.renderToString(<div><h1 style={{fontSize: '10px'}}>{url}</h1></div>)
                //     }
                //   },
            },
        },
        // Main tooltip
        tooltip: {
            backgroundColor: 'red',
            borderColor: "transparent",
            borderWidth: 0,
            borderRadius: 2,
            shadow: 'none',
            outside: false,
            useHTML: true,
            className: 'fa-sankey-diagram--tooltip',
            formatter() {
                const self = this;
                // console.log("tooltip-->", self)
                return ReactDOMServer.renderToString(
                    <CustomTooltip data={self} />
                );
            },
        },
        credits: {
            enabled: false
        },

        series: [
            {
                type: 'sankey',
                keys: reverseChart ? ['to', 'from', 'weight'] : ['from', 'to', 'weight'],
                // keys: ['from', 'to', 'weight'],
                data: chartData ? chartData : [],
                // data: transformDataFn(sankeyData),
                clip:false,
                centerInCategory: true,
                className: "fa-sankey-diagram",

                // tooltip: {
                //     headerFormat: undefined,
                //     enabled: false,
                //     outside: false,
                //     className: 'custom-div',
                //     nodeFormat: undefined, 
                //     shadow: false,
                //     useHTML: true,
                //     pointFormatter: function () {
                //         return ReactDOMServer.renderToString(<CustomTooltip />);
                //     },
                // },
            },
        ],
    };



 
    const drawChart = useCallback(() => { 
        Highcharts.chart("fa-sankey-container", {
            ...options
        })
    });

    useEffect(() => {
        if (chartData && !_.isEmpty(chartData))
        {
            drawChart();
        }
      }, [reverseChart, chartData ]);

    return (
        <> 
                    <div className='mt-16 mb-10'>

                        {/* {StepArraowGenerator(activeQuery)} */}


                        {/* <HighchartsReact
                            highcharts={Highcharts}
                            options={options}
                        /> */}

                        {(chartData && _.isEmpty(chartData)) ?  (activeQuery?.status == 'building' || activeQuery?.status == 'saved') ? <DataBuildMessage /> : <Empty /> : 
                    <div className='fa-sankey-container' id="fa-sankey-container" />
                        }


                    </div> 
        </>
    );
}

export default Sankey;
