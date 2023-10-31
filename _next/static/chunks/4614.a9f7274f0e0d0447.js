"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[4614],{6157:function(e,t,l){l.d(t,{a:function(){return addHtmlLabel}});var r=l(6715);function addHtmlLabel(e,t){var l=e.append("foreignObject").attr("width","100000"),a=l.append("xhtml:div");a.attr("xmlns","http://www.w3.org/1999/xhtml");var o=t.label;switch(typeof o){case"function":a.insert(o);break;case"object":a.insert(function(){return o});break;default:a.html(o)}r.bg(a,t.labelStyle),a.style("display","inline-block"),a.style("white-space","nowrap");var n=a.node().getBoundingClientRect();return l.attr("width",n.width).attr("height",n.height),l}},6715:function(e,t,l){l.d(t,{$p:function(){return applyClass},O1:function(){return edgeToId},WR:function(){return applyTransition},bF:function(){return isSubgraph},bg:function(){return applyStyle}});var r=l(2701),a=l(8246);function isSubgraph(e,t){return!!e.children(t).length}function edgeToId(e){return escapeId(e.v)+":"+escapeId(e.w)+":"+escapeId(e.name)}var o=/:/g;function escapeId(e){return e?String(e).replace(o,"\\:"):""}function applyStyle(e,t){t&&e.attr("style",t)}function applyClass(e,t,l){t&&e.attr("class",t).attr("class",l+" "+e.attr("class"))}function applyTransition(e,t){var l=t.graph();if(r.Z(l)){var o=l.transition;if(a.Z(o))return o(e)}return e}},4614:function(e,t,l){l.d(t,{diagram:function(){return n}});var r=l(5116),a=l(97),o=l(6388);l(6357),l(8472),l(9500),l(6576),l(7693),l(7608),l(1699);let n={parser:r.p,db:r.f,renderer:a.f,styles:a.a,init:e=>{e.flowchart||(e.flowchart={}),e.flowchart.arrowMarkerAbsolute=e.arrowMarkerAbsolute,(0,o.p)({flowchart:{arrowMarkerAbsolute:e.arrowMarkerAbsolute}}),a.f.setConf(e.flowchart),r.f.clear(),r.f.setGen("gen-2")}}},97:function(e,t,l){l.d(t,{a:function(){return flowStyles},f:function(){return b}});var r=l(8472),a=l(6357),o=l(6388),n=l(5220),i=l(6157),s=l(3445),d=l(1739),methods_channel=(e,t)=>s.Z.lang.round(d.Z.parse(e)[t]),c=l(6442);let p={},addVertices=function(e,t,l,r,a,n){let s=r.select(`[id="${l}"]`),d=Object.keys(e);d.forEach(function(l){let r;let d=e[l],c="default";d.classes.length>0&&(c=d.classes.join(" ")),c+=" flowchart-label";let p=(0,o.k)(d.styles),b=void 0!==d.text?d.text:d.id;if(o.l.info("vertex",d,d.labelType),"markdown"===d.labelType)o.l.info("vertex",d,d.labelType);else if((0,o.m)((0,o.c)().flowchart.htmlLabels)){let e={label:b.replace(/fa[blrs]?:fa-[\w-]+/g,e=>`<i class='${e.replace(":"," ")}'></i>`)};(r=(0,i.a)(s,e).node()).parentNode.removeChild(r)}else{let e=a.createElementNS("http://www.w3.org/2000/svg","text");e.setAttribute("style",p.labelStyle.replace("color:","fill:"));let t=b.split(o.e.lineBreakRegex);for(let l of t){let t=a.createElementNS("http://www.w3.org/2000/svg","tspan");t.setAttributeNS("http://www.w3.org/XML/1998/namespace","xml:space","preserve"),t.setAttribute("dy","1em"),t.setAttribute("x","1"),t.textContent=l,e.appendChild(t)}r=e}let u=0,f="";switch(d.type){case"round":u=5,f="rect";break;case"square":case"group":default:f="rect";break;case"diamond":f="question";break;case"hexagon":f="hexagon";break;case"odd":case"odd_right":f="rect_left_inv_arrow";break;case"lean_right":f="lean_right";break;case"lean_left":f="lean_left";break;case"trapezoid":f="trapezoid";break;case"inv_trapezoid":f="inv_trapezoid";break;case"circle":f="circle";break;case"ellipse":f="ellipse";break;case"stadium":f="stadium";break;case"subroutine":f="subroutine";break;case"cylinder":f="cylinder";break;case"doublecircle":f="doublecircle"}t.setNode(d.id,{labelStyle:p.labelStyle,shape:f,labelText:b,labelType:d.labelType,rx:u,ry:u,class:c,style:p.style,id:d.id,link:d.link,linkTarget:d.linkTarget,tooltip:n.db.getTooltip(d.id)||"",domId:n.db.lookUpDomId(d.id),haveCallback:d.haveCallback,width:"group"===d.type?500:void 0,dir:d.dir,type:d.type,props:d.props,padding:(0,o.c)().flowchart.padding}),o.l.info("setNode",{labelStyle:p.labelStyle,labelType:d.labelType,shape:f,labelText:b,rx:u,ry:u,class:c,style:p.style,id:d.id,domId:n.db.lookUpDomId(d.id),width:"group"===d.type?500:void 0,type:d.type,dir:d.dir,props:d.props,padding:(0,o.c)().flowchart.padding})})},addEdges=function(e,t,l){let r,n;o.l.info("abc78 edges = ",e);let i=0,s={};if(void 0!==e.defaultStyle){let t=(0,o.k)(e.defaultStyle);r=t.style,n=t.labelStyle}e.forEach(function(l){i++;let d="L-"+l.start+"-"+l.end;void 0===s[d]?s[d]=0:s[d]++,o.l.info("abc78 new entry",d,s[d]);let c=d+"-"+s[d];o.l.info("abc78 new link id to be used is",d,c,s[d]);let b="LS-"+l.start,u="LE-"+l.end,f={style:"",labelStyle:""};switch(f.minlen=l.length||1,"arrow_open"===l.type?f.arrowhead="none":f.arrowhead="normal",f.arrowTypeStart="arrow_open",f.arrowTypeEnd="arrow_open",l.type){case"double_arrow_cross":f.arrowTypeStart="arrow_cross";case"arrow_cross":f.arrowTypeEnd="arrow_cross";break;case"double_arrow_point":f.arrowTypeStart="arrow_point";case"arrow_point":f.arrowTypeEnd="arrow_point";break;case"double_arrow_circle":f.arrowTypeStart="arrow_circle";case"arrow_circle":f.arrowTypeEnd="arrow_circle"}let w="",h="";switch(l.stroke){case"normal":w="fill:none;",void 0!==r&&(w=r),void 0!==n&&(h=n),f.thickness="normal",f.pattern="solid";break;case"dotted":f.thickness="normal",f.pattern="dotted",f.style="fill:none;stroke-width:2px;stroke-dasharray:3;";break;case"thick":f.thickness="thick",f.pattern="solid",f.style="stroke-width: 3.5px;fill:none;";break;case"invisible":f.thickness="invisible",f.pattern="solid",f.style="stroke-width: 0;fill:none;"}if(void 0!==l.style){let e=(0,o.k)(l.style);w=e.style,h=e.labelStyle}f.style=f.style+=w,f.labelStyle=f.labelStyle+=h,void 0!==l.interpolate?f.curve=(0,o.n)(l.interpolate,a.c_6):void 0!==e.defaultInterpolate?f.curve=(0,o.n)(e.defaultInterpolate,a.c_6):f.curve=(0,o.n)(p.curve,a.c_6),void 0===l.text?void 0!==l.style&&(f.arrowheadStyle="fill: #333"):(f.arrowheadStyle="fill: #333",f.labelpos="c"),f.labelType=l.labelType,f.label=l.text.replace(o.e.lineBreakRegex,"\n"),void 0===l.style&&(f.style=f.style||"stroke: #333; stroke-width: 1.5px;fill:none;"),f.labelStyle=f.labelStyle.replace("color:","fill:"),f.id=c,f.classes="flowchart-link "+b+" "+u,t.setEdge(l.start,l.end,f,i)})},draw=async function(e,t,l,i){let s,d;o.l.info("Drawing flowchart");let c=i.db.getDirection();void 0===c&&(c="TD");let{securityLevel:p,flowchart:b}=(0,o.c)(),u=b.nodeSpacing||50,f=b.rankSpacing||50;"sandbox"===p&&(s=(0,a.Ys)("#i"+t));let w="sandbox"===p?(0,a.Ys)(s.nodes()[0].contentDocument.body):(0,a.Ys)("body"),h="sandbox"===p?s.nodes()[0].contentDocument:document,g=new r.k({multigraph:!0,compound:!0}).setGraph({rankdir:c,nodesep:u,ranksep:f,marginx:0,marginy:0}).setDefaultEdgeLabel(function(){return{}}),y=i.db.getSubGraphs();o.l.info("Subgraphs - ",y);for(let e=y.length-1;e>=0;e--)d=y[e],o.l.info("Subgraph - ",d),i.db.addVertex(d.id,{text:d.title,type:d.labelType},"group",void 0,d.classes,d.dir);let k=i.db.getVertices(),x=i.db.getEdges();o.l.info("Edges",x);let m=0;for(m=y.length-1;m>=0;m--){d=y[m],(0,a.td_)("cluster").append("text");for(let e=0;e<d.nodes.length;e++)o.l.info("Setting up subgraphs",d.nodes[e],d.id),g.setParent(d.nodes[e],d.id)}addVertices(k,g,t,w,h,i),addEdges(x,g);let v=w.select(`[id="${t}"]`),S=w.select("#"+t+" g");if(await (0,n.r)(S,g,["point","circle","cross"],"flowchart",t),o.u.insertTitle(v,"flowchartTitleText",b.titleTopMargin,i.db.getDiagramTitle()),(0,o.o)(g,v,b.diagramPadding,b.useMaxWidth),i.db.indexNodes("subGraph"+m),!b.htmlLabels){let e=h.querySelectorAll('[id="'+t+'"] .edgeLabel .label');for(let t of e){let e=t.getBBox(),l=h.createElementNS("http://www.w3.org/2000/svg","rect");l.setAttribute("rx",0),l.setAttribute("ry",0),l.setAttribute("width",e.width),l.setAttribute("height",e.height),t.insertBefore(l,t.firstChild)}}let T=Object.keys(k);T.forEach(function(e){let l=k[e];if(l.link){let r=(0,a.Ys)("#"+t+' [id="'+e+'"]');if(r){let e=h.createElementNS("http://www.w3.org/2000/svg","a");e.setAttributeNS("http://www.w3.org/2000/svg","class",l.classes.join(" ")),e.setAttributeNS("http://www.w3.org/2000/svg","href",l.link),e.setAttributeNS("http://www.w3.org/2000/svg","rel","noopener"),"sandbox"===p?e.setAttributeNS("http://www.w3.org/2000/svg","target","_top"):l.linkTarget&&e.setAttributeNS("http://www.w3.org/2000/svg","target",l.linkTarget);let t=r.insert(function(){return e},":first-child"),a=r.select(".label-container");a&&t.append(function(){return a.node()});let o=r.select(".label");o&&t.append(function(){return o.node()})}}})},b={setConf:function(e){let t=Object.keys(e);for(let l of t)p[l]=e[l]},addVertices,addEdges,getClasses:function(e,t){return t.db.getClasses()},draw},fade=(e,t)=>{let l=methods_channel(e,"r"),r=methods_channel(e,"g"),a=methods_channel(e,"b");return c.Z(l,r,a,t)},flowStyles=e=>`.label {
    font-family: ${e.fontFamily};
    color: ${e.nodeTextColor||e.textColor};
  }
  .cluster-label text {
    fill: ${e.titleColor};
  }
  .cluster-label span,p {
    color: ${e.titleColor};
  }

  .label text,span,p {
    fill: ${e.nodeTextColor||e.textColor};
    color: ${e.nodeTextColor||e.textColor};
  }

  .node rect,
  .node circle,
  .node ellipse,
  .node polygon,
  .node path {
    fill: ${e.mainBkg};
    stroke: ${e.nodeBorder};
    stroke-width: 1px;
  }
  .flowchart-label text {
    text-anchor: middle;
  }
  // .flowchart-label .text-outer-tspan {
  //   text-anchor: middle;
  // }
  // .flowchart-label .text-inner-tspan {
  //   text-anchor: start;
  // }

  .node .label {
    text-align: center;
  }
  .node.clickable {
    cursor: pointer;
  }

  .arrowheadPath {
    fill: ${e.arrowheadColor};
  }

  .edgePath .path {
    stroke: ${e.lineColor};
    stroke-width: 2.0px;
  }

  .flowchart-link {
    stroke: ${e.lineColor};
    fill: none;
  }

  .edgeLabel {
    background-color: ${e.edgeLabelBackground};
    rect {
      opacity: 0.5;
      background-color: ${e.edgeLabelBackground};
      fill: ${e.edgeLabelBackground};
    }
    text-align: center;
  }

  /* For html labels only */
  .labelBkg {
    background-color: ${fade(e.edgeLabelBackground,.5)};
    // background-color: 
  }

  .cluster rect {
    fill: ${e.clusterBkg};
    stroke: ${e.clusterBorder};
    stroke-width: 1px;
  }

  .cluster text {
    fill: ${e.titleColor};
  }

  .cluster span,p {
    color: ${e.titleColor};
  }
  /* .cluster div {
    color: ${e.titleColor};
  } */

  div.mermaidTooltip {
    position: absolute;
    text-align: center;
    max-width: 200px;
    padding: 2px;
    font-family: ${e.fontFamily};
    font-size: 12px;
    background: ${e.tertiaryColor};
    border: 1px solid ${e.border2};
    border-radius: 2px;
    pointer-events: none;
    z-index: 100;
  }

  .flowchartTitleText {
    text-anchor: middle;
    font-size: 18px;
    fill: ${e.textColor};
  }
`}}]);