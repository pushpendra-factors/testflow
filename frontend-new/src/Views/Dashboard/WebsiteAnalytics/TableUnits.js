import React from "react";
import { Text, SVG } from "../../../components/factorsComponents";
import WebsiteAnalyticsTable from "./WebsiteAnalyticsTable";

function TableUnits({ tableUnits, data, setwidgetModal, resultState }) {
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
                <div className={"py-5 flex justify-between items-start w-full"}>
                  <div className={"w-full flex flex-1 flex-col h-full"}>
                    <div
                      style={{
                        borderBottom: "1px solid rgb(231, 233, 237)",
                      }}
                      className="flex items-center justify-between px-6 pb-4"
                    >
                      <div className="flex flex-col">
                        <div
                          className="flex cursor-pointer items-center"
                          onClick={() =>
                            setwidgetModal({ unit, data: resultState.data })
                          }
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
                          <SVG color="#8692A3" size={20} name="expand" />
                        </div>
                        {/* <div>
                          <Text
                            ellipsis
                            type={"paragraph"}
                            mini
                            color={"grey"}
                            extraClass={"m-0"}
                          >
                            {unit.description}
                          </Text>
                        </div> */}
                      </div>
                    </div>
                    <div
                      className={`w-full px-6 flex flex-1 flex-col  justify-center`}
                    >
                      <WebsiteAnalyticsTable
                        title={unit.title}
                        tableData={data[unit.id]}
                        isWidgetModal={false}
                      />
                    </div>
                  </div>
                </div>
              </div>
              );
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
