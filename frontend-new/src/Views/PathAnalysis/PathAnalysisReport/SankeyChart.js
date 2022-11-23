import React, { useEffect, useState, useCallback } from 'react';
import Highcharts from 'highcharts'; 
import * as highchartsSankey from 'highcharts/modules/sankey';
import { Timeline} from 'antd'; 
import ReactDOMServer from 'react-dom/server'; 

const StepArraowGenerator = (activeQuery) => {
    let eventName = activeQuery?.event?.label;
    let count = Number(activeQuery?.steps)
    let isReverse = activeQuery?.event_type == "startswith" ? true : false;

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
            <div className='custom-div-style'>
                <div className='wrapper'>
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
        let isReverse = activeQuery?.event_type == "startswith" ? false : true
        setReverseChart(isReverse)
        setChartData(transformDataFn(sankeyData, isReverse)) 
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
    const transformDataFn = (data, isReverse) => { 
        console.log('input chart data-->', data); 
        if (data) {
            let results = data;
            let finalArr = [];
            let final = [];
            let title = '';
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
            console.log('transformed chart data-->', finalArr); 
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
            height: chartData ? (chartData.length < 75 ? 300 :  chartData.length* 4) : 400,
        },

        plotOptions: {
            sankey: {
                nodeWidth: 60,
                colorByPoint: false,
                curveFactor: 0.5,
                nodePadding: 40,
                borderRadius: 4,
                centerInCategory: true,
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
        // Vishnu use this tooltip key to render tooltip
        tooltip: {
            backgroundColor: 'red',
            borderWidth: 0,
            borderRadius: 2,
            shadow: 'none',
            borderColor: "transparent",
            useHTML: true,
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
                keys: reverseChart ? ['to', 'from', 'weight'] : ['from', 'to', 'weight'],
                // keys: ['from', 'to', 'weight'],
                data: chartData ? chartData : [],
                // data: transformDataFn(sankeyData),
                type: 'sankey',
                tooltip: {
                    headerFormat: undefined,
                    enabled: false,
                    outside: true,
                    className: 'custom-div',
                    nodeFormat: undefined,
                    // pointFormat:'{point.series.name} 22→ {point.toNode.name}: <b>{point.options.weight}</b>',
                    // shared: true,
                    shadow: false,
                    useHTML: true,
                    pointFormatter: function () {
                        return ReactDOMServer.renderToString(<CustomTooltip />);
                    },
                },
            },
        ],
    };

 
    const drawChart = useCallback(() => { 
        Highcharts.chart("fa-sankey-container", {
            ...options
        })
    });

    useEffect(() => {
        drawChart();
      }, [reverseChart, chartData ]);

    return (
        <> 
                    <div className='mt-16 mb-10'>

                        {/* {StepArraowGenerator(activeQuery)} */}


                        {/* <HighchartsReact
                            highcharts={Highcharts}
                            options={options}
                        /> */}

                    <div className='fa-sankey-container' id="fa-sankey-container" />


                    </div> 
        </>
    );
}

export default Sankey;
