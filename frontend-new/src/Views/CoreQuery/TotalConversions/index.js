import React, { useEffect, useRef } from 'react';
import c3 from 'c3';
import styles from './index.module.scss';

function TotalConversions() {

    const chartRef = useRef(null);

    const categories = ['Google', 'Facebook', 'G2', 'Capterra', 'Email'];

    useEffect(() => {
        c3.generate({
            bindto: chartRef.current,
            data: {
                columns: [
                    ['First Touch', 50, 140, 230, 20, 210],
                    ['Last Touch', 130, 120, 110, 100, 90],
                    ['Max Engaged', 100, 90, 80, 70, 60],
                    ['Min Engaged', 120, 80, 40, 40, 63],
                ],
                type: 'bar',
                colors: {
                    'First Touch': '#014694',
                    'Last Touch': '#009FA1',
                    'Max Engaged': '#96CC6E',
                    'Min Engaged': '#008FA1',
                },
            },
            transition: {
                duration: 1000
            },
            // onrendered: showAnnotation,
            bar: {
                width: {
                    ratio: 0.3,
                },
                zerobased: false,
            },
            axis: {
                x: {
                    type: 'category',
                    categories,
                },
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
    }, [categories]);

    return (
        <div className={styles.totalConversionsChart} style={{ margin: '0.25rem' }} ref={chartRef} />
    )
}

export default TotalConversions;