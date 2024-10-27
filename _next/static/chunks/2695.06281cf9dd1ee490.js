"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[2695],{1219:(e,t,i)=>{i.d(t,{Z:()=>l});var s=i(3445),r=i(1739);let l=(e,t)=>s.Z.lang.round(r.Z.parse(e)[t])},9593:(e,t,i)=>{i.d(t,{Z:()=>r});var s=i(1066);let r=function(e){return(0,s.Z)(e,4)}},2695:(e,t,i)=>{i.d(t,{diagram:()=>F});var s,r,l=i(9523),n=i(9593),o=i(1219),a=i(6442),c=i(1724),u=i(8472),h=i(6357);i(7693),i(7608),i(1699);var d=function(){var e=function(e,t,i,s){for(i=i||{},s=e.length;s--;i[e[s]]=t);return i},t=[1,7],i=[1,13],s=[1,14],r=[1,15],l=[1,19],n=[1,16],o=[1,17],a=[1,18],c=[8,30],u=[8,21,28,29,30,31,32,40,44,47],h=[1,23],d=[1,24],g=[8,15,16,21,28,29,30,31,32,40,44,47],y=[8,15,16,21,27,28,29,30,31,32,40,44,47],p=[1,49],b={trace:function(){},yy:{},symbols_:{error:2,spaceLines:3,SPACELINE:4,NL:5,separator:6,SPACE:7,EOF:8,start:9,BLOCK_DIAGRAM_KEY:10,document:11,stop:12,statement:13,link:14,LINK:15,START_LINK:16,LINK_LABEL:17,STR:18,nodeStatement:19,columnsStatement:20,SPACE_BLOCK:21,blockStatement:22,classDefStatement:23,cssClassStatement:24,styleStatement:25,node:26,SIZE:27,COLUMNS:28,"id-block":29,end:30,block:31,NODE_ID:32,nodeShapeNLabel:33,dirList:34,DIR:35,NODE_DSTART:36,NODE_DEND:37,BLOCK_ARROW_START:38,BLOCK_ARROW_END:39,classDef:40,CLASSDEF_ID:41,CLASSDEF_STYLEOPTS:42,DEFAULT:43,class:44,CLASSENTITY_IDS:45,STYLECLASS:46,style:47,STYLE_ENTITY_IDS:48,STYLE_DEFINITION_DATA:49,$accept:0,$end:1},terminals_:{2:"error",4:"SPACELINE",5:"NL",7:"SPACE",8:"EOF",10:"BLOCK_DIAGRAM_KEY",15:"LINK",16:"START_LINK",17:"LINK_LABEL",18:"STR",21:"SPACE_BLOCK",27:"SIZE",28:"COLUMNS",29:"id-block",30:"end",31:"block",32:"NODE_ID",35:"DIR",36:"NODE_DSTART",37:"NODE_DEND",38:"BLOCK_ARROW_START",39:"BLOCK_ARROW_END",40:"classDef",41:"CLASSDEF_ID",42:"CLASSDEF_STYLEOPTS",43:"DEFAULT",44:"class",45:"CLASSENTITY_IDS",46:"STYLECLASS",47:"style",48:"STYLE_ENTITY_IDS",49:"STYLE_DEFINITION_DATA"},productions_:[0,[3,1],[3,2],[3,2],[6,1],[6,1],[6,1],[9,3],[12,1],[12,1],[12,2],[12,2],[11,1],[11,2],[14,1],[14,4],[13,1],[13,1],[13,1],[13,1],[13,1],[13,1],[13,1],[19,3],[19,2],[19,1],[20,1],[22,4],[22,3],[26,1],[26,2],[34,1],[34,2],[33,3],[33,4],[23,3],[23,3],[24,3],[25,3]],performAction:function(e,t,i,s,r,l,n){var o=l.length-1;switch(r){case 4:s.getLogger().debug("Rule: separator (NL) ");break;case 5:s.getLogger().debug("Rule: separator (Space) ");break;case 6:s.getLogger().debug("Rule: separator (EOF) ");break;case 7:s.getLogger().debug("Rule: hierarchy: ",l[o-1]),s.setHierarchy(l[o-1]);break;case 8:s.getLogger().debug("Stop NL ");break;case 9:s.getLogger().debug("Stop EOF ");break;case 10:s.getLogger().debug("Stop NL2 ");break;case 11:s.getLogger().debug("Stop EOF2 ");break;case 12:s.getLogger().debug("Rule: statement: ",l[o]),"number"==typeof l[o].length?this.$=l[o]:this.$=[l[o]];break;case 13:s.getLogger().debug("Rule: statement #2: ",l[o-1]),this.$=[l[o-1]].concat(l[o]);break;case 14:s.getLogger().debug("Rule: link: ",l[o],e),this.$={edgeTypeStr:l[o],label:""};break;case 15:s.getLogger().debug("Rule: LABEL link: ",l[o-3],l[o-1],l[o]),this.$={edgeTypeStr:l[o],label:l[o-1]};break;case 18:let a=parseInt(l[o]),c=s.generateId();this.$={id:c,type:"space",label:"",width:a,children:[]};break;case 23:s.getLogger().debug("Rule: (nodeStatement link node) ",l[o-2],l[o-1],l[o]," typestr: ",l[o-1].edgeTypeStr);let u=s.edgeStrToEdgeData(l[o-1].edgeTypeStr);this.$=[{id:l[o-2].id,label:l[o-2].label,type:l[o-2].type,directions:l[o-2].directions},{id:l[o-2].id+"-"+l[o].id,start:l[o-2].id,end:l[o].id,label:l[o-1].label,type:"edge",directions:l[o].directions,arrowTypeEnd:u,arrowTypeStart:"arrow_open"},{id:l[o].id,label:l[o].label,type:s.typeStr2Type(l[o].typeStr),directions:l[o].directions}];break;case 24:s.getLogger().debug("Rule: nodeStatement (abc88 node size) ",l[o-1],l[o]),this.$={id:l[o-1].id,label:l[o-1].label,type:s.typeStr2Type(l[o-1].typeStr),directions:l[o-1].directions,widthInColumns:parseInt(l[o],10)};break;case 25:s.getLogger().debug("Rule: nodeStatement (node) ",l[o]),this.$={id:l[o].id,label:l[o].label,type:s.typeStr2Type(l[o].typeStr),directions:l[o].directions,widthInColumns:1};break;case 26:s.getLogger().debug("APA123",this?this:"na"),s.getLogger().debug("COLUMNS: ",l[o]),this.$={type:"column-setting",columns:"auto"===l[o]?-1:parseInt(l[o])};break;case 27:s.getLogger().debug("Rule: id-block statement : ",l[o-2],l[o-1]),s.generateId(),this.$={...l[o-2],type:"composite",children:l[o-1]};break;case 28:s.getLogger().debug("Rule: blockStatement : ",l[o-2],l[o-1],l[o]);let h=s.generateId();this.$={id:h,type:"composite",label:"",children:l[o-1]};break;case 29:s.getLogger().debug("Rule: node (NODE_ID separator): ",l[o]),this.$={id:l[o]};break;case 30:s.getLogger().debug("Rule: node (NODE_ID nodeShapeNLabel separator): ",l[o-1],l[o]),this.$={id:l[o-1],label:l[o].label,typeStr:l[o].typeStr,directions:l[o].directions};break;case 31:s.getLogger().debug("Rule: dirList: ",l[o]),this.$=[l[o]];break;case 32:s.getLogger().debug("Rule: dirList: ",l[o-1],l[o]),this.$=[l[o-1]].concat(l[o]);break;case 33:s.getLogger().debug("Rule: nodeShapeNLabel: ",l[o-2],l[o-1],l[o]),this.$={typeStr:l[o-2]+l[o],label:l[o-1]};break;case 34:s.getLogger().debug("Rule: BLOCK_ARROW nodeShapeNLabel: ",l[o-3],l[o-2]," #3:",l[o-1],l[o]),this.$={typeStr:l[o-3]+l[o],label:l[o-2],directions:l[o-1]};break;case 35:case 36:this.$={type:"classDef",id:l[o-1].trim(),css:l[o].trim()};break;case 37:this.$={type:"applyClass",id:l[o-1].trim(),styleClass:l[o].trim()};break;case 38:this.$={type:"applyStyles",id:l[o-1].trim(),stylesStr:l[o].trim()}}},table:[{9:1,10:[1,2]},{1:[3]},{11:3,13:4,19:5,20:6,21:t,22:8,23:9,24:10,25:11,26:12,28:i,29:s,31:r,32:l,40:n,44:o,47:a},{8:[1,20]},e(c,[2,12],{13:4,19:5,20:6,22:8,23:9,24:10,25:11,26:12,11:21,21:t,28:i,29:s,31:r,32:l,40:n,44:o,47:a}),e(u,[2,16],{14:22,15:h,16:d}),e(u,[2,17]),e(u,[2,18]),e(u,[2,19]),e(u,[2,20]),e(u,[2,21]),e(u,[2,22]),e(g,[2,25],{27:[1,25]}),e(u,[2,26]),{19:26,26:12,32:l},{11:27,13:4,19:5,20:6,21:t,22:8,23:9,24:10,25:11,26:12,28:i,29:s,31:r,32:l,40:n,44:o,47:a},{41:[1,28],43:[1,29]},{45:[1,30]},{48:[1,31]},e(y,[2,29],{33:32,36:[1,33],38:[1,34]}),{1:[2,7]},e(c,[2,13]),{26:35,32:l},{32:[2,14]},{17:[1,36]},e(g,[2,24]),{11:37,13:4,14:22,15:h,16:d,19:5,20:6,21:t,22:8,23:9,24:10,25:11,26:12,28:i,29:s,31:r,32:l,40:n,44:o,47:a},{30:[1,38]},{42:[1,39]},{42:[1,40]},{46:[1,41]},{49:[1,42]},e(y,[2,30]),{18:[1,43]},{18:[1,44]},e(g,[2,23]),{18:[1,45]},{30:[1,46]},e(u,[2,28]),e(u,[2,35]),e(u,[2,36]),e(u,[2,37]),e(u,[2,38]),{37:[1,47]},{34:48,35:p},{15:[1,50]},e(u,[2,27]),e(y,[2,33]),{39:[1,51]},{34:52,35:p,39:[2,31]},{32:[2,15]},e(y,[2,34]),{39:[2,32]}],defaultActions:{20:[2,7],23:[2,14],50:[2,15],52:[2,32]},parseError:function(e,t){if(t.recoverable)this.trace(e);else{var i=Error(e);throw i.hash=t,i}},parse:function(e){var t=this,i=[0],s=[],r=[null],l=[],n=this.table,o="",a=0,c=0,u=l.slice.call(arguments,1),h=Object.create(this.lexer),d={yy:{}};for(var g in this.yy)Object.prototype.hasOwnProperty.call(this.yy,g)&&(d.yy[g]=this.yy[g]);h.setInput(e,d.yy),d.yy.lexer=h,d.yy.parser=this,void 0===h.yylloc&&(h.yylloc={});var y=h.yylloc;l.push(y);var p=h.options&&h.options.ranges;"function"==typeof d.yy.parseError?this.parseError=d.yy.parseError:this.parseError=Object.getPrototypeOf(this).parseError;for(var b,x,S,L,f,_,m,k,E={};;){if(x=i[i.length-1],this.defaultActions[x]?S=this.defaultActions[x]:(null==b&&(b=function(){var e;return"number"!=typeof(e=s.pop()||h.lex()||1)&&(e instanceof Array&&(e=(s=e).pop()),e=t.symbols_[e]||e),e}()),S=n[x]&&n[x][b]),void 0===S||!S.length||!S[0]){var w="";for(f in k=[],n[x])this.terminals_[f]&&f>2&&k.push("'"+this.terminals_[f]+"'");w=h.showPosition?"Parse error on line "+(a+1)+":\n"+h.showPosition()+"\nExpecting "+k.join(", ")+", got '"+(this.terminals_[b]||b)+"'":"Parse error on line "+(a+1)+": Unexpected "+(1==b?"end of input":"'"+(this.terminals_[b]||b)+"'"),this.parseError(w,{text:h.match,token:this.terminals_[b]||b,line:h.yylineno,loc:y,expected:k})}if(S[0]instanceof Array&&S.length>1)throw Error("Parse Error: multiple actions possible at state: "+x+", token: "+b);switch(S[0]){case 1:i.push(b),r.push(h.yytext),l.push(h.yylloc),i.push(S[1]),b=null,c=h.yyleng,o=h.yytext,a=h.yylineno,y=h.yylloc;break;case 2:if(_=this.productions_[S[1]][1],E.$=r[r.length-_],E._$={first_line:l[l.length-(_||1)].first_line,last_line:l[l.length-1].last_line,first_column:l[l.length-(_||1)].first_column,last_column:l[l.length-1].last_column},p&&(E._$.range=[l[l.length-(_||1)].range[0],l[l.length-1].range[1]]),void 0!==(L=this.performAction.apply(E,[o,c,a,d.yy,S[1],r,l].concat(u))))return L;_&&(i=i.slice(0,-1*_*2),r=r.slice(0,-1*_),l=l.slice(0,-1*_)),i.push(this.productions_[S[1]][0]),r.push(E.$),l.push(E._$),m=n[i[i.length-2]][i[i.length-1]],i.push(m);break;case 3:return!0}}return!0}};function x(){this.yy={}}return b.lexer={EOF:1,parseError:function(e,t){if(this.yy.parser)this.yy.parser.parseError(e,t);else throw Error(e)},setInput:function(e,t){return this.yy=t||this.yy||{},this._input=e,this._more=this._backtrack=this.done=!1,this.yylineno=this.yyleng=0,this.yytext=this.matched=this.match="",this.conditionStack=["INITIAL"],this.yylloc={first_line:1,first_column:0,last_line:1,last_column:0},this.options.ranges&&(this.yylloc.range=[0,0]),this.offset=0,this},input:function(){var e=this._input[0];return this.yytext+=e,this.yyleng++,this.offset++,this.match+=e,this.matched+=e,e.match(/(?:\r\n?|\n).*/g)?(this.yylineno++,this.yylloc.last_line++):this.yylloc.last_column++,this.options.ranges&&this.yylloc.range[1]++,this._input=this._input.slice(1),e},unput:function(e){var t=e.length,i=e.split(/(?:\r\n?|\n)/g);this._input=e+this._input,this.yytext=this.yytext.substr(0,this.yytext.length-t),this.offset-=t;var s=this.match.split(/(?:\r\n?|\n)/g);this.match=this.match.substr(0,this.match.length-1),this.matched=this.matched.substr(0,this.matched.length-1),i.length-1&&(this.yylineno-=i.length-1);var r=this.yylloc.range;return this.yylloc={first_line:this.yylloc.first_line,last_line:this.yylineno+1,first_column:this.yylloc.first_column,last_column:i?(i.length===s.length?this.yylloc.first_column:0)+s[s.length-i.length].length-i[0].length:this.yylloc.first_column-t},this.options.ranges&&(this.yylloc.range=[r[0],r[0]+this.yyleng-t]),this.yyleng=this.yytext.length,this},more:function(){return this._more=!0,this},reject:function(){return this.options.backtrack_lexer?(this._backtrack=!0,this):this.parseError("Lexical error on line "+(this.yylineno+1)+". You can only invoke reject() in the lexer when the lexer is of the backtracking persuasion (options.backtrack_lexer = true).\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},less:function(e){this.unput(this.match.slice(e))},pastInput:function(){var e=this.matched.substr(0,this.matched.length-this.match.length);return(e.length>20?"...":"")+e.substr(-20).replace(/\n/g,"")},upcomingInput:function(){var e=this.match;return e.length<20&&(e+=this._input.substr(0,20-e.length)),(e.substr(0,20)+(e.length>20?"...":"")).replace(/\n/g,"")},showPosition:function(){var e=this.pastInput(),t=Array(e.length+1).join("-");return e+this.upcomingInput()+"\n"+t+"^"},test_match:function(e,t){var i,s,r;if(this.options.backtrack_lexer&&(r={yylineno:this.yylineno,yylloc:{first_line:this.yylloc.first_line,last_line:this.last_line,first_column:this.yylloc.first_column,last_column:this.yylloc.last_column},yytext:this.yytext,match:this.match,matches:this.matches,matched:this.matched,yyleng:this.yyleng,offset:this.offset,_more:this._more,_input:this._input,yy:this.yy,conditionStack:this.conditionStack.slice(0),done:this.done},this.options.ranges&&(r.yylloc.range=this.yylloc.range.slice(0))),(s=e[0].match(/(?:\r\n?|\n).*/g))&&(this.yylineno+=s.length),this.yylloc={first_line:this.yylloc.last_line,last_line:this.yylineno+1,first_column:this.yylloc.last_column,last_column:s?s[s.length-1].length-s[s.length-1].match(/\r?\n?/)[0].length:this.yylloc.last_column+e[0].length},this.yytext+=e[0],this.match+=e[0],this.matches=e,this.yyleng=this.yytext.length,this.options.ranges&&(this.yylloc.range=[this.offset,this.offset+=this.yyleng]),this._more=!1,this._backtrack=!1,this._input=this._input.slice(e[0].length),this.matched+=e[0],i=this.performAction.call(this,this.yy,this,t,this.conditionStack[this.conditionStack.length-1]),this.done&&this._input&&(this.done=!1),i)return i;if(this._backtrack)for(var l in r)this[l]=r[l];return!1},next:function(){if(this.done)return this.EOF;this._input||(this.done=!0),this._more||(this.yytext="",this.match="");for(var e,t,i,s,r=this._currentRules(),l=0;l<r.length;l++)if((i=this._input.match(this.rules[r[l]]))&&(!t||i[0].length>t[0].length)){if(t=i,s=l,this.options.backtrack_lexer){if(!1!==(e=this.test_match(i,r[l])))return e;if(!this._backtrack)return!1;t=!1;continue}if(!this.options.flex)break}return t?!1!==(e=this.test_match(t,r[s]))&&e:""===this._input?this.EOF:this.parseError("Lexical error on line "+(this.yylineno+1)+". Unrecognized text.\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},lex:function(){return this.next()||this.lex()},begin:function(e){this.conditionStack.push(e)},popState:function(){return this.conditionStack.length-1>0?this.conditionStack.pop():this.conditionStack[0]},_currentRules:function(){return this.conditionStack.length&&this.conditionStack[this.conditionStack.length-1]?this.conditions[this.conditionStack[this.conditionStack.length-1]].rules:this.conditions.INITIAL.rules},topState:function(e){return(e=this.conditionStack.length-1-Math.abs(e||0))>=0?this.conditionStack[e]:"INITIAL"},pushState:function(e){this.begin(e)},stateStackSize:function(){return this.conditionStack.length},options:{},performAction:function(e,t,i,s){switch(i){case 0:return 10;case 1:return e.getLogger().debug("Found space-block"),31;case 2:return e.getLogger().debug("Found nl-block"),31;case 3:return e.getLogger().debug("Found space-block"),29;case 4:e.getLogger().debug(".",t.yytext);break;case 5:e.getLogger().debug("_",t.yytext);break;case 6:return 5;case 7:return t.yytext=-1,28;case 8:return t.yytext=t.yytext.replace(/columns\s+/,""),e.getLogger().debug("COLUMNS (LEX)",t.yytext),28;case 9:case 77:case 78:case 100:this.pushState("md_string");break;case 10:return"MD_STR";case 11:case 35:case 80:this.popState();break;case 12:this.pushState("string");break;case 13:e.getLogger().debug("LEX: POPPING STR:",t.yytext),this.popState();break;case 14:return e.getLogger().debug("LEX: STR end:",t.yytext),"STR";case 15:return t.yytext=t.yytext.replace(/space\:/,""),e.getLogger().debug("SPACE NUM (LEX)",t.yytext),21;case 16:return t.yytext="1",e.getLogger().debug("COLUMNS (LEX)",t.yytext),21;case 17:return 43;case 18:return"LINKSTYLE";case 19:return"INTERPOLATE";case 20:return this.pushState("CLASSDEF"),40;case 21:return this.popState(),this.pushState("CLASSDEFID"),"DEFAULT_CLASSDEF_ID";case 22:return this.popState(),this.pushState("CLASSDEFID"),41;case 23:return this.popState(),42;case 24:return this.pushState("CLASS"),44;case 25:return this.popState(),this.pushState("CLASS_STYLE"),45;case 26:return this.popState(),46;case 27:return this.pushState("STYLE_STMNT"),47;case 28:return this.popState(),this.pushState("STYLE_DEFINITION"),48;case 29:return this.popState(),49;case 30:return this.pushState("acc_title"),"acc_title";case 31:return this.popState(),"acc_title_value";case 32:return this.pushState("acc_descr"),"acc_descr";case 33:return this.popState(),"acc_descr_value";case 34:this.pushState("acc_descr_multiline");break;case 36:return"acc_descr_multiline_value";case 37:return 30;case 38:case 39:case 41:case 42:case 45:return this.popState(),e.getLogger().debug("Lex: (("),"NODE_DEND";case 40:return this.popState(),e.getLogger().debug("Lex: ))"),"NODE_DEND";case 43:return this.popState(),e.getLogger().debug("Lex: (-"),"NODE_DEND";case 44:return this.popState(),e.getLogger().debug("Lex: -)"),"NODE_DEND";case 46:return this.popState(),e.getLogger().debug("Lex: ]]"),"NODE_DEND";case 47:return this.popState(),e.getLogger().debug("Lex: ("),"NODE_DEND";case 48:return this.popState(),e.getLogger().debug("Lex: ])"),"NODE_DEND";case 49:case 50:return this.popState(),e.getLogger().debug("Lex: /]"),"NODE_DEND";case 51:return this.popState(),e.getLogger().debug("Lex: )]"),"NODE_DEND";case 52:return this.popState(),e.getLogger().debug("Lex: )"),"NODE_DEND";case 53:return this.popState(),e.getLogger().debug("Lex: ]>"),"NODE_DEND";case 54:return this.popState(),e.getLogger().debug("Lex: ]"),"NODE_DEND";case 55:return e.getLogger().debug("Lexa: -)"),this.pushState("NODE"),36;case 56:return e.getLogger().debug("Lexa: (-"),this.pushState("NODE"),36;case 57:return e.getLogger().debug("Lexa: ))"),this.pushState("NODE"),36;case 58:case 60:case 61:case 62:case 65:return e.getLogger().debug("Lexa: )"),this.pushState("NODE"),36;case 59:return e.getLogger().debug("Lex: ((("),this.pushState("NODE"),36;case 63:return e.getLogger().debug("Lexc: >"),this.pushState("NODE"),36;case 64:return e.getLogger().debug("Lexa: (["),this.pushState("NODE"),36;case 66:case 67:case 68:case 69:case 70:case 71:case 72:return this.pushState("NODE"),36;case 73:return e.getLogger().debug("Lexa: ["),this.pushState("NODE"),36;case 74:return this.pushState("BLOCK_ARROW"),e.getLogger().debug("LEX ARR START"),38;case 75:return e.getLogger().debug("Lex: NODE_ID",t.yytext),32;case 76:return e.getLogger().debug("Lex: EOF",t.yytext),8;case 79:return"NODE_DESCR";case 81:e.getLogger().debug("Lex: Starting string"),this.pushState("string");break;case 82:e.getLogger().debug("LEX ARR: Starting string"),this.pushState("string");break;case 83:return e.getLogger().debug("LEX: NODE_DESCR:",t.yytext),"NODE_DESCR";case 84:e.getLogger().debug("LEX POPPING"),this.popState();break;case 85:e.getLogger().debug("Lex: =>BAE"),this.pushState("ARROW_DIR");break;case 86:return t.yytext=t.yytext.replace(/^,\s*/,""),e.getLogger().debug("Lex (right): dir:",t.yytext),"DIR";case 87:return t.yytext=t.yytext.replace(/^,\s*/,""),e.getLogger().debug("Lex (left):",t.yytext),"DIR";case 88:return t.yytext=t.yytext.replace(/^,\s*/,""),e.getLogger().debug("Lex (x):",t.yytext),"DIR";case 89:return t.yytext=t.yytext.replace(/^,\s*/,""),e.getLogger().debug("Lex (y):",t.yytext),"DIR";case 90:return t.yytext=t.yytext.replace(/^,\s*/,""),e.getLogger().debug("Lex (up):",t.yytext),"DIR";case 91:return t.yytext=t.yytext.replace(/^,\s*/,""),e.getLogger().debug("Lex (down):",t.yytext),"DIR";case 92:return t.yytext="]>",e.getLogger().debug("Lex (ARROW_DIR end):",t.yytext),this.popState(),this.popState(),"BLOCK_ARROW_END";case 93:return e.getLogger().debug("Lex: LINK","#"+t.yytext+"#"),15;case 94:case 95:case 96:return e.getLogger().debug("Lex: LINK",t.yytext),15;case 97:case 98:case 99:return e.getLogger().debug("Lex: START_LINK",t.yytext),this.pushState("LLABEL"),16;case 101:return e.getLogger().debug("Lex: Starting string"),this.pushState("string"),"LINK_LABEL";case 102:return this.popState(),e.getLogger().debug("Lex: LINK","#"+t.yytext+"#"),15;case 103:case 104:return this.popState(),e.getLogger().debug("Lex: LINK",t.yytext),15;case 105:return e.getLogger().debug("Lex: COLON",t.yytext),t.yytext=t.yytext.slice(1),27}},rules:[/^(?:block-beta\b)/,/^(?:block\s+)/,/^(?:block\n+)/,/^(?:block:)/,/^(?:[\s]+)/,/^(?:[\n]+)/,/^(?:((\u000D\u000A)|(\u000A)))/,/^(?:columns\s+auto\b)/,/^(?:columns\s+[\d]+)/,/^(?:["][`])/,/^(?:[^`"]+)/,/^(?:[`]["])/,/^(?:["])/,/^(?:["])/,/^(?:[^"]*)/,/^(?:space[:]\d+)/,/^(?:space\b)/,/^(?:default\b)/,/^(?:linkStyle\b)/,/^(?:interpolate\b)/,/^(?:classDef\s+)/,/^(?:DEFAULT\s+)/,/^(?:\w+\s+)/,/^(?:[^\n]*)/,/^(?:class\s+)/,/^(?:(\w+)+((,\s*\w+)*))/,/^(?:[^\n]*)/,/^(?:style\s+)/,/^(?:(\w+)+((,\s*\w+)*))/,/^(?:[^\n]*)/,/^(?:accTitle\s*:\s*)/,/^(?:(?!\n||)*[^\n]*)/,/^(?:accDescr\s*:\s*)/,/^(?:(?!\n||)*[^\n]*)/,/^(?:accDescr\s*\{\s*)/,/^(?:[\}])/,/^(?:[^\}]*)/,/^(?:end\b\s*)/,/^(?:\(\(\()/,/^(?:\)\)\))/,/^(?:[\)]\))/,/^(?:\}\})/,/^(?:\})/,/^(?:\(-)/,/^(?:-\))/,/^(?:\(\()/,/^(?:\]\])/,/^(?:\()/,/^(?:\]\))/,/^(?:\\\])/,/^(?:\/\])/,/^(?:\)\])/,/^(?:[\)])/,/^(?:\]>)/,/^(?:[\]])/,/^(?:-\))/,/^(?:\(-)/,/^(?:\)\))/,/^(?:\))/,/^(?:\(\(\()/,/^(?:\(\()/,/^(?:\{\{)/,/^(?:\{)/,/^(?:>)/,/^(?:\(\[)/,/^(?:\()/,/^(?:\[\[)/,/^(?:\[\|)/,/^(?:\[\()/,/^(?:\)\)\))/,/^(?:\[\\)/,/^(?:\[\/)/,/^(?:\[\\)/,/^(?:\[)/,/^(?:<\[)/,/^(?:[^\(\[\n\-\)\{\}\s\<\>:]+)/,/^(?:$)/,/^(?:["][`])/,/^(?:["][`])/,/^(?:[^`"]+)/,/^(?:[`]["])/,/^(?:["])/,/^(?:["])/,/^(?:[^"]+)/,/^(?:["])/,/^(?:\]>\s*\()/,/^(?:,?\s*right\s*)/,/^(?:,?\s*left\s*)/,/^(?:,?\s*x\s*)/,/^(?:,?\s*y\s*)/,/^(?:,?\s*up\s*)/,/^(?:,?\s*down\s*)/,/^(?:\)\s*)/,/^(?:\s*[xo<]?--+[-xo>]\s*)/,/^(?:\s*[xo<]?==+[=xo>]\s*)/,/^(?:\s*[xo<]?-?\.+-[xo>]?\s*)/,/^(?:\s*~~[\~]+\s*)/,/^(?:\s*[xo<]?--\s*)/,/^(?:\s*[xo<]?==\s*)/,/^(?:\s*[xo<]?-\.\s*)/,/^(?:["][`])/,/^(?:["])/,/^(?:\s*[xo<]?--+[-xo>]\s*)/,/^(?:\s*[xo<]?==+[=xo>]\s*)/,/^(?:\s*[xo<]?-?\.+-[xo>]?\s*)/,/^(?::\d+)/],conditions:{STYLE_DEFINITION:{rules:[29],inclusive:!1},STYLE_STMNT:{rules:[28],inclusive:!1},CLASSDEFID:{rules:[23],inclusive:!1},CLASSDEF:{rules:[21,22],inclusive:!1},CLASS_STYLE:{rules:[26],inclusive:!1},CLASS:{rules:[25],inclusive:!1},LLABEL:{rules:[100,101,102,103,104],inclusive:!1},ARROW_DIR:{rules:[86,87,88,89,90,91,92],inclusive:!1},BLOCK_ARROW:{rules:[77,82,85],inclusive:!1},NODE:{rules:[38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,78,81],inclusive:!1},md_string:{rules:[10,11,79,80],inclusive:!1},space:{rules:[],inclusive:!1},string:{rules:[13,14,83,84],inclusive:!1},acc_descr_multiline:{rules:[35,36],inclusive:!1},acc_descr:{rules:[33],inclusive:!1},acc_title:{rules:[31],inclusive:!1},INITIAL:{rules:[0,1,2,3,4,5,6,7,8,9,12,15,16,17,18,19,20,24,27,30,32,34,37,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,93,94,95,96,97,98,99,105],inclusive:!0}}},x.prototype=b,b.Parser=x,new x}();d.parser=d;let g={},y=[],p={},b="color",x="fill",S=(0,l.c)(),L={},f=e=>l.e.sanitizeText(e,S),_=function(e,t=""){void 0===L[e]&&(L[e]={id:e,styles:[],textStyles:[]});let i=L[e];null!=t&&t.split(",").forEach(e=>{let t=e.replace(/([^;]*);/,"$1").trim();if(e.match(b)){let e=t.replace(x,"bgFill").replace(b,x);i.textStyles.push(e)}i.styles.push(t)})},m=function(e,t=""){let i=g[e];null!=t&&(i.styles=t.split(","))},k=function(e,t){e.split(",").forEach(function(e){let i=g[e];if(void 0===i){let t=e.trim();g[t]={id:t,type:"na",children:[]},i=g[t]}i.classes||(i.classes=[]),i.classes.push(t)})},E=(e,t)=>{let i=e.flat(),s=[];for(let e of i){if(e.label&&(e.label=f(e.label)),"classDef"===e.type){_(e.id,e.css);continue}if("applyClass"===e.type){k(e.id,(null==e?void 0:e.styleClass)||"");continue}if("applyStyles"===e.type){(null==e?void 0:e.stylesStr)&&m(e.id,null==e?void 0:e.stylesStr);continue}if("column-setting"===e.type)t.columns=e.columns||-1;else if("edge"===e.type)p[e.id]?p[e.id]++:p[e.id]=1,e.id=p[e.id]+"-"+e.id,y.push(e);else{e.label||("composite"===e.type?e.label="":e.label=e.id);let t=!g[e.id];if(t?g[e.id]=e:("na"!==e.type&&(g[e.id].type=e.type),e.label!==e.id&&(g[e.id].label=e.label)),e.children&&E(e.children,e),"space"===e.type){let t=e.width||1;for(let i=0;i<t;i++){let t=(0,n.Z)(e);t.id=t.id+"-"+i,g[t.id]=t,s.push(t)}}else t&&s.push(e)}}t.children=s},w=[],v={id:"root",type:"composite",children:[],columns:-1},D=0,$=(e,t)=>{let i=o.Z,s=i(e,"r"),r=i(e,"g"),l=i(e,"b");return a.Z(s,r,l,t)};function N(e,t,i=!1){var s,r,n;let o;let a="default";((null==(s=null==e?void 0:e.classes)?void 0:s.length)||0)>0&&(a=((null==e?void 0:e.classes)||[]).join(" ")),a+=" flowchart-label";let c=0,u="";switch(e.type){case"round":c=5,u="rect";break;case"composite":c=0,u="composite",o=0;break;case"square":case"group":default:u="rect";break;case"diamond":u="question";break;case"hexagon":u="hexagon";break;case"block_arrow":u="block_arrow";break;case"odd":case"rect_left_inv_arrow":u="rect_left_inv_arrow";break;case"lean_right":u="lean_right";break;case"lean_left":u="lean_left";break;case"trapezoid":u="trapezoid";break;case"inv_trapezoid":u="inv_trapezoid";break;case"circle":u="circle";break;case"ellipse":u="ellipse";break;case"stadium":u="stadium";break;case"subroutine":u="subroutine";break;case"cylinder":u="cylinder";break;case"doublecircle":u="doublecircle"}let h=(0,l.k)((null==e?void 0:e.styles)||[]),d=e.label,g=e.size||{width:0,height:0,x:0,y:0};return{labelStyle:h.labelStyle,shape:u,labelText:d,rx:c,ry:c,class:a,style:h.style,id:e.id,directions:e.directions,width:g.width,height:g.height,x:g.x,y:g.y,positioned:i,intersect:void 0,type:e.type,padding:o??((null==(n=null==(r=(0,l.F)())?void 0:r.block)?void 0:n.padding)||0)}}async function I(e,t,i){let s=N(t,i,!1);if("group"===s.type)return;let r=await (0,c.e)(e,s),l=r.node().getBBox(),n=i.getBlock(s.id);n.size={width:l.width,height:l.height,x:0,y:0,node:r},i.setBlock(n),r.remove()}async function T(e,t,i){let s=N(t,i,!0);"space"!==i.getBlock(s.id).type&&(await (0,c.e)(e,s),t.intersect=null==s?void 0:s.intersect,(0,c.p)(s))}async function z(e,t,i,s){for(let r of t)await s(e,r,i),r.children&&await z(e,r.children,i,s)}async function C(e,t,i){await z(e,t,i,I)}async function O(e,t,i){await z(e,t,i,T)}async function A(e,t,i,s,r){let l=new u.k({multigraph:!0,compound:!0});for(let e of(l.setGraph({rankdir:"TB",nodesep:10,ranksep:10,marginx:8,marginy:8}),i))e.size&&l.setNode(e.id,{width:e.size.width,height:e.size.height,intersect:e.intersect});for(let i of t)if(i.start&&i.end){let t=s.getBlock(i.start),n=s.getBlock(i.end);if((null==t?void 0:t.size)&&(null==n?void 0:n.size)){let s=t.size,o=n.size,a=[{x:s.x,y:s.y},{x:s.x+(o.x-s.x)/2,y:s.y+(o.y-s.y)/2},{x:o.x,y:o.y}];await (0,c.h)(e,{v:i.start,w:i.end,name:i.id},{...i,arrowTypeEnd:i.arrowTypeEnd,arrowTypeStart:i.arrowTypeStart,points:a,classes:"edge-thickness-normal edge-pattern-solid flowchart-link LS-a1 LE-b1"},void 0,"block",l,r),i.label&&(await (0,c.f)(e,{...i,label:i.label,labelStyle:"stroke: #333; stroke-width: 1.5px;fill:none;",arrowTypeEnd:i.arrowTypeEnd,arrowTypeStart:i.arrowTypeStart,points:a,classes:"edge-thickness-normal edge-pattern-solid flowchart-link LS-a1 LE-b1"}),await (0,c.j)({...i,x:a[1].x,y:a[1].y},{originalPath:a}))}}}let R=(null==(r=null==(s=(0,l.c)())?void 0:s.block)?void 0:r.padding)||8,B=e=>{let t=0,i=0;for(let s of e.children){let{width:r,height:n,x:o,y:a}=s.size||{width:0,height:0,x:0,y:0};l.l.debug("getMaxChildSize abc95 child:",s.id,"width:",r,"height:",n,"x:",o,"y:",a,s.type),"space"!==s.type&&(r>t&&(t=r/(e.widthInColumns||1)),n>i&&(i=n))}return{width:t,height:i}},F={parser:d,db:{getConfig:()=>(0,l.F)().block,typeStr2Type:function(e){switch(l.l.debug("typeStr2Type",e),e){case"[]":return"square";case"()":return l.l.debug("we have a round"),"round";case"(())":return"circle";case">]":return"rect_left_inv_arrow";case"{}":return"diamond";case"{{}}":return"hexagon";case"([])":return"stadium";case"[[]]":return"subroutine";case"[()]":return"cylinder";case"((()))":return"doublecircle";case"[//]":return"lean_right";case"[\\\\]":return"lean_left";case"[/\\]":return"trapezoid";case"[\\/]":return"inv_trapezoid";case"<[]>":return"block_arrow";default:return"na"}},edgeTypeStr2Type:function(e){return(l.l.debug("typeStr2Type",e),"=="===e)?"thick":"normal"},edgeStrToEdgeData:function(e){switch(e.trim()){case"--x":return"arrow_cross";case"--o":return"arrow_circle";default:return"arrow_point"}},getLogger:()=>console,getBlocksFlat:()=>[...Object.values(g)],getBlocks:()=>w||[],getEdges:()=>y,setHierarchy:e=>{v.children=e,E(e,v),w=v.children},getBlock:e=>g[e],setBlock:e=>{g[e.id]=e},getColumns:e=>{let t=g[e];return t?t.columns?t.columns:t.children?t.children.length:-1:-1},getClasses:function(){return L},clear:()=>{l.l.debug("Clear called"),(0,l.v)(),g={root:v={id:"root",type:"composite",children:[],columns:-1}},w=[],L={},y=[],p={}},generateId:()=>(D++,"id-"+Math.random().toString(36).substr(2,12)+"-"+D)},renderer:{draw:async function(e,t,i,s){let r;let{securityLevel:n,block:o}=(0,l.F)(),a=s.db;"sandbox"===n&&(r=(0,h.Ys)("#i"+t));let u="sandbox"===n?(0,h.Ys)(r.nodes()[0].contentDocument.body):(0,h.Ys)("body"),d="sandbox"===n?u.select(`[id="${t}"]`):(0,h.Ys)(`[id="${t}"]`);(0,c.a)(d,["point","circle","cross"],s.type,t);let g=a.getBlocks(),y=a.getBlocksFlat(),p=a.getEdges(),b=d.insert("g").attr("class","block");await C(b,g,a);let x=function(e){let t=e.getBlock("root");if(!t)return;!function e(t,i,s=0,r=0){var n,o,a,c,u,h,d,g,y,p,b;l.l.debug("setBlockSizes abc95 (start)",t.id,null==(n=null==t?void 0:t.size)?void 0:n.x,"block width =",null==t?void 0:t.size,"sieblingWidth",s),(null==(o=null==t?void 0:t.size)?void 0:o.width)||(t.size={width:s,height:r,x:0,y:0});let x=0,S=0;if((null==(a=t.children)?void 0:a.length)>0){for(let s of t.children)e(s,i);let n=B(t);for(let e of(x=n.width,S=n.height,l.l.debug("setBlockSizes abc95 maxWidth of",t.id,":s children is ",x,S),t.children))e.size&&(l.l.debug(`abc95 Setting size of children of ${t.id} id=${e.id} ${x} ${S} ${e.size}`),e.size.width=x*(e.widthInColumns||1)+R*((e.widthInColumns||1)-1),e.size.height=S,e.size.x=0,e.size.y=0,l.l.debug(`abc95 updating size of ${t.id} children child:${e.id} maxWidth:${x} maxHeight:${S}`));for(let s of t.children)e(s,i,x,S);let o=t.columns||-1,a=0;for(let e of t.children)a+=e.widthInColumns||1;let g=t.children.length;o>0&&o<a&&(g=o),t.widthInColumns;let y=Math.ceil(a/g),p=g*(x+R)+R,b=y*(S+R)+R;if(p<s){l.l.debug(`Detected to small siebling: abc95 ${t.id} sieblingWidth ${s} sieblingHeight ${r} width ${p}`),p=s,b=r;let e=(s-g*R-R)/g,i=(r-y*R-R)/y;for(let s of(l.l.debug("Size indata abc88",t.id,"childWidth",e,"maxWidth",x),l.l.debug("Size indata abc88",t.id,"childHeight",i,"maxHeight",S),l.l.debug("Size indata abc88 xSize",g,"padding",R),t.children))s.size&&(s.size.width=e,s.size.height=i,s.size.x=0,s.size.y=0)}if(l.l.debug(`abc95 (finale calc) ${t.id} xSize ${g} ySize ${y} columns ${o}${t.children.length} width=${Math.max(p,(null==(c=t.size)?void 0:c.width)||0)}`),p<((null==(u=null==t?void 0:t.size)?void 0:u.width)||0)){p=(null==(h=null==t?void 0:t.size)?void 0:h.width)||0;let e=o>0?Math.min(t.children.length,o):t.children.length;if(e>0){let i=(p-e*R-R)/e;for(let e of(l.l.debug("abc95 (growing to fit) width",t.id,p,null==(d=t.size)?void 0:d.width,i),t.children))e.size&&(e.size.width=i)}}t.size={width:p,height:b,x:0,y:0}}l.l.debug("setBlockSizes abc94 (done)",t.id,null==(g=null==t?void 0:t.size)?void 0:g.x,null==(y=null==t?void 0:t.size)?void 0:y.width,null==(p=null==t?void 0:t.size)?void 0:p.y,null==(b=null==t?void 0:t.size)?void 0:b.height)}(t,e,0,0),function e(t,i){var s,r,n,o,a,c,u,h,d,g,y,p,b,x,S,L,f;l.l.debug(`abc85 layout blocks (=>layoutBlocks) ${t.id} x: ${null==(s=null==t?void 0:t.size)?void 0:s.x} y: ${null==(r=null==t?void 0:t.size)?void 0:r.y} width: ${null==(n=null==t?void 0:t.size)?void 0:n.width}`);let _=t.columns||-1;if(l.l.debug("layoutBlocks columns abc95",t.id,"=>",_,t),t.children&&t.children.length>0){let i=(null==(a=null==(o=null==t?void 0:t.children[0])?void 0:o.size)?void 0:a.width)||0,s=t.children.length*i+(t.children.length-1)*R;l.l.debug("widthOfChildren 88",s,"posX");let r=0;l.l.debug("abc91 block?.size?.x",t.id,null==(c=null==t?void 0:t.size)?void 0:c.x);let n=(null==(u=null==t?void 0:t.size)?void 0:u.x)?(null==(h=null==t?void 0:t.size)?void 0:h.x)+(-(null==(d=null==t?void 0:t.size)?void 0:d.width)/2||0):-R,S=0;for(let i of t.children){if(!i.size)continue;let{width:s,height:o}=i.size,{px:a,py:c}=function(e,t){if(0===e||!Number.isInteger(e))throw Error("Columns must be an integer !== 0.");if(t<0||!Number.isInteger(t))throw Error("Position must be a non-negative integer."+t);if(e<0)return{px:t,py:0};if(1===e)return{px:0,py:t};let i=Math.floor(t/e);return{px:t%e,py:i}}(_,r);if(c!=S&&(S=c,n=(null==(g=null==t?void 0:t.size)?void 0:g.x)?(null==(y=null==t?void 0:t.size)?void 0:y.x)+(-(null==(p=null==t?void 0:t.size)?void 0:p.width)/2||0):-R,l.l.debug("New row in layout for block",t.id," and child ",i.id,S)),l.l.debug(`abc89 layout blocks (child) id: ${i.id} Pos: ${r} (px, py) ${a},${c} (${null==(b=null==t?void 0:t.size)?void 0:b.x},${null==(x=null==t?void 0:t.size)?void 0:x.y}) parent: ${t.id} width: ${s}${R}`),t.size){let e=s/2;i.size.x=n+R+e,l.l.debug(`abc91 layout blocks (calc) px, pyid:${i.id} startingPos=X${n} new startingPosX${i.size.x} ${e} padding=${R} width=${s} halfWidth=${e} => x:${i.size.x} y:${i.size.y} ${i.widthInColumns} (width * (child?.w || 1)) / 2 ${s*((null==i?void 0:i.widthInColumns)||1)/2}`),n=i.size.x+e,i.size.y=t.size.y-t.size.height/2+c*(o+R)+o/2+R,l.l.debug(`abc88 layout blocks (calc) px, pyid:${i.id}startingPosX${n}${R}${e}=>x:${i.size.x}y:${i.size.y}${i.widthInColumns}(width * (child?.w || 1)) / 2${s*((null==i?void 0:i.widthInColumns)||1)/2}`)}i.children&&e(i),r+=(null==i?void 0:i.widthInColumns)||1,l.l.debug("abc88 columnsPos",i,r)}}l.l.debug(`layout blocks (<==layoutBlocks) ${t.id} x: ${null==(S=null==t?void 0:t.size)?void 0:S.x} y: ${null==(L=null==t?void 0:t.size)?void 0:L.y} width: ${null==(f=null==t?void 0:t.size)?void 0:f.width}`)}(t),l.l.debug("getBlocks",JSON.stringify(t,null,2));let{minX:i,minY:s,maxX:r,maxY:n}=function e(t,{minX:i,minY:s,maxX:r,maxY:l}={minX:0,minY:0,maxX:0,maxY:0}){if(t.size&&"root"!==t.id){let{x:e,y:n,width:o,height:a}=t.size;e-o/2<i&&(i=e-o/2),n-a/2<s&&(s=n-a/2),e+o/2>r&&(r=e+o/2),n+a/2>l&&(l=n+a/2)}if(t.children)for(let n of t.children)({minX:i,minY:s,maxX:r,maxY:l}=e(n,{minX:i,minY:s,maxX:r,maxY:l}));return{minX:i,minY:s,maxX:r,maxY:l}}(t);return{x:i,y:s,width:r-i,height:n-s}}(a);if(await O(b,g,a),await A(b,p,y,a,t),x){let e=Math.max(1,Math.round(.125*(x.width/x.height))),t=x.height+e+10,i=x.width+10,{useMaxWidth:s}=o;(0,l.i)(d,t,i,!!s),l.l.debug("Here Bounds",x,x),d.attr("viewBox",`${x.x-5} ${x.y-5} ${x.width+10} ${x.height+10}`)}(0,h.PKp)(h.K2I)},getClasses:function(e,t){return t.db.getClasses()}},styles:e=>`.label {
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
    background-color: ${$(e.edgeLabelBackground,.5)};
    // background-color:
  }

  .node .cluster {
    // fill: ${$(e.mainBkg,.5)};
    fill: ${$(e.clusterBkg,.5)};
    stroke: ${$(e.clusterBorder,.2)};
    box-shadow: rgba(50, 50, 93, 0.25) 0px 13px 27px -5px, rgba(0, 0, 0, 0.3) 0px 8px 16px -8px;
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
`}}}]);