import React, { useEffect, useState, useRef } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, notification, Tooltip, Input, Row, Col, Tag } from 'antd';
import { useSelector } from 'react-redux';
import { Text } from 'Components/factorsComponents';
import { udpateProjectSettings, fetchProjectSettings } from 'Reducers/global';
import { FilterIps } from './types';

const ExcludeIpBlock = ({
    mode = 'view',
    udpateProjectSettings,
    fetchProjectSettings
}: ExcludeIpBlockProps) => {

    const activeProject = useSelector((state) => state.global.active_project);
    const currentProjectSettings = useSelector((state) => state.global.currentProjectSettings);
    const [filterIps, setFilterIps] = useState<FilterIps | null>(null)
    const [ipToExclude, setIpToExclude] = useState<string>('');
    const [errorState, setErrorState] = useState<string | undefined>(undefined);

    useEffect(() => {
        if (currentProjectSettings) {
            const excludedIpList = currentProjectSettings?.filter_ips?.block_ips
            setFilterIps(new FilterIps(excludedIpList));
        }
    }, [currentProjectSettings])

    const renderIpList = () => {
        const chunks = filterIps?.getFilterIpsByChunks(6);
        return (<div>
            {chunks?.map((ipChunk, index) => {
                if (index > 3 && mode === 'view') return;
                return (<>
                    <Row>{ipChunk.map((ip, ipindex) => {
                        const lastItem = index === chunks.length - 1 && ipindex === ipChunk.length - 1;
                        return (
                            mode === 'view' ? renderIpText(ip, lastItem) : renderIpTag(ip)
                        );
                    })}</Row>
                    {index === 3 && mode === 'view' ? (<Text type={'title'} level={6} extraClass={'mr-1'} weight={'bold'}>...</Text>) : null}
                </>);
            })}
        </div>);
    }

    const blockIpandSaveSettings = () => {
        const setIpResponse = filterIps?.setIp(ipToExclude);
        if (setIpResponse === true) {
            udpateProjectSettings(activeProject.id, { filter_ips: filterIps?.getFilterIpPayload() }).then((res) => {
                fetchProjectSettings(activeProject.id);
            })
        } else {
            setErrorState(setIpResponse);
        }
    }


    const renderExcludeInput = () => {
        return (<Row className={`mb-2`}>
            <Col span={4} className={`mr-2`}>
                <Text level={6} type={'paragraph'} color={'grey'}>
                    IP Address equals
                </Text>
            </Col>
            <Col span={6} className={`mr-1`}>
                <Input onChange={(ev) => setIpToExclude(ev.target.value)}></Input>
            </Col>
            <Col span={6}>
                <Button onClick={() => blockIpandSaveSettings()}> Exclude</Button>
            </Col>

        </Row>)
    }

    const renderIpTag = (ip: String) => {
        return (<Tag closable={true} onClose={() => filterIps?.removeIp(ip)}>{ip}</Tag>)
    }

    const renderIpText = (ip: String, lastItem: boolean = false) => {
        return (
            <Text type={'title'} level={6} extraClass={'mr-1'} weight={'bold'}>{ip}{lastItem ? null : ', '}</Text>
        )
    }

    return (<>
        <Row>
            <Text type={'title'} level={7} extraClass={'mb-2'}>
                Internal Traffic Blocking
            </Text>
        </Row>
        {mode === 'edit' ? renderExcludeInput() : null}
        <Row>
            {renderIpList()}
        </Row>

    </>)
};

const mapDispatchToProps = (dispatch) =>
    bindActionCreators(
        {
            udpateProjectSettings,
            fetchProjectSettings
        },
        dispatch
    );
const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type BlockStates = 'view' | 'edit';
type ExcludeIpBlockParams = {
    mode: BlockStates;
};

type ExcludeIpBlockProps = ExcludeIpBlockParams & ReduxProps;

export default connector(ExcludeIpBlock);
