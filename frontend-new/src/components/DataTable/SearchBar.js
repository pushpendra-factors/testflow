import React from "react";
import styles from "./index.module.scss";
import { Input, Button } from "antd";
import { SVG } from "../factorsComponents";
import { CSVLink } from "react-csv";

function SearchBar({
  searchText,
  handleSearchTextChange,
  searchBar,
  getCSVData,
  downloadBtnRef,
}) {
  let csvData = { data: [], fileName: "data" };

  if (getCSVData) {
    csvData = getCSVData();
  }

  const downloadBtn = (
    <div
      ref={downloadBtnRef}
      className="flex flex-1 items-center justify-end cursor-pointer"
    >
      <Button
        size={"large"}
        style={{ display: "flex", alignItems: "center" }}
        icon={<SVG name={"download"} size={24} color={"grey"} />}
        type="text"
      >
        <CSVLink
          id="csvLink"
          style={{ color: "#0E2647" }}
          onClick={() => {
            if (!csvData.data.length) return false;
          }}
          filename={csvData.fileName}
          data={csvData.data}
        >
          Download CSV
        </CSVLink>
      </Button>
    </div>
  );

  return (
    <div className={`${styles.searchBar}`}>
      {!searchBar ? (
        <div className="flex pb-1 w-full">
          <div className={"flex items-center w-3/4 cursor-pointer"}>
            <div className="mr-2">
              <svg
                width="17"
                height="18"
                viewBox="0 0 17 18"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fillRule="evenodd"
                  clipRule="evenodd"
                  d="M13.6661 9.22917C13.4267 9.97472 13.0656 10.6658 12.6064 11.279L16.7071 15.3797C17.0976 15.7703 17.0976 16.4034 16.7071 16.794C16.3166 17.1845 15.6834 17.1845 15.2929 16.794L11.1921 12.6932C10.0236 13.5684 8.57232 14.0868 7 14.0868C3.13401 14.0868 0 10.9528 0 7.08679C0 3.2208 3.13401 0.086792 7 0.086792C8.6281 0.086792 10.1264 0.642616 11.3154 1.57486C12.9498 2.85628 14 4.84889 14 7.08679C14 7.83405 13.8829 8.55396 13.6661 9.22917ZM11.7397 8.68317C11.0736 10.6618 9.20321 12.0868 7 12.0868C4.23858 12.0868 2 9.84822 2 7.08679C2 4.32537 4.23858 2.08679 7 2.08679C8.23286 2.08679 9.36151 2.533 10.2333 3.27273C11.3141 4.18987 12 5.55823 12 7.08679C12 7.64501 11.9085 8.18186 11.7397 8.68317Z"
                  fill="black"
                />
              </svg>
            </div>
            <div className={styles.breakupHeading}>Break-up</div>
          </div>
          {downloadBtn}
        </div>
      ) : (
        <Input
          onChange={(e) => handleSearchTextChange(e.target.value)}
          value={searchText}
          className={`${styles.inputSearchBar} ${
            !searchText.length
              ? styles.inputPlaceHolderFont
              : styles.inputTextFont
          }`}
          size="large"
          placeholder="Search"
          prefix={
            <svg
              width="17"
              height="18"
              viewBox="0 0 17 18"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M13.6661 9.22917C13.4267 9.97472 13.0656 10.6658 12.6064 11.279L16.7071 15.3797C17.0976 15.7703 17.0976 16.4034 16.7071 16.794C16.3166 17.1845 15.6834 17.1845 15.2929 16.794L11.1921 12.6932C10.0236 13.5684 8.57232 14.0868 7 14.0868C3.13401 14.0868 0 10.9528 0 7.08679C0 3.2208 3.13401 0.086792 7 0.086792C8.6281 0.086792 10.1264 0.642616 11.3154 1.57486C12.9498 2.85628 14 4.84889 14 7.08679C14 7.83405 13.8829 8.55396 13.6661 9.22917ZM11.7397 8.68317C11.0736 10.6618 9.20321 12.0868 7 12.0868C4.23858 12.0868 2 9.84822 2 7.08679C2 4.32537 4.23858 2.08679 7 2.08679C8.23286 2.08679 9.36151 2.533 10.2333 3.27273C11.3141 4.18987 12 5.55823 12 7.08679C12 7.64501 11.9085 8.18186 11.7397 8.68317Z"
                fill="black"
              />
            </svg>
          }
          suffix={downloadBtn}
        />
      )}
    </div>
  );
}

export default SearchBar;
