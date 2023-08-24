import React, { useEffect, useState, useRef, useCallback } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, notification, Tooltip, Input, Row, Col, Tag } from 'antd';
import { useSelector } from 'react-redux';
import { SVG, Text } from 'Components/factorsComponents';
import { udpateProjectSettings, fetchProjectSettings } from 'Reducers/global';
import { FilterIps } from './types';

const ExcludeIpBlock = ({
  udpateProjectSettings,
  fetchProjectSettings
}: ExcludeIpBlockProps) => {
  const activeProject = useSelector((state) => state.global.active_project);
  const currentProjectSettings = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const [filterIps, setFilterIps] = useState<FilterIps | null>(null);
  const [ipToExclude, setIpToExclude] = useState<string>('');
  const [errorState, setErrorState] = useState<string | undefined>(undefined);
  const [editMode, setEditMode] = useState<boolean>(false);

  useEffect(() => {
    if (currentProjectSettings) {
      const excludedIpList = currentProjectSettings?.filter_ips?.block_ips;
      setFilterIps(new FilterIps(excludedIpList));
    }
  }, [currentProjectSettings]);

  const renderIpList = () => {
    const chunks = filterIps?.getFilterIpsByChunks(6);
    return (
      <div>
        {chunks?.map((ipChunk, index) => {
          if (index > 3 && !editMode) return null;
          return (
            <>
              <Row>
                {ipChunk.map((ip, ipindex) => {
                  const lastItem =
                    index === chunks.length - 1 &&
                    ipindex === ipChunk.length - 1;
                  return !editMode
                    ? renderIpText(ip, lastItem)
                    : renderIpTag(ip);
                })}
              </Row>
              {index === 3 && !editMode ? (
                <Text
                  type={'title'}
                  level={6}
                  extraClass={'mr-1'}
                  weight={'bold'}
                >
                  ...
                </Text>
              ) : null}
            </>
          );
        })}
      </div>
    );
  };

  const SaveSettings = useCallback(() => {
    udpateProjectSettings(activeProject.id, {
      filter_ips: filterIps?.getFilterIpPayload()
    }).then(() => {
      fetchProjectSettings(activeProject.id);
    });
  }, [activeProject?.id, filterIps]);

  const blockIpandSaveSettings = () => {
    const setIpResponse = filterIps?.setIp(ipToExclude);
    if (setIpResponse === true) {
      SaveSettings();
    } else {
      setErrorState(setIpResponse);
    }
  };

  const renderExcludeInput = () => {
    return (
      <Row className='mb-4'>
        <Col className={`flex items-center`}>
          <Input
            placeholder='Enter IP Address'
            onChange={(ev) => setIpToExclude(ev.target.value)}
          ></Input>
          <Button className='ml-2' onClick={() => blockIpandSaveSettings()}>
            Block
          </Button>
        </Col>
      </Row>
    );
  };

  const renderIpTag = (ip: String) => {
    return (
      <Tag closable={true} onClose={() => filterIps?.removeIp(ip)}>
        {ip}
      </Tag>
    );
  };

  const renderIpText = (ip: String, lastItem: boolean = false) => {
    return (
      <Text type={'title'} level={7} extraClass={'mr-1'} weight={'bold'}>
        {ip}
        {lastItem ? null : ', '}
      </Text>
    );
  };

  return (
    <>
      <Row className='flex justify-between items-start'>
        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mb-2'}>
          Blocked IPs
        </Text>
        {editMode ? (
          <Button
            type='primary'
            size='large'
            icon={<SVG name='edit' size='20' color='white' />}
            onClick={() => {
              setEditMode(false);
              SaveSettings();
            }}
          >
            Save
          </Button>
        ) : (
          <Button
            type='default'
            size='large'
            icon={<SVG name='edit' size='20' color='grey' />}
            onClick={() => {
              setEditMode(true);
            }}
          >
            Edit
          </Button>
        )}
      </Row>
      {editMode ? renderExcludeInput() : null}
      <Row>{renderIpList()}</Row>
    </>
  );
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

type ExcludeIpBlockProps = ReduxProps;

export default connector(ExcludeIpBlock);
