import React, { useCallback, useRef, useEffect } from 'react';
import c3 from 'c3';
import * as d3 from 'd3';
import moment from 'moment';
import styles from './index.module.scss';

function LineChart({ chartData, appliedColors, queries }) {

    const chartRef = useRef(null);

    const xAxisValues = chartData.find(elem => elem[0] === 'x').slice(1);

    let xAxisCount = Math.ceil(xAxisValues.length/2);

    if(xAxisCount > 10) {
        xAxisCount = 10;
    }

    const modVal = Math.ceil(xAxisValues.length / xAxisCount);

    const finalXaxisValues = [];
    let j = 1;

    for (let i = 0; i < xAxisCount && j < xAxisValues.length; i++) {
        finalXaxisValues.push(xAxisValues[1 + i * modVal]);
        j = 1 + (i + 1) * modVal;
    }

    const colors = {};

    queries.forEach((query, index)=>{
        colors[query] = appliedColors[index];
    })

    const drawChart = useCallback(() => {
        c3.generate({
            bindto: chartRef.current,
            size: {
                height: 350
            },
            padding: {
                left: 50,
                bottom: 24,
                right: 10
            },
            data: {
                x: 'x',
                columns: chartData,
                colors,
                onmouseover: (d) => {
                    d3.select(chartRef.current)
                        .selectAll('.c3-chart-line.c3-target')
                        .nodes()
                        .forEach(node => {
                            node.classList.add('c3-defocused')
                        })
                    d3.select(chartRef.current)
                        .select(`.c3-chart-line.c3-target.c3-target-${d.name.split(" ").join('-')}`)
                        .nodes()
                        .forEach(node => {
                            node.classList.remove('c3-defocused');
                            node.classList.add('c3-focused');
                        })
                },
                onmouseout: () => {
                    d3.select(chartRef.current)
                        .selectAll('.c3-chart-line.c3-target')
                        .nodes()
                        .forEach(node => {
                            node.classList.remove('c3-defocused')
                            node.classList.remove('c3-focused')
                        })
                },
            },
            axis: {
                x: {
                    type: 'timeseries',
                    tick: {
                        values: finalXaxisValues,
                        format: (d) => {
                            return moment(d).format('MMM D')
                        }
                    }
                },
                y: {
                    tick: {
                        count: 6,
                        format: function (d) {
                            return parseInt(d);
                        }
                    },
                }
            },
            onrendered: () => {
                d3.select(chartRef.current).select('.c3-axis.c3-axis-x').selectAll('.tick').select('tspan').attr('dy', '16px');
            },
            legend: {
                show:false
            },
            grid: {
                y: {
                    show: true
                }
            },
            tooltip: {
                grouped: false
            }
        });
    }, [chartData, finalXaxisValues, colors]);

    const displayChart = useCallback(() => {
        drawChart();
    }, [drawChart]);


    useEffect(() => {
        displayChart();
    }, [displayChart]);

    return (
        <div className={styles.lineChart} ref={chartRef} />
    )
}

export default LineChart;