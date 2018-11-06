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
   	.h5 {
   	color:#009900;
   	}
	h4 {
    color:#4876FF;
    }
	.table-css {
		text-align: center;
	    color:#000033;
	}
	.key-prob {
		color:#0000CC;
		padding-right:2px;
	}
	.title-css {
	float: left;
	}
	.comment-css {
	background-color:#e0ffff;
	}
	.body-css {
	 background-color:#fcfdf8;
	}
	</style>
</head>
<body style="padding-top:55px;">
{{str2html .Nav}}
<div class="container-fluid">
<div class="row">

<div class="col-md-6 col-md-offset-2">
    <p class="crumbs"><a href="/">首页</a> &gt;  <a href="{{.HrefSub}}">{{.SubName}}</a></p>
    <h3>{{.Title}}</h3>
    <p><span class = "key-prob">作者:{{.UserName}}</span>
    <span class = "key-prob">类型:{{.Type}}</span>
    <span class = "key-prob">时间: {{.Date}}</span></p>

    <div id="story" class="body-css">
        {{str2html .Story}}
    </div>
	
	<h4>相关阅读</h4>
     {{range .Recommend}}
     <p class="p-css"><a class="a" href="cure{{.Id}}" target="_blank">{{.Title}}</a> </p>
     {{end}}

    <p class="h5">网友留言(最近100条)</p>
    
    {{range .Comments}}
       <p class="comment-css">{{.UserName}}  {{.Time}}</p>
       <p>{{.Content}}</p>
    {{end}}

    <p class="h5">你的回应 ...... (提示：字数不能超过255)</p>
	<textarea id="text" class="form-control" rows="3"></textarea>
	<script>
	function comment() {
		if ("游客1" == getUser()) {
			document.getElementById("botinfo").innerHTML=("您还没登录，不能留言，请先登录...");
			return
		}
			
		var text = document.getElementById("text").value;
		if (text.length < 1) {
			document.getElementById("botinfo").innerHTML=("错误：评论为空，请输入评论！");
			return 
		} else if (text.length > 255) {
			document.getElementById("botinfo").innerHTML=("提升：评论不能超过255个字，当前字数:" + text.length);
			return
		}
		var btn = $("#commit"); 
		
			
		$.ajax({
		          async:true,
		          cache:false,
		          timeout:10000,
		          type:"POST",
		          url:"/comment",
		          data:{
		          	sid_:{{.Sid}},
					aid_:{{.Aid}},
			        comment_:text,
			    	},
		          error:function(jqXHR, textStatus, errorThrown){
		            if(textStatus=="timeout"){
		              document.getElementById("botinfo").innerHTML=("提交点评超时，请稍后再提交...");
		            }else{
		              document.getElementById("botinfo").innerHTML=("提交点评失败!");
		            }
		          },
		          success:function(msg){
		          	if ("[0]" != msg.substr(0,3)) {
						document.getElementById("botinfo").innerHTML=(msg);
		          	} else {
					    btn.attr("disabled", true);
						document.getElementById("botinfo").innerHTML=("你的评论已经成功发布! 待系统审核后才能显示，请不要重复提交。");
						location.reload();
		          	}
		          }
		        });
			}
	</script>
	<button type="button" class="btn btn-default btn-sm" style="float:right;" id="commit" onclick="comment()">加上去</button>
	<p id="botinfo" class="alert alert-info">网站提示：点击右边按钮提交评论！</p>
	<script>
	if ("游客" == getUser()) {
		document.getElementById("botinfo").innerHTML=("您还没登录，不能留言，请先登录...");
	}
	</script>
</div>

</div>
</div>

</body>
</html>
