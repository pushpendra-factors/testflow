import React from "react";
import { Text, SVG } from "../../../components/factorsComponents";
import WebsiteAnalyticsTable from "./WebsiteAnalyticsTable";

function TableUnits({ tableUnits, data }) {
  return (
    <>
      {tableUnits.map((unit) => {
        if (data[unit.id]) {
          return (
            <div
              key={unit.id}
              className={`py-4 px-2 flex widget-card-top-div w-full`}
            >
              <div
                id={`card-${unit.id}`}
                className={"fa-dashboard--widget-card w-full flex"}
              >
                <div
                  className={
                    "px-8 py-4 flex justify-between items-start w-full"
                  }
                >
                  <div className={"w-full"}>
                    <div className="flex items-center justify-between">
                      <div className="flex flex-col">
                        <div
                          className="flex cursor-pointer items-center"
                          // onClick={() =>
                          //   setwidgetModal({ unit, data: resultState.data })
                          // }
                        >
                          <Text
                            ellipsis
                            type={"title"}
                            level={5}
                            weight={"bold"}
                            extraClass={"m-0 mr-1"}
                          >
                            {unit.title}
                          </Text>
                          {/* <SVG color="#8692A3" size={20} name="expand" /> */}
                        </div>
                        <div>
                          <Text
                            ellipsis
                            type={"paragraph"}
                            mini
                            color={"grey"}
                            extraClass={"m-0"}
                          >
                            {unit.description}
                          </Text>
                        </div>
                      </div>
                      {/* <div>
                        <Dropdown overlay={getMenu()} trigger={["hover"]}>
                          <Button type="text" icon={<MoreOutlined />} />
                        </Dropdown>
                      </div> */}
                    </div>
                    <div className="mt-4">
											<WebsiteAnalyticsTable unit={unit} tableData={data[unit.id]} />
										</div>
                  </div>
                </div>
              </div>
              {/* <div
                id={`resize-${unit.id}`}
                className={"fa-widget-card--resize-container"}
              >
                <span className={"fa-widget-card--resize-contents"}>
                  {unit.cardSize === 0 ? (
                    <a href="#!" onClick={changeCardSize.bind(this, 1)}>
                      <RightOutlined />
                    </a>
                  ) : null}
                  {unit.cardSize === 1 ? (
                    <a href="#!" onClick={changeCardSize.bind(this, 0)}>
                      <LeftOutlined />
                    </a>
                  ) : null}
                </span>
              </div> */}
            </div>
          );
        } else {
          return null;
        }
      })}
    </>
  );
}

export default TableUnits;
