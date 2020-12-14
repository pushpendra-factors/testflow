/* eslint-disable */
import React, { useCallback, useState, useEffect } from "react";
import { Text, SVG } from "../../components/factorsComponents";
import { Row, Col, Table, Avatar, Button, Dropdown, Menu } from "antd";
import { MoreOutlined } from "@ant-design/icons";
import Header from "../AppLayout/Header";
import SearchBar from "../../components/SearchBar";
import { useSelector, useDispatch } from "react-redux";
import moment from "moment";
import { getStateQueryFromRequestQuery } from "../CoreQuery/utils";
import { INITIALIZE_GROUPBY } from "../../reducers/coreQuery/actions";
import ConfirmationModal from "../../components/ConfirmationModal";
import { deleteQuery } from "../../reducers/coreQuery/services";
import { typeOf } from "react-is";
import {
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
} from "../../utils/constants";
import { SHOW_ANALYTICS_RESULT } from "../../reducers/types";

const coreQueryoptions = [
  {
    title: "Events",
    icon: "events_cq",
    desc: "Create charts from events and related properties",
  },
  {
    title: "Funnels",
    icon: "funnels_cq",
    desc: "Find how users are navigating a defined path",
  },
  {
    title: "Campaigns",
    icon: "campaigns_cq",
    desc: "Find the effect of your marketing campaigns",
  },
  {
    title: "Attributions",
    icon: "attributions_cq",
    desc: "Analyse Multi Touch Attributions",
  },
  {
    title: "Templates",
    icon: "templates_cq",
    desc: "A list of advanced queries crafter by experts",
  },
];

const columns = [
  {
    title: "Type",
    dataIndex: "type",
    width: 60,
    key: "type",
  },
  {
    title: "Title of the Query",
    dataIndex: "title",
    key: "title",
  },
  {
    title: "Created By",
    dataIndex: "author",
    key: "author",
    render: (text) => (
      <div className="flex items-center">
        <Avatar src="assets/avatar/avatar.png" className={"mr-2"} />
        &nbsp; {text}{" "}
      </div>
    ),
  },
  {
    title: "Date",
    dataIndex: "date",
    key: "date",
  },
];

