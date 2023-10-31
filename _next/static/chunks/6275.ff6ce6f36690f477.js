"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[6275],{6275:function(t,e,i){i.d(e,{diagram:function(){return v}});var r=i(6388),a=i(6357),n=i(4296);i(7693),i(7608),i(1699);var s=function(){var o=function(t,e,i,r){for(i=i||{},r=t.length;r--;i[t[r]]=e);return i},t=[6,8,10,11,12,14,16,17,18],e=[1,9],i=[1,10],r=[1,11],a=[1,12],n=[1,13],s=[1,14],l={trace:function(){},yy:{},symbols_:{error:2,start:3,journey:4,document:5,EOF:6,line:7,SPACE:8,statement:9,NEWLINE:10,title:11,acc_title:12,acc_title_value:13,acc_descr:14,acc_descr_value:15,acc_descr_multiline_value:16,section:17,taskName:18,taskData:19,$accept:0,$end:1},terminals_:{2:"error",4:"journey",6:"EOF",8:"SPACE",10:"NEWLINE",11:"title",12:"acc_title",13:"acc_title_value",14:"acc_descr",15:"acc_descr_value",16:"acc_descr_multiline_value",17:"section",18:"taskName",19:"taskData"},productions_:[0,[3,3],[5,0],[5,2],[7,2],[7,1],[7,1],[7,1],[9,1],[9,2],[9,2],[9,1],[9,1],[9,2]],performAction:function(t,e,i,r,a,n,s){var l=n.length-1;switch(a){case 1:return n[l-1];case 2:case 6:case 7:this.$=[];break;case 3:n[l-1].push(n[l]),this.$=n[l-1];break;case 4:case 5:this.$=n[l];break;case 8:r.setDiagramTitle(n[l].substr(6)),this.$=n[l].substr(6);break;case 9:this.$=n[l].trim(),r.setAccTitle(this.$);break;case 10:case 11:this.$=n[l].trim(),r.setAccDescription(this.$);break;case 12:r.addSection(n[l].substr(8)),this.$=n[l].substr(8);break;case 13:r.addTask(n[l-1],n[l]),this.$="task"}},table:[{3:1,4:[1,2]},{1:[3]},o(t,[2,2],{5:3}),{6:[1,4],7:5,8:[1,6],9:7,10:[1,8],11:e,12:i,14:r,16:a,17:n,18:s},o(t,[2,7],{1:[2,1]}),o(t,[2,3]),{9:15,11:e,12:i,14:r,16:a,17:n,18:s},o(t,[2,5]),o(t,[2,6]),o(t,[2,8]),{13:[1,16]},{15:[1,17]},o(t,[2,11]),o(t,[2,12]),{19:[1,18]},o(t,[2,4]),o(t,[2,9]),o(t,[2,10]),o(t,[2,13])],defaultActions:{},parseError:function(t,e){if(e.recoverable)this.trace(t);else{var i=Error(t);throw i.hash=e,i}},parse:function(t){var e=this,i=[0],r=[],a=[null],n=[],s=this.table,l="",c=0,h=0,u=n.slice.call(arguments,1),y=Object.create(this.lexer),p={yy:{}};for(var d in this.yy)Object.prototype.hasOwnProperty.call(this.yy,d)&&(p.yy[d]=this.yy[d]);y.setInput(t,p.yy),p.yy.lexer=y,p.yy.parser=this,void 0===y.yylloc&&(y.yylloc={});var f=y.yylloc;n.push(f);var g=y.options&&y.options.ranges;function lex(){var t;return"number"!=typeof(t=r.pop()||y.lex()||1)&&(t instanceof Array&&(t=(r=t).pop()),t=e.symbols_[t]||t),t}"function"==typeof p.yy.parseError?this.parseError=p.yy.parseError:this.parseError=Object.getPrototypeOf(this).parseError;for(var x,m,k,_,b,w,v,$,T={};;){if(m=i[i.length-1],this.defaultActions[m]?k=this.defaultActions[m]:(null==x&&(x=lex()),k=s[m]&&s[m][x]),void 0===k||!k.length||!k[0]){var M="";for(b in $=[],s[m])this.terminals_[b]&&b>2&&$.push("'"+this.terminals_[b]+"'");M=y.showPosition?"Parse error on line "+(c+1)+":\n"+y.showPosition()+"\nExpecting "+$.join(", ")+", got '"+(this.terminals_[x]||x)+"'":"Parse error on line "+(c+1)+": Unexpected "+(1==x?"end of input":"'"+(this.terminals_[x]||x)+"'"),this.parseError(M,{text:y.match,token:this.terminals_[x]||x,line:y.yylineno,loc:f,expected:$})}if(k[0]instanceof Array&&k.length>1)throw Error("Parse Error: multiple actions possible at state: "+m+", token: "+x);switch(k[0]){case 1:i.push(x),a.push(y.yytext),n.push(y.yylloc),i.push(k[1]),x=null,h=y.yyleng,l=y.yytext,c=y.yylineno,f=y.yylloc;break;case 2:if(w=this.productions_[k[1]][1],T.$=a[a.length-w],T._$={first_line:n[n.length-(w||1)].first_line,last_line:n[n.length-1].last_line,first_column:n[n.length-(w||1)].first_column,last_column:n[n.length-1].last_column},g&&(T._$.range=[n[n.length-(w||1)].range[0],n[n.length-1].range[1]]),void 0!==(_=this.performAction.apply(T,[l,h,c,p.yy,k[1],a,n].concat(u))))return _;w&&(i=i.slice(0,-1*w*2),a=a.slice(0,-1*w),n=n.slice(0,-1*w)),i.push(this.productions_[k[1]][0]),a.push(T.$),n.push(T._$),v=s[i[i.length-2]][i[i.length-1]],i.push(v);break;case 3:return!0}}return!0}};function Parser(){this.yy={}}return l.lexer={EOF:1,parseError:function(t,e){if(this.yy.parser)this.yy.parser.parseError(t,e);else throw Error(t)},setInput:function(t,e){return this.yy=e||this.yy||{},this._input=t,this._more=this._backtrack=this.done=!1,this.yylineno=this.yyleng=0,this.yytext=this.matched=this.match="",this.conditionStack=["INITIAL"],this.yylloc={first_line:1,first_column:0,last_line:1,last_column:0},this.options.ranges&&(this.yylloc.range=[0,0]),this.offset=0,this},input:function(){var t=this._input[0];return this.yytext+=t,this.yyleng++,this.offset++,this.match+=t,this.matched+=t,t.match(/(?:\r\n?|\n).*/g)?(this.yylineno++,this.yylloc.last_line++):this.yylloc.last_column++,this.options.ranges&&this.yylloc.range[1]++,this._input=this._input.slice(1),t},unput:function(t){var e=t.length,i=t.split(/(?:\r\n?|\n)/g);this._input=t+this._input,this.yytext=this.yytext.substr(0,this.yytext.length-e),this.offset-=e;var r=this.match.split(/(?:\r\n?|\n)/g);this.match=this.match.substr(0,this.match.length-1),this.matched=this.matched.substr(0,this.matched.length-1),i.length-1&&(this.yylineno-=i.length-1);var a=this.yylloc.range;return this.yylloc={first_line:this.yylloc.first_line,last_line:this.yylineno+1,first_column:this.yylloc.first_column,last_column:i?(i.length===r.length?this.yylloc.first_column:0)+r[r.length-i.length].length-i[0].length:this.yylloc.first_column-e},this.options.ranges&&(this.yylloc.range=[a[0],a[0]+this.yyleng-e]),this.yyleng=this.yytext.length,this},more:function(){return this._more=!0,this},reject:function(){return this.options.backtrack_lexer?(this._backtrack=!0,this):this.parseError("Lexical error on line "+(this.yylineno+1)+". You can only invoke reject() in the lexer when the lexer is of the backtracking persuasion (options.backtrack_lexer = true).\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},less:function(t){this.unput(this.match.slice(t))},pastInput:function(){var t=this.matched.substr(0,this.matched.length-this.match.length);return(t.length>20?"...":"")+t.substr(-20).replace(/\n/g,"")},upcomingInput:function(){var t=this.match;return t.length<20&&(t+=this._input.substr(0,20-t.length)),(t.substr(0,20)+(t.length>20?"...":"")).replace(/\n/g,"")},showPosition:function(){var t=this.pastInput(),e=Array(t.length+1).join("-");return t+this.upcomingInput()+"\n"+e+"^"},test_match:function(t,e){var i,r,a;if(this.options.backtrack_lexer&&(a={yylineno:this.yylineno,yylloc:{first_line:this.yylloc.first_line,last_line:this.last_line,first_column:this.yylloc.first_column,last_column:this.yylloc.last_column},yytext:this.yytext,match:this.match,matches:this.matches,matched:this.matched,yyleng:this.yyleng,offset:this.offset,_more:this._more,_input:this._input,yy:this.yy,conditionStack:this.conditionStack.slice(0),done:this.done},this.options.ranges&&(a.yylloc.range=this.yylloc.range.slice(0))),(r=t[0].match(/(?:\r\n?|\n).*/g))&&(this.yylineno+=r.length),this.yylloc={first_line:this.yylloc.last_line,last_line:this.yylineno+1,first_column:this.yylloc.last_column,last_column:r?r[r.length-1].length-r[r.length-1].match(/\r?\n?/)[0].length:this.yylloc.last_column+t[0].length},this.yytext+=t[0],this.match+=t[0],this.matches=t,this.yyleng=this.yytext.length,this.options.ranges&&(this.yylloc.range=[this.offset,this.offset+=this.yyleng]),this._more=!1,this._backtrack=!1,this._input=this._input.slice(t[0].length),this.matched+=t[0],i=this.performAction.call(this,this.yy,this,e,this.conditionStack[this.conditionStack.length-1]),this.done&&this._input&&(this.done=!1),i)return i;if(this._backtrack)for(var n in a)this[n]=a[n];return!1},next:function(){if(this.done)return this.EOF;this._input||(this.done=!0),this._more||(this.yytext="",this.match="");for(var t,e,i,r,a=this._currentRules(),n=0;n<a.length;n++)if((i=this._input.match(this.rules[a[n]]))&&(!e||i[0].length>e[0].length)){if(e=i,r=n,this.options.backtrack_lexer){if(!1!==(t=this.test_match(i,a[n])))return t;if(!this._backtrack)return!1;e=!1;continue}if(!this.options.flex)break}return e?!1!==(t=this.test_match(e,a[r]))&&t:""===this._input?this.EOF:this.parseError("Lexical error on line "+(this.yylineno+1)+". Unrecognized text.\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},lex:function(){return this.next()||this.lex()},begin:function(t){this.conditionStack.push(t)},popState:function(){return this.conditionStack.length-1>0?this.conditionStack.pop():this.conditionStack[0]},_currentRules:function(){return this.conditionStack.length&&this.conditionStack[this.conditionStack.length-1]?this.conditions[this.conditionStack[this.conditionStack.length-1]].rules:this.conditions.INITIAL.rules},topState:function(t){return(t=this.conditionStack.length-1-Math.abs(t||0))>=0?this.conditionStack[t]:"INITIAL"},pushState:function(t){this.begin(t)},stateStackSize:function(){return this.conditionStack.length},options:{"case-insensitive":!0},performAction:function(t,e,i,r){switch(i){case 0:case 1:case 3:case 4:break;case 2:return 10;case 5:return 4;case 6:return 11;case 7:return this.begin("acc_title"),12;case 8:return this.popState(),"acc_title_value";case 9:return this.begin("acc_descr"),14;case 10:return this.popState(),"acc_descr_value";case 11:this.begin("acc_descr_multiline");break;case 12:this.popState();break;case 13:return"acc_descr_multiline_value";case 14:return 17;case 15:return 18;case 16:return 19;case 17:return":";case 18:return 6;case 19:return"INVALID"}},rules:[/^(?:%(?!\{)[^\n]*)/i,/^(?:[^\}]%%[^\n]*)/i,/^(?:[\n]+)/i,/^(?:\s+)/i,/^(?:#[^\n]*)/i,/^(?:journey\b)/i,/^(?:title\s[^#\n;]+)/i,/^(?:accTitle\s*:\s*)/i,/^(?:(?!\n||)*[^\n]*)/i,/^(?:accDescr\s*:\s*)/i,/^(?:(?!\n||)*[^\n]*)/i,/^(?:accDescr\s*\{\s*)/i,/^(?:[\}])/i,/^(?:[^\}]*)/i,/^(?:section\s[^#:\n;]+)/i,/^(?:[^#:\n;]+)/i,/^(?::[^#\n;]+)/i,/^(?::)/i,/^(?:$)/i,/^(?:.)/i],conditions:{acc_descr_multiline:{rules:[12,13],inclusive:!1},acc_descr:{rules:[10],inclusive:!1},acc_title:{rules:[8],inclusive:!1},INITIAL:{rules:[0,1,2,3,4,5,6,7,9,11,14,15,16,17,18,19],inclusive:!0}}},Parser.prototype=l,l.Parser=Parser,new Parser}();s.parser=s;let l="",c=[],h=[],u=[],updateActors=function(){let t=[];h.forEach(e=>{e.people&&t.push(...e.people)});let e=new Set(t);return[...e].sort()},compileTasks=function(){let t=!0;for(let[e,i]of u.entries())u[e].processed,t=t&&i.processed;return t},y={getConfig:()=>(0,r.c)().journey,clear:function(){c.length=0,h.length=0,l="",u.length=0,(0,r.t)()},setDiagramTitle:r.q,getDiagramTitle:r.r,setAccTitle:r.s,getAccTitle:r.g,setAccDescription:r.b,getAccDescription:r.a,addSection:function(t){l=t,c.push(t)},getSections:function(){return c},getTasks:function(){let t=compileTasks(),e=0;for(;!t&&e<100;)t=compileTasks(),e++;return h.push(...u),h},addTask:function(t,e){let i=e.substr(1).split(":"),r=0,a=[];1===i.length?(r=Number(i[0]),a=[]):(r=Number(i[0]),a=i[1].split(","));let n=a.map(t=>t.trim()),s={section:l,type:l,people:n,task:t,score:r};u.push(s)},addTaskOrg:function(t){let e={section:l,type:l,description:t,task:t,classes:[]};h.push(e)},getActors:function(){return updateActors()}},drawRect=function(t,e){return(0,n.d)(t,e)},drawFace=function(t,e){let i=t.append("circle").attr("cx",e.cx).attr("cy",e.cy).attr("class","face").attr("r",15).attr("stroke-width",2).attr("overflow","visible"),r=t.append("g");function smile(t){let i=(0,a.Nb1)().startAngle(Math.PI/2).endAngle(3*(Math.PI/2)).innerRadius(7.5).outerRadius(15/2.2);t.append("path").attr("class","mouth").attr("d",i).attr("transform","translate("+e.cx+","+(e.cy+2)+")")}function sad(t){let i=(0,a.Nb1)().startAngle(3*Math.PI/2).endAngle(5*(Math.PI/2)).innerRadius(7.5).outerRadius(15/2.2);t.append("path").attr("class","mouth").attr("d",i).attr("transform","translate("+e.cx+","+(e.cy+7)+")")}function ambivalent(t){t.append("line").attr("class","mouth").attr("stroke",2).attr("x1",e.cx-5).attr("y1",e.cy+7).attr("x2",e.cx+5).attr("y2",e.cy+7).attr("class","mouth").attr("stroke-width","1px").attr("stroke","#666")}return r.append("circle").attr("cx",e.cx-5).attr("cy",e.cy-5).attr("r",1.5).attr("stroke-width",2).attr("fill","#666").attr("stroke","#666"),r.append("circle").attr("cx",e.cx+5).attr("cy",e.cy-5).attr("r",1.5).attr("stroke-width",2).attr("fill","#666").attr("stroke","#666"),e.score>3?smile(r):e.score<3?sad(r):ambivalent(r),i},drawCircle=function(t,e){let i=t.append("circle");return i.attr("cx",e.cx),i.attr("cy",e.cy),i.attr("class","actor-"+e.pos),i.attr("fill",e.fill),i.attr("stroke",e.stroke),i.attr("r",e.r),void 0!==i.class&&i.attr("class",i.class),void 0!==e.title&&i.append("title").text(e.title),i},drawText=function(t,e){return(0,n.f)(t,e)},p=-1,d=function(){function byText(t,e,i,r,a,n,s,l){let c=e.append("text").attr("x",i+a/2).attr("y",r+n/2+5).style("font-color",l).style("text-anchor","middle").text(t);_setTextAttrs(c,s)}function byTspan(t,e,i,r,a,n,s,l,c){let{taskFontSize:h,taskFontFamily:u}=l,y=t.split(/<br\s*\/?>/gi);for(let t=0;t<y.length;t++){let l=t*h-h*(y.length-1)/2,p=e.append("text").attr("x",i+a/2).attr("y",r).attr("fill",c).style("text-anchor","middle").style("font-size",h).style("font-family",u);p.append("tspan").attr("x",i+a/2).attr("dy",l).text(y[t]),p.attr("y",r+n/2).attr("dominant-baseline","central").attr("alignment-baseline","central"),_setTextAttrs(p,s)}}function byFo(t,e,i,r,a,n,s,l){let c=e.append("switch"),h=c.append("foreignObject").attr("x",i).attr("y",r).attr("width",a).attr("height",n).attr("position","fixed"),u=h.append("xhtml:div").style("display","table").style("height","100%").style("width","100%");u.append("div").attr("class","label").style("display","table-cell").style("text-align","center").style("vertical-align","middle").text(t),byTspan(t,c,i,r,a,n,s,l),_setTextAttrs(u,s)}function _setTextAttrs(t,e){for(let i in e)i in e&&t.attr(i,e[i])}return function(t){return"fo"===t.textPlacement?byFo:"old"===t.textPlacement?byText:byTspan}}(),f={drawRect,drawCircle,drawSection:function(t,e,i){let r=t.append("g"),a=(0,n.g)();a.x=e.x,a.y=e.y,a.fill=e.fill,a.width=i.width*e.taskCount+i.diagramMarginX*(e.taskCount-1),a.height=i.height,a.class="journey-section section-type-"+e.num,a.rx=3,a.ry=3,drawRect(r,a),d(i)(e.text,r,a.x,a.y,a.width,a.height,{class:"journey-section section-type-"+e.num},i,e.colour)},drawText,drawLabel:function(t,e){function genPoints(t,e,i,r,a){return t+","+e+" "+(t+i)+","+e+" "+(t+i)+","+(e+r-a)+" "+(t+i-1.2*a)+","+(e+r)+" "+t+","+(e+r)}let i=t.append("polygon");i.attr("points",genPoints(e.x,e.y,50,20,7)),i.attr("class","labelBox"),e.y=e.y+e.labelMargin,e.x=e.x+.5*e.labelMargin,drawText(t,e)},drawTask:function(t,e,i){let r=e.x+i.width/2,a=t.append("g");p++,a.append("line").attr("id","task"+p).attr("x1",r).attr("y1",e.y).attr("x2",r).attr("y2",450).attr("class","task-line").attr("stroke-width","1px").attr("stroke-dasharray","4 2").attr("stroke","#666"),drawFace(a,{cx:r,cy:300+(5-e.score)*30,score:e.score});let s=(0,n.g)();s.x=e.x,s.y=e.y,s.fill=e.fill,s.width=i.width,s.height=i.height,s.class="task task-type-"+e.num,s.rx=3,s.ry=3,drawRect(a,s);let l=e.x+14;e.people.forEach(t=>{let i=e.actors[t].color,r={cx:l,cy:e.y,r:7,fill:i,stroke:"#000",title:t,pos:e.actors[t].position};drawCircle(a,r),l+=10}),d(i)(e.task,a,s.x,s.y,s.width,s.height,{class:"task"},i,e.colour)},drawBackgroundRect:function(t,e){(0,n.a)(t,e)},initGraphics:function(t){t.append("defs").append("marker").attr("id","arrowhead").attr("refX",5).attr("refY",2).attr("markerWidth",6).attr("markerHeight",4).attr("orient","auto").append("path").attr("d","M 0,0 V 4 L6,2 Z")}},g={};function drawActorLegend(t){let e=(0,r.c)().journey,i=60;Object.keys(g).forEach(r=>{let a=g[r].color,n={cx:20,cy:i,r:7,fill:a,stroke:"#000",pos:g[r].position};f.drawCircle(t,n);let s={x:40,y:i+7,fill:"#666",text:r,textMargin:5|e.boxTextMargin};f.drawText(t,s),i+=20})}let x=(0,r.c)().journey,m=x.leftMargin,k={data:{startx:void 0,stopx:void 0,starty:void 0,stopy:void 0},verticalPos:0,sequenceItems:[],init:function(){this.sequenceItems=[],this.data={startx:void 0,stopx:void 0,starty:void 0,stopy:void 0},this.verticalPos=0},updateVal:function(t,e,i,r){void 0===t[e]?t[e]=i:t[e]=r(i,t[e])},updateBounds:function(t,e,i,a){let n=(0,r.c)().journey,s=this,l=0;function updateFn(r){return function(c){l++;let h=s.sequenceItems.length-l+1;s.updateVal(c,"starty",e-h*n.boxMargin,Math.min),s.updateVal(c,"stopy",a+h*n.boxMargin,Math.max),s.updateVal(k.data,"startx",t-h*n.boxMargin,Math.min),s.updateVal(k.data,"stopx",i+h*n.boxMargin,Math.max),"activation"!==r&&(s.updateVal(c,"startx",t-h*n.boxMargin,Math.min),s.updateVal(c,"stopx",i+h*n.boxMargin,Math.max),s.updateVal(k.data,"starty",e-h*n.boxMargin,Math.min),s.updateVal(k.data,"stopy",a+h*n.boxMargin,Math.max))}}this.sequenceItems.forEach(updateFn())},insert:function(t,e,i,r){let a=Math.min(t,i),n=Math.max(t,i),s=Math.min(e,r),l=Math.max(e,r);this.updateVal(k.data,"startx",a,Math.min),this.updateVal(k.data,"starty",s,Math.min),this.updateVal(k.data,"stopx",n,Math.max),this.updateVal(k.data,"stopy",l,Math.max),this.updateBounds(a,s,n,l)},bumpVerticalPos:function(t){this.verticalPos=this.verticalPos+t,this.data.stopy=this.verticalPos},getVerticalPos:function(){return this.verticalPos},getBounds:function(){return this.data}},_=x.sectionFills,b=x.sectionColours,drawTasks=function(t,e,i){let a=(0,r.c)().journey,n="",s=2*a.height+a.diagramMarginY,l=i+s,c=0,h="#CCC",u="black",y=0;for(let[i,r]of e.entries()){if(n!==r.section){h=_[c%_.length],y=c%_.length,u=b[c%b.length];let s=0,l=r.section;for(let t=i;t<e.length;t++)if(e[t].section==l)s+=1;else break;let p={x:i*a.taskMargin+i*a.width+m,y:50,text:r.section,fill:h,num:y,colour:u,taskCount:s};f.drawSection(t,p,a),n=r.section,c++}let s=r.people.reduce((t,e)=>(g[e]&&(t[e]=g[e]),t),{});r.x=i*a.taskMargin+i*a.width+m,r.y=l,r.width=a.diagramMarginX,r.height=a.diagramMarginY,r.colour=u,r.fill=h,r.num=y,r.actors=s,f.drawTask(t,r,a),k.insert(r.x,r.y,r.x+r.width+a.taskMargin,450)}},w={setConf:function(t){let e=Object.keys(t);e.forEach(function(e){x[e]=t[e]})},draw:function(t,e,i,n){let s;let l=(0,r.c)().journey,c=(0,r.c)().securityLevel;"sandbox"===c&&(s=(0,a.Ys)("#i"+e));let h="sandbox"===c?(0,a.Ys)(s.nodes()[0].contentDocument.body):(0,a.Ys)("body");k.init();let u=h.select("#"+e);f.initGraphics(u);let y=n.db.getTasks(),p=n.db.getDiagramTitle(),d=n.db.getActors();for(let t in g)delete g[t];let x=0;d.forEach(t=>{g[t]={color:l.actorColours[x%l.actorColours.length],position:x},x++}),drawActorLegend(u),k.insert(0,0,m,50*Object.keys(g).length),drawTasks(u,y,0);let _=k.getBounds();p&&u.append("text").text(p).attr("x",m).attr("font-size","4ex").attr("font-weight","bold").attr("y",25);let b=_.stopy-_.starty+2*l.diagramMarginY,w=m+_.stopx+2*l.diagramMarginX;(0,r.i)(u,b,w,l.useMaxWidth),u.append("line").attr("x1",m).attr("y1",4*l.height).attr("x2",w-m-4).attr("y2",4*l.height).attr("stroke-width",4).attr("stroke","black").attr("marker-end","url(#arrowhead)");let v=p?70:0;u.attr("viewBox",`${_.startx} -25 ${w} ${b+v}`),u.attr("preserveAspectRatio","xMinYMin meet"),u.attr("height",b+v+25)}},v={parser:s,db:y,renderer:w,styles:t=>`.label {
    font-family: 'trebuchet ms', verdana, arial, sans-serif;
    font-family: var(--mermaid-font-family);
    color: ${t.textColor};
  }
  .mouth {
    stroke: #666;
  }

  line {
    stroke: ${t.textColor}
  }

  .legend {
    fill: ${t.textColor};
  }

  .label text {
    fill: #333;
  }
  .label {
    color: ${t.textColor}
  }

  .face {
    ${t.faceColor?`fill: ${t.faceColor}`:"fill: #FFF8DC"};
    stroke: #999;
  }

  .node rect,
  .node circle,
  .node ellipse,
  .node polygon,
  .node path {
    fill: ${t.mainBkg};
    stroke: ${t.nodeBorder};
    stroke-width: 1px;
  }

  .node .label {
    text-align: center;
  }
  .node.clickable {
    cursor: pointer;
  }

  .arrowheadPath {
    fill: ${t.arrowheadColor};
  }

  .edgePath .path {
    stroke: ${t.lineColor};
    stroke-width: 1.5px;
  }

  .flowchart-link {
    stroke: ${t.lineColor};
    fill: none;
  }

  .edgeLabel {
    background-color: ${t.edgeLabelBackground};
    rect {
      opacity: 0.5;
    }
    text-align: center;
  }

  .cluster rect {
  }

  .cluster text {
    fill: ${t.titleColor};
  }

  div.mermaidTooltip {
    position: absolute;
    text-align: center;
    max-width: 200px;
    padding: 2px;
    font-family: 'trebuchet ms', verdana, arial, sans-serif;
    font-family: var(--mermaid-font-family);
    font-size: 12px;
    background: ${t.tertiaryColor};
    border: 1px solid ${t.border2};
    border-radius: 2px;
    pointer-events: none;
    z-index: 100;
  }

  .task-type-0, .section-type-0  {
    ${t.fillType0?`fill: ${t.fillType0}`:""};
  }
  .task-type-1, .section-type-1  {
    ${t.fillType0?`fill: ${t.fillType1}`:""};
  }
  .task-type-2, .section-type-2  {
    ${t.fillType0?`fill: ${t.fillType2}`:""};
  }
  .task-type-3, .section-type-3  {
    ${t.fillType0?`fill: ${t.fillType3}`:""};
  }
  .task-type-4, .section-type-4  {
    ${t.fillType0?`fill: ${t.fillType4}`:""};
  }
  .task-type-5, .section-type-5  {
    ${t.fillType0?`fill: ${t.fillType5}`:""};
  }
  .task-type-6, .section-type-6  {
    ${t.fillType0?`fill: ${t.fillType6}`:""};
  }
  .task-type-7, .section-type-7  {
    ${t.fillType0?`fill: ${t.fillType7}`:""};
  }

  .actor-0 {
    ${t.actor0?`fill: ${t.actor0}`:""};
  }
  .actor-1 {
    ${t.actor1?`fill: ${t.actor1}`:""};
  }
  .actor-2 {
    ${t.actor2?`fill: ${t.actor2}`:""};
  }
  .actor-3 {
    ${t.actor3?`fill: ${t.actor3}`:""};
  }
  .actor-4 {
    ${t.actor4?`fill: ${t.actor4}`:""};
  }
  .actor-5 {
    ${t.actor5?`fill: ${t.actor5}`:""};
  }
`,init:t=>{w.setConf(t.journey),y.clear()}}},4296:function(t,e,i){i.d(e,{a:function(){return drawBackgroundRect},b:function(){return drawEmbeddedImage},c:function(){return drawImage},d:function(){return drawRect},e:function(){return getTextObj},f:function(){return drawText},g:function(){return getNoteRect}});var r=i(7608),a=i(6388);let drawRect=(t,e)=>{let i=t.append("rect");if(i.attr("x",e.x),i.attr("y",e.y),i.attr("fill",e.fill),i.attr("stroke",e.stroke),i.attr("width",e.width),i.attr("height",e.height),void 0!==e.rx&&i.attr("rx",e.rx),void 0!==e.ry&&i.attr("ry",e.ry),void 0!==e.attrs)for(let t in e.attrs)i.attr(t,e.attrs[t]);return void 0!==e.class&&i.attr("class",e.class),i},drawBackgroundRect=(t,e)=>{let i={x:e.startx,y:e.starty,width:e.stopx-e.startx,height:e.stopy-e.starty,fill:e.fill,stroke:e.stroke,class:"rect"},r=drawRect(t,i);r.lower()},drawText=(t,e)=>{let i=e.text.replace(a.H," "),r=t.append("text");r.attr("x",e.x),r.attr("y",e.y),r.attr("class","legend"),r.style("text-anchor",e.anchor),void 0!==e.class&&r.attr("class",e.class);let n=r.append("tspan");return n.attr("x",e.x+2*e.textMargin),n.text(i),r},drawImage=(t,e,i,a)=>{let n=t.append("image");n.attr("x",e),n.attr("y",i);let s=(0,r.Nm)(a);n.attr("xlink:href",s)},drawEmbeddedImage=(t,e,i,a)=>{let n=t.append("use");n.attr("x",e),n.attr("y",i);let s=(0,r.Nm)(a);n.attr("xlink:href",`#${s}`)},getNoteRect=()=>({x:0,y:0,width:100,height:100,fill:"#EDF2AE",stroke:"#666",anchor:"start",rx:0,ry:0}),getTextObj=()=>({x:0,y:0,width:100,height:100,"text-anchor":"start",style:"#666",textMargin:0,rx:0,ry:0,tspan:!0})}}]);