import {   
    CHART_TYPE_STACKED_AREA,
    CHART_TYPE_LINECHART,
    CHART_TYPE_STACKED_BAR,
    CHART_TYPE_SPARKLINES,
    CHART_TYPE_BARCHART,
    CHART_TYPE_SCATTER_PLOT,
    CHART_TYPE_HORIZONTAL_BAR_CHART,
    CHART_TYPE_PIVOT_CHART 
} from "../utils/constants";

export const CHART_TYPES_DROPDOWN_CONSTANTS = {
    [CHART_TYPE_SPARKLINES ]:"A sparkline presents the general trend in variation of a metric. It helps you understand data fast and easy.",
    [CHART_TYPE_LINECHART]:"A line chart connects a series of data points with a continuous line. A classic way to observe a variable change over time.",
    [CHART_TYPE_STACKED_AREA]:"A stacked area chart has several area series stacked over each other. The height of a series reflects its value. ",
    [CHART_TYPE_HORIZONTAL_BAR_CHART]:"Bar charts present categorical data with heights that are proportional to their values. They're great for visual comparisons amongst different properties that make up an overall dataset. ",
    [CHART_TYPE_BARCHART]:"Column charts present categorical data with heights that are proportional to their values. They're great for visual comparisons amongst different properties that make up an overall dataset. ",
    [CHART_TYPE_SCATTER_PLOT]:"Scatter plots show you the relationship between 2 variables. See how the occurrence of a variable impacts an outcome. ",
    [CHART_TYPE_STACKED_BAR]:"A Stacked Column shows you a vertically stacked data series. They're great to observe how each of several variables and their sum change.",
    [CHART_TYPE_PIVOT_CHART]:"A pivot chart is an easy way to summarise large amounts of data in a friendly way. "
};