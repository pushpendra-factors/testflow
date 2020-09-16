import React, { useState } from 'react';
import FunnelsResultPage from './FunnelsResultPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button } from 'antd';
import { SVG, Text } from '../../components/factorsComponents';
import EventsAnalytics from '../EventsAnalytics';

function CoreQuery() {
    const [drawerVisible, setDrawerVisible] = useState(false);
    const [queryType, setQueryType] = useState('event');
    const [showResult, setShowResult] = useState(false);
    // const [showResult, setShowResult] = useState(true);
    const [queries, setQueries] = useState([]);
    // const [queries, setQueries] = useState(['Applied Coupon', 'Cart Updated', 'Checkout', 'Add to Wishlist', 'Paid']);

    const [showFunnels, setShowFunnels] = useState(true);

    const queryChange = (newEvent, index, changeType = 'add') => {
        const queryupdated = [...queries];
        if (queryupdated[index]) {
            if (changeType === 'add') {
                queryupdated[index] = newEvent;
            } else {
                queryupdated.splice(index, 1);
            }

        } else {
            queryupdated.push(newEvent);
        }
        setQueries(queryupdated);
    }

    const runQuery = () => {
        setShowResult(true);
        closeDrawer();
    }

    const closeDrawer = () => {
        setDrawerVisible(false);
    }

    const title = () => {
        return (<div className={`flex justify-between items-center`}>
            <div className={`flex`}>
                <SVG name="teamfeed"></SVG>
                <Text type={'title'} level={4} weight={`bold`} extraClass={`ml-2 m-0`}>Find event funnel for</Text>
            </div>
            <div className={`flex justify-end items-center`}>
                <Button type="text"><SVG name="play"></SVG>Help</Button>
                <Button type="text" onClick={() => closeDrawer()}><SVG name="times"></SVG></Button>
            </div>

        </div>)
    }

    let result = (<EventsAnalytics showFunnels={showFunnels} setShowFunnels={setShowFunnels} queries={queries.map(elem => elem.label)} />);
    // let result = (<EventsAnalytics showFunnels={showFunnels} setShowFunnels={setShowFunnels} queries={queries.map(elem => elem)} />);

    if (showFunnels) {
        result = (<FunnelsResultPage showFunnels={showFunnels} setShowFunnels={setShowFunnels} setDrawerVisible={setDrawerVisible} queries={queries.map(elem => elem.label)} />);
        // result = (<FunnelsResultPage showFunnels={showFunnels} setShowFunnels={setShowFunnels} setDrawerVisible={setDrawerVisible} queries={queries.map(elem => elem)} />);
    }

    return (
        <>
            <Drawer
                title={title()}
                placement="left"
                closable={false}
                visible={drawerVisible}
                onClose={closeDrawer}
                getContainer={false}
                width={"600px"}
                className={`fa-drawer`}
            >

                <QueryComposer
                    queries={queries}
                    runQuery={runQuery}
                    eventChange={queryChange}
                    queryType={queryType}
                />
            </Drawer>

            {showResult ? (
                <>
                    {result}
                </>

            ) : (
                    <CoreQueryHome setDrawerVisible={setDrawerVisible} />
                )}

        </>
    )
}

export default CoreQuery;