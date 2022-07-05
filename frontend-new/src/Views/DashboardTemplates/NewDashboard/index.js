import React, { useState, useCallback, useEffect } from "react";
import { Row, Col, Tabs, Modal, notification, Input, Checkbox} from "antd";
import { useSelector, useDispatch } from "react-redux";
import { Text ,SVG} from "../../../components/factorsComponents";
import styles from "../index.module.scss";
import AddReportsTab from "./AddReportsTab";
import AddDashboardDetails from "./AddDashboardDetails";


function NewDashboard({
    AddDashboardDetailsVisible,
    setAddDashboardDetailsVisible,
    AddReportsVisible,
    setAddReportsVisible
}){
    const queries = [];
    for(let k=1;k<19;k++){
      queries.push({id:k,title:`${k}-Submission Form`});
    }
    // const [title, setTitle] = useState("");
    // const [description, setDescription] = useState("");
    const [selectedQueries, setSelectedQueries] = useState([]);
    return(
        <>
            <AddDashboardDetails 
                AddDashboardDetailsVisible={AddDashboardDetailsVisible}
                setAddDashboardDetailsVisible={setAddDashboardDetailsVisible}
                AddReportsVisible={AddReportsVisible}
                setAddReportsVisible={setAddReportsVisible}
            />
            <AddReportsTab
                AddReportsVisible={AddReportsVisible}
                setAddReportsVisible={setAddReportsVisible}
                selectedQueries={selectedQueries}
                setSelectedQueries={setSelectedQueries} 
                queries={queries}                               
            />
        </>
    );
}

export default NewDashboard;