<!DOCTYPE html>
<html>
<head>
   <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <title>Gentlesys {{.Title}}</title>
  <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
   <style>
	.key-prob {
		color:#009900;
		font-size:20px;
		padding-right:20px;
	}
	.edit_prob {
		color:#009900;
	}
	.des-center {
	    text-align: center;
	    font-size:20px;
	    color:#4876FF;
	    padding-bottom:5px;
	}
	.body-css {
     padding-top:55px;
	 background-color:#fcfdf8;
	}
	</style>
</head>
<body class="body-css">
{{str2html .Navigation}}
<div class="container-fluid">
<div class="row">
    <hr>
    <div class="col-md-10 col-md-offset-2">
    <p id="info" class="alert"><strong>提示:{{.Info}}</strong></p>
    <p style="padding:5px 15px;">
    <span><strong>请选择在哪个版块主题中查找</strong></span>
    <span><select id="subType" class="select">
	{{range .SubType}}
    <option value="{{.UniqueId}}">{{.Name}}</option>
    {{end}}
    </select>
    </span>
    </p>
    </div>
    
    <div class="col-md-10 col-md-offset-2">
    <form class="form-horizontal" role="form">
    <div class="col-md-6">
    <div class="input-group">
    <div class="input-group-btn">
    
    <select id="sType" onchange="gradeChange()" class="form-control" style="
    width:140px;
    padding:3px 3px;
    border-top-left-radius: 5px;
    border-bottom-left-radius: 5px;
    background-color:#EFEFEF;
    background-size:13px 13px;
    appearance:none;
    -moz-appearance:none;
    -webkit-appearance:none;
    ">
    <option value="1">按用户名查找帖子</option>
    <option value="2">按日期时间查找帖子</option>
    <option value="3">按用户名查找回复</option>
    </select>
    </div>
    <input id="intext" type="text" class="form-control" placeholder="请输入用户名...">
    <span class="input-group-btn">
    <button class="btn btn-default" type="button" onclick="findinfo()">开始查找</button>
    </span>
    </div>
    </div>
    </form>
    </div>
</div>
<hr/>
<div class="col-md-offset-2" style="padding:1px 15px;">
     <ul class="pagination pagination-sm">
        <li><span class="btn btn-default" role="button">结果索引</span></li>
    	<li><a href="{{.PrePage}}">&laquo;</a></li>
    	 {{range .RecordIndexs}}
        	<li class="{{.IsActive}}"><a href="{{.Ref}}">{{.Title}}</a></li>
    	{{end}}
    	<li><a href="{{.NextPage}}">&raquo;</a></li>
    </ul>
    {{if .IsTopic}}
        {{range .TopicsList}}
            <p>时间{{.Date}}&nbsp&nbsp作者&nbsp&nbsp{{.UserName}}&nbsp&nbsp[标题]&nbsp&nbsp<a href="/browse?sid={{$.Sid}}&aid={{.Id}}" target="_blank">{{.Title}}(阅读次数{{.ReadTimes}}) </a>&nbsp&nbsp&nbsp&nbsp[禁用与否]{{.Disable}}&nbsp&nbsp&nbsp&nbsp <button type="submit" class="btn-xs btn-warning" onclick="disable({{$.Sid}},{{.Id}})">(禁贴或开启)</button></p>
        {{end}}
         <script>
         function disable(sid,aid){
		$.ajax({
	          async:true,
	          cache:false,
	          timeout:10000,
	          type:"POST",
	          url:"/disable",
	          data:{
	          	subid_:sid,
	          	aid_:aid,
		    	},
	          error:function(jqXHR, textStatus, errorThrown){
	            if(textStatus=="timeout"){
	              $("#info").text("操作超时，请重试...");
	            }else{
	              $("#info").text("设置帖子禁用状态失败!");
	            }
	          },
	          success:function(msg){
			  	$("#info").text(msg);
				$("#intext").attr("placeholder",msg)
	          }	            
	        });
		}
        </script>
    {{else}}
        {{range .CommentsList}}
            <hr/>
            <p><a href="/browse?sid={{.SubId}}&aid={{.Aid}}&page={{idtope .Commentdata}}" target="_blank">版块{{.SubId}}&nbsp&nbsp文章id&nbsp&nbsp{{.Aid}}&nbsp&nbsp</a>第#{{.Commentdata.Id}}楼&nbsp&nbsp
            {{if isdeled .Commentdata}}
            <span>回复已被禁用</span>
            {{else}}
            <button type="submit" class="btn-xs btn-warning" onclick="remove({{.SubId}},{{.Aid}},{{$.Uid}},{{.Commentdata.Id}})">(禁该回复)</button></p>
            {{end}}
            <p>时间{{.Commentdata.Time}}&nbsp&nbsp作者&nbsp&nbsp{{.Commentdata.UserName}}&nbsp&nbsp</p>
            <div>[内容]&nbsp&nbsp{{str2html .Commentdata.Content}}</div>
        {{end}}
        <script>
        function remove(sid,aid,uid,cid){
		$.ajax({
	          async:true,
	          cache:false,
	          timeout:10000,
	          type:"POST",
	          url:"/remove",
	          data:{
	          	subId_:sid,
				artiId_:aid,
	          	userId_:uid,
	          	cid_:cid,
	          	pages_:{{.PageNum}},
		    	},
	          error:function(jqXHR, textStatus, errorThrown){
	            if(textStatus=="timeout"){
	              $("#info").text("操作超时，请重试...");
	            }else{
	              $("#info").text("设置帖子禁用状态失败!");
	            }
	          },
	          success:function(msg){
			  	$("#info").text(msg);
				$("#intext").attr("placeholder",msg)
	          }	            
	        });
		}
        </script>
    {{end}}
    <hr/>
 </div>
</div>
<script>
		function findinfo() {
			var options=$("#sType option:selected");
			var type = options.val(); 

			var options=$("#subType option:selected");
			var subId = options.val(); 
		
			var key = document.getElementById("intext").value;
			if (type == 1) {
				if (key.length < 1) {
					$("#intext").attr("placeholder","没有输入用户名称")
					return
				}
			} 
			else if (type == 2)
			{
				if (key.length < 1) {
					$("#intext").attr("placeholder","没有输入日期时间，格式 2018-01-01")
					return
				}
			}
			else if (type == 3) {
				if (key.length < 3) {
					$("#intext").attr("placeholder","没有输入用户名称")
					return
				}
			}
		
			$.ajax({
	          async:true,
	          cache:false,
	          timeout:10000,
	          type:"POST",
	          url:"/{{.ManageUrl}}",
	          data:{
	          	subid_:subId,
	          	type_:type,
				key_:key,
		    	},
	          error:function(jqXHR, textStatus, errorThrown){
	            if(textStatus=="timeout"){
	              $("#info").text("查找超时，请重试...");
	            }else{
	              $("#info").text("查找失败!");
	            }
	          },
	          success:function(msg){
	          	if ("[0]" != msg.substr(0,3)) {
					$("#info").val(msg);
	          	} else {
					window.location.href=msg.substr(3)
	          	}
	          }
	            
	        });
		}

		function gradeChange() {
		var options=$("#sType option:selected");
		var type = options.val(); 
			if (type == 1 || type == 3) {
				$("#intext").attr("placeholder","请输入用户名")
			} else if (type == 2) {
				$("#intext").attr("placeholder","请输入日期时间，格式 2018-01-01")
			}
			
		}
</script>
</div>
</body>
</html>


