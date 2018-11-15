<!DOCTYPE html>
<html>
<head>
   <meta charset="utf-8">
   <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <meta name="description" content="{{.Title}}"/>
   <title>{{.Title}} Powered by Gentlesys!</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
   <style>
	.body-css {
		padding-top:55px;
	}
	.root-css {
	 background-color:#fcfdf8;
	}
	p{
   	color:#009900;
	}
   </style>
</head>

<body class="body-css">
{{str2html .Navigation}}

<div class="container-fluid">
	<div class="row root-css col-lg-10 col-lg-offset-1">
    <hr>
     <table class="table table-condensed"> 
        <tbody>
         <tr>
            <td>
            <p><strong>用户名称: </strong>{{.Name}}</p>
            <p><strong>级别：</strong>{{.Level}}<strong> 状态：</strong>{{.Status}}</p>
            <p><strong>注册日期: </strong>{{.Birth}}</p>
            <p><strong>最近登录日期: </strong>{{.Lastlog}}</p>
            <p><strong>邮箱: </strong>{{.Mail}}</p>
            <p><strong>发帖数量: </strong>{{.TlArticleNums}} <strong> 评论次数: </strong>{{.TlCommentTimes}}</p>
            </td>
         </tr>
         {{if .IsAdmin}}
         <tr>
          <td>
            <p id="info"><strong>提示:一切正常{{.Info}}</strong></p>
            <p><strong>以下内容仅管理员可见</strong></p>
            <p><strong>禁用或解禁: </strong><button type="submit" class="btn-xs btn-warning" onclick="setings(1)">点击设置</button></p>
            <p><strong>提升等级：</strong><button type="submit" class="btn-xs btn-warning" onclick="setings(2)">点击提升</button></p>
          </td>
         </tr>
         <script>
         function setings(type) {
         $.ajax({
	          async:true,
	          cache:false,
	          timeout:10000,
	          type:"POST",
	          url:"/user",
	          data:{
	          	userId_:{{.UserId}},
	          	type_:type,
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
	          }	            
	        });
		}
		</script>
         {{end}}
         </tbody>
       </table>
	</div>
</div>
</body>
</html>

