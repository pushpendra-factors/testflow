import React from "react";
import SkeletonCard from ".";

const blankArr = (new Array(100)).fill(0).map((e,i)=>i)
export default function(){
    return <div className="flex w-full flex-wrap gap-5">
        {blankArr.map((e)=><SkeletonCard key={e} index={e}/>)}
    </div>
}