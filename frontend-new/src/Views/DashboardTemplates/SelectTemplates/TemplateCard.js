import React from "react";
import { Text ,SVG} from "../../../components/factorsComponents";
import styles from "../index.module.scss";

function TemplateCard({id,title,description,setTemplateSelected}){
    const handleTemplateCard=()=>{
        setTemplateSelected(()=>{return id});
    }
    return(
        // <Col span={8}>
            <div className={`${styles.cardnew}`} onClick={()=>handleTemplateCard()}  key={id}>
                <img alt='template' src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/template-icon-1.png' className={'mb-2 justify-center w-full'} />
                <Text type={'title'} level={6} color={'grey-2'} weight={'bold'} extraClass={'m-0 p-2'}>
                    {title}
                </Text>
                <Text type={'paragraph'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0 p-2'}>
                    {description}
                </Text>
            </div>
        // </Col>
    );
}
export default TemplateCard;