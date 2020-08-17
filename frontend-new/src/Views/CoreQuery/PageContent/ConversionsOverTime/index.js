import React, { useRef, useEffect, useCallback } from 'react';
import styles from './index.module.scss';
import c3 from 'c3';
import * as d3 from 'd3';

function ConversionsOverTime() {

    const chartRef = useRef(null);

    const categories = [
        {
            name: 'Chennai',
            conversion_rate: "5%"
        },
        {
            name: 'Mumbai',
            conversion_rate: "10%"
        },
        {
            name: 'New Delhi',
            conversion_rate: "8%"
        },
        {
            name: 'Amritsar',
            conversion_rate: "20%"
        },
        {
            name: 'Jalandhar',
            conversion_rate: "2%"
        }
    ]

    const showConverionRates = useCallback(() => {
        let yGridLines = d3.select(chartRef.current).select('g.c3-ygrids').selectAll('line').nodes();
    }, []);

    const showVerticalGridLines = useCallback(() => {
        let yGridLines = d3.select(chartRef.current).select('g.c3-ygrids').selectAll('line').nodes();
        let top, bottom, height;
        let topGridLine = yGridLines[yGridLines.length - 1];
        top = topGridLine.getBoundingClientRect().y;
        let bottomGridLine = yGridLines[0];
        bottom = bottomGridLine.getBoundingClientRect().y;
        height = bottom - top;
        d3.select(chartRef.current).select('g.c3-shapes-Fifth-Event').selectAll('path').nodes()
            .forEach((elem, index) => {
                let position = elem.getBoundingClientRect();
                let verticalLine = document.getElementById(categories[index].name);
                verticalLine.style.left = `${position.x + position.width - 1}px`;
                verticalLine.style.height = `${height}px`;
                verticalLine.style.top = `${top}px`;
            });
        showConverionRates();
    }, [categories, showConverionRates]);

    useEffect(() => {

        setTimeout(() => {
            c3.generate({
                size: {
                    height: 350,
                },
                bindto: chartRef.current,
                data: {
                    columns: [
                        ['First Event', 100, 100, 100, 100, 100],
                        ['Second Event', 70, 50, 60, 50, 60],
                        ['Third Event', 30, 40, 30, 40, 40],
                        ['Fouth Event', 10, 15, 15, 30, 20],
                        ['Fifth Event', 5, 10, 8, 20, 2]
                    ],
                    type: 'bar',
                    colors: {
                        'First Event': '#014694',
                        'Second Event': '#008BAE',
                        'Third Event': '#52C07C',
                        'Fouth Event': '#F1C859',
                        'Fifth Event': '#DE7542'
                    },
                },
                transition: {
                    duration: 1000
                },
                onrendered: showVerticalGridLines,
                bar: {
                    space: 0.03,
                },
                axis: {
                    x: {
                        type: 'category',
                        categories: categories.map(elem => elem.name),
                    },
                    y: {
                        max: 100,
                        tick: {
                            values: [0, 20, 40, 60, 80, 100],
                            format: (d) => {
                                if (d) {
                                    return d + '%';
                                } else {
                                    return d
                                }

                            },
                        }
                    }
                },
                tooltip: {
                    grouped: false,
                },
                grid: {
                    y: {
                        show: true,
                    },
                },
            });
        }, 0)
    }, [categories, showVerticalGridLines]);

    let first = 19;

    return (
        <>
            {
                categories.map(elem => {
                    return (
                        <div className={`absolute border-l border-solid ${styles.verticalGridLines}`} key={elem.name} id={elem.name}></div>
                    );
                })
            }
            {/* {
                categories.map((elem, index) => {
                    return (
                        <div style={{top: "11.5rem", lineHeight: "1.25rem", left: 17 + (19*index) + 'rem'}} className="absolute text-lg">
                            <div className="font-bold flex justify-end">{elem.conversion_rate}</div>
                            <div>Conversion</div>
                        </div>
                    );
                })
            } */}
            <div className={styles.conversionsOverTimeChart} ref={chartRef} />
        </>
    )
}

export default ConversionsOverTime;