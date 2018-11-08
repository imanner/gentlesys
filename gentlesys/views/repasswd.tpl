<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8"> 
   <title>Gentlesys {{.Title}}</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="/static/bootsp/js/jquery.md5.js"></script>
    <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
    <style>
    .text {
   	color:#009900;
   	}
   	.text1 {
   	color:#009900;
   	font-size:24px;
   	}
	</style>
</head>
<body style="padding-top:55px;">
{{str2html .Navigation}}
<div class="container-fluid">
<div class="row">
    <div class="col-md-6 col-md-offset-2">
        <div>
        <p class="text1">重置账号 [ {{.User}} ] 的密码</p>
        </div>
        
        <form class="form-horizontal" role="form">
        	<div class="row form-group">
        		<label class="col-md-2 control-label text">输入新密码</label>
        		<div class="col-md-4">
        			<input type="password" class="form-control" id="passwd" 
        				   placeholder="6-12位字母或数字">
        		</div>
        	</div>
        	
        	<div class="row form-group">
        	<label class="col-md-2 control-label text">确认新密码</label>
        		<div class="col-md-4">
        			<input type="password" class="form-control" id="confirm" 
        				   placeholder="确认密码">
        		</div>
        	</div>
        </form>

         <div class="col-md-offset-2">
		<button id="register" type="submit" class="btn btn-default" onclick="back()">设置密码</button>
        <p id="info" class="alert">网站提示：输入新密码重置</p>
        </div>
        
    </div>
    <script>
		function back() {
			var passwd = document.getElementById("passwd").value;
			var confirm = document.getElementById("confirm").value;

			if (passwd.length < 6) {
				document.getElementById("info").innerHTML=("密码不能小于6位长度！");
				return 
			} else if(passwd.length > 32) {
				document.getElementById("info").innerHTML=("密码长度不能超过32个字符！");
				return 
			} else if (passwd != confirm) {
				document.getElementById("info").innerHTML=("两次输入密码不一致！");
				return 
			} 

			var md5Pwd=$.md5(passwd);

			document.getElementById("info").innerHTML=("正在重置中...");
			
			var btn = $("#register"); 
			btn.attr("disabled", true);

			$.ajax({
		          async:true,
		          cache:false,
		          timeout:10000,
		          type:"POST",
		          url:"/updatepd",
		          data:{
		           id_:{{.Id}},
				   passwd_:md5Pwd,
			      },
		          error:function(jqXHR, textStatus, errorThrown){
		            if(textStatus=="timeout"){
		              document.getElementById("info").innerHTML=("重置超时，请重试...");
		            }else{
		              document.getElementById("info").innerHTML=("重置失败!");
		            }
					btn.attr("disabled", false);
		          },
		          success:function(msg){
		          	if ("[0]" != msg.substr(0,3)) {
						btn.attr("disabled", false);
						document.getElementById("info").innerHTML=(msg);
		          	} else {
						document.getElementById("info").innerHTML=("重置成功！");
						window.location.href="/"
		          	}
		            
		          }
		        });
		}
    </script>
        
</div>
</div>
</body>
</html>

