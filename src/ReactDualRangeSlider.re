/* module ReactDualRangeslider = { */
/* [@bs.module "react-dual-rangeslider"] */
/* [@bs.module]
   external rangeSlider: ReasonReact.reactClass =
     "react-dual-rangeslider/dist/RangeSlider.js"; */

[@bs.module]
/* external rangeSlider: ReasonReact.reactClass = "react-dual-range-slider"; */
external reactClass: ReasonReact.reactClass = "react-dual-range-slider";
/*
 let make = (~min: , ~max: float, ~minRange: float, ~step: float) =>
   ReasonReact.wrapJsForReason(
     ~reactClass=rangeSlider,
     ~props={"min": min, "max": max, "minRange": minRange, "step": step},
   ); */

/* }; */

let make = (~limits: array(int), ~values: array(int), children) =>
  ReasonReact.wrapJsForReason(
    ~reactClass,
    ~props=Js.Nullable.{"limits": limits, "values": values},
    children,
  );