function CoreQuery({
  setDrawerVisible,
  setQueryType,
  setQueries,
  setRowClicked,
  setQueryOptions,
  location,
}) {
  const queriesState = useSelector((state) => state.queries);
  const [deleteModal, showDeleteModal] = useState(false);
  const [activeRow, setActiveRow] = useState(null);
  const dispatch = useDispatch();

  const getFormattedRow = (q) => {
    let svgName = "funnels_cq";
    let requestQuery = q.query;
    if (requestQuery.query_group) {
      svgName = "events_cq";
    }

    return {
      key: q.id,
      type: <SVG name={svgName} size={24} />,
      title: q.title,
      author: q.created_by_name,
      date: (
        <div className="flex justify-between items-center">
          <div>{moment(q.created_at).format("MMM DD, YYYY")}</div>
          <div>
            <Dropdown overlay={getMenu(q)} trigger={["hover"]}>
              <Button type="text" icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        </div>
      ),
      query: requestQuery,
      actions: "",
    };
  };

  const confirmDelete = useCallback(() => {
    deleteQuery(dispatch, activeRow);
    setActiveRow(null);
    showDeleteModal(false);
  }, [activeRow]);

  const handleDelete = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setActiveRow(row);
    showDeleteModal(true);
  }, []);

  const handleViewResult = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setQueryToState(getFormattedRow(row));
  }, []);

  const setQueryToState = useCallback((record) => {
    let equivalentQuery;
    if (record.query.query_group) {
      equivalentQuery = getStateQueryFromRequestQuery(
        record.query.query_group[0]
      );
    } else {
      equivalentQuery = getStateQueryFromRequestQuery(record.query);
    }
    dispatch({ type: INITIALIZE_GROUPBY, payload: equivalentQuery.breakdown });
    setQueries(equivalentQuery.events);
    setQueryType(equivalentQuery.queryType);
    setQueryOptions((currData) => {
      return {
        ...currData,
        groupBy: [
          ...equivalentQuery.breakdown.global,
          ...equivalentQuery.breakdown.event,
        ],
      };
    });
    setRowClicked(equivalentQuery.queryType);
  }, []);

  const getMenu = (row) => {
    return (
      <Menu>
        <Menu.Item key="0">
          <a onClick={handleViewResult.bind(this, row)} href="#!">
            View Results
          </a>
        </Menu.Item>
        <Menu.Item key="1">
          <a onClick={(e) => e.stopPropagation()} href="#!">
            Copy Link
          </a>
        </Menu.Item>
        <Menu.Item key="2">
          <a onClick={handleDelete.bind(this, row)} href="#!">
            Delete Query
          </a>
        </Menu.Item>
      </Menu>
    );
  };

  useEffect(() => {
    if (location.state && location.state.global_search) {
      setQueryToState(location.state.query);
      location.state = undefined;
    } else {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    }
  }, [location.state, setQueryToState]);

  const data = queriesState.data
    .filter((q) => parseInt(q.type) === 2)
    .map((q) => {
      return getFormattedRow(q);
    });

  const setQueryTypeTab = (item) => {
    setDrawerVisible(true);

    if (item.title === "Funnels") {
      setQueryType(QUERY_TYPE_FUNNEL);
    }

    if (item.title === "Events") {
      setQueryType(QUERY_TYPE_EVENT);
    }

    if (item.title === "Attributions") {
      setQueryType(QUERY_TYPE_ATTRIBUTION);
    }
  };

  return (
    <>
      <ConfirmationModal
        visible={deleteModal}
        confirmationText="Are you sure you want to delete this query?"
        onOk={confirmDelete}
        onCancel={showDeleteModal.bind(this, false)}
        title="Delete Query"
        okText="Confirm"
        cancelText="Cancel"
      />
      <Header>
        <div className="w-full h-full py-4 flex flex-col justify-center items-center">
          <SearchBar setQueryToState={setQueryToState} />
        </div>
      </Header>
      <div className={"fa-container mt-24"}>
        <Row gutter={[24, 24]} justify="center">
          <Col span={20}>
            <Text type={"title"} level={2} weight={"bold"} extraClass={"m-0"}>
              Core Query
            </Text>
            <Text
              type={"title"}
              level={5}
              weight={"regular"}
              color={"grey"}
              extraClass={"m-0"}
            >
              Use these tools to Analyse and get to the bottom of User Behaviors
              and Marketing Funnels
            </Text>
          </Col>
        </Row>
        <Row gutter={[24, 24]} justify="center" className={"mt-10"}>
          {coreQueryoptions.map((item, index) => {
            return (
              <Col span={4} key={index}>
                <div onClick={() => {
                  setDrawerVisible(true);
                  switch(item.title){
                    case 'Funnels' : setQueryType('funnel'); break;
                    case 'Events' : setQueryType('event'); break;
                    case 'Campaigns' : setQueryType('campaigns'); break;
                    case 'Attributions' : setQueryType('attributions'); break;
                    case 'Templates' : setQueryType('templates'); break;
                    default: setQueryType('funnel'); break;
                  }
                  // item.title === 'Funnels' ? setQueryType('funnel') : setQueryType('event');
                }} className="fai--custom-card flex justify-center items-center flex-col ">
                  <div className={'fai--custom-card--icon'}><SVG name={item.icon} size={48} /> </div>
                  <div className="flex justify-start items-center flex-col before-hover">
                    <Text
                      type={"title"}
                      level={3}
                      weight={"bold"}
                      extraClass={"fai--custom-card--title"}
                    >
                      {item.title}
                    </Text>
                  </div>
                  <div className="flex justify-start items-center flex-col after-hover">
                    <div
                      className={
                        "fai--custom-card--content flex-col flex justify-start items-center"
                      }
                    >
                      <Text
                        type={"title"}
                        level={7}
                        weight={"bold"}
                        extraClass={"fai--custom-card--desc"}
                      >
                        {item.desc}
                      </Text>
                      <a className={"fai--custom-card--cta"}>
                        New Query <SVG name={"next"} size={20} />{" "}
                      </a>
                    </div>
                  </div>
                </div>
              </Col>
            );
          })}
        </Row>

        <Row justify="center" className={"mt-12"}>
          <Col span={20}>
            <Row className={"flex justify-between items-center"}>
              <Col span={10}>
                <Text
                  type={"title"}
                  level={4}
                  weight={"bold"}
                  extraClass={"m-0"}
                >
                  Saved Queries
                </Text>
              </Col>
              <Col span={5}>
                <div className={"flex flex-row justify-end items-end "}>
                  <Button
                    icon={<SVG name={"help"} size={12} color={"grey"} />}
                    type="text"
                  >
                    Learn More
                  </Button>
                </div>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row justify="center" className={"mt-2 mb-20"}>
          <Col span={20}>
            <Table
              onRow={(record) => {
                return {
                  onClick: (e) => {
                    setQueryToState(record);
                  },
                };
              }}
              loading={queriesState.loading}
              className="ant-table--custom"
              columns={columns}
              dataSource={data}
              pagination={false}
              rowClassName="cursor-pointer"
            />
          </Col>
        </Row>
      </div>
    </>
  );
}

export default CoreQuery;